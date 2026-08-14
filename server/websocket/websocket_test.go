package websocket_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"server/api"
	"server/auth"
	"server/database"
	"server/models"
	"server/protocol"
	"server/repository"
	"server/services"
	ws "server/websocket"
)

func setupTestServer(t *testing.T) (*httptest.Server, *auth.Service, *repository.Repository) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.MigrateDB(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repos := repository.New(db)
	svcs := services.New(repos, 24*time.Hour)
	authSvc := auth.NewService(repos.Users, "test-ws-secret", 1*time.Hour)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	wsServer := ws.NewServer(repos, logger)
	router := api.NewRouter(logger, authSvc, svcs, wsServer, "*")

	ts := httptest.NewServer(router)
	t.Cleanup(ts.Close)

	return ts, authSvc, repos
}

func connectWS(t *testing.T, serverURL, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{
			"Authorization": {"Bearer " + token},
		},
	})
	if err != nil {
		t.Fatalf("ws dial failed: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "done") })
	return conn
}

func connectWSWithQueryToken(t *testing.T, serverURL, token string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + strings.TrimPrefix(serverURL, "http") + "/ws?token=" + token
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("ws query-token dial failed: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "done") })
	return conn
}

// readMessage reads from the WS connection, skipping any server-originated
// presence envelopes (snapshot / live updates), and returns the first
// non-presence envelope.
func readMessage(t *testing.T, conn *websocket.Conn, ctx context.Context) protocol.Envelope {
	t.Helper()
	for {
		_, rData, err := conn.Read(ctx)
		if err != nil {
			t.Fatalf("read failed: %v", err)
		}
		var env protocol.Envelope
		if err := json.Unmarshal(rData, &env); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if env.Type != protocol.TypePresence && env.Type != protocol.TypePresenceSnapshot {
			return env
		}
	}
}

func TestWebSocketDirectMessaging(t *testing.T) {
	ts, authSvc, repos := setupTestServer(t)

	aliceResp, err := authSvc.Register("alice", "password123")
	if err != nil {
		t.Fatalf("register alice: %v", err)
	}
	aliceLogin, _ := authSvc.Login("alice", "password123")

	bobResp, err := authSvc.Register("bob", "password123")
	if err != nil {
		t.Fatalf("register bob: %v", err)
	}
	bobLogin, _ := authSvc.Login("bob", "password123")

	// Direct messages now require an accepted relationship.
	now := time.Now().Unix()
	repos.Relationships.Create(&models.Relationship{
		RequesterID: aliceResp.User.ID,
		RecipientID: bobResp.User.ID,
		Status:      models.RelationshipAccepted,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	aliceConn := connectWSWithQueryToken(t, ts.URL, aliceLogin.Token)
	bobConn := connectWSWithQueryToken(t, ts.URL, bobLogin.Token)

	payload, _ := json.Marshal(protocol.DirectMessagePayload{
		RecipientID: bobResp.User.ID,
		Content:     "Hello Bob!",
	})
	msgData, _ := json.Marshal(protocol.Envelope{
		Version: 1,
		Type:    protocol.TypeMessage,
		Payload: payload,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := aliceConn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("alice send failed: %v", err)
	}

	env := readMessage(t, bobConn, ctx)
	if env.Type != protocol.TypeMessage {
		t.Errorf("expected type message, got %s", env.Type)
	}

	dm, err := protocol.ParseDirectMessage(&env)
	if err != nil {
		t.Fatalf("parse dm: %v", err)
	}
	if dm.Content != "Hello Bob!" {
		t.Errorf("expected 'Hello Bob!', got %q", dm.Content)
	}
	_ = aliceResp
}

func TestWebSocketOfflineQueue(t *testing.T) {
	ts, authSvc, repos := setupTestServer(t)

	aliceResp, _ := authSvc.Register("alice2", "password123")
	aliceLogin, _ := authSvc.Login("alice2", "password123")
	bobResp, _ := authSvc.Register("bob2", "password123")
	bobLogin, _ := authSvc.Login("bob2", "password123")

	// Direct messages now require an accepted relationship.
	now := time.Now().Unix()
	repos.Relationships.Create(&models.Relationship{
		RequesterID: aliceResp.User.ID,
		RecipientID: bobResp.User.ID,
		Status:      models.RelationshipAccepted,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	aliceConn := connectWS(t, ts.URL, aliceLogin.Token)

	payload, _ := json.Marshal(protocol.DirectMessagePayload{
		RecipientID: bobResp.User.ID,
		Content:     "Offline message for Bob",
	})
	msgData, _ := json.Marshal(protocol.Envelope{
		Version: 1,
		Type:    protocol.TypeMessage,
		Payload: payload,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := aliceConn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("alice send failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	bobConn := connectWS(t, ts.URL, bobLogin.Token)

	env := readMessage(t, bobConn, ctx)
	dm, err := protocol.ParseDirectMessage(&env)
	if err != nil {
		t.Fatalf("parse dm: %v", err)
	}
	if dm.Content != "Offline message for Bob" {
		t.Errorf("expected 'Offline message for Bob', got %q", dm.Content)
	}
	_ = aliceResp
}

func TestWebSocketGroupFanOut(t *testing.T) {
	ts, authSvc, repos := setupTestServer(t)

	aliceResp, _ := authSvc.Register("alice3", "password123")
	aliceLogin, _ := authSvc.Login("alice3", "password123")
	bobResp, _ := authSvc.Register("bob3", "password123")
	bobLogin, _ := authSvc.Login("bob3", "password123")
	charlieResp, _ := authSvc.Register("charlie3", "password123")

	groupID := "test-group-1"
	repos.Groups.Create(&models.Group{ID: groupID, CreatedAt: time.Now().Unix()})
	repos.Groups.AddMember(&models.GroupMember{GroupID: groupID, UserID: aliceResp.User.ID, JoinedAt: time.Now().Unix()})
	repos.Groups.AddMember(&models.GroupMember{GroupID: groupID, UserID: bobResp.User.ID, JoinedAt: time.Now().Unix()})
	repos.Groups.AddMember(&models.GroupMember{GroupID: groupID, UserID: charlieResp.User.ID, JoinedAt: time.Now().Unix()})

	aliceConn := connectWS(t, ts.URL, aliceLogin.Token)
	bobConn := connectWS(t, ts.URL, bobLogin.Token)

	payload, _ := json.Marshal(protocol.GroupMessagePayload{
		GroupID: groupID,
		Content: "Hello Group!",
	})
	msgData, _ := json.Marshal(protocol.Envelope{
		Version: 1,
		Type:    protocol.TypeGroupMessage,
		Payload: payload,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := aliceConn.Write(ctx, websocket.MessageText, msgData); err != nil {
		t.Fatalf("alice send failed: %v", err)
	}

	env := readMessage(t, bobConn, ctx)
	gm, _ := protocol.ParseGroupMessage(&env)
	if gm.Content != "Hello Group!" {
		t.Errorf("expected 'Hello Group!', got %q", gm.Content)
	}

	time.Sleep(100 * time.Millisecond)

	charlieLogin, _ := authSvc.Login("charlie3", "password123")
	charlieConn := connectWS(t, ts.URL, charlieLogin.Token)

	env = readMessage(t, charlieConn, ctx)
	gm, _ = protocol.ParseGroupMessage(&env)
	if gm.Content != "Hello Group!" {
		t.Errorf("expected 'Hello Group!', got %q", gm.Content)
	}
}

func TestPresenceSnapshotAndBroadcast(t *testing.T) {
	ts, authSvc, repos := setupTestServer(t)

	aliceResp, _ := authSvc.Register("alice-presence", "password123")
	aliceLogin, _ := authSvc.Login("alice-presence", "password123")
	bobResp, _ := authSvc.Register("bob-presence", "password123")
	bobLogin, _ := authSvc.Login("bob-presence", "password123")

	now := time.Now().Unix()
	repos.Relationships.Create(&models.Relationship{
		RequesterID: aliceResp.User.ID,
		RecipientID: bobResp.User.ID,
		Status:      models.RelationshipAccepted,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	// Connect Alice first — she gets a presence snapshot (empty, Bob offline).
	aliceConn := connectWS(t, ts.URL, aliceLogin.Token)
	time.Sleep(50 * time.Millisecond)

	// Connect Bob — Bob should receive a snapshot with Alice online,
	// and Alice should receive a live "bob online" presence update.
	bobConn := connectWS(t, ts.URL, bobLogin.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Read Bob's presence snapshot
	_, rData, err := bobConn.Read(ctx)
	if err != nil {
		t.Fatalf("bob read snapshot failed: %v", err)
	}
	var snapEnv protocol.Envelope
	if err := json.Unmarshal(rData, &snapEnv); err != nil {
		t.Fatalf("unmarshal snapshot env: %v", err)
	}
	if snapEnv.Type != protocol.TypePresenceSnapshot {
		t.Fatalf("expected presence_snapshot, got %s", snapEnv.Type)
	}
	var snapshot protocol.PresenceSnapshotPayload
	if err := json.Unmarshal(snapEnv.Payload, &snapshot); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if len(snapshot.Online) != 1 || snapshot.Online[0] != aliceResp.User.ID {
		t.Errorf("expected alice online in snapshot, got %v", snapshot.Online)
	}

	// Read Alice: skip her empty snapshot, then read the live presence update.
	aliceConn.Read(ctx) // presence snapshot (empty)
	_, rData, err = aliceConn.Read(ctx)
	if err != nil {
		t.Fatalf("alice read presence failed: %v", err)
	}
	var presEnv protocol.Envelope
	if err := json.Unmarshal(rData, &presEnv); err != nil {
		t.Fatalf("unmarshal presence env: %v", err)
	}
	if presEnv.Type != protocol.TypePresence {
		t.Fatalf("expected presence, got %s", presEnv.Type)
	}
	var presence protocol.PresencePayload
	if err := json.Unmarshal(presEnv.Payload, &presence); err != nil {
		t.Fatalf("unmarshal presence: %v", err)
	}
	if presence.UserID != bobResp.User.ID || presence.Status != "online" {
		t.Errorf("expected bob online, got %+v", presence)
	}
}

func TestPresenceOfflineBroadcast(t *testing.T) {
	ts, authSvc, repos := setupTestServer(t)

	aliceResp, _ := authSvc.Register("alice-presence2", "password123")
	aliceLogin, _ := authSvc.Login("alice-presence2", "password123")
	bobResp, _ := authSvc.Register("bob-presence2", "password123")
	bobLogin, _ := authSvc.Login("bob-presence2", "password123")

	now := time.Now().Unix()
	repos.Relationships.Create(&models.Relationship{
		RequesterID: aliceResp.User.ID,
		RecipientID: bobResp.User.ID,
		Status:      models.RelationshipAccepted,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	aliceConn := connectWS(t, ts.URL, aliceLogin.Token)
	bobConn := connectWS(t, ts.URL, bobLogin.Token)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Drain Alice's buffer: snapshot + bob-online presence
	aliceConn.Read(ctx) // snapshot (empty)
	aliceConn.Read(ctx) // presence (bob online)

	// Disconnect Bob — Alice should receive a "bob offline" presence update.
	bobConn.Close(websocket.StatusNormalClosure, "disconnecting")
	time.Sleep(200 * time.Millisecond)

	_, rData, err := aliceConn.Read(ctx)
	if err != nil {
		t.Fatalf("alice read presence failed: %v", err)
	}
	var env protocol.Envelope
	if err := json.Unmarshal(rData, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
	if env.Type != protocol.TypePresence {
		t.Fatalf("expected presence, got %s", env.Type)
	}
	var presence protocol.PresencePayload
	if err := json.Unmarshal(env.Payload, &presence); err != nil {
		t.Fatalf("unmarshal presence: %v", err)
	}
	if presence.UserID != bobResp.User.ID || presence.Status != "offline" {
		t.Errorf("expected bob offline, got %+v", presence)
	}
}

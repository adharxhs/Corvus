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
	svcs := services.New(repos)
	authSvc := auth.NewService(repos.Users, "test-ws-secret", 1*time.Hour)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	wsServer := ws.NewServer(repos, logger)
	router := api.NewRouter(logger, authSvc, svcs, wsServer)

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

func TestWebSocketDirectMessaging(t *testing.T) {
	ts, authSvc, _ := setupTestServer(t)

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

	aliceConn := connectWS(t, ts.URL, aliceLogin.Token)
	bobConn := connectWS(t, ts.URL, bobLogin.Token)

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

	_, rData, err := bobConn.Read(ctx)
	if err != nil {
		t.Fatalf("bob read failed: %v", err)
	}

	var env protocol.Envelope
	if err := json.Unmarshal(rData, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
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
	ts, authSvc, _ := setupTestServer(t)

	aliceResp, _ := authSvc.Register("alice2", "password123")
	aliceLogin, _ := authSvc.Login("alice2", "password123")
	bobResp, _ := authSvc.Register("bob2", "password123")
	bobLogin, _ := authSvc.Login("bob2", "password123")

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

	_, rData, err := bobConn.Read(ctx)
	if err != nil {
		t.Fatalf("bob read offline message failed: %v", err)
	}

	var env protocol.Envelope
	if err := json.Unmarshal(rData, &env); err != nil {
		t.Fatalf("unmarshal env: %v", err)
	}
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

	_, rData, err := bobConn.Read(ctx)
	if err != nil {
		t.Fatalf("bob read failed: %v", err)
	}
	var env protocol.Envelope
	json.Unmarshal(rData, &env)
	gm, _ := protocol.ParseGroupMessage(&env)
	if gm.Content != "Hello Group!" {
		t.Errorf("expected 'Hello Group!', got %q", gm.Content)
	}

	time.Sleep(100 * time.Millisecond)

	charlieLogin, _ := authSvc.Login("charlie3", "password123")
	charlieConn := connectWS(t, ts.URL, charlieLogin.Token)

	_, rData, err = charlieConn.Read(ctx)
	if err != nil {
		t.Fatalf("charlie read offline group message failed: %v", err)
	}
	json.Unmarshal(rData, &env)
	gm, _ = protocol.ParseGroupMessage(&env)
	if gm.Content != "Hello Group!" {
		t.Errorf("expected 'Hello Group!', got %q", gm.Content)
	}
}

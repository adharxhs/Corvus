package tests

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"server/database"
	"server/models"
	"server/repository"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "corvus.db"))
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(db) })
	if err := database.MigrateDB(db); err != nil {
		t.Fatalf("running migrations: %v", err)
	}
	return db
}

func TestDatabaseMigrateCreatesSchema(t *testing.T) {
	db := openTestDB(t)

	var applied int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("reading schema_migrations: %v", err)
	}
	if applied != 10 {
		t.Errorf("expected 10 applied migrations, got %d", applied)
	}

	expected := []string{
		"users",
		"prekey_bundles",
		"pending_messages",
		"groups",
		"group_members",
		"sender_key_distribution",
		"relationships",
		"group_invites",
		"profile_pictures",
		"group_profile_pictures",
	}
	for _, name := range expected {
		var count int
		err := db.QueryRow(
			`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`, name,
		).Scan(&count)
		if err != nil {
			t.Fatalf("checking table %s: %v", name, err)
		}
		if count != 1 {
			t.Errorf("table %s not found", name)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := openTestDB(t)

	if err := database.MigrateDB(db); err != nil {
		t.Fatalf("rerunning migrations: %v", err)
	}

	var applied int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("reading schema_migrations: %v", err)
	}
	if applied != 10 {
		t.Errorf("expected 10 applied migrations after rerun, got %d", applied)
	}
}

func TestForeignKeyEnforcement(t *testing.T) {
	db := openTestDB(t)

	_, err := db.Exec(
		`INSERT INTO pending_messages (id, recipient_id, ciphertext) VALUES (?, ?, ?)`,
		"msg-1", "missing-user", []byte("ciphertext"),
	)
	if err == nil {
		t.Fatal("expected foreign key violation for unknown recipient")
	}

	_, err = db.Exec(
		`INSERT INTO group_members (group_id, user_id, joined_at) VALUES (?, ?, ?)`,
		"group-1", "missing-user", 1,
	)
	if err == nil {
		t.Fatal("expected foreign key violation for unknown group")
	}
}

func TestRepositoriesRoundTrip(t *testing.T) {
	db := openTestDB(t)
	repo := repository.New(db)

	now := int64(1722400000)

	alice := &models.User{ID: "u-1", Username: "alice", PasswordHash: "hash", CreatedAt: now}
	if err := repo.Users.Create(alice); err != nil {
		t.Fatalf("creating user: %v", err)
	}

	got, err := repo.Users.GetByUsername("alice")
	if err != nil {
		t.Fatalf("getting user by username: %v", err)
	}
	if got.ID != "u-1" || got.PasswordHash != "hash" || got.CreatedAt != now {
		t.Errorf("unexpected user: %+v", got)
	}

	dup := &models.User{ID: "u-2", Username: "alice", PasswordHash: "hash2", CreatedAt: now}
	if err := repo.Users.Create(dup); err == nil {
		t.Error("expected duplicate username error")
	}

	if _, err := repo.Users.GetByUsername("nobody"); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows for missing user, got %v", err)
	}

	bundle := &models.PrekeyBundle{
		UserID:                "u-1",
		IdentityKey:           []byte("ik"),
		SignedPrekey:          []byte("spk"),
		SignedPrekeySignature: []byte("sig"),
		OneTimePrekey:         []byte("opk"),
	}
	if err := repo.Prekeys.Upsert(bundle); err != nil {
		t.Fatalf("upserting prekey bundle: %v", err)
	}
	gotBundle, err := repo.Prekeys.GetByUserID("u-1")
	if err != nil {
		t.Fatalf("getting prekey bundle: %v", err)
	}
	if string(gotBundle.IdentityKey) != "ik" || string(gotBundle.OneTimePrekey) != "opk" {
		t.Errorf("unexpected prekey bundle: %+v", gotBundle)
	}

	bundle.OneTimePrekey = []byte("opk2")
	if err := repo.Prekeys.Upsert(bundle); err != nil {
		t.Fatalf("replacing prekey bundle: %v", err)
	}
	gotBundle, err = repo.Prekeys.GetByUserID("u-1")
	if err != nil {
		t.Fatalf("getting replaced prekey bundle: %v", err)
	}
	if string(gotBundle.OneTimePrekey) != "opk2" {
		t.Errorf("expected replaced one-time prekey, got %q", gotBundle.OneTimePrekey)
	}

	group := &models.Group{ID: "g-1", CreatedAt: now}
	if err := repo.Groups.Create(group); err != nil {
		t.Fatalf("creating group: %v", err)
	}
	if err := repo.Groups.AddMember(&models.GroupMember{GroupID: "g-1", UserID: "u-1", JoinedAt: now}); err != nil {
		t.Fatalf("adding member: %v", err)
	}
	members, err := repo.Groups.ListMembers("g-1")
	if err != nil {
		t.Fatalf("listing members: %v", err)
	}
	if len(members) != 1 || members[0].UserID != "u-1" {
		t.Errorf("unexpected members: %+v", members)
	}
	if err := repo.Groups.RemoveMember("g-1", "u-1"); err != nil {
		t.Fatalf("removing member: %v", err)
	}
	if members, _ = repo.Groups.ListMembers("g-1"); len(members) != 0 {
		t.Errorf("expected no members, got %+v", members)
	}

	msg := &models.PendingMessage{ID: "m-1", RecipientID: "u-1", Ciphertext: []byte("ct")}
	if err := repo.Messages.Insert(msg); err != nil {
		t.Fatalf("inserting pending message: %v", err)
	}
	queue, err := repo.Messages.FetchByRecipient("u-1")
	if err != nil {
		t.Fatalf("fetching pending messages: %v", err)
	}
	if len(queue) != 1 || string(queue[0].Ciphertext) != "ct" {
		t.Errorf("unexpected queue: %+v", queue)
	}
	if err := repo.Messages.MarkDelivered("m-1"); err != nil {
		t.Fatalf("marking message delivered: %v", err)
	}
	if err := repo.Messages.DeleteDelivered(); err != nil {
		t.Fatalf("deleting delivered messages: %v", err)
	}
	if queue, _ = repo.Messages.FetchByRecipient("u-1"); len(queue) != 0 {
		t.Errorf("expected empty queue, got %+v", queue)
	}

	sk := &models.SenderKeyDistribution{
		ID: "s-1", GroupID: "g-1", RecipientID: "u-1", Ciphertext: []byte("skct"),
	}
	if err := repo.SenderKeys.Enqueue(sk); err != nil {
		t.Fatalf("enqueueing sender key distribution: %v", err)
	}
	pending, err := repo.SenderKeys.FetchByRecipient("u-1")
	if err != nil {
		t.Fatalf("fetching sender key distributions: %v", err)
	}
	if len(pending) != 1 || string(pending[0].Ciphertext) != "skct" {
		t.Errorf("unexpected sender key distributions: %+v", pending)
	}
}

func TestWithTxRollback(t *testing.T) {
	db := openTestDB(t)
	repo := repository.New(db)

	now := int64(1722400000)
	err := repo.WithTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec(
			`INSERT INTO users (id, username, password_hash, created_at) VALUES (?, ?, ?, ?)`,
			"u-1", "bob", "hash", now,
		); err != nil {
			return err
		}
		return errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected transaction to fail")
	}

	if _, err := repo.Users.GetByUsername("bob"); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected rolled-back user to be absent, got %v", err)
	}
}

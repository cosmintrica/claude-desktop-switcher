package bridge

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// The tables and columns claude-bridge 0.3.0 creates that this package reads.
const bridgeSchema = `
CREATE TABLE sessions (key TEXT PRIMARY KEY, short TEXT NOT NULL, session_id TEXT, name TEXT, account TEXT, pid INTEGER,
	first_seen INTEGER NOT NULL, last_seen INTEGER NOT NULL, closed_at INTEGER, status TEXT);
CREATE TABLE messages (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER NOT NULL, from_key TEXT NOT NULL,
	from_label TEXT NOT NULL, to_key TEXT, to_label TEXT, body TEXT NOT NULL, reply_to INTEGER, wake INTEGER NOT NULL DEFAULT 0);
CREATE TABLE receipts (message_id INTEGER NOT NULL, session_key TEXT NOT NULL, read_at INTEGER, woke_at INTEGER,
	PRIMARY KEY (message_id, session_key));
CREATE TABLE claims (resource TEXT PRIMARY KEY, label TEXT NOT NULL, kind TEXT NOT NULL, owner_key TEXT NOT NULL,
	note TEXT, created_at INTEGER NOT NULL, expires_at INTEGER);
CREATE TABLE handoffs (id INTEGER PRIMARY KEY AUTOINCREMENT, created_at INTEGER NOT NULL, from_key TEXT NOT NULL,
	from_label TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, status TEXT NOT NULL DEFAULT 'open');
`

func TestReadShowsTheActiveSessionsAndWhatWaitsForThem(t *testing.T) {
	home, claudeDir := t.TempDir(), t.TempDir()
	now := time.Date(2026, 11, 9, 1, 0, 0, 0, time.UTC)
	ms := func(ago time.Duration) int64 { return now.Add(-ago).UnixMilli() }
	db, err := sql.Open("sqlite", filepath.Join(home, "bridge.db"))
	if err != nil {
		t.Fatal(err)
	}
	statements := []struct {
		query string
		args  []any
	}{
		{bridgeSchema, nil},
		{`INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []any{"local_a", "a", "cli-a", "Site design", "office@", 101, ms(time.Hour), ms(10 * time.Second), nil, nil}},
		{`INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []any{"local_b", "b", "cli-b", "", "private@", 102, ms(30 * time.Minute), ms(30 * time.Second), nil, nil}},
		{`INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []any{"local_c", "c", "cli-c", "Gone quiet", "office@", 103, ms(time.Hour), ms(5 * time.Minute), nil, nil}},
		{`INSERT INTO sessions VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, []any{"local_d", "d", "cli-d", "Closed", "office@", 104, ms(time.Hour), ms(time.Second), ms(time.Second), nil}},
		// 1 to a, unread: counts. 2 to a, read. 3 to c, which is not active.
		// 4 from a to everyone, after b joined, unread by b: counts once.
		// 5 from c to everyone, read by a and b. 6 from a to everyone, before b joined.
		{`INSERT INTO messages (id, created_at, from_key, from_label, to_key, body) VALUES
			(1, ?, 'local_b', 'b', 'local_a', 'x'), (2, ?, 'local_b', 'b', 'local_a', 'x'), (3, ?, 'local_a', 'a', 'local_c', 'x'),
			(4, ?, 'local_a', 'a', NULL, 'x'), (5, ?, 'local_c', 'c', NULL, 'x'), (6, ?, 'local_a', 'a', NULL, 'x')`,
			[]any{ms(time.Minute), ms(time.Minute), ms(time.Minute), ms(10 * time.Minute), ms(10 * time.Minute), ms(45 * time.Minute)}},
		{`INSERT INTO receipts VALUES (1, 'local_a', NULL, NULL), (2, 'local_a', ?, NULL), (3, 'local_c', NULL, NULL), (5, 'local_a', ?, NULL), (5, 'local_b', ?, NULL)`,
			[]any{ms(time.Minute), ms(time.Minute), ms(time.Minute)}},
		{`INSERT INTO claims VALUES ('f1', 'f1', 'file', 'local_a', NULL, ?, NULL), ('f2', 'f2', 'file', 'local_a', NULL, ?, ?), ('f3', 'f3', 'file', 'local_c', NULL, ?, NULL)`,
			[]any{ms(time.Hour), ms(time.Hour), ms(time.Minute), ms(time.Hour)}},
		{`INSERT INTO handoffs (created_at, from_key, from_label, title, body, status) VALUES (?, 'local_a', 'a', 'next', 'x', 'open'), (?, 'local_a', 'a', 'done', 'x', 'taken')`,
			[]any{ms(time.Hour), ms(time.Hour)}},
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	db.Close()
	records := map[string]string{
		"101.json": `{"pid": 101, "status": "busy", "hostSessionId": "local_a"}`,
		"102.json": `{"pid": 102, "status": "idle", "sessionId": "cli-b"}`,
	}
	if err := os.MkdirAll(filepath.Join(claudeDir, "sessions"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, body := range records {
		if err := os.WriteFile(filepath.Join(claudeDir, "sessions", name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	state, err := Read(home, claudeDir, now)
	if err != nil {
		t.Fatal(err)
	}
	want := State{
		Sessions: []Session{{Name: "Site design", Account: "office@", Busy: true}, {Account: "private@"}},
		Unread:   2,
		Claims:   1,
		Handoffs: 1,
	}
	if !reflect.DeepEqual(state, want) {
		t.Fatalf("state = %+v\nwant %+v", state, want)
	}
}

func TestReadSaysWhenTheBridgeIsNotInstalled(t *testing.T) {
	if _, err := Read(t.TempDir(), t.TempDir(), time.Now()); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("err = %v", err)
	}
}

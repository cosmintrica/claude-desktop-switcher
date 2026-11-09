// Package bridge reads what claude-bridge knows right now. claude-bridge is a
// local MCP server through which the Claude Code sessions of every account on
// this computer talk; its processes keep one SQLite file whose schema only
// ever grows, and this package only reads it.
package bridge

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/TinoA/claude-desktop-switcher/internal/sessions"
	_ "modernc.org/sqlite"
)

// liveFor is how recently a session must have checked in to count as active,
// as claude-bridge counts it.
const liveFor = 2 * time.Minute

// State is what the bridge knows now.
type State struct {
	Sessions []Session
	Unread   int // messages some active session has not read yet
	Claims   int // files, folders or tasks held by active sessions
	Handoffs int // handoffs nobody has taken yet
}

// Session is one active session on the bridge.
type Session struct {
	Name, Account string
	Busy          bool
}

// Read reads the bridge's database in home (~/.claude-bridge) and, for each
// active session, Claude Code's record in claudeDir (~/.claude) of whether it
// runs a turn. An error matching os.ErrNotExist means the bridge is not
// installed.
func Read(home, claudeDir string, now time.Time) (State, error) {
	file := filepath.Join(home, "bridge.db")
	if _, err := os.Stat(file); err != nil {
		return State{}, err
	}
	db, err := sql.Open("sqlite", readOnly(file))
	if err != nil {
		return State{}, err
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	live := now.Add(-liveFor).UnixMilli()
	rows, err := db.QueryContext(ctx, `SELECT key, COALESCE(session_id, ''), COALESCE(name, ''), COALESCE(account, ''), COALESCE(pid, 0)
		FROM sessions WHERE closed_at IS NULL AND last_seen >= ? ORDER BY last_seen DESC`, live)
	if err != nil {
		return State{}, err
	}
	defer rows.Close()
	var state State
	for rows.Next() {
		var key, id, name, account string
		var pid int
		if err := rows.Scan(&key, &id, &name, &account, &pid); err != nil {
			return State{}, err
		}
		state.Sessions = append(state.Sessions, Session{Name: name, Account: account, Busy: sessions.Busy(claudeDir, pid, key, id)})
	}
	if err := rows.Err(); err != nil {
		return State{}, err
	}
	counts := []struct {
		into  *int
		query string
		args  []any
	}{
		// As claude-bridge counts a session's unread messages: those sent to it,
		// and those sent to everyone by another session since it joined, that it
		// has not marked read.
		{&state.Unread, `SELECT COUNT(DISTINCT m.id) FROM messages m JOIN sessions s ON s.closed_at IS NULL AND s.last_seen >= ?
			WHERE (m.to_key = s.key OR (m.to_key IS NULL AND m.from_key <> s.key AND m.created_at >= s.first_seen))
			AND NOT EXISTS (SELECT 1 FROM receipts r WHERE r.message_id = m.id AND r.session_key = s.key AND r.read_at IS NOT NULL)`, []any{live}},
		{&state.Claims, `SELECT COUNT(*) FROM claims c JOIN sessions s ON s.key = c.owner_key
			WHERE s.closed_at IS NULL AND s.last_seen >= ? AND (c.expires_at IS NULL OR c.expires_at > ?)`, []any{live, now.UnixMilli()}},
		{&state.Handoffs, `SELECT COUNT(*) FROM handoffs WHERE status = 'open'`, nil},
	}
	for _, count := range counts {
		if err := db.QueryRowContext(ctx, count.query, count.args...).Scan(count.into); err != nil {
			return State{}, err
		}
	}
	return state, nil
}

func readOnly(file string) string {
	path := filepath.ToSlash(file)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	dsn := &url.URL{Scheme: "file", Path: path}
	query := dsn.Query()
	query.Set("mode", "ro")
	query.Add("_pragma", "busy_timeout(2000)")
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

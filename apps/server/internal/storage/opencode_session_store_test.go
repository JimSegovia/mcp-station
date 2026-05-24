package storage

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	oldDB := DB
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	DB = db
	migrate()

	t.Cleanup(func() {
		DB.Close()
		DB = oldDB
	})
}

func TestCleanExpiredSessionsUsesLatestMessageTimestamp(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	cutoff := now.Add(-24 * time.Hour)
	stale := cutoff.Add(-2 * time.Hour).Format(time.RFC3339)
	fresh := cutoff.Add(2 * time.Hour).Format(time.RFC3339)

	if _, err := DB.Exec(`INSERT INTO opencode_sessions (id, session_id, tool_name, prompt, created_at) VALUES (?, ?, ?, ?, ?)`,
		"old-1", "session-shared", "opencode_ask", "first", stale); err != nil {
		t.Fatalf("insert shared old message: %v", err)
	}
	if _, err := DB.Exec(`INSERT INTO opencode_sessions (id, session_id, tool_name, prompt, created_at) VALUES (?, ?, ?, ?, ?)`,
		"new-1", "session-shared", "opencode_ask", "second", fresh); err != nil {
		t.Fatalf("insert shared fresh message: %v", err)
	}
	if _, err := DB.Exec(`INSERT INTO opencode_sessions (id, session_id, tool_name, prompt, created_at) VALUES (?, ?, ?, ?, ?)`,
		"old-2", "session-stale", "opencode_ask", "stale", stale); err != nil {
		t.Fatalf("insert stale message: %v", err)
	}
	if err := UpsertOpenCodeSessionBinding(XiaoZhiTestScopeKey(), "session-stale", now.Add(-48*time.Hour)); err != nil {
		t.Fatalf("upsert binding: %v", err)
	}

	cleaned, err := CleanExpiredSessions(24 * time.Hour)
	if err != nil {
		t.Fatalf("clean sessions: %v", err)
	}

	if len(cleaned) != 1 || cleaned[0] != "session-stale" {
		t.Fatalf("unexpected cleaned sessions: %#v", cleaned)
	}

	var remaining int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM opencode_sessions WHERE session_id = ?`, "session-shared").Scan(&remaining); err != nil {
		t.Fatalf("count remaining shared session: %v", err)
	}
	if remaining != 2 {
		t.Fatalf("expected shared session history to remain, got %d", remaining)
	}

	binding, err := GetOpenCodeSessionBinding(XiaoZhiTestScopeKey())
	if err != nil {
		t.Fatalf("get binding: %v", err)
	}
	if binding != nil {
		t.Fatalf("expected stale binding to be removed, got %#v", binding)
	}
}

func TestOpenCodeSessionBindingFreshness(t *testing.T) {
	setupTestDB(t)

	now := time.Now().UTC()
	if err := UpsertOpenCodeSessionBinding(XiaoZhiTestScopeKey(), "session-1", now); err != nil {
		t.Fatalf("upsert binding: %v", err)
	}

	binding, err := GetOpenCodeSessionBinding(XiaoZhiTestScopeKey())
	if err != nil {
		t.Fatalf("get binding: %v", err)
	}
	if !IsOpenCodeSessionBindingFresh(binding, 15*time.Minute, now.Add(10*time.Minute)) {
		t.Fatalf("expected binding to be fresh")
	}
	if IsOpenCodeSessionBindingFresh(binding, 15*time.Minute, now.Add(16*time.Minute)) {
		t.Fatalf("expected binding to expire after idle window")
	}
}

func XiaoZhiTestScopeKey() string {
	return "xiaozhi_mcp_test"
}

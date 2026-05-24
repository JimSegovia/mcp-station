package opencode

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mcp-station/server/internal/storage"
	_ "modernc.org/sqlite"
)

func setupScopedSessionTestDB(t *testing.T) {
	t.Helper()

	oldDB := storage.DB
	dbPath := filepath.Join(t.TempDir(), "test.db")

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	db.SetMaxOpenConns(1)
	storage.DB = db
	for _, query := range []string{
		`CREATE TABLE IF NOT EXISTS opencode_sessions (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			tool_name TEXT NOT NULL DEFAULT '',
			prompt TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS opencode_session_bindings (
			scope_key TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			last_used_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
	} {
		if _, err := storage.DB.Exec(query); err != nil {
			t.Fatalf("create test schema: %v", err)
		}
	}

	t.Cleanup(func() {
		storage.DB.Close()
		storage.DB = oldDB
	})
}

type scopedSessionHarness struct {
	mu              sync.Mutex
	nextID          int
	createdSessions []string
	deleted         map[string]bool
}

func newScopedSessionServer(t *testing.T) (*httptest.Server, *scopedSessionHarness) {
	t.Helper()

	h := &scopedSessionHarness{
		nextID:  1,
		deleted: make(map[string]bool),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.mu.Lock()
		defer h.mu.Unlock()

		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/session":
			id := fmt.Sprintf("session-%d", h.nextID)
			h.nextID++
			h.createdSessions = append(h.createdSessions, id)
			writeJSON(t, w, SessionInfo{ID: id, Title: id})
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/session/") && strings.HasSuffix(r.URL.Path, "/message"):
			sessionID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/session/"), "/message")
			if h.deleted[sessionID] {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			writeJSON(t, w, MessageResponse{
				Info: MessageInfo{ID: "msg-" + sessionID, Role: "assistant"},
				Parts: []MessagePart{
					{Type: "text", Text: "reply from " + sessionID},
				},
			})
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/session/"):
			sessionID := strings.TrimPrefix(r.URL.Path, "/session/")
			if h.deleted[sessionID] {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			writeJSON(t, w, SessionItem{ID: sessionID, Title: sessionID, Status: "ready"})
		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(server.Close)
	return server, h
}

func TestAskWithScopedSessionReusesSessionWithinIdleWindow(t *testing.T) {
	setupScopedSessionTestDB(t)
	server, _ := newScopedSessionServer(t)

	client := NewClient(server.URL, "", "")
	scopeKey := "xiaozhi_mcp"

	first, err := client.AskWithScopedSession(scopeKey, "hola", "build", "", 15*time.Minute)
	if err != nil {
		t.Fatalf("first scoped ask: %v", err)
	}
	second, err := client.AskWithScopedSession(scopeKey, "seguimos", "build", "", 15*time.Minute)
	if err != nil {
		t.Fatalf("second scoped ask: %v", err)
	}

	if first.SessionID != second.SessionID {
		t.Fatalf("expected same session, got %s and %s", first.SessionID, second.SessionID)
	}
}

func TestAskWithScopedSessionCreatesNewSessionAfterIdleExpiration(t *testing.T) {
	setupScopedSessionTestDB(t)
	server, _ := newScopedSessionServer(t)

	client := NewClient(server.URL, "", "")
	scopeKey := "xiaozhi_mcp"

	first, err := client.AskWithScopedSession(scopeKey, "hola", "build", "", 15*time.Minute)
	if err != nil {
		t.Fatalf("first scoped ask: %v", err)
	}

	staleTime := time.Now().UTC().Add(-16 * time.Minute).Format(time.RFC3339)
	if _, err := storage.DB.Exec(`UPDATE opencode_session_bindings SET last_used_at = ? WHERE scope_key = ?`, staleTime, scopeKey); err != nil {
		t.Fatalf("stale binding: %v", err)
	}

	second, err := client.AskWithScopedSession(scopeKey, "nuevo contexto", "build", "", 15*time.Minute)
	if err != nil {
		t.Fatalf("second scoped ask after idle: %v", err)
	}

	if first.SessionID == second.SessionID {
		t.Fatalf("expected new session after idle expiration")
	}
}

func TestAskWithScopedSessionRecreatesDeletedSession(t *testing.T) {
	setupScopedSessionTestDB(t)
	server, harness := newScopedSessionServer(t)

	client := NewClient(server.URL, "", "")
	scopeKey := "xiaozhi_mcp"

	first, err := client.AskWithScopedSession(scopeKey, "hola", "build", "", 15*time.Minute)
	if err != nil {
		t.Fatalf("first scoped ask: %v", err)
	}

	harness.mu.Lock()
	harness.deleted[first.SessionID] = true
	harness.mu.Unlock()

	second, err := client.AskWithScopedSession(scopeKey, "recupera", "build", "", 15*time.Minute)
	if err != nil {
		t.Fatalf("scoped ask after deleted session: %v", err)
	}

	if first.SessionID == second.SessionID {
		t.Fatalf("expected deleted session to be replaced")
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode json: %v", err)
	}
}

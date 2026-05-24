package storage

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type OpenCodeSession struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	ToolName  string `json:"toolName"`
	Prompt    string `json:"prompt"`
	CreatedAt string `json:"createdAt"`
}

type OpenCodeSessionBinding struct {
	ScopeKey   string `json:"scopeKey"`
	SessionID  string `json:"sessionId"`
	CreatedAt  string `json:"createdAt"`
	LastUsedAt string `json:"lastUsedAt"`
}

func TrackOpenCodeSession(sessionID, toolName, prompt string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	id := "ocs-" + uuid.NewString()[:8]

	_, err := DB.Exec(`INSERT INTO opencode_sessions (id, session_id, tool_name, prompt, created_at) VALUES (?, ?, ?, ?, ?)`,
		id, sessionID, toolName, prompt, now)
	return err
}

func ListTrackedSessions() ([]OpenCodeSession, error) {
	rows, err := DB.Query(`SELECT id, session_id, tool_name, prompt, created_at FROM opencode_sessions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []OpenCodeSession
	for rows.Next() {
		var s OpenCodeSession
		if err := rows.Scan(&s.ID, &s.SessionID, &s.ToolName, &s.Prompt, &s.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func DeleteTrackedSession(trackID string) error {
	_, err := DB.Exec(`DELETE FROM opencode_sessions WHERE id = ?`, trackID)
	return err
}

func DeleteTrackedSessionsBySessionID(sessionID string) error {
	_, err := DB.Exec(`DELETE FROM opencode_sessions WHERE session_id = ?`, sessionID)
	return err
}

func GetOpenCodeSessionBinding(scopeKey string) (*OpenCodeSessionBinding, error) {
	row := DB.QueryRow(`SELECT scope_key, session_id, created_at, last_used_at FROM opencode_session_bindings WHERE scope_key = ?`, scopeKey)

	var binding OpenCodeSessionBinding
	if err := row.Scan(&binding.ScopeKey, &binding.SessionID, &binding.CreatedAt, &binding.LastUsedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &binding, nil
}

func UpsertOpenCodeSessionBinding(scopeKey, sessionID string, now time.Time) error {
	ts := now.UTC().Format(time.RFC3339)
	_, err := DB.Exec(`
		INSERT INTO opencode_session_bindings (scope_key, session_id, created_at, last_used_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(scope_key) DO UPDATE SET
			session_id = excluded.session_id,
			created_at = excluded.created_at,
			last_used_at = excluded.last_used_at
	`, scopeKey, sessionID, ts, ts)
	return err
}

func TouchOpenCodeSessionBinding(scopeKey string, now time.Time) error {
	ts := now.UTC().Format(time.RFC3339)
	_, err := DB.Exec(`UPDATE opencode_session_bindings SET last_used_at = ? WHERE scope_key = ?`, ts, scopeKey)
	return err
}

func DeleteOpenCodeSessionBinding(scopeKey string) error {
	_, err := DB.Exec(`DELETE FROM opencode_session_bindings WHERE scope_key = ?`, scopeKey)
	return err
}

func DeleteOpenCodeSessionBindingsBySessionID(sessionID string) error {
	_, err := DB.Exec(`DELETE FROM opencode_session_bindings WHERE session_id = ?`, sessionID)
	return err
}

func IsOpenCodeSessionBindingFresh(binding *OpenCodeSessionBinding, maxAge time.Duration, now time.Time) bool {
	if binding == nil {
		return false
	}

	lastUsedAt, err := time.Parse(time.RFC3339, binding.LastUsedAt)
	if err != nil {
		return false
	}

	return now.UTC().Sub(lastUsedAt.UTC()) <= maxAge
}

func CleanExpiredSessions(maxAge time.Duration) ([]string, error) {
	cutoff := time.Now().UTC().Add(-maxAge).Format(time.RFC3339)

	rows, err := DB.Query(`
		SELECT session_id
		FROM opencode_sessions
		GROUP BY session_id
		HAVING MAX(created_at) < ?
	`, cutoff)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessionIDs []string
	for rows.Next() {
		var sid string
		if err := rows.Scan(&sid); err != nil {
			return nil, err
		}
		sessionIDs = append(sessionIDs, sid)
	}

	for _, sid := range sessionIDs {
		DB.Exec(`DELETE FROM opencode_sessions WHERE session_id = ?`, sid)
		DeleteOpenCodeSessionBindingsBySessionID(sid)
	}

	return sessionIDs, nil
}

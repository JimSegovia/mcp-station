package storage

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init() {
	dir := filepath.Join(".", "data")
	os.MkdirAll(dir, 0755)

	var err error
	DB, err = sql.Open("sqlite", filepath.Join(dir, "station.db"))
	if err != nil {
		log.Fatalf("storage: open db: %v", err)
	}

	DB.SetMaxOpenConns(1)

	migrate()
	seed()
}

func migrate() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS integrations (
			id INTEGER PRIMARY KEY DEFAULT 1,
			endpoint TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'empty',
			tools_json TEXT NOT NULL DEFAULT '[]',
			last_connected_at TEXT,
			last_error TEXT,
			server_port INTEGER NOT NULL DEFAULT 8090,
			protocol_version TEXT NOT NULL DEFAULT 'MCP/JSON-RPC 2.0',
			uptime_started_at TEXT,
			latency_ms INTEGER NOT NULL DEFAULT 0,
			health_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS modules (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			status TEXT NOT NULL DEFAULT 'ok',
			last_error TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS servers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'websocket',
			endpoint TEXT NOT NULL DEFAULT '',
			enabled INTEGER NOT NULL DEFAULT 1,
			status TEXT NOT NULL DEFAULT 'disconnected',
			tools_json TEXT NOT NULL DEFAULT '[]',
			last_connected_at TEXT,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		)`,
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
	}

	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			log.Fatalf("storage: migrate: %v", err)
		}
	}
}

func seed() {
	var count int
	DB.QueryRow("SELECT COUNT(*) FROM integrations").Scan(&count)
	if count > 0 {
		return
	}

	now := time.Now().UTC().Format(time.RFC3339)

	DB.Exec(`INSERT OR IGNORE INTO integrations (id, endpoint, status, server_port, protocol_version, created_at, updated_at) VALUES (1, '', 'empty', 8090, 'MCP/JSON-RPC 2.0', ?, ?)`, now, now)

	log.Println("storage: seeded database")
}

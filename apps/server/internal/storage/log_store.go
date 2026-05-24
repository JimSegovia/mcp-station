package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/mcp-station/server/internal/model"
)

func GetLogs(query model.LogQuery) ([]model.LogEntry, error) {
	where := []string{}
	args := []interface{}{}

	if query.Type != "" {
		where = append(where, "type = ?")
		args = append(args, query.Type)
	}
	if query.Source != "" {
		where = append(where, "source = ?")
		args = append(args, query.Source)
	}
	if query.Result != "" {
		where = append(where, "result = ?")
		args = append(args, query.Result)
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	limit := 50
	if query.Limit > 0 {
		limit = query.Limit
	}
	args = append(args, limit)

	sql := fmt.Sprintf("SELECT id, timestamp, type, source, message, result FROM logs %s ORDER BY timestamp DESC LIMIT ?", whereClause)

	rows, err := DB.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.LogEntry
	for rows.Next() {
		var l model.LogEntry
		if err := rows.Scan(&l.ID, &l.Timestamp, &l.Type, &l.Source, &l.Message, &l.Result); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	if logs == nil {
		logs = []model.LogEntry{}
	}

	return logs, nil
}

func AddLog(logType, source, message, result string) (*model.LogEntry, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	id := "log-" + uuid.NewString()[:8]

	_, err := DB.Exec(`INSERT INTO logs (id, timestamp, type, source, message, result) VALUES (?, ?, ?, ?, ?, ?)`,
		id, now, logType, source, message, result)
	if err != nil {
		return nil, err
	}

	return &model.LogEntry{
		ID:        id,
		Timestamp: now,
		Type:      logType,
		Source:    source,
		Message:   message,
		Result:    result,
	}, nil
}

func DeleteLogs() error {
	_, err := DB.Exec("DELETE FROM logs")
	return err
}

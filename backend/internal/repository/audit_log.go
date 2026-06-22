package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
)

// AuditLogRow is the response shape for a single audit log entry with the
// actor name/email joined from the users table.
type AuditLogRow struct {
	ID         int64             `json:"id"`
	ActorID    *string           `json:"actor_id"`
	ActorName  *string           `json:"actor_name"`
	ActorEmail *string           `json:"actor_email"`
	TargetType string            `json:"target_type"`
	TargetID   string            `json:"target_id"`
	Action     string            `json:"action"`
	Metadata   *json.RawMessage  `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
}

// AuditLogFilter carries optional filters for ListAuditLog.
type AuditLogFilter struct {
	ActorID    string
	From       string // RFC3339 or YYYY-MM-DD, passed through to SQL as-is
	To         string
	TargetType string
	Q          string // case-insensitive action keyword search
	Cursor     string
	Limit      int
}

// ListAuditLog returns audit log entries with an optional LEFT JOIN to the
// users table for actor display. Cursor-paginated by audit_log.id (BIGSERIAL).
func (r *Repository) ListAuditLog(ctx context.Context, filter AuditLogFilter) ([]AuditLogRow, string, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	query := `SELECT al.id, al.actor_id, u.name, u.email,
		al.target_type, al.target_id, al.action, al.metadata, al.created_at
		FROM audit_log al
		LEFT JOIN users u ON u.id = al.actor_id
		WHERE 1=1`
	args := []any{}
	argNum := 1

	if filter.ActorID != "" {
		query += fmt.Sprintf(` AND al.actor_id = $%d`, argNum)
		args = append(args, filter.ActorID)
		argNum++
	}
	if filter.From != "" {
		query += fmt.Sprintf(` AND al.created_at >= $%d::timestamptz`, argNum)
		args = append(args, filter.From)
		argNum++
	}
	if filter.To != "" {
		query += fmt.Sprintf(` AND al.created_at <= $%d::timestamptz`, argNum)
		args = append(args, filter.To)
		argNum++
	}
	if filter.TargetType != "" {
		query += fmt.Sprintf(` AND al.target_type = $%d`, argNum)
		args = append(args, filter.TargetType)
		argNum++
	}
	if filter.Q != "" {
		query += fmt.Sprintf(` AND al.action ILIKE $%d`, argNum)
		args = append(args, "%"+filter.Q+"%")
		argNum++
	}
	if filter.Cursor != "" {
		query += fmt.Sprintf(` AND al.id < $%d`, argNum)
		args = append(args, filter.Cursor)
		argNum++
	}

	query += fmt.Sprintf(` ORDER BY al.created_at DESC, al.id DESC LIMIT $%d`, argNum)
	args = append(args, filter.Limit+1)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var entries []AuditLogRow
	var nextCursor int64

	for rows.Next() {
		var e AuditLogRow
		var metadataBytes []byte
		if err := rows.Scan(
			&e.ID, &e.ActorID, &e.ActorName, &e.ActorEmail,
			&e.TargetType, &e.TargetID, &e.Action, &metadataBytes, &e.CreatedAt,
		); err != nil {
			return nil, "", err
		}
		if metadataBytes != nil {
			rm := json.RawMessage(metadataBytes)
			e.Metadata = &rm
		}
		if len(entries) < filter.Limit {
			entries = append(entries, e)
		} else {
			nextCursor = e.ID
		}
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	nextCursorStr := ""
	if nextCursor > 0 {
		nextCursorStr = strconv.FormatInt(nextCursor, 10)
	}

	return entries, nextCursorStr, nil
}

// InsertAuditLogMeta writes an audit log row with a nullable actor, a nullable
// JSONB metadata payload, and optional transaction support. Pass nil for actorID
// to set SQL NULL; pass nil or empty map for metadata to omit.
func (r *Repository) InsertAuditLogMeta(ctx context.Context, tx pgx.Tx, actorID *string, targetType, targetID, action string, metadata map[string]any) error {
	var metadataJSON []byte
	if len(metadata) > 0 {
		var err error
		metadataJSON, err = json.Marshal(metadata)
		if err != nil {
			return err
		}
	}

	var err error
	if tx != nil {
		_, err = tx.Exec(ctx,
			`INSERT INTO audit_log (actor_id, target_type, target_id, action, metadata) VALUES ($1, $2, $3, $4, $5)`,
			actorID, targetType, targetID, action, metadataJSON,
		)
	} else {
		_, err = r.pool.Exec(ctx,
			`INSERT INTO audit_log (actor_id, target_type, target_id, action, metadata) VALUES ($1, $2, $3, $4, $5)`,
			actorID, targetType, targetID, action, metadataJSON,
		)
	}
	return err
}

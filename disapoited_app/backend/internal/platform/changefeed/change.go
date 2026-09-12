// Package changefeed appends user-scoped changes inside the caller transaction.
package changefeed

import (
	"context"
	"encoding/json"
	"mypocket/internal/platform/commandtx"
)

func Append(ctx context.Context, tx commandtx.Queryer, userID, entity, id, operation string, version int64, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	var cursor int64
	err = tx.QueryRowContext(ctx, `INSERT INTO sync_cursors(user_id,next_cursor) VALUES($1,2)
 ON CONFLICT(user_id) DO UPDATE SET next_cursor=sync_cursors.next_cursor+1
 RETURNING next_cursor-1`, userID).Scan(&cursor)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO sync_changes(user_id,cursor,entity_type,entity_id,operation,version,payload_json)
 VALUES($1,$2,$3,$4,$5,$6,$7)`, userID, cursor, entity, id, operation, version, body)
	return err
}

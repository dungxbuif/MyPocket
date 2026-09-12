package identity

import (
	"context"
	"database/sql"
	"fmt"
)

func RequireOwner(ctx context.Context, conn *sql.DB, userID string, objectID string) error {
	var exists bool
	if err := conn.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM ownership_harness_objects
			WHERE id = $1 AND user_id = $2
		)
	`, objectID, userID).Scan(&exists); err != nil {
		return fmt.Errorf("check ownership: %w", err)
	}
	if !exists {
		return ErrForbidden
	}
	return nil
}

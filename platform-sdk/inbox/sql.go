// Package inbox deduplicates message effects in the same PostgreSQL transaction.
package inbox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Execute inserts the unique marker before fn, so concurrent duplicates serialize.
// The caller MUST roll back tx on any error and commit only on success.
func Execute(ctx context.Context, tx *sql.Tx, consumer, messageID string, fn func(context.Context) error) (bool, error) {
	if tx == nil || consumer == "" || messageID == "" || fn == nil {
		return false, errors.New("invalid inbox request")
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO platform.inbox (consumer,message_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, consumer, messageID)
	if err != nil {
		return false, fmt.Errorf("claim inbox: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if n == 0 {
		return true, nil
	}
	return false, fn(ctx)
}

package postgres

import (
	"context"
	"database/sql"
)

// TxManager shows the intended transaction boundary. The v1 scaffold executes
// the callback directly until SQL-backed repositories are implemented.
type TxManager struct {
	DB *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{DB: db}
}

func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if fn == nil {
		return nil
	}
	return fn(ctx)
}

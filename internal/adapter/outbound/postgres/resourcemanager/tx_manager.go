package postgres

import "context"

type TxManager struct{}

func (TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

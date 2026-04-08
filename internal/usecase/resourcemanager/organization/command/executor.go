package organizationcommand

import (
	"context"

	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
	"github.com/m8platform/platform/internal/usecase/resourcemanager/port"
)

type CommandExecutor struct {
	TxManager        port.TxManager
	IdempotencyStore port.IdempotencyStore
}

func (e CommandExecutor) Execute(ctx context.Context, scope string, idempotencyKey string, fn func(context.Context) error) error {
	run := func(txCtx context.Context) error {
		reservation, err := usecasecommon.ReserveIdempotency(txCtx, e.IdempotencyStore, scope, idempotencyKey, usecasecommon.DefaultIdempotencyTTL)
		if err != nil {
			return err
		}

		if err := fn(txCtx); err != nil {
			return err
		}
		return usecasecommon.CompleteIdempotency(txCtx, e.IdempotencyStore, reservation)
	}

	if e.TxManager == nil {
		return run(ctx)
	}
	return e.TxManager.WithinTx(ctx, run)
}

// Package idempotency implements transaction-scoped PostgreSQL request deduplication.
package idempotency

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/m8-team/platform-sdk/fault"
	"time"
)

type Key struct{ Tenant, Operation, Value string }
type Request struct {
	Key       Key
	Hash      []byte
	Retention time.Duration
}

// Execute must share tx with every business write and outbox insertion in fn.
// PROCESSING is never committed independently. A crashed transaction rolls back;
// concurrent requests wait on the unique constraint, bounded by ctx's deadline.
// Result serialization and schema versioning belong to the caller.
func Execute(ctx context.Context, tx *sql.Tx, req Request, fn func(context.Context) ([]byte, error)) ([]byte, bool, error) {
	if tx == nil || fn == nil || req.Key.Tenant == "" || req.Key.Operation == "" || req.Key.Value == "" || len(req.Hash) != 32 || req.Retention <= 0 {
		return nil, false, errors.New("invalid idempotency request")
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO platform.idempotency (tenant,operation,key,request_hash,status,expires_at) VALUES ($1,$2,$3,$4,'PROCESSING',clock_timestamp()+$5::bigint*interval '1 millisecond') ON CONFLICT DO NOTHING`, req.Key.Tenant, req.Key.Operation, req.Key.Value, req.Hash, req.Retention.Milliseconds())
	if err != nil {
		return nil, false, fmt.Errorf("claim idempotency: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return nil, false, err
	}
	if n == 0 {
		var hash, payload []byte
		var status string
		var expired bool
		err = tx.QueryRowContext(ctx, `SELECT request_hash,status,result,expires_at<=clock_timestamp() FROM platform.idempotency WHERE tenant=$1 AND operation=$2 AND key=$3`, req.Key.Tenant, req.Key.Operation, req.Key.Value).Scan(&hash, &status, &payload, &expired)
		if err != nil {
			return nil, false, fmt.Errorf("read idempotency: %w", err)
		}
		if !bytes.Equal(hash, req.Hash) {
			return nil, false, fault.New(fault.Conflict, "idempotency key used with a different request", nil)
		}
		if expired {
			return nil, false, fault.New(fault.FailedPrecondition, "idempotency key expired; use a new key", nil)
		}
		if status != "SUCCEEDED" {
			return nil, false, fault.New(fault.Conflict, "request is not complete", nil)
		}
		return payload, true, nil
	}
	payload, err := fn(ctx)
	if err != nil {
		return nil, false, err
	}
	_, err = tx.ExecContext(ctx, `UPDATE platform.idempotency SET status='SUCCEEDED',result=$4,updated_at=clock_timestamp() WHERE tenant=$1 AND operation=$2 AND key=$3`, req.Key.Tenant, req.Key.Operation, req.Key.Value, payload)
	if err != nil {
		return nil, false, fmt.Errorf("complete idempotency: %w", err)
	}
	return payload, false, nil
}

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/m8-team/platform-sdk/fault"
	"github.com/m8-team/platform-sdk/idempotency"
	"github.com/m8-team/platform-sdk/metadata"
	"github.com/m8-team/platform-sdk/outbox"
	"github.com/m8-team/platform-sdk/transaction"
	"github.com/m8-team/product-service/internal/product/domain"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"time"
)

type Repository struct {
	db          *sql.DB
	tracer      trace.Tracer
	propagation propagation.TextMapPropagator
}

func New(db *sql.DB, tracer trace.Tracer, p propagation.TextMapPropagator) *Repository {
	return &Repository{db, tracer, p}
}
func (r *Repository) Get(ctx context.Context, tenant domain.TenantID, id domain.ID) (domain.Product, error) {
	ctx, span := r.tracer.Start(ctx, "product.get")
	defer span.End()
	var product domain.Product
	err := r.db.QueryRowContext(ctx, `SELECT id,tenant,name FROM product.products WHERE tenant=$1 AND id=$2`, tenant, id).Scan(&product.ID, &product.Tenant, &product.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return product, fault.New(fault.NotFound, "product not found", err)
	}
	if err != nil {
		return product, fmt.Errorf("get product: %w", err)
	}
	return product, nil
}
func (r *Repository) Create(ctx context.Context, product domain.Product, key string, hash [32]byte) (domain.Product, error) {
	ctx, span := r.tracer.Start(ctx, "product.create.transaction")
	defer span.End()
	var encoded []byte
	err := transaction.Run(ctx, r.db, func(tx *sql.Tx) error {
		var err error
		encoded, _, err = idempotency.Execute(ctx, tx, idempotency.Request{Key: idempotency.Key{Tenant: string(product.Tenant), Operation: "product.create.v1", Value: key}, Hash: hash[:], Retention: 24 * time.Hour}, func(ctx context.Context) ([]byte, error) {
			if _, err := tx.ExecContext(ctx, `INSERT INTO product.products (id,tenant,name) VALUES ($1,$2,$3)`, product.ID, product.Tenant, product.Name); err != nil {
				return nil, fmt.Errorf("insert product: %w", err)
			}
			payload, err := json.Marshal(product)
			if err != nil {
				return nil, err
			}
			headers := propagation.MapCarrier{"tenant_id": string(product.Tenant), "schema_version": "1", "version": "1", "occurred_at": time.Now().UTC().Format(time.RFC3339Nano), "correlation_id": metadata.RequestID(ctx), "causation_id": metadata.RequestID(ctx)}
			r.propagation.Inject(ctx, headers)
			err = outbox.Insert(ctx, tx, outbox.Event{ID: string(product.ID) + ":created", Topic: "product.events", AggregateType: "product", AggregateID: string(product.ID), Type: "company.product.created.v1", Payload: payload, Headers: headers})
			return payload, err
		})
		return err
	})
	if err != nil {
		return domain.Product{}, err
	}
	var result domain.Product
	if err = json.Unmarshal(encoded, &result); err != nil {
		return result, fmt.Errorf("decode stored product: %w", err)
	}
	return result, nil
}

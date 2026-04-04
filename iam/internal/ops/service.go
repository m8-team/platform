package ops

import (
	"context"

	opsv1 "github.com/m8platform/platform/iam/gen/proto/saas/iam/ops/v1"
	"github.com/m8platform/platform/iam/internal/core"
	"github.com/m8platform/platform/iam/internal/storage/ydb"
)

type Service struct {
	opsv1.UnimplementedOperationsServiceServer

	store core.DocumentStore
}

func NewService(store core.DocumentStore) *Service {
	return &Service{store: store}
}

func (s *Service) ListOperations(ctx context.Context, req *opsv1.ListOperationsRequest) (*opsv1.ListOperationsResponse, error) {
	operations, next, err := core.ListProto(ctx, s.store, ydb.TableOperations, req.GetTenantId(), int(req.GetPageSize()), req.GetPageToken(), func() *opsv1.Operation {
		return &opsv1.Operation{}
	})
	if err != nil {
		return nil, err
	}
	return &opsv1.ListOperationsResponse{Operations: operations, NextPageToken: next}, nil
}

func (s *Service) GetOperation(ctx context.Context, req *opsv1.GetOperationRequest) (*opsv1.Operation, error) {
	operation := &opsv1.Operation{}
	if err := core.LoadProto(ctx, s.store, ydb.TableOperations, req.GetOperationId(), operation); err != nil {
		return nil, err
	}
	return operation, nil
}

package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/m8platform/platform/internal/access/app/command"
	"github.com/m8platform/platform/internal/access/app/ports"
	"github.com/m8platform/platform/internal/access/app/query"
	"github.com/m8platform/platform/internal/access/domain"
)

var ErrPermissionEngineRequired = errors.New("permission engine is required")

type CheckPermissionHandler struct {
	engine ports.PermissionEngine
}

func NewCheckPermissionHandler(engine ports.PermissionEngine) (*CheckPermissionHandler, error) {
	if engine == nil {
		return nil, ErrPermissionEngineRequired
	}

	return &CheckPermissionHandler{engine: engine}, nil
}

func (h *CheckPermissionHandler) Handle(
	ctx context.Context,
	cmd command.CheckPermissionCommand,
) (*query.CheckPermissionResult, error) {
	request, err := cmd.ToDomain()
	if err != nil {
		return nil, err
	}

	decision, err := h.engine.CheckPermission(ctx, request)
	if err != nil {
		failure, ok := classifyEngineFailure(ctx, err)
		if !ok {
			return nil, fmt.Errorf("check permission: %w", err)
		}

		decision, err = domain.NewEngineFailureDecision(request, failure, err.Error())
		if err != nil {
			return nil, err
		}

		result := query.NewCheckPermissionResult(decision)
		return &result, nil
	}

	if err := decision.EnsureModelRevision(request.ModelRevision()); err != nil {
		return nil, err
	}

	result := query.NewCheckPermissionResult(decision)
	return &result, nil
}

func classifyEngineFailure(ctx context.Context, err error) (domain.EngineFailureKind, bool) {
	if errors.Is(err, ports.ErrPermissionEngineTimeout) ||
		errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return domain.EngineFailureTimeout, true
	}

	if errors.Is(err, ports.ErrPermissionEngineUnavailable) {
		return domain.EngineFailureUnavailable, true
	}

	return "", false
}

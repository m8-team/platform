package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/m8-platform/m8/internal/platform/idempotency"
	"github.com/m8-platform/m8/internal/platform/operation"
	"github.com/m8-platform/m8/services/m8-authentication/internal/domain"
)

var (
	ErrClientNotFound   = errors.New("client not found")
	ErrSubjectNotFound  = errors.New("subject not found")
	ErrDependencyFailed = errors.New("dependency unavailable")
)

type ClientRepository interface {
	GetClient(context.Context, string) (domain.Client, bool, error)
}

type IdentityGateway interface {
	ResolveSubject(context.Context, string) (string, bool, error)
}

type RiskGateway interface {
	Evaluate(context.Context, string, string, string) (domain.RiskDecision, error)
}

type TransactionRepository interface {
	Save(context.Context, domain.Transaction) error
	Get(context.Context, string) (domain.Transaction, bool, error)
}

type OperationRepository interface {
	Save(context.Context, operation.Operation) error
	Get(context.Context, string) (operation.Operation, bool, error)
}

type OutboxRepository interface {
	Add(context.Context, domain.Event) error
}

type UnitOfWork interface {
	Do(context.Context, func(context.Context) error) error
}

type IDGenerator interface {
	NewID(prefix string) string
}

type Clock interface {
	Now() time.Time
}

type StartReauthenticationCommand struct {
	ClientID       string
	SubjectHint    string
	RequestedAAL   string
	Reason         string
	IdempotencyKey string
	CorrelationID  string
}

type StartReauthenticationResult struct {
	AuthenticationID string
	OperationID      string
	Reused           bool
}

type Service struct {
	Clients      ClientRepository
	Identity     IdentityGateway
	Risk         RiskGateway
	Transactions TransactionRepository
	Operations   OperationRepository
	Outbox       OutboxRepository
	Idempotency  idempotency.Store
	UOW          UnitOfWork
	IDs          IDGenerator
	Clock        Clock
	TTL          time.Duration
}

func (s Service) StartReauthentication(
	ctx context.Context,
	cmd StartReauthenticationCommand,
) (StartReauthenticationResult, error) {
	requestHash := hash(cmd.ClientID + "|" + cmd.SubjectHint + "|" + cmd.RequestedAAL + "|" + cmd.Reason)
	if rec, ok := s.Idempotency.Get("AUTH-FR-017", cmd.IdempotencyKey); ok {
		if rec.RequestHash != requestHash {
			return StartReauthenticationResult{}, idempotency.ErrConflict
		}
		op, found, err := s.Operations.Get(ctx, rec.ResultID)
		if err != nil || !found {
			return StartReauthenticationResult{}, ErrDependencyFailed
		}
		return StartReauthenticationResult{
			AuthenticationID: op.ResultID,
			OperationID:      op.ID,
			Reused:           true,
		}, nil
	}

	client, found, err := s.Clients.GetClient(ctx, cmd.ClientID)
	if err != nil {
		return StartReauthenticationResult{}, ErrDependencyFailed
	}
	if !found {
		return StartReauthenticationResult{}, ErrClientNotFound
	}
	if !client.Enabled || !client.CIBA {
		return StartReauthenticationResult{}, domain.ErrClientDisabled
	}

	subjectID, found, err := s.Identity.ResolveSubject(ctx, cmd.SubjectHint)
	if err != nil {
		return StartReauthenticationResult{}, ErrDependencyFailed
	}
	if !found {
		return StartReauthenticationResult{}, ErrSubjectNotFound
	}

	decision, err := s.Risk.Evaluate(ctx, client.ID, subjectID, cmd.RequestedAAL)
	if err != nil {
		return StartReauthenticationResult{}, ErrDependencyFailed
	}
	if !decision.Allowed {
		return StartReauthenticationResult{}, domain.ErrRiskDenied
	}

	now := s.Clock.Now()
	authenticationID := s.IDs.NewID("authn")
	operationID := s.IDs.NewID("operations")
	tx := domain.NewTransaction(authenticationID, client.ID, subjectID, decision.RequiredAAL, now, s.TTL)
	op := operation.Operation{
		ID:              operationID,
		Type:            "START_REAUTHENTICATION",
		State:           operation.Running,
		ProgressPercent: 10,
		Stage:           "CHALLENGE_PENDING",
		ResultID:        authenticationID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	event := domain.Event{
		ID:                s.IDs.NewID("events"),
		Type:              "m8.authentication.authentication_started.v1",
		AggregateID:       authenticationID,
		AggregateRevision: tx.Revision,
		CorrelationID:     cmd.CorrelationID,
		OccurredAt:        now,
	}

	err = s.UOW.Do(ctx, func(txCtx context.Context) error {
		if err := s.Transactions.Save(txCtx, tx); err != nil {
			return err
		}
		if err := s.Operations.Save(txCtx, op); err != nil {
			return err
		}
		if err := s.Outbox.Add(txCtx, event); err != nil {
			return err
		}
		return s.Idempotency.Put("AUTH-FR-017", cmd.IdempotencyKey, idempotency.Record{
			RequestHash: requestHash,
			ResultID:    operationID,
		})
	})
	if err != nil {
		return StartReauthenticationResult{}, err
	}

	return StartReauthenticationResult{
		AuthenticationID: authenticationID,
		OperationID:      operationID,
	}, nil
}

func hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

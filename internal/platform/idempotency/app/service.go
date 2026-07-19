package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/platform/idempotency/domain"
)

// Store is the application port implemented by persistence adapters:
// Postgres, YDB, Redis-backed cache, in-memory tests, etc.
type Store interface {
	Begin(
		ctx context.Context,
		scope domain.Scope,
		key domain.Key,
		fingerprint domain.Fingerprint,
		options domain.BeginOptions,
	) (*domain.BeginResult, error)

	Commit(
		ctx context.Context,
		scope domain.Scope,
		key domain.Key,
		leaseToken string,
		result domain.Result,
	) (*domain.Record, error)

	Abort(
		ctx context.Context,
		scope domain.Scope,
		key domain.Key,
		leaseToken string,
		err error,
		retryable bool,
	) (*domain.Record, error)

	Touch(
		ctx context.Context,
		scope domain.Scope,
		key domain.Key,
		leaseToken string,
		extendBy time.Duration,
	) error

	Get(
		ctx context.Context,
		scope domain.Scope,
		key domain.Key,
	) (*domain.Record, error)
}

// Clock is injected to make lease and TTL behavior deterministic in tests.
type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

// LeaseTokenGenerator creates opaque tokens used to prove execution ownership.
type LeaseTokenGenerator interface {
	Generate() (string, error)
}

type RandomLeaseTokenGenerator struct {
	Bytes int
}

func (g RandomLeaseTokenGenerator) Generate() (string, error) {
	n := g.Bytes
	if n <= 0 {
		n = 32
	}

	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate lease token: %w", err)
	}

	return hex.EncodeToString(buf), nil
}

type Config struct {
	Owner               string
	DefaultTTL          time.Duration
	DefaultLockTTL      time.Duration
	DefaultReplayPolicy domain.ReplayPolicy
}

func (c Config) withDefaults() Config {
	if c.DefaultTTL <= 0 {
		c.DefaultTTL = 24 * time.Hour
	}

	if c.DefaultLockTTL <= 0 {
		c.DefaultLockTTL = 30 * time.Second
	}

	if c.DefaultReplayPolicy == "" {
		c.DefaultReplayPolicy = domain.ReplayPolicySuccessAndFinalErrors
	}

	return c
}

type Option func(*Service)

func WithClock(clock Clock) Option {
	return func(s *Service) {
		if clock != nil {
			s.clock = clock
		}
	}
}

func WithLeaseTokenGenerator(generator LeaseTokenGenerator) Option {
	return func(s *Service) {
		if generator != nil {
			s.leaseTokens = generator
		}
	}
}

// Service contains idempotency application logic.
// It is transport-neutral and has no dependency on HTTP, gRPC or SQL.
type Service struct {
	store       Store
	clock       Clock
	leaseTokens LeaseTokenGenerator
	config      Config
}

func NewService(store Store, config Config, options ...Option) (*Service, error) {
	if store == nil {
		return nil, errors.New("idempotency store is required")
	}

	config = config.withDefaults()

	if !config.DefaultReplayPolicy.IsValid() {
		return nil, fmt.Errorf("%w: %q", domain.ErrInvalidReplayPolicy, config.DefaultReplayPolicy)
	}

	service := &Service{
		store:       store,
		clock:       SystemClock{},
		leaseTokens: RandomLeaseTokenGenerator{},
		config:      config,
	}

	for _, option := range options {
		option(service)
	}

	return service, nil
}

type BeginCommand struct {
	Scope       domain.Scope
	Key         domain.Key
	Fingerprint domain.Fingerprint

	TTL          time.Duration
	LockTTL      time.Duration
	Owner        string
	ReplayPolicy domain.ReplayPolicy
	Labels       map[string]string
}

func (s *Service) Begin(ctx context.Context, command BeginCommand) (*domain.BeginResult, error) {
	if err := command.Scope.Validate(); err != nil {
		return nil, err
	}

	if err := command.Key.Validate(); err != nil {
		return nil, err
	}

	if err := command.Fingerprint.Validate(); err != nil {
		return nil, err
	}

	options := domain.BeginOptions{
		TTL:          firstDuration(command.TTL, s.config.DefaultTTL),
		LockTTL:      firstDuration(command.LockTTL, s.config.DefaultLockTTL),
		Owner:        firstString(command.Owner, s.config.Owner),
		ReplayPolicy: firstReplayPolicy(command.ReplayPolicy, s.config.DefaultReplayPolicy),
		Labels:       cloneLabels(command.Labels),
	}

	if options.Owner == "" {
		options.Owner = "unknown"
	}

	if !options.ReplayPolicy.IsValid() {
		return nil, fmt.Errorf("%w: %q", domain.ErrInvalidReplayPolicy, options.ReplayPolicy)
	}

	return s.store.Begin(
		ctx,
		command.Scope,
		command.Key,
		command.Fingerprint,
		options,
	)
}

type CommitCommand struct {
	Scope      domain.Scope
	Key        domain.Key
	LeaseToken string
	Result     domain.Result
}

func (s *Service) Commit(ctx context.Context, command CommitCommand) (*domain.Record, error) {
	if err := command.Scope.Validate(); err != nil {
		return nil, err
	}

	if err := command.Key.Validate(); err != nil {
		return nil, err
	}

	if command.LeaseToken == "" {
		return nil, domain.ErrLeaseRequired
	}

	if err := command.Result.Validate(); err != nil {
		return nil, err
	}

	return s.store.Commit(
		ctx,
		command.Scope,
		command.Key,
		command.LeaseToken,
		command.Result,
	)
}

type AbortCommand struct {
	Scope      domain.Scope
	Key        domain.Key
	LeaseToken string
	Error      error
	Retryable  bool
}

func (s *Service) Abort(ctx context.Context, command AbortCommand) (*domain.Record, error) {
	if err := command.Scope.Validate(); err != nil {
		return nil, err
	}

	if err := command.Key.Validate(); err != nil {
		return nil, err
	}

	if command.LeaseToken == "" {
		return nil, domain.ErrLeaseRequired
	}

	if command.Error == nil {
		command.Error = errors.New("idempotency execution aborted")
	}

	return s.store.Abort(
		ctx,
		command.Scope,
		command.Key,
		command.LeaseToken,
		command.Error,
		command.Retryable,
	)
}

type TouchCommand struct {
	Scope      domain.Scope
	Key        domain.Key
	LeaseToken string
	ExtendBy   time.Duration
}

func (s *Service) Touch(ctx context.Context, command TouchCommand) error {
	if err := command.Scope.Validate(); err != nil {
		return err
	}

	if err := command.Key.Validate(); err != nil {
		return err
	}

	if command.LeaseToken == "" {
		return domain.ErrLeaseRequired
	}

	extendBy := firstDuration(command.ExtendBy, s.config.DefaultLockTTL)

	return s.store.Touch(
		ctx,
		command.Scope,
		command.Key,
		command.LeaseToken,
		extendBy,
	)
}

type GetCommand struct {
	Scope domain.Scope
	Key   domain.Key
}

func (s *Service) Get(ctx context.Context, command GetCommand) (*domain.Record, error) {
	if err := command.Scope.Validate(); err != nil {
		return nil, err
	}

	if err := command.Key.Validate(); err != nil {
		return nil, err
	}

	return s.store.Get(ctx, command.Scope, command.Key)
}

type ExecuteCommand struct {
	Begin BeginCommand
}

// Handler is called only when Begin returns EXECUTE or RECOVER.
// The handler receives the lease token that must be used by Commit/Abort.
type Handler func(ctx context.Context, leaseToken string, record *domain.Record) (domain.Result, error)

type ExecuteResult struct {
	Decision domain.Decision
	Record   *domain.Record
	Result   *domain.Result
}

// Execute is a convenience helper for simple synchronous commands.
// For LRO commands, prefer explicit Begin -> create operation -> Commit(OperationResult).
func (s *Service) Execute(
	ctx context.Context,
	command ExecuteCommand,
	handler Handler,
) (*ExecuteResult, error) {
	if handler == nil {
		return nil, errors.New("idempotency handler is required")
	}

	begin, err := s.Begin(ctx, command.Begin)
	if err != nil {
		return nil, err
	}

	switch begin.Decision {
	case domain.DecisionReplay:
		return &ExecuteResult{
			Decision: begin.Decision,
			Record:   begin.Record,
			Result:   begin.Record.Result,
		}, nil

	case domain.DecisionInProgress:
		return &ExecuteResult{
				Decision: begin.Decision,
				Record:   begin.Record,
			}, &domain.InProgressError{
				Scope:      command.Begin.Scope,
				Key:        command.Begin.Key,
				RetryAfter: begin.RetryAfter.String(),
			}

	case domain.DecisionConflict:
		return &ExecuteResult{
			Decision: begin.Decision,
			Record:   begin.Record,
		}, domain.ErrConflict

	case domain.DecisionExpired:
		return &ExecuteResult{
			Decision: begin.Decision,
			Record:   begin.Record,
		}, domain.ErrExpired

	case domain.DecisionRecover, domain.DecisionExecute:
		result, handlerErr := handler(ctx, begin.LeaseToken, begin.Record)
		if handlerErr != nil {
			_, abortErr := s.Abort(ctx, AbortCommand{
				Scope:      command.Begin.Scope,
				Key:        command.Begin.Key,
				LeaseToken: begin.LeaseToken,
				Error:      handlerErr,
				Retryable:  true,
			})
			if abortErr != nil {
				return nil, errors.Join(handlerErr, abortErr)
			}

			return nil, handlerErr
		}

		record, commitErr := s.Commit(ctx, CommitCommand{
			Scope:      command.Begin.Scope,
			Key:        command.Begin.Key,
			LeaseToken: begin.LeaseToken,
			Result:     result,
		})
		if commitErr != nil {
			return nil, commitErr
		}

		return &ExecuteResult{
			Decision: begin.Decision,
			Record:   record,
			Result:   &result,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported idempotency decision: %s", begin.Decision)
	}
}

// NewLeaseToken is exposed for stores that want the application layer
// to generate tokens before atomic claim. Stores may also generate tokens
// internally if they own Begin completely.
func (s *Service) NewLeaseToken() (string, error) {
	return s.leaseTokens.Generate()
}

func firstDuration(value time.Duration, fallback time.Duration) time.Duration {
	if value > 0 {
		return value
	}

	return fallback
}

func firstString(value string, fallback string) string {
	if value != "" {
		return value
	}

	return fallback
}

func firstReplayPolicy(value domain.ReplayPolicy, fallback domain.ReplayPolicy) domain.ReplayPolicy {
	if value != "" {
		return value
	}

	return fallback
}

func cloneLabels(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return nil
	}

	out := make(map[string]string, len(labels))
	for key, value := range labels {
		out[key] = value
	}

	return out
}

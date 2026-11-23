// Package catalog provides an authenticated, deny-by-default publication boundary
// for module-owned authorization definitions. Persistence is supplied by an adapter.
package catalog

import (
	"context"
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/shared"
)
var (
	ErrCatalogUnavailable = errors.New("authorization catalog unavailable")
	ErrUnauthorizedPublisher = errors.New("module namespace publication denied")
	ErrInvalidPublication = errors.New("invalid module catalog publication")
)
// PrincipalResolver derives the actor from verified workload identity (mTLS/OIDC).
// Caller-supplied module headers, strings and manifest metadata are not credentials.
type PrincipalResolver interface {
	AuthenticatedPublisher(ctx context.Context) (shared.Subject,error)
}
// NamespaceAuthorizer verifies that actor is allowed to publish the module's
// namespace. Manifest.Publisher() matching the actor is necessary but insufficient.
type NamespaceAuthorizer interface {
	CanPublishModule(ctx context.Context,actor shared.Subject,moduleID string)(bool,error)
}
// ManifestRepository MUST enforce compare-and-swap, release-version immutability,
// retry idempotency and durable publication in an authoritative transaction.
// Neither the caller nor this interface writes another module's resource records.
type ManifestRepository interface {
	Publish(ctx context.Context,m module.Manifest,expectedRevision int64,idempotencyKey string)error
}
type Service struct {
	resolver PrincipalResolver
	authorizer NamespaceAuthorizer
	repository ManifestRepository
}
func New(resolver PrincipalResolver,authorizer NamespaceAuthorizer,repository ManifestRepository)(*Service,error){
	if resolver==nil||authorizer==nil||repository==nil{return nil,ErrCatalogUnavailable}
	return &Service{resolver:resolver,authorizer:authorizer,repository:repository},nil
}
func (s *Service) Publish(ctx context.Context,m module.Manifest,expectedRevision int64,idempotencyKey string)error{
	if s==nil||s.resolver==nil||s.authorizer==nil||s.repository==nil{return ErrCatalogUnavailable}
	if err:=ctx.Err();err!=nil{return err}
	if m.Validate()!=nil||expectedRevision<0||strings.TrimSpace(idempotencyKey)!=idempotencyKey||
	idempotencyKey==""||len(idempotencyKey)>256{return ErrInvalidPublication}
	actor,err:=s.resolver.AuthenticatedPublisher(ctx)
	if err!=nil{return ErrUnauthorizedPublisher}
	if actor.Validate()!=nil||actor.Type()!="service_account"||actor!=m.Publisher(){return ErrUnauthorizedPublisher}
	allowed,err:=s.authorizer.CanPublishModule(ctx,actor,m.ModuleID())
	if err!=nil||!allowed{return ErrUnauthorizedPublisher}
	return s.repository.Publish(ctx,m,expectedRevision,idempotencyKey)
}

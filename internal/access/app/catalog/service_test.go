package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

type resolverStub struct {
	principal shared.Subject
	err       error
}

func (r resolverStub) AuthenticatedPublisher(context.Context) (shared.Subject, error) {
	return r.principal, r.err
}

type authorizerStub struct {
	allowed bool
	err     error
}

func (a authorizerStub) CanPublishModule(context.Context, shared.Subject, string) (bool, error) {
	return a.allowed, a.err
}

type repositoryStub struct{ saved int }

func (r *repositoryStub) Publish(context.Context, module.Manifest, int64, string) error {
	r.saved++
	return nil
}

func fixture(t *testing.T) (module.Manifest, shared.Subject) {
	t.Helper()
	publisher, _ := shared.NewSubject("service_account", "finance-service")
	resourceType, _ := resource.NewTypeDefinition("finance", "finance.invoice", []authorization.ScopeKind{authorization.ScopeInstance})
	action, _ := permission.NewPermission("finance.invoice.view")
	definition, _ := permission.NewDefinition(action, "finance.invoice", "use")
	m, err := module.NewManifest(module.ManifestInput{
		ModuleID: "finance", Version: "1.0.0", Publisher: publisher,
		Resources: []resource.TypeDefinition{resourceType}, Permissions: []permission.Definition{definition},
	})
	if err != nil {
		t.Fatal(err)
	}
	return m, publisher
}
func TestPublisherMustOwnNamespace(t *testing.T) {
	manifest, identity := fixture(t)
	repository := &repositoryStub{}
	untrusted, _ := shared.NewSubject("service_account", "attacker")
	service, err := New(resolverStub{principal: untrusted}, authorizerStub{allowed: true}, repository)
	if err != nil {
		t.Fatal(err)
	}
	if err = service.Publish(context.Background(), manifest, 0, "request-1"); !errors.Is(err, ErrUnauthorizedPublisher) {
		t.Fatal(err)
	}
	service, _ = New(resolverStub{principal: identity}, authorizerStub{allowed: false}, repository)
	if err = service.Publish(context.Background(), manifest, 0, "request-1"); !errors.Is(err, ErrUnauthorizedPublisher) {
		t.Fatal(err)
	}
	if repository.saved != 0 {
		t.Fatal("unauthorized publisher mutated catalog")
	}
	service, _ = New(resolverStub{principal: identity}, authorizerStub{allowed: true}, repository)
	if err = service.Publish(context.Background(), manifest, 0, "request-1"); err != nil {
		t.Fatal(err)
	}
	if repository.saved != 1 {
		t.Fatal("authorized manifest was not saved")
	}
	if err = service.Publish(context.Background(), manifest, 0, ""); !errors.Is(err, ErrInvalidPublication) {
		t.Fatal(err)
	}
}

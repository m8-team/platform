package resource

import (
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidResourceReference = errors.New("invalid protected resource reference")

// Reference is an owner-bound pointer. The object is still written exclusively
// by its authoritative module; catalog registration is not object creation.
type Reference struct {
	tenantID      shared.TenantID
	kind          string
	id            string
	ownerModule   string
	ownerRevision int64
}

func NewReference(tenant shared.TenantID, kind, id, owner string, ownerRevision int64) (Reference, error) {
	r := Reference{tenantID: tenant, kind: strings.TrimSpace(kind), id: strings.TrimSpace(id),
		ownerModule: strings.TrimSpace(owner), ownerRevision: ownerRevision}
	return r, r.Validate()
}
func (r Reference) Validate() error {
	if r.tenantID.Validate() != nil || !shared.ValidQualifiedName(r.kind) || !shared.ValidModuleName(r.ownerModule) ||
		!strings.HasPrefix(r.kind, r.ownerModule+".") || r.id == "" || len(r.id) > 256 || r.ownerRevision < 1 {
		return ErrInvalidResourceReference
	}
	return nil
}
func (r Reference) TenantID() shared.TenantID { return r.tenantID }
func (r Reference) Kind() string              { return r.kind }
func (r Reference) ID() string                { return r.id }
func (r Reference) OwnerModule() string       { return r.ownerModule }
func (r Reference) OwnerRevision() int64      { return r.ownerRevision }

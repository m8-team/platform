// Package ydbcatalog persists Access Catalog releases, command receipts and
// read-side projections in a single serializable YDB transaction.
package ydbcatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/module"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidManifest = errors.New("invalid persisted authorization manifest")

const maxManifestSize = 4 << 20

// A purpose-built storage DTO avoids serializing private domain fields
// (which encoding/json would otherwise silently omit).
type manifestDTO struct {
	ModuleID      string          `json:"module_id"`
	Version       string          `json:"version"`
	PublisherType string          `json:"publisher_type"`
	PublisherID   string          `json:"publisher_id"`
	Resources     []resourceDTO   `json:"resources"`
	Permissions   []permissionDTO `json:"permissions"`
	Roles         []roleDTO       `json:"roles"`
}
type resourceDTO struct {
	Name   string                    `json:"name"`
	Scopes []authorization.ScopeKind `json:"scopes"`
}
type permissionDTO struct {
	Name            string `json:"name"`
	ResourceType    string `json:"resource_type"`
	GraphPermission string `json:"graph_permission"`
}
type roleDTO struct {
	Name        string                    `json:"name"`
	Version     string                    `json:"version"`
	Permissions []string                  `json:"permissions"`
	Scopes      []authorization.ScopeKind `json:"scopes"`
}

func marshalManifest(m module.Manifest) ([]byte, error) {
	if err := m.Validate(); err != nil {
		return nil, ErrInvalidManifest
	}
	d := manifestDTO{ModuleID: m.ModuleID(), Version: m.Version(),
		PublisherType: m.Publisher().Type(), PublisherID: m.Publisher().ID()}
	for _, r := range m.Resources() {
		d.Resources = append(d.Resources, resourceDTO{Name: r.Name(), Scopes: r.Scopes()})
	}
	for _, p := range m.Permissions() {
		d.Permissions = append(d.Permissions, permissionDTO{
			Name: p.Permission().Name(), ResourceType: p.ResourceType(), GraphPermission: p.GraphPermission()})
	}
	for _, r := range m.Roles() {
		entry := roleDTO{Name: r.Ref().Name, Version: r.Ref().Version, Scopes: r.AllowedScopes()}
		for _, p := range r.Permissions() {
			entry.Permissions = append(entry.Permissions, p.Name())
		}
		d.Roles = append(d.Roles, entry)
	}
	encoded, err := json.Marshal(d)
	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", ErrInvalidManifest, err)
	}
	if len(encoded) > maxManifestSize {
		return nil, ErrInvalidManifest
	}
	return encoded, nil
}
func unmarshalManifest(encoded []byte) (module.Manifest, error) {
	if len(encoded) == 0 || len(encoded) > maxManifestSize {
		return module.Manifest{}, ErrInvalidManifest
	}
	var d manifestDTO
	if err := json.Unmarshal(encoded, &d); err != nil {
		return module.Manifest{}, ErrInvalidManifest
	}
	actor, err := shared.NewSubject(d.PublisherType, d.PublisherID)
	if err != nil {
		return module.Manifest{}, ErrInvalidManifest
	}
	in := module.ManifestInput{ModuleID: d.ModuleID, Version: d.Version, Publisher: actor}
	for _, r := range d.Resources {
		typ, err := resource.NewTypeDefinition(d.ModuleID, r.Name, r.Scopes)
		if err != nil {
			return module.Manifest{}, ErrInvalidManifest
		}
		in.Resources = append(in.Resources, typ)
	}
	for _, p := range d.Permissions {
		action, err := permission.NewPermission(p.Name)
		if err != nil {
			return module.Manifest{}, ErrInvalidManifest
		}
		definition, err := permission.NewDefinition(action, p.ResourceType, p.GraphPermission)
		if err != nil {
			return module.Manifest{}, ErrInvalidManifest
		}
		in.Permissions = append(in.Permissions, definition)
	}
	for _, r := range d.Roles {
		var actions []permission.Permission
		for _, name := range r.Permissions {
			action, err := permission.NewPermission(name)
			if err != nil {
				return module.Manifest{}, ErrInvalidManifest
			}
			actions = append(actions, action)
		}
		template, err := role.NewTemplate(d.ModuleID, role.TemplateRef{Name: r.Name, Version: r.Version}, actions, r.Scopes)
		if err != nil {
			return module.Manifest{}, ErrInvalidManifest
		}
		in.Roles = append(in.Roles, template)
	}
	result, err := module.NewManifest(in)
	if err != nil {
		return module.Manifest{}, ErrInvalidManifest
	}
	return result, nil
}
func requestHash(payload []byte, expected int64) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("publish/catalog/v1:%d:%s", expected, payload)))
	return hex.EncodeToString(sum[:])
}

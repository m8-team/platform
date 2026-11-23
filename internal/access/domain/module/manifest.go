// Package module owns declarations of authorization capabilities by trusted M8 modules.
package module

import (
	"errors"
	"strings"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/resource"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidManifest = errors.New("invalid authorization module manifest")

// Manifest is a versioned, publish-once catalog snapshot. Source modules own
// their resources; Access owns only their authorization definitions.
type Manifest struct{
	moduleID string
	version string
	publisher shared.Subject
	resources []resource.TypeDefinition
	permissions []permission.Definition
	roles []role.Template
}
type ManifestInput struct{
	ModuleID string
	Version string
	Publisher shared.Subject
	Resources []resource.TypeDefinition
	Permissions []permission.Definition
	Roles []role.Template
}
func NewManifest(in ManifestInput)(Manifest,error){
	m:=Manifest{moduleID:strings.TrimSpace(in.ModuleID),version:strings.TrimSpace(in.Version),
	publisher:in.Publisher,resources:append([]resource.TypeDefinition(nil),in.Resources...),
	permissions:append([]permission.Definition(nil),in.Permissions...),roles:append([]role.Template(nil),in.Roles...)}
	return m,m.Validate()
}
func (m Manifest) Validate()error{
	if !shared.ValidModuleName(m.moduleID)||!shared.ValidReleaseVersion(m.version)||
	m.publisher.Validate()!=nil||m.publisher.Type()!="service_account"||
	len(m.resources)==0||len(m.resources)>256||len(m.permissions)==0||len(m.permissions)>1024||len(m.roles)>256{
		return ErrInvalidManifest
	}
	types:=map[string]resource.TypeDefinition{}
	for _,d:=range m.resources{
		if d.Validate()!=nil||d.ModuleID()!=m.moduleID{return ErrInvalidManifest}
		if _,ok:=types[d.Name()];ok{return ErrInvalidManifest}
		types[d.Name()]=d
	}
	actions:=map[string]resource.TypeDefinition{}
	for _,d:=range m.permissions{
		if d.Validate()!=nil||!strings.HasPrefix(d.Permission().Name(),m.moduleID+"."){return ErrInvalidManifest}
		typ,exists:=types[d.ResourceType()];if !exists{return ErrInvalidManifest}
		// A public action must resolve to one resource type; a collision cannot be
		// used to widen the same role permission across unrelated resources.
		if _,duplicate:=actions[d.Permission().Name()];duplicate{return ErrInvalidManifest}
		actions[d.Permission().Name()]=typ
	}
	templates:=map[string]bool{}
	for _,t:=range m.roles{
		if t.Validate()!=nil||t.ModuleID()!=m.moduleID||templates[t.Ref().Key()]{return ErrInvalidManifest}
		templates[t.Ref().Key()]=true
		for _,action:=range t.Permissions(){
			typ,found:=actions[action.Name()]
			if !found{return ErrInvalidManifest}
			// A role template cannot advertise a broader scope than any
			// protected resource type for which it includes an action.
			for _,scope:=range t.AllowedScopes(){if !typ.Supports(scope){return ErrInvalidManifest}}
		}
	}
	return nil
}
func (m Manifest) ModuleID()string{return m.moduleID}
func (m Manifest) Version()string{return m.version}
func (m Manifest) Publisher()shared.Subject{return m.publisher}
func (m Manifest) Resources()[]resource.TypeDefinition{return append([]resource.TypeDefinition(nil),m.resources...)}
func (m Manifest) Permissions()[]permission.Definition{return append([]permission.Definition(nil),m.permissions...)}
func (m Manifest) Roles()[]role.Template{return append([]role.Template(nil),m.roles...)}

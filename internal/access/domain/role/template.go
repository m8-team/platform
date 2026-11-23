package role

import (
	"errors"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/permission"
	"github.com/m8-team/platform/internal/access/domain/shared"
)
var ErrInvalidRoleTemplate = errors.New("invalid versioned role template")

// TemplateRef pins a module-published role by both identity and version.
type TemplateRef struct {
	Name string
	Version string
}
func (r TemplateRef) Validate() error {
	if !shared.ValidQualifiedName(r.Name)||!shared.ValidReleaseVersion(r.Version){return ErrInvalidRoleTemplate}
	return nil
}
func (r TemplateRef) Key() string{return r.Name+"@"+r.Version}

// Template is immutable catalog content; changing permissions publishes a NEW version.
// Tenant-specific grants live in RoleBinding/Assignment, not in a Template.
type Template struct {
	moduleID string
	ref TemplateRef
	permissions []permission.Permission
	allowedScopes []authorization.ScopeKind
}
func NewTemplate(moduleID string,ref TemplateRef,permissions []permission.Permission,scopes []authorization.ScopeKind)(Template,error) {
	t:=Template{moduleID:strings.TrimSpace(moduleID),ref:ref,permissions:append([]permission.Permission(nil),permissions...),
	allowedScopes:append([]authorization.ScopeKind(nil),scopes...)}
	return t,t.Validate()
}
func (t Template) Validate() error {
	if !shared.ValidModuleName(t.moduleID)||t.ref.Validate()!=nil||
	!strings.HasPrefix(t.ref.Name,t.moduleID+".")||permission.ValidateSet(t.permissions,true)!=nil||
	len(t.allowedScopes)==0||len(t.allowedScopes)>8{return ErrInvalidRoleTemplate}
	seen:=make(map[authorization.ScopeKind]bool,len(t.allowedScopes))
	for _,scope:=range t.allowedScopes{if !scope.Valid()||seen[scope]{return ErrInvalidRoleTemplate};seen[scope]=true}
	return nil
}
func (t Template) ModuleID()string{return t.moduleID}
func (t Template) Ref()TemplateRef{return t.ref}
func (t Template) Permissions()[]permission.Permission{return append([]permission.Permission(nil),t.permissions...)}
func (t Template) AllowedScopes()[]authorization.ScopeKind{return append([]authorization.ScopeKind(nil),t.allowedScopes...)}
func (t Template) SupportsScope(kind authorization.ScopeKind)bool {
	for _,s:=range t.allowedScopes{if s==kind{return true}}
	return false
}

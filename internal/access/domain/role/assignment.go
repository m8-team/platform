package role

import (
	"errors"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/shared"
)
var (
	ErrInvalidAssignment = errors.New("invalid access role/profile assignment")
	ErrAssignmentNotActive = errors.New("assignment is not active")
)
type AssignmentTargetKind string
const (
	AssignmentTemplate AssignmentTargetKind = "TEMPLATE"
	AssignmentBundle AssignmentTargetKind = "BUNDLE"
)
type AssignmentTarget struct {
	Kind AssignmentTargetKind
	Name string
	Version string
}
func (a AssignmentTarget) Validate()error{
	if !shared.ValidQualifiedName(a.Name)||!shared.ValidReleaseVersion(a.Version)||
	(a.Kind!=AssignmentTemplate&&a.Kind!=AssignmentBundle){return ErrInvalidAssignment}
	return nil
}
func (a AssignmentTarget) Key()string{return string(a.Kind)+":"+a.Name+"@"+a.Version}

type AssignmentState string
const (
	AssignmentActive AssignmentState = "ACTIVE"
	AssignmentRevoked AssignmentState = "REVOKED"
)
// Assignment pins one template/profile version to a specific subject and scope.
// Its independent identity prevents revoking another overlapping assignment.
type Assignment struct{
	id string
	tenantID shared.TenantID
	subject shared.Subject
	target AssignmentTarget
	scope authorization.Scope
	policyIDs []string
	revision shared.Revision
	state AssignmentState
	revokedAt time.Time
}
type AssignmentInput struct{
	ID string
	TenantID shared.TenantID
	Subject shared.Subject
	Target AssignmentTarget
	Scope authorization.Scope
	PolicyIDs []string
}
func NewAssignment(in AssignmentInput)(Assignment,error){
	a:=Assignment{id:strings.TrimSpace(in.ID),tenantID:in.TenantID,subject:in.Subject,
	target:in.Target,scope:in.Scope,policyIDs:append([]string(nil),in.PolicyIDs...),revision:1,state:AssignmentActive}
	return a,a.Validate()
}
func (a Assignment) Validate()error{
	if !shared.ValidUUID(a.id)||a.tenantID.Validate()!=nil||a.subject.Validate()!=nil||a.target.Validate()!=nil||
	a.scope.Validate()!=nil||a.scope.TenantID()!=a.tenantID||a.revision.Validate()!=nil{return ErrInvalidAssignment}
	if a.state!=AssignmentActive&&a.state!=AssignmentRevoked{return ErrInvalidAssignment}
	if (a.state==AssignmentRevoked)!=(!a.revokedAt.IsZero()){return ErrInvalidAssignment}
	seen:=make(map[string]bool,len(a.policyIDs))
	if len(a.policyIDs)>64{return ErrInvalidAssignment}
	for _,id:=range a.policyIDs{if !shared.ValidUUID(id)||seen[id]{return ErrInvalidAssignment};seen[id]=true}
	return nil
}
func (a Assignment) Revoke(expected int64,at time.Time)(Assignment,error){
	if a.state!=AssignmentActive{return Assignment{},ErrAssignmentNotActive}
	if at.IsZero(){return Assignment{},ErrInvalidAssignment}
	next,err:=a.revision.Next(expected);if err!=nil{return Assignment{},err}
	updated:=a;updated.revision=next;updated.state=AssignmentRevoked;updated.revokedAt=at.UTC()
	return updated,updated.Validate()
}
func (a Assignment) ID()string{return a.id}
func (a Assignment) TenantID()shared.TenantID{return a.tenantID}
func (a Assignment) Subject()shared.Subject{return a.subject}
func (a Assignment) Target()AssignmentTarget{return a.target}
func (a Assignment) Scope()authorization.Scope{return a.scope}
func (a Assignment) PolicyIDs()[]string{return append([]string(nil),a.policyIDs...)}
func (a Assignment) Revision()int64{return a.revision.Int64()}
func (a Assignment) State()AssignmentState{return a.state}
func (a Assignment) RevokedAt()time.Time{return a.revokedAt}

// Package authorization contains engine-independent permission-check contracts.
package authorization

import (
 "errors"
 "strings"
 "github.com/m8-team/platform/internal/access/domain/permission"
 "github.com/m8-team/platform/internal/access/domain/shared"
)

var (
 ErrInvalidCheck = errors.New("invalid permission check")
 ErrInvalidFailMode = errors.New("invalid access fail mode")
 ErrFailOpenForCriticalCheck = errors.New("fail-open forbidden for critical checks")
 ErrFailOpenReferenceRequired = errors.New("fail-open requires an approved reference")
 ErrInvalidZedToken = errors.New("invalid SpiceDB ZedToken")
)
type FailMode string
const (
 FailDeny FailMode = "DENY"
 FailAllow FailMode = "ALLOW"
)
func (f FailMode) WithDefault() FailMode {if f=="" {return FailDeny};return f}
func (f FailMode) Valid() bool{return f==FailDeny||f==FailAllow}
type CheckPermissionInput struct {
 TenantID shared.TenantID
 MinZedToken string
 Subject shared.Subject
 Permission permission.Permission
 Resource shared.Resource
 ModelRevision shared.ModelRevision
 FailMode FailMode
 Critical bool
 FailOpenReference string
}
type CheckPermissionRequest struct {
 tenantID shared.TenantID
 minZedToken string
 subject shared.Subject
 permission permission.Permission
 resource shared.Resource
 modelRevision shared.ModelRevision
 failMode FailMode
 critical bool
 failOpenReference string
}
func NewCheckPermissionRequest(input CheckPermissionInput)(CheckPermissionRequest,error){
 c:=CheckPermissionRequest{
  tenantID:input.TenantID,minZedToken:strings.TrimSpace(input.MinZedToken),
  subject:input.Subject,permission:input.Permission,resource:input.Resource,
  modelRevision:input.ModelRevision,failMode:input.FailMode.WithDefault(),
  critical:input.Critical,failOpenReference:strings.TrimSpace(input.FailOpenReference),
 }
 return c,c.Validate()
}
func (c CheckPermissionRequest) Validate() error {
 if c.tenantID.Validate()!=nil||c.subject.Validate()!=nil||c.permission.Validate()!=nil||
  c.resource.Validate()!=nil||c.modelRevision.Validate()!=nil{return ErrInvalidCheck}
 if len(c.minZedToken)>4096{return ErrInvalidZedToken}
 if !c.failMode.Valid(){return ErrInvalidFailMode}
 if c.failMode==FailAllow {
  if c.critical{return ErrFailOpenForCriticalCheck}
  if c.failOpenReference==""{return ErrFailOpenReferenceRequired}
 }
 return nil
}
func (c CheckPermissionRequest) TenantID() shared.TenantID{return c.tenantID}
func (c CheckPermissionRequest) MinZedToken() string{return c.minZedToken}
func (c CheckPermissionRequest) Subject() shared.Subject{return c.subject}
func (c CheckPermissionRequest) Permission() permission.Permission{return c.permission}
func (c CheckPermissionRequest) Resource() shared.Resource{return c.resource}
func (c CheckPermissionRequest) ModelRevision() shared.ModelRevision{return c.modelRevision}
func (c CheckPermissionRequest) FailMode() FailMode{return c.failMode}
func (c CheckPermissionRequest) Critical() bool{return c.critical}
func (c CheckPermissionRequest) FailOpenReference() string{return c.failOpenReference}

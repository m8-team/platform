package permission

import (
 "errors"
 "strings"
 "github.com/m8-team/platform/internal/access/domain/shared"
)

var (
 ErrInvalidModel = errors.New("invalid authorization model")
 ErrModelPublished = errors.New("published authorization model is immutable")
 ErrDuplicateDefinition = errors.New("duplicate permission definition")
)

type ModelState string
const (
 ModelDraft ModelState = "DRAFT"
 ModelPublished ModelState = "PUBLISHED"
)

// AuthorizationModel is a tenant-scoped aggregate. Publication freezes a snapshot;
// further edits require a new draft, never an in-place mutation of a published model.
type AuthorizationModel struct {
 id string
 tenantID shared.TenantID
 state ModelState
 revision shared.Revision
 modelRevision shared.ModelRevision
 definitions []Definition
}

func NewAuthorizationModel(id string, tenant shared.TenantID, definitions []Definition) (AuthorizationModel,error) {
 m:=AuthorizationModel{id:strings.TrimSpace(id),tenantID:tenant,state:ModelDraft,revision:1,definitions:append([]Definition(nil),definitions...)}
 return m,m.Validate()
}
func (m AuthorizationModel) Validate() error {
 if !shared.ValidUUID(m.id) || m.tenantID.Validate()!=nil || m.revision.Validate()!=nil { return ErrInvalidModel }
 if m.state!=ModelDraft && m.state!=ModelPublished { return ErrInvalidModel }
 if m.state==ModelPublished && m.modelRevision.Validate()!=nil { return ErrInvalidModel }
 seen:=make(map[string]struct{},len(m.definitions))
 for _,d:=range m.definitions {
  if err:=d.Validate();err!=nil {return err}
  if _,ok:=seen[d.Key()];ok{return ErrDuplicateDefinition}
  seen[d.Key()]=struct{}{}
 }
 return nil
}
func (m AuthorizationModel) WithDefinitions(expected int64, definitions []Definition) (AuthorizationModel,error) {
 if m.state!=ModelDraft{return AuthorizationModel{},ErrModelPublished}
 next,err:=m.revision.Next(expected);if err!=nil{return AuthorizationModel{},err}
 updated:=m;updated.revision=next;updated.definitions=append([]Definition(nil),definitions...)
 if err:=updated.Validate();err!=nil{return AuthorizationModel{},err}
 return updated,nil
}
func (m AuthorizationModel) Publish(expected int64, revision shared.ModelRevision) (AuthorizationModel,error) {
 if m.state!=ModelDraft{return AuthorizationModel{},ErrModelPublished}
 if err:=revision.Validate();err!=nil{return AuthorizationModel{},err}
 if len(m.definitions)==0{return AuthorizationModel{},ErrInvalidModel}
 next,err:=m.revision.Next(expected);if err!=nil{return AuthorizationModel{},err}
 updated:=m;updated.revision=next;updated.state=ModelPublished;updated.modelRevision=revision
 return updated,updated.Validate()
}
func (m AuthorizationModel) ID() string{return m.id}
func (m AuthorizationModel) TenantID() shared.TenantID{return m.tenantID}
func (m AuthorizationModel) State() ModelState{return m.state}
func (m AuthorizationModel) Revision() int64{return m.revision.Int64()}
func (m AuthorizationModel) ModelRevision() shared.ModelRevision{return m.modelRevision}
func (m AuthorizationModel) Definitions() []Definition{return append([]Definition(nil),m.definitions...)}

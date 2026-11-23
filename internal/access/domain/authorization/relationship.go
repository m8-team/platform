package authorization

import (
 "errors"
 "regexp"
 "strings"
 "time"
 "github.com/m8-team/platform/internal/access/domain/shared"
)
var (
 ErrInvalidRelationship = errors.New("invalid authorization relationship")
 ErrInvalidRelationName = errors.New("invalid graph relation name")
)
var relationPattern=regexp.MustCompile("^[a-z][a-z0-9_]{0,127}$")
// Relationship is a derived, tenant-bound graph tuple, never an authoritative grant.
// Both the graph resource AND subject namespaces must be scoped to the tenant.
type Relationship struct {
 tenantID shared.TenantID
 resource shared.Resource
 relation string
 subject shared.Subject
 subjectRelation string
 conditionRef string
 expiresAt time.Time
}
type RelationshipInput struct {
 TenantID shared.TenantID
 Resource shared.Resource
 Relation string
 Subject shared.Subject
 SubjectRelation string
 ConditionRef string
 ExpiresAt time.Time
}
func NewRelationship(in RelationshipInput)(Relationship,error){
 r:=Relationship{tenantID:in.TenantID,resource:in.Resource,relation:strings.TrimSpace(in.Relation),
 subject:in.Subject,subjectRelation:strings.TrimSpace(in.SubjectRelation),
 conditionRef:strings.TrimSpace(in.ConditionRef)}
 if !in.ExpiresAt.IsZero(){r.expiresAt=in.ExpiresAt.UTC()}
 return r,r.Validate()
}
func (r Relationship) Validate()error{
 if r.tenantID.Validate()!=nil||r.resource.Validate()!=nil||r.subject.Validate()!=nil{return ErrInvalidRelationship}
 if !relationPattern.MatchString(r.relation){return ErrInvalidRelationName}
 if r.subjectRelation!=""&&!relationPattern.MatchString(r.subjectRelation){return ErrInvalidRelationName}
 if r.conditionRef!=""&&!relationPattern.MatchString(r.conditionRef){return ErrInvalidRelationship}
 return nil
}
type RelationshipKey struct {
 TenantID shared.TenantID
 ResourceType string
 ResourceID string
 Relation string
 SubjectType string
 SubjectID string
 SubjectRelation string
}
func (r Relationship) Key()RelationshipKey{
 return RelationshipKey{r.tenantID,r.resource.Type(),r.resource.ID(),r.relation,r.subject.Type(),r.subject.ID(),r.subjectRelation}
}
func (r Relationship) TenantID()shared.TenantID{return r.tenantID}
func (r Relationship) Resource()shared.Resource{return r.resource}
func (r Relationship) Relation()string{return r.relation}
func (r Relationship) Subject()shared.Subject{return r.subject}
func (r Relationship) SubjectRelation()string{return r.subjectRelation}
func (r Relationship) ConditionRef()string{return r.conditionRef}
func (r Relationship) ExpiresAt()time.Time{return r.expiresAt}

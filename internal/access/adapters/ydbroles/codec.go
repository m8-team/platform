package ydbroles

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidSnapshot = errors.New("invalid role management snapshot")

const maxRecordBytes = 1 << 20

type roleRefDTO struct {
	Name string `json:"name"`
	Version string `json:"version"`
}
type bundleDTO struct {
	Name string `json:"name"`
	Version string `json:"version"`
	Roles []roleRefDTO `json:"roles"`
	Includes []roleRefDTO `json:"includes"`
}
type scopeDTO struct {
	Kind authorization.ScopeKind `json:"kind"`
	TenantID string `json:"tenant_id"`
	ResourceType string `json:"resource_type"`
	ResourceID string `json:"resource_id,omitempty"`
	OrganizationID string `json:"organization_id,omitempty"`
	CollectionIDs []string `json:"collection_ids,omitempty"`
	RootType string `json:"root_type,omitempty"`
	RootID string `json:"root_id,omitempty"`
}
type assignmentDTO struct {
	ID string `json:"id"`
	TenantID string `json:"tenant_id"`
	SubjectType string `json:"subject_type"`
	SubjectID string `json:"subject_id"`
	TargetKind role.AssignmentTargetKind `json:"target_kind"`
	TargetName string `json:"target_name"`
	TargetVersion string `json:"target_version"`
	Scopes []scopeDTO `json:"scopes"`
	PolicyIDs []string `json:"policy_ids,omitempty"`
	Revision int64 `json:"revision"`
	Status role.ProfileAssignmentStatus `json:"status"`
	RevokedAt string `json:"revoked_at,omitempty"`
}

func encodeJSON(value any)(string,error){
	encoded,err:=json.Marshal(value)
	if err!=nil||len(encoded)==0||len(encoded)>maxRecordBytes{return "",ErrInvalidSnapshot}
	return string(encoded),nil
}
func decodeJSON(encoded string,destination any)error{
	if len(encoded)==0||len(encoded)>maxRecordBytes{return ErrInvalidSnapshot}
	if err:=json.Unmarshal([]byte(encoded),destination);err!=nil{return ErrInvalidSnapshot}
	return nil
}
func encodeBundle(bundle role.Bundle)(string,error){
	if bundle.Validate()!=nil{return "",ErrInvalidSnapshot}
	d:=bundleDTO{Name:bundle.Ref().Name,Version:bundle.Ref().Version}
	for _,ref:=range bundle.Roles(){d.Roles=append(d.Roles,roleRefDTO{Name:ref.Name,Version:ref.Version})}
	for _,ref:=range bundle.Includes(){d.Includes=append(d.Includes,roleRefDTO{Name:ref.Name,Version:ref.Version})}
	return encodeJSON(d)
}
func decodeBundle(encoded string)(role.Bundle,error){
	var d bundleDTO
	if err:=decodeJSON(encoded,&d);err!=nil{return role.Bundle{},err}
	roles:=make([]role.TemplateRef,0,len(d.Roles))
	children:=make([]role.BundleRef,0,len(d.Includes))
	for _,ref:=range d.Roles{roles=append(roles,role.TemplateRef{Name:ref.Name,Version:ref.Version})}
	for _,ref:=range d.Includes{children=append(children,role.BundleRef{Name:ref.Name,Version:ref.Version})}
	b,err:=role.NewBundle(role.BundleRef{Name:d.Name,Version:d.Version},roles,children)
	if err!=nil{return role.Bundle{},ErrInvalidSnapshot}
	return b,nil
}
func encodeAssignment(a role.ProfileAssignment)(string,error){
	if a.Validate()!=nil{return "",ErrInvalidSnapshot}
	d:=assignmentDTO{
		ID:a.ID(),TenantID:string(a.TenantID()),SubjectType:a.Subject().Type(),SubjectID:a.Subject().ID(),
		TargetKind:a.Target().Kind,TargetName:a.Target().Name,TargetVersion:a.Target().Version,
		PolicyIDs:a.PolicyIDs(),Revision:a.Revision(),Status:a.Status(),
	}
	if !a.RevokedAt().IsZero(){d.RevokedAt=a.RevokedAt().UTC().Format(time.RFC3339Nano)}
	for _,s:=range a.Scopes(){
		entry:=scopeDTO{
			Kind:s.Kind(),TenantID:string(s.TenantID()),ResourceType:s.ResourceType(),ResourceID:s.ResourceID(),
			OrganizationID:s.OrganizationID(),CollectionIDs:s.CollectionIDs(),
		}
		if s.Kind()==authorization.ScopeSubtree{entry.RootType=s.Root().Type();entry.RootID=s.Root().ID()}
		d.Scopes=append(d.Scopes,entry)
	}
	return encodeJSON(d)
}
func decodeAssignment(encoded string)(role.ProfileAssignment,error){
	var d assignmentDTO
	if err:=decodeJSON(encoded,&d);err!=nil{return role.ProfileAssignment{},err}
	tenant,err:=shared.NewTenantID(d.TenantID)
	if err!=nil{return role.ProfileAssignment{},ErrInvalidSnapshot}
	subject,err:=shared.NewSubject(d.SubjectType,d.SubjectID)
	if err!=nil{return role.ProfileAssignment{},ErrInvalidSnapshot}
	input:=role.ProfileAssignmentInput{
		ID:d.ID,TenantID:tenant,Subject:subject,
		Target:role.AssignmentTarget{Kind:d.TargetKind,Name:d.TargetName,Version:d.TargetVersion},
		PolicyIDs:d.PolicyIDs,
	}
	for _,s:=range d.Scopes{
		scopeTenant,err:=shared.NewTenantID(s.TenantID)
		if err!=nil{return role.ProfileAssignment{},ErrInvalidSnapshot}
		param:=authorization.ScopeInput{
			Kind:s.Kind,TenantID:scopeTenant,ResourceType:s.ResourceType,ResourceID:s.ResourceID,
			OrganizationID:s.OrganizationID,CollectionIDs:s.CollectionIDs,
		}
		if s.Kind()==authorization.ScopeSubtree{
			param.Root,err=shared.NewResource(s.RootType,s.RootID)
			if err!=nil{return role.ProfileAssignment{},ErrInvalidSnapshot}
		}else if s.RootType!=""||s.RootID!="" {
			return role.ProfileAssignment{},ErrInvalidSnapshot
		}
		scope,err:=authorization.NewScope(param)
		if err!=nil{return role.ProfileAssignment{},ErrInvalidSnapshot}
		input.Scopes=append(input.Scopes,scope)
	}
	var revokedAt time.Time
	if d.RevokedAt!=""{
		revokedAt,err=time.Parse(time.RFC3339Nano,d.RevokedAt)
		if err!=nil{return role.ProfileAssignment{},ErrInvalidSnapshot}
	}
	a,err:=role.RestoreProfileAssignment(input,d.Revision,d.Status,revokedAt)
	if err!=nil{return role.ProfileAssignment{},ErrInvalidSnapshot}
	return a,nil
}
func commandDigest(op string,expected int64,encoded string)string {
	sum:=sha256.Sum256([]byte(fmt.Sprintf("role-management/v1:%s:%d:%s",op,expected,encoded)))
	return hex.EncodeToString(sum[:])
}

package projection

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/domain"
	"github.com/m8-team/platform/internal/access/domain/authorization"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrInvalidProfileProjection = errors.New("invalid profile assignment projection source")
	ErrUnsupportedProfileScope = errors.New("profile scope requires authoritative owner/containment graph")
	ErrUnverifiedProfilePolicy = errors.New("contextual profile policy requires trusted PDP evaluator")
)

// ProfileCapability is a projection intent compiled from a published immutable
// template and checked owner proof, then persisted with the aggregate revision.
// Kafka is only a notification; it MUST NOT supply or alter these capabilities.
type ProfileCapability struct {
	ResourceType string `json:"resource_type"`
	ResourceID string `json:"resource_id"`
	Action string `json:"action"`
	OwnerRevision int64 `json:"owner_revision"`
}

// CapabilityObjectID intentionally hashes an unambiguous structured tuple.
// The tenant is separately prefixed by SpiceDB graphID. It is never possible
// for a new module's permissions to match a different resource/action tuple.
func CapabilityObjectID(resourceType,resourceID,action string)(string,error){
	if !shared.ValidQualifiedName(resourceType)||!shared.ValidQualifiedName(action)||
		resourceID==""||len(resourceID)>256||strings.TrimSpace(resourceID)!=resourceID{return "",ErrInvalidProfileProjection}
	raw,err:=json.Marshal([]string{resourceType,resourceID,action})
	if err!=nil{return "",err}
	sum:=sha256.Sum256(raw)
	return "v1_"+hex.EncodeToString(sum[:]),nil
}

func(c ProfileCapability) Validate()error{
	if c.OwnerRevision<1{return ErrInvalidProfileProjection}
	_,err:=CapabilityObjectID(c.ResourceType,c.ResourceID,c.Action)
	return err
}

// ProfileAssignmentState is a source-owned, revision-pinned, integrity-checked
// snapshot. Active permissions MUST be certified by policyjournal and a
// distributed fenced checkpoint; the snapshot alone never authorizes.
type ProfileAssignmentState struct{
	Assignment role.ProfileAssignment
	Capabilities []ProfileCapability
}

func(s ProfileAssignmentState) Validate()error{
	if s.Assignment.Validate()!=nil{return ErrInvalidProfileProjection}
	if len(s.Capabilities)>999{return ErrInvalidProfileProjection}
	if s.Assignment.Status()!=role.ProfileAssignmentActive{
		if len(s.Capabilities)!=0{return ErrInvalidProfileProjection}
		return nil
	}
	if len(s.Assignment.PolicyIDs())!=0{return ErrUnverifiedProfilePolicy}
	if len(s.Capabilities)==0{return ErrInvalidProfileProjection}
	scopes:=map[string]authorization.Scope{}
	for _,scope:=range s.Assignment.Scopes(){
		// Only INSTANCE is sound with the current graph schema. Do not
		// flatten organization/tenant/subtree/collection into arbitrary IDs.
		if scope.Kind()!=authorization.ScopeInstance{return ErrUnsupportedProfileScope}
		scopes[scope.ResourceType()]=scope
	}
	seen:=map[string]bool{}
	for _,capability:=range s.Capabilities{
		if capability.Validate()!=nil{return ErrInvalidProfileProjection}
		scope,ok:=scopes[capability.ResourceType]
		if !ok||scope.ResourceID()!=capability.ResourceID{return ErrInvalidProfileProjection}
		id,err:=CapabilityObjectID(capability.ResourceType,capability.ResourceID,capability.Action)
		if err!=nil||seen[id]{return ErrInvalidProfileProjection}
		seen[id]=true
	}
	if len(seen)<len(scopes){return ErrInvalidProfileProjection}
	return nil
}

func(p *Projector) profileAssignmentRelationships(_ context.Context,state *ProfileAssignmentState)([]domain.Relationship,error){
	if state==nil{return nil,nil}
	if err:=state.Validate();err!=nil{return nil,err}
	a:=state.Assignment
	if a.Status()!=role.ProfileAssignmentActive{return nil,nil}
	tenant,err:=domain.NewTenantID(string(a.TenantID()))
	if err!=nil{return nil,err}
	if a.Subject().Type()!="user"&&a.Subject().Type()!="service_account"{return nil,ErrUnsupportedSubject}
	holder,err:=domain.NewResource("access_profile_assignment",a.ID())
	if err!=nil{return nil,err}
	subject,err:=domain.NewSubject(a.Subject().Type(),a.Subject().ID())
	if err!=nil{return nil,err}
	holderEdge,err:=newRelationship(tenant,holder,"holder",subject,"",time.Time{})
	if err!=nil{return nil,err}
	out:=[]domain.Relationship{holderEdge}
	grantee,err:=domain.NewSubject("access_profile_assignment",a.ID())
	if err!=nil{return nil,err}
	for _,cap:=range state.Capabilities{
		id,err:=CapabilityObjectID(cap.ResourceType,cap.ResourceID,cap.Action)
		if err!=nil{return nil,err}
		target,err:=domain.NewResource("access_capability",id)
		if err!=nil{return nil,err}
		edge,err:=newRelationship(tenant,target,"grantee",grantee,"holder",time.Time{})
		if err!=nil{return nil,err}
		out=append(out,edge)
	}
	// Stable tuple order makes retries/checkpoint reconstruction deterministic.
	sort.Slice(out,func(i,j int)bool{return fmt.Sprint(out[i].Key())<fmt.Sprint(out[j].Key())})
	if len(out)>maxUpdates{return nil,ErrProjectionTooLarge}
	return out,nil
}

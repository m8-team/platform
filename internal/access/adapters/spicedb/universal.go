package spicedb

import (
	"context"

	pb "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

// CheckUniversalCapability uses an exact capability object and a fully
// consistent graph check. No caller-supplied ZedToken is accepted or merged.
// Contextual/partial permissions are intentionally denied until the trusted
// caveat evaluator is implemented.
func(c *Client) CheckUniversalCapability(ctx context.Context,tenant domain.TenantID,subject domain.Subject,resourceType,resourceID,action string)(bool,error){
	if c==nil||c.permissions==nil{return false,ErrInvalidConfig}
	if err:=ctx.Err();err!=nil{return false,err}
	if tenant.Validate()!=nil||subject.Validate()!=nil{return false,ErrInvalidGraphReference}
	if subject.Type()!="user"&&subject.Type()!="service_account"{return false,ErrInvalidGraphReference}
	id,err:=projection.CapabilityObjectID(resourceType,resourceID,action)
	if err!=nil{return false,ErrInvalidGraphReference}
	obj,err:=graphObject(tenant,"access_capability",id)
	if err!=nil{return false,err}
	s,err:=graphSubject(tenant,subject,"")
	if err!=nil{return false,err}
	response,err:=c.permissions.CheckPermission(ctx,&pb.CheckPermissionRequest{
		Resource:obj,Permission:"use",Subject:s,
		Consistency:&pb.Consistency{Requirement:&pb.Consistency_FullyConsistent{FullyConsistent:true}},
	})
	if err!=nil{return false,classifyRPCFailure(ctx,err)}
	if response==nil{return false,ErrInvalidResponse}
	switch response.Permissionship{
	case pb.CheckPermissionResponse_PERMISSIONSHIP_HAS_PERMISSION:return true,nil
	case pb.CheckPermissionResponse_PERMISSIONSHIP_NO_PERMISSION:return false,nil
	case pb.CheckPermissionResponse_PERMISSIONSHIP_CONDITIONAL_PERMISSION:return false,nil
	default:return false,ErrInvalidResponse
	}
}

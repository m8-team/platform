package grpcadapter

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"github.com/m8-team/platform/internal/resourcemanager/v1engine"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizationRelationshipServer struct {
	pb.UnimplementedOrganizationRelationshipServiceServer
	resources *ResourceServer
}

func NewOrganizationRelationshipServer(resources *ResourceServer) *OrganizationRelationshipServer {
	return &OrganizationRelationshipServer{resources: resources}
}

func (s *OrganizationRelationshipServer) guard() error {
	if s.resources == nil { return status.Error(codes.Unavailable, "Resource Manager unavailable") }
	return s.resources.guard()
}

func relationshipType(t pb.OrganizationRelationshipType) v1engine.RelationshipType {
	switch t {
	case pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_PARENT_SUBSIDIARY:
		return v1engine.RelationshipParentSubsidiary
	case pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_AFFILIATE:
		return v1engine.RelationshipAffiliate
	case pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_PARTNER:
		return v1engine.RelationshipPartner
	case pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_SERVICE_PROVIDER:
		return v1engine.RelationshipServiceProvider
	default:
		return ""
	}
}
func protoRelationshipType(t v1engine.RelationshipType) pb.OrganizationRelationshipType {
	switch t {
	case v1engine.RelationshipParentSubsidiary:
		return pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_PARENT_SUBSIDIARY
	case v1engine.RelationshipAffiliate:
		return pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_AFFILIATE
	case v1engine.RelationshipPartner:
		return pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_PARTNER
	case v1engine.RelationshipServiceProvider:
		return pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_SERVICE_PROVIDER
	}
	return pb.OrganizationRelationshipType_ORGANIZATION_RELATIONSHIP_TYPE_UNSPECIFIED
}
func protoRelationship(r v1engine.OrganizationRelationship) *pb.OrganizationRelationship {
	out:=&pb.OrganizationRelationship{
		Id:r.ID,TenantId:r.TenantID,
		FromOrganizationId:r.FromOrganizationID,ToOrganizationId:r.ToOrganizationID,
		Type:protoRelationshipType(r.Type),OwnershipBasisPoints:r.OwnershipBasisPoints,
		ValidFrom:timestamppb.New(time.Unix(r.ValidFrom,0).UTC()),
		EvidenceRef:r.EvidenceRef,Revision:r.Revision,
		State:pb.OrganizationRelationshipState_ORGANIZATION_RELATIONSHIP_STATE_ACTIVE,
		CreateTime:timestamppb.New(r.CreatedAt), UpdateTime:timestamppb.New(r.UpdatedAt),
	}
	if r.ValidTo!=0 { out.ValidTo=timestamppb.New(time.Unix(r.ValidTo,0).UTC()) }
	if r.State=="CLOSED" { out.State=pb.OrganizationRelationshipState_ORGANIZATION_RELATIONSHIP_STATE_CLOSED }
	return out
}
func validTimestamp(t *timestamppb.Timestamp) bool {
	return t != nil && t.CheckValid()==nil
}
func (s *OrganizationRelationshipServer) CreateOrganizationRelationship(ctx context.Context,req *pb.CreateOrganizationRelationshipRequest) (*pb.OrganizationRelationship,error) {
	if err:=s.guard();err!=nil{return nil,err}
	if req==nil || req.GetRelationship()==nil || req.GetIdempotencyKey()=="" {
		return nil,status.Error(codes.InvalidArgument,"relationship and idempotency_key required")
	}
	in:=req.GetRelationship()
	if !validTimestamp(in.GetValidFrom()) || (in.GetValidTo()!=nil && !validTimestamp(in.GetValidTo())) {
		return nil,status.Error(codes.InvalidArgument,"invalid temporal validity")
	}
	if in.GetId()!="" || in.GetRevision()!=0 || in.GetState()!=0 || in.GetCreateTime()!=nil || in.GetUpdateTime()!=nil {
		return nil,status.Error(codes.InvalidArgument,"server-assigned fields cannot be specified")
	}
	r:=v1engine.OrganizationRelationship{
		TenantID:in.GetTenantId(),FromOrganizationID:in.GetFromOrganizationId(),
		ToOrganizationID:in.GetToOrganizationId(), Type:relationshipType(in.GetType()),
		OwnershipBasisPoints:in.GetOwnershipBasisPoints(),
		ValidFrom:in.GetValidFrom().AsTime().Unix(),EvidenceRef:in.GetEvidenceRef(),
	}
	if in.GetValidTo()!=nil {r.ValidTo=in.GetValidTo().AsTime().Unix()}
	if err:=r.Validate();err!=nil{return nil,asErr(err)}
	out,err:=s.resources.store.CreateRelationship(ctx,r,req.GetIdempotencyKey())
	if err!=nil{return nil,asErr(err)}
	return protoRelationship(out),nil
}
func (s *OrganizationRelationshipServer) GetOrganizationRelationship(ctx context.Context,req *pb.GetOrganizationRelationshipRequest) (*pb.OrganizationRelationship,error) {
	if err:=s.guard();err!=nil{return nil,err}
	if req==nil || req.GetTenantId()=="" || req.GetId()=="" {return nil,status.Error(codes.InvalidArgument,"tenant_id and id required")}
	out,err:=s.resources.store.GetRelationship(ctx,req.GetTenantId(),req.GetId())
	if err!=nil{return nil,asErr(err)}
	return protoRelationship(out),nil
}

type relationshipPage struct {
	TenantID string `json:"tenant"`
	OrganizationID string `json:"org"`
	IncludeClosed bool `json:"closed"`
	Offset int `json:"offset"`
}
func parseRelationshipPage(req *pb.ListOrganizationRelationshipsRequest) (int,error) {
	if req.GetPageToken()=="" {return 0,nil}
	data,err:=base64.RawURLEncoding.DecodeString(req.GetPageToken())
	if err!=nil || len(data)>4096 {return 0,status.Error(codes.InvalidArgument,"invalid relationship page token")}
	var page relationshipPage
	if err=json.Unmarshal(data,&page);err!=nil ||
		page.TenantID!=req.GetTenantId() || page.OrganizationID!=req.GetOrganizationId() ||
		page.IncludeClosed!=req.GetIncludeClosed() || page.Offset<0 || page.Offset>10000000 {
		return 0,status.Error(codes.InvalidArgument,"page token does not match query scope")
	}
	return page.Offset,nil
}
func encodeRelationshipPage(req *pb.ListOrganizationRelationshipsRequest,offset int) string {
	data,_:=json.Marshal(relationshipPage{req.GetTenantId(),req.GetOrganizationId(),req.GetIncludeClosed(),offset})
	return base64.RawURLEncoding.EncodeToString(data)
}

func (s *OrganizationRelationshipServer) ListOrganizationRelationships(ctx context.Context,req *pb.ListOrganizationRelationshipsRequest) (*pb.ListOrganizationRelationshipsResponse,error) {
	if err:=s.guard();err!=nil{return nil,err}
	if req==nil || req.GetTenantId()=="" {return nil,status.Error(codes.InvalidArgument,"tenant_id required")}
	if req.GetPageSize()<0 || req.GetPageSize()>500 {return nil,status.Error(codes.InvalidArgument,"page_size must be 1..500")}
	offset,err:=parseRelationshipPage(req)
	if err!=nil{return nil,err}
	rows,err:=s.resources.store.ListRelationships(ctx,req.GetTenantId(),req.GetOrganizationId(),req.GetIncludeClosed())
	if err!=nil{return nil,asErr(err)}
	limit:=int(req.GetPageSize())
	if limit==0 {limit=100}
	out:=&pb.ListOrganizationRelationshipsResponse{}
	if offset>=len(rows) {
		if offset>0 && offset>len(rows) {return nil,status.Error(codes.Aborted,"relationship page is stale")}
		return out,nil
	}
	end:=offset+limit
	if end>len(rows){end=len(rows)}
	for _,r:=range rows[offset:end] {out.Relationships=append(out.Relationships,protoRelationship(r))}
	if end<len(rows){out.NextPageToken=encodeRelationshipPage(req,end)}
	return out,nil
}
func (s *OrganizationRelationshipServer) UpdateOrganizationRelationship(ctx context.Context,req *pb.UpdateOrganizationRelationshipRequest) (*pb.OrganizationRelationship,error) {
	if err:=s.guard();err!=nil{return nil,err}
	if req==nil || req.GetTenantId()=="" || req.GetId()=="" ||
		req.GetExpectedRevision()<=0 || req.GetIdempotencyKey()=="" {
		return nil,status.Error(codes.InvalidArgument,"id, tenant_id, revision and idempotency_key required")
	}
	if req.GetValidTo()!=nil && !validTimestamp(req.GetValidTo()) {
		return nil,status.Error(codes.InvalidArgument,"invalid valid_to")
	}
	u:=v1engine.RelationshipUpdate{
		TenantID:req.GetTenantId(),ID:req.GetId(),ExpectedRevision:req.GetExpectedRevision(),
		OwnershipBasisPoints:req.GetOwnershipBasisPoints(),EvidenceRef:req.GetEvidenceRef(),
	}
	if req.GetValidTo()!=nil {u.ValidTo=req.GetValidTo().AsTime().Unix()}
	r,err:=s.resources.store.UpdateRelationship(ctx,u,req.GetIdempotencyKey())
	if err!=nil{return nil,asErr(err)}
	return protoRelationship(r),nil
}
func (s *OrganizationRelationshipServer) CloseOrganizationRelationship(ctx context.Context,req *pb.CloseOrganizationRelationshipRequest) (*pb.OrganizationRelationship,error) {
	if err:=s.guard();err!=nil{return nil,err}
	if req==nil || req.GetTenantId()=="" || req.GetId()=="" ||
		req.GetExpectedRevision()<=0 || req.GetIdempotencyKey()=="" || !validTimestamp(req.GetClosedAt()) {
		return nil,status.Error(codes.InvalidArgument,"id, tenant_id, revision, key, closed_at required")
	}
	r,err:=s.resources.store.CloseRelationship(ctx,req.GetTenantId(),req.GetId(),
		req.GetExpectedRevision(),req.GetClosedAt().AsTime().Unix(),req.GetIdempotencyKey())
	if err!=nil{return nil,asErr(err)}
	return protoRelationship(r),nil
}


func (s *OrganizationRelationshipServer) ListOrganizationRelationshipVersions(ctx context.Context,
	req *pb.ListOrganizationRelationshipVersionsRequest) (*pb.ListOrganizationRelationshipVersionsResponse,error) {
	if err:=s.guard();err!=nil{return nil,err}
	if req==nil || req.GetTenantId()=="" || req.GetId()=="" {
		return nil,status.Error(codes.InvalidArgument,"tenant_id and id required")
	}
	if req.GetPageSize()<0 || req.GetPageSize()>500 {
		return nil,status.Error(codes.InvalidArgument,"page_size must be 1..500")
	}
	rows,err:=s.resources.store.ListRelationshipVersions(ctx,req.GetTenantId(),req.GetId())
	if err!=nil{return nil,asErr(err)}
	offset:=0
	if req.GetPageToken()!="" {
		raw,err:=base64.RawURLEncoding.DecodeString(req.GetPageToken())
		if err!=nil || len(raw)>2048{return nil,status.Error(codes.InvalidArgument,"invalid history token")}
		var token struct{Tenant,ID string;Offset int}
		if err:=json.Unmarshal(raw,&token);err!=nil || token.Tenant!=req.GetTenantId() || token.ID!=req.GetId() ||
			token.Offset<0 || token.Offset>len(rows) {
			return nil,status.Error(codes.InvalidArgument,"history token does not match scope")
		}
		offset=token.Offset
	}
	result:=&pb.ListOrganizationRelationshipVersionsResponse{}
	limit:=int(req.GetPageSize())
	if limit==0{limit=100}
	end:=offset+limit
	if end>len(rows){end=len(rows)}
	for _,r:=range rows[offset:end] {result.Versions=append(result.Versions,protoRelationship(r))}
	if end<len(rows){
		raw,_:=json.Marshal(struct{Tenant,ID string;Offset int}{req.GetTenantId(),req.GetId(),end})
		result.NextPageToken=base64.RawURLEncoding.EncodeToString(raw)
	}
	return result,nil
}

// Verify the generated contract is implemented at compile time.
var _ pb.OrganizationRelationshipServiceServer = (*OrganizationRelationshipServer)(nil)

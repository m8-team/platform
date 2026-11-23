package ydbroles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

// StoredAssignment is a durable assignment request, NEVER an authorization
// decision. PENDING_PROJECTION must be denied by all PDP/PEP consumers.
type StoredAssignment struct {
	Assignment role.ProfileAssignment
}

func(s *Store) getAssignment(ctx context.Context,q interface{
	QueryRowContext(context.Context,string,...any)*sql.Row
},tenant shared.TenantID,id string)(role.ProfileAssignment,error){
	var revision int64
	var status,payload string
	err:=q.QueryRowContext(ctx,s.q(selectAssignment),assignmentArgs(tenant,id)...).Scan(&revision,&status,&payload)
	if errors.Is(err,sql.ErrNoRows){return role.ProfileAssignment{},ErrNotFound}
	if err!=nil{return role.ProfileAssignment{},err}
	a,err:=decodeAssignment(payload)
	if err!=nil||a.TenantID()!=tenant||a.ID()!=id||a.Revision()!=revision||
		string(a.Status())!=status{return role.ProfileAssignment{},ErrInvalidSnapshot}
	return a,nil
}
func(s *Store) GetAssignment(ctx context.Context,tenant shared.TenantID,id string)(role.ProfileAssignment,error){
	if s==nil||s.db==nil{return role.ProfileAssignment{},ErrInvalidStore}
	if validateTenant(tenant)!=nil||!shared.ValidUUID(id){return role.ProfileAssignment{},ErrInvalidCommand}
	if err:=ctx.Err();err!=nil{return role.ProfileAssignment{},err}
	return s.getAssignment(ctx,s.db,tenant,id)
}

const selectAssignmentEvent = "DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;DECLARE $stream AS Utf8; SELECT payload_json FROM `m8-access/role_management_outbox` WHERE stream=$stream AND tenant_id=$tenant AND aggregate_id=$id AND revision=$revision;"

func(s *Store) replayAssignment(ctx context.Context,tx *sql.Tx,tenant shared.TenantID,id string,revision int64)(role.ProfileAssignment,error){
	var payload string
	err:=tx.QueryRowContext(ctx,s.q(selectAssignmentEvent),
		sql.Named("stream","assignment"),sql.Named("tenant",string(tenant)),sql.Named("id",id),sql.Named("revision",revision)).
		Scan(&payload)
	if errors.Is(err,sql.ErrNoRows){return role.ProfileAssignment{},ErrInvalidSnapshot}
	if err!=nil{return role.ProfileAssignment{},err}
	a,err:=decodeAssignment(payload)
	if err!=nil||a.TenantID()!=tenant||a.ID()!=id||a.Revision()!=revision{return role.ProfileAssignment{},ErrInvalidSnapshot}
	return a,nil
}
func(s *Store) receipt(ctx context.Context,tx *sql.Tx,tenant shared.TenantID,id,key,digest string)(role.ProfileAssignment,bool,error){
	var savedHash string
	var revision int64
	err:=tx.QueryRowContext(ctx,s.q(selectAssignmentReceipt),
		append(assignmentArgs(tenant,id),sql.Named("key",key))...).Scan(&savedHash,&revision)
	if errors.Is(err,sql.ErrNoRows){return role.ProfileAssignment{},false,nil}
	if err!=nil{return role.ProfileAssignment{},false,err}
	if savedHash!=digest{return role.ProfileAssignment{},false,ErrAssignmentConflict}
	record,err:=s.replayAssignment(ctx,tx,tenant,id,revision)
	return record,true,err
}
func(s *Store) saveAssignmentTx(ctx context.Context,tx *sql.Tx,a role.ProfileAssignment,payload string,now interface{} )error{
	_,err:=tx.ExecContext(ctx,s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;DECLARE $status AS Utf8;DECLARE $payload AS Utf8;\nUPSERT INTO `m8-access/profile_assignments` (tenant_id,assignment_id,revision,status,assignment_json) VALUES ($tenant,$id,$revision,$status,$payload);"),
		sql.Named("tenant",string(a.TenantID())),sql.Named("id",a.ID()),
		sql.Named("revision",a.Revision()),sql.Named("status",string(a.Status())),sql.Named("payload",payload))
	if err!=nil{return err}
	_,err=tx.ExecContext(ctx,s.q("DECLARE $tenant AS Utf8;DECLARE $type AS Utf8;DECLARE $subject AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Int64;DECLARE $status AS Utf8;DECLARE $payload AS Utf8;\nUPSERT INTO `m8-access/profile_assignment_subjects` (tenant_id,subject_type,subject_id,assignment_id,revision,status,assignment_json) VALUES ($tenant,$type,$subject,$id,$revision,$status,$payload);"),
		sql.Named("tenant",string(a.TenantID())),sql.Named("type",a.Subject().Type()),sql.Named("subject",a.Subject().ID()),
		sql.Named("id",a.ID()),sql.Named("revision",a.Revision()),sql.Named("status",string(a.Status())),sql.Named("payload",payload))
	return err
}
func(s *Store) saveReceiptTx(ctx context.Context,tx *sql.Tx,tenant shared.TenantID,id,key,digest string,revision int64)error{
	_,err:=tx.ExecContext(ctx,s.q("DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $key AS Utf8;DECLARE $hash AS Utf8;DECLARE $revision AS Int64;\nUPSERT INTO `m8-access/profile_assignment_receipts` (tenant_id,assignment_id,request_key,request_hash,revision) VALUES ($tenant,$id,$key,$hash,$revision);"),
		sql.Named("tenant",string(tenant)),sql.Named("id",id),sql.Named("key",key),
		sql.Named("hash",digest),sql.Named("revision",revision))
	return err
}

// CreateAssignment writes a STAGED request and its CQRS projection/outbox.
// It does NOT write a policy journal mutation or an effective SpiceDB grant.
func(s *Store) CreateAssignment(ctx context.Context,a role.ProfileAssignment,key string)(role.ProfileAssignment,error){
	if s==nil||s.db==nil{return role.ProfileAssignment{},ErrInvalidStore}
	if a.Validate()!=nil||a.Status()!=role.ProfileAssignmentPending||a.Revision()!=1||!validKey(key){
		return role.ProfileAssignment{},ErrInvalidCommand
	}
	payload,err:=encodeAssignment(a)
	if err!=nil{return role.ProfileAssignment{},ErrInvalidCommand}
	digest:=commandDigest("create",0,payload)
	var out role.ProfileAssignment
	err=s.tx(ctx,func(tx *sql.Tx)error{
		if saved,exists,err:=s.receipt(ctx,tx,a.TenantID(),a.ID(),key,digest);err!=nil{
			return err
		}else if exists{
			out=saved
			return nil
		}
		existing,err:=s.getAssignment(ctx,tx,a.TenantID(),a.ID())
		if err==nil&&existing.Validate()==nil{return ErrAssignmentConflict}
		if !errors.Is(err,ErrNotFound){return err}
		now,err:=s.serverTime(ctx,tx)
		if err!=nil{return err}
		if err:=s.saveAssignmentTx(ctx,tx,a,payload,now);err!=nil{return err}
		if err:=s.saveReceiptTx(ctx,tx,a.TenantID(),a.ID(),key,digest,1);err!=nil{return err}
		if err:=s.appendOutboxTx(ctx,tx,"assignment",string(a.TenantID()),a.ID(),1,
			"ProfileAssignmentStaged",payload,now);err!=nil{return err}
		out=a
		return nil
	})
	if err!=nil{return role.ProfileAssignment{},err}
	return out,nil
}

// RevokeAssignment is a CAS mutation of ONE independent assignment ID.
// It is safe to revoke a staged assignment: it has never been effective.
// Active assignment revocation will need a future tenant policy barrier writer.
func(s *Store) RevokeAssignment(ctx context.Context,tenant shared.TenantID,id string,expected int64,key string)(role.ProfileAssignment,error){
	if s==nil||s.db==nil{return role.ProfileAssignment{},ErrInvalidStore}
	if validateTenant(tenant)!=nil||!shared.ValidUUID(id)||expected<1||!validKey(key){
		return role.ProfileAssignment{},ErrInvalidCommand
	}
	digest:=commandDigest("revoke",expected,fmt.Sprintf("%s/%s",tenant,id))
	var out role.ProfileAssignment
	err:=s.tx(ctx,func(tx *sql.Tx)error{
		if saved,exists,err:=s.receipt(ctx,tx,tenant,id,key,digest);err!=nil{
			return err
		}else if exists{
			out=saved
			return nil
		}
		a,err:=s.getAssignment(ctx,tx,tenant,id)
		if err!=nil{return err}
		if a.Revision()!=expected{return shared.ErrRevisionConflict}
		now,err:=s.serverTime(ctx,tx)
		if err!=nil{return err}
		revoked,err:=a.Revoke(expected,now)
		if err!=nil{return err}
		payload,err:=encodeAssignment(revoked)
		if err!=nil{return err}
		if err:=s.saveAssignmentTx(ctx,tx,revoked,payload,now);err!=nil{return err}
		if err:=s.saveReceiptTx(ctx,tx,tenant,id,key,digest,revoked.Revision());err!=nil{return err}
		if err:=s.appendOutboxTx(ctx,tx,"assignment",string(tenant),id,revoked.Revision(),
			"ProfileAssignmentRevoked",payload,now);err!=nil{return err}
		out=revoked
		return nil
	})
	if err!=nil{return role.ProfileAssignment{},err}
	return out,nil
}

// ListBySubject returns an authorized read-model slice, ordered by UUID.
// The application MUST authorize tenant + subject before calling this port.
// No results ever represent effective authorization.
func(s *Store) ListBySubject(ctx context.Context,tenant shared.TenantID,subject shared.Subject,afterID string,limit int)([]role.ProfileAssignment,error){
	if s==nil||s.db==nil{return nil,ErrInvalidStore}
	if validateTenant(tenant)!=nil||subject.Validate()!=nil||limit<1||limit>100||
		(afterID!=""&&!shared.ValidUUID(afterID)){return nil,ErrInvalidCommand}
	rows,err:=s.db.QueryContext(ctx,s.q("DECLARE $tenant AS Utf8;DECLARE $type AS Utf8;DECLARE $subject AS Utf8;DECLARE $after AS Utf8;DECLARE $limit AS Uint64;\nSELECT assignment_id,revision,status,assignment_json FROM `m8-access/profile_assignment_subjects` WHERE tenant_id=$tenant AND subject_type=$type AND subject_id=$subject AND assignment_id>$after ORDER BY assignment_id LIMIT $limit;"),
		sql.Named("tenant",string(tenant)),sql.Named("type",subject.Type()),sql.Named("subject",subject.ID()),
		sql.Named("after",afterID),sql.Named("limit",uint64(limit)))
	if err!=nil{return nil,err}
	defer rows.Close()
	out:=make([]role.ProfileAssignment,0,limit)
	for rows.Next(){
		var id,status,payload string
		var rev int64
		if err:=rows.Scan(&id,&rev,&status,&payload);err!=nil{return nil,err}
		a,err:=decodeAssignment(payload)
		if err!=nil||a.TenantID()!=tenant||a.Subject()!=subject||
			a.ID()!=id||a.Revision()!=rev||string(a.Status())!=status||
			(afterID!=""&&strings.Compare(id,afterID)<=0){return nil,ErrInvalidSnapshot}
		out=append(out,a)
		afterID=id
	}
	if err:=rows.Err();err!=nil{return nil,err}
	return out,nil
}

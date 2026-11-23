package ydbroles

import (
	"context"
	"database/sql"
	"errors"
	"math"

	"github.com/m8-team/platform/internal/access/app/rolemanagement"
	"github.com/m8-team/platform/internal/access/domain/role"
	"github.com/m8-team/platform/internal/access/domain/shared"
)

// PublishedBundle is a version-pinned definition with its independent
// per-profile publication revision, not a tenant policy decision.
type PublishedBundle = rolemanagement.BundleRecord

var _ rolemanagement.BundleRepository = (*Store)(nil)
var _ rolemanagement.AssignmentRepository = (*Store)(nil)

func(s *Store) getBundle(ctx context.Context,q interface{
	QueryRowContext(context.Context,string,...any)*sql.Row
},ref role.BundleRef)(PublishedBundle,error){
	var revision int64
	var payload string
	err:=q.QueryRowContext(ctx,s.q(selectBundle),sql.Named("name",ref.Name),sql.Named("version",ref.Version)).Scan(&revision,&payload)
	if errors.Is(err,sql.ErrNoRows){return PublishedBundle{},ErrNotFound}
	if err!=nil{return PublishedBundle{},err}
	bundle,err:=decodeBundle(payload)
	if err!=nil||bundle.Ref()!=ref||revision<1{return PublishedBundle{},ErrInvalidSnapshot}
	return PublishedBundle{Bundle:bundle,Revision:revision},nil
}
func(s *Store) GetBundle(ctx context.Context,ref role.BundleRef)(PublishedBundle,error){
	if s==nil||s.db==nil{return PublishedBundle{},ErrInvalidStore}
	if ref.Validate()!=nil{return PublishedBundle{},ErrInvalidCommand}
	if err:=ctx.Err();err!=nil{return PublishedBundle{},err}
	return s.getBundle(ctx,s.db,ref)
}

// PublishBundle performs complete-version publication and replay deduplication.
// Validated cross-bundle dependencies are an app-layer prerequisite.
// YDB ensures a concurrent competing publication cannot both win the same CAS.
func(s *Store) PublishBundle(ctx context.Context,b role.Bundle,expected int64,key string)(PublishedBundle,error){
	if s==nil||s.db==nil{return PublishedBundle{},ErrInvalidStore}
	if err:=ctx.Err();err!=nil{return PublishedBundle{},err}
	if b.Validate()!=nil||expected<0||!validKey(key){return PublishedBundle{},ErrInvalidCommand}
	payload,err:=encodeBundle(b)
	if err!=nil{return PublishedBundle{},ErrInvalidCommand}
	ref:=b.Ref()
	digest:=commandDigest("publish-bundle",expected,payload)
	var out PublishedBundle
	err=s.tx(ctx,func(tx *sql.Tx)error{
		var receiptHash,receiptVersion string
		var receiptRev int64
		err:=tx.QueryRowContext(ctx,s.q(selectBundleReceipt),sql.Named("name",ref.Name),sql.Named("key",key)).
			Scan(&receiptHash,&receiptVersion,&receiptRev)
		if err==nil{
			if receiptHash!=digest||receiptVersion!=ref.Version{return ErrRoleConflict}
			record,err:=s.getBundle(ctx,tx,ref)
			if err!=nil{return err}
			if record.Revision!=receiptRev{return ErrInvalidSnapshot}
			out=record
			return nil
		}
		if !errors.Is(err,sql.ErrNoRows){return err}
		var head int64
		err=tx.QueryRowContext(ctx,s.q(selectBundleHead),sql.Named("name",ref.Name)).Scan(&head)
		if err!=nil&&!errors.Is(err,sql.ErrNoRows){return err}
		if head!=expected{return shared.ErrRevisionConflict}
		if head==math.MaxInt64{return ErrRoleConflict}
		var storedRev int64
		var storedPayload string
		err=tx.QueryRowContext(ctx,s.q(selectBundle),sql.Named("name",ref.Name),sql.Named("version",ref.Version)).
			Scan(&storedRev,&storedPayload)
		if err==nil{return ErrRoleConflict}
		if !errors.Is(err,sql.ErrNoRows){return err}
		now,err:=s.serverTime(ctx,tx)
		if err!=nil{return err}
		next:=head+1
		_,err=tx.ExecContext(ctx,s.q("DECLARE $name AS Utf8;DECLARE $revision AS Int64;DECLARE $updated AS Timestamp;\nUPSERT INTO `m8-access/role_bundle_heads` (profile_name,head_revision,updated_at) VALUES ($name,$revision,$updated);"),
			sql.Named("name",ref.Name),sql.Named("revision",next),sql.Named("updated",now))
		if err!=nil{return err}
		_,err=tx.ExecContext(ctx,s.q("DECLARE $name AS Utf8;DECLARE $version AS Utf8;DECLARE $revision AS Int64;DECLARE $payload AS Utf8;DECLARE $hash AS Utf8;DECLARE $created AS Timestamp;\nUPSERT INTO `m8-access/role_bundles` (profile_name,version,revision,bundle_json,bundle_hash,created_at) VALUES ($name,$version,$revision,$payload,$hash,$created);"),
			sql.Named("name",ref.Name),sql.Named("version",ref.Version),sql.Named("revision",next),
			sql.Named("payload",payload),sql.Named("hash",digest),sql.Named("created",now))
		if err!=nil{return err}
		_,err=tx.ExecContext(ctx,s.q("DECLARE $name AS Utf8;DECLARE $revision AS Int64;DECLARE $version AS Utf8;DECLARE $payload AS Utf8;\nUPSERT INTO `m8-access/role_bundle_read_index` (profile_name,revision,version,bundle_json) VALUES ($name,$revision,$version,$payload);"),
			sql.Named("name",ref.Name),sql.Named("revision",next),sql.Named("version",ref.Version),
			sql.Named("payload",payload))
		if err!=nil{return err}
		_,err=tx.ExecContext(ctx,s.q("DECLARE $name AS Utf8;DECLARE $key AS Utf8;DECLARE $hash AS Utf8;DECLARE $version AS Utf8;DECLARE $revision AS Int64;\nUPSERT INTO `m8-access/role_bundle_receipts` (profile_name,request_key,request_hash,version,revision) VALUES ($name,$key,$hash,$version,$revision);"),
			sql.Named("name",ref.Name),sql.Named("key",key),sql.Named("hash",digest),
			sql.Named("version",ref.Version),sql.Named("revision",next))
		if err!=nil{return err}
		// Global bundle definitions are never tenant entitlement mutations.
		if err:=s.appendOutboxTx(ctx,tx,"bundle","global",ref.Name,next,"RoleBundlePublished",payload,now);err!=nil{return err}
		out=PublishedBundle{Bundle:b,Revision:next}
		return nil
	})
	if err!=nil{return PublishedBundle{},err}
	return out,nil
}

func(s *Store) ListBundleVersions(ctx context.Context,name string,after int64,limit int)([]PublishedBundle,error){
	if s==nil||s.db==nil{return nil,ErrInvalidStore}
	if !shared.ValidQualifiedName(name)||after<0||limit<1||limit>100{return nil,ErrInvalidCommand}
	rows,err:=s.db.QueryContext(ctx,s.q("DECLARE $name AS Utf8;DECLARE $after AS Int64;DECLARE $limit AS Uint64;\nSELECT revision,version,bundle_json FROM `m8-access/role_bundle_read_index` WHERE profile_name=$name AND revision>$after ORDER BY revision LIMIT $limit;"),
		sql.Named("name",name),sql.Named("after",after),sql.Named("limit",uint64(limit)))
	if err!=nil{return nil,err}
	defer rows.Close()
	out:=make([]PublishedBundle,0,limit)
	for rows.Next(){
		var revision int64
		var version,payload string
		if err:=rows.Scan(&revision,&version,&payload);err!=nil{return nil,err}
		bundle,err:=decodeBundle(payload)
		if err!=nil||bundle.Ref().Name!=name||bundle.Ref().Version!=version||revision<=after{
			return nil,ErrInvalidSnapshot
		}
		out=append(out,PublishedBundle{Bundle:bundle,Revision:revision})
		after=revision
	}
	if err:=rows.Err();err!=nil{return nil,err}
	return out,nil
}

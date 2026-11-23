// Package ydbroles stores immutable composite profiles and staged scoped
// assignments without turning stored records into effective permissions.
package ydbroles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrInvalidStore       = errors.New("invalid YDB role management store")
	ErrRoleConflict       = errors.New("profile version or idempotency conflict")
	ErrAssignmentConflict = errors.New("assignment version or idempotency conflict")
	ErrNotFound           = errors.New("role management record not found")
	ErrInvalidCommand     = errors.New("invalid role management command")
)

var safePrefix = regexp.MustCompile(`^/[A-Za-z0-9_-]+(?:/[A-Za-z0-9_-]+)*$`)

// Store owns the write model, materialized read views and transactional event
// journal. It is not a PDP read adapter. Only app-layer authenticated commands
// should access the Store; no handlers are registered here.
type Store struct {
	db *sql.DB
	prefix string
}

func New(db *sql.DB, prefix string) (*Store,error) {
	if db == nil || !safePrefix.MatchString(prefix) {
		return nil,ErrInvalidStore
	}
	return &Store{db:db,prefix:prefix},nil
}

func(s *Store) q(query string) string {
	return fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s",s.prefix,query)
}
func(s *Store) tx(ctx context.Context,fn func(*sql.Tx)error)error{
	tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelSerializable})
	if err!=nil{return err}
	defer tx.Rollback()
	if err:=fn(tx);err!=nil{return err}
	return tx.Commit()
}
func(s *Store) serverTime(ctx context.Context,tx *sql.Tx)(time.Time,error){
	var now time.Time
	if err:=tx.QueryRowContext(ctx,s.q("SELECT CurrentUtcTimestamp();")).Scan(&now);err!=nil{
		return time.Time{},err
	}
	return now.UTC(),nil
}
func validKey(key string)bool{return key!=""&&len(key)<=256&&strings.TrimSpace(key)==key}
func validateTenant(tenant shared.TenantID)error{
	if err:=tenant.Validate();err!=nil{return ErrInvalidCommand}
	return nil
}

const selectBundleHead = "DECLARE $name AS Utf8; SELECT head_revision FROM `m8-access/role_bundle_heads` WHERE profile_name=$name;"
const selectBundleReceipt = "DECLARE $name AS Utf8;DECLARE $key AS Utf8; SELECT request_hash,version,revision FROM `m8-access/role_bundle_receipts` WHERE profile_name=$name AND request_key=$key;"
const selectBundle = "DECLARE $name AS Utf8;DECLARE $version AS Utf8; SELECT revision,bundle_json FROM `m8-access/role_bundles` WHERE profile_name=$name AND version=$version;"
const selectAssignment = "DECLARE $tenant AS Utf8;DECLARE $id AS Utf8; SELECT revision,status,assignment_json FROM `m8-access/profile_assignments` WHERE tenant_id=$tenant AND assignment_id=$id;"
const selectAssignmentReceipt = "DECLARE $tenant AS Utf8;DECLARE $id AS Utf8;DECLARE $key AS Utf8; SELECT request_hash,revision FROM `m8-access/profile_assignment_receipts` WHERE tenant_id=$tenant AND assignment_id=$id AND request_key=$key;"

func assignmentArgs(tenant shared.TenantID,id string)[]any{
	return []any{sql.Named("tenant",string(tenant)),sql.Named("id",id)}
}

// appendOutboxTx is deliberately part of the caller's serializable
// transaction. Its events do not advance the tenant policy consistency head.
func(s *Store) appendOutboxTx(ctx context.Context,tx *sql.Tx,stream,tenant,aggregate string,revision int64,eventKind,payload string,now time.Time)error{
	if stream!="bundle"&&stream!="assignment"{return ErrInvalidCommand}
	eventID:=fmt.Sprintf("%s/%s/%s/%d",stream,tenant,aggregate,revision)
	_,err:=tx.ExecContext(ctx,s.q("DECLARE $stream AS Utf8;DECLARE $tenant AS Utf8;DECLARE $aggregate AS Utf8;DECLARE $revision AS Int64;DECLARE $event AS Utf8;DECLARE $kind AS Utf8;DECLARE $payload AS Utf8;DECLARE $published AS Bool;DECLARE $created AS Timestamp;\nUPSERT INTO `m8-access/role_management_outbox` (stream,tenant_id,aggregate_id,revision,event_id,event_kind,payload_json,published,created_at) VALUES ($stream,$tenant,$aggregate,$revision,$event,$kind,$payload,$published,$created);"),
		sql.Named("stream",stream),sql.Named("tenant",tenant),sql.Named("aggregate",aggregate),
		sql.Named("revision",revision),sql.Named("event",eventID),sql.Named("kind",eventKind),
		sql.Named("payload",payload),sql.Named("published",false),sql.Named("created",now))
	return err
}

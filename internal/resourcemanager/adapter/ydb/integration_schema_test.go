package ydb

import (
    "context"
    "database/sql"
    "fmt"
    "os"
    "strings"

    ydbsdk "github.com/ydb-platform/ydb-go-sdk/v3"
)

// prepareIntegrationSchema runs only against a disposable test database.
// It does not use production migration credentials and must never be enabled
// against a shared or production database.
func prepareIntegrationSchema(ctx context.Context, db *sql.DB, dsn, prefix string) error {
    if os.Getenv("M8_RM_TEST_YDB_BOOTSTRAP")!="1" { return nil }
    if !strings.HasPrefix(prefix,"/local/") {return fmt.Errorf("test bootstrap only permits /local/ paths")}
    driver,err:=ydbsdk.Open(ctx,dsn)
    if err!=nil{return fmt.Errorf("bootstrap YDB driver: %w",err)}
    defer driver.Close(ctx)
    segments:=strings.Split(strings.TrimPrefix(prefix,"/local/"),"/")
    current:="/local"
    for _,segment:=range segments {
        current+="/"+segment
        if err:=driver.Scheme().MakeDirectory(ctx,current);err!=nil{
            if _,listErr:=driver.Scheme().ListDirectory(ctx,current);listErr!=nil{
                return fmt.Errorf("make %s: %w",current,err)
            }
        }
    }
    folder:=prefix+"/m8-resource-manager"
    if err:=driver.Scheme().MakeDirectory(ctx,folder);err!=nil{
        if _,listErr:=driver.Scheme().ListDirectory(ctx,folder);listErr!=nil{
            return fmt.Errorf("make %s: %w",folder,err)
        }
    }
    tables:=[]string{
        "CREATE TABLE `m8-resource-manager/org_structure_heads` ("+
         "tenant_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,active_revision Uint64 NOT NULL,"+
         "active_snapshot_id Utf8 NOT NULL,updated_at Timestamp NOT NULL,"+
         "PRIMARY KEY(tenant_id,organization_id));",
        "CREATE TABLE `m8-resource-manager/org_units` ("+
         "tenant_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,org_unit_id Utf8 NOT NULL,"+
         "authoritative_source_id Utf8,create_time Timestamp NOT NULL,"+
         "PRIMARY KEY(tenant_id,organization_id,org_unit_id));",
        "CREATE TABLE `m8-resource-manager/org_unit_snapshots` ("+
         "tenant_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,snapshot_id Utf8 NOT NULL,"+
         "org_unit_id Utf8 NOT NULL,parent_org_unit_id Utf8,display_name Utf8 NOT NULL,"+
         "lifecycle_state Utf8 NOT NULL,authoritative_source_id Utf8,"+
         "PRIMARY KEY(tenant_id,organization_id,snapshot_id,org_unit_id));",
        "CREATE TABLE `m8-resource-manager/org_unit_closure` ("+
         "tenant_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,snapshot_id Utf8 NOT NULL,"+
         "ancestor_id Utf8 NOT NULL,descendant_id Utf8 NOT NULL,depth Uint32 NOT NULL,"+
         "PRIMARY KEY(tenant_id,organization_id,snapshot_id,ancestor_id,descendant_id));",
        "CREATE TABLE `m8-resource-manager/org_unit_change_sets` ("+
         "tenant_id Utf8 NOT NULL,organization_id Utf8 NOT NULL,change_set_id Utf8 NOT NULL,"+
         "source_id Utf8,base_revision Uint64 NOT NULL,target_revision Uint64,staged_snapshot_id Utf8,"+
         "request_hash String NOT NULL,status Utf8 NOT NULL,effective_at Timestamp NOT NULL,"+
         "recorded_at Timestamp NOT NULL,changes JsonDocument NOT NULL,impact_evidence JsonDocument,"+
         "created_at Timestamp NOT NULL,updated_at Timestamp NOT NULL,"+
         "PRIMARY KEY(tenant_id,organization_id,change_set_id));",
        "CREATE TABLE `m8-resource-manager/outbox` ("+
         "event_id Utf8 NOT NULL,aggregate_type Utf8 NOT NULL,aggregate_id Utf8 NOT NULL,"+
         "aggregate_revision Uint64 NOT NULL,event_type Utf8 NOT NULL,topic Utf8 NOT NULL,"+
         "partition_key Utf8 NOT NULL,payload JsonDocument NOT NULL,status Utf8 NOT NULL,"+
         "attempt_count Uint32 NOT NULL,created_at Timestamp NOT NULL,PRIMARY KEY(event_id));",
    }
    for _,ddl:=range tables{
        if _,err:=db.ExecContext(ctx,fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s",prefix,ddl));err!=nil{
            return fmt.Errorf("create YDB test table: %w",err)
        }
    }
    return nil
}

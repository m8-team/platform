// Package ydb contains the durable resource catalog. All resource IDs are
// immutable and the organization root is provisioned inside the same YDB
// transaction as the Organization and its idempotency/outbox records.
package ydb

import (
 "context"
 "crypto/sha256"
 "database/sql"
 "encoding/hex"
 "encoding/json"
 "errors"
 "fmt"
 "time"

 "github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

type Catalog struct{ store *Store }

func NewCatalog(db *sql.DB, prefix string) (*Catalog,error) {
 s,err:=NewStore(db,prefix);if err!=nil{return nil,err}
 return &Catalog{store:s},nil
}

func (c *Catalog) stmt(query string) string{return c.store.q(query)}
func resourceTable()string{return "`m8-resource-manager/resource_records`"}
func createTable()string{return "`m8-resource-manager/resource_create_requests`"}
func orgHeadsTable()string{return "`m8-resource-manager/org_structure_heads`"}
func outboxTable()string{return "`m8-resource-manager/outbox`"}
func envKeysTable()string{return "`m8-resource-manager/resource_environment_keys`"}

func scanResource(row *sql.Row)(v1engine.Record,error){
 var r v1engine.Record
 var k,ownerKind string
 err:=row.Scan(&k,&r.ID,&r.TenantID,&ownerKind,&r.Owner.ID,&r.OrganizationID,&r.Name,&r.Description,&r.State,&r.Revision,&r.EnvironmentKey,&r.CreatedAt,&r.UpdatedAt)
 if err!=nil{if errors.Is(err,sql.ErrNoRows){return r,v1engine.ErrNotFound};return r,err}
 r.Kind=v1engine.Kind(k);r.Owner.Kind=v1engine.Kind(ownerKind)
 return r,nil
}
func (c *Catalog) get(ctx context.Context,q queryer,kind v1engine.Kind,id string)(v1engine.Record,error){
 return scanResource(q.QueryRowContext(ctx,c.stmt(fmt.Sprintf(
  "DECLARE $kind AS Utf8; DECLARE $id AS Utf8; SELECT kind,id,tenant_id,owner_kind,owner_id,organization_id,display_name,description,state,revision,environment_key,created_at,updated_at FROM %s WHERE kind=$kind AND id=$id;",
  resourceTable())),sql.Named("kind",string(kind)),sql.Named("id",id)))
}
func (c *Catalog) Get(ctx context.Context,kind v1engine.Kind,id string)(v1engine.Record,error) {
 if id=="" || kind==""{return v1engine.Record{},v1engine.ErrInvalid}
 return c.get(ctx,c.store.db,kind,id)
}
func createHash(r v1engine.Record) (string,error) {
 payload,err:=json.Marshal(struct{
  Kind v1engine.Kind
  Owner v1engine.Owner
  Name string
  Description string
  EnvironmentKey string
 }{r.Kind,r.Owner,r.Name,r.Description,r.EnvironmentKey})
 if err!=nil{return "",err}
 sum:=sha256.Sum256(payload);return hex.EncodeToString(sum[:]),nil
}
func (c *Catalog) lookupRequest(ctx context.Context,q queryer,scope,key string)(string,string,string,error){
 var hash,kind,id string
 err:=q.QueryRowContext(ctx,c.stmt(fmt.Sprintf(
  "DECLARE $scope AS Utf8; DECLARE $key AS Utf8; SELECT request_hash,resource_kind,resource_id FROM %s WHERE scope=$scope AND idempotency_key=$key;",createTable())),
  sql.Named("scope",scope),sql.Named("key",key)).Scan(&hash,&kind,&id)
 if errors.Is(err,sql.ErrNoRows){return "","","",nil}
 return hash,kind,id,err
}
// Create uses one serializable YDB transaction for all resource kinds.
// An Organization additionally receives an empty OrgUnit revision-1 head in
// that exact transaction, never a separately acknowledged bootstrap.
func (c *Catalog) Create(ctx context.Context,r v1engine.Record,idem string)(v1engine.Record,error){
 if c==nil||c.store==nil||idem==""||r.Name==""||!v1engine.AllowedParent(r.Kind,r.Owner.Kind) {return v1engine.Record{},v1engine.ErrInvalid}
 if r.Kind!=v1engine.Tenant&&r.Owner.ID=="" {return v1engine.Record{},v1engine.ErrInvalid}
 if r.Kind==v1engine.Environment&&r.EnvironmentKey==""{return v1engine.Record{},v1engine.ErrInvalid}
 hash,err:=createHash(r);if err!=nil{return v1engine.Record{},err}
 tx,err:=c.store.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelSerializable})
 if err!=nil{return v1engine.Record{},err}
 defer tx.Rollback()
 // Tenant creation is platform-scoped; all other keys are scoped to a parent
 // tenant determined from the YDB-authoritative parent row.
 scope:="tenant:create"
 if r.Kind!=v1engine.Tenant{
  parent,err:=c.get(ctx,tx,r.Owner.Kind,r.Owner.ID)
  if err!=nil{return v1engine.Record{},err}
  if parent.State!="ACTIVE"{return v1engine.Record{},v1engine.ErrConflict}
  if r.Kind==v1engine.Group||r.Kind==v1engine.Organization{r.TenantID=parent.ID}else{r.TenantID=parent.TenantID}
  r.OrganizationID=parent.OrganizationID
  if parent.Kind==v1engine.Organization {r.OrganizationID=parent.ID}
  scope=r.TenantID
 }
 previousHash,previousKind,previousID,err:=c.lookupRequest(ctx,tx,scope,idem)
 if err!=nil{return v1engine.Record{},err}
 if previousID!="" {
  if previousHash!=hash{return v1engine.Record{},v1engine.ErrConflict}
  return c.get(ctx,tx,v1engine.Kind(previousKind),previousID)
 }
 if r.Kind==v1engine.Environment {
  var existing string
  err=tx.QueryRowContext(ctx,c.stmt(fmt.Sprintf("DECLARE $org AS Utf8; DECLARE $key AS Utf8; SELECT environment_id FROM %s WHERE organization_id=$org AND environment_key=$key;",envKeysTable())),sql.Named("org",r.OrganizationID),sql.Named("key",r.EnvironmentKey)).Scan(&existing)
  if err==nil {return v1engine.Record{},v1engine.ErrConflict}
  if !errors.Is(err,sql.ErrNoRows){return v1engine.Record{},err}
 }
 id,err:=newID();if err!=nil{return v1engine.Record{},err}
 now:=time.Now().UTC()
 r.ID=id;r.State="ACTIVE";r.Revision=1;r.CreatedAt=now;r.UpdatedAt=now
 if r.Kind==v1engine.Tenant{r.TenantID=id}
 _,err=tx.ExecContext(ctx,c.stmt(fmt.Sprintf(
 "DECLARE $kind AS Utf8;DECLARE $id AS Utf8;DECLARE $tenant AS Utf8;DECLARE $ownerKind AS Utf8;DECLARE $owner AS Utf8;DECLARE $org AS Utf8;DECLARE $name AS Utf8;DECLARE $description AS Utf8;DECLARE $state AS Utf8;DECLARE $revision AS Int64;DECLARE $env AS Utf8;DECLARE $now AS Timestamp; INSERT INTO %s (kind,id,tenant_id,owner_kind,owner_id,organization_id,display_name,description,state,revision,environment_key,created_at,updated_at) VALUES($kind,$id,$tenant,$ownerKind,$owner,$org,$name,$description,$state,$revision,$env,$now,$now);",resourceTable())),
 sql.Named("kind",string(r.Kind)),sql.Named("id",id),sql.Named("tenant",r.TenantID),
 sql.Named("ownerKind",string(r.Owner.Kind)),sql.Named("owner",r.Owner.ID),
 sql.Named("org",r.OrganizationID),sql.Named("name",r.Name),sql.Named("description",r.Description),
 sql.Named("state",r.State),sql.Named("revision",r.Revision),sql.Named("env",r.EnvironmentKey),sql.Named("now",now))
 if err!=nil{return v1engine.Record{},err}
 if r.Kind==v1engine.Organization{
  // Empty initial snapshot: no OrgUnit rows or closure rows are necessary.
  // The Organization resource ID is stable and is its structure namespace.
  snapshot,err:=newID();if err!=nil{return v1engine.Record{},err}
  _,err=tx.ExecContext(ctx,c.stmt(fmt.Sprintf(
   "DECLARE $tenant AS Utf8;DECLARE $org AS Utf8;DECLARE $snap AS Utf8;DECLARE $now AS Timestamp; INSERT INTO %s (tenant_id,organization_id,active_revision,active_snapshot_id,updated_at) VALUES($tenant,$org,1u,$snap,$now);",orgHeadsTable())),
   sql.Named("tenant",r.TenantID),sql.Named("org",r.ID),sql.Named("snap",snapshot),sql.Named("now",now))
  if err!=nil{return v1engine.Record{},err}
 }
 if r.Kind==v1engine.Environment{
  _,err=tx.ExecContext(ctx,c.stmt(fmt.Sprintf("DECLARE $org AS Utf8;DECLARE $key AS Utf8;DECLARE $id AS Utf8;INSERT INTO %s (organization_id,environment_key,environment_id) VALUES($org,$key,$id);",envKeysTable())),sql.Named("org",r.OrganizationID),sql.Named("key",r.EnvironmentKey),sql.Named("id",id))
  if err!=nil{return v1engine.Record{},err}
 }
 _,err=tx.ExecContext(ctx,c.stmt(fmt.Sprintf("DECLARE $scope AS Utf8;DECLARE $key AS Utf8;DECLARE $hash AS String;DECLARE $kind AS Utf8;DECLARE $id AS Utf8;DECLARE $now AS Timestamp; INSERT INTO %s (scope,idempotency_key,request_hash,resource_kind,resource_id,created_at) VALUES($scope,$key,$hash,$kind,$id,$now);",createTable())),
  sql.Named("scope",scope),sql.Named("key",idem),sql.Named("hash",[]byte(hash)),sql.Named("kind",string(r.Kind)),sql.Named("id",id),sql.Named("now",now))
 if err!=nil{return v1engine.Record{},err}
 evtID,err:=newID();if err!=nil{return v1engine.Record{},err}
 data,err:=json.Marshal(struct{ TenantID,ResourceKind,ResourceID,OrganizationID string }{r.TenantID,string(r.Kind),id,r.OrganizationID})
 if err!=nil{return v1engine.Record{},err}
 _,err=tx.ExecContext(ctx,c.stmt(fmt.Sprintf("DECLARE $event AS Utf8;DECLARE $kind AS Utf8;DECLARE $id AS Utf8;DECLARE $tenant AS Utf8;DECLARE $payload AS Utf8;DECLARE $now AS Timestamp; INSERT INTO %s (event_id,aggregate_type,aggregate_id,aggregate_revision,event_type,topic,partition_key,payload,status,attempt_count,created_at) VALUES($event,$kind,$id,1u,'m8.resourcemanager.resource.created.v1','m8.platform.resource-manager.resources.v1',$tenant,Unwrap(CAST($payload AS JsonDocument)),'PENDING',0u,$now);",outboxTable())),
  sql.Named("event",evtID),sql.Named("kind",string(r.Kind)),sql.Named("id",id),sql.Named("tenant",r.TenantID),sql.Named("payload",string(data)),sql.Named("now",now))
 if err!=nil{return v1engine.Record{},err}
 if err=tx.Commit();err!=nil{return v1engine.Record{},err}
 return r,nil
}
func (c *Catalog) List(ctx context.Context,kind v1engine.Kind,organizationID string)([]v1engine.Record,error) {
 if organizationID==""{return nil,v1engine.ErrInvalid}
 rows,err:=c.store.db.QueryContext(ctx,c.stmt(fmt.Sprintf("DECLARE $org AS Utf8; SELECT kind,id,tenant_id,owner_kind,owner_id,organization_id,display_name,description,state,revision,environment_key,created_at,updated_at FROM %s WHERE organization_id=$org OR (kind='organization' AND id=$org) ORDER BY id;",resourceTable())),sql.Named("org",organizationID))
 if err!=nil{return nil,err}
 defer rows.Close()
 out:=[]v1engine.Record{}
 for rows.Next(){
  var r v1engine.Record
  var k,p string
  if err=rows.Scan(&k,&r.ID,&r.TenantID,&p,&r.Owner.ID,&r.OrganizationID,&r.Name,&r.Description,&r.State,&r.Revision,&r.EnvironmentKey,&r.CreatedAt,&r.UpdatedAt);err!=nil{return nil,err}
  r.Kind=v1engine.Kind(k);r.Owner.Kind=v1engine.Kind(p)
  if kind==""||kind==r.Kind{out=append(out,r)}
 }
 return out,rows.Err()
}

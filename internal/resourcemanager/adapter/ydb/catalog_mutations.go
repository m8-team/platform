package ydb

import (
 "context"
 "database/sql"
 "encoding/json"
 "errors"
 "fmt"
 "sort"
 "strings"
 "time"

 "github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

func (c *Catalog) transaction(ctx context.Context,run func(*sql.Tx)error)error{
 tx,err:=c.store.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelSerializable})
 if err!=nil{return err}
 defer tx.Rollback()
 if err=run(tx);err!=nil{return err}
 return tx.Commit()
}
func (c *Catalog) event(ctx context.Context,tx *sql.Tx,r v1engine.Record,eventType string)error{
 id,err:=newID();if err!=nil{return err}
 payload,err:=json.Marshal(struct{Kind string;ID string;Revision int64;TenantID string}{string(r.Kind),r.ID,r.Revision,r.TenantID})
 if err!=nil{return err}
 _,err=tx.ExecContext(ctx,c.stmt(fmt.Sprintf("DECLARE $event AS Utf8;DECLARE $kind AS Utf8;DECLARE $id AS Utf8;DECLARE $revision AS Uint64;DECLARE $type AS Utf8;DECLARE $tenant AS Utf8;DECLARE $payload AS Utf8;DECLARE $at AS Timestamp;INSERT INTO %s (event_id,aggregate_type,aggregate_id,aggregate_revision,event_type,topic,partition_key,payload,status,attempt_count,created_at) VALUES($event,$kind,$id,$revision,$type,'m8.platform.resource-manager.resources.v1',$tenant,Unwrap(CAST($payload AS JsonDocument)),'PENDING',0u,$at);",outboxTable())),
 sql.Named("event",id),sql.Named("kind",string(r.Kind)),sql.Named("id",r.ID),
 sql.Named("revision",uint64(r.Revision)),sql.Named("type",eventType),sql.Named("tenant",r.TenantID),sql.Named("payload",string(payload)),sql.Named("at",time.Now().UTC()))
 return err
}
func (c *Catalog) save(ctx context.Context,tx *sql.Tx,r v1engine.Record)error{
 _,err:=tx.ExecContext(ctx,c.stmt(fmt.Sprintf(
 "DECLARE $kind AS Utf8;DECLARE $id AS Utf8;DECLARE $name AS Utf8;DECLARE $description AS Utf8;DECLARE $state AS Utf8;DECLARE $revision AS Int64;DECLARE $ownerKind AS Utf8;DECLARE $owner AS Utf8;DECLARE $org AS Utf8;DECLARE $updated AS Timestamp;UPDATE %s SET display_name=$name,description=$description,state=$state,revision=$revision,owner_kind=$ownerKind,owner_id=$owner,organization_id=$org,updated_at=$updated WHERE kind=$kind AND id=$id;",resourceTable())),
 sql.Named("kind",string(r.Kind)),sql.Named("id",r.ID),sql.Named("name",r.Name),sql.Named("description",r.Description),sql.Named("state",r.State),sql.Named("revision",r.Revision),sql.Named("ownerKind",string(r.Owner.Kind)),sql.Named("owner",r.Owner.ID),sql.Named("org",r.OrganizationID),sql.Named("updated",r.UpdatedAt))
 return err
}
func (c *Catalog) Update(ctx context.Context,kind v1engine.Kind,id string,rev int64,name,description string)(v1engine.Record,error){
 if rev<=0||strings.TrimSpace(name)==""{return v1engine.Record{},v1engine.ErrInvalid}
 var out v1engine.Record
 err:=c.transaction(ctx,func(tx *sql.Tx)error{
  r,err:=c.get(ctx,tx,kind,id);if err!=nil{return err}
  if r.State!="ACTIVE"||r.Revision!=rev{return v1engine.ErrConflict}
  r.Name=name;r.Description=description;r.Revision++;r.UpdatedAt=time.Now().UTC()
  if err=c.save(ctx,tx,r);err!=nil{return err}
  if err=c.event(ctx,tx,r,"m8.resourcemanager.resource.updated.v1");err!=nil{return err}
  out=r;return nil
 })
 return out,err
}
func(c *Catalog) hasChild(ctx context.Context,tx *sql.Tx,parent v1engine.Owner)(bool,error){
 var id string
 err:=tx.QueryRowContext(ctx,c.stmt(fmt.Sprintf(
 "DECLARE $kind AS Utf8;DECLARE $id AS Utf8;SELECT id FROM %s WHERE owner_kind=$kind AND owner_id=$id AND state!='DELETED' LIMIT 1;",resourceTable())),
 sql.Named("kind",string(parent.Kind)),sql.Named("id",parent.ID)).Scan(&id)
 if errors.Is(err,sql.ErrNoRows){return false,nil}
 if err!=nil{return false,err}
 return true,nil
}
func(c *Catalog) ChangeState(ctx context.Context,kind v1engine.Kind,id string,rev int64,state string)(v1engine.Record,error){
 if rev<=0||(state!="SUSPENDED"&&state!="ACTIVE"&&state!="ARCHIVED"&&state!="DELETED"){return v1engine.Record{},v1engine.ErrInvalid}
 var out v1engine.Record
 err:=c.transaction(ctx,func(tx *sql.Tx)error{
  r,err:=c.get(ctx,tx,kind,id);if err!=nil{return err}
  if r.Revision!=rev||r.State=="DELETED"{return v1engine.ErrConflict}
  if state=="DELETED"||state=="ARCHIVED"{
   yes,err:=c.hasChild(ctx,tx,v1engine.Owner{Kind:kind,ID:id});if err!=nil{return err}
   if yes{return v1engine.ErrConflict}
  }
  r.State=state;r.Revision++;r.UpdatedAt=time.Now().UTC()
  if err=c.save(ctx,tx,r);err!=nil{return err}
  if err=c.event(ctx,tx,r,"m8.resourcemanager.resource.state_changed.v1");err!=nil{return err}
  out=r;return nil
 })
 return out,err
}
func(c *Catalog) MoveProject(ctx context.Context,id string,parent v1engine.Owner,rev int64)(v1engine.Record,error){
 if !v1engine.AllowedParent(v1engine.Project,parent.Kind)||parent.ID==""||rev<=0{return v1engine.Record{},v1engine.ErrInvalid}
 var out v1engine.Record
 err:=c.transaction(ctx,func(tx *sql.Tx)error{
  r,err:=c.get(ctx,tx,v1engine.Project,id);if err!=nil{return err}
  p,err:=c.get(ctx,tx,parent.Kind,parent.ID);if err!=nil{return err}
  if r.Revision!=rev||r.State!="ACTIVE"||p.State!="ACTIVE"{return v1engine.ErrConflict}
  org:=p.OrganizationID
  if p.Kind==v1engine.Organization{org=p.ID}
  if org!=r.OrganizationID||r.TenantID!=p.TenantID{return v1engine.ErrInvalid}
  r.Owner=parent;r.Revision++;r.UpdatedAt=time.Now().UTC()
  if err=c.save(ctx,tx,r);err!=nil{return err}
  if err=c.event(ctx,tx,r,"m8.resourcemanager.resource.moved.v1");err!=nil{return err}
  out=r;return nil
 })
 return out,err
}
func(c *Catalog) Resolve(ctx context.Context,kind v1engine.Kind,id string)([]v1engine.Owner,error){
 var out []v1engine.Owner
 visited:=map[string]bool{}
 for i:=0;i<20;i++{
  r,err:=c.Get(ctx,kind,id);if err!=nil{return nil,err}
  key:=string(kind)+":"+id
  if visited[key]{return nil,v1engine.ErrInvalid}
  visited[key]=true
  out=append([]v1engine.Owner{{Kind:kind,ID:id}},out...)
  if r.Owner.Kind==""{return out,nil}
  kind,id=r.Owner.Kind,r.Owner.ID
 }
 return nil,v1engine.ErrInvalid
}
func(c *Catalog) SetGroupLink(ctx context.Context,groupID,orgID string)error{
 if groupID==""||orgID==""{return v1engine.ErrInvalid}
 return c.transaction(ctx,func(tx *sql.Tx)error{
  g,err:=c.get(ctx,tx,v1engine.Group,groupID);if err!=nil{return err}
  o,err:=c.get(ctx,tx,v1engine.Organization,orgID);if err!=nil{return err}
  if g.State!="ACTIVE"||o.State!="ACTIVE"||g.TenantID!=o.TenantID{return v1engine.ErrInvalid}
  _,err=tx.ExecContext(ctx,c.stmt("DECLARE $group AS Utf8;DECLARE $org AS Utf8;DECLARE $tenant AS Utf8;DECLARE $now AS Timestamp;UPSERT INTO `m8-resource-manager/resource_group_affiliations` (group_id,organization_id,tenant_id,created_at) VALUES($group,$org,$tenant,$now);"),
   sql.Named("group",groupID),sql.Named("org",orgID),sql.Named("tenant",o.TenantID),sql.Named("now",time.Now().UTC()))
  if err!=nil{return err}
  return c.event(ctx,tx,g,"m8.resourcemanager.group_affiliation.changed.v1")
 })
}
func(c *Catalog) RemoveGroupLink(ctx context.Context,groupID,orgID string)error{
 if groupID==""||orgID==""{return v1engine.ErrInvalid}
 return c.transaction(ctx,func(tx *sql.Tx)error{
  var found string
  err:=tx.QueryRowContext(ctx,c.stmt("DECLARE $group AS Utf8;DECLARE $org AS Utf8;SELECT group_id FROM `m8-resource-manager/resource_group_affiliations` WHERE group_id=$group AND organization_id=$org;"),
   sql.Named("group",groupID),sql.Named("org",orgID)).Scan(&found)
  if errors.Is(err,sql.ErrNoRows){return v1engine.ErrNotFound}
  if err!=nil{return err}
  g,err:=c.get(ctx,tx,v1engine.Group,groupID);if err!=nil{return err}
  _,err=tx.ExecContext(ctx,c.stmt("DECLARE $group AS Utf8;DECLARE $org AS Utf8;DELETE FROM `m8-resource-manager/resource_group_affiliations` WHERE group_id=$group AND organization_id=$org;"),
   sql.Named("group",groupID),sql.Named("org",orgID))
  if err!=nil{return err}
  return c.event(ctx,tx,g,"m8.resourcemanager.group_affiliation.removed.v1")
 })
}
func(c *Catalog) ListGroupOrganizations(ctx context.Context,groupID string)([]string,error){
 if _,err:=c.Get(ctx,v1engine.Group,groupID);err!=nil{return nil,err}
 rows,err:=c.store.db.QueryContext(ctx,c.stmt("DECLARE $group AS Utf8;SELECT organization_id FROM `m8-resource-manager/resource_group_affiliations` WHERE group_id=$group;"),sql.Named("group",groupID))
 if err!=nil{return nil,err}
 defer rows.Close()
 out:=[]string{}
 for rows.Next(){var id string;if err=rows.Scan(&id);err!=nil{return nil,err};out=append(out,id)}
 sort.Strings(out)
 return out,rows.Err()
}

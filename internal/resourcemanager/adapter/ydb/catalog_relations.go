package ydb

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "sort"
 "time"

 "github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

func(c *Catalog) SetBinding(ctx context.Context,resource v1engine.Owner,envID,version,config string)(v1engine.Binding,error){
 if resource.ID==""||resource.Kind==""||envID==""||version==""{return v1engine.Binding{},v1engine.ErrInvalid}
 var out v1engine.Binding
 err:=c.transaction(ctx,func(tx *sql.Tx)error{
  r,err:=c.get(ctx,tx,resource.Kind,resource.ID);if err!=nil{return err}
  env,err:=c.get(ctx,tx,v1engine.Environment,envID);if err!=nil{return err}
  if r.State!="ACTIVE"||env.State!="ACTIVE"||r.OrganizationID==""||r.OrganizationID!=env.OrganizationID{return v1engine.ErrInvalid}
  var id string
  err=tx.QueryRowContext(ctx,c.stmt(
   "DECLARE $kind AS Utf8;DECLARE $resource AS Utf8;DECLARE $env AS Utf8;SELECT id FROM `m8-resource-manager/resource_environment_bindings` WHERE resource_kind=$kind AND resource_id=$resource AND environment_id=$env LIMIT 1;"),
   sql.Named("kind",string(resource.Kind)),sql.Named("resource",resource.ID),sql.Named("env",envID)).Scan(&id)
  if err!=nil&&!errors.Is(err,sql.ErrNoRows){return err}
  now:=time.Now().UTC()
  if id!="" {
   var b v1engine.Binding
   b.ID=id;b.Resource=resource;b.EnvironmentID=envID
   err=tx.QueryRowContext(ctx,c.stmt(
    "DECLARE $id AS Utf8;SELECT version_ref,configuration_ref,revision FROM `m8-resource-manager/resource_environment_bindings` WHERE id=$id;"),sql.Named("id",id)).Scan(&b.Version,&b.ConfigRef,&b.Revision)
   if err!=nil{return err}
   b.Version=version;b.ConfigRef=config;b.Revision++
   _,err=tx.ExecContext(ctx,c.stmt(
    "DECLARE $id AS Utf8;DECLARE $version AS Utf8;DECLARE $config AS Utf8;DECLARE $revision AS Int64;DECLARE $now AS Timestamp;UPDATE `m8-resource-manager/resource_environment_bindings` SET version_ref=$version,configuration_ref=$config,revision=$revision,updated_at=$now WHERE id=$id;"),
    sql.Named("id",id),sql.Named("version",version),sql.Named("config",config),sql.Named("revision",b.Revision),sql.Named("now",now))
   out=b
  }else{
   id,err=newID();if err!=nil{return err}
   out=v1engine.Binding{ID:id,Resource:resource,EnvironmentID:envID,Version:version,ConfigRef:config,Revision:1}
   _,err=tx.ExecContext(ctx,c.stmt(
    "DECLARE $id AS Utf8;DECLARE $kind AS Utf8;DECLARE $resource AS Utf8;DECLARE $org AS Utf8;DECLARE $env AS Utf8;DECLARE $version AS Utf8;DECLARE $config AS Utf8;DECLARE $now AS Timestamp;INSERT INTO `m8-resource-manager/resource_environment_bindings` (id,resource_kind,resource_id,organization_id,environment_id,version_ref,configuration_ref,revision,created_at,updated_at) VALUES($id,$kind,$resource,$org,$env,$version,$config,1,$now,$now);"),
    sql.Named("id",id),sql.Named("kind",string(resource.Kind)),sql.Named("resource",resource.ID),sql.Named("org",r.OrganizationID),sql.Named("env",envID),sql.Named("version",version),sql.Named("config",config),sql.Named("now",now))
  }
  if err!=nil{return err}
  return c.event(ctx,tx,r,"m8.resourcemanager.environment_binding.changed.v1")
 })
 return out,err
}
func(c *Catalog) RemoveBinding(ctx context.Context,id string,rev int64)error{
 if id==""||rev<=0{return v1engine.ErrInvalid}
 return c.transaction(ctx,func(tx *sql.Tx)error{
  var rk,ri string
  var revision int64
  err:=tx.QueryRowContext(ctx,c.stmt("DECLARE $id AS Utf8;SELECT resource_kind,resource_id,revision FROM `m8-resource-manager/resource_environment_bindings` WHERE id=$id;"),sql.Named("id",id)).Scan(&rk,&ri,&revision)
  if errors.Is(err,sql.ErrNoRows){return v1engine.ErrNotFound}
  if err!=nil{return err}
  if revision!=rev{return v1engine.ErrConflict}
  r,err:=c.get(ctx,tx,v1engine.Kind(rk),ri);if err!=nil{return err}
  _,err=tx.ExecContext(ctx,c.stmt("DECLARE $id AS Utf8;DELETE FROM `m8-resource-manager/resource_environment_bindings` WHERE id=$id;"),sql.Named("id",id))
  if err!=nil{return err}
  return c.event(ctx,tx,r,"m8.resourcemanager.environment_binding.removed.v1")
 })
}
func(c *Catalog) ListBindings(ctx context.Context,resource v1engine.Owner,environmentID string)([]v1engine.Binding,error){
 if resource.ID==""&&environmentID==""{return nil,v1engine.ErrInvalid}
 var query string
 args:=[]any{}
 if resource.ID!=""{
  query="DECLARE $kind AS Utf8;DECLARE $resource AS Utf8;SELECT id,resource_kind,resource_id,environment_id,version_ref,configuration_ref,revision FROM `m8-resource-manager/resource_environment_bindings` WHERE resource_kind=$kind AND resource_id=$resource;"
  args=append(args,sql.Named("kind",string(resource.Kind)),sql.Named("resource",resource.ID))
 }else{
  query="DECLARE $env AS Utf8;SELECT id,resource_kind,resource_id,environment_id,version_ref,configuration_ref,revision FROM `m8-resource-manager/resource_environment_bindings` WHERE environment_id=$env;"
  args=append(args,sql.Named("env",environmentID))
 }
 rows,err:=c.store.db.QueryContext(ctx,c.stmt(query),args...)
 if err!=nil{return nil,fmt.Errorf("list bindings: %w",err)}
 defer rows.Close()
 out:=[]v1engine.Binding{}
 for rows.Next(){
  var b v1engine.Binding
  var kind string
  if err=rows.Scan(&b.ID,&kind,&b.Resource.ID,&b.EnvironmentID,&b.Version,&b.ConfigRef,&b.Revision);err!=nil{return nil,err}
  b.Resource.Kind=v1engine.Kind(kind)
  if environmentID==""||b.EnvironmentID==environmentID{out=append(out,b)}
 }
 sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID})
 return out,rows.Err()
}

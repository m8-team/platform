package ydbcatalog

import (
 "context"
 "database/sql"
 "errors"

 "github.com/m8-team/platform/internal/access/app/catalog"
 "github.com/m8-team/platform/internal/access/domain/shared"
)
const listIndex = "DECLARE $module AS Utf8;DECLARE $after AS Int64;DECLARE $snapshot AS Int64;DECLARE $limit AS Uint64;\nSELECT revision,version,manifest_json FROM \x60m8-access/catalog_read_index\x60 WHERE module_id=$module AND revision>$after AND revision<=$snapshot ORDER BY revision LIMIT $limit;"

// List is the CQRS read path. Its first page pins the catalog head, and
// subsequent pages use a signed principal-bound cursor with that same head.
func(s *Store) List(ctx context.Context,principal shared.Subject,moduleID string,pageSize int,pageToken string)([]catalog.PublishedRecord,string,error){
 if s==nil||s.db==nil{return nil,"",ErrInvalidStore}
 if principal.Validate()!=nil||!shared.ValidModuleName(moduleID)||pageSize<1||pageSize>100||len(pageToken)>2048{
  return nil,"",catalog.ErrCatalogArgument
 }
 if err:=ctx.Err();err!=nil{return nil,"",err}
 var after,snapshot int64
 if pageToken!=""{
  parsed,err:=s.decodeCursor(pageToken,principal,moduleID)
  if err!=nil{return nil,"",catalog.ErrCatalogArgument}
  after=parsed.After
  snapshot=parsed.Snapshot
 }else{
  err:=s.db.QueryRowContext(ctx,s.q(selectHead),sql.Named("module",moduleID)).Scan(&snapshot)
  if errors.Is(err,sql.ErrNoRows){return []catalog.PublishedRecord{},"",nil}
  if err!=nil{return nil,"",err}
  if snapshot<1{return nil,"",catalog.ErrCatalogStoreInvalid}
 }
 rows,err:=s.db.QueryContext(ctx,s.q(listIndex),sql.Named("module",moduleID),
  sql.Named("after",after),sql.Named("snapshot",snapshot),sql.Named("limit",uint64(pageSize+1)))
 if err!=nil{return nil,"",err}
 defer rows.Close()
 records:=make([]catalog.PublishedRecord,0,pageSize)
 var last int64
 hasNext:=false
 for rows.Next(){
  var revision int64
  var version,jsonText string
  if err:=rows.Scan(&revision,&version,&jsonText);err!=nil{return nil,"",err}
  if revision<=after||revision>snapshot||!shared.ValidReleaseVersion(version){return nil,"",catalog.ErrCatalogStoreInvalid}
  if len(records)==pageSize{hasNext=true;break}
  manifest,err:=unmarshalManifest([]byte(jsonText))
  if err!=nil||manifest.ModuleID()!=moduleID||manifest.Version()!=version{return nil,"",catalog.ErrCatalogStoreInvalid}
  record:=catalog.PublishedRecord{Manifest:manifest,Revision:revision}
  if record.Validate()!=nil||revision<=last{return nil,"",catalog.ErrCatalogStoreInvalid}
  records=append(records,record)
  last=revision
 }
 if err:=rows.Err();err!=nil{return nil,"",err}
 if !hasNext{return records,"",nil}
 next,err:=s.encodeCursor(principal,moduleID,last,snapshot)
 if err!=nil{return nil,"",err}
 return records,next,nil
}

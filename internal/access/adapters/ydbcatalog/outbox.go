package ydbcatalog

import (
 "context"
 "database/sql"
 "errors"
 "time"

 "github.com/m8-team/platform/internal/access/app/catalog"
 "github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidOutboxEvent = errors.New("invalid Access Catalog outbox event")

// Event is an immutable fact. Emission to Kafka is at-least-once;
// downstream consumers must deduplicate by EventID and source revision.
type Event struct {
 ModuleID string
 Revision int64
 EventID string
 Kind string
 ManifestJSON string
 CreatedAt time.Time
}
func(e Event) Validate()error{
 if !shared.ValidModuleName(e.ModuleID)||e.Revision<1||e.EventID==""||
  e.Kind!="ModuleManifestPublished"||e.CreatedAt.IsZero(){return ErrInvalidOutboxEvent}
 manifest,err:=unmarshalManifest([]byte(e.ManifestJSON))
 if err!=nil||manifest.ModuleID()!=e.ModuleID{return ErrInvalidOutboxEvent}
 return nil
}

const pendingEvents = "DECLARE $module AS Utf8;DECLARE $after AS Int64;DECLARE $limit AS Uint64;\nSELECT revision,event_id,event_kind,payload_json,created_at FROM \x60m8-access/catalog_outbox\x60 WHERE module_id=$module AND revision>$after AND published=FALSE ORDER BY revision LIMIT $limit;"
const selectEvent = "DECLARE $module AS Utf8;DECLARE $revision AS Int64;\nSELECT event_id,published FROM \x60m8-access/catalog_outbox\x60 WHERE module_id=$module AND revision=$revision;"

// Pending returns durable, ordered events for one verified module namespace.
// There is no implicit broker ACK here: ACK is allowed only after publishing.
func(s *Store) Pending(ctx context.Context,moduleID string,after int64,limit int)([]Event,error){
 if s==nil||s.db==nil{return nil,ErrInvalidStore}
 if !shared.ValidModuleName(moduleID)||after<0||limit<1||limit>100{return nil,catalog.ErrCatalogArgument}
 rows,err:=s.db.QueryContext(ctx,s.q(pendingEvents),sql.Named("module",moduleID),
  sql.Named("after",after),sql.Named("limit",uint64(limit)))
 if err!=nil{return nil,err}
 defer rows.Close()
 events:=make([]Event,0,limit)
 for rows.Next(){
  var e Event
  e.ModuleID=moduleID
  if err:=rows.Scan(&e.Revision,&e.EventID,&e.Kind,&e.ManifestJSON,&e.CreatedAt);err!=nil{return nil,err}
  if err:=e.Validate();err!=nil{return nil,err}
  events=append(events,e)
 }
 if err:=rows.Err();err!=nil{return nil,err}
 return events,nil
}
// MarkPublished may only be invoked by an authorized outbox dispatcher after
// the external broker confirms delivery. Repeated ACKs are harmless.
func(s *Store) MarkPublished(ctx context.Context,moduleID string,revision int64,eventID string)error{
 if s==nil||s.db==nil{return ErrInvalidStore}
 if !shared.ValidModuleName(moduleID)||revision<1||eventID==""{return catalog.ErrCatalogArgument}
 return s.tx(ctx,func(tx *sql.Tx)error{
  var observed string
  var published bool
  err:=tx.QueryRowContext(ctx,s.q(selectEvent),sql.Named("module",moduleID),sql.Named("revision",revision)).
   Scan(&observed,&published)
  if errors.Is(err,sql.ErrNoRows){return catalog.ErrCatalogNotFound}
  if err!=nil{return err}
  if observed!=eventID{return ErrInvalidOutboxEvent}
  if published{return nil}
  _,err=tx.ExecContext(ctx,s.q("DECLARE $module AS Utf8;DECLARE $revision AS Int64;DECLARE $published AS Bool;\nUPDATE \x60m8-access/catalog_outbox\x60 SET published=$published WHERE module_id=$module AND revision=$revision;"),
   sql.Named("module",moduleID),sql.Named("revision",revision),sql.Named("published",true))
  return err
 })
}

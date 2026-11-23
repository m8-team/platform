//go:build integration

package ydbcatalog

import (
 "context"
 "database/sql"
 "errors"
 "fmt"
 "os"
 "path/filepath"
 "strings"
 "sync"
 "testing"
 "time"

 "github.com/m8-team/platform/internal/access/app/catalog"
 "github.com/m8-team/platform/internal/access/domain/module"
 "github.com/m8-team/platform/internal/access/domain/shared"
 ydbsdk "github.com/ydb-platform/ydb-go-sdk/v3"
)

func testStore(t *testing.T)*Store{
 t.Helper()
 if os.Getenv("M8_ACCESS_LIVE_YDB")!="1"{t.Skip("requires disposable YDB")}
 dsn:=os.Getenv("M8_ACCESS_TEST_YDB_DSN")
 base:=os.Getenv("M8_ACCESS_TEST_YDB_PREFIX")
 if !strings.HasPrefix(base,"/local/m8_access_ci")||!strings.Contains(dsn,"localhost:2136/local"){
  t.Fatal("refuse to test outside disposable local YDB")
 }
 suffix:=strings.NewReplacer("/","_"," ","_").Replace(strings.ToLower(t.Name()))
 prefix:=base+"_catalog_"+suffix
 ctx,cancel:=context.WithTimeout(context.Background(),120*time.Second)
 t.Cleanup(cancel)
 db,err:=sql.Open("ydb",dsn)
 if err!=nil{t.Fatal(err)}
 t.Cleanup(func(){_ =db.Close()})
 for attempt:=0;attempt<40;attempt++{
  probe,done:=context.WithTimeout(ctx,3*time.Second)
  err=db.PingContext(probe)
  done()
  if err==nil{break}
  if ctx.Err()!=nil{break}
  time.Sleep(time.Second)
 }
 if err!=nil{t.Fatal(err)}
 sdk,err:=ydbsdk.Open(ctx,dsn)
 if err!=nil{t.Fatal(err)}
 defer sdk.Close(ctx)
 for _,path:=range []string{base,prefix,prefix+"/m8-access"}{
  if err:=sdk.Scheme().MakeDirectory(ctx,path);err!=nil{
   if _,exists:=sdk.Scheme().ListDirectory(ctx,path);exists!=nil{t.Fatalf("YDB directory %s: %v",path,err)}
  }
 }
 sqlBytes,err:=os.ReadFile(filepath.Join("schema.sql"))
 if err!=nil{t.Fatal(err)}
 for _,fragment:=range strings.Split(string(sqlBytes),";"){
  var lines []string
  for _,line:=range strings.Split(fragment,"\n"){
   line=strings.TrimSpace(line)
   if line!=""&&!strings.HasPrefix(line,"--"){lines=append(lines,line)}
  }
  if len(lines)==0{continue}
  stmt:=strings.Join(lines,"\n")+";"
  _,err=db.ExecContext(ctx,fmt.Sprintf("PRAGMA TablePathPrefix(%q);\n%s",prefix,stmt))
  if err!=nil{t.Fatalf("catalog migration failed: %v; statement %s",err,stmt)}
 }
 store,err:=New(db,prefix,[]byte(strings.Repeat("test-cursor-secret-",4)),time.Hour)
 if err!=nil{t.Fatal(err)}
 return store
}
func TestLiveYDBCatalogPublishReadIndexIdempotencyAndOutbox(t *testing.T){
 store:=testStore(t)
 ctx,cancel:=context.WithTimeout(context.Background(),60*time.Second)
 defer cancel()
 first:=fixture(t,"1.0.0")
 record,err:=store.Publish(ctx,first,0,"key-1")
 if err!=nil||record.Revision!=1{t.Fatalf("first publish: %+v %v",record,err)}
 retry,err:=store.Publish(ctx,first,0,"key-1")
 if err!=nil||retry.Revision!=1{t.Fatalf("idempotency replay: %+v %v",retry,err)}
 second:=fixture(t,"1.0.1")
 if _,err:=store.Publish(ctx,second,0,"key-2");!errors.Is(err,shared.ErrRevisionConflict){t.Fatalf("stale head accepted: %v",err)}
 if _,err:=store.Publish(ctx,second,1,"key-1");!errors.Is(err,catalog.ErrCatalogConflict){t.Fatalf("key reuse accepted: %v",err)}
 if _,err:=store.Publish(ctx,first,1,"key-new");!errors.Is(err,catalog.ErrCatalogConflict){t.Fatalf("immutable release overwritten: %v",err)}
 created,err:=store.Publish(ctx,second,1,"key-2")
 if err!=nil||created.Revision!=2{t.Fatalf("second version publish: %v",err)}
 actor:=first.Publisher()
 page,next,err:=store.List(ctx,actor,"finance",1,"")
 if err!=nil||len(page)!=1||page[0].Revision!=1||next==""{t.Fatalf("first page: %v %+v %q",err,page,next)}
 page2,last,err:=store.List(ctx,actor,"finance",1,next)
 if err!=nil||len(page2)!=1||page2[0].Revision!=2||last!=""{t.Fatalf("second page: %v %+v",err,page2)}
 other,_:=shared.NewSubject("service_account","attacker")
 if _,_,err:=store.List(ctx,other,"finance",1,next);!errors.Is(err,catalog.ErrCatalogArgument){t.Fatalf("reused another principal cursor: %v",err)}
 loaded,err:=store.Get(ctx,"finance","1.0.0")
 if err!=nil||loaded.Revision!=1{t.Fatalf("missing immutable release: %v",err)}
 events,err:=store.Pending(ctx,"finance",0,10)
 if err!=nil||len(events)!=2||events[0].Revision!=1||events[1].Revision!=2{t.Fatalf("outbox: %v %+v",err,events)}
 if err:=store.MarkPublished(ctx,"finance",1,"wrong");!errors.Is(err,ErrInvalidOutboxEvent){t.Fatal(err)}
 if err:=store.MarkPublished(ctx,"finance",1,events[0].EventID);err!=nil{t.Fatal(err)}
 if err:=store.MarkPublished(ctx,"finance",1,events[0].EventID);err!=nil{t.Fatal(err)}
 pending,err:=store.Pending(ctx,"finance",0,10)
 if err!=nil||len(pending)!=1||pending[0].Revision!=2{t.Fatalf("outbox ack: %v %+v",err,pending)}
}
func TestLiveYDBCatalogConcurrentPublishHasSingleWinner(t *testing.T){
 store:=testStore(t)
 ctx,cancel:=context.WithTimeout(context.Background(),60*time.Second)
 defer cancel()
 if _,err:=store.Publish(ctx,fixture(t,"1.0.0"),0,"seed");err!=nil{t.Fatal(err)}
 manifests:=[]module.Manifest{fixture(t,"1.0.1"),fixture(t,"1.0.2")}
 var wg sync.WaitGroup
 results:=make(chan error,2)
 for i,manifest:=range manifests{
  version:=[]string{"1.0.1","1.0.2"}[i]
  wg.Add(1)
  go func(m module.Manifest,v string){
   defer wg.Done()
   _,err:=store.Publish(ctx,m,1,"key-"+v)
   results<-err
  }(manifest,version)
 }
 wg.Wait();close(results)
 successes:=0
 for err:=range results {if err==nil{successes++}}
 if successes!=1{t.Fatalf("concurrent CAS winners: %d (expected exactly one)",successes)}
 actor,_:=shared.NewSubject("service_account","finance")
 list,next,err:=store.List(ctx,actor,"finance",10,"")
 if err!=nil||len(list)!=2||next!=""{t.Fatalf("read-side atomic snapshot: %v %+v",err,list)}
 pending,err:=store.Pending(ctx,"finance",0,10)
 if err!=nil||len(pending)!=2{t.Fatalf("partial outbox state: %v %+v",err,pending)}
}

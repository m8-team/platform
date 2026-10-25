package main

import (
 "net/http"
 "net/http/httptest"
 "strings"
 "testing"

 pb "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
 grpcserver "github.com/m8-team/platform/internal/platform/server/grpc"
 "go.uber.org/fx"
 "google.golang.org/grpc"
 "google.golang.org/protobuf/encoding/protojson"
)

func appTestConfig(allow bool)Config{
 return Config{HTTP:HTTPConfig{Address:"127.0.0.1:0"},HealthHTTP:HealthHTTPConfig{Address:"127.0.0.1:0"},GRPC:grpcserver.Config{Address:"127.0.0.1:0"},AllowUnauthenticated:allow}
}
func TestNewAppBuilds(t *testing.T){
 app:=NewApp(appTestConfig(false));if err:=app.Err();err!=nil{t.Fatal(err)}
}
func TestRegistersCanonicalResourceManagerService(t *testing.T){
 var server *grpc.Server
 app:=fx.New(append(appOptions(appTestConfig(false)),fx.Populate(&server))...)
 if err:=app.Err();err!=nil{t.Fatal(err)}
 if _,found:=server.GetServiceInfo()[pb.ResourceManagerService_ServiceDesc.ServiceName];!found{t.Fatal("canonical ResourceManagerService is not registered")}
}
func buildHTTPHandler(t *testing.T,allow bool)http.Handler{
 t.Helper();var wrapped resourceManagerHTTPHandler
 app:=fx.New(append(appOptions(appTestConfig(allow)),fx.Populate(&wrapped))...)
 if err:=app.Err();err!=nil{t.Fatal(err)}
 if wrapped.Handler==nil{t.Fatal("nil gateway")}
 return wrapped.Handler
}
func TestCanonicalV1GatewayCreatesTenant(t *testing.T){
 handler:=buildHTTPHandler(t,true)
 req:=httptest.NewRequest(http.MethodPost,"/resource-manager/v1/resources",strings.NewReader(`{"resource":{"tenant":{"metadata":{"name":"Demo"}}},"idempotencyKey":"create-demo-1"}`))
 req.Header.Set("Content-Type","application/json");response:=httptest.NewRecorder()
 handler.ServeHTTP(response,req)
 if response.Code!=http.StatusOK{t.Fatalf("status=%d body=%s",response.Code,response.Body.String())}
 var got pb.ResourceRecord;if err:=protojson.Unmarshal(response.Body.Bytes(),&got);err!=nil{t.Fatal(err)}
 if got.GetTenant()==nil||got.GetTenant().GetMetadata().GetId()==""{t.Fatalf("response=%s",response.Body.String())}
}
func TestCanonicalV1GatewayDeniesAccessByDefault(t *testing.T){
 handler:=buildHTTPHandler(t,false);req:=httptest.NewRequest(http.MethodPost,"/resource-manager/v1/resources",strings.NewReader(`{"resource":{"tenant":{"metadata":{"name":"Denied"}}},"idempotencyKey":"denied"}`))
 req.Header.Set("Content-Type","application/json");response:=httptest.NewRecorder()
 handler.ServeHTTP(response,req)
 if response.Code!=http.StatusForbidden{t.Fatalf("status=%d body=%s",response.Code,response.Body.String())}
}
func TestResourceGatewayDoesNotExposeHealth(t *testing.T){
 req:=httptest.NewRequest("GET","/livez",nil);response:=httptest.NewRecorder();buildHTTPHandler(t,false).ServeHTTP(response,req)
 if response.Code!=http.StatusNotFound{t.Fatalf("status=%d",response.Code)}
}
func TestHealthRoutesRemainSeparate(t *testing.T){
 var wrapped healthHTTPHandler
 app:=fx.New(append(appOptions(appTestConfig(false)),fx.Populate(&wrapped))...)
 if err:=app.Err();err!=nil{t.Fatal(err)}
 req:=httptest.NewRequest("GET","/livez",nil);response:=httptest.NewRecorder();wrapped.ServeHTTP(response,req)
 if response.Code!=http.StatusOK{t.Fatalf("status=%d",response.Code)}
}

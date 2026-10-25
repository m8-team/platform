package v1engine
import("net/http/httptest";"strings";"testing")
func TestHTTPAuthenticationAndCreate(t *testing.T){
 h:=NewHandler(NewStore(),"local-secret");path:="/resource-manager/v1/resources";payload:=`{"resource":{"tenant":{"metadata":{"name":"Demo"}}},"idempotencyKey":"req1"}`
 req:=httptest.NewRequest("POST",path,strings.NewReader(payload));rr:=httptest.NewRecorder();h.ServeHTTP(rr,req);if rr.Code!=401{t.Fatalf("no auth: %d",rr.Code)}
 req=httptest.NewRequest("POST",path,strings.NewReader(payload));req.Header.Set("Authorization","Bearer local-secret");rr=httptest.NewRecorder();h.ServeHTTP(rr,req);if rr.Code!=201{t.Fatalf("create: %d %s",rr.Code,rr.Body.String())};if !strings.Contains(rr.Body.String(),"Demo"){t.Fatal(rr.Body.String())}
}
func TestHTTPRejectUnknownJSON(t *testing.T){
 h:=NewHandler(NewStore(),"secret");req:=httptest.NewRequest("POST","/resource-manager/v1/resources",strings.NewReader(`{"unknown":true}`));req.Header.Set("Authorization","Bearer secret");rr:=httptest.NewRecorder();h.ServeHTTP(rr,req);if rr.Code!=400{t.Fatalf("%d %s",rr.Code,rr.Body.String())}
}

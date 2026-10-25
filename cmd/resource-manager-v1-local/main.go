// Resource Manager canonical v1 local API development runner.
// This is NOT a production endpoint: data is in memory and IAM is not integrated.
package main

import (
 "fmt"
 "log"
 "net"
 "net/http"
 "os"
 "strings"

 "github.com/m8-team/platform/internal/resourcemanager/v1engine"
)

func main() {
 token:=strings.TrimSpace(os.Getenv("M8_RM_DEV_TOKEN"))
 if len(token)<32 {log.Fatal("M8_RM_DEV_TOKEN must contain at least 32 characters; development only")}
 addr:="127.0.0.1:18080"
 if host,_,err:=net.SplitHostPort(addr);err!=nil||host!="127.0.0.1"{log.Fatal("local runner must bind loopback")}
 handler:=v1engine.NewHandler(v1engine.NewStore(),token)
 fmt.Println("M8 Resource Manager v1 local HTTP API:",addr)
 log.Fatal(http.ListenAndServe(addr,handler))
}

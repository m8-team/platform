package ydbcatalog

import (
 "bytes"
 "encoding/base64"
 "fmt"
 "go/format"
 "os"
 "path/filepath"
 "strings"
 "testing"
)
func TestTemporaryGofmtExport(t *testing.T) {
 paths,err:=filepath.Glob("*.go")
 if err!=nil{t.Fatal(err)}
 total:=0
 for _,path:=range paths {
  if path=="gofmt_export_test.go"{continue}
  body,err:=os.ReadFile(path)
  if err!=nil{t.Fatal(err)}
  formatted,err:=format.Source(body)
  if err!=nil{t.Fatalf("%s: %v",path,err)}
  if bytes.Equal(body,formatted){continue}
  encoded:=base64.StdEncoding.EncodeToString(formatted)
  chunks:=(len(encoded)+2999)/3000
  for i:=0;i<chunks;i++{
   end:=(i+1)*3000
   if end>len(encoded){end=len(encoded)}
   fmt.Printf("M8FMT|internal/access/adapters/ydbcatalog/%s|%d|%d|%s\n",strings.TrimSpace(path),i,chunks,encoded[i*3000:end])
  }
  total++
 }
 if total>0{t.Fatalf("%d Go files require gofmt",total)}
}

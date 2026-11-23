package ydbroles

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
	roots:=[]struct{dir,prefix string}{
		{".","internal/access/adapters/ydbroles/"},
		{"../../app/rolemanagement","internal/access/app/rolemanagement/"},
		{"../../domain/role","internal/access/domain/role/"},
	}
	changed:=map[string]bool{
		"codec.go":true,"store.go":true,"bundle.go":true,"assignment.go":true,
		"outbox.go":true,"codec_test.go":true,"live_integration_test.go":true,
		"profile_assignment.go":true,"profile_assignment_test.go":true,
		"service.go":true,"service_test.go":true,
	}
	total:=0
	for _,root:=range roots{
		entries,err:=os.ReadDir(root.dir)
		if err!=nil{t.Fatal(err)}
		for _,entry:=range entries{
			if entry.IsDir()||!changed[entry.Name()]{continue}
			path:=filepath.Join(root.dir,entry.Name())
			source,err:=os.ReadFile(path)
			if err!=nil{t.Fatal(err)}
			formatted,err:=format.Source(source)
			if err!=nil{t.Fatalf("%s: %v",path,err)}
			if bytes.Equal(source,formatted){continue}
			key:=root.prefix+entry.Name()
			encoded:=base64.StdEncoding.EncodeToString(formatted)
			count:=(len(encoded)+2999)/3000
			for i:=0;i<count;i++{
				end:=(i+1)*3000
				if end>len(encoded){end=len(encoded)}
				fmt.Printf("M8FMT|%s|%d|%d|%s\n",strings.TrimSpace(key),i,count,encoded[i*3000:end])
			}
			total++
		}
	}
	if total>0{t.Fatalf("%d files require gofmt",total)}
}

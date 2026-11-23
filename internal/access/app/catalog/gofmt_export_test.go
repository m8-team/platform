package catalog

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemporaryGofmtExport(t *testing.T) {
	total:=0
	for _,root:=range []struct{dir,prefix string}{
		{".","internal/access/app/catalog/"},
		{"../../adapters/cataloggrpc","internal/access/adapters/cataloggrpc/"},
	}{
		err:=filepath.WalkDir(root.dir,func(path string,entry fs.DirEntry,walkErr error)error{
			if walkErr!=nil{return walkErr}
			if entry.IsDir()||!strings.HasSuffix(path,".go")||entry.Name()=="gofmt_export_test.go"{return nil}
			body,err:=os.ReadFile(path)
			if err!=nil{return err}
			formatted,err:=format.Source(body)
			if err!=nil{return fmt.Errorf("%s: %w",path,err)}
			if bytes.Equal(body,formatted){return nil}
			key:=root.prefix+entry.Name()
			b64:=base64.StdEncoding.EncodeToString(formatted)
			chunks:=(len(b64)+2999)/3000
			for i:=0;i<chunks;i++ {
				end:=(i+1)*3000
				if end>len(b64){end=len(b64)}
				fmt.Printf("M8FMT|%s|%d|%d|%s\n",key,i,chunks,b64[i*3000:end])
			}
			total++
			return nil
		})
		if err!=nil{t.Fatal(err)}
	}
	if total>0 {t.Fatalf("%d files require gofmt",total)}
}

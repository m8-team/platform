package permission

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

func TestTemporaryExportGofmt(t *testing.T) {
	count:=0
	for _,root:=range []string{"..","../../app/catalog"}{
		err:=filepath.WalkDir(root,func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr!=nil{return walkErr}
			if entry.IsDir()||!strings.HasSuffix(path,".go")||filepath.Base(path)=="gofmt_dump_test.go"{return nil}
			raw,err:=os.ReadFile(path);if err!=nil{return err}
			formatted,err:=format.Source(raw);if err!=nil{return fmt.Errorf("%s: %w",path,err)}
			if bytes.Equal(raw,formatted){return nil}
			rel,err:=filepath.Rel("../..",path);if err!=nil{return err}
			fmt.Printf("M8FMT|%s|%s\n",filepath.ToSlash(rel),base64.StdEncoding.EncodeToString(formatted))
			count++
			return nil
		})
		if err!=nil{t.Fatal(err)}
	}
	t.Fatalf("temporary gofmt export: %d changed files",count)
}

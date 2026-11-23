package ydbcatalog

import (
 "crypto/hmac"
 "crypto/sha256"
 "encoding/base64"
 "encoding/json"
 "errors"
 "strings"

 "github.com/m8-team/platform/internal/access/app/catalog"
 "github.com/m8-team/platform/internal/access/domain/shared"
)

var ErrInvalidCursor = errors.New("invalid or expired catalog cursor")

// Cursor is authenticated, principal-bound, module-bound and snapshot-pinned.
// It deliberately contains no tenant data because the module catalog is global.
type cursor struct{
 Module string `json:"m"`
 SubjectType string `json:"t"`
 SubjectID string `json:"s"`
 After int64 `json:"a"`
 Snapshot int64 `json:"r"`
 ExpiresAt int64 `json:"e"`
}
func(s *Store) sign(body []byte)[]byte{
 mac:=hmac.New(sha256.New,s.tokenKey)
 _,_=mac.Write(body)
 return mac.Sum(nil)
}
func(s *Store) encodeCursor(p shared.Subject,moduleID string,after,snapshot int64)(string,error){
 if after<1||snapshot<=after||p.Validate()!=nil||!shared.ValidModuleName(moduleID){return "",ErrInvalidCursor}
 payload,err:=json.Marshal(cursor{
  Module:moduleID,SubjectType:p.Type(),SubjectID:p.ID(),After:after,Snapshot:snapshot,
  ExpiresAt:s.now().Add(s.cursorTTL).Unix(),
 })
 if err!=nil{return "",ErrInvalidCursor}
 text:=base64.RawURLEncoding.EncodeToString(payload)+"."+base64.RawURLEncoding.EncodeToString(s.sign(payload))
 if len(text)>2048{return "",ErrInvalidCursor}
 return text,nil
}
func(s *Store) decodeCursor(token string,p shared.Subject,moduleID string)(cursor,error){
 var empty cursor
 if token==""||len(token)>2048||p.Validate()!=nil||!shared.ValidModuleName(moduleID){return empty,catalog.ErrCatalogArgument}
 chunks:=strings.Split(token,".")
 if len(chunks)!=2{return empty,ErrInvalidCursor}
 body,err:=base64.RawURLEncoding.DecodeString(chunks[0])
 if err!=nil||len(body)==0||len(body)>1500{return empty,ErrInvalidCursor}
 signature,err:=base64.RawURLEncoding.DecodeString(chunks[1])
 if err!=nil||!hmac.Equal(signature,s.sign(body)){return empty,ErrInvalidCursor}
 var value cursor
 if err:=json.Unmarshal(body,&value);err!=nil{return empty,ErrInvalidCursor}
 if value.Module!=moduleID||value.SubjectType!=p.Type()||value.SubjectID!=p.ID()||
  value.After<1||value.Snapshot<=value.After||value.ExpiresAt<=s.now().Unix()||
  value.ExpiresAt>s.now().Add(s.cursorTTL).Unix(){return empty,ErrInvalidCursor}
 return value,nil
}

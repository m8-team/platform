package contracttest

import (
	"testing"

	rm "github.com/m8-team/go-genproto/m8/platform/resourcemanager/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestCreateOrganizationRejectsServerOwnedFields(t *testing.T) {
	for _, field := range []string{`"id":"abc"`, `"state":"ACTIVE"`, `"version":1`, `"createTime":"2026-09-05T00:00:00Z"`} {
		t.Run(field, func(t *testing.T) {
			var request rm.CreateOrganizationRequest
			if err := protojson.Unmarshal([]byte(`{"organization":{`+field+`}}`), &request); err == nil {
				t.Fatal("server-owned field accepted")
			}
		})
	}
	var request rm.CreateOrganizationRequest
	if err := protojson.Unmarshal([]byte(`{"organization":{"name":"Acme","labels":{"env":"prod"}}}`), &request); err != nil {
		t.Fatal(err)
	}
}

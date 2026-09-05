package contracttest

import (
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	iam "github.com/m8-team/go-genproto/m8/platform/iam/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Exercise the declared CEL guards against the generated wire models, so input
// hints cannot accidentally be treated as authoritative context or echoed back.
func TestAuthenticationTrustBoundaries(t *testing.T) {
	for _, tt := range []struct {
		name    string
		message proto.Message
		valid   bool
	}{
		{"empty request", &iam.CreateRequest{}, true},
		{"input hints", &iam.CreateRequest{StartContext: &iam.AuthenticationStartContext{}}, true},
		{"resolved context", &iam.CreateRequest{Context: &iam.AuthenticationContext{}}, false},
		{"both contexts", &iam.CreateRequest{Context: &iam.AuthenticationContext{}, StartContext: &iam.AuthenticationStartContext{}}, false},
		{"anonymous snapshot", &iam.Authentication{}, true},
		{"masked snapshot", &iam.Authentication{PublicSubject: &iam.AuthenticationPublicSubject{MaskedIdentifier: "a***@example.com"}}, true},
		{"raw email snapshot", &iam.Authentication{Subject: &iam.AuthenticationSubject{Identifier: &iam.AuthenticationSubject_Email{Email: "alice@example.com"}}}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.message.ProtoReflect().Descriptor()
			rules := proto.GetExtension(d.Options(), validate.E_Message).(*validate.MessageRules)
			if len(rules.GetCel()) == 0 {
				t.Fatal("missing trust boundary rule")
			}
			env, err := cel.NewEnv(cel.Types(tt.message), cel.Variable("this", cel.ObjectType(string(d.FullName()))))
			if err != nil {
				t.Fatal(err)
			}
			for _, rule := range rules.GetCel() {
				ast, issues := env.Compile(rule.GetExpression())
				if issues.Err() != nil {
					t.Fatal(issues.Err())
				}
				program, err := env.Program(ast)
				if err != nil {
					t.Fatal(err)
				}
				got, _, err := program.Eval(map[string]any{"this": tt.message})
				if err != nil {
					t.Fatal(err)
				}
				if (got == types.True) != tt.valid {
					t.Fatalf("%s: got %v, want %v", rule.GetId(), got, tt.valid)
				}
			}
		})
	}
}

func TestAuthenticationStartJSON(t *testing.T) {
	var request iam.CreateRequest
	err := protojson.Unmarshal([]byte(`{"startContext":{"oidc":{"nonce":"nonce","codeChallenge":"challenge","codeChallengeMethod":"S256"}},"options":{"methodId":"password"}}`), &request)
	if err != nil {
		t.Fatal(err)
	}
	if request.GetStartContext().GetOidc().GetCodeChallenge() != "challenge" || request.GetContext() != nil {
		t.Fatal("OIDC input did not reach the untrusted start context")
	}
	field := request.GetOptions().ProtoReflect().Descriptor().Fields().ByName("provider_id")
	rules := proto.GetExtension(field.Options(), validate.E_Field).(*validate.FieldRules)
	if request.GetOptions().GetProviderId() != "" || rules.GetIgnore() != validate.Ignore_IGNORE_IF_ZERO_VALUE || !rules.GetString().GetUuid() {
		t.Fatal("provider must allow omission while validating supplied UUIDs")
	}
}

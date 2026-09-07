package connect

import (
	rpc "connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/m8-team/platform-sdk/fault"
	"github.com/m8-team/platform-sdk/logging"
	"github.com/m8-team/platform-sdk/telemetry"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"log/slog"
	"net/http"
	"testing"
)

type auth struct{}

func (auth) Authenticate(context.Context, http.Header) (Principal, error) {
	return Principal{SubjectID: "subject", TenantID: "tenant"}, nil
}
func (auth) Check(context.Context, CheckRequest) error { return nil }
func TestPanicAndMapping(t *testing.T) {
	p, err := telemetry.New(context.Background(), logging.Info{Name: "test"}, telemetry.Config{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := p.Shutdown(context.Background()); err != nil {
			t.Error(err)
		}
	})
	interceptor, err := New(Config{Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Telemetry: p, Authentication: auth{}, Authorization: auth{}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = interceptor.WrapUnary(func(context.Context, rpc.AnyRequest) (rpc.AnyResponse, error) { panic("secret") })(context.Background(), rpc.NewRequest(&emptypb.Empty{}))
	if rpc.CodeOf(err) != rpc.CodeInternal || err.Error() != "internal: internal error" {
		t.Fatal(err)
	}
	cause := errors.New("database password=secret")
	mapped := ToError(fmt.Errorf("adapter: %w", fault.New(fault.NotFound, "product not found", cause)))
	if mapped.Code() != rpc.CodeNotFound || mapped.Message() != "product not found" {
		t.Fatal(mapped)
	}
	if ToError(cause).Message() != "internal error" {
		t.Fatal("internal detail leaked")
	}
}

package wiring

import (
	"context"
	"errors"
	"testing"

	"github.com/m8-team/platform/internal/access/adapters/authzen"
)

func TestProductionUniversalRuntimeDoesNotAllowPartialWiring(t *testing.T) {
	_,err:=NewProductionUniversalPDP(context.Background(),nil,nil,nil,nil,nil,nil,nil,nil,nil,"model-v1")
	if !errors.Is(err,authzen.ErrInvalidConfiguration){
		t.Fatalf("partially wired Universal PDP was accepted: %v",err)
	}
}

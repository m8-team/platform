package authzen

import (
    "context"
    "net/http"

    accesshttp "github.com/m8-team/platform/internal/access/adapters/http"
)

// NewHTTPHandler explicitly wires the real AuthZEN HTTP route to the
// consistency-checked implementation. TLS and trusted identity middleware
// must wrap the returned handler before it is exposed outside a service.
func NewHTTPHandler(ctx context.Context,service *Service)(http.Handler,error){
    if service==nil{return nil,ErrInvalidConfiguration}
    return accesshttp.NewHandler(ctx,service)
}

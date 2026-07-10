package examples

import "context"

type AuthenticationClient interface {
	StartAuthentication(context.Context, StartAuthenticationRequest) (Operation, error)
}

type StartAuthenticationRequest struct {
	ClientID       string
	SubjectHint    string
	Reason         string
	RequestedAAL   string
	IdempotencyKey string
}

type Operation struct{ Name string }

func StartReauthentication(ctx context.Context, client AuthenticationClient) (Operation, error) {
	return client.StartAuthentication(ctx, StartAuthenticationRequest{
		ClientID:       "clients/example",
		SubjectHint:    "user@example.test",
		Reason:         "REFRESH_UNAVAILABLE",
		RequestedAAL:   "AAL2",
		IdempotencyKey: "generated-per-logical-operation",
	})
}

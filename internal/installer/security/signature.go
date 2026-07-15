package security

import (
	"context"
	"errors"
)

var ErrSignatureVerificationFailed = errors.New("signature verification failed")

type SignedPayload struct {
	Payload   []byte
	Signature []byte
	KeyRef    string
}

type SignatureVerifier interface {
	Verify(ctx context.Context, payload SignedPayload) error
}

type NoopVerifier struct{}

func (NoopVerifier) Verify(ctx context.Context, payload SignedPayload) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(payload.Signature) == 0 {
		return ErrSignatureVerificationFailed
	}
	return nil
}

package idempotency

import (
	"crypto/sha256"
	"encoding/hex"
)

func Key(scope string, requestID string) string {
	sum := sha256.Sum256([]byte(scope + ":" + requestID))
	return hex.EncodeToString(sum[:])
}

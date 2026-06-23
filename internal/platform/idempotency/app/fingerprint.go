package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"m8/internal/platform/idempotency/domain"
)

const (
	FingerprintAlgorithmSHA256 = "sha256"

	CanonicalJSONV1  = "canonical-json-v1"
	CanonicalRawV1   = "raw-bytes-v1"
	CanonicalProtoV1 = "proto-deterministic-v1"
	DefaultSchemaV1  = "v1"
)

// FingerprintInput is a transport-neutral description of a request.
// The caller should pass a stable route name, not a raw URL.
type FingerprintInput struct {
	Method        string
	Route         string
	TenantID      string
	ActorID       string
	SchemaVersion string

	// Body contains request payload.
	// For JSON, it is decoded and encoded again to normalize whitespace and map key order.
	// For Raw and Proto, the bytes are hashed as provided.
	Body []byte

	// SelectedQuery contains only query parameters that are semantically part
	// of the command. Do not include volatile tracking/query params.
	SelectedQuery map[string][]string

	// SelectedHeaders contains only stable headers that affect command semantics.
	// Do not include Authorization, Date, User-Agent, X-Request-Id or trace headers.
	SelectedHeaders map[string]string
}

type JSONFingerprinter struct{}

func (JSONFingerprinter) Fingerprint(_ context.Context, input FingerprintInput) (domain.Fingerprint, error) {
	payload, err := canonicalJSONPayload(input)
	if err != nil {
		return domain.Fingerprint{}, err
	}

	sum := sha256.Sum256(payload)

	return domain.Fingerprint{
		Algorithm:               FingerprintAlgorithmSHA256,
		Hash:                    hex.EncodeToString(sum[:]),
		CanonicalizationVersion: CanonicalJSONV1,
		Method:                  input.Method,
		Route:                   input.Route,
		SchemaVersion:           firstSchemaVersion(input.SchemaVersion),
	}, nil
}

type RawFingerprinter struct{}

func (RawFingerprinter) Fingerprint(_ context.Context, input FingerprintInput) (domain.Fingerprint, error) {
	payload, err := canonicalEnvelope(input, input.Body)
	if err != nil {
		return domain.Fingerprint{}, err
	}

	sum := sha256.Sum256(payload)

	return domain.Fingerprint{
		Algorithm:               FingerprintAlgorithmSHA256,
		Hash:                    hex.EncodeToString(sum[:]),
		CanonicalizationVersion: CanonicalRawV1,
		Method:                  input.Method,
		Route:                   input.Route,
		SchemaVersion:           firstSchemaVersion(input.SchemaVersion),
	}, nil
}

// ProtoFingerprinter expects Body to already be produced by deterministic
// protobuf marshal in the transport adapter.
type ProtoFingerprinter struct{}

func (ProtoFingerprinter) Fingerprint(_ context.Context, input FingerprintInput) (domain.Fingerprint, error) {
	payload, err := canonicalEnvelope(input, input.Body)
	if err != nil {
		return domain.Fingerprint{}, err
	}

	sum := sha256.Sum256(payload)

	return domain.Fingerprint{
		Algorithm:               FingerprintAlgorithmSHA256,
		Hash:                    hex.EncodeToString(sum[:]),
		CanonicalizationVersion: CanonicalProtoV1,
		Method:                  input.Method,
		Route:                   input.Route,
		SchemaVersion:           firstSchemaVersion(input.SchemaVersion),
	}, nil
}

func FingerprintJSON(ctx context.Context, input FingerprintInput) (domain.Fingerprint, error) {
	return JSONFingerprinter{}.Fingerprint(ctx, input)
}

func FingerprintRaw(ctx context.Context, input FingerprintInput) (domain.Fingerprint, error) {
	return RawFingerprinter{}.Fingerprint(ctx, input)
}

func FingerprintProto(ctx context.Context, input FingerprintInput) (domain.Fingerprint, error) {
	return ProtoFingerprinter{}.Fingerprint(ctx, input)
}

type fingerprintEnvelope struct {
	Method          string              `json:"method,omitempty"`
	Route           string              `json:"route"`
	TenantID        string              `json:"tenant_id"`
	ActorID         string              `json:"actor_id,omitempty"`
	SchemaVersion   string              `json:"schema_version"`
	SelectedQuery   map[string][]string `json:"selected_query,omitempty"`
	SelectedHeaders map[string]string   `json:"selected_headers,omitempty"`
	Body            json.RawMessage     `json:"body,omitempty"`
	BodyBase64      []byte              `json:"body_base64,omitempty"`
}

func canonicalJSONPayload(input FingerprintInput) ([]byte, error) {
	var body any
	if len(bytes.TrimSpace(input.Body)) > 0 {
		decoder := json.NewDecoder(bytes.NewReader(input.Body))
		decoder.UseNumber()

		if err := decoder.Decode(&body); err != nil {
			return nil, fmt.Errorf("decode JSON request body for idempotency fingerprint: %w", err)
		}
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal canonical JSON request body for idempotency fingerprint: %w", err)
	}

	return canonicalEnvelope(input, bodyBytes)
}

func canonicalEnvelope(input FingerprintInput, body []byte) ([]byte, error) {
	envelope := fingerprintEnvelope{
		Method:          input.Method,
		Route:           input.Route,
		TenantID:        input.TenantID,
		ActorID:         input.ActorID,
		SchemaVersion:   firstSchemaVersion(input.SchemaVersion),
		SelectedQuery:   input.SelectedQuery,
		SelectedHeaders: input.SelectedHeaders,
	}

	if len(body) > 0 {
		if json.Valid(body) {
			envelope.Body = json.RawMessage(body)
		} else {
			envelope.BodyBase64 = body
		}
	}

	out, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal idempotency fingerprint envelope: %w", err)
	}

	return out, nil
}

func firstSchemaVersion(value string) string {
	if value != "" {
		return value
	}

	return DefaultSchemaV1
}

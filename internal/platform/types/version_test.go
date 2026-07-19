package types

import (
	"encoding"
	"encoding/json"
	"errors"
	"math"
	"strconv"
	"testing"
)

var (
	_ encoding.TextMarshaler   = Version(0)
	_ encoding.TextUnmarshaler = (*Version)(nil)
)

func TestVersionCreationAndMethods(t *testing.T) {
	fromUint, err := NewVersion(7)
	if err != nil {
		t.Fatalf("NewVersion() error = %v", err)
	}
	fromExplicitUint, err := NewVersionFromUint64(7)
	if err != nil {
		t.Fatalf("NewVersionFromUint64() error = %v", err)
	}
	fromInt, err := NewVersionFromInt64(7)
	if err != nil {
		t.Fatalf("NewVersionFromInt64() error = %v", err)
	}
	parsed, err := ParseVersion(" 7 ")
	if err != nil {
		t.Fatalf("ParseVersion() error = %v", err)
	}

	versions := map[string]Version{
		"initial":              NewInitialVersion(),
		"from uint":            fromUint,
		"from explicit uint64": fromExplicitUint,
		"from int64":           fromInt,
		"parsed":               parsed,
		"must parsed":          MustParseVersion("7"),
	}
	for name, version := range versions {
		t.Run(name, func(t *testing.T) {
			if version.IsZero() {
				t.Fatal("version is zero")
			}
			if err := version.Validate(); err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
		})
	}

	if NewInitialVersion() != InitialVersion {
		t.Fatalf("NewInitialVersion() = %s, want %s", NewInitialVersion(), InitialVersion)
	}
	if fromUint.Int64() != 7 || fromUint.Uint64() != 7 || fromUint.String() != "7" {
		t.Fatalf(
			"version conversions = int64 %d, uint64 %d, string %q",
			fromUint.Int64(),
			fromUint.Uint64(),
			fromUint.String(),
		)
	}
	if !fromUint.Equal(parsed) || fromUint.Equal(InitialVersion) {
		t.Fatalf("Equal() produced an unexpected result")
	}

	maximum, err := NewVersion(uint64(math.MaxInt64))
	if err != nil {
		t.Fatalf("NewVersion(MaxInt64) error = %v", err)
	}
	if maximum.Int64() != math.MaxInt64 {
		t.Fatalf("maximum Int64() = %d, want %d", maximum.Int64(), int64(math.MaxInt64))
	}
}

func TestVersionErrors(t *testing.T) {
	tests := []struct {
		name    string
		run     func() error
		wantErr error
	}{
		{
			name:    "new zero",
			run:     func() error { _, err := NewVersion(0); return err },
			wantErr: ErrZeroVersion,
		},
		{
			name:    "new unsigned overflow",
			run:     func() error { _, err := NewVersion(uint64(math.MaxInt64) + 1); return err },
			wantErr: ErrVersionOverflow,
		},
		{
			name:    "new negative int64",
			run:     func() error { _, err := NewVersionFromInt64(-1); return err },
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "parse empty",
			run:     func() error { _, err := ParseVersion(" "); return err },
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "parse malformed",
			run:     func() error { _, err := ParseVersion("one"); return err },
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "parse negative",
			run:     func() error { _, err := ParseVersion("-1"); return err },
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "parse zero",
			run:     func() error { _, err := ParseVersion("0"); return err },
			wantErr: ErrZeroVersion,
		},
		{
			name:    "parse plus sign",
			run:     func() error { _, err := ParseVersion("+1"); return err },
			wantErr: ErrInvalidVersion,
		},
		{
			name:    "parse leading zero",
			run:     func() error { _, err := ParseVersion("01"); return err },
			wantErr: ErrInvalidVersion,
		},
		{
			name: "parse overflow",
			run: func() error {
				_, err := ParseVersion(strconv.FormatUint(uint64(math.MaxInt64)+1, 10))
				return err
			},
			wantErr: ErrVersionOverflow,
		},
		{
			name:    "validate zero",
			run:     Version(0).Validate,
			wantErr: ErrZeroVersion,
		},
		{
			name:    "validate negative",
			run:     Version(-1).Validate,
			wantErr: ErrInvalidVersion,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.run(); !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionNext(t *testing.T) {
	next, err := InitialVersion.Next()
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if next != 2 {
		t.Fatalf("Next() = %s, want 2", next)
	}

	tests := []struct {
		name    string
		version Version
		wantErr error
	}{
		{name: "zero", version: 0, wantErr: ErrZeroVersion},
		{name: "negative", version: -1, wantErr: ErrInvalidVersion},
		{name: "maximum", version: Version(math.MaxInt64), wantErr: ErrVersionOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := tt.version.Next(); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Next() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestVersionTextAndJSONRoundTrip(t *testing.T) {
	version := MustParseVersion("42")

	text, err := version.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText() error = %v", err)
	}
	if string(text) != "42" {
		t.Fatalf("MarshalText() = %q, want 42", text)
	}

	var fromText Version
	if err := fromText.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText() error = %v", err)
	}
	if !fromText.Equal(version) {
		t.Fatalf("UnmarshalText() = %s, want %s", fromText, version)
	}

	encoded, err := json.Marshal(version)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(encoded) != `"42"` {
		t.Fatalf("json.Marshal() = %s, want %q", encoded, "42")
	}

	var fromJSON Version
	if err := json.Unmarshal(encoded, &fromJSON); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if !fromJSON.Equal(version) {
		t.Fatalf("json.Unmarshal() = %s, want %s", fromJSON, version)
	}
}

func TestVersionUnmarshalRejectsInvalidInputWithoutMutation(t *testing.T) {
	version := MustParseVersion("42")
	original := version

	if err := version.UnmarshalText([]byte("invalid")); !errors.Is(err, ErrInvalidVersion) {
		t.Fatalf("UnmarshalText() error = %v, want %v", err, ErrInvalidVersion)
	}
	if !version.Equal(original) {
		t.Fatalf("Version changed after failed unmarshal: got %s, want %s", version, original)
	}

	var nilVersion *Version
	if err := nilVersion.UnmarshalText([]byte("1")); !errors.Is(err, ErrInvalidVersion) {
		t.Fatalf("nil UnmarshalText() error = %v, want %v", err, ErrInvalidVersion)
	}

	if _, err := Version(0).MarshalText(); !errors.Is(err, ErrZeroVersion) {
		t.Fatalf("zero MarshalText() error = %v, want %v", err, ErrZeroVersion)
	}
}

func TestMustParseVersionPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("panic = nil, want non-nil")
		}
	}()

	MustParseVersion("invalid")
}

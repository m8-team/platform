package types

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Version is a positive optimistic-lock version that fits the platform's
// protobuf and database int64 boundaries. The zero value represents an omitted
// optional precondition and is not a valid stored resource version.
type Version int64

// InitialVersion is the first version assigned to a persisted resource.
const InitialVersion Version = 1

var (
	// ErrInvalidVersion indicates a negative or malformed version.
	ErrInvalidVersion = errors.New("invalid version")
	// ErrZeroVersion indicates that a required stored version was omitted.
	ErrZeroVersion = errors.New("version is zero")
	// ErrVersionOverflow indicates that a version exceeds math.MaxInt64.
	ErrVersionOverflow = errors.New("version overflow")
)

// NewInitialVersion returns the first valid resource version.
func NewInitialVersion() Version {
	return InitialVersion
}

// NewVersion creates a Version from an unsigned integer after checking the
// platform int64 boundary.
func NewVersion(value uint64) (Version, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("%w: %d", ErrVersionOverflow, value)
	}

	return NewVersionFromInt64(int64(value))
}

// NewVersionFromUint64 is retained as an explicit boundary constructor.
func NewVersionFromUint64(value uint64) (Version, error) {
	return NewVersion(value)
}

// NewVersionFromInt64 creates a positive Version from protobuf or database
// int64 values.
func NewVersionFromInt64(value int64) (Version, error) {
	version := Version(value)
	if err := version.Validate(); err != nil {
		return 0, err
	}

	return version, nil
}

// ParseVersion parses a canonical base-10 version. Surrounding whitespace is
// ignored, while signs and leading zeroes are rejected.
func ParseVersion(value string) (Version, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("%w: empty", ErrInvalidVersion)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		if errors.Is(err, strconv.ErrRange) {
			return 0, fmt.Errorf("%w: %q", ErrVersionOverflow, value)
		}
		return 0, fmt.Errorf("%w: %q", ErrInvalidVersion, value)
	}
	if strconv.FormatInt(parsed, 10) != value {
		return 0, fmt.Errorf("%w: %q: expected canonical decimal", ErrInvalidVersion, value)
	}

	return NewVersionFromInt64(parsed)
}

// MustParseVersion is ParseVersion for constants and fixtures. It panics on
// invalid input.
func MustParseVersion(value string) Version {
	version, err := ParseVersion(value)
	if err != nil {
		panic(err)
	}

	return version
}

// Validate verifies that v is a positive stored resource version.
func (v Version) Validate() error {
	switch {
	case v < 0:
		return fmt.Errorf("%w: %d", ErrInvalidVersion, v)
	case v == 0:
		return ErrZeroVersion
	default:
		return nil
	}
}

// IsZero reports whether v is the omitted optional-precondition value.
func (v Version) IsZero() bool {
	return v == 0
}

// Int64 returns the protobuf and database representation of v.
func (v Version) Int64() int64 {
	return int64(v)
}

// Uint64 returns the unsigned representation of v. Callers should validate v
// before converting values that may have come from an untrusted boundary.
func (v Version) Uint64() uint64 {
	return uint64(v)
}

// String returns the canonical base-10 representation of v.
func (v Version) String() string {
	return strconv.FormatInt(int64(v), 10)
}

// Equal reports whether v and other represent the same version.
func (v Version) Equal(other Version) bool {
	return v == other
}

// Next returns the next resource version without overflowing int64.
func (v Version) Next() (Version, error) {
	if err := v.Validate(); err != nil {
		return 0, err
	}

	if v == math.MaxInt64 {
		return 0, fmt.Errorf("%w: %d", ErrVersionOverflow, v)
	}

	return v + 1, nil
}

// MarshalText implements encoding.TextMarshaler using canonical decimal text.
func (v Version) MarshalText() ([]byte, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}

	return []byte(v.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler using ParseVersion. The
// receiver is changed only after the complete input has been validated.
func (v *Version) UnmarshalText(text []byte) error {
	if v == nil {
		return fmt.Errorf("%w: nil receiver", ErrInvalidVersion)
	}

	parsed, err := ParseVersion(string(text))
	if err != nil {
		return err
	}

	*v = parsed
	return nil
}

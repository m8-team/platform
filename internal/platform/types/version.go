package types

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Version uint64

const InitialVersion Version = 1

var (
	ErrInvalidVersion  = errors.New("invalid version")
	ErrZeroVersion     = errors.New("version is zero")
	ErrVersionOverflow = errors.New("version overflow")
)

func NewInitialVersion() Version {
	return InitialVersion
}

func NewVersion(value uint64) (Version, error) {
	version := Version(value)

	if err := version.Validate(); err != nil {
		return 0, err
	}

	return version, nil
}

func NewVersionFromUint64(value uint64) (Version, error) {
	if value > math.MaxInt64 {
		return 0, fmt.Errorf("%w: %d", ErrVersionOverflow, value)
	}

	return NewVersion(value)
}

func ParseVersion(value string) (Version, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("%w: empty", ErrInvalidVersion)
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: %q", ErrInvalidVersion, value)
	}

	return NewVersion(uint64(parsed))
}

func MustParseVersion(value string) Version {
	version, err := ParseVersion(value)
	if err != nil {
		panic(err)
	}

	return version
}

func (v Version) Validate() error {
	switch {
	case v == 0:
		return ErrZeroVersion
	case v < 0:
		return fmt.Errorf("%w: %d", ErrInvalidVersion, v)
	default:
		return nil
	}
}

func (v Version) IsZero() bool {
	return v == 0
}

func (v Version) Int64() int64 {
	return int64(v)
}

func (v Version) Uint64() uint64 {
	return uint64(v)
}

func (v Version) String() string {
	return strconv.FormatInt(int64(v), 10)
}

func (v Version) Equal(other Version) bool {
	return v == other
}

func (v Version) Next() (Version, error) {
	if err := v.Validate(); err != nil {
		return 0, err
	}

	if v == math.MaxInt64 {
		return 0, fmt.Errorf("%w: %d", ErrVersionOverflow, v)
	}

	return v + 1, nil
}

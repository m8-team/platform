package common

import (
	"fmt"
	"slices"
)

// ValidateMask ensures that every path in the update mask is supported.
func ValidateMask(paths []string, allowed []string) error {
	if len(paths) == 0 {
		return ErrInvalidMask
	}

	for _, path := range paths {
		if !slices.Contains(allowed, path) {
			return fmt.Errorf("%w: %s", ErrInvalidMask, path)
		}
	}

	return nil
}

package organizationcmd

import (
	"fmt"
	"slices"

	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
)

func validateMask(mask []string, allowed []string) error {
	if len(mask) == 0 {
		return usecasecommon.ErrInvalidMask
	}
	for _, path := range mask {
		if !slices.Contains(allowed, path) {
			return fmt.Errorf("%w: %s", usecasecommon.ErrInvalidMask, path)
		}
	}
	return nil
}

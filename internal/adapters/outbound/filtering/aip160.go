package filtering

import (
	"fmt"
	"strings"

	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
)

type AIP160Validator struct{}

func (AIP160Validator) Validate(raw string) error {
	if len(raw) > 1024 {
		return fmt.Errorf("%w: filter exceeds 1024 characters", usecasecommon.ErrInvalidInput)
	}
	if strings.Contains(raw, ";") {
		return fmt.Errorf("%w: filter contains unsupported separator", usecasecommon.ErrInvalidInput)
	}
	return nil
}

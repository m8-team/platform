package ordering

import (
	"fmt"
	"strings"

	usecasecommon "github.com/m8platform/platform/internal/usecase/resourcemanager/common"
)

type AIP132Validator struct{}

func (AIP132Validator) Validate(raw string) error {
	if len(raw) > 128 {
		return fmt.Errorf("%w: order_by exceeds 128 characters", usecasecommon.ErrInvalidInput)
	}
	if strings.Contains(raw, ";") {
		return fmt.Errorf("%w: order_by contains unsupported separator", usecasecommon.ErrInvalidInput)
	}
	return nil
}

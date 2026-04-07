package ordering

import (
	"fmt"
	"strings"
)

type Parser struct{}

func (Parser) Validate(raw string) error {
	if len(raw) > 128 {
		return fmt.Errorf("order_by exceeds 128 characters")
	}
	if strings.Contains(raw, ";") {
		return fmt.Errorf("order_by parser placeholder rejects ';'")
	}
	return nil
}

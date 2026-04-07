package filters

import (
	"fmt"
	"strings"
)

type Parser struct{}

func (Parser) Validate(raw string) error {
	if len(raw) > 1024 {
		return fmt.Errorf("filter exceeds 1024 characters")
	}
	if strings.Contains(raw, ";") {
		return fmt.Errorf("filter parser placeholder rejects ';'")
	}
	return nil
}

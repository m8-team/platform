package organization

import (
	"fmt"

	"github.com/m8-team/platform/internal/platform/types"
)

// ID is the stable, module-owned identifier of an Organization.
type ID struct {
	value types.ID
}

// NewID creates a new non-zero Organization identifier.
func NewID() ID {
	return ID{value: types.New()}
}

// ParseID parses a UUID string into an Organization identifier.
func ParseID(value string) (ID, error) {
	if value == "" {
		return ID{}, ErrEmptyOrganizationID
	}

	parsed, err := types.Parse(value)
	if err != nil {
		return ID{}, fmt.Errorf("%w: %q: %v", ErrInvalidOrganizationID, value, err)
	}

	return ID{value: parsed}, nil
}

// MustParseID parses value and panics when it is not a valid Organization ID.
// It is intended for constants and test fixtures.
func MustParseID(value string) ID {
	id, err := ParseID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func (id ID) String() string {
	return id.value.String()
}

func (id ID) IsZero() bool {
	return id.value.IsZero()
}

func (id ID) Validate() error {
	if id.IsZero() {
		return ErrEmptyOrganizationID
	}

	return nil
}

func (id ID) Equal(other ID) bool {
	return id.value.Equal(other.value)
}

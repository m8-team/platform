package types

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ID uuid.UUID

var (
	ErrInvalidID = errors.New("invalid id")
	ErrZeroID    = errors.New("id is zero")
)

func NewID() ID {
	return ID(uuid.New())
}

func NewIDFromUUID(value uuid.UUID) (ID, error) {
	id := ID(value)

	if err := id.Validate(); err != nil {
		return ID{}, err
	}

	return id, nil
}

func ParseID(value string) (ID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return ID{}, fmt.Errorf("%w: %q", ErrInvalidID, value)
	}

	id := ID(parsed)

	if err := id.Validate(); err != nil {
		return ID{}, err
	}

	return id, nil
}

func MustParseID(value string) ID {
	id, err := ParseID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func (id ID) UUID() uuid.UUID {
	return uuid.UUID(id)
}

func (id ID) String() string {
	return uuid.UUID(id).String()
}

func (id ID) IsZero() bool {
	return uuid.UUID(id) == uuid.Nil
}

func (id ID) Validate() error {
	if id.IsZero() {
		return ErrZeroID
	}

	return nil
}

func (id ID) Equal(other ID) bool {
	return id == other
}

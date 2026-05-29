package id

import (
	"fmt"

	"github.com/google/uuid"
)

type ID string

func New(value string) (ID, error) {
	if value == "" {
		return "", ErrEmptyID
	}

	if _, err := uuid.Parse(value); err != nil {
		return "", fmt.Errorf("%w: %s", ErrInvalidID, value)
	}

	return ID(value), nil
}

func MustNew(value string) ID {
	id, err := New(value)
	if err != nil {
		panic(err)
	}

	return id
}

func NewUUID() ID {
	return ID(uuid.NewString())
}

func (id ID) String() string {
	return string(id)
}

func (id ID) IsZero() bool {
	return id == ""
}

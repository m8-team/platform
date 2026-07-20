package system

import (
	"time"

	"github.com/google/uuid"
	"github.com/m8-team/platform/internal/resourcemanager/domain/organization"
	"github.com/m8-team/platform/internal/resourcemanager/domain/workspace"
)

type Clock struct{}

func NewClock() *Clock {
	return &Clock{}
}

func (*Clock) Now() time.Time {
	return time.Now().UTC()
}

type IDGenerator struct{}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

func (*IDGenerator) NewID() organization.ID {
	return organization.NewID()
}

func (*IDGenerator) NewWorkspaceID() workspace.ID {
	return workspace.NewID()
}

func (*IDGenerator) NewOperationID() string {
	return uuid.NewString()
}

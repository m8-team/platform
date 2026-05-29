package organization

import (
	"time"

	"github.com/m8platform/platform/pkg/platform"
)

const ResourceType = "resourcemanager.organization"

type Organization struct {
	id          ID
	state       State
	name        string
	description string
	labels      Labels

	createTime time.Time
	updateTime time.Time
	deleteTime *time.Time
	purgeTime  *time.Time

	version platform.Version

	events []Event
}

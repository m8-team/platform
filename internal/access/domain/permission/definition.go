package permission

import (
	"errors"
	"regexp"
	"strings"
)

var ErrInvalidDefinition = errors.New("invalid permission definition")
var graphName = regexp.MustCompile("^[a-z][a-z0-9_]{0,127}$")

// Definition maps a public action and resource kind to a graph permission.
// Its publication is managed by AuthorizationModel, not by the SpiceDB adapter.
type Definition struct {
	permission      Permission
	resourceType    string
	graphPermission string
}

func NewDefinition(p Permission, resourceType, graphPermission string) (Definition, error) {
	d := Definition{permission: p, resourceType: strings.TrimSpace(resourceType), graphPermission: strings.TrimSpace(graphPermission)}
	return d, d.Validate()
}
func (d Definition) Validate() error {
	if err := d.permission.Validate(); err != nil {
		return err
	}
	if !graphName.MatchString(d.resourceType) || !graphName.MatchString(d.graphPermission) {
		return ErrInvalidDefinition
	}
	return nil
}
func (d Definition) Permission() Permission  { return d.permission }
func (d Definition) ResourceType() string    { return d.resourceType }
func (d Definition) GraphPermission() string { return d.graphPermission }
func (d Definition) Key() string             { return d.resourceType + "/" + d.permission.Name() }

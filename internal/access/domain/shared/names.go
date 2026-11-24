package shared

import "regexp"

// Qualified names are module-owned identifiers: finance.invoice or finance.invoice.view.
var (
	moduleNamePattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	qualifiedNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}(\.[a-z][a-z0-9_]{0,63})+$`)
	versionPattern       = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`)
)

func ValidModuleName(value string) bool { return moduleNamePattern.MatchString(value) }
func ValidQualifiedName(value string) bool {
	return qualifiedNamePattern.MatchString(value) && len(value) <= 128
}
func ValidReleaseVersion(value string) bool {
	return versionPattern.MatchString(value) && len(value) <= 48
}

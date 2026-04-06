package migrator

import (
	"regexp"
	"strings"
)

var (
	createTablePattern = regexp.MustCompile(`(?im)^\s*CREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+` + "`?" + `([A-Za-z0-9_./-]+)` + "`?")
	nonCreatePattern   = regexp.MustCompile(`(?im)\bALTER\s+TABLE\b|\bDROP\s+TABLE\b|\bUPSERT\s+INTO\b|\bINSERT\s+INTO\b|\bUPDATE\b|\bDELETE\s+FROM\b`)
)

func ParseExpectedTables(sql string) []string {
	matches := createTablePattern.FindAllStringSubmatch(sql, -1)
	if len(matches) == 0 {
		return nil
	}
	unique := make(map[string]struct{}, len(matches))
	tables := make([]string, 0, len(matches))
	for _, match := range matches {
		name := strings.TrimSpace(match[1])
		if name == "" {
			continue
		}
		if _, seen := unique[name]; seen {
			continue
		}
		unique[name] = struct{}{}
		tables = append(tables, name)
	}
	return tables
}

func IsCreateOnlyMigration(sql string) bool {
	if len(ParseExpectedTables(sql)) == 0 {
		return false
	}
	return !nonCreatePattern.MatchString(sql)
}

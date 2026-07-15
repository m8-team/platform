package diagnostics

import (
	"regexp"
	"strings"
)

var sensitiveKeyPattern = regexp.MustCompile(`(?i)(password|passwd|secret|token|private[_-]?key|client[_-]?secret|authorization|cookie|otp|refresh[_-]?token|access[_-]?token)`)

func RedactMap(input map[string]string) map[string]string {
	if input == nil {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		if sensitiveKeyPattern.MatchString(key) || looksLikeBearer(value) {
			output[key] = "[REDACTED]"
			continue
		}
		output[key] = value
	}
	return output
}

func RedactText(input string) string {
	lines := strings.Split(input, "\n")
	for i, line := range lines {
		if sensitiveKeyPattern.MatchString(line) || looksLikeBearer(line) {
			lines[i] = "[REDACTED]"
		}
	}
	return strings.Join(lines, "\n")
}

func looksLikeBearer(value string) bool {
	return strings.Contains(strings.ToLower(value), "bearer ") || strings.HasPrefix(value, "eyJ")
}

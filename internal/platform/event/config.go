package event

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

var (
	ErrInvalidConfiguration = errors.New("invalid event configuration")
	ErrInvalidTopic         = errors.New("invalid event topic")
	ErrInvalidEvent         = errors.New("invalid event")
	ErrEventTooLarge        = errors.New("event exceeds size limit")
	ErrBatchTooLarge        = errors.New("event batch exceeds limit")
)

// Config is trusted startup configuration, never request input.
// Topics maps logical bounded-context names to absolute YDB topic paths.
type Config struct {
	Topics         map[string]string
	MaxEventBytes  int
	MaxBatchEvents int
	MaxBatchBytes  int
}

func DefaultConfig(topics map[string]string) Config {
	return Config{
		Topics:        maps.Clone(topics),
		MaxEventBytes: 256 << 10, MaxBatchEvents: 100, MaxBatchBytes: 4 << 20,
	}
}

func (c Config) Validate() error {
	if c.MaxEventBytes <= 0 ||
		c.MaxBatchEvents <= 0 || c.MaxBatchBytes < c.MaxEventBytes || len(c.Topics) == 0 {
		return ErrInvalidConfiguration
	}
	for name, topic := range c.Topics {
		if !validName(name) || !strings.HasPrefix(topic, "/") ||
			topic == "/" || path.Clean(topic) != topic || !utf8.ValidString(topic) ||
			strings.ContainsFunc(topic, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) {
			return fmt.Errorf("%w: invalid topic mapping", ErrInvalidConfiguration)
		}
	}
	return nil
}

func validName(s string) bool {
	if len(s) == 0 || len(s) > 128 || s == "." || s == ".." {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}

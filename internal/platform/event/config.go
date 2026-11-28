package event

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"regexp"
	"strings"
	"time"
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

var topicNamePattern = regexp.MustCompile(
	`^([a-z0-9_-][a-z0-9_.-]?|\.[a-z0-9_-]|[a-z0-9_.-]{3,128})$`,
)

// Config is trusted startup configuration, never request input.
// Topics maps logical bounded-context names to absolute YDB topic paths.
type Config struct {
	Topics            map[string]string
	MinTopicRetention time.Duration
	MaxEventBytes     int
	MaxBatchEvents    int
	MaxBatchBytes     int
}

func DefaultConfig(topics map[string]string) Config {
	return Config{
		Topics:        maps.Clone(topics),
		MaxEventBytes: 512 << 10, MaxBatchEvents: 100, MaxBatchBytes: 8 << 20,
	}
}

func (c Config) Validate() error {
	if c.MinTopicRetention < 0 || c.MaxEventBytes <= 0 ||
		c.MaxBatchEvents <= 0 || c.MaxBatchBytes < c.MaxEventBytes || len(c.Topics) == 0 {
		return ErrInvalidConfiguration
	}
	paths := make(map[string]struct{}, len(c.Topics))
	for name, topic := range c.Topics {
		if !validName(name) || !validTopicPath(topic) {
			return fmt.Errorf("%w: invalid topic mapping", ErrInvalidConfiguration)
		}
		if _, exists := paths[topic]; exists {
			return fmt.Errorf("%w: duplicate physical topic mapping", ErrInvalidConfiguration)
		}
		paths[topic] = struct{}{}
	}
	return nil
}

func validName(s string) bool {
	return topicNamePattern.MatchString(s)
}

func validTopicPath(topic string) bool {
	return path.IsAbs(topic) &&
		topic != "/" &&
		path.Clean(topic) == topic &&
		utf8.ValidString(topic) &&
		!strings.ContainsFunc(topic, func(r rune) bool {
			return unicode.IsSpace(r) || unicode.IsControl(r)
		})
}

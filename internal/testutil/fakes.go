package testutil

import "time"

type FakeClock struct {
	Current time.Time
}

func (c FakeClock) Now() time.Time {
	return c.Current
}

type SequenceUUIDGenerator struct {
	Values []string
	index  int
}

func (g *SequenceUUIDGenerator) NewString() string {
	if len(g.Values) == 0 {
		return ""
	}
	if g.index >= len(g.Values) {
		return g.Values[len(g.Values)-1]
	}
	value := g.Values[g.index]
	g.index++
	return value
}

func StringPointer(value string) *string {
	v := value
	return &v
}

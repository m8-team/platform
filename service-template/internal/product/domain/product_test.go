package domain

import (
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	for _, name := range []string{"", " ", " name", strings.Repeat("x", 201)} {
		if _, err := New("id", "tenant", name); err == nil {
			t.Fatalf("accepted %q", name)
		}
	}
	if _, err := New("id", "tenant", "Coffee"); err != nil {
		t.Fatal(err)
	}
}
func FuzzProduct(f *testing.F) {
	f.Add("Coffee")
	f.Add("")
	f.Fuzz(func(t *testing.T, name string) {
		p, err := New("id", "tenant", name)
		if err == nil && p.Name != name {
			t.Fatal("name changed")
		}
	})
}

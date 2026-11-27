package role

import (
	"errors"

	"github.com/m8-team/platform/internal/access/domain/shared"
)

var (
	ErrInvalidBundle    = errors.New("invalid composite access profile")
	ErrUnresolvedBundle = errors.New("unresolved profile dependency")
	ErrCyclicBundle     = errors.New("cyclic composite access profile")
)

type BundleRef struct {
	Name    string
	Version string
}

func (r BundleRef) Validate() error {
	if !shared.ValidQualifiedName(r.Name) || !shared.ValidReleaseVersion(r.Version) {
		return ErrInvalidBundle
	}
	return nil
}
func (r BundleRef) Key() string { return r.Name + "@" + r.Version }

// Bundle is a version-pinned, immutable access profile, e.g. profiles.accountant.
// It does not itself authorize access; an Assignment is always required.
type Bundle struct {
	ref      BundleRef
	roles    []TemplateRef
	includes []BundleRef
}

func NewBundle(ref BundleRef, roles []TemplateRef, includes []BundleRef) (Bundle, error) {
	b := Bundle{ref: ref, roles: append([]TemplateRef(nil), roles...), includes: append([]BundleRef(nil), includes...)}
	return b, b.Validate()
}
func (b Bundle) Validate() error {
	if b.ref.Validate() != nil || len(b.roles)+len(b.includes) == 0 || len(b.roles)+len(b.includes) > 100 {
		return ErrInvalidBundle
	}
	seen := map[string]bool{}
	for _, r := range b.roles {
		if r.Validate() != nil || seen["r:"+r.Key()] {
			return ErrInvalidBundle
		}
		seen["r:"+r.Key()] = true
	}
	for _, r := range b.includes {
		if r.Validate() != nil || r == b.ref || seen["b:"+r.Key()] {
			return ErrInvalidBundle
		}
		seen["b:"+r.Key()] = true
	}
	return nil
}
func (b Bundle) Ref() BundleRef        { return b.ref }
func (b Bundle) Roles() []TemplateRef  { return append([]TemplateRef(nil), b.roles...) }
func (b Bundle) Includes() []BundleRef { return append([]BundleRef(nil), b.includes...) }

// ExpandProfiles verifies the complete pinned graph, rejects missing versions,
// detects cycles, and returns a deduplicated list of role-template references.
// An application must also resolve every TemplateRef from the published catalog.
func ExpandProfiles(root BundleRef, available []Bundle) ([]TemplateRef, error) {
	if root.Validate() != nil {
		return nil, ErrInvalidBundle
	}
	if len(available) > 1000 {
		return nil, ErrInvalidBundle
	}
	index := make(map[string]Bundle, len(available))
	for _, b := range available {
		if err := b.Validate(); err != nil {
			return nil, err
		}
		if _, exists := index[b.ref.Key()]; exists {
			return nil, ErrInvalidBundle
		}
		index[b.ref.Key()] = b
	}
	visiting := map[string]bool{}
	visited := map[string]bool{}
	roles := map[string]bool{}
	var output []TemplateRef
	var walk func(BundleRef, int) error
	walk = func(ref BundleRef, depth int) error {
		if depth > 32 {
			return ErrInvalidBundle
		}
		key := ref.Key()
		if visiting[key] {
			return ErrCyclicBundle
		}
		if visited[key] {
			return nil
		}
		b, ok := index[key]
		if !ok {
			return ErrUnresolvedBundle
		}
		visiting[key] = true
		for _, child := range b.includes {
			if err := walk(child, depth+1); err != nil {
				return err
			}
		}
		for _, r := range b.roles {
			if !roles[r.Key()] {
				roles[r.Key()] = true
				output = append(output, r)
			}
		}
		visiting[key] = false
		visited[key] = true
		return nil
	}
	if err := walk(root, 0); err != nil {
		return nil, err
	}
	return output, nil
}

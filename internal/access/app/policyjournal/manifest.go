// Package policyjournal describes the authoritative transaction manifest for
// Access policy changes. A role edit MUST include one job for EVERY affected
// RoleBinding; a grant lifecycle mutation includes the distinct AccessGrant.
package policyjournal

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/m8-team/platform/internal/access/app/projection"
	"github.com/m8-team/platform/internal/access/domain"
)

var ErrInvalidManifest = errors.New("invalid Access policy mutation manifest")

// Job is a committed projection target, not a caller-supplied relationship.
type Job struct {
	Key              projection.Key
	RequiredRevision int64
}

func isPolicyTarget(kind projection.Kind) bool {
	return kind == projection.KindGrant || kind == projection.KindRoleBinding || kind == projection.KindProfileAssignment
}

// Mutation MUST be written within the SAME serializable YDB transaction as
// the authoritative Access grant, role or role binding mutation + outbox.
// Bootstrap is only for an audited full entitlement inventory/migration.
// ExpectedHead is a tenant-scoped CAS value; events may be processed out of order.
type Mutation struct {
	TenantID     domain.TenantID
	ExpectedHead uint64
	ID           string
	Bootstrap    bool
	Jobs         []Job
}

func (m Mutation) Validate() error {
	if m.TenantID.Validate() != nil || strings.TrimSpace(m.ID) != m.ID || m.ID == "" ||
		len(m.ID) > 256 || len(m.Jobs) == 0 || len(m.Jobs) > 1000 ||
		(m.Bootstrap && m.ExpectedHead != 0) {
		return ErrInvalidManifest
	}
	unique := make(map[projection.Key]bool, len(m.Jobs))
	for _, job := range m.Jobs {
		if job.Key.Validate() != nil || job.Key.TenantID != m.TenantID ||
			!isPolicyTarget(job.Key.Kind) ||
			job.RequiredRevision < 1 || unique[job.Key] {
			return ErrInvalidManifest
		}
		unique[job.Key] = true
	}
	return nil
}

// Digest binds the idempotency key to exact fanout contents. Sorting makes
// repeated commits independent of map and list iteration order.
func (m Mutation) Digest() (string, error) {
	if err := m.Validate(); err != nil {
		return "", err
	}
	entries := make([]string, 0, len(m.Jobs))
	for _, job := range m.Jobs {
		entries = append(entries, fmt.Sprintf("%s:%d", job.Key.String(), job.RequiredRevision))
	}
	sort.Strings(entries)
	payload := fmt.Sprintf("%s|%d|%s|%t|%s", m.TenantID, m.ExpectedHead, m.ID, m.Bootstrap, strings.Join(entries, "|"))
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:]), nil
}

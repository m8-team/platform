package domain

import (
	"errors"
	"time"
)

type State string

const (
	ChallengePending State = "CHALLENGE_PENDING"
	Authenticated    State = "AUTHENTICATED"
	Failed           State = "FAILED"
	Cancelled        State = "CANCELLED"
	Expired          State = "EXPIRED"
)

var (
	ErrClientDisabled = errors.New("client disabled")
	ErrRiskDenied     = errors.New("risk denied")
)

type Transaction struct {
	ID                      string
	ClientID                string
	SubjectID               string
	RequestedAssuranceLevel string
	State                   State
	CreatedAt               time.Time
	ExpiresAt               time.Time
	Revision                uint64
}

func NewTransaction(
	id, clientID, subjectID, aal string,
	now time.Time,
	ttl time.Duration,
) Transaction {
	return Transaction{
		ID:                      id,
		ClientID:                clientID,
		SubjectID:               subjectID,
		RequestedAssuranceLevel: aal,
		State:                   ChallengePending,
		CreatedAt:               now,
		ExpiresAt:               now.Add(ttl),
		Revision:                1,
	}
}

type Client struct {
	ID      string
	Enabled bool
	CIBA    bool
}

type RiskDecision struct {
	Allowed           bool
	RequiredAAL       string
	RequiredChallenge string
}

type Event struct {
	ID                string
	Type              string
	AggregateID       string
	AggregateRevision uint64
	CorrelationID     string
	OccurredAt        time.Time
}

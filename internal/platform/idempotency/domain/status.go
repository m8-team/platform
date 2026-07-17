package domain

import "fmt"

// Status describes the lifecycle state of an idempotency record.
type Status string

const (
	StatusReceived        Status = "RECEIVED"
	StatusProcessing      Status = "PROCESSING"
	StatusCompleted       Status = "COMPLETED"
	StatusFailedRetryable Status = "FAILED_RETRYABLE"
	StatusFailedFinal     Status = "FAILED_FINAL"
	StatusUnknown         Status = "UNKNOWN"
	StatusExpired         Status = "EXPIRED"
)

func (s Status) String() string {
	return string(s)
}

func (s Status) IsTerminal() bool {
	switch s {
	case StatusCompleted, StatusFailedFinal, StatusExpired:
		return true
	default:
		return false
	}
}

func (s Status) CanReplay() bool {
	switch s {
	case StatusCompleted, StatusFailedFinal:
		return true
	default:
		return false
	}
}

func (s Status) CanExecute() bool {
	switch s {
	case StatusReceived, StatusFailedRetryable:
		return true
	default:
		return false
	}
}

func (s Status) IsValid() bool {
	switch s {
	case StatusReceived,
		StatusProcessing,
		StatusCompleted,
		StatusFailedRetryable,
		StatusFailedFinal,
		StatusUnknown,
		StatusExpired:
		return true
	default:
		return false
	}
}

func ParseStatus(value string) (Status, error) {
	status := Status(value)
	if !status.IsValid() {
		return "", fmt.Errorf("%w: %q", ErrInvalidStatus, value)
	}

	return status, nil
}

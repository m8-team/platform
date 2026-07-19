package organization

import "testing"

func TestState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		state State
		text  string
		valid bool
	}{
		{state: StateUnspecified, text: "STATE_UNSPECIFIED", valid: false},
		{state: StateCreating, text: "CREATING", valid: true},
		{state: StateActive, text: "ACTIVE", valid: true},
		{state: StateSuspended, text: "SUSPENDED", valid: true},
		{state: StateDeleting, text: "DELETING", valid: true},
		{state: StateDeleted, text: "DELETED", valid: true},
		{state: StateFailed, text: "FAILED", valid: true},
		{state: State(99), text: "State(99)", valid: false},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			if got := tt.state.String(); got != tt.text {
				t.Fatalf("String() = %q, want %q", got, tt.text)
			}
			if got := tt.state.IsValid(); got != tt.valid {
				t.Fatalf("IsValid() = %t, want %t", got, tt.valid)
			}
		})
	}
}

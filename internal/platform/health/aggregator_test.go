package health

import "testing"

func TestAggregate(t *testing.T) {
	tests := []struct {
		name    string
		results []Result
		want    Status
	}{
		{
			name: "empty results",
			want: StatusHealthy,
		},
		{
			name: "all healthy",
			results: []Result{
				{Status: StatusHealthy, Criticality: CriticalityRequired},
				{Status: StatusHealthy, Criticality: CriticalityOptional},
			},
			want: StatusHealthy,
		},
		{
			name: "optional unhealthy",
			results: []Result{
				{Status: StatusHealthy, Criticality: CriticalityRequired},
				{Status: StatusUnhealthy, Criticality: CriticalityOptional},
			},
			want: StatusDegraded,
		},
		{
			name: "required unhealthy",
			results: []Result{
				{Status: StatusUnhealthy, Criticality: CriticalityRequired},
				{Status: StatusHealthy, Criticality: CriticalityOptional},
			},
			want: StatusUnhealthy,
		},
		{
			name: "required unknown",
			results: []Result{
				{Status: StatusUnknown, Criticality: CriticalityRequired},
			},
			want: StatusUnhealthy,
		},
		{
			name: "optional unknown",
			results: []Result{
				{Status: StatusUnknown, Criticality: CriticalityOptional},
			},
			want: StatusDegraded,
		},
		{
			name: "degraded only",
			results: []Result{
				{Status: StatusDegraded, Criticality: CriticalityRequired},
			},
			want: StatusDegraded,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Aggregate(tt.results); got != tt.want {
				t.Fatalf("Aggregate() = %s, want %s", got, tt.want)
			}
		})
	}
}

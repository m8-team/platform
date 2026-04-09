package requestmetadata

import (
	"context"
	"errors"
	"testing"
)

type stubGenerator struct {
	ids   []string
	index int
}

func (g *stubGenerator) NewID() string {
	id := g.ids[g.index]
	g.index++
	return id
}

type mapCarrier map[string]string

func (c mapCarrier) Get(key string) string {
	return c[key]
}

func TestMetadataNormalize(t *testing.T) {
	got := Metadata{
		Actor:          " actor ",
		CorrelationID:  " corr ",
		IdempotencyKey: " idem ",
		RequestID:      " req ",
		Source:         " HTTP ",
	}.Normalize()

	want := Metadata{
		Actor:          "actor",
		CorrelationID:  "corr",
		IdempotencyKey: "idem",
		RequestID:      "req",
		Source:         SourceHTTP,
	}

	if got != want {
		t.Fatalf("Normalize() = %#v, want %#v", got, want)
	}
}

func TestMetadataHasIdempotencyKey(t *testing.T) {
	if (Metadata{IdempotencyKey: "   "}).HasIdempotencyKey() {
		t.Fatal("HasIdempotencyKey() = true, want false")
	}
	if !(Metadata{IdempotencyKey: " idem "}).HasIdempotencyKey() {
		t.Fatal("HasIdempotencyKey() = false, want true")
	}
}

func TestNewForCLI(t *testing.T) {
	gen := &stubGenerator{ids: []string{" corr ", " req "}}

	got := NewForCLI(" actor ", gen)
	want := Metadata{
		Actor:         "actor",
		CorrelationID: "corr",
		RequestID:     "req",
		Source:        SourceCLI,
	}

	if got != want {
		t.Fatalf("NewForCLI() = %#v, want %#v", got, want)
	}
}

func TestNewForWorker(t *testing.T) {
	gen := &stubGenerator{ids: []string{" req "}}

	got := NewForWorker(" actor ", " corr ", " idem ", gen)
	want := Metadata{
		Actor:          "actor",
		CorrelationID:  "corr",
		IdempotencyKey: "idem",
		RequestID:      "req",
		Source:         SourceWorker,
	}

	if got != want {
		t.Fatalf("NewForWorker() = %#v, want %#v", got, want)
	}
}

func TestFromCarrierPrefersFirstNonEmptyValue(t *testing.T) {
	carrier := mapCarrier{
		HeaderCorrelationID:    " corr-header ",
		MetadataCorrelationID:  "corr-meta",
		MetadataRequestID:      " req-meta ",
		HeaderIdempotencyKey:   " idem-header ",
		MetadataIdempotencyKey: "idem-meta",
	}

	got := FromCarrier(" actor ", " HTTP ", carrier, nil)
	want := Metadata{
		Actor:          "actor",
		CorrelationID:  "corr-header",
		RequestID:      "req-meta",
		IdempotencyKey: "idem-header",
		Source:         SourceHTTP,
	}

	if got != want {
		t.Fatalf("FromCarrier() = %#v, want %#v", got, want)
	}
}

func TestFromCarrierGeneratesMissingIDs(t *testing.T) {
	gen := &stubGenerator{ids: []string{" corr ", " req "}}

	got := FromCarrier("actor", SourceGRPC, nil, gen)
	want := Metadata{
		Actor:         "actor",
		CorrelationID: "corr",
		RequestID:     "req",
		Source:        SourceGRPC,
	}

	if got != want {
		t.Fatalf("FromCarrier() = %#v, want %#v", got, want)
	}
}

func TestValidateForQuery(t *testing.T) {
	tests := []struct {
		name string
		meta Metadata
		want error
	}{
		{
			name: "invalid source",
			meta: Metadata{
				Actor:         "actor",
				CorrelationID: "corr",
				RequestID:     "req",
				Source:        "queue",
			},
			want: ErrInvalidSource,
		},
		{
			name: "missing actor",
			meta: Metadata{
				CorrelationID: "corr",
				RequestID:     "req",
				Source:        SourceHTTP,
			},
			want: ErrMissingActor,
		},
		{
			name: "missing correlation id",
			meta: Metadata{
				Actor:     "actor",
				RequestID: "req",
				Source:    SourceHTTP,
			},
			want: ErrMissingCorrelationID,
		},
		{
			name: "missing request id",
			meta: Metadata{
				Actor:         "actor",
				CorrelationID: "corr",
				Source:        SourceHTTP,
			},
			want: ErrMissingRequestID,
		},
		{
			name: "valid",
			meta: Metadata{
				Actor:         "actor",
				CorrelationID: "corr",
				RequestID:     "req",
				Source:        SourceHTTP,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.meta.ValidateForQuery()
			if !errors.Is(err, tt.want) {
				t.Fatalf("ValidateForQuery() error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestValidateForCommand(t *testing.T) {
	meta := Metadata{
		Actor:         "actor",
		CorrelationID: "corr",
		RequestID:     "req",
		Source:        SourceHTTP,
	}

	if err := meta.ValidateForCommand(false); err != nil {
		t.Fatalf("ValidateForCommand(false) error = %v, want nil", err)
	}

	err := meta.ValidateForCommand(true)
	if !errors.Is(err, ErrMissingIdempotencyKey) {
		t.Fatalf("ValidateForCommand(true) error = %v, want %v", err, ErrMissingIdempotencyKey)
	}
}

func TestContextRoundTrip(t *testing.T) {
	ctx := IntoContext(context.Background(), Metadata{
		Actor:         " actor ",
		CorrelationID: " corr ",
		RequestID:     " req ",
		Source:        " cli ",
	})

	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext() ok = false, want true")
	}

	want := Metadata{
		Actor:         "actor",
		CorrelationID: "corr",
		RequestID:     "req",
		Source:        SourceCLI,
	}

	if got != want {
		t.Fatalf("FromContext() = %#v, want %#v", got, want)
	}
}

func TestMustFromContextPanicsWhenMissing(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("MustFromContext() did not panic")
		}
	}()

	MustFromContext(context.Background())
}

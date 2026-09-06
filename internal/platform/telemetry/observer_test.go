package telemetry

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestObserverRecordsOutcomeWithoutSensitiveError(t *testing.T) {
	var logs bytes.Buffer
	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	defer provider.Shutdown(context.Background())
	observer := New(slog.New(slog.NewJSONHandler(&logs, nil)), provider)
	_, finish := observer.Start(context.Background(), "organization.create")
	finish(errors.New("secret password and email"))
	if strings.Contains(logs.String(), "secret") {
		t.Fatal("sensitive error leaked")
	}
	if len(recorder.Ended()) != 1 {
		t.Fatal("span not finished")
	}
	response := httptest.NewRecorder()
	observer.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	var metrics map[string]map[string]float64
	if err := json.Unmarshal(response.Body.Bytes(), &metrics); err != nil {
		t.Fatal(err)
	}
	if metrics["organization.create"]["calls"] != 1 || metrics["organization.create"]["failures"] != 1 {
		t.Fatalf("metrics = %v", metrics)
	}
}

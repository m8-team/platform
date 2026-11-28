package main

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReconciliationFreshness(t *testing.T) {
	var status reconciliationStatus
	now := time.Now()
	if err := status.ready(now); err == nil {
		t.Fatal("ready before successful reconciliation")
	}
	status.record(now, nil)
	status.record(now.Add(time.Second), errors.New("Kubernetes unavailable"))
	if err := status.ready(now.Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := status.ready(now.Add(time.Minute)); err == nil {
		t.Fatal("ready with stale ownership observation")
	}
	if status.failures.Load() != 1 {
		t.Fatal("expected one recorded error")
	}
}

func TestRejectDuplicatePredecessor(t *testing.T) {
	old := uuid.NewString()
	plan := deployment{Namespace: "test", StatefulSet: "publisher", UID: uuid.NewString(), Slots: 1,
		Predecessors: []incarnation{{UID: old, Slots: 2}, {UID: old, Slots: 3}}}
	if err := plan.validate(); err == nil {
		t.Fatal("duplicate predecessor accepted")
	}
}

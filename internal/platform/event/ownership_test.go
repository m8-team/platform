package event

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReplicaSlotAcceptsOnlyCurrentStatefulSetPod(t *testing.T) {
	const setUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const podUID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	slot := ReplicaSlot{"namespace", setUID, "publisher-0", "/local/owners"}
	if err := slot.validate(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, namespace, owner, terminating string
		status                              int
		want                                bool
	}{
		{name: "matching", namespace: "namespace", owner: setUID, want: true},
		{name: "wrong namespace", namespace: "other", owner: setUID},
		{name: "wrong owner", namespace: "namespace", owner: podUID},
		{name: "terminating", namespace: "namespace", owner: setUID, terminating: `,"deletionTimestamp":"2026-01-01T00:00:00Z"`},
		{name: "forbidden", status: http.StatusForbidden},
		{name: "missing", status: http.StatusNotFound},
		{name: "redirect", status: http.StatusTemporaryRedirect},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.status != 0 {
					w.Header().Set("Location", "/redirected")
					w.WriteHeader(tc.status)
					return
				}
				fmt.Fprintf(w, `{"metadata":{"uid":%q,"name":"publisher-0","namespace":%q%s,"ownerReferences":[{"apiVersion":"apps/v1","kind":"StatefulSet","controller":true,"uid":%q}]}}`,
					podUID, tc.namespace, tc.terminating, tc.owner)
			}))
			defer server.Close()
			got, err := slot.currentPodUID(t.Context(), server.Client(), server.URL)
			if tc.want {
				if err != nil || got != podUID {
					t.Fatalf("got uid=%q error=%v", got, err)
				}
			} else if got != "" || !errors.Is(err, ErrReplicaOwnership) {
				t.Fatalf("invalid Pod response accepted: uid=%q error=%v", got, err)
			}
		})
	}
}

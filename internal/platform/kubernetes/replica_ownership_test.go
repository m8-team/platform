package kubernetes

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestReplicaSlotAcceptsOnlyCurrentStatefulSetPod(t *testing.T) {
	const setUID = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	const podUID = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
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
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if strings.Contains(r.URL.Path, "/statefulsets/") {
					return response(http.StatusOK, fmt.Sprintf(`{"metadata":{"uid":%q,"name":"publisher","namespace":"namespace"}}`, setUID)), nil
				}
				if tc.status != 0 {
					return response(tc.status, ""), nil
				}
				return response(http.StatusOK, fmt.Sprintf(`{"metadata":{"uid":%q,"name":"publisher-0","namespace":%q%s,"ownerReferences":[{"apiVersion":"apps/v1","kind":"StatefulSet","controller":true,"uid":%q}]}}`,
					podUID, tc.namespace, tc.terminating, tc.owner)), nil
			})}
			got, retired, err := ObserveReplica(t.Context(), client, "https://kubernetes.default.svc", "namespace", "publisher", setUID, "publisher-0")
			if retired {
				t.Fatal("unexpected retirement")
			}
			if tc.want {
				if err != nil || got != podUID {
					t.Fatalf("got uid=%q error=%v", got, err)
				}
			} else if got != "" || (err == nil) != (tc.status == http.StatusNotFound || tc.terminating != "" || tc.name == "wrong owner") {
				t.Fatalf("invalid Pod response accepted: uid=%q error=%v", got, err)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func response(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}

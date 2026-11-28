package event

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/m8-team/platform/internal/platform/kubernetes"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Query_V1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Topic_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Query"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
	"google.golang.org/protobuf/proto"
)

const kubernetesFixtureNamespace = "event-handoff"

func TestIntegrationKubernetesReplicaHandoff(t *testing.T) {
	fixture := os.Getenv("M8_EVENT_KUBERNETES_FIXTURE")
	if fixture == "" {
		t.Skip("run the disposable kind handoff fixture")
	}
	ctx, db, path := liveDatabase(t)
	createTestTopic(t, ctx, db, path, 1)
	state := createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
	createTestTable(t, ctx, db, path+"_owner", "replica Utf8 NOT NULL, pod_uid Utf8 NOT NULL, generation Uint64 NOT NULL, PRIMARY KEY (replica)")
	guard := createTestTable(t, ctx, db, path+"_guard", "replica Utf8 NOT NULL, revision Uint64 NOT NULL, PRIMARY KEY (replica)")
	readFile := func(name string) []byte {
		t.Helper()
		data, err := os.ReadFile(fixture + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	command := func(name string, args ...string) string {
		t.Helper()
		out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
		if err != nil {
			t.Fatalf("fixture command %s failed: %v: %s", name, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	ctl := func(args ...string) string {
		return command("kubectl", append([]string{"--kubeconfig", fixture + "/kubeconfig", "--context", "kind-m8-event-handoff", "-n", kubernetesFixtureNamespace}, args...)...)
	}
	const publisherUser = "eventfixturepublisher"
	password := "EventAa1" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if err := db.Query().Exec(ctx, "CREATE USER "+publisherUser+" PASSWORD '"+password+"';"); err != nil {
		t.Fatal("could not create publisher fixture credentials")
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		_ = db.Query().Exec(cleanup, "DROP USER "+publisherUser+";")
	})
	for _, grant := range []string{
		"GRANT CONNECT ON `/local`",
		"GRANT SELECT ON `" + path + "_owner`",
		"GRANT SELECT, UPDATE ROW ON " + state + ", " + guard + ", `" + path + "`",
	} {
		if err := db.Query().Exec(ctx, grant+" TO "+publisherUser+";"); err != nil {
			t.Fatalf("grant disposable publisher access: %v", err)
		}
	}
	passwordFile := fixture + "/publisher-password"
	if err := os.WriteFile(passwordFile, []byte(password), 0600); err != nil {
		t.Fatal("could not prepare publisher fixture secret")
	}
	ctl("create", "secret", "generic", "publisher-ydb", "--from-file=password="+passwordFile)
	if err := os.Remove(passwordFile); err != nil {
		t.Fatal("could not unlink temporary password file")
	}
	// Build the Pod command from fixture values. No application service is exposed.
	setUID := ctl("get", "statefulset", "publisher", "-o", "jsonpath={.metadata.uid}")
	ctl("set", "env", "statefulset/publisher", "M8_EVENT_WORKER_TOPIC="+path, "M8_EVENT_WORKER_STATE="+state,
		"M8_EVENT_WORKER_TABLE="+path+"_owner", "M8_EVENT_WORKER_GUARD="+path+"_guard", "M8_EVENT_WORKER_SET_UID="+setUID)
	ctl("scale", "statefulset/publisher", "--replicas=1")
	waitPod := func() {
		ctl("wait", "--for=create", "pod/publisher-0", "--timeout=30s")
		ctl("wait", "--for=condition=Ready", "pod/publisher-0", "--timeout=30s")
	}
	waitPod()
	pod := func(field string) string { return ctl("get", "pod", "publisher-0", "-o", "jsonpath={"+field+"}") }
	oldUID, oldNode := pod(".metadata.uid"), pod(".spec.nodeName")
	if !validPodUID(oldUID) || !strings.HasPrefix(oldNode, "m8-event-handoff-worker") {
		t.Fatal("unexpected fixture Pod")
	}
	slot := ReplicaSlot{kubernetesFixtureNamespace, setUID, "publisher-0", path + "_owner"}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(readFile("ca.pem")) {
		t.Fatal("invalid Kubernetes test CA")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}
	defer transport.CloseIdleConnections()
	kube := &http.Client{Transport: bearerTransport{transport, strings.TrimSpace(string(readFile("reader-token")))}}
	api := strings.TrimSpace(string(readFile("server")))
	owner, err := NewTransactor(db, 8*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	delayedOwner, err := NewTransactor(db, 45*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	// Hold a real response containing the old UID across both assignments.
	observed, release := make(chan struct{}), make(chan struct{}, 1)
	var firstObservation atomic.Bool
	delayedKube := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		response, err := kube.Transport.RoundTrip(r)
		if err == nil && strings.Contains(r.URL.Path, "/pods/") && firstObservation.CompareAndSwap(false, true) {
			close(observed)
			select {
			case <-release:
			case <-r.Context().Done():
				_ = response.Body.Close()
				return nil, r.Context().Err()
			}
		}
		return response, err
	})}
	observe := func(client *http.Client) func(context.Context) (string, bool, error) {
		return func(ctx context.Context) (string, bool, error) {
			return kubernetes.ObserveReplica(ctx, client, api, slot.Namespace, "publisher", slot.StatefulSetUID, slot.PodName)
		}
	}
	delayedResult := make(chan error, 1)
	go func() { delayedResult <- slot.Reconcile(ctx, delayedOwner, observe(delayedKube)) }()
	defer close(release)
	select {
	case <-observed:
	case err := <-delayedResult:
		t.Fatalf("first observer exited before receiving the old Pod: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := slot.Reconcile(ctx, owner, observe(kube)); err != nil {
		t.Fatal(err)
	}
	call := func(node, endpoint string) string {
		t.Helper()
		if !strings.HasPrefix(node, "m8-event-handoff-worker") {
			t.Fatal("invalid fixture worker")
		}
		return command("docker", "exec", node, "curl", "--max-time", "65", "-sS", "http://127.0.0.1:18091/"+endpoint)
	}
	if got := call(oldNode, "bind"); got != oldUID {
		t.Fatalf("old worker bound %q", got)
	}
	first, next, sentinel := testEvent(t), testEvent(t), testEvent(t)
	first.Id, next.Id, sentinel.Id = uuid.NewString(), uuid.NewString(), uuid.NewString()
	if got := call(oldNode, "start?id="+first.Id); got != "held" {
		t.Fatalf("old writer: %s", got)
	}
	ctl("cordon", oldNode)
	controlIP := ctl("get", "node", "m8-event-handoff-control-plane", "-o", "jsonpath={.status.addresses[?(@.type=='InternalIP')].address}")
	if net.ParseIP(controlIP) == nil {
		t.Fatal("invalid control plane IP")
	}
	for _, rule := range [][]string{
		{"INPUT", "-s", controlIP, "-p", "tcp", "--sport", "6443", "-j", "DROP"},
		{"OUTPUT", "-d", controlIP, "-p", "tcp", "--dport", "6443", "-j", "DROP"},
	} {
		command("docker", append([]string{"exec", oldNode, "iptables", "-I"}, rule...)...)
		t.Cleanup(func() {
			_ = exec.Command("docker", append([]string{"exec", oldNode, "iptables", "-D"}, rule...)...).Run()
		})
	}
	probe, stop := context.WithTimeout(ctx, 4*time.Second)
	defer stop()
	if err := exec.CommandContext(probe, "docker", "exec", oldNode, "curl", "--insecure", "--max-time", "2", "-sS", "https://"+controlIP+":6443/livez").Run(); err == nil {
		t.Fatal("isolated node can still reach apiserver")
	}
	ctl("delete", "pod", "publisher-0", "--force", "--grace-period=0", "--wait=false")
	waitPod()
	newUID, newNode := pod(".metadata.uid"), pod(".spec.nodeName")
	if !validPodUID(newUID) || newUID == oldUID || newNode == oldNode {
		t.Fatal("expected new UID on a different worker")
	}
	if err := slot.Reconcile(ctx, owner, observe(kube)); err != nil {
		t.Fatal(err)
	}
	release <- struct{}{}
	if err := <-delayedResult; err != nil || delayedOwner.Stats().Retries == 0 {
		t.Fatalf("a delayed controller must retry against the current Pod: %v", err)
	}
	if got := call(newNode, "bind"); got != newUID {
		t.Fatalf("replacement bound %q", got)
	}
	if got := call(newNode, "write?id="+next.Id); got != "committed" {
		t.Fatalf("replacement write: %s", got)
	}
	if got := call(oldNode, "resume"); got != "stale" {
		t.Fatalf("old transaction: %s", got)
	}
	if got := call(oldNode, "bind"); got != "stale" {
		t.Fatalf("old Pod rebound: %s", got)
	}
	if got := call(oldNode, "claim"); got != "denied" {
		t.Fatalf("old Pod could change ownership: %s", got)
	}
	if got := call(newNode, "write?id="+sentinel.Id); got != "committed" {
		t.Fatalf("replacement reuse: %s", got)
	}
	crashAndRestart := func(endpoint, exitCode string, interrupted, after *eventv1.Event) {
		t.Helper()
		restarts := pod(".status.containerStatuses[0].restartCount")
		crash := exec.CommandContext(ctx, "docker", "exec", newNode, "curl", "--max-time", "15",
			"-fsS", "http://127.0.0.1:18091/"+endpoint+"?id="+interrupted.Id)
		if err := crash.Run(); err == nil {
			t.Fatal("publisher observed the commit response")
		}
		wait, cancel := context.WithTimeout(ctx, 55*time.Second)
		defer cancel()
		for pod(".status.containerStatuses[0].restartCount") == restarts {
			select {
			case <-wait.Done():
				t.Fatal("kubelet did not restart publisher")
			case <-time.After(250 * time.Millisecond):
			}
		}
		waitPod()
		if pod(".metadata.uid") != newUID || pod(".spec.nodeName") != newNode {
			t.Fatal("expected a new process in the very same Pod")
		}
		if code := pod(".status.containerStatuses[0].lastState.terminated.exitCode"); code != exitCode {
			t.Fatalf("publisher did not exit at the expected commit boundary: %s", code)
		}
		if got := call(newNode, "bind"); got != newUID {
			t.Fatalf("restarted process bound %q", got)
		}
		if got := call(newNode, "write?id="+after.Id); got != "committed" {
			t.Fatalf("publish after unknown outcome: %s", got)
		}
	}
	crashed, restarted := testEvent(t), testEvent(t)
	crashed.Id, restarted.Id = uuid.NewString(), uuid.NewString()
	crashAndRestart("crash", "42", crashed, restarted)
	t.Log("publisher process restarted with the same Pod UID after YDB commit but before the application received the response")
	events := []*eventv1.Event{next, sentinel, crashed, restarted}
	unknown, recovered := testEvent(t), testEvent(t)
	unknown.Id, recovered.Id = uuid.NewString(), uuid.NewString()
	crashAndRestart("crash-send", "43", unknown, recovered)
	// The restarted publisher commits before the test learns the old outcome.
	outcome, err := db.Query().QueryRow(ctx, "DECLARE $id AS Utf8; SELECT COUNT(*) FROM "+state+" WHERE id = $id;",
		query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(unknown.Id).Build()))
	if err != nil {
		t.Fatal(err)
	}
	var survived uint64
	if err := outcome.Scan(&survived); err != nil || survived > 1 {
		t.Fatalf("unexpected interrupted commit outcome: %d, %v", survived, err)
	}
	if survived == 1 {
		events = append(events, unknown)
	}
	events = append(events, recovered)
	t.Logf("publisher restarted after sending commit without observing any response; interrupted mutation committed=%t", survived == 1)
	beforeCommit := fixtureWorkerEvent(t, uuid.NewString())
	afterCrash := fixtureWorkerEvent(t, uuid.NewString())
	crashAndRestart("crash-before", "44", beforeCommit, afterCrash)
	events = append(events, afterCrash)
	row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	var count uint64
	if err := row.Scan(&count); err != nil || count != uint64(len(events)) {
		t.Fatalf("expected %d committed rows, got %d: %v", len(events), count, err)
	}
	for i, e := range events {
		row, err := db.Query().QueryRow(ctx, "DECLARE $id AS Utf8; SELECT id FROM "+state+" WHERE id = $id;", query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(e.Id).Build()))
		if err != nil {
			t.Fatal(err)
		}
		var got string
		if err := row.Scan(&got); err != nil || got != e.Id {
			t.Fatalf("expected replacement state row: %v", err)
		}
		events[i] = fixtureWorkerEvent(t, e.Id)
	}
	assertStableProducerEvents(t, ctx, db, path, fmt.Sprintf("m8-%x-0", sha256.Sum256([]byte(slot.key()))), events)
	// Keep the old binding in this process even when Kubernetes removes the Pod.
	oldClaim, err := slot.Bind(ctx, owner, newUID)
	if err != nil {
		t.Fatal(err)
	}
	oldPublisher, err := NewPublisher(owner, DefaultConfig(map[string]string{"events": path}), oldClaim, path+"_guard")
	if err != nil {
		t.Fatal(err)
	}
	ctl("scale", "statefulset/publisher", "--replicas=0")
	ctl("wait", "--for=delete", "pod/publisher-0", "--timeout=35s")
	if err := slot.Reconcile(ctx, owner, observe(kube)); err != nil {
		t.Fatal(err)
	}
	assertFenceRejected := func(p *Publisher) {
		t.Helper()
		if err := p.WithinTransaction(ctx, func(context.Context, query.TxActor, func(string, ...*eventv1.Event) error) error {
			t.Error("retired or vacant controller slot admitted a business mutation")
			return nil
		}); !errors.Is(err, ErrReplicaOwnership) {
			t.Fatalf("stale ownership: %v", err)
		}
	}
	assertFenceRejected(oldPublisher)
	ctl("scale", "statefulset/publisher", "--replicas=1")
	waitPod()
	scaledUID := pod(".metadata.uid")
	if scaledUID == newUID {
		t.Fatal("scale-up reused the old Pod")
	}
	restartedOwner, err := NewTransactor(db, 8*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := slot.Reconcile(ctx, restartedOwner, observe(kube)); err != nil {
		t.Fatal(err)
	}
	scaledClaim, err := slot.Bind(ctx, restartedOwner, scaledUID)
	if err != nil || scaledClaim.generation != oldClaim.generation+2 {
		t.Fatalf("scale-down and scale-up should advance two generations: %v", err)
	}
	scaledPublisher, err := NewPublisher(restartedOwner, DefaultConfig(map[string]string{"events": path}), scaledClaim, path+"_guard")
	if err != nil || scaledPublisher.producer != oldPublisher.producer {
		t.Fatalf("scale-up should reuse the producer: %v", err)
	}
	assertFenceRejected(oldPublisher)

	manifest := filepath.Join(fixture, "statefulset.json")
	tmpl := ctl("get", "statefulset/publisher", "-o", "json")
	var spec map[string]json.RawMessage
	if err := json.Unmarshal([]byte(tmpl), &spec); err != nil {
		t.Fatal(err)
	}
	spec["metadata"] = json.RawMessage(`{"name":"publisher"}`)
	delete(spec, "status")
	var settings map[string]json.RawMessage
	if err := json.Unmarshal(spec["spec"], &settings); err != nil {
		t.Fatal(err)
	}
	settings["replicas"] = json.RawMessage("0")
	spec["spec"], err = json.Marshal(settings)
	if err != nil {
		t.Fatal(err)
	}
	recreate, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(manifest, recreate, 0600); err != nil {
		t.Fatal(err)
	}
	ctl("delete", "statefulset", "publisher", "--wait=true", "--timeout=35s")
	ctl("wait", "--for=delete", "pod/publisher-0", "--timeout=35s")
	ctl("create", "-f", manifest)
	recreatedUID := ctl("get", "statefulset", "publisher", "-o", "jsonpath={.metadata.uid}")
	if recreatedUID == setUID {
		t.Fatal("controller incarnation did not change")
	}
	staged, finish := make(chan struct{}), make(chan struct{})
	oldResult := make(chan error, 1)
	doomed := fixtureWorkerEvent(t, uuid.NewString())
	go func() {
		attempted := false
		oldResult <- scaledPublisher.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor, send func(string, ...*eventv1.Event) error) error {
			if attempted {
				return errors.New("retired ownership reached the callback on retry")
			}
			attempted = true
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+state+" (id) VALUES ($id);",
				query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(doomed.Id).Build())); err != nil {
				return err
			}
			if err := send("events", doomed); err != nil {
				return err
			}
			close(staged)
			select {
			case <-finish:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	select {
	case <-staged:
	case err := <-oldResult:
		close(finish)
		t.Fatalf("old incarnation failed before retirement: %v", err)
	case <-ctx.Done():
		close(finish)
		t.Fatal(ctx.Err())
	}
	retireErr := slot.Reconcile(ctx, restartedOwner, observe(kube))
	close(finish)
	if retireErr != nil {
		t.Fatal(retireErr)
	}
	if err := <-oldResult; !errors.Is(err, ErrReplicaOwnership) {
		t.Fatalf("retirement did not fence the already staged attempt: %v", err)
	}
	if err := slot.Reconcile(ctx, restartedOwner, func(context.Context) (string, bool, error) {
		t.Error("retired incarnation tried to observe the Pod again")
		return scaledUID, false, nil
	}); err != nil {
		t.Fatal(err)
	}
	assertFenceRejected(scaledPublisher)
	verifyFence(t, ctx, restartedOwner, scaledPublisher, retiredReplica, scaledClaim.generation+1, uint64(len(events)))

	ctl("set", "env", "statefulset/publisher", "M8_EVENT_WORKER_SET_UID="+recreatedUID)
	ctl("scale", "statefulset/publisher", "--replicas=1")
	waitPod()
	freshSlot := slot
	freshSlot.StatefulSetUID = recreatedUID
	if err := freshSlot.Reconcile(ctx, restartedOwner, func(ctx context.Context) (string, bool, error) {
		return kubernetes.ObserveReplica(ctx, kube, api, slot.Namespace, "publisher", recreatedUID, slot.PodName)
	}); err != nil {
		t.Fatal(err)
	}
	freshUID := pod(".metadata.uid")
	freshNode := pod(".spec.nodeName")
	if got := call(freshNode, "bind"); got != freshUID {
		t.Fatalf("recreated replica bound %q", got)
	}
	freshEvent := fixtureWorkerEvent(t, uuid.NewString())
	if got := call(freshNode, "write?id="+freshEvent.Id); got != "committed" {
		t.Fatalf("recreated replica write: %s", got)
	}
	freshClaim, err := freshSlot.Bind(ctx, restartedOwner, freshUID)
	if err != nil {
		t.Fatal(err)
	}
	freshPublisher, err := NewPublisher(restartedOwner, DefaultConfig(map[string]string{"events": path}), freshClaim, path+"_guard")
	if err != nil || freshPublisher.producer == oldPublisher.producer {
		t.Fatalf("recreated StatefulSet must have its own producer identity: %v", err)
	}
	verifyFence(t, ctx, restartedOwner, freshPublisher, freshUID, 1, 1)
	reader, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := reader.Close(ctx); err != nil {
			t.Error(err)
		}
	}()
	for range events {
		if _, err := reader.ReadMessage(ctx); err != nil {
			t.Fatal(err)
		}
	}
	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := io.ReadAll(msg)
	if err != nil {
		t.Fatal(err)
	}
	got := new(eventv1.Event)
	if err := proto.Unmarshal(payload, got); err != nil || !proto.Equal(got, freshEvent) ||
		msg.ProducerID != freshPublisher.producer+"-0" || msg.Offset != int64(len(events)) {
		t.Fatalf("unexpected event at recreation boundary: %v", err)
	}
	ends, _, err := topicEnds(ctx, db, path)
	if err != nil || ends != int64(len(events)+1) {
		t.Fatalf("topic end after recreation: %d, %v", ends, err)
	}
	row, err = db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	if err := row.Scan(&count); err != nil || count != uint64(len(events)+1) {
		t.Fatalf("unexpected business state after recreation: %d, %v", count, err)
	}

	t.Logf("old UID %s was isolated on %s, replacement UID %s committed on %s", oldUID, oldNode, newUID, newNode)
}

type bearerTransport struct {
	next  http.RoundTripper
	token string
}

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func (b bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	copy := r.Clone(r.Context())
	copy.Header.Set("Authorization", "Bearer "+b.token)
	return b.next.RoundTrip(copy)
}

func fixtureWorkerEvent(t *testing.T, id string) *eventv1.Event {
	t.Helper()
	e := testEvent(t)
	e.Id = id
	e.OccurredAt.Seconds, e.OccurredAt.Nanos = 1700000000, 0
	return e
}

func TestKubernetesReplicaWorker(t *testing.T) {
	if os.Getenv("M8_EVENT_WORKER") != "1" {
		t.Skip("runs inside a disposable Kubernetes Pod")
	}
	uid := os.Getenv("POD_UID")
	ip := os.Getenv("M8_EVENT_WORKER_YDB_IP")
	if !validPodUID(uid) || net.ParseIP(ip) == nil {
		t.Fatal("invalid Pod fixture configuration")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	held := make(chan struct{})
	resume := make(chan struct{}, 1)
	var delay, paused atomic.Bool
	var crashBeforeCommit, crashOnCommit atomic.Bool
	var crashOnSend crashCommitSend
	intercept := func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, next grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		s, err := next(ctx, desc, conn, method, opts...)
		if err == nil && delay.Load() && method == Ydb_Topic_V1.TopicService_StreamWrite_FullMethodName {
			return heldProducerStream{s, held, resume, &paused}, nil
		}
		return s, err
	}
	unary := func(ctx context.Context, method string, req, reply any, conn *grpc.ClientConn, next grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		err := next(ctx, method, req, reply, conn, opts...)
		if err == nil && method == Ydb_Query_V1.QueryService_CommitTransaction_FullMethodName && crashOnCommit.Load() {
			if result, ok := reply.(*Ydb_Query.CommitTransactionResponse); ok && result.GetStatus() == Ydb.StatusIds_SUCCESS {
				os.Exit(42) // The application never observes the commit result.
			}
		}
		return err
	}
	dial := func(ctx context.Context, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "tcp", net.JoinHostPort(ip, "32536"))
	}
	password, err := os.ReadFile("/var/run/publisher-ydb/password")
	if err != nil {
		t.Fatal("read publisher fixture credentials")
	}
	db, err := ydb.Open(ctx, "grpc://"+net.JoinHostPort(ip, "32536")+"/local",
		ydb.WithStaticCredentials("eventfixturepublisher", string(password)),
		ydb.With(config.WithGrpcOptions(grpc.WithContextDialer(dial), grpc.WithChainStreamInterceptor(intercept),
			grpc.WithChainUnaryInterceptor(unary), grpc.WithStatsHandler(&crashOnSend))))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close(ctx)
	owner, err := NewTransactor(db, 55*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	var bound atomic.Pointer[Publisher]
	publish := func(id string) error {
		claim := bound.Load()
		if claim == nil {
			return ErrReplicaOwnership
		}
		return claim.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor, send func(string, ...*eventv1.Event) error) error {
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+os.Getenv("M8_EVENT_WORKER_STATE")+" (id) VALUES ($id);", query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(id).Build())); err != nil {
				return err
			}
			if err := send("events", fixtureWorkerEvent(t, id)); err != nil {
				return err
			}
			if crashBeforeCommit.Load() {
				os.Exit(44)
			}
			return nil
		})
	}
	respond := func(w http.ResponseWriter, err error) {
		switch {
		case err == nil:
			fmt.Fprintln(w, "committed")
		case errors.Is(err, ErrReplicaOwnership):
			fmt.Fprintln(w, "stale")
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/bind", func(w http.ResponseWriter, r *http.Request) {
		slot := ReplicaSlot{kubernetesFixtureNamespace, os.Getenv("M8_EVENT_WORKER_SET_UID"), "publisher-0", os.Getenv("M8_EVENT_WORKER_TABLE")}
		claim, err := slot.Bind(r.Context(), owner, uid)
		if err != nil {
			respond(w, err)
			return
		}
		publisher, err := NewPublisher(owner, DefaultConfig(map[string]string{"events": os.Getenv("M8_EVENT_WORKER_TOPIC")}), claim, os.Getenv("M8_EVENT_WORKER_GUARD"))
		if err != nil {
			respond(w, err)
			return
		}
		bound.Store(publisher)
		fmt.Fprintln(w, uid)
	})
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		delay.Store(true)
		id := r.URL.Query().Get("id")
		go func() { result <- publish(id) }()
		select {
		case <-held:
			fmt.Fprintln(w, "held")
		case err := <-result:
			respond(w, err)
		case <-r.Context().Done():
		}
	})
	mux.HandleFunc("/resume", func(w http.ResponseWriter, r *http.Request) {
		resume <- struct{}{}
		select {
		case err := <-result:
			respond(w, err)
		case <-r.Context().Done():
		}
	})
	mux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) { respond(w, publish(r.URL.Query().Get("id"))) })
	mux.HandleFunc("/crash-before", func(w http.ResponseWriter, r *http.Request) {
		crashBeforeCommit.Store(true)
		respond(w, publish(r.URL.Query().Get("id")))
	})
	mux.HandleFunc("/crash", func(w http.ResponseWriter, r *http.Request) {
		crashOnCommit.Store(true)
		respond(w, publish(r.URL.Query().Get("id")))
	})
	mux.HandleFunc("/crash-send", func(w http.ResponseWriter, r *http.Request) {
		crashOnSend.armed.Store(true)
		respond(w, publish(r.URL.Query().Get("id")))
	})
	mux.HandleFunc("/claim", func(w http.ResponseWriter, r *http.Request) {
		slot := ReplicaSlot{kubernetesFixtureNamespace, os.Getenv("M8_EVENT_WORKER_SET_UID"), "publisher-0", os.Getenv("M8_EVENT_WORKER_TABLE")}
		check, cancel := context.WithTimeout(r.Context(), 6*time.Second)
		defer cancel()
		var denied error
		err := db.Query().Do(check, func(ctx context.Context, session query.Session) error {
			denied = session.Exec(ctx, declareReplicaKey+"DECLARE $uid AS Utf8; "+
				"UPDATE `"+slot.TablePath+"` SET pod_uid = $uid WHERE replica = $replica;",
				query.WithParameters(ydb.ParamsBuilder().Param("$replica").Text(slot.key()).Param("$uid").Text(uid).Build()))
			return nil // Preserve the first authorization response without SDK retries.
		})
		accessDenied := ydb.IsOperationError(denied, Ydb.StatusIds_UNAUTHORIZED)
		// YDB 26.3 can wrap a table ACL rejection as an ABORTED resolve error.
		ydb.IterateByIssues(denied, func(message string, _ Ydb.StatusIds_StatusCode, _ uint32) {
			accessDenied = accessDenied || strings.Contains(message, "status: AccessDenied")
		})
		if err == nil && accessDenied {
			fmt.Fprintln(w, "denied")
			return
		}
		respond(w, errors.Join(err, denied))
	})
	server := &http.Server{Addr: "127.0.0.1:18091", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { <-ctx.Done(); _ = server.Close() }()
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		t.Fatal(err)
	}
}

type crashCommitSend struct{ armed atomic.Bool }

func (*crashCommitSend) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context { return ctx }
func (*crashCommitSend) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}
func (*crashCommitSend) HandleConn(context.Context, stats.ConnStats) {}

func (h *crashCommitSend) HandleRPC(_ context.Context, event stats.RPCStats) {
	if output, ok := event.(*stats.OutPayload); ok {
		if _, commit := output.Payload.(*Ydb_Query.CommitTransactionRequest); commit && h.armed.Load() {
			os.Exit(43)
		}
	}
}

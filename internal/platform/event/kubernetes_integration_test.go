package event

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-genproto/Ydb_Topic_V1"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/config"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"google.golang.org/grpc"
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
		"GRANT SELECT, UPDATE ROW ON " + state + ", `" + path + "`",
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
		"M8_EVENT_WORKER_TABLE="+path+"_owner", "M8_EVENT_WORKER_SET_UID="+setUID)
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
		if err == nil && firstObservation.CompareAndSwap(false, true) {
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
	delayedResult := make(chan error, 1)
	go func() { delayedResult <- slot.Reconcile(ctx, delayedOwner, delayedKube, api) }()
	defer close(release)
	select {
	case <-observed:
	case err := <-delayedResult:
		t.Fatalf("first observer exited before receiving the old Pod: %v", err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := slot.Reconcile(ctx, owner, kube, api); err != nil {
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
	if err := slot.Reconcile(ctx, owner, kube, api); err != nil {
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
	row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+state)
	if err != nil {
		t.Fatal(err)
	}
	var count uint64
	if err := row.Scan(&count); err != nil || count != 2 {
		t.Fatalf("expected exactly two committed rows, got %d: %v", count, err)
	}
	for _, e := range []*eventv1.Event{next, sentinel} {
		row, err := db.Query().QueryRow(ctx, "DECLARE $id AS Utf8; SELECT id FROM "+state+" WHERE id = $id;", query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(e.Id).Build()))
		if err != nil {
			t.Fatal(err)
		}
		var got string
		if err := row.Scan(&got); err != nil || got != e.Id {
			t.Fatalf("expected replacement state row: %v", err)
		}
	}
	assertStableProducerEvents(t, ctx, db, path, kubernetesFixtureProducer+"-0", []*eventv1.Event{fixtureWorkerEvent(t, next.Id), fixtureWorkerEvent(t, sentinel.Id)})
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

const kubernetesFixtureProducer = "m8.test.kubernetes.0.0"

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
	intercept := func(ctx context.Context, desc *grpc.StreamDesc, conn *grpc.ClientConn, method string, next grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		s, err := next(ctx, desc, conn, method, opts...)
		if err == nil && delay.Load() && method == Ydb_Topic_V1.TopicService_StreamWrite_FullMethodName {
			return heldProducerStream{s, held, resume, &paused}, nil
		}
		return s, err
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
		ydb.With(config.WithGrpcOptions(grpc.WithContextDialer(dial), grpc.WithChainStreamInterceptor(intercept))))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close(ctx)
	owner, err := NewTransactor(db, 55*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	result := make(chan error, 1)
	var bound atomic.Pointer[PodOwnership]
	publish := func(id string) error {
		claim := bound.Load()
		if claim == nil {
			return ErrReplicaOwnership
		}
		return owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
			if err := claim.Check(ctx, tx); err != nil {
				return err
			}
			if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+os.Getenv("M8_EVENT_WORKER_STATE")+" (id) VALUES ($id);", query.WithParameters(ydb.ParamsBuilder().Param("$id").Text(id).Build())); err != nil {
				return err
			}
			return writeWithProducer(ctx, db, tx, os.Getenv("M8_EVENT_WORKER_TOPIC"), kubernetesFixtureProducer, fixtureWorkerEvent(t, id))
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
		bound.Store(&claim)
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

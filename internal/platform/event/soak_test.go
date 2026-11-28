package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	eventv1 "github.com/m8-team/platform/internal/proto/m8/platform/event/v1"
	"github.com/ydb-platform/ydb-go-sdk/v3"
	"github.com/ydb-platform/ydb-go-sdk/v3/query"
	"github.com/ydb-platform/ydb-go-sdk/v3/topic/topicoptions"
	"google.golang.org/protobuf/proto"
)

// TestSoakProducerMetadata is opt-in because it drives the real server for
// longer than the normal live suite. It observes retained source IDs inside
// the partition tablet, independently of client producer IDs and process RSS.
func TestSoakProducerMetadata(t *testing.T) {
	if os.Getenv("M8_EVENT_TEST_SOAK_DURATION") == "" {
		t.Skip("set M8_EVENT_TEST_SOAK_DURATION and M8_EVENT_TEST_MONITOR_URL for a disposable YDB")
	}
	duration, err := time.ParseDuration(os.Getenv("M8_EVENT_TEST_SOAK_DURATION"))
	if err != nil || duration < time.Minute || duration > 30*time.Minute {
		t.Fatal("soak duration must be between 1m and 30m")
	}
	for _, mode := range []string{"random", "stable", "protected", "protected_churn"} {
		t.Run(mode, func(t *testing.T) { testSoakProducerMetadata(t, duration, mode) })
	}
}

func testSoakProducerMetadata(t *testing.T, duration time.Duration, mode string) {
	_, db, path := liveDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), duration+time.Minute)
	defer cancel()
	partitions := int64(1)
	switch os.Getenv("M8_EVENT_TEST_SOAK_PARTITIONS") {
	case "", "1":
	case "4":
		if mode != "protected" && mode != "protected_churn" {
			t.Fatal("four-partition soak requires a protected mode")
		}
		partitions = 4
	default:
		t.Fatal("soak partitions must be 1 or 4")
	}
	createTestTopic(t, ctx, db, path, partitions)
	owner, err := NewTransactor(db, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	var publishers []*Publisher
	var initial, restarted *Publisher
	var state string
	workerCount := 6
	churn := mode == "protected_churn"
	protected := mode == "protected" || churn
	if protected {
		workerCount = 4
		state = createTestTable(t, ctx, db, path+"_state", "id Utf8 NOT NULL, PRIMARY KEY (id)")
		cfg := DefaultConfig(map[string]string{"soak": path})
		first := livePublisher(t, ctx, db, path, cfg, 15*time.Second)
		publishers = append(publishers, first)
		for i := 1; i < workerCount; i++ {
			slot := first.ownership.slot
			slot.PodName = fmt.Sprintf("publisher-%d", i)
			publishers = append(publishers, bindTestPublisher(t, ctx, db, slot, first.guard, cfg, 15*time.Second))
		}
		if churn {
			initial = first
			_, restartDB, _ := liveDatabase(t)
			restartOwner, err := NewTransactor(restartDB, 15*time.Second)
			if err != nil {
				t.Fatal(err)
			}
			claim, err := first.ownership.slot.Bind(ctx, restartOwner, first.ownership.uid)
			if err != nil {
				t.Fatal(err)
			}
			restarted, err = newPublisher(restartOwner, cfg, claim, first.guard)
			if err != nil || restarted.producer != first.producer {
				t.Fatalf("new process did not reuse the slot identity: %v", err)
			}
		}
	}
	template := testEvent(t)

	maxCount, ttl, sample := observePartitionSourceIDs(t, ctx, path, int(partitions))
	t.Logf("tablet SourceIdMaxCounts=%d SourceIdLifetimeSeconds=%d", maxCount, ttl)
	if got := sample(ctx); got != 0 {
		t.Fatalf("empty fixture has %d retained source IDs", got)
	}

	var issued, commits atomic.Uint64
	started := time.Now()
	workCtx, stop := context.WithDeadline(ctx, started.Add(duration))
	defer stop()
	const targetRate = 40 // transactions per second on the one measured partition
	schedule := time.NewTicker(time.Second / targetRate)
	defer schedule.Stop()
	t.Logf("target_rate=%d transactions/s", targetRate)
	var workers sync.WaitGroup
	failures := make(chan error, workerCount)
	slotCommits := make([]uint64, workerCount)
	for lane := range workerCount {
		producer := ""
		if mode == "stable" {
			producer = fmt.Sprintf("m8.test.metadata.0.%d", lane)
		}
		workers.Go(func() {
			if churn && lane > 0 {
				delay := []time.Duration{0, duration / 4, duration / 2, duration / 2}[lane]
				enable := time.NewTimer(delay)
				defer enable.Stop()
				select {
				case <-workCtx.Done():
					return
				case <-enable.C:
					t.Logf("slot %d joined the protected load after %s", lane, time.Since(started).Round(time.Second))
				}
			}
			for {
				select {
				case <-workCtx.Done():
					return
				case <-schedule.C:
				}
				e := proto.CloneOf(template)
				index := issued.Add(1)
				e.Id = soakEventID(index)
				if partitions > 1 {
					e.AggregateId = fmt.Sprintf("aggregate-%d", index%128)
				}
				var err error
				if protected {
					if churn && lane == 0 && publishers[0] != restarted && time.Since(started) >= duration*3/4 {
						publishers[0] = restarted
						t.Logf("slot 0 restarted with its previous identity after %s", time.Since(started).Round(time.Second))
					}
					err = publishers[lane].WithinTransaction(ctx, func(ctx context.Context, tx *Transaction) error {
						if err := tx.Exec(ctx, "DECLARE $id AS Utf8; INSERT INTO "+state+" (id) VALUES ($id);",
							ydb.ParamsBuilder().Param("$id").Text(e.Id).Build()); err != nil {
							return err
						}
						return tx.Publish("soak", e)
					})
				} else {
					err = owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
						return writeWithProducer(ctx, db, tx, path, producer, e)
					})
				}
				if err != nil {
					failures <- err
					return
				}
				slotCommits[lane]++
				commits.Add(1)
			}
		})
	}
	defer func() {
		cancel()
		workers.Wait()
	}()
	done := make(chan struct{})
	go func() {
		workers.Wait()
		close(done)
	}()
	ticker := time.NewTicker(min(time.Minute, duration/3))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			before := commits.Load()
			t.Logf("elapsed=%s confirmed_before=%d retained=%d confirmed_after=%d",
				time.Since(started).Round(time.Second), before, sample(ctx), commits.Load())
		case <-done:
			if len(failures) != 0 {
				t.Fatalf("soak failed after %d commits: %v", commits.Load(), <-failures)
			}
			n := commits.Load()
			retained := sample(ctx)
			offset, _, err := topicEnds(ctx, db, path)
			t.Logf("elapsed=%s confirmed=%d retained=%d topic_end=%d", time.Since(started).Round(time.Millisecond), n, retained, offset)
			wantRetained := n
			if mode != "random" {
				wantRetained = uint64(workerCount) * uint64(partitions)
			}
			if n < 100 || offset != int64(n) || retained != int(wantRetained) {
				t.Fatalf("unexpected loss or early eviction: commits=%d, retained=%d, offset=%d, err=%v", n, retained, offset, err)
			}
			if churn && publishers[0] != restarted {
				t.Fatal("workload did not exercise a same-slot restart")
			}
			if protected {
				row, err := db.Query().QueryRow(ctx, "SELECT COUNT(*) FROM "+state)
				var count uint64
				if err != nil {
					t.Fatal(err)
				}
				if err := row.Scan(&count); err != nil || count != n {
					t.Fatalf("soak state count=%d, commits=%d, error=%v", count, n, err)
				}
				for i, p := range publishers {
					stats := p.owner.Stats()
					laneWait := p.Stats().LaneWait
					if churn && i == 0 {
						before := initial.owner.Stats()
						stats.Retries += before.Retries
						stats.CommitConflicts += before.CommitConflicts
						laneWait += initial.Stats().LaneWait
					}
					t.Logf("slot commits=%d retries=%d commit_conflicts=%d lane_wait=%s", slotCommits[i], stats.Retries, stats.CommitConflicts, laneWait)
					verifyFence(t, ctx, p.owner, p, p.ownership.uid, 1, slotCommits[i])
				}
			}
			reader, err := db.Topic().StartReader(testConsumer, topicoptions.ReadTopic(path))
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				closeCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
				defer stop()
				if err := reader.Close(closeCtx); err != nil {
					t.Error(err)
				}
			}()
			expected := make(map[string]string, n)
			for i := uint64(1); i <= n; i++ {
				aggregate := template.AggregateId
				if partitions > 1 {
					aggregate = fmt.Sprintf("aggregate-%d", i%128)
				}
				expected[soakEventID(i)] = aggregate
			}
			producers := make(map[string]struct{}, wantRetained)
			partitionOffsets := make(map[int64]int64, partitions)
			for range n {
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					t.Fatal(err)
				}
				data, err := io.ReadAll(msg)
				if err != nil {
					t.Fatal(err)
				}
				got := new(eventv1.Event)
				if err := proto.Unmarshal(data, got); err != nil {
					t.Fatal(err)
				}
				template.Id = got.Id
				var ok bool
				template.AggregateId, ok = expected[got.Id]
				partition := msg.PartitionID()
				offset := partitionOffsets[partition]
				if !ok || msg.Offset != offset || !proto.Equal(got, template) {
					t.Fatalf("unexpected or duplicate event %s at offset %d (want %d)", got.Id, msg.Offset, offset)
				}
				partitionOffsets[partition]++
				delete(expected, got.Id)
				producers[msg.ProducerID] = struct{}{}
			}
			if len(expected) != 0 || len(producers) != int(wantRetained) {
				t.Fatalf("missing events=%d, distinct producer IDs=%d, want=%d", len(expected), len(producers), wantRetained)
			}
			if len(partitionOffsets) != int(partitions) {
				t.Fatalf("read %d partitions, want %d", len(partitionOffsets), partitions)
			}
			t.Logf("verified %d event IDs and payloads across %d partitions, %d distinct producer IDs", n, partitions, len(producers))
			return
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
}

func soakEventID(index uint64) string {
	return fmt.Sprintf("34b169b7-3969-47f1-bc60-%012x", index)
}

// observeSourceIDs reads effective settings and counts records in the
// isolated partition tablet's monitoring page.
func observeSourceIDs(t *testing.T, ctx context.Context, path string) (maxCount int, ttlSeconds int64, count func(context.Context) int) {
	t.Helper()
	return observePartitionSourceIDs(t, ctx, path, 1)
}

func observePartitionSourceIDs(t *testing.T, ctx context.Context, path string, partitions int) (int, int64, func(context.Context) int) {
	t.Helper()
	monitor, err := url.Parse(os.Getenv("M8_EVENT_TEST_MONITOR_URL"))
	if err != nil || monitor.Scheme != "http" || monitor.Hostname() != "127.0.0.1" ||
		monitor.Port() == "" || monitor.User != nil || monitor.Path != "" ||
		monitor.RawQuery != "" || monitor.Fragment != "" {
		t.Fatal("soak monitor must be a bare http://127.0.0.1:<port> origin")
	}
	type tabletDescription struct {
		PathDescription struct {
			PersQueueGroup struct {
				Partitions []struct {
					TabletID string `json:"TabletId"`
				}
				PQTabletConfig struct {
					PartitionConfig struct {
						SourceIdMaxCounts       int
						SourceIdLifetimeSeconds json.Number
					}
				}
			}
		}
	}
	client := &http.Client{
		Timeout: 12 * time.Second,
		CheckRedirect: func(req *http.Request, _ []*http.Request) error {
			if req.URL.Scheme != monitor.Scheme || req.URL.Host != monitor.Host {
				return fmt.Errorf("tablet monitor redirected outside the fixture")
			}
			return nil
		},
	}
	description := monitorGET(t, ctx, client, monitor.String()+"/viewer/json/describe?path="+url.QueryEscape(path))
	var desc tabletDescription
	if err := json.Unmarshal(description, &desc); err != nil {
		t.Fatal(err)
	}
	group := desc.PathDescription.PersQueueGroup
	if len(group.Partitions) != partitions {
		t.Fatalf("expected %d isolated partitions, got %d", partitions, len(group.Partitions))
	}
	pages := make(map[string]struct{}, partitions)
	for _, partition := range group.Partitions {
		if partition.TabletID == "" {
			t.Fatal("fixture partition has no tablet")
		}
		pages[monitor.String()+"/tablets/app?TabletID="+url.QueryEscape(partition.TabletID)] = struct{}{}
	}
	if len(pages) != partitions {
		t.Fatal("fixture source sampling requires one tablet per partition")
	}
	settings := group.PQTabletConfig.PartitionConfig
	ttl, err := settings.SourceIdLifetimeSeconds.Int64()
	if err != nil || ttl <= 0 || settings.SourceIdMaxCounts <= 0 {
		t.Fatalf("invalid Source ID settings: %+v, %v", settings, err)
	}
	sample := func(ctx context.Context) int {
		var total int
		for page := range pages {
			html := monitorGET(t, ctx, client, page)
			_, after, ok := bytes.Cut(html, []byte("<caption>Source ids</caption>"))
			_, rows, hasBody := bytes.Cut(after, []byte("<tbody>"))
			rows, _, closed := bytes.Cut(rows, []byte("</tbody>"))
			if !ok || !hasBody || !closed {
				t.Fatal("partition tablet has no Source ids table")
			}
			total += bytes.Count(rows, []byte("<tr>"))
		}
		return total
	}
	return settings.SourceIdMaxCounts, ttl, sample
}

func monitorGET(t *testing.T, ctx context.Context, client *http.Client, endpoint string) []byte {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("monitor returned %s", resp.Status)
	}
	const maxBody = 32 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil || len(data) > maxBody {
		t.Fatalf("read monitor response: size=%d, err=%v", len(data), err)
	}
	return data
}

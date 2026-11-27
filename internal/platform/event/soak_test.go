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

	"github.com/ydb-platform/ydb-go-sdk/v3/query"
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
	monitor, err := url.Parse(os.Getenv("M8_EVENT_TEST_MONITOR_URL"))
	if err != nil || monitor.Scheme != "http" || monitor.Hostname() != "127.0.0.1" ||
		monitor.Port() == "" || monitor.User != nil || monitor.Path != "" ||
		monitor.RawQuery != "" || monitor.Fragment != "" {
		t.Fatal("soak monitor must be a bare http://127.0.0.1:<port> origin")
	}
	_, db, path := liveDatabase(t)
	ctx, cancel := context.WithTimeout(context.Background(), duration+time.Minute)
	defer cancel()
	createTestTopic(t, ctx, db, path, 1)
	p, err := NewPublisher(db, DefaultConfig(map[string]string{"soak": path}))
	if err != nil {
		t.Fatal(err)
	}
	owner, err := NewTransactor(db, 15*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	template := testEvent(t)

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
	if len(group.Partitions) != 1 || group.Partitions[0].TabletID == "" {
		t.Fatal("expected a single isolated PersQueue tablet")
	}
	settings := group.PQTabletConfig.PartitionConfig
	t.Logf("tablet SourceIdMaxCounts=%d SourceIdLifetimeSeconds=%s", settings.SourceIdMaxCounts, settings.SourceIdLifetimeSeconds)
	page := monitor.String() + "/tablets/app?TabletID=" + url.QueryEscape(group.Partitions[0].TabletID)
	sample := func() int {
		html := monitorGET(t, ctx, client, page)
		_, after, ok := bytes.Cut(html, []byte("<caption>Source ids</caption>"))
		_, rows, hasBody := bytes.Cut(after, []byte("<tbody>"))
		rows, _, closed := bytes.Cut(rows, []byte("</tbody>"))
		if !ok || !hasBody || !closed {
			t.Fatal("partition tablet has no Source ids table")
		}
		return bytes.Count(rows, []byte("<tr>"))
	}
	if got := sample(); got != 0 {
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
	const workerCount = 6
	failures := make(chan error, workerCount)
	for range workerCount {
		workers.Go(func() {
			for {
				select {
				case <-workCtx.Done():
					return
				case <-schedule.C:
				}
				e := proto.CloneOf(template)
				e.Id = fmt.Sprintf("34b169b7-3969-47f1-bc60-%012x", issued.Add(1))
				err := owner.WithinTransaction(ctx, func(ctx context.Context, tx query.TxActor) error {
					return p.Publish(ctx, tx, "soak", e)
				})
				if err != nil {
					failures <- err
					return
				}
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
				time.Since(started).Round(time.Second), before, sample(), commits.Load())
		case <-done:
			if len(failures) != 0 {
				t.Fatalf("soak failed after %d commits: %v", commits.Load(), <-failures)
			}
			n := commits.Load()
			retained := sample()
			offset, _, err := topicEnds(ctx, db, path)
			t.Logf("elapsed=%s confirmed=%d retained=%d topic_end=%d", time.Since(started).Round(time.Millisecond), n, retained, offset)
			if n < 100 || offset != int64(n) || retained < int(n) {
				t.Fatalf("unexpected loss or early eviction: commits=%d, retained=%d, offset=%d, err=%v", n, retained, offset, err)
			}
			if err != nil {
				t.Fatal(err)
			}
			return
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
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

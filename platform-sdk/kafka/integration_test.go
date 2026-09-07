package kafka

import (
	"context"
	"crypto/rand"
	"github.com/m8-team/platform-sdk/logging"
	"github.com/m8-team/platform-sdk/retry"
	"github.com/m8-team/platform-sdk/telemetry"
	"github.com/twmb/franz-go/pkg/kgo"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBrokerAcknowledgementAndCommit(t *testing.T) {
	brokers := os.Getenv("TEST_KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("TEST_KAFKA_BROKERS absent: broker integration not run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	providers, err := telemetry.New(ctx, logging.Info{Name: "test"}, telemetry.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer providers.Shutdown(context.Background())
	producer, err := NewProducer(providers, kgo.SeedBrokers(strings.Split(brokers, ",")...), kgo.RecordRetries(1), kgo.RecordDeliveryTimeout(5*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	defer producer.Client.Close()
	id := rand.Text()
	if err = producer.Send(ctx, Message{Topic: "product.events", Key: []byte(id), Value: []byte(id)}); err != nil {
		t.Fatal(err)
	}
	client, err := NewConsumer(kgo.SeedBrokers(strings.Split(brokers, ",")...), kgo.ConsumerGroup("sdk-test-"+id), kgo.ConsumeTopics("product.events"), kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	var record *kgo.Record
	for record == nil && ctx.Err() == nil {
		fetch := client.PollRecords(ctx, 1)
		if errs := fetch.Errors(); len(errs) > 0 {
			t.Fatal(errs[0].Err)
		}
		for _, r := range fetch.Records() {
			if string(r.Key) == id {
				record = r
			}
		}
		if record == nil {
			client.AllowRebalance()
		}
	}
	if record == nil {
		t.Fatal("published record not fetched")
	}
	calls := 0
	consumer := Consumer{Client: client, Telemetry: providers, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)), Timeout: time.Second, Retry: retry.Policy{MaxAttempts: 1}, Handler: func(context.Context, Message) error { calls++; return nil }}
	err = consumer.process(ctx, record)
	client.AllowRebalance()
	if err != nil || calls != 1 {
		t.Fatalf("calls=%d err=%v", calls, err)
	}
	// CommitRecords must update the group's committed position past this record.
	if got := client.CommittedOffsets()[record.Topic][record.Partition].Offset; got != record.Offset+1 {
		t.Fatalf("offset=%d want=%d", got, record.Offset+1)
	}
}

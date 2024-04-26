package main

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/phuslu/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kfake"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kphuslog"
)

func TestCollector(t *testing.T) {
	log.DefaultLogger.SetLevel(log.InfoLevel)
	if testing.Verbose() {
		log.DefaultLogger.SetLevel(log.DebugLevel)
	}

	numBrokers := 3

	c, err := kfake.NewCluster(
		kfake.NumBrokers(numBrokers),
		kfake.DefaultNumPartitions(3),
		kfake.SeedTopics(0, "topic1", "topic2"),
	)

	if err != nil {
		t.Fatal(err, "failed to create cluster")
	}

	defer c.Close()

	var conf Config
	for _, broker := range c.ListenAddrs() {
		conf.Kafka.Servers = append(conf.Kafka.Servers, Address(broker))
	}

	client, err := kgo.NewClient(append(
		franz(conf).Opts(),
		kgo.ConsumerGroup("dummy-cg"),
		kgo.ConsumeTopics("topic1"),
		kgo.WithLogger(kphuslog.New(&log.Logger{Level: log.ErrorLevel})),
	)...)
	if err != nil {
		t.Fatal(err, "failed to create client")
	}

	ctx := context.Background()
	e := NewExporter(conf)

	// produce something and consume it
	_ = client.ProduceSync(ctx, &kgo.Record{Topic: "topic1", Value: []byte("hello")}).FirstErr()
	_ = client.ProduceSync(ctx, &kgo.Record{Topic: "topic1", Value: []byte("world")}).FirstErr()

	if err := client.ProduceSync(ctx, &kgo.Record{Topic: "topic2", Value: []byte("world")}).FirstErr(); err != nil {
		t.Fatal(err, "failed to produce")
	}

	fetches := client.PollRecords(ctx, 1)
	if err := fetches.Err(); err != nil {
		t.Fatal(err, "failed to poll")
	}

	fetches.EachRecord(func(r *kgo.Record) {
		t.Log("record", string(r.Value))
	})

	if err := client.CommitRecords(ctx, fetches.Records()...); err != nil {
		t.Fatal(err, "failed to commit")
	}

	time.Sleep(100 * time.Millisecond)

	resp, err := kadm.NewClient(client).FetchOffsets(ctx, "dummy-cg")
	if err != nil {
		t.Fatal(err, "failed to fetch offsets")
	}

	resp.Each(func(r kadm.OffsetResponse) {
		t.Log("offset", r.Topic, r.Partition, r.Offset.At)
	})

	if err := e.export(ctx); err != nil {
		t.Fatal(err, "failed to export")
	}

	metricFamily, err := e.metrics.reg.Gather()
	if err != nil {
		t.Fatal(err, "failed to gather metrics")
	}

	if len(metricFamily) == 0 {
		t.Fatal("no metrics found")
	}

	for _, mf := range metricFamily {
		if *mf.Name == "kafka_brokers" {
			if mf.Metric[0].Gauge.GetValue() != float64(numBrokers) {
				t.Fatal("expected 3 brokers, got", mf.Metric[0].Gauge.GetValue())
			}
		}
	}

	trw, treq := httptest.NewRecorder(), httptest.NewRequest("GET", "/metrics", nil)
	promhttp.HandlerFor(e.metrics.reg, promhttp.HandlerOpts{Registry: e.metrics.reg}).ServeHTTP(trw, treq)

	t.Log(trw.Body.String())
}

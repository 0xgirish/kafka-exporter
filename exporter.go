package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/0xgirish/kafka-exporter/pkg/fail"
	"github.com/phuslu/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/twmb/franz-go/pkg/kadm"
)

type exporter struct {
	d time.Duration

	metrics *metrics
	client  *kadm.Client

	onErrors fail.OnErrors

	config Config

	clientRefreshTime time.Time
}

func NewExporter(conf Config) *exporter {
	reg := prometheus.NewRegistry()

	return &exporter{
		d: conf.RefreshInterval,

		metrics: newMetrics(reg),
		client:  conf.Franzgo(),

		onErrors:          fail.OnErrors{Max: conf.ContinuousFailures},
		config:            conf,
		clientRefreshTime: time.Now(),
	}
}

func (e *exporter) Start(ctx context.Context) error {
	t := time.NewTicker(e.d)
	defer t.Stop()

	for {
		if e.onErrors.Fail() {
			// if we have too many errors, we should stop collecting metrics and fail the container
			// so sre can investigate the issue
			return fmt.Errorf("too many errors! recent: %w", e.onErrors.Recent())
		}

		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := e.export(context.Background()); err != nil {
				e.onErrors.Record(err)
				log.Error().Err(err).Msg("failed to export metrics")
				continue
			}

			e.onErrors.Record(nil)

			if e.onErrors.Failing() && time.Since(e.clientRefreshTime) > 2*time.Minute {
				log.Warn().Err(e.onErrors.Recent()).Msg("failing, re-initializing client")
				e.client.Close()
				e.client = e.config.Franzgo()
				e.clientRefreshTime = time.Now()
			}
		}
	}
}

func (e *exporter) export(ctx context.Context) error {
	defer e.Recover()

	metadata, err := e.client.Metadata(ctx)
	if err != nil {
		return err
	}

	// broker metrics
	e.metrics.broker.controller.Set(float64(metadata.Controller))
	e.metrics.broker.brokers.Set(float64(len(metadata.Brokers)))
	for _, broker := range metadata.Brokers {
		rackID := "unknown"
		if broker.Rack != nil {
			rackID = *broker.Rack
		}

		e.metrics.broker.brokerInfo.With(prometheus.Labels{
			"id":      strconv.Itoa(int(broker.NodeID)),
			"address": fmt.Sprintf("%s:%d", broker.Host, broker.Port),
			"rack":    rackID,
		}).Set(1)
	}

	// topic metrics, except for offsets
	topics := make([]string, 0, len(metadata.Topics))
	for _, topic := range metadata.Topics {
		topics = append(topics, topic.Topic)

		e.metrics.topic.partitions.With(prometheus.Labels{
			"topic": topic.Topic,
		}).Set(float64(len(topic.Partitions)))

		for _, partition := range topic.Partitions {

			if partition.Err != nil {
				log.Error().Err(partition.Err).Msg("failed to get partition info")
				e.onErrors.Record(partition.Err)
				continue
			}

			e.metrics.topic.partitionLeader.With(prometheus.Labels{
				"topic":     topic.Topic,
				"partition": strconv.Itoa(int(partition.Partition)),
			}).Set(float64(partition.Leader))

			e.metrics.topic.partitionReplicas.With(prometheus.Labels{
				"topic":     topic.Topic,
				"partition": strconv.Itoa(int(partition.Partition)),
			}).Set(float64(len(partition.Replicas)))

			e.metrics.topic.partitionISR.With(prometheus.Labels{
				"topic":     topic.Topic,
				"partition": strconv.Itoa(int(partition.Partition)),
			}).Set(float64(len(partition.ISR)))

			e.metrics.topic.partitionUnderRep.With(prometheus.Labels{
				"topic":     topic.Topic,
				"partition": strconv.Itoa(int(partition.Partition)),
			}).Set(float64(len(partition.OfflineReplicas)))

			isPreferred := 0
			if len(partition.Replicas) > 0 && partition.Leader == partition.Replicas[0] {
				isPreferred = 1
			}

			e.metrics.topic.partitionLeaderIsPreferred.With(prometheus.Labels{
				"topic":     topic.Topic,
				"partition": strconv.Itoa(int(partition.Partition)),
			}).Set(float64(isPreferred))
		}
	}

	// offset metrics
	offsetsMetrics := func(listOffsets func(ctx context.Context, topics ...string) (kadm.ListedOffsets, error), isEnd bool) {
		topicOffsets, err := listOffsets(ctx, topics...)
		if err != nil {
			log.Error().Err(err).Msg("failed to list end offsets")
			e.onErrors.Record(err)
			return
		}

		for _, offsets := range topicOffsets {
			for _, offset := range offsets {
				if offset.Err != nil {
					log.Error().Err(offset.Err).Msg("failed to get end offset")
					e.onErrors.Record(offset.Err)
					continue
				}

				metric := e.metrics.topic.partitionOldestOffset
				if isEnd {
					metric = e.metrics.topic.partitionCurrentOffset
				}

				metric.With(prometheus.Labels{
					"topic":     offset.Topic,
					"partition": strconv.Itoa(int(offset.Partition)),
				}).Set(float64(offset.Offset))
			}
		}
	}

	offsetsMetrics(e.client.ListEndOffsets, true)
	offsetsMetrics(e.client.ListStartOffsets, false)

	// consumer group metrics
	groupLags, err := e.client.Lag(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to get consumer group lags")
		return err
	}

	for _, groupLag := range groupLags {
		if groupLag.FetchErr != nil || groupLag.DescribeErr != nil {
			e.onErrors.Record(groupLag.FetchErr)
			e.onErrors.Record(groupLag.DescribeErr)

			log.Error().
				AnErr("fetch_err", groupLag.FetchErr).
				AnErr("describe_err", groupLag.DescribeErr).
				Msg("failed to get consumer group lag")
			continue
		}

		e.metrics.group.members.With(prometheus.Labels{
			"consumergroup": groupLag.Group,
		}).Set(float64(len(groupLag.Members)))

		e.metrics.group.coordinator.With(prometheus.Labels{
			"consumergroup": groupLag.Group,
		}).Set(float64(groupLag.Coordinator.NodeID))

		if len(groupLag.Lag) == 0 {
			log.Warn().Str("consumergroup", groupLag.Group).Msg("no lag information found for consumer group")
			continue
		}

		for _, memberLags := range groupLag.Lag {
			for _, memberLag := range memberLags {
				if memberLag.Err != nil {
					log.Error().Err(memberLag.Err).Msg("failed to get consumer group lag")
					e.onErrors.Record(memberLag.Err)
					continue
				}

				e.metrics.group.lag.With(prometheus.Labels{
					"consumergroup": groupLag.Group,
					"topic":         memberLag.Topic,
					"partition":     strconv.Itoa(int(memberLag.Partition)),
				}).Set(float64(memberLag.Lag))

				e.metrics.group.currentOffset.With(prometheus.Labels{
					"consumergroup": groupLag.Group,
					"topic":         memberLag.Topic,
					"partition":     strconv.Itoa(int(memberLag.Partition)),
				}).Set(float64(memberLag.Commit.At))
			}
		}
	}

	return nil
}

func (e *exporter) Recover() {
	if r := recover(); r != nil {
		e.onErrors.Record(fmt.Errorf("panic: %v", r))
		log.Error().Stack().Msgf("Recovered from panic: %v", r)
	}
}

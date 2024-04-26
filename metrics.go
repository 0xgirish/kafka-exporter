package main

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	broker brokerMetrics
	topic  topicMetrics
	group  consumerGroupMetrics

	reg *prometheus.Registry
}

func newMetrics(reg *prometheus.Registry) *metrics {
	m := &metrics{
		broker: brokerMetrics{
			brokers: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "kafka_brokers",
				Help: "Number of brokers in the Kafka cluster",
			}),
			brokerInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_broker_info",
				Help: "Information about the broker (node_id, host, rack_id* (if present) )",
			}, []string{"id", "address", "rack"}),
			controller: prometheus.NewGauge(prometheus.GaugeOpts{
				Name: "kafka_broker_controller",
				Help: "ID of the broker that is currently the controller for the Kafka cluster",
			}),
		},
		topic: topicMetrics{
			partitions: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partitions",
				Help: "Number of partitions for this Topic",
			}, []string{"topic"}),
			partitionLeader: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partition_leader",
				Help: "ID of the broker that is currently the leader for this Topic/Partition",
			}, []string{"topic", "partition"}),
			partitionReplicas: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partition_replicas",
				Help: "Number of Replicas for this Topic/Partition",
			}, []string{"topic", "partition"}),
			partitionISR: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partition_in_sync_replicas",
				Help: "Number of In-Sync Replicas for this Topic/Partition",
			}, []string{"topic", "partition"}),
			partitionUnderRep: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partition_under_replicated_partition",
				Help: "1 if Topic/Partition is under Replicated, 0 otherwise",
			}, []string{"topic", "partition"}),
			partitionLeaderIsPreferred: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partition_leader_is_preferred",
				Help: "1 if the current broker is the preferred leader for this Topic/Partition, 0 otherwise",
			}, []string{"topic", "partition"}),
			partitionCurrentOffset: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partition_current_offset",
				Help: "Current Offset of a Topic/Partition",
			}, []string{"topic", "partition"}),
			partitionOldestOffset: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_partition_oldest_offset",
				Help: "Oldest Offset of a Topic/Partition",
			}, []string{"topic", "partition"}),
			isInternal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_topic_is_internal",
				Help: "1 if the Topic is an internal Topic, 0 otherwise",
			}, []string{"topic"}),
		},
		group: consumerGroupMetrics{
			members: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_consumergroup_members",
				Help: "Number of members in the consumer group",
			}, []string{"consumergroup"}),
			coordinator: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_consumergroup_coordinator",
				Help: "ID of the broker that is currently the coordinator for the consumer group",
			}, []string{"consumergroup"}),
			lag: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_consumergroup_lag",
				Help: "Current Approximate Lag of a ConsumerGroup at Topic/Partition",
			}, []string{"consumergroup", "topic", "partition"}),
			currentOffset: prometheus.NewGaugeVec(prometheus.GaugeOpts{
				Name: "kafka_consumergroup_current_offset",
				Help: "Current Offset of a ConsumerGroup at Topic/Partition",
			}, []string{"consumergroup", "topic", "partition"}),
		},
		reg: reg,
	}

	m.register(reg)
	return m
}

func (m *metrics) register(reg *prometheus.Registry) {
	reg.MustRegister(
		m.broker.brokers,
		m.broker.brokerInfo,
		m.broker.controller,
		m.topic.partitions,
		m.topic.partitionLeader,
		m.topic.partitionReplicas,
		m.topic.partitionISR,
		m.topic.partitionUnderRep,
		m.topic.partitionLeaderIsPreferred,
		m.topic.partitionCurrentOffset,
		m.topic.partitionOldestOffset,
		m.topic.isInternal,
		m.group.members,
		m.group.coordinator,
		m.group.lag,
		m.group.currentOffset,
	)
}

type brokerMetrics struct {
	brokers    prometheus.Gauge
	brokerInfo *prometheus.GaugeVec
	controller prometheus.Gauge
}

type topicMetrics struct {
	partitions                 *prometheus.GaugeVec
	partitionReplicas          *prometheus.GaugeVec
	partitionISR               *prometheus.GaugeVec
	partitionUnderRep          *prometheus.GaugeVec
	partitionLeader            *prometheus.GaugeVec
	partitionLeaderIsPreferred *prometheus.GaugeVec
	partitionCurrentOffset     *prometheus.GaugeVec
	partitionOldestOffset      *prometheus.GaugeVec
	isInternal                 *prometheus.GaugeVec
}

type consumerGroupMetrics struct {
	members       *prometheus.GaugeVec
	coordinator   *prometheus.GaugeVec
	lag           *prometheus.GaugeVec
	currentOffset *prometheus.GaugeVec
}

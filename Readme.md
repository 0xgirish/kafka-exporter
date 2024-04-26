# Kafka Exporter
Kafka exporter for Prometheus.

## Usage
```sh
$ kafka-exporter --help
Kafka exporter for Prometheus.
Usage: kafka-exporter --kafka.servers BROKER_ADDRESS [--sasl.enabled] [--sasl.username SASL.USERNAME] [--sasl.password SASL.PASSWORD] [--sasl.mechanism SASL.MECHANISM] [--tls.enabled] [--tls.insecure-skip-tls-verify] [--listen.address ADDRESS] [--refresh.interval DURATION] [--continuous.failures CONTINUOUS.FAILURES] [--log.level LOG.LEVEL]

Options:
  --kafka.servers BROKER_ADDRESS
                         Address of the Kafka brokers
  --sasl.enabled         Enable SASL authentication [default: false]
  --sasl.username SASL.USERNAME
                         Username for SASL authentication [env: SASL_USERNAME]
  --sasl.password SASL.PASSWORD
                         Password for SASL authentication [env: SASL_PASSWORD]
  --sasl.mechanism SASL.MECHANISM
                         SASL mechanism to use [default: PLAIN]
  --tls.enabled          Enable TLS [default: false]
  --tls.insecure-skip-tls-verify
                         Skip TLS verification [default: false]
  --listen.address ADDRESS
                         Address to listen on for serving Prometheus metrics [default: :9308]
  --refresh.interval DURATION
                         Interval at which to refresh the metrics from Kafka [default: 30s]
  --continuous.failures CONTINUOUS.FAILURES
                         Number of continuous failures before exiting [default: 10]
  --log.level LOG.LEVEL
                         Log level [default: debug]
  --help, -h             display this help and exit
```

## Metrics

### Broker
1. `kafka_brokers` - Number of brokers in the Kafka cluster
2. `kafka_broker_info` - Information about the broker (node_id, host, rack_id)
3. `kafka_broker_controller` - Broker ID of the controller

### Topic
1. `kafka_topic_partitions` - Number of partitions for a topic
2. `kafka_topic_partition_replicas` - Number of replicas for a partition
3. `kafka_topic_partition_leader` - Leader of a partition
4. `kafka_topic_partition_in_sync_replicas` - Number of in-sync replicas for a partition
5. `kafka_topic_partition_leader_is_preferred` - Whether the leader is preferred for a partition
6. `kafka_topic_partition_under_replicated_partition` - Whether a partition is under replicated
7. `kafka_topic_partition_current_offset` - Current offset of a partition
8. `kafka_topic_partition_oldest_offset` - Oldest offset of a partition
9. `kafka_topic_is_internal` - Whether a topic is internal

### Consumer Group
1. `kafka_consumergroup_current_offset` - Current offset of a consumer group
2. `kafka_consumergroup_lag` - Lag of a consumer group
3. `kafka_consumergroup_coordinator` - Broker ID of the coordinator for a consumer group
4. `kafka_consumergroup_members` - Number of members in a consumer group
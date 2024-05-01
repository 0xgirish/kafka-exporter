module github.com/0xgirish/kafka-exporter

go 1.22

replace github.com/twmb/franz-go/pkg/kadm => github.com/0xgirish/franz-go/pkg/kadm v1.11.1

require (
	github.com/alexflint/go-arg v1.4.3
	github.com/phuslu/log v1.0.92
	github.com/prometheus/client_golang v1.19.0
	github.com/twmb/franz-go v1.16.1
	github.com/twmb/franz-go/pkg/kadm v1.11.0
	github.com/twmb/franz-go/pkg/kfake v0.0.0-20240412162337-6a58760afaa7
	github.com/twmb/franz-go/plugin/kphuslog v1.0.0
)

require (
	github.com/alexflint/go-scalar v1.1.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.19 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.7.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)

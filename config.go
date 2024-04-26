package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/phuslu/log"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
	"github.com/twmb/franz-go/pkg/sasl/scram"
	"github.com/twmb/franz-go/plugin/kphuslog"
)

type Config struct {
	Kafka

	ListenAddress      Address       `arg:"--listen.address" help:"Address to listen on for serving Prometheus metrics" default:":9308" placeholder:"ADDRESS"`
	RefreshInterval    time.Duration `arg:"--refresh.interval" help:"Interval at which to refresh the metrics from Kafka" default:"30s" placeholder:"DURATION"`
	ContinuousFailures int           `arg:"--continuous.failures" help:"Number of continuous failures before exiting" default:"10"`
	LogLevel           string        `arg:"--log.level" help:"Log level" default:"debug"`
}

type Kafka struct {
	Servers []Address `arg:"--kafka.servers,required" help:"Address of the Kafka brokers" placeholder:"BROKER_ADDRESS"`
	SASL
	TLS
}

type SASL struct {
	Enabled   bool   `arg:"--sasl.enabled" help:"Enable SASL authentication" default:"false"`
	Username  string `arg:"--sasl.username,env:SASL_USERNAME" help:"Username for SASL authentication"`
	Password  string `arg:"--sasl.password,env:SASL_PASSWORD" help:"Password for SASL authentication"`
	Mechanism string `arg:"--sasl.mechanism" help:"SASL mechanism to use" default:"PLAIN"`
}

type TLS struct {
	Enabled               bool `arg:"--tls.enabled" help:"Enable TLS" default:"false"`
	InsecureSkipTLSVerify bool `arg:"--tls.insecure-skip-tls-verify" help:"Skip TLS verification" default:"false"`
}

func (Config) Description() string {
	return "Kafka exporter for Prometheus."
}

func (c Config) Franzgo() *kadm.Client {
	return kadm.NewClient(franz(c))
}

type Address string

func (s *Address) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid server address: %s", text)
	}

	host := parts[0]
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid port number: %s", parts[1])
	}

	*s = Address(fmt.Sprintf("%s:%d", host, port))
	return nil
}

func franz(config Config) *kgo.Client {
	brokers := make([]string, 0, len(config.Servers))
	for _, server := range config.Servers {
		brokers = append(brokers, string(server))
	}

	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.WithLogger(kphuslog.New(&log.Logger{Level: log.InfoLevel})),
	}

	if config.SASL.Enabled {
		switch config.SASL.Mechanism {
		case "PLAIN":
			opts = append(opts, kgo.SASL(plain.Auth{
				User: config.SASL.Username,
				Pass: config.SASL.Password,
			}.AsMechanism()))
		case "SCRAM-SHA-256":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: config.SASL.Username,
				Pass: config.SASL.Password,
			}.AsSha256Mechanism()))
		case "SCRAM-SHA-512":
			opts = append(opts, kgo.SASL(scram.Auth{
				User: config.SASL.Username,
				Pass: config.SASL.Password,
			}.AsSha512Mechanism()))
		}
	}

	if config.TLS.Enabled {
		opts = append(opts, kgo.DialTLSConfig(&tls.Config{
			InsecureSkipVerify: config.TLS.InsecureSkipTLSVerify,
			ClientAuth:         tls.NoClientCert,
		}))
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		log.Panic().Err(err).Msg("failed to create Kafka client")
	}

	if err := client.Ping(context.Background()); err != nil {
		log.Panic().Err(err).Msg("failed to ping Kafka server")
	}

	return client
}

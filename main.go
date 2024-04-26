package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/0xgirish/kafka-exporter/pkg/sighandler"
	"github.com/alexflint/go-arg"
	"github.com/phuslu/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			// if containers are restarted too quickly, we might make too many requests to the Kafka server,
			// so we sleep for a while
			time.Sleep(10 * time.Second)
			log.Panic().Any("recover", r).Msg("recovered from panic")
		}
	}()

	var config Config
	arg.MustParse(&config)

	switch config.LogLevel {
	case "debug":
		log.DefaultLogger.SetLevel(log.DebugLevel)
	case "info":
		log.DefaultLogger.SetLevel(log.InfoLevel)
	case "warn", "warning":
		log.DefaultLogger.SetLevel(log.WarnLevel)
	case "error":
		log.DefaultLogger.SetLevel(log.ErrorLevel)
	default:
		log.DefaultLogger.SetLevel(log.DebugLevel)
	}

	ctx := sighandler.WithCancelOnSigInt(context.Background())
	exporter := NewExporter(config)

	server := &http.Server{
		Addr: string(config.ListenAddress),
		Handler: func() http.Handler {
			mux := http.NewServeMux()
			mux.Handle("/metrics",
				promhttp.HandlerFor(exporter.metrics.reg, promhttp.HandlerOpts{Registry: exporter.metrics.reg}))
			return mux
		}(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Panic().Err(err).Msg("failed to start server")
		}
	}()

	// blocking call
	if err := exporter.Start(ctx); err != nil {
		log.Panic().Err(err).Msg("failure while collecting metrics")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Panic().Err(err).Msg("failed to gracefully shutdown server")
	}
}

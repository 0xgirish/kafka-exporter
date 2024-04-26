package sighandler

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/phuslu/log"
)

// WithCancelOnSigInt creates a child context that can be cancelled when
// a SIGINT signal (Ctrl+C) is received. The parent context is passed as
// an argument and the function returns the created child context.
//
// The function starts a goroutine to listen for the SIGINT signal and
// calls the cancel function of the child context when the signal is received.
//
// Example usage:
//
//	parentCtx := context.Background()
//	childCtx := WithCancelOnSigInt(parentCtx)
//	// use childCtx in your code...
func WithCancelOnSigInt(parent context.Context) context.Context {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancelCause(parent)

	// Create a channel to receive SIGINT (Ctrl+C) signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start a goroutine that will call the cancel function
	// upon receiving a SIGINT signal
	go func() {
		signal := <-sigCh
		log.Warn().Any("signal", signal).Msg("received termination signal, cancelling context")
		cancel(errors.New("received termination signal, cancelling context"))
	}()

	return ctx
}

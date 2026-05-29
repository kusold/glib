package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// SignalContext returns a context canceled when one of the provided signals is
// received. When no signals are provided, os.Interrupt and SIGTERM are used.
func SignalContext(ctx context.Context, signals ...os.Signal) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
	}
	return signal.NotifyContext(ctx, signals...)
}

package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Handler is catch SIGINT and SIGTERM signal.
func Handler(ctx context.Context, done chan struct{}) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		signal.Reset()
		return nil
	case <-done:
		return nil
	case sig := <-sigCh:
		return fmt.Errorf("signal received: %s", sig.String())
	}
}

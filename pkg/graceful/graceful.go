// package graceful controls execution and graceful termination of a group of goroutines
// that must gracefully terminate in the frame of defined timeout.
//
// Candidate to move to separate open-source repository.
package graceful

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Config contains configuration options for the graceful Run function.
//
// It is useful as zero value.
type Config struct {
	// Timeout to wait for graceful termination, default 5 seconds.
	Timeout time.Duration
	// Optional context to base on. For example if you want to control
	// Cancellation of the goroutines group externally.
	Context *context.Context
	// Signals to hook on, by default SIGINT, SIGTERM, SIGHUP.
	Signals []os.Signal
}

// ErrUngraceful is returned by Run if it was finished ungracefully.
//
// OriginalError will contain the error of the first goroutine that
// exits (may be nil).
type ErrUngraceful struct {
	OriginalError error
}

func (e ErrUngraceful) Error() string {
	if e.OriginalError == nil {
		return ungracefulMessage
	}
	return fmt.Sprintf("%v; %v", e.OriginalError, ungracefulMessage)
}

const (
	defaultTimeout    = 5 * time.Second
	ungracefulMessage = "one ore more goroutines failed to gracefully finish"
)

// Run runs a set of goroutines that accept context and return error.
// On interruption signal it will cancel context passed to funcs signalling them to finish.
// If any of the routines returns with error, the context is cancelled as well.
//
// There's a timeout to wait for all routines to terminate after which Run returns.
// If any of the routines failed to terminate in time, Run will return ErrUngraceful,
// which contains original error that caused the termination.
// Otherwise it will just return original error (which can be nil).
//
// Example:
//
// err := graceful.Run(graceful.Config{Timeout: 10 * time.Second}, server, scheduler)
//
// Blocks until any of goroutines returns an error or expected OS signal is caught.
func Run(cfg Config, funcs ...func(context.Context) error) error {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}
	ctx := context.Background()
	if cfg.Context != nil {
		ctx = *cfg.Context
	}
	if len(cfg.Signals) == 0 {
		cfg.Signals = []os.Signal{os.Interrupt, os.Kill, syscall.SIGHUP}
	}
	ctx, cancel := signal.NotifyContext(ctx, cfg.Signals...)

	var wg sync.WaitGroup
	errChan := make(chan error, len(funcs))

	for _, f := range funcs {
		wg.Add(1)
		go func(fn func(context.Context) error) {
			errChan <- fn(ctx)
			wg.Done()
		}(f)
	}

	err := <-errChan
	cancel()

	doneChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	select {
	case <-time.After(cfg.Timeout):
		return ErrUngraceful{
			OriginalError: err,
		}

	case <-doneChan:
		return err
	}
}

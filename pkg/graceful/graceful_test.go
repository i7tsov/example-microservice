package graceful

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGracefulRun(t *testing.T) {
	err1 := errors.New("error one")
	err2 := errors.New("error two")
	err3 := errors.New("error three")
	cases := []struct {
		timeout     time.Duration
		name        string
		workers     []worker
		expectError error
	}{
		{
			name:    "ungraceful",
			timeout: 10 * time.Millisecond,
			workers: []worker{
				{endAfter: 10 * time.Millisecond, err: err1},
				{ungraceful: true},
			},
			expectError: ErrUngraceful{OriginalError: err1},
		},
		{
			name:    "graceful1",
			timeout: 10 * time.Millisecond,
			workers: []worker{
				{endAfter: 10 * time.Millisecond, err: err1},
				{endAfter: 100 * time.Millisecond, err: err2},
				{endAfter: 100 * time.Millisecond, err: err3},
			},
			expectError: err1,
		},
		{
			name:    "graceful2",
			timeout: 10 * time.Millisecond,
			workers: []worker{
				{endAfter: 100 * time.Millisecond, err: err1},
				{endAfter: 10 * time.Millisecond, err: err2},
				{endAfter: 100 * time.Millisecond, err: err3},
			},
			expectError: err2,
		},
		{
			name:    "graceful3",
			timeout: 10 * time.Millisecond,
			workers: []worker{
				{endAfter: 100 * time.Millisecond, err: err1},
				{endAfter: 100 * time.Millisecond, err: err2},
				{endAfter: 10 * time.Millisecond, err: err3},
			},
			expectError: err3,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			routines := make([]func(context.Context) error, 0, len(c.workers))
			for _, w := range c.workers {
				routines = append(routines, w.Run)
			}
			err := Run(Config{Timeout: c.timeout}, routines...)
			if err.Error() != c.expectError.Error() {
				t.Logf("Expected error '%v' doesn't match actual error: '%v'", err.Error(), c.expectError.Error())
				t.Fail()
			}
		})
	}

}

type worker struct {
	endAfter   time.Duration
	ungraceful bool
	err        error
}

func (w worker) Run(ctx context.Context) error {
	if w.ungraceful {
		select {}
	}
	if w.endAfter != 0 {
		select {
		case <-time.After(w.endAfter):
			return w.err
		case <-ctx.Done():
			return w.err
		}
	} else {
		<-ctx.Done()
		return w.err
	}
}

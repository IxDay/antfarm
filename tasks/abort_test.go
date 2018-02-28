package tasks

import (
	"context"
	"github.com/ixday/antfarm"
	"os"
	"testing"
	"time"
)

func unexpectedErr(t *testing.T, got, expected error) {
	if got != expected {
		t.Errorf("unexpected error type, got: %s, expected: %s", got, expected)
	}
}

func runner() (func(), chan struct{}, chan error) {
	start := make(chan struct{})
	stop := make(chan error)

	fn := func() {
		stop <- antfarm.Runner{}.
			Task("abort", Abort()).
			Task("infinite", antfarm.TaskFunc(func(ctx context.Context) error {
				close(start)
				for i := 0; i < 5; i++ { // 250ms
					select {
					case <-ctx.Done():
						return nil
					default:
						time.Sleep(50 * time.Millisecond)
					}
				}
				return nil
			})).
			Start("abort", "infinite")
	}
	return fn, start, stop
}

func TestInterrupt(t *testing.T) {
	fn, start, stop := runner()

	go fn()
	<-start
	p, err := os.FindProcess(os.Getpid())
	unexpectedErr(t, err, nil)
	unexpectedErr(t, p.Signal(os.Interrupt), nil)
	err = <-stop
	unexpectedErr(t, err, ErrInterrupt)
}

func TestNoInterrupt(t *testing.T) {
	fn, _, stop := runner()

	go fn()
	err := <-stop
	unexpectedErr(t, err, nil)
}

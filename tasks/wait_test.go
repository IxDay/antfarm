package tasks

import (
	"context"
	"testing"
	"time"
)

func TestWait(t *testing.T) {
	duration := 50 * time.Millisecond

	start := time.Now()

	Wait(duration).Start(context.Background())

	end := time.Now()

	if end.Sub(start) < duration {
		t.Errorf("wait did not act as a barrier, waiting time superior as elapsed")
	}
}

func TestWaitCancel(t *testing.T) {
	duration := 500 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())

	start := time.Now()

	go Wait(duration).Start(ctx)
	cancel() // stop right away
	<-ctx.Done()

	end := time.Now()

	if end.Sub(start) > duration {
		t.Errorf("wait was not canceled, elapsed time is superior as waiting")
	}
}

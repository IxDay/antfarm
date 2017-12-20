package antfarm

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

type buffer []string

func (b *buffer) NewTask(message string) Task {
	return TaskFunc(func(ctx context.Context) error {
		*b = append(*b, message)
		return nil
	})
}

func Error(err error) Task { return TaskFunc(func(_ context.Context) error { return err }) }

func compare(t *testing.T, a1, a2 []string) {
	if len(a1) != len(a2) {
		t.Errorf("arrays don't have same size")
	}
	for i, a := range a1 {
		if a != a2[i] {
			t.Errorf("value at position %d differs, got: %s, expected: %s", i, a, a2[i])
		}
	}
}

func unexpectedErr(t *testing.T, got, expected error) {
	if got != expected {
		t.Errorf("unexpected error type, got: %s, expected: %s", got, expected)
	}
}

func TestDependencyOrder(t *testing.T) {
	buffer := &buffer{}
	Runner{}.
		Task("world", buffer.NewTask("world"), "bar", "foo").
		Task("foo", buffer.NewTask("foo")).
		Task("bar", buffer.NewTask("bar"), "foo").
		Start("world")
	compare(t, []string(*buffer), []string{"foo", "bar", "world"})
}

func TestDependencyErr(t *testing.T) {
	runner := Runner{}.
		Task("foo", Noop(), "bar").
		Task("bar", Noop(), "foo").
		Task("baz", Noop(), "quz")

	unexpectedErr(t, runner.Start("baz"), ErrDepNotFound)
	unexpectedErr(t, runner.Start("bar"), ErrDepCircular)
}

func TestErrorPropagation(t *testing.T) {
	ErrBar := fmt.Errorf("bar")
	ErrBaz := fmt.Errorf("baz")

	runner := Runner{}.
		Task("foo", Noop()).
		Task("bar", Error(ErrBar), "foo").
		Task("baz", Error(ErrBaz), "bar")

	unexpectedErr(t, runner.Start("baz"), ErrBar)
}

func TestInterrupt(t *testing.T) {
	var err error
	var stop = make(chan bool)
	var start = make(chan bool)

	go func() {
		defer close(stop)
		err = Runner{}.
			Task("infinite", TaskFunc(func(ctx context.Context) error {
				close(start)
				for {
					select {
					case <-ctx.Done():
						return nil
					default:
						time.Sleep(1 * time.Second)
					}
				}
			})).
			Start("infinite")
	}()
	<-start
	p, err := os.FindProcess(os.Getpid())
	unexpectedErr(t, err, nil)
	unexpectedErr(t, p.Signal(os.Interrupt), nil)
	<-stop
	unexpectedErr(t, err, ErrInterrupt)
}

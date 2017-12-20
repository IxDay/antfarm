package antfarm

import (
	"context"
	"fmt"
	"testing"
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

func TestDependencyOrder(t *testing.T) {
	buffer := &buffer{}
	Runner{}.
		Task("world", buffer.NewTask("world"), "bar", "foo").
		Task("foo", buffer.NewTask("foo")).
		Task("bar", buffer.NewTask("bar"), "foo").
		Start("world")
	compare(t, []string(*buffer), []string{"foo", "bar", "world"})
}

func TestErrorPropagation(t *testing.T) {
	ErrBar := fmt.Errorf("bar")
	ErrBaz := fmt.Errorf("baz")

	err := Runner{}.
		Task("foo", Noop()).
		Task("bar", Error(ErrBar), "foo").
		Task("baz", Error(ErrBaz), "bar").
		Start("baz")
	if err != ErrBar {
		t.Errorf("unexpected error type, got: %s, expected: %s", err, ErrBar)
	}
}

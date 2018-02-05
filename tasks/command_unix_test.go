package tasks

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	err := Command("/bin/true").Start(context.Background())
	if err != nil {
		t.Errorf("task should not have returned an error, got: %s", err)
	}
}

func TestCommandOutput(t *testing.T) {
	reader, writer, err := os.Pipe()
	expected := "foo"
	text := make([]byte, len(expected)+1)

	if err != nil {
		t.Fatalf("%s", err)
	}
	defer reader.Close()
	defer writer.Close()

	go Command("/bin/echo", func(cmd *exec.Cmd) {
		cmd.Stdout = writer
		cmd.Args = append(cmd.Args, "foo")
	}).Start(context.Background())

	if _, err = reader.Read(text); err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if expected += "\n"; string(text) != expected {
		t.Errorf("unexpected output, got: %q, wanted: %q", text, expected)
	}
}

func TestCommandAbort(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(50*time.Millisecond, cancel)

	err := Command("/bin/yes").Start(ctx)
	if err == nil {
		t.Fatalf("expected an error when aborting command")
	}
	if err.Error() != "signal: killed" { // this is not really robust
		t.Fatalf("unexpected error type, got: %s", err)
	}
}

package tasks

import (
	"context"
	"os"
	"testing"
)

func TestPrint(t *testing.T) {
	expected := "Test Print"
	text := make([]byte, len(expected)) // add \n to length

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("%s", err)
	}
	defer reader.Close()
	defer writer.Close()

	go Print(expected, func(opt *PrintOpts) { opt.Writer = writer }).Start(context.Background())

	_, err = reader.Read(text)
	if err != nil {
		t.Fatalf("%s", err)
	}
	if expected != string(text) {
		t.Errorf("expected: %q, got: %q", expected, text)
	}
}

package tasks

import (
	"context"
	"github.com/ixday/antfarm"
	"io"
	"os"
)

type PrintOpts struct {
	io.Writer
}

func Print(msg string, options ...func(*PrintOpts)) antfarm.Task {
	opts := PrintOpts{os.Stdout}
	for _, option := range options {
		option(&opts)
	}
	return antfarm.TaskFunc(func(_ context.Context) error {
		_, err := opts.Writer.Write([]byte(msg))
		return err
	})
}

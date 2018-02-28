package tasks

import (
	"context"
	"fmt"
	"github.com/ixday/antfarm"
	"os"
	"os/signal"
)

var (
	ErrInterrupt = fmt.Errorf("Aborting due to ^C...")
)

func Abort() antfarm.Task {
	return antfarm.TaskFunc(func(ctx context.Context) error {
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt)
		select {
		case <-interrupt:
			return ErrInterrupt
		case <-ctx.Done():
		}
		return nil
	})
}

package tasks

import (
	"context"
	"fmt"
	"github.com/ixday/antfarm"
	"time"
)

func Wait(d time.Duration) antfarm.Task {
	return antfarm.TaskFunc(func(ctx context.Context) error {
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
			fmt.Printf("waited for %s\n", d)
		case <-ctx.Done():
			if ok := timer.Stop(); ok {
				fmt.Println("aborted waiting")
			}
		}
		return nil
	})
}

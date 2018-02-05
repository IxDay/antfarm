package tasks

import (
	"context"
	"github.com/ixday/antfarm"
	"os/exec"
)

func Command(name string, options ...func(*exec.Cmd)) antfarm.Task {
	return antfarm.TaskFunc(func(ctx context.Context) error {
		cmd := exec.CommandContext(ctx, name)
		for _, option := range options {
			option(cmd)
		}
		return cmd.Run()
	})
}

package antfarm

import (
	"context"
)

type (
	Task interface {
		Start(context.Context) error
	}
	Provisioner interface {
		Expect() (bool, error)
		Task
		Abort()
	}
	TaskFunc func(context.Context) error
)

func (tf TaskFunc) Start(ctx context.Context) error { return tf(ctx) }

func Provision(provisioner Provisioner) Task {
	return TaskFunc(func(ctx context.Context) error {
		if ok, err := provisioner.Expect(); err != nil || !ok {
			return err
		}

		if err := provisioner.Start(ctx); err != nil {
			provisioner.Abort()
			return err
		}
		return nil
	})
}

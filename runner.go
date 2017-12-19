package antfarm

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

type (
	Task interface {
		Start(context.Context) error
	}

	Node struct {
		Name string
		Task Task
		Deps []string
	}

	Runner map[string]Node

	state struct {
		context.CancelFunc
		Done chan bool
	}
)

type TaskFunc func(context.Context) error

func (tf TaskFunc) Start(ctx context.Context) error { return tf(ctx) }

func in(value string, array []string) bool {
	for _, elt := range array {
		if value == elt {
			return true
		}
	}
	return false
}

func reverse(array []string) []string {
	out := make([]string, len(array))
	l := len(array) - 1
	for j := range out {
		out[j] = array[l-j]
	}
	return out
}

func (r Runner) Task(name string, task Task, deps ...string) Runner {
	r[name] = Node{name, task, deps}
	return r
}

func (r Runner) Resolve(node Node) ([]string, error) {
	var seen, resolved []string
	var resolve func(node Node) error

	resolve = func(node Node) error {
		seen = append(seen, node.Name)
		for _, dep := range node.Deps {
			if !in(dep, resolved) {
				if in(dep, seen) {
					return fmt.Errorf("circular dependency detected")
				}
				node, ok := r[dep]
				if !ok {
					return fmt.Errorf("dependency not found")
				}
				if err := resolve(node); err != nil {
					return err
				}
			}
		}
		resolved = append(resolved, node.Name)
		return nil
	}
	return resolved, resolve(node)
}

func clean(running map[string]state, resolved []string, done chan error) (err error) {
	cleaned := make(chan bool)
	for {
		select {
		case e := <-done:
			if e == nil || err != nil { // ensure only one cleaning is running
				continue
			}
			err = e // save error
			go func() {
				defer close(cleaned)                     // close channel only when all tasks are canceled
				for _, name := range reverse(resolved) { // ensure all task are closed
					running[name].CancelFunc()
					<-running[name].Done
				}
			}()
		case <-running[""].Done:
			if err == nil { // if an error exists, there's a cleaning running
				return
			}
		case <-cleaned:
			return
		}
	}
}

func abort() Task {
	return TaskFunc(func(ctx context.Context) error {
		interrupt := make(chan os.Signal)
		signal.Notify(interrupt, os.Interrupt)
		select {
		case <-interrupt:
			fmt.Println()
			return fmt.Errorf("Aborting due to ^C...")
		case <-ctx.Done():
		}
		return nil
	})
}

func noop() Task { return TaskFunc(func(_ context.Context) error { return nil }) }

func (runner Runner) Start(tasks ...string) error {
	done := make(chan error)
	running := map[string]state{}
	root := runner.Task("abort", abort()).Task("", noop(), tasks...)[""]
	resolved, err := runner.Resolve(root)

	if err != nil {
		return err
	}

	for _, name := range append(resolved, "abort") {
		ctx, cancel := context.WithCancel(context.Background())
		running[name] = state{cancel, make(chan bool)}

		go func(ctx context.Context, node Node, s state) {
			defer close(s.Done)
			for _, dep := range node.Deps {
				select {
				case <-ctx.Done(): // if interrupt don't wait any longer
					return
				case <-running[dep].Done: // wait for dependencies to finish
				}
			}
			done <- node.Task.Start(ctx) // start job
		}(ctx, runner[name], running[name])
	}

	return clean(running, resolved, done)
}

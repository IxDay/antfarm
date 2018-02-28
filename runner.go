package antfarm

import (
	"context"
	"fmt"
)

var (
	ErrDepNotFound = fmt.Errorf("dependency not found")
	ErrDepCircular = fmt.Errorf("circular dependency detected")
	ErrFinished    = fmt.Errorf("task running finished")
)

type (
	Node struct {
		Name string
		Task Task
		Deps []string
	}

	state struct {
		context context.Context
		cancel  context.CancelFunc
		done    chan struct{}
	}

	Runner struct {
		Resolver
		Tasks map[string]Node
	}
)

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

func newState(task Task) state {
	ctx, cancel := context.WithCancel(context.Background())
	if lt, ok := task.(LongTask); ok {
		return state{ctx, func() { lt.Cancel(); cancel() }, make(chan struct{})}
	} else {
		return state{ctx, cancel, make(chan struct{})}
	}
}

func NewRunner(options ...func(*Runner)) *Runner {
	runner := &Runner{ResolverFunc(Resolve), make(map[string]Node)}
	for _, option := range options {
		option(runner)
	}
	return runner
}

func (r *Runner) Task(name string, task Task, deps ...string) *Runner {
	r.Tasks[name] = Node{name, task, deps}
	return r
}

func Resolve(node Node, tasks map[string]Node) ([]string, error) {
	var seen, resolved []string
	var resolve func(node Node) error

	resolve = func(node Node) error {
		seen = append(seen, node.Name)
		for _, dep := range node.Deps {
			if !in(dep, resolved) {
				if in(dep, seen) {
					return ErrDepCircular
				}
				node, ok := tasks[dep]
				if !ok {
					return ErrDepNotFound
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
	cleaned := make(chan struct{})
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
					running[name].cancel()
					<-running[name].done
				}
			}()
		case <-running[""].done:
			if err == nil { // if an error exists, there's a cleaning running
				return
			}
		case <-cleaned:
			return
		}
	}
}

func (runner *Runner) Start(tasks ...string) error {
	done := make(chan error)
	running := map[string]state{}
	root := runner.Task("", TaskFunc(func(_ context.Context) error {
		return ErrFinished
	}), tasks...).Tasks[""]
	resolved, err := runner.Resolve(root, runner.Tasks)
	if err != nil {
		return err
	}

	for _, name := range resolved {
		running[name] = newState(runner.Tasks[name].Task)
		go func(name string) {
			state := running[name]
			defer close(state.done)

			for _, dep := range runner.Tasks[name].Deps {
				select {
				case <-state.context.Done(): // if interrupt don't wait any longer
					return
				case <-running[dep].done: // wait for dependencies to finish
				}
			}
			done <- runner.Tasks[name].Task.Start(state.context) // start job
		}(name)
	}

	return clean(running, resolved, done)
}

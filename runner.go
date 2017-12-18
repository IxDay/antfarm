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
		context.Context
		context.CancelFunc
	}

	Runner struct {
		Tree map[string]Node
		Done chan error
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

type TaskFunc func(context.Context) error

func (tf TaskFunc) Start(ctx context.Context) error { return tf(ctx) }

func NewRunner() *Runner {
	return &Runner{map[string]Node{}, make(chan error)}
}

func (r *Runner) Task(name string, task Task, deps ...string) *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	r.Tree[name] = Node{name, task, deps, ctx, cancel}
	return r
}

func (r *Runner) Resolve(node Node) ([]string, error) {
	var seen, resolved []string
	var resolve func(node Node) error

	resolve = func(node Node) error {
		seen = append(seen, node.Name)
		for _, dep := range node.Deps {
			if !in(dep, resolved) {
				if in(dep, seen) {
					return fmt.Errorf("circular dependency detected")
				}
				node, ok := r.Tree[dep]
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

func (r *Runner) clean(resolved []string, err error) {
	// stop in reverse order (can be parallelized)
	go func() {
		for i := len(resolved) - 1; i >= 0; i-- {
			r.Tree[resolved[i]].CancelFunc()
		}
		r.Done <- err
	}()
}

func (runner *Runner) Start(tasks ...string) error {
	root := runner.Task("", Noop(), tasks...).Tree[""]
	interrupt := make(chan os.Signal)
	resolved, err := runner.Resolve(root)

	if err != nil {
		return err
	}

	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for _ = range interrupt {
			runner.clean(resolved, fmt.Errorf("Aborting due to ^C..."))
		}
	}()

	for _, name := range resolved {
		go func(node Node) {
			for _, dep := range node.Deps {
				<-runner.Tree[dep].Context.Done() // wait for dependencies to finish
			}
			select {
			case <-node.Context.Done():
				return
			default:
				if err = node.Task.Start(node.Context); err != nil { // start job
					runner.clean(resolved, err)
				} else {
					node.CancelFunc()
				}
			}
			return
		}(runner.Tree[name])
	}

	select {
	case err = <-runner.Done:
		fmt.Println(err)
	case <-root.Context.Done():
	}
	return err
}

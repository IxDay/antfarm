package main

import (
	"fmt"
	"os"
	"os/signal"
)

type (
	Task interface {
		Start() error
		Stop(error) error
	}

	Node struct {
		Name string
		Task Task
		Done chan bool
		Deps []string
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

func NewRunner() *Runner {
	return &Runner{map[string]Node{}, make(chan error)}
}

func (r *Runner) Task(name string, task Task, deps ...string) *Runner {
	r.Tree[name] = Node{name, task, make(chan bool), deps}
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
			// have to store errors
			r.Tree[resolved[i]].Task.Stop(err)
		}
		r.Done <- err
	}()
}

func (runner *Runner) Start(tasks ...string) error {
	scheduler := make(map[string]chan bool, len(runner.Tree))
	root := runner.Task("", Noop{}, tasks...).Tree[""]
	interrupt := make(chan os.Signal)
	resolved, err := runner.Resolve(root)

	if err != nil {
		return err
	}

	signal.Notify(interrupt, os.Interrupt)
	go func() {
		for _ = range interrupt {
			runner.clean(resolved, fmt.Errorf("Received an interrupt..."))
		}
	}()

	for _, name := range resolved {
		scheduler[name] = runner.Tree[name].Done
		go func(node Node) {
			for _, dep := range node.Deps {
				<-scheduler[dep] // wait for dependencies to finish
			}
			if err = node.Task.Start(); err != nil { // start job
				runner.clean(resolved, err)
			}
			close(node.Done)
			return
		}(runner.Tree[name])
	}
	select {
	case <-root.Done:
		runner.clean(resolved, nil)
	case err = <-runner.Done:
	}
	return err
}

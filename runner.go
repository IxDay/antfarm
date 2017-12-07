package main

import (
	"context"
	"fmt"
)

type (
	Task interface {
		Start() error
		Stop() error
	}

	Node struct {
		Name string
		Task Task
		Done chan bool
		Deps []string
	}

	Runner struct {
		Tree map[string]Node
	}

	Resolver struct {
		seen, resolved []string
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
	return &Runner{map[string]Node{}}
}

func (r *Runner) Task(name string, task Task, deps ...string) *Runner {
	r.Tree[name] = Node{name, task, make(chan bool), deps}
	return r
}

func (r *Resolver) Resolve(node Node, tree map[string]Node) error {
	r.seen = append(r.seen, node.Name)
	for _, dep := range node.Deps {
		if !in(dep, r.resolved) {
			if in(dep, r.seen) {
				return fmt.Errorf("circular dependency detected")
			}
			node, ok := tree[dep]
			if !ok {
				return fmt.Errorf("dependency not found")
			}
			if err := r.Resolve(node, tree); err != nil {
				return err
			}
		}
	}
	r.resolved = append(r.resolved, node.Name)
	return nil
}

func (r *Runner) Start(tasks ...string) error {
	resolver := Resolver{}
	scheduler := make(map[string]chan bool, len(r.Tree))
	ctx, cancel := context.WithCancel(context.Background())
	root := r.Task("", Noop{}, tasks...).Tree[""]

	if err := resolver.Resolve(root, r.Tree); err != nil {
		return err
	}

	for _, name := range resolver.resolved {
		scheduler[name] = r.Tree[name].Done
		go func(node Node) {
			for _, dep := range node.Deps {
				<-scheduler[dep] // wait for dependencies to finish
			}
			if err := node.Task.Start(); err != nil { // start job
				cancel()
			}
			close(node.Done)
			return
		}(r.Tree[name])
		go func(node Node) {
			for {
				select {
				case <-ctx.Done():
					fmt.Printf("aborting: %s", node.Name)
					node.Task.Stop()
					return // returning not to leak the goroutine
				}
			}
		}(r.Tree[name])
	}
	<-root.Done

	for i := len(resolver.resolved) - 1; i >= 0; i-- {
		r.Tree[resolver.resolved[i]].Task.Stop() // stop in reverse order (can be parallelized)
	}
	return nil
}

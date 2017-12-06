package main

import (
	"context"
	"fmt"
	"sync"
)

type Node struct {
	Name string
	Task Task
	Deps []string
}

type Runner struct {
	Tree map[string]Node
}

func NewRunner() *Runner {
	return &Runner{map[string]Node{}}
}

func (r *Runner) Task(name string, task Task, deps ...string) *Runner {
	r.Tree[name] = Node{name, task, deps}
	return r
}

func InArray(value string, array []string) bool {
	for _, elt := range array {
		if value == elt {
			return true
		}
	}
	return false
}

type Resolver struct {
	seen, resolved []string
}

func (r *Resolver) Resolve(node Node, tree map[string]Node) error {
	r.seen = append(r.seen, node.Name)
	for _, dep := range node.Deps {
		if !InArray(dep, r.resolved) {
			if InArray(dep, r.seen) {
				return fmt.Errorf("circular dependency detected")
			}
			if err := r.Resolve(tree[dep], tree); err != nil {
				return err
			}
		}
	}
	r.resolved = append(r.resolved, node.Name)
	return nil
}

func (r *Runner) Start(tasks ...string) error {
	var resolver Resolver

	if err := resolver.Resolve(Node{"", Noop{}, tasks}, r.Tree); err != nil {
		return err
	}
	return nil
}

func (r *Runner) start() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	for _, node := range r.Tree {
		wg.Add(1)
		go func(task Task) {
			ok, err := task.Expect()
			if err != nil {
				cancel()
			}
			if !ok {
				task.Start()
			}
			wg.Done()
		}(node.Task)
		go func(task Task) {
			for {
				select {
				case <-ctx.Done():
					fmt.Println("stopping")
					task.Stop()
					wg.Done()
					return // returning not to leak the goroutine
				}
			}
		}(node.Task)
	}
	wg.Wait()
}

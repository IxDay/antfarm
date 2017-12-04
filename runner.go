package main

import (
	"context"
	"fmt"
	"sync"
)

type Runner struct {
	tasks map[string]Task
}

func NewRunner() *Runner {
	return &Runner{map[string]Task{}}
}

func (r *Runner) Task(name string, task Task) *Runner {
	r.tasks[name] = task
	return r
}

func (r *Runner) Start() {
	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())

	for _, task := range r.tasks {
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
		}(task)
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
		}(task)
	}
	cancel()
	wg.Wait()
}

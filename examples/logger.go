package main

import (
	"context"
	"fmt"
	"github.com/ixday/antfarm"
	"github.com/ixday/antfarm/tasks"
	"time"
)

type runner struct{ antfarm.Runner }

func NewRunner() runner { return runner{antfarm.Runner{}} }
func (r runner) Task(name string, task antfarm.Task, deps ...string) runner {
	r.Runner = r.Runner.Task(name, antfarm.TaskFunc(func(ctx context.Context) error {
		fmt.Printf("starting task: %s...\n", name)
		if err := task.Start(ctx); err != nil {
			return err
		}
		fmt.Printf("finished task: %s...\n", name)
		return nil
	}), deps...)
	return r
}

func main() {
	fmt.Println(NewRunner().
		Task("wait", tasks.Wait(5*time.Second)).
		Task("world", tasks.Print("Hello World!"), "bar", "foo").
		Task("foo", tasks.Print("Hello Foo!")).
		Task("bar", tasks.Print("Hello Bar!"), "foo", "wait").
		Start("world"))
}

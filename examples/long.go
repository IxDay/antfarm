package main

import (
	"context"
	"fmt"
	"github.com/ixday/antfarm"
	"github.com/ixday/antfarm/tasks"
	"time"
)

type LongTask struct{}

func (lt LongTask) Start(_ context.Context) error {
	fmt.Println("starting...")
	return nil
}
func (lt LongTask) Cancel() { fmt.Println("tearing down...") }

func main() {
	fmt.Println(antfarm.Runner{}.
		Task("long", LongTask{}).
		Task("wait", tasks.Wait(2*time.Second), "long").
		Start("wait"))
}

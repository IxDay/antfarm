package main

import (
	"fmt"
	"github.com/ixday/antfarm"
	"github.com/ixday/antfarm/tasks"
	"time"
)

func addAbort(runner *antfarm.Runner) {
	runner.Resolver = antfarm.ResolverFunc(
		func(node antfarm.Node, nodes map[string]antfarm.Node) ([]string, error) {
			// you can test if the value does not already exist
			runner.Task("abort", tasks.Abort())

			resolved, err := antfarm.Resolve(node, nodes)
			return append([]string{"abort"}, resolved...), err
		},
	)
}

func main() {
	fmt.Println(antfarm.NewRunner(addAbort).
		Task("wait", tasks.Wait(2*time.Second)).
		Start("wait"))
}

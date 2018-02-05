package main

import (
	"fmt"
	"github.com/ixday/antfarm"
	"github.com/ixday/antfarm/tasks"
	"io"
	"os"
	"os/exec"
	"time"
)

type optFn func(cmd *exec.Cmd)

func cmdArgs(args ...string) optFn {
	return func(cmd *exec.Cmd) { cmd.Args = append(cmd.Args, args...) }
}
func cmdStdout(w io.Writer) optFn { return func(cmd *exec.Cmd) { cmd.Stdout = w } }

func main() {

	fmt.Println(antfarm.Runner{}.
		Task("wait", tasks.Wait(5*time.Second)).
		Task("world", tasks.Print("Hello World!"), "bar", "foo").
		Task("foo", tasks.Print("Hello Foo!")).
		Task("bar", tasks.Print("Hello Bar!"), "foo", "wait").
		Task("exec", tasks.Command("echo", cmdStdout(os.Stdout), cmdArgs("Hello World!"))).
		Start(os.Args[1:]...))
}

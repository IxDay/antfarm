package main

import (
	"fmt"
	"github.com/ixday/antfarm"
	"os"
	"os/exec"
	"time"
)

func setOpts(cmd *exec.Cmd) {
	cmd.Args = append(cmd.Args, "Hello World!")
	cmd.Stdout = os.Stdout
}

func main() {

	fmt.Println(antfarm.Runner{}.
		Task("wait", antfarm.Wait(5*time.Second)).
		Task("world", antfarm.Print("Hello World!"), "bar", "foo").
		Task("foo", antfarm.Print("Hello Foo!")).
		Task("bar", antfarm.Print("Hello Bar!"), "foo", "wait").
		Task("exec", antfarm.Command("echo", setOpts)).
		Start(os.Args[1:]...))
}

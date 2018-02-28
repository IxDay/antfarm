package main

import (
	"fmt"
	"github.com/ixday/antfarm"
	"github.com/ixday/antfarm/tasks"
	"time"
)

func main() {
	rulz := tasks.NewSSH(func(config *tasks.SSHConfig) {
		config.User = "root"
		config.Host = "rulz.xyz"
	})

	fmt.Println(antfarm.Runner{}.
		Task("wait", tasks.Wait(5*time.Second)).
		Task("world", tasks.Print("Hello World!\n"), "bar", "foo").
		Task("foo", tasks.Print("Hello Foo!\n")).
		Task("bar", tasks.Print("Hello Bar!\n"), "foo", "wait").
		Task("ls", rulz.Run("ls -la")).
		Task("echo", rulz.Run("echo foo"), "ls").
		Start("world"))
}

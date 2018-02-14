package main

import (
	"fmt"
	"github.com/ixday/antfarm"
	"github.com/ixday/antfarm/tasks"
	"os"
	"time"
)

func main() {
	rulz, err := tasks.NewSSH(func(config *tasks.SSHConfig) {
		config.User = "root"
		config.Host = "rulz.xyz"
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(antfarm.Runner{}.
		Task("wait", tasks.Wait(5*time.Second)).
		Task("world", tasks.Print("Hello World!"), "bar", "foo").
		Task("foo", tasks.Print("Hello Foo!")).
		Task("bar", tasks.Print("Hello Bar!"), "foo", "wait").
		Task("ssh", rulz.Run("ls -la")).
		Start("ssh"))
}

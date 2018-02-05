package main

import (
	"fmt"
	"github.com/ixday/antfarm"
	"github.com/ixday/antfarm/tasks"
	"time"
)

func main() {

	fmt.Println(antfarm.Runner{}.
		Task("wait", tasks.Wait(5*time.Second)).
		Task("copy", tasks.FileCopy("/home/max/games/Champions of Norrath (USA).iso", "./out")).
		Task("world", tasks.Print("Hello World!"), "bar", "foo").
		Task("foo", tasks.Print("Hello Foo!")).
		Task("bar", tasks.Print("Hello Bar!"), "foo", "wait").
		Start("copy"))
}

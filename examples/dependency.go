package main

import (
	"fmt"
	"github.com/ixday/antfarm"
	"time"
)

func main() {

	fmt.Println(antfarm.Runner{}.
		Task("wait", antfarm.Wait(5*time.Second)).
		Task("copy", antfarm.FileCopy("/home/max/games/Champions of Norrath (USA).iso", "./out")).
		Task("world", antfarm.Print("Hello World!"), "bar", "foo").
		Task("foo", antfarm.Print("Hello Foo!")).
		Task("bar", antfarm.Print("Hello Bar!"), "foo", "wait").
		Start("copy"))
}

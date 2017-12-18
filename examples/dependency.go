package main

import (
	"github.com/ixday/antfarm"
	"time"
)

func main() {

	antfarm.NewRunner().
		Task("wait", antfarm.Wait(5*time.Second)).
		Task("world", antfarm.Print("Hello World!"), "bar", "foo").
		Task("foo", antfarm.Print("Hello Foo!")).
		Task("bar", antfarm.Print("Hello Bar!"), "foo", "wait").
		Start("world")
}

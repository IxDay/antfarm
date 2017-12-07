package main

import (
	"github.com/ixday/antfarm"
)

func main() {

	antfarm.NewRunner().
		Task("world", antfarm.Print{"Hello World!"}, "bar", "foo").
		Task("foo", antfarm.Print{"Hello Foo!"}).
		Task("bar", antfarm.Print{"Hello Bar!"}, "foo").
		Start("world")
}

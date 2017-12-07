package main

import (
	"fmt"
	"time"
)

func main() {

	err := NewRunner().
		Task("wait", NewWait(5*time.Second), "foo", "bar").
		Task("world", Print{"Hello World!"}, "wait").
		Task("foo", Print{"Hello Foo!"}).
		Task("bar", Print{"Hello Bar!"}, "foo").
		Start("world")
	fmt.Println(err)
}

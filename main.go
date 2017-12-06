package main

import (
	"fmt"
	"time"
)

func main() {

	err := NewRunner().
		Task("wait", NewWait(5*time.Second), "foo", "bar", "world").
		Task("world", Print{"Hello World!"}).
		Task("foo", Print{"Hello Foo!"}).
		Task("bar", Print{"Hello Bar!"}, "foo").
		Start("wait")
	fmt.Println(err)
}

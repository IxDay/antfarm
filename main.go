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
		Task("copy", FileCopy("main.go", "/tmp/main.go")).
		Start("wait")
	fmt.Println(err)
}

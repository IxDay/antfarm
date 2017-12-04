package main

import (
	"time"
)

func main() {

	NewRunner().
		Task("wait", NewWait(5*time.Second)).
		Task("world", Print{"Hello World!"}).
		Task("foo", Print{"Hello Foo!"}).
		Task("bar", Print{"Hello Bar!"}).
		Start()
}

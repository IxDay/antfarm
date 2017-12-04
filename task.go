package main

type Task interface {
	Start() error
	Stop() error
	Expect() (bool, error)
}

type LongTask interface {
	Task
	Ready() error
}

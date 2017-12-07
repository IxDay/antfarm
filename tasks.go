package main

import (
	"fmt"
	"time"
)

type Noop struct{}

func (n Noop) Stop() error  { return nil }
func (n Noop) Start() error { return nil }

type Print struct{ message string }

func (p Print) Stop() error { return nil }
func (p Print) Start() error {
	fmt.Println(p.message)
	return nil
}

type Wait struct {
	duration time.Duration
	timer    *time.Timer
}

func NewWait(d time.Duration) *Wait {
	return &Wait{duration: d}
}

func (w *Wait) Stop() error {
	if w.timer == nil {
		return nil
	}
	if ok := w.timer.Stop(); ok {
		fmt.Println("aborted")
	}
	return nil
}
func (w *Wait) Start() error {
	w.timer = time.NewTimer(w.duration)
	<-w.timer.C
	fmt.Printf("waited for %s\n", w.duration)
	return nil
}

package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

type Noop struct{}

func (n Noop) Stop(err error) error { return nil }
func (n Noop) Start() error         { return nil }

type Print struct{ message string }

func (p Print) Stop(err error) error { return nil }
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

func (w *Wait) Stop(err error) error {
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

type Provisioner interface {
	Destination() string
	Provision() error
}

type Provision struct {
	Provisioner
}

func (p Provision) Start() error {
	if _, err := os.Stat(p.Destination()); os.IsNotExist(err) {
		return p.Provision()
	}
	return nil
}

func (p Provision) Stop(err error) error {
	if err != nil {
	}
	return nil
}

type fileCopy struct {
	source, destination string
}

func (fc fileCopy) Destination() string { return fc.destination }
func (fc fileCopy) Provision() error {
	in, err := os.Open(fc.source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(fc.destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

func FileCopy(src, dest string) Task { return Provision{fileCopy{src, dest}} }

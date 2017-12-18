package antfarm

import (
	"context"
	"fmt"
	//"io"
	//"os"
	"time"
)

func Noop() Task { return TaskFunc(func(_ context.Context) error { return nil }) }

func Print(msg string) Task {
	return TaskFunc(func(_ context.Context) error {
		fmt.Println(msg)
		return nil
	})
}

func Wait(d time.Duration) Task {
	return TaskFunc(func(ctx context.Context) error {
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
			fmt.Printf("waited for %s\n", d)
		case <-ctx.Done():
			if ok := timer.Stop(); ok {
				fmt.Println("aborted waiting")
			}
		}
		return nil
	})
}

//type Provisioner interface {
//	Destination() string
//	Provision() error
//}
//
//type Provision struct {
//	Provisioner
//}
//
//func (p Provision) Start() error {
//	if _, err := os.Stat(p.Destination()); os.IsNotExist(err) {
//		return p.Provision()
//	}
//	return nil
//}
//
//func (p Provision) Stop(err error) error {
//	if err != nil {
//	}
//	return nil
//}
//
//type fileCopy struct {
//	source, destination string
//}
//
//func (fc fileCopy) Destination() string { return fc.destination }
//func (fc fileCopy) Provision() error {
//	in, err := os.Open(fc.source)
//	if err != nil {
//		return err
//	}
//	defer in.Close()
//
//	out, err := os.Create(fc.destination)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//
//	_, err = io.Copy(out, in)
//	if err != nil {
//		return err
//	}
//	return out.Close()
//}
//
//func FileCopy(src, dest string) Task { return Provision{fileCopy{src, dest}} }

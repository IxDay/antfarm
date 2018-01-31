package antfarm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
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

func Command(name string, options ...func(*exec.Cmd)) Task {
	return TaskFunc(func(ctx context.Context) error {
		cmd := exec.CommandContext(ctx, name)
		for _, option := range options {
			option(cmd)
		}
		return cmd.Run()
	})
}

type Provisioner interface {
	Expect() (bool, error)
	Task
	Abort()
}

func Provision(provisioner Provisioner) Task {
	return TaskFunc(func(ctx context.Context) error {
		if ok, err := provisioner.Expect(); err != nil || !ok {
			return err
		}

		if err := provisioner.Start(ctx); err != nil {
			provisioner.Abort()
			return err
		}
		return nil
	})
}

type fileCopy struct{ source, destination string }

func FileCopy(src, dest string) Task { return Provision(fileCopy{src, dest}) }

func (fc fileCopy) Abort() { os.Remove(fc.destination) }
func (fc fileCopy) Expect() (ok bool, err error) {
	_, err = os.Stat(fc.destination)
	if os.IsNotExist(err) {
		return true, nil
	}
	return
}
func (fc fileCopy) Start(ctx context.Context) error {
	buf := make([]byte, 32*1024)
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

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			nr, err := in.Read(buf)

			if nr > 0 {
				nw, ew := out.Write(buf[0:nr])
				if ew != nil {
					return ew
				}
				if nr != nw {
					return io.ErrShortWrite
				}
			}

			if err != nil {
				if err != io.EOF {
					return err
				}
				return nil
			}
		}
	}

	return nil
}

type fileCopyMD5 struct{ fileCopy }

func hash_file_md5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}

func (fcmd5 fileCopyMD5) Expect() (ok bool, err error) {
	hashSrc, err := hash_file_md5(fcmd5.source)
	if err != nil {
		return
	}
	hashDest, err := hash_file_md5(fcmd5.destination)
	if err != nil {
		return
	}
	ok = hashSrc != hashDest
	return
}

func FileCopyMD5(src, dest string) Task { return Provision(fileCopyMD5{fileCopy{src, dest}}) }

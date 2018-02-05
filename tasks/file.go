package tasks

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/ixday/antfarm"
	"io"
	"os"
)

type (
	readerFunc  func(p []byte) (n int, err error)
	fileCopy    struct{ source, destination string }
	fileCopyMD5 struct{ fileCopy }
)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

func (fc fileCopy) Abort() { os.Remove(fc.destination) }
func (fc fileCopy) Expect() (ok bool, err error) {
	_, err = os.Stat(fc.destination)
	if os.IsNotExist(err) {
		return true, nil
	}
	return
}

func (fc fileCopy) Start(ctx context.Context) error {
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
	_, err = io.Copy(out, readerFunc(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return in.Read(p)
		}
	}))
	return err
}

func md5HashF(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return md5Hash(file)
}

func md5Hash(reader io.Reader) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)[:16]), nil
}

func (fcmd5 fileCopyMD5) Expect() (bool, error) {
	hashSrc, err := md5HashF(fcmd5.source)
	if err != nil {
		return false, err
	}
	hashDest, err := md5HashF(fcmd5.destination)
	if err != nil && !os.IsNotExist(err) {
		return false, err
	}
	return hashSrc != hashDest, nil
}

func FileCopy(src, dest string) antfarm.Task { return antfarm.Provision(fileCopy{src, dest}) }
func FileCopyMD5(src, dest string) antfarm.Task {
	return antfarm.Provision(fileCopyMD5{fileCopy{src, dest}})
}

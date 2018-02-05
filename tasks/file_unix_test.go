// Interacting with FS here, better scope to OS

package tasks

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func helperEnv(t *testing.T, fn func(*os.File)) {
	t.Helper()
	dir, err := ioutil.TempDir("", "")
	helperMust(t, err)
	defer func() { helperMust(t, os.RemoveAll(dir)) }()
	f, err := ioutil.TempFile(dir, "")
	helperMust(t, err)
	defer f.Close()
	_, err = f.Write([]byte("foo bar"))
	helperMust(t, err)
	fn(f)
}

// prepare one file in temp dir and generate a random destination file name
func helperFileCopy(t *testing.T, fn func(*os.File, *os.File)) {
	t.Helper()
	helperEnv(t, func(src *os.File) {
		dest, err := ioutil.TempFile(filepath.Dir(src.Name()), "")
		helperMust(t, err)
		dest.Close()

		helperMust(t, os.Remove(dest.Name())) // ensure file does not exist anymore
		fn(src, dest)
	})
}

// prepare two identical file in temp dir
func helperFileCopyMD5(t *testing.T, fn func(*os.File, *os.File)) {
	t.Helper()
	helperEnv(t, func(src *os.File) {
		dest, err := ioutil.TempFile(filepath.Dir(src.Name()), "")
		helperMust(t, err)
		_, err = io.Copy(dest, src)
		helperMust(t, err)
		fn(src, dest)
	})
}

func helperMust(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestMD5(t *testing.T) {
	expected := "327b6f07435811239bc47e1544353273"

	helperEnv(t, func(f *os.File) {
		if hash, err := md5HashF(f.Name()); err != nil {
			t.Fatal(err)
		} else if hash != expected {
			t.Errorf("unexpected hash result, want: %s, got: %s", expected, hash)
		}
	})
}

func TestMD5NoFile(t *testing.T) {
	helperEnv(t, func(f *os.File) {
		helperMust(t, os.Remove(f.Name()))
		expected := fmt.Sprintf("open %s: no such file or directory", f.Name())
		if _, err := md5HashF(f.Name()); err.Error() != expected { // this is not robust
			t.Errorf("unexpected error type, want: %s, got: %s", expected, err)
		}
	})
}

func TestMD5ReaderClosed(t *testing.T) {
	helperEnv(t, func(f *os.File) {
		f.Close()
		expected := fmt.Sprintf("read %s: file already closed", f.Name())
		if _, err := md5Hash(f); err.Error() != expected { // this is not robust
			t.Errorf("unexpected error type, want: %s, got: %s", expected, err)
		}
	})
}

func TestFileCopy(t *testing.T) {
	helperFileCopy(t, func(src, dest *os.File) {
		err := FileCopy(src.Name(), dest.Name()).Start(context.Background())
		if err != nil {
			t.Errorf("unexpected error from task run, got: %s", err)
		}
		hashSrc, err := md5HashF(src.Name())
		if err != nil {
			t.Errorf("unexpected error from hashing src file, got: %s", err)
		}

		hashDest, err := md5HashF(dest.Name())
		if err != nil {
			t.Errorf("unexpected error from hashing dest file, got: %s", err)
		}
		if hashSrc != hashDest {
			t.Errorf("dest file does not have correct md5, got: %s, want: %s", hashDest, hashSrc)
		}
	})
}

func TestFileCopyAbort(t *testing.T) {
	ch := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	helperFileCopy(t, func(src, dest *os.File) {
		go func() { ch <- FileCopy(src.Name(), dest.Name()).Start(ctx) }()

		cancel()
		err := <-ch
		if err != context.Canceled {
			t.Errorf("unexpected error type, got: %s, want: %s", err, context.Canceled)
		}
		if _, err := os.Stat(dest.Name()); !os.IsNotExist(err) {
			t.Errorf("abort should have removed the dest file, got: %s", err)
		}
	})
}

func TestFileCopySrcDoesNotExist(t *testing.T) {

	helperFileCopy(t, func(src, dest *os.File) {
		helperMust(t, os.Remove(src.Name()))
		err := FileCopy(src.Name(), dest.Name()).Start(context.Background())
		if err == nil || err.Error() != fmt.Sprintf("open %s: no such file or directory", src.Name()) {
			t.Errorf("unexpected error type, got: %s", err.Error())
		}
	})
}

func TestFileCopyDestExist(t *testing.T) {
	helperFileCopy(t, func(src, dest *os.File) {
		dest, err := os.Create(dest.Name())
		helperMust(t, err)
		helperMust(t, FileCopy(src.Name(), dest.Name()).Start(context.Background()))

		bytes, err := ioutil.ReadAll(dest)
		helperMust(t, err)
		if string(bytes) != "" {
			t.Errorf("unexpected content on destination file, got: '%s', expected: '%s'", bytes, "")
		}
	})
}

func TestFileCopyDestNotReachable(t *testing.T) {
	helperEnv(t, func(src *os.File) {
		helperEnv(t, func(dest *os.File) {
			expected := fmt.Sprintf("open %s: permission denied", dest.Name())
			helperMust(t, os.Remove(dest.Name()))
			helperMust(t, os.Chmod(filepath.Dir(dest.Name()), 0500))
			err := FileCopy(src.Name(), dest.Name()).Start(context.Background())
			if err.Error() != expected {
				t.Errorf("unexpected error type, got: %s, want: %s", err, expected)
			}
		})
	})
}

func TestFileCopyMD5(t *testing.T) {
	helperFileCopy(t, func(src *os.File, dest *os.File) {
		helperMust(t, FileCopyMD5(src.Name(), dest.Name()).Start(context.Background()))

		hashSrc, err := md5HashF(src.Name())
		helperMust(t, err)
		hashDest, err := md5HashF(dest.Name())
		helperMust(t, err)

		if hashSrc != hashDest {
			t.Errorf("src and dest hash are not identical, got: %s and %s", hashSrc, hashDest)
		}
	})
}

func TestFileCopyMD5SrcNotReachable(t *testing.T) {
	helperFileCopyMD5(t, func(src, dest *os.File) {
		expected := fmt.Sprintf("open %s: permission denied", src.Name())
		helperMust(t, os.Chmod(src.Name(), 0000))

		err := FileCopyMD5(src.Name(), dest.Name()).Start(context.Background())
		if err.Error() != expected {
			t.Errorf("unexpected error type, got: %s, want: %s", err, expected)
		}
	})
}

func TestFileCopyMD5DestNotReachable(t *testing.T) {
	helperFileCopyMD5(t, func(src, dest *os.File) {
		expected := fmt.Sprintf("open %s: permission denied", dest.Name())
		helperMust(t, os.Chmod(dest.Name(), 0000))

		err := FileCopyMD5(src.Name(), dest.Name()).Start(context.Background())
		if err.Error() != expected {
			t.Errorf("unexpected error type, got: %s, want: %s", err, expected)
		}
	})
}

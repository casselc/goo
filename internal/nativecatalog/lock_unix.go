//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package nativecatalog

import (
	"context"
	"errors"
	"os"
	"syscall"
	"time"
)

// Directory locks coordinate independent Goo processes without lock files or GET
// writes. The configured local saved directory must not be replaced while serving.
func lockDirectory(ctx context.Context, path string, exclusive bool) (func(), error) {
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	mode := syscall.LOCK_SH
	if exclusive {
		mode = syscall.LOCK_EX
	}
	for {
		e = syscall.Flock(int(f.Fd()), mode|syscall.LOCK_NB)
		if e == nil {
			return func() { f.Close() }, nil
		}
		if !errors.Is(e, syscall.EWOULDBLOCK) && !errors.Is(e, syscall.EAGAIN) {
			f.Close()
			return nil, e
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			f.Close()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

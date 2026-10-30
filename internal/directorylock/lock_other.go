//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package directorylock

import (
	"context"
	"errors"
)

func Acquire(context.Context, string, bool) (func(), error) {
	return nil, errors.New("directory admission requires local filesystem advisory locks on a supported Unix platform")
}

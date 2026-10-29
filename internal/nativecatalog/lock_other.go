//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package nativecatalog

import (
	"context"
	"errors"
)

func lockDirectory(context.Context, string, bool) (func(), error) {
	return nil, errors.New("native catalog publication requires local filesystem advisory directory locks on a supported Unix platform")
}

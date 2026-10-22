// Package immutablefile publishes caller-encoded bytes without replacing a winner.
// Callers own trusted paths, validation, identity, size limits and winner verification.
package immutablefile

import (
	"errors"
	"os"
	"path/filepath"
)

// Publish writes and syncs a temporary sibling, links it to name, and syncs dir.
// created is false when name already exists; no existing bytes are inspected or
// replaced. A successful link may precede a directory-sync error, so an error is
// not proof of non-publication. This is not a power-loss durability guarantee.
func Publish(dir, name, prefix string, data []byte) (created bool, err error) {
	return publish(dir, name, prefix, data, true, true)
}

// PublishNew refuses an existing final name with the original link error.
// Unlike Publish, it does not sync the directory after a collision.
func PublishNew(dir, name, prefix string, data []byte) error {
	_, err := publish(dir, name, prefix, data, false, true)
	return err
}

// PublishFileOnly keeps the idempotent link policy but omits directory sync.
// It preserves the historical custody/YAML write contract; callers still own
// winner verification. Successful return does not establish crash durability.
func PublishFileOnly(dir, name, prefix string, data []byte) (bool, error) {
	return publish(dir, name, prefix, data, true, false)
}

func publish(dir, name, prefix string, data []byte, acceptExisting, syncDirectory bool) (created bool, err error) {
	if err = os.MkdirAll(dir, 0700); err != nil {
		return false, err
	}
	f, err := os.CreateTemp(dir, prefix)
	if err != nil {
		return false, err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return false, err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return false, err
	}
	if err = f.Close(); err != nil {
		return false, err
	}
	err = os.Link(f.Name(), filepath.Join(dir, name))
	if err != nil && (!acceptExisting || !errors.Is(err, os.ErrExist)) {
		return false, err
	}
	created = err == nil
	if !syncDirectory {
		return created, nil
	}
	d, err := os.Open(dir)
	if err != nil {
		return created, err
	}
	defer d.Close()
	return created, d.Sync()
}

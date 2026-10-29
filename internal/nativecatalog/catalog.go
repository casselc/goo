// Package nativecatalog owns bounded native definition discovery and publication.
// It does not execute telemetry or modify an existing immutable definition.
package nativecatalog

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/casselc/goo/internal/immutablefile"
	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

const Limit = 256
const FileBytes = 32768

var ErrCapacity = errors.New("native package catalog capacity reached; existing revisions remain accessible by exact ID")
var revisionPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Store struct{ BuiltinDir, SavedDir string }
type Entry struct {
	Definition measurements.Definition
	Revision   string
	Builtin    bool
}

func read(path string) (measurements.Definition, string, error) {
	var zero measurements.Definition
	info, e := os.Lstat(path)
	if e != nil {
		return zero, "", e
	}
	if !info.Mode().IsRegular() || info.Size() > FileBytes {
		return zero, "", errors.New("package file exceeds limit or is not regular")
	}
	f, e := os.Open(path)
	if e != nil {
		return zero, "", e
	}
	defer f.Close()
	info, e = f.Stat()
	if e != nil || !info.Mode().IsRegular() {
		return zero, "", errors.New("package file is not regular")
	}
	b, e := io.ReadAll(io.LimitReader(f, FileBytes+1))
	if e != nil {
		return zero, "", e
	}
	if len(b) > FileBytes {
		return zero, "", errors.New("package file exceeds limit")
	}
	d, e := measurements.Read(string(b))
	if e != nil {
		return zero, "", e
	}
	rev, e := d.Revision()
	return d, rev, e
}
func scan(ctx context.Context, dir string) (map[string]measurements.Definition, int, error) {
	all := map[string]measurements.Definition{}
	if dir == "" {
		return all, 0, nil
	}
	f, e := os.Open(dir)
	if errors.Is(e, os.ErrNotExist) {
		return all, 0, nil
	}
	if e != nil {
		return nil, 0, e
	}
	defer f.Close()
	files, e := f.ReadDir(Limit + 1)
	if e != nil && e != io.EOF {
		return nil, 0, e
	}
	if len(files) > Limit {
		return nil, 0, ErrCapacity
	}
	for _, file := range files {
		if e := ctx.Err(); e != nil {
			return nil, 0, e
		}
		if filepath.Ext(file.Name()) != ".yaml" {
			continue
		}
		d, rev, e := read(filepath.Join(dir, file.Name()))
		if e != nil {
			return nil, 0, e
		}
		name := strings.TrimSuffix(file.Name(), ".yaml")
		if revisionPattern.MatchString(name) && name != rev {
			return nil, 0, errors.New("saved package filename differs from its revision")
		}
		all[rev] = d
	}
	return all, len(files), nil
}
func (s Store) catalog(ctx context.Context) ([]Entry, int, error) {
	builtin, _, e := scan(ctx, s.BuiltinDir)
	if e != nil {
		return nil, 0, e
	}
	saved, physical, e := scan(ctx, s.SavedDir)
	if e != nil {
		return nil, 0, e
	}
	all := map[string]Entry{}
	for rev, d := range builtin {
		all[rev] = Entry{d, rev, true}
	}
	for rev, d := range saved {
		all[rev] = Entry{d, rev, all[rev].Builtin}
	}
	if len(all) > Limit {
		return nil, 0, ErrCapacity
	}
	entries := make([]Entry, 0, len(all))
	for _, entry := range all {
		entries = append(entries, entry)
	}
	slices.SortFunc(entries, func(a, b Entry) int { return strings.Compare(a.Revision, b.Revision) })
	return entries, physical, nil
}
func (s Store) List(ctx context.Context) ([]Entry, error) {
	unlock, e := lockDirectory(ctx, s.SavedDir, false)
	if errors.Is(e, os.ErrNotExist) {
		s.SavedDir = ""
		entries, _, err := s.catalog(ctx)
		return entries, err
	}
	if e != nil {
		return nil, e
	}
	defer unlock()
	entries, _, e := s.catalog(ctx)
	return entries, e
}
func verify(path, revision string) error {
	_, actual, e := read(path)
	if e != nil || actual != revision {
		return errors.New("existing measurement package is unavailable, invalid, or has a different revision")
	}
	return nil
}
func (s Store) Publish(ctx context.Context, d measurements.Definition) error {
	revision, e := d.Revision()
	if e != nil {
		return e
	}
	b, e := yaml.Marshal(d)
	if e != nil {
		return e
	}
	if e = os.MkdirAll(s.SavedDir, 0700); e != nil {
		return e
	}
	unlock, e := lockDirectory(ctx, s.SavedDir, true)
	if e != nil {
		return e
	}
	defer unlock()
	path := filepath.Join(s.SavedDir, revision+".yaml")
	if _, e = os.Lstat(path); e == nil {
		return verify(path, revision)
	} else if !errors.Is(e, os.ErrNotExist) {
		return e
	}
	if len(b) > FileBytes {
		return fmt.Errorf("canonical package YAML exceeds %d bytes", FileBytes)
	}
	entries, physical, e := s.catalog(ctx)
	if e != nil {
		return e
	}
	exists := false
	for _, entry := range entries {
		exists = exists || entry.Revision == revision
	}
	if physical >= Limit || len(entries) >= Limit && !exists {
		return ErrCapacity
	}
	if e = ctx.Err(); e != nil {
		return e
	}
	created, e := immutablefile.PublishFileOnly(s.SavedDir, revision+".yaml", ".package-", b)
	if e != nil {
		return e
	}
	if !created {
		return verify(path, revision)
	}
	return nil
}

// Resolve bypasses catalog capacity only for an exact saved hash filename. It
// still verifies contents and refuses a corrupt winner before builtin fallback.
func (s Store) Resolve(ctx context.Context, revision string) (Entry, error) {
	if !revisionPattern.MatchString(revision) {
		return Entry{}, errors.New("invalid native package revision")
	}
	unlock, e := lockDirectory(ctx, s.SavedDir, false)
	if errors.Is(e, os.ErrNotExist) {
		builtin, _, err := scan(ctx, s.BuiltinDir)
		if err != nil {
			return Entry{}, err
		}
		if d, ok := builtin[revision]; ok {
			return Entry{d, revision, true}, nil
		}
		return Entry{}, os.ErrNotExist
	}
	if e != nil {
		return Entry{}, e
	}
	defer unlock()
	d, actual, e := read(filepath.Join(s.SavedDir, revision+".yaml"))
	if e == nil {
		if actual != revision {
			return Entry{}, errors.New("native package revision mismatch")
		}
		return Entry{d, revision, false}, nil
	}
	if !errors.Is(e, os.ErrNotExist) {
		return Entry{}, e
	}
	builtin, _, e := scan(ctx, s.BuiltinDir)
	if e != nil {
		return Entry{}, e
	}
	if d, ok := builtin[revision]; ok {
		return Entry{d, revision, true}, nil
	}
	return Entry{}, os.ErrNotExist
}

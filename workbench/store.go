// Package workbench stores authored context separately from telemetry evidence.
package workbench

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const MaxBytes = 128 << 10
const CatalogLimit = 10000

var ID = regexp.MustCompile(`^[a-f0-9]{64}$`)

type Store struct{ Dir string }

func text(s string, n int, required, multiline bool) bool {
	return utf8.ValidString(s) && len(s) <= n && (!required || strings.TrimSpace(s) != "") && strings.IndexFunc(s, func(r rune) bool { return unicode.IsControl(r) && (!multiline || r != '\n' && r != '\t' && r != '\r') }) < 0
}
func digest(v any) (string, error) {
	b, e := json.Marshal(v)
	if e != nil {
		return "", e
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
func (s Store) read(kind, id string, v any) error {
	if !ID.MatchString(id) {
		return errors.New("invalid reference ID")
	}
	f, e := os.Open(filepath.Join(s.Dir, kind, id+".json"))
	if e != nil {
		return e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, MaxBytes+1))
	if e != nil {
		return e
	}
	if len(b) > MaxBytes {
		return errors.New("artifact exceeds 128 KiB")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if e = d.Decode(v); e != nil {
		return e
	}
	if d.Decode(new(any)) != io.EOF {
		return errors.New("expected one artifact")
	}
	return nil
}
func (s Store) publish(kind, id string, v any) error {
	b, e := json.Marshal(v)
	if e != nil {
		return e
	}
	if len(b) > MaxBytes {
		return errors.New("artifact exceeds 128 KiB")
	}
	dir := filepath.Join(s.Dir, kind)
	if e = os.MkdirAll(dir, 0700); e != nil {
		return e
	}
	f, e := os.CreateTemp(dir, ".draft-")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	if _, e = f.Write(b); e != nil {
		f.Close()
		return e
	}
	if e = f.Sync(); e != nil {
		f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Link(f.Name(), filepath.Join(dir, id+".json")); e != nil && !errors.Is(e, os.ErrExist) {
		return e
	}
	d, e := os.Open(dir)
	if e != nil {
		return e
	}
	defer d.Close()
	return d.Sync()
}
func (s Store) ids(kind string) ([]string, error) {
	d, e := os.Open(filepath.Join(s.Dir, kind))
	if errors.Is(e, os.ErrNotExist) {
		return []string{}, nil
	}
	if e != nil {
		return nil, e
	}
	defer d.Close()
	entries, e := d.ReadDir(CatalogLimit + 1)
	if e != nil && e != io.EOF {
		return nil, e
	}
	if len(entries) > CatalogLimit {
		return nil, fmt.Errorf("%s catalog scan exceeds %d entries; direct IDs remain available", kind, CatalogLimit)
	}
	ids := []string{}
	for _, entry := range entries {
		id := strings.TrimSuffix(entry.Name(), ".json")
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") && ID.MatchString(id) {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	return ids, nil
}
func validTime(s string) bool { _, e := time.Parse(time.RFC3339Nano, s); return e == nil }

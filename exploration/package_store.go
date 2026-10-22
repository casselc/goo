package exploration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/casselc/goo/internal/immutablefile"
)

var regexpRevision = regexp.MustCompile(`^[a-f0-9]{64}$`)

func stringsTrimJSON(name string) string {
	if !strings.HasSuffix(name, ".json") {
		return ""
	}
	return strings.TrimSuffix(name, ".json")
}

// PackageStore publishes immutable definitions, not evidence or source manifests.
type PackageStore struct{ Dir string }
type SavedPackage struct {
	Revision   string           `json:"revision"`
	Definition TelemetryPackage `json:"definition"`
}

func (s PackageStore) Save(p TelemetryPackage) (*SavedPackage, error) {
	id, e := p.Revision()
	if e != nil {
		return nil, e
	}
	data, e := json.Marshal(p)
	if e != nil {
		return nil, e
	}
	if len(data) > PackageBytes {
		return nil, fmt.Errorf("definition exceeds 32 KiB")
	}
	if _, e = immutablefile.Publish(s.Dir, id+".json", ".package-", data); e != nil {
		return nil, e
	}
	return s.Load(id)
}
func (s PackageStore) Load(id string) (*SavedPackage, error) {
	if !regexpRevision.MatchString(id) {
		return nil, fmt.Errorf("invalid package revision")
	}
	f, e := os.Open(filepath.Join(s.Dir, id+".json"))
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, PackageBytes+1))
	if e != nil {
		return nil, e
	}
	if len(b) > PackageBytes {
		return nil, fmt.Errorf("stored package exceeds limit")
	}
	p, e := DecodePackageJSON(b)
	if e != nil {
		return nil, e
	}
	actual, e := p.Revision()
	if e != nil || actual != id {
		return nil, fmt.Errorf("stored package digest mismatch")
	}
	return &SavedPackage{id, p}, nil
}
func (s PackageStore) List() ([]SavedPackage, error) {
	entries, e := os.ReadDir(s.Dir)
	if errors.Is(e, os.ErrNotExist) {
		return []SavedPackage{}, nil
	}
	if e != nil {
		return nil, e
	}
	out := []SavedPackage{}
	for _, entry := range entries {
		if entry.IsDir() || !regexpRevision.MatchString(stringsTrimJSON(entry.Name())) {
			continue
		}
		if len(out) >= 200 {
			return nil, fmt.Errorf("package library exceeds 200 revisions; use revision lookup or manage local storage")
		}
		p, e := s.Load(stringsTrimJSON(entry.Name()))
		if e != nil {
			return nil, e
		}
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Definition.Name != out[j].Definition.Name {
			return out[i].Definition.Name < out[j].Definition.Name
		}
		return out[i].Revision < out[j].Revision
	})
	return out, nil
}

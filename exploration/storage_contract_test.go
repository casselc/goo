package exploration

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPackageStorageContract(t *testing.T) {
	wire, e := os.ReadFile("../testdata/storage/telemetry-package.json")
	if e != nil {
		t.Fatal(e)
	}
	var ids struct {
		Revision string `json:"telemetryPackageRevision"`
	}
	b, e := os.ReadFile("../testdata/storage/identities.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &ids); e != nil {
		t.Fatal(e)
	}
	p, e := DecodePackageJSON(wire)
	if e != nil {
		t.Fatal(e)
	}
	id, e := p.Revision()
	if e != nil || id != ids.Revision {
		t.Fatalf("identity changed: %s %v", id, e)
	}
	s := PackageStore{Dir: t.TempDir()}
	saved, e := s.Save(p)
	if e != nil || saved.Revision != id {
		t.Fatal(saved, e)
	}
	file := filepath.Join(s.Dir, id+".json")
	check := func(want []byte) {
		t.Helper()
		got, e := os.ReadFile(file)
		if e != nil || !bytes.Equal(got, want) {
			t.Fatalf("stored bytes changed: %s %v", got, e)
		}
	}
	check(wire)
	if _, e = (PackageStore{Dir: s.Dir}).Load(id); e != nil {
		t.Fatal("reopen", e)
	}
	if _, e = s.Save(p); e != nil {
		t.Fatal("duplicate", e)
	}
	check(wire)
	if _, e = s.Load("../escape"); e == nil {
		t.Fatal("invalid ID accepted")
	}
	for name, bad := range map[string][]byte{"truncated": wire[:len(wire)/2], "foreign field": append(append([]byte{}, wire[:len(wire)-1]...), []byte(`,"foreign":1}`)...), "changed digest": bytes.Replace(wire, []byte("API spans"), []byte("API changed"), 1)} {
		t.Run(name, func(t *testing.T) {
			if e := os.WriteFile(file, bad, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := s.Load(id); e == nil {
				t.Fatal("corrupt file accepted")
			}
			if _, e := s.Save(p); e == nil {
				t.Fatal("duplicate concealed corrupt winner")
			}
			got, e := os.ReadFile(file)
			if e != nil || !bytes.Equal(got, bad) {
				t.Fatal("existing corrupt file overwritten", e)
			}
		})
	}
	entries, e := os.ReadDir(s.Dir)
	if e != nil || len(entries) != 1 {
		t.Fatal("temporary files leaked", entries, e)
	}
}

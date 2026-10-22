package workbench

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWorkbenchStorageContract(t *testing.T) {
	wire, e := os.ReadFile("../testdata/storage/deployment.json")
	if e != nil {
		t.Fatal(e)
	}
	var m Deployment
	if e = json.Unmarshal(wire, &m); e != nil {
		t.Fatal(e)
	}
	var ids struct{ DeploymentID string }
	b, e := os.ReadFile("../testdata/storage/identities.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &ids); e != nil {
		t.Fatal(e)
	}
	id := m.ID
	m.ID = ""
	hash, e := digest(m)
	m.ID = id
	if e != nil || hash != id || id != ids.DeploymentID {
		t.Fatal("identity changed", hash, e)
	}
	s := Store{Dir: t.TempDir()}
	if e = s.publish("deployments", id, m); e != nil {
		t.Fatal(e)
	}
	file := filepath.Join(s.Dir, "deployments", id+".json")
	check := func(want []byte) {
		t.Helper()
		got, e := os.ReadFile(file)
		if e != nil || !bytes.Equal(got, want) {
			t.Fatal("stored bytes changed", e)
		}
	}
	check(wire)
	if _, e = (Store{Dir: s.Dir}).LoadDeployment(id); e != nil {
		t.Fatal("reopen", e)
	}
	changed := m
	changed.Title = "Do not overwrite"
	if e = s.publish("deployments", id, changed); e != nil {
		t.Fatal("duplicate", e)
	}
	check(wire)
	if _, e = s.LoadDeployment("../escape"); e == nil {
		t.Fatal("invalid ID accepted")
	}
	for name, bad := range map[string][]byte{"truncated": wire[:len(wire)/2], "foreign field": append(append([]byte{}, wire[:len(wire)-1]...), []byte(`,"foreign":1}`)...), "changed digest": bytes.Replace(wire, []byte("API rollout"), []byte("API changed"), 1)} {
		t.Run(name, func(t *testing.T) {
			if e := os.WriteFile(file, bad, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := s.LoadDeployment(id); e == nil {
				t.Fatal("corrupt file accepted")
			}
			if e := s.publish("deployments", id, m); e != nil {
				t.Fatal(e)
			}
			check(bad)
			if _, e := s.LoadDeployment(id); e == nil {
				t.Fatal("duplicate repaired corruption implicitly")
			}
		})
	}
	entries, e := os.ReadDir(filepath.Dir(file))
	if e != nil || len(entries) != 1 {
		t.Fatal("temporary files leaked", entries, e)
	}
}

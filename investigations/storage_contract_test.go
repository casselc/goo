package investigations

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInvestigationCaptureStorageContract(t *testing.T) {
	wire, e := os.ReadFile("../testdata/storage/investigation-capture.json")
	if e != nil {
		t.Fatal(e)
	}
	var c Capture
	if e = json.Unmarshal(wire, &c); e != nil {
		t.Fatal(e)
	}
	ids, e := os.ReadFile("../testdata/storage/identities.json")
	if e != nil {
		t.Fatal(e)
	}
	var expected map[string]string
	if e = json.Unmarshal(ids, &expected); e != nil {
		t.Fatal(e)
	}
	id, canonical, e := digest(c)
	if e != nil || id != c.ID || id != expected["investigationCaptureID"] {
		t.Fatal("identity changed", id, e)
	}
	if !bytes.Contains(canonical, []byte("18446744073709551615")) {
		t.Fatal("canonical number rounded")
	}
	reordered := c
	reordered.Evidence.Operations = json.RawMessage(`{"count":18446744073709551615,"basis":"one","spans":[]}`)
	same, _, e := digest(reordered)
	if e != nil || same != id {
		t.Fatal("embedded key order changed identity", e)
	}
	s := Store{Dir: t.TempDir()}
	file := filepath.Join(s.Dir, id+".json")
	if e = os.WriteFile(file, wire, 0600); e != nil {
		t.Fatal(e)
	}
	loaded, e := (Store{Dir: s.Dir}).Load(id)
	if e != nil || !bytes.Equal(loaded.Evidence.Operations, c.Evidence.Operations) {
		t.Fatal("reopen changed evidence", e)
	}
	roundtrip, e := json.Marshal(loaded)
	if e != nil || !bytes.Equal(roundtrip, wire) {
		t.Fatal("loaded response changed", e)
	}
	q := c.Question
	q.ToMS++
	if _, e = s.Save(q, c.Selection, c.Evidence, id); e == nil || e.Error() != "refresh must retain the original question and package revision" {
		t.Fatal("parent mismatch accepted", e)
	}
	child, e := s.Save(c.Question, c.Selection, c.Evidence, id)
	if e != nil || child.Parent != id {
		t.Fatal("valid refresh", e)
	}
	got, e := os.ReadFile(file)
	if e != nil || !bytes.Equal(got, wire) {
		t.Fatal("parent overwritten", e)
	}
	for name, bad := range map[string][]byte{"truncated": wire[:len(wire)/2], "digest mismatch": bytes.Replace(wire, []byte("18446744073709551615"), []byte("18446744073709551614"), 1), "trailing value": append(append([]byte{}, wire...), []byte(` {}`)...)} {
		t.Run(name, func(t *testing.T) {
			if e := os.WriteFile(file, bad, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := s.Load(id); e == nil {
				t.Fatal("corruption accepted")
			}
		})
	}
	// This loader historically ignores unknown fields; keep that policy explicit.
	unknown := append(append([]byte{}, wire[:len(wire)-1]...), []byte(`,"foreign":1}`)...)
	if e = os.WriteFile(file, unknown, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Load(id); e != nil {
		t.Fatal("unknown-field policy changed", e)
	}
	if _, e = s.Load("../escape"); e == nil {
		t.Fatal("invalid reference accepted")
	}
}

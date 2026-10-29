package exploration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplorationCaptureStorageContract(t *testing.T) {
	wire, e := os.ReadFile("../testdata/storage/exploration-capture.json")
	if e != nil {
		t.Fatal(e)
	}
	var c SavedCapture
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
	id, _, e := captureDigest(c)
	if e != nil || id != c.ID || id != expected["explorationCaptureID"] {
		t.Fatal("identity changed", id, e)
	}
	s := CaptureStore{Dir: t.TempDir()}
	saved, e := s.publish(c)
	if e != nil || saved.ID != id {
		t.Fatal(saved, e)
	}
	file := filepath.Join(s.Dir, id+".json")
	got, e := os.ReadFile(file)
	if e != nil || !bytes.Equal(got, wire) {
		t.Fatal("stored bytes changed", e)
	}
	if _, e = s.publish(c); e != nil {
		t.Fatal("duplicate", e)
	}
	reopened, e := (CaptureStore{Dir: s.Dir}).Load(id)
	if e != nil {
		t.Fatal(e)
	}
	if string(reopened.Explore.Rows[0]) != string(c.Explore.Rows[0]) || !strings.Contains(string(reopened.Explore.Rows[0]), "18446744073709551615") {
		t.Fatal("large integer changed")
	}
	page, e := reopened.Page("")
	if e != nil {
		t.Fatal(e)
	}
	response, e := json.Marshal(page)
	if e != nil {
		t.Fatal(e)
	}
	want, e := os.ReadFile("../testdata/storage/exploration-capture-page.json")
	if e != nil || !bytes.Equal(response, want) {
		t.Fatal("page response changed", e)
	}
	changed := c.Question
	q := *changed.Explore
	q.Environment = "other"
	changed.Explore = &q
	if _, e = s.Create(context.Background(), Source{}, c.Title, changed, id); e == nil || e.Error() != "refresh must retain the original question" {
		t.Fatal("changed refresh question not rejected before source read", e)
	}
	if _, e = s.Load("../escape"); e == nil {
		t.Fatal("invalid reference accepted")
	}
	if _, e = reopened.Page(captureCursor(strings.Repeat("0", 64), 0)); e == nil {
		t.Fatal("foreign cursor accepted")
	}
	for name, bad := range map[string][]byte{"truncated": wire[:len(wire)/2], "unknown field": append(append([]byte{}, wire[:len(wire)-1]...), []byte(`,"foreign":1}`)...), "digest mismatch": bytes.Replace(wire, []byte("API evidence"), []byte("Changed evidence"), 1), "trailing value": append(append([]byte{}, wire...), []byte(` {}`)...)} {
		t.Run(name, func(t *testing.T) {
			if e := os.WriteFile(file, bad, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := s.Load(id); e == nil {
				t.Fatal("corruption accepted")
			}
			if _, e := s.publish(c); e == nil {
				t.Fatal("duplicate concealed corrupt winner")
			}
			got, e := os.ReadFile(file)
			if e != nil || !bytes.Equal(got, bad) {
				t.Fatal("existing file overwritten", e)
			}
		})
	}
	entries, e := os.ReadDir(s.Dir)
	if e != nil || len(entries) != 2 || entries[0].Name() != ".catalog-v2" || !entries[0].IsDir() || entries[1].Name() != id+".json" {
		t.Fatal("temporary leak", e)
	}
}

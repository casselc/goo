package investigations

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestImmutableRevisions(t *testing.T) {
	s := Store{t.TempDir()}
	q := Question{"checkout", "v1", 60000, 120000}
	e := Evidence{Operations: json.RawMessage(`{"spans":[],"basis":"one","count":18446744073709551615}`)}
	a, err := s.Save(q, json.RawMessage(`{"filter":{"outcome":"error"}}`), e, "")
	if err != nil {
		t.Fatal(err)
	}
	// A new store object reopens evidence without depending on any source or memory.
	got, err := (Store{s.Dir}).Load(a.ID)
	if err != nil || string(got.Evidence.Operations) != string(a.Evidence.Operations) {
		t.Fatal(got, err)
	}
	e.Operations = json.RawMessage(`{"spans":[],"basis":"late"}`)
	b, err := s.Save(q, a.Selection, e, a.ID)
	if err != nil || b.Parent != a.ID || b.ID == a.ID || b.Question != a.Question {
		t.Fatal(b, err)
	}
	again, err := s.Load(a.ID)
	if err != nil || string(again.Evidence.Operations) != string(a.Evidence.Operations) {
		t.Fatal("predecessor changed", err)
	}
	q.ToMS++
	if _, err = s.Save(q, a.Selection, e, a.ID); err == nil {
		t.Fatal("changed question accepted")
	}
	if _, err = s.Load("../../escape"); err == nil {
		t.Fatal("invalid ID accepted")
	}
	path := filepath.Join(s.Dir, a.ID+".json")
	raw, _ := os.ReadFile(path)
	var v map[string]any
	json.Unmarshal(raw, &v)
	v["capturedAt"] = "tampered"
	raw, _ = json.Marshal(v)
	os.WriteFile(path, raw, 0600)
	if _, err = s.Load(a.ID); err == nil {
		t.Fatal("tampered receipt accepted")
	}
}
func TestConcurrentSaves(t *testing.T) {
	s := Store{t.TempDir()}
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, e := s.Save(Question{"p", "v", 60000, 120000}, nil, Evidence{Operations: json.RawMessage(`{}`)}, "")
			if e != nil {
				t.Error(e)
				return
			}
			if _, e = s.Load(c.ID); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
}

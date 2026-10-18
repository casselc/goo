package workbench

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func fixtureDeployment(at int64) DeploymentInput {
	return DeploymentInput{Environment: "prod", Service: "gateway", Title: "Release rollout", Revision: "v2", OccurredAtMS: at}
}
func TestDeploymentScopeCustodyAndIntegrity(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	a, e := s.CreateDeployment(fixtureDeployment(100))
	if e != nil {
		t.Fatal(e)
	}
	b, e := s.CreateDeployment(fixtureDeployment(200))
	if e != nil {
		t.Fatal(e)
	}
	other := fixtureDeployment(150)
	other.Environment = "other"
	s.CreateDeployment(other)
	r, e := s.Deployments(DeploymentScope{"prod", 100, 200})
	if e != nil || len(r.Items) != 1 || r.Items[0].ID != a.ID || r.Items[0].OccurredAtMS != 100 || !validTime(a.RecordedAt) {
		t.Fatal(r, e)
	}
	if a.RecordedAt == b.RecordedAt {
		t.Fatal("custody collapsed")
	}
	bad := fixtureDeployment(100)
	bad.URL = "javascript:alert(1)"
	if _, e = s.CreateDeployment(bad); e == nil {
		t.Fatal("executable URL accepted")
	}
	bad.URL = "https://user:password@example.test"
	if _, e = s.CreateDeployment(bad); e == nil {
		t.Fatal("credential URL accepted")
	}
	bad.URL = "https://example.test/releases/v2"
	if _, e = s.CreateDeployment(bad); e != nil {
		t.Fatal(e)
	}
	path := filepath.Join(s.Dir, "deployments", a.ID+".json")
	data, _ := os.ReadFile(path)
	os.WriteFile(path, []byte(strings.Replace(string(data), "v2", "v3", 1)), 0600)
	if _, e = s.LoadDeployment(a.ID); e == nil {
		t.Fatal("edited receipt accepted")
	}
	if _, e = s.Deployments(DeploymentScope{"prod", 100, 200}); e == nil {
		t.Fatal("invalid context silently became empty")
	}
}
func TestDocumentReferencesBranchingAndPreservedOrder(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	m, _ := s.CreateDeployment(fixtureDeployment(100))
	draft := DocumentInput{Title: "Regression investigation", Cards: []Card{{Kind: "note", Role: "hypothesis", Text: "Deployment could be related.\nCheck downstream timing."}, {Kind: "deployment", Reference: m.ID}}}
	first, e := s.CreateDocument(draft, "", nil)
	if e != nil {
		t.Fatal(e)
	}
	draft.Cards = append(draft.Cards, Card{Kind: "note", Role: "conclusion", Text: "Need more evidence."})
	var wg sync.WaitGroup
	children := make(chan *Document, 2)
	errors := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			input := draft
			input.Cards = append([]Card(nil), draft.Cards...)
			input.Cards[len(input.Cards)-1].Text = fmt.Sprintf("Alternative %d", i)
			d, e := s.CreateDocument(input, first.ID, nil)
			children <- d
			errors <- e
		}()
	}
	wg.Wait()
	close(children)
	close(errors)
	for e := range errors {
		if e != nil {
			t.Fatal(e)
		}
	}
	ids := map[string]bool{}
	for d := range children {
		if d.Parent != first.ID || d.Root != first.ID || len(d.Cards) != 3 {
			t.Fatal(d)
		}
		ids[d.ID] = true
	}
	if len(ids) != 2 {
		t.Fatal("branches collapsed")
	}
	original, e := s.LoadDocument(first.ID)
	if e != nil || len(original.Cards) != 2 {
		t.Fatal(original, e)
	}
	bad := draft
	bad.Cards = []Card{{Kind: "capture", Reference: strings.Repeat("a", 64)}}
	if _, e = s.CreateDocument(bad, "", func(id string) error { return fmt.Errorf("missing") }); e == nil {
		t.Fatal("missing reference accepted")
	}
	if _, e = s.CreateDocument(bad, "", nil); e == nil {
		t.Fatal("unchecked reference accepted")
	}
	if _, e = s.CreateDocument(bad, "", func(id string) error { return nil }); e != nil {
		t.Fatal(e)
	}
	list, e := s.Documents("")
	if e != nil || len(list.Items) != 4 {
		t.Fatal(list, e)
	}
	data, _ := json.Marshal(first)
	forged := strings.Replace(string(data), "Regression investigation", "Changed investigation", 1)
	os.WriteFile(filepath.Join(s.Dir, "documents", first.ID+".json"), []byte(forged), 0600)
	if _, e = s.LoadDocument(first.ID); e == nil {
		t.Fatal("corrupt narrative accepted")
	}
}
func TestMarkerLimitAndDocumentPagination(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	for i := 0; i < 201; i++ {
		if _, e := s.CreateDeployment(fixtureDeployment(int64(i))); e != nil {
			t.Fatal(e)
		}
	}
	if r, e := s.Deployments(DeploymentScope{"prod", 0, 1000}); e == nil || r != nil {
		t.Fatal("partial context presented as complete")
	}
	draft := DocumentInput{Title: "Notes", Cards: []Card{{Kind: "note", Role: "finding", Text: "Observed delay"}}}
	for i := 0; i < 22; i++ {
		if _, e := s.CreateDocument(draft, "", nil); e != nil {
			t.Fatal(e)
		}
	}
	a, e := s.Documents("")
	if e != nil || len(a.Items) != 20 || a.NextCursor == "" {
		t.Fatal(a, e)
	}
	b, e := s.Documents(a.NextCursor)
	if e != nil || len(b.Items) != 2 || b.NextCursor != "" {
		t.Fatal(b, e)
	}
}

package workbench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/casselc/goo/identity"
)

func meaningInput() MeaningInput {
	return MeaningInput{Target: MeaningTarget{Kind: "service", Source: "clickstack", Environment: "prod", Service: "checkout", Namespace: &identity.Namespace{Mode: "missing"}}, Tags: []string{"team:payments", "incident:latency"}, Note: "Investigate", State: "active", Applies: &MeaningWindow{FromMS: 100, ToMS: 200}}
}
func TestMeaningBranchesWithdrawalsAndImmutableHistory(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	in := meaningInput()
	root, e := s.CreateMeaning(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	in.Tags[0] = "mutated"
	in.Target.Namespace.Mode = "value"
	in.Applies.FromMS = 0
	if root.Tags[1] != "team:payments" || root.Target.Namespace.Mode != "missing" || root.Applies.FromMS != 100 {
		t.Fatal("caller input aliases revision")
	}
	children := make([]*MeaningRevision, 2)
	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range children {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			v := meaningInput()
			v.Parent = root.ID
			v.Note = []string{"First review", "Independent review"}[i]
			children[i], errs[i] = s.CreateMeaning(v, nil)
		}(i)
	}
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			t.Fatal(e)
		}
	}
	page, e := s.Meanings(&root.Target, "team:payments", "", false)
	if e != nil || len(page.Items) != 2 {
		t.Fatal(page, e)
	}
	v := meaningInput()
	v.Parent = children[0].ID
	v.State = "withdrawn"
	v.Tags = []string{}
	v.Note = ""
	withdraw, e := s.CreateMeaning(v, nil)
	if e != nil {
		t.Fatal(e)
	}
	page, e = s.Meanings(&root.Target, "team:payments", "", false)
	if e != nil || len(page.Items) != 1 || page.Items[0].Revision.ID != children[1].ID {
		t.Fatal(page, e)
	}
	hist, e := s.Meanings(&root.Target, "", "", true)
	if e != nil || len(hist.Items) != 4 {
		t.Fatal(hist, e)
	}
	old, e := s.LoadMeaning(root.ID)
	if e != nil || old.Note != "Investigate" {
		t.Fatal(old, e)
	}
	v = meaningInput()
	v.Parent = withdraw.ID
	v.Note = "Restored"
	if _, e = s.CreateMeaning(v, nil); e != nil {
		t.Fatal(e)
	}
	v.Target.Namespace = &identity.Namespace{Mode: "value", Value: ""}
	if _, e = s.CreateMeaning(v, nil); e == nil {
		t.Fatal("retargeted parent")
	}
	raw, _ := json.Marshal(old)
	var corrupt map[string]any
	json.Unmarshal(raw, &corrupt)
	corrupt["note"] = "tampered"
	raw, _ = json.Marshal(corrupt)
	os.WriteFile(filepath.Join(s.Dir, "meanings", root.ID+".json"), raw, 0600)
	if _, e = s.LoadMeaning(root.ID); e == nil {
		t.Fatal("accepted tampered revision")
	}
}
func TestMeaningValidationAndExactScope(t *testing.T) {
	cases := []func(*MeaningInput){
		func(v *MeaningInput) { v.Target.Namespace = nil }, func(v *MeaningInput) { v.Target.Source = "other" }, func(v *MeaningInput) { v.Target.ID = strings.Repeat("a", 64) },
		func(v *MeaningInput) { v.Applies = nil }, func(v *MeaningInput) { v.Applies.ToMS = v.Applies.FromMS }, func(v *MeaningInput) { v.Applies.ToMS = 8640000000000001 },
		func(v *MeaningInput) { v.Tags = []string{"UPPER"} }, func(v *MeaningInput) { v.Tags = []string{"same", "same"} }, func(v *MeaningInput) { v.Tags = nil; v.Note = " \n\t" },
		func(v *MeaningInput) { v.Note = strings.Repeat("é", 2001) }, func(v *MeaningInput) { v.State = "withdrawn" }, func(v *MeaningInput) { v.Parent = "../escape" },
	}
	for i, change := range cases {
		v := meaningInput()
		change(&v)
		if v.Validate() == nil {
			t.Errorf("invalid case %d", i)
		}
	}
	s := Store{Dir: t.TempDir()}
	for _, ns := range []*identity.Namespace{{Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "shop"}} {
		v := meaningInput()
		v.Target.Namespace = ns
		if _, e := s.CreateMeaning(v, nil); e != nil {
			t.Fatal(e)
		}
		p, e := s.Meanings(&v.Target, "", "", false)
		if e != nil || len(p.Items) != 1 {
			t.Fatal(p, e)
		}
	}
	v := meaningInput()
	v.Target = MeaningTarget{Kind: "package", ID: strings.Repeat("a", 64)}
	v.Applies = nil
	if _, e := s.CreateMeaning(v, nil); e == nil {
		t.Fatal("unresolved package accepted")
	}
	calls := 0
	if _, e := s.CreateMeaning(v, func(id string) error {
		calls++
		if id != v.Target.ID {
			t.Fatal(id)
		}
		return nil
	}); e != nil || calls != 1 {
		t.Fatal(e, calls)
	}
	v.Target.Kind = "document"
	if _, e := s.CreateMeaning(v, nil); e == nil {
		t.Fatal("nonexistent document")
	}
}
func TestMeaningCatalogPaginationAndTagSupersession(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	for i := 0; i < 23; i++ {
		if _, e := s.CreateMeaning(meaningInput(), nil); e != nil {
			t.Fatal(e)
		}
	}
	first, e := s.Meanings(nil, "team:payments", "", false)
	if e != nil || len(first.Items) != 20 || first.NextCursor == "" {
		t.Fatal(first, e)
	}
	last, e := s.Meanings(nil, "team:payments", first.NextCursor, false)
	if e != nil || len(last.Items) != 3 || last.NextCursor != "" {
		t.Fatal(last, e)
	}
	if _, e = s.Meanings(nil, "absent", first.NextCursor, false); e == nil {
		t.Fatal("cross-filter cursor")
	}
	v := meaningInput()
	v.Parent = first.NextCursor
	v.Tags = []string{"new"}
	if _, e = s.CreateMeaning(v, nil); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Meanings(nil, "team:payments", first.NextCursor, false); e == nil {
		t.Fatal("obsolete cursor should require restart")
	}
	p, e := s.Meanings(nil, "new", "", false)
	if e != nil || len(p.Items) != 1 {
		t.Fatal(p, e)
	}
}

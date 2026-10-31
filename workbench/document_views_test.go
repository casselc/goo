package workbench

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDocumentPinsSavedViewAndKeepsEvidenceSeparate(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	view, err := store.CreateExplorationView(timelineComposition())
	if err != nil {
		t.Fatal(err)
	}
	original := *view
	captureID := strings.Repeat("c", 64)
	in := DocumentInput{Title: "Scoped finding", Cards: []Card{{Kind: "view", Reference: view.ID, Caption: "Count timeline", Text: "Captured records changed; cause is unproven."}, {Kind: "capture", Reference: captureID}}}
	resolved := 0
	doc, err := store.CreateDocument(in, "", func(id string) error {
		resolved++
		if id != captureID {
			t.Fatal("definition sent to evidence resolver")
		}
		return nil
	})
	if err != nil || resolved != 1 {
		t.Fatal(doc, err, resolved)
	}
	in.Cards = append([]Card(nil), in.Cards...)
	in.Cards[0].Text = "Revised finding"
	child, err := store.CreateDocument(in, doc.ID, func(string) error { return nil })
	if err != nil || child.Parent != doc.ID || child.Root != doc.ID {
		t.Fatal(child, err)
	}
	old, err := store.LoadDocument(doc.ID)
	if err != nil || old.Cards[0].Text == "Revised finding" || old.Cards[0].Reference != view.ID {
		t.Fatal(old, err)
	}
	again, err := store.LoadExplorationView(view.ID)
	if err != nil || !reflect.DeepEqual(*again, original) {
		t.Fatal("definition mutated", err)
	}
	// Reopening a document verifies its own custody, not availability of referenced data.
	if err = os.Remove(filepath.Join(store.Dir, "exploration-views", view.ID+".json")); err != nil {
		t.Fatal(err)
	}
	if _, err = store.LoadDocument(doc.ID); err != nil {
		t.Fatal("held document depends on live definition availability", err)
	}
	if _, err = store.CreateDocument(in, "", func(string) error { return nil }); err == nil {
		t.Fatal("publication accepted unavailable definition")
	}
}

func TestViewCardValidationAndDefinitionIntegrity(t *testing.T) {
	id := strings.Repeat("a", 64)
	valid := Card{Kind: "view", Reference: id, Caption: "Saved definition", Text: "雪\nFinding"}
	if err := valid.Validate(); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*Card){func(c *Card) { c.Reference = "latest" }, func(c *Card) { c.Step = &Step{} }, func(c *Card) { c.Key = id }, func(c *Card) { c.Evidence = []string{id} }, func(c *Card) { c.Review = id }, func(c *Card) { c.ReadBasis = "basis" }, func(c *Card) { c.Role = "finding" }, func(c *Card) { c.Text = strings.Repeat("x", 4001) }, func(c *Card) { c.Caption = strings.Repeat("x", 161) }} {
		bad := valid
		mutate(&bad)
		if bad.Validate() == nil {
			t.Fatalf("invalid definition card accepted: %#v", bad)
		}
	}
	store := Store{Dir: t.TempDir()}
	if _, err := store.CreateDocument(DocumentInput{Title: "Missing", Cards: []Card{valid}}, "", nil); err == nil {
		t.Fatal("missing definition accepted")
	}
	v, err := store.CreateExplorationView(compositionView())
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(store.Dir, "exploration-views", v.ID+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	raw = []byte(strings.Replace(string(raw), v.Title, "tampered", 1))
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = store.CreateDocument(DocumentInput{Title: "Tampered", Cards: []Card{{Kind: "view", Reference: v.ID}}}, "", nil); err == nil {
		t.Fatal("tampered definition accepted")
	}
}

package workbench

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestQuestionFirstInvestigationAndPreservedSummary(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	in := DocumentInput{Title: "Delayed work", Question: "Did accepted work finish?", Finding: "Not yet established", OpenQuestions: "Could telemetry be incomplete?", Cards: nil}
	a, e := s.CreateDocument(in, "", nil)
	if e != nil {
		t.Fatal(e)
	}
	if a.Cards == nil {
		t.Fatal("question-only evidence list not normalized")
	}
	in.Finding = "Captured completion observed"
	b, e := s.CreateDocument(in, a.ID, nil)
	if e != nil || b.Parent != a.ID || b.Root != a.ID || b.Question != a.Question {
		t.Fatal(b, e)
	}
	old, e := s.LoadDocument(a.ID)
	if e != nil || old.Finding != "Not yet established" {
		t.Fatal("overwrote original summary", old, e)
	}
	in.OpenQuestions = strings.Repeat("x", 4001)
	if in.Validate() == nil {
		t.Fatal("unbounded brief accepted")
	}
	if (DocumentInput{Title: "Empty"}).Validate() == nil {
		t.Fatal("empty investigation accepted")
	}
}
func TestLegacyDocumentHashRemainsReadable(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	// Exact pre-brief JSON shape/order, independent of the expanded Document type.
	type legacyInput struct {
		Title string `json:"title"`
		Cards []Card `json:"cards"`
	}
	type legacyDocument struct {
		ID         string `json:"id"`
		Version    int    `json:"version"`
		Parent     string `json:"parent,omitempty"`
		Root       string `json:"root,omitempty"`
		RecordedAt string `json:"recordedAt"`
		legacyInput
	}
	old := legacyDocument{Version: 1, RecordedAt: "2026-10-18T00:00:00Z", legacyInput: legacyInput{Title: "Legacy", Cards: []Card{{Kind: "note", Role: "finding", Text: "Observed"}}}}
	id, e := digest(old)
	if e != nil {
		t.Fatal(e)
	}
	old.ID = id
	if e = s.publish("documents", id, old); e != nil {
		t.Fatal(e)
	}
	loaded, e := s.LoadDocument(id)
	if e != nil || loaded.ID != id || loaded.Question != "" {
		t.Fatal("legacy receipt changed", e)
	}
	b, _ := json.Marshal(loaded)
	if strings.Contains(string(b), "openQuestions") {
		t.Fatal("empty fields changed legacy shape")
	}
}

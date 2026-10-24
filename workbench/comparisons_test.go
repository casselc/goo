package workbench

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func comparisonCaptures() (*exploration.SavedCapture, *exploration.SavedCapture) {
	makeCapture := func(id, count string) *exploration.SavedCapture {
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
		return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{ID: id, Version: 1, Title: "Stored spans", CapturedAt: "2026-10-19T00:00:00Z", Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Question: q, Count: count, State: "observed", Basis: strings.Repeat("c", 64)}}
	}
	return makeCapture(strings.Repeat("a", 64), "2"), makeCapture(strings.Repeat("b", 64), "4")
}
func TestSavedComparisonFreezesDecisionOrderAndRevisedFinding(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	a, b := comparisonCaptures()
	expected, e := exploration.CompareEvidence(a, b)
	if e != nil {
		t.Fatal(e)
	}
	load := func(id string) (*exploration.SavedCapture, error) {
		if id == a.ID {
			return a, nil
		}
		if id == b.ID {
			return b, nil
		}
		return nil, fmt.Errorf("missing")
	}
	receipt, e := s.CreateComparison(a.ID, b.ID, expected, load)
	if e != nil {
		t.Fatal(e)
	}
	if receipt.Result.CountDifference != "2" || receipt.Result.Baseline.Receipt.ID != a.ID || receipt.Algorithm != ComparisonAlgorithm {
		t.Fatal(receipt)
	}
	in := DocumentInput{Title: "Latency finding", Cards: []Card{{Kind: "comparison", Reference: receipt.ID, Text: "Observed, not causal."}}}
	original, e := s.CreateDocument(in, "", nil)
	if e != nil {
		t.Fatal(e)
	}
	in.Cards = append([]Card(nil), in.Cards...)
	in.Cards[0].Text = "Revised finding"
	child, e := s.CreateDocument(in, original.ID, nil)
	if e != nil || child.Parent != original.ID {
		t.Fatal(child, e)
	}
	old, e := s.LoadDocument(original.ID)
	if e != nil || old.Cards[0].Text != "Observed, not causal." {
		t.Fatal(old, e)
	}
	// Source removal or evaluator changes cannot redefine the saved snapshot.
	a.Explore.Count = "999"
	b.Explore.Count = "1"
	again, e := s.LoadComparison(receipt.ID)
	if e != nil || !reflect.DeepEqual(again, receipt) {
		t.Fatal(again, e)
	}
	forged := *expected
	forged.State = "incompatible"
	if _, e = s.CreateComparison(a.ID, b.ID, &forged, load); e != ErrComparisonChanged {
		t.Fatal("forged decision accepted", e)
	}
	raw, e := os.ReadFile(filepath.Join(s.Dir, "comparisons", receipt.ID+".json"))
	if e != nil {
		t.Fatal(e)
	}
	var disk ComparisonReceipt
	json.Unmarshal(raw, &disk)
	disk.Result.CountDifference = "999"
	raw, _ = json.Marshal(disk)
	os.WriteFile(filepath.Join(s.Dir, "comparisons", receipt.ID+".json"), raw, 0600)
	if _, e = s.LoadComparison(receipt.ID); e == nil {
		t.Fatal("modified snapshot accepted")
	}
}
func TestComparisonCardValidationAndPublicationFailure(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	id := strings.Repeat("a", 64)
	for _, c := range []Card{{Kind: "comparison", Reference: "x"}, {Kind: "comparison", Reference: id, Text: strings.Repeat("x", 4001)}, {Kind: "note", Role: "finding", Text: "ok", Reference: id}, {Kind: "capture", Reference: id, Text: "finding"}} {
		if c.Validate() == nil {
			t.Fatal("bad card accepted", c)
		}
	}
	if (Card{Kind: "comparison", Reference: id, Text: "雪\nTyped finding"}).Validate() != nil {
		t.Fatal("valid finding rejected")
	}
	if _, e := s.CreateDocument(DocumentInput{Title: "Missing", Cards: []Card{{Kind: "comparison", Reference: id}}}, "", nil); e == nil {
		t.Fatal("missing comparison reference accepted")
	}
	if _, e := s.CreateComparison(id, id, nil, nil); e == nil {
		t.Fatal("same ID accepted")
	}
	a, b := comparisonCaptures()
	b.Question.Explore.Environment = "staging"
	b.Explore.Question.Environment = "staging"
	expected, _ := exploration.CompareEvidence(a, b)
	receipt, e := s.CreateComparison(a.ID, b.ID, expected, func(id string) (*exploration.SavedCapture, error) {
		if id == a.ID {
			return a, nil
		}
		return b, nil
	})
	if e != nil || receipt.Result.State != "incompatible" || receipt.Result.CountDifference != "" {
		t.Fatal(receipt, e)
	}
	// Frozen receipts remain valid when document publication fails.
	if _, e = s.CreateDocument(DocumentInput{Title: ""}, "", nil); e == nil {
		t.Fatal("bad document accepted")
	}
	if _, e = s.LoadComparison(receipt.ID); e != nil {
		t.Fatal("receipt lost", e)
	}
}

func TestLegacyComparisonReceiptStillReopensWithoutReevaluation(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	a, b := comparisonCaptures()
	expected, _ := exploration.CompareEvidence(a, b)
	r, e := s.CreateComparison(a.ID, b.ID, expected, func(id string) (*exploration.SavedCapture, error) {
		if id == a.ID {
			return a, nil
		}
		return b, nil
	})
	if e != nil {
		t.Fatal(e)
	}
	for _, algorithm := range []string{"goo/evidence-comparison/v1", "goo/evidence-comparison/v2"} {
		r.ID = ""
		r.Algorithm = algorithm
		id, err := digest(r)
		if err != nil {
			t.Fatal(err)
		}
		r.ID = id
		if err = s.publish("comparisons", id, r); err != nil {
			t.Fatal(err)
		}
		a.Explore.Count = "99"
		got, err := s.LoadComparison(id)
		if err != nil || got.Algorithm != algorithm || got.Result.CountDifference != "2" {
			t.Fatal(got, err)
		}
	}
}

func TestNumericEvidenceReceiptFreezesVersionedMeaning(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	a, b := comparisonCaptures()
	binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.items"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	raw := func(s string) *string { return &s }
	for i, c := range []*exploration.SavedCapture{a, b} {
		c.Question.Explore.Numeric = &binding
		c.Explore.Question.Numeric = &binding
		values := []*string{raw("-1"), raw("0")}
		if i == 1 {
			values = append(values, raw("1"), nil)
		}
		summary, e := exploration.EvaluateNumericValues(binding, values)
		if e != nil {
			t.Fatal(e)
		}
		c.Explore.Numeric = &summary
	}
	expected, e := exploration.CompareEvidence(a, b)
	if e != nil {
		t.Fatal(e)
	}
	load := func(id string) (*exploration.SavedCapture, error) {
		if id == a.ID {
			return a, nil
		}
		return b, nil
	}
	receipt, e := store.CreateComparison(a.ID, b.ID, expected, load)
	if e != nil || receipt.Algorithm != NumericComparisonAlgorithm || receipt.Result.Measures == nil || receipt.Result.Measures.Numeric == nil {
		t.Fatal("numeric version/meaning lost", receipt, e)
	}
	original, _ := json.Marshal(receipt)
	reopened, e := store.LoadComparison(receipt.ID)
	if e != nil {
		t.Fatal(e)
	}
	again, _ := json.Marshal(reopened)
	if string(original) != string(again) {
		t.Fatal("reopen changed immutable numeric bytes")
	}
	// A legacy receipt freezes its historical count-only result even when the new
	// numeric comparison engine could now derive additional measure differences.
	legacy := *receipt
	legacy.ID = ""
	legacy.Algorithm = ComparisonAlgorithm
	cloned, _ := json.Marshal(receipt.Result)
	legacy.Result = nil
	json.Unmarshal(cloned, &legacy.Result)
	legacy.Result.Measures = nil
	id, e := digest(legacy)
	if e != nil {
		t.Fatal(e)
	}
	legacy.ID = id
	if e = store.publish("comparisons", id, legacy); e != nil {
		t.Fatal(e)
	}
	frozen, e := store.LoadComparison(id)
	if e != nil || frozen.Result.Measures != nil || frozen.Algorithm != ComparisonAlgorithm {
		t.Fatal("legacy evidence reinterpreted", e)
	}
	a.Explore.State = "changed_between_reads"
	unknown, e := exploration.CompareEvidence(a, b)
	if e != nil || unknown.State != "unknown" || unknown.Measures.Numeric.MeanDifference != "" {
		t.Fatal("unknown comparison asserted arithmetic", e)
	}
	unknownReceipt, e := store.CreateComparison(a.ID, b.ID, unknown, load)
	if e != nil || unknownReceipt.Algorithm != NumericComparisonAlgorithm {
		t.Fatal("new unknown evidence cannot reopen", e)
	}
	bad := *unknownReceipt
	bad.ID = ""
	bad.Algorithm = ComparisonAlgorithm
	badID, e := digest(bad)
	if e != nil {
		t.Fatal(e)
	}
	bad.ID = badID
	if e = store.publish("comparisons", badID, bad); e != nil {
		t.Fatal(e)
	}
	if _, e = store.LoadComparison(badID); e == nil {
		t.Fatal("legacy version accepted a new numeric decision")
	}
}

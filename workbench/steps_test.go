package workbench

import (
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"os"
	"strings"
	"testing"
)

func TestMixedQuestionStepsPreserveRevisionsAndScope(t *testing.T) {
	b, e := os.ReadFile("../config/measurements/worker-utilization.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(b))
	if e != nil {
		t.Fatal(e)
	}
	q := exploration.Query{Signal: "spans", Environment: d.Environment, FromMS: 100000, ToMS: 220000, Limit: 100}
	scope := StepScope{q.Environment, q.FromMS, q.ToMS}
	record := Step{Version: 1, Kind: "record", Scope: scope, View: "Records", Record: &q}
	metric := Step{Version: 1, Kind: "measurement", Scope: scope, View: "Overview", Measurement: &d}
	store := Store{Dir: t.TempDir()}
	in := DocumentInput{Title: "Mixed signals", Cards: []Card{{Kind: "question", Step: &record}, {Kind: "question", Step: &metric}}}
	first, e := store.CreateDocument(in, "", nil)
	if e != nil {
		t.Fatal(e)
	}
	before, _ := json.Marshal(first)
	in.Cards = in.Cards[1:]
	child, e := store.CreateDocument(in, first.ID, nil)
	if e != nil || child.Parent != first.ID {
		t.Fatal(child, e)
	}
	old, e := store.LoadDocument(first.ID)
	after, _ := json.Marshal(old)
	if e != nil || string(before) != string(after) {
		t.Fatal("reopening changed original intent", e)
	}
	bad := record
	bad.Scope.ToMS++
	if bad.Validate() == nil {
		t.Fatal("foreign record scope accepted")
	}
	bad = metric
	bad.Record = &q
	if bad.Validate() == nil {
		t.Fatal("mixed payload accepted")
	}
	bad = metric
	bad.Scope.Environment = "foreign"
	if bad.Validate() == nil {
		t.Fatal("foreign measurement environment accepted")
	}
	if (Card{Kind: "note", Role: "finding", Text: "x", Step: &metric}).Validate() == nil {
		t.Fatal("question hidden in note")
	}
}

func TestLinkedEvidenceRequiresResolverAndPreservesHistory(t *testing.T) {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
	step := Step{Version: 1, Kind: "record", Scope: StepScope{"prod", 1000, 2000}, View: "Records", Record: &q}
	key := strings.Repeat("a", 64)
	one, two := strings.Repeat("b", 64), strings.Repeat("c", 64)
	card := Card{Kind: "question", Key: key, Step: &step, Evidence: []string{one}}
	store := Store{Dir: t.TempDir()}
	in := DocumentInput{Title: "Evidence history", Cards: []Card{card}}
	if _, e := store.CreateDocument(in, "", nil); e == nil {
		t.Fatal("unchecked evidence accepted")
	}
	if _, e := store.CreateDocument(in, "", func(string) error { return errors.New("missing") }); e == nil {
		t.Fatal("missing evidence accepted")
	}
	resolver := func(id string) error {
		if id != one && id != two {
			return errors.New("missing")
		}
		return nil
	}
	first, e := store.CreateDocument(in, "", resolver)
	if e != nil {
		t.Fatal(e)
	}
	in.Cards[0].Evidence = append([]string{one}, two)
	child, e := store.CreateDocument(in, first.ID, resolver)
	if e != nil || len(child.Cards[0].Evidence) != 2 {
		t.Fatal(child, e)
	}
	old, e := store.LoadDocument(first.ID)
	if e != nil || len(old.Cards[0].Evidence) != 1 || old.Cards[0].Evidence[0] != one {
		t.Fatal("history overwritten")
	}
	in.Cards[0].Evidence = []string{one, one}
	if in.Validate() == nil {
		t.Fatal("duplicate evidence accepted")
	}
	in.Cards[0].Evidence = nil
	in.Cards = append(in.Cards, in.Cards[0])
	if in.Validate() == nil {
		t.Fatal("duplicate question key accepted")
	}
}

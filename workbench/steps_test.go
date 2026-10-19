package workbench

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"os"
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

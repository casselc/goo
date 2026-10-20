package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestHegelReusableQuestionRevisions(t *testing.T) {
	raw, e := os.ReadFile("../config/measurements/checkout-calls.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	hegel.Test(t, func(tc *hegel.T) {
		offset := hegel.Draw(tc, hegel.Integers(1, 999999))
		base := int64(1790000000000000000) + int64(offset)
		scope := workbench.StepScope{Environment: d.Environment, FromMS: base / 1000000, ToMS: (base + 2000000000 + 999999) / 1000000, FromNS: strconv.FormatInt(base, 10), ToNS: strconv.FormatInt(base+2000000000, 10)}
		step := workbench.Step{Version: 1, Kind: "measurement", Scope: scope, View: "Overview", Measurement: &d}
		id := strings.Repeat("a", 64)
		store := workbench.Store{Dir: t.TempDir()}
		input := workbench.DocumentInput{Title: "Reusable question", Cards: []workbench.Card{{Kind: "question", Key: strings.Repeat("b", 64), Step: &step, Evidence: []string{id}}}}
		resolver := func(ref string) error {
			if ref != id {
				return fmt.Errorf("missing")
			}
			return nil
		}
		first, e := store.CreateDocument(input, "", resolver)
		if e != nil {
			tc.Fatal(e)
		}
		snapshot, _ := json.Marshal(first)
		// A new question window is a revision; the stored exact question remains intact.
		next := step
		next.Scope.FromNS = ""
		next.Scope.ToNS = ""
		next.Scope.FromMS += 60000
		next.Scope.ToMS += 60000
		input.Cards = []workbench.Card{{Kind: "question", Key: strings.Repeat("c", 64), Step: &next}}
		child, e := store.CreateDocument(input, first.ID, resolver)
		if e != nil || child.Parent != first.ID {
			tc.Fatal(e)
		}
		old, e := store.LoadDocument(first.ID)
		after, _ := json.Marshal(old)
		if e != nil || string(snapshot) != string(after) || old.Cards[0].Step.Scope.FromNS != scope.FromNS || old.Cards[0].Evidence[0] != id {
			tc.Fatal("question or evidence changed")
		}
	}, options(100)...)
}
func TestQuintReusableQuestionsLive(t *testing.T) {
	file := os.Getenv("GOO_REUSABLE_QUESTIONS_LEDGER")
	if file == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Events []struct {
				Action           string
				Reads, Revisions int
			}
			Original, Saved, Rerun workbench.Document
			Captures               []exploration.CapturePage
			Questions              []measurements.Question
			Fresh                  measurements.Question
			After                  int
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Events[0].Reads++
		case 1:
			l.Saved.Cards[1].Evidence[0] = l.Captures[1].Receipt.ID
		case 2:
			l.Saved.Cards[2].Key = l.Saved.Cards[1].Key
		case 3:
			l.Saved.Cards[1].Step.Scope.FromNS = ""
		case 4:
			l.Fresh.Definition.Source.Metric = "wrong"
		case 5:
			l.Original.Cards[0].Text = "changed"
		}
		coherent := l.After == 2 && len(l.Saved.Cards) == 3 && len(l.Rerun.Cards) == 3 && reflect.DeepEqual(l.Original.Cards[0], l.Saved.Cards[0]) && l.Saved.Cards[1].Key != l.Saved.Cards[2].Key && l.Rerun.Parent == l.Saved.ID && reflect.DeepEqual(l.Fresh.Definition, l.Questions[0].Definition) && l.Fresh.FromNS == "" && len(l.Rerun.Cards[1].Evidence) == 2 && l.Rerun.Cards[1].Evidence[0] == l.Captures[0].Receipt.ID && reflect.DeepEqual(l.Rerun.Cards[1].Step, l.Saved.Cards[1].Step)
		for i, q := range l.Questions {
			c := l.Saved.Cards[i+1]
			coherent = coherent && len(c.Evidence) == 1 && c.Evidence[0] == l.Captures[i].Receipt.ID && c.Step.Scope.FromNS == q.FromNS && c.Step.Scope.ToNS == q.ToNS && reflect.DeepEqual(*c.Step.Measurement, q.Definition)
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.reads", event.Reads, "obs.revisions", event.Revisions, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/reusable_questions.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("live trace rejected")
	}
	for i := 0; i < 6; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

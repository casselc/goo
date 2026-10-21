package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestQuintTraceFindingLedger(t *testing.T) {
	name := os.Getenv("GOO_TRACE_FINDING_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Q                         exploration.Query
			Captures                  []exploration.CapturePage
			Doc                       workbench.Document
			BeforeReopen, AfterReopen int
			TraceID                   string
			Orphan                    exploration.CapturePage
		}
		if json.Unmarshal(raw, &l) != nil || len(l.Captures) != 2 || len(l.Doc.Cards) != 3 {
			t.Fatal("invalid live ledger")
		}
		switch mutation {
		case 0:
			l.Captures[1].Receipt.Question.Explore.Environment = "foreign"
		case 1:
			l.Captures[1].Receipt.Question.Explore.FromMS--
		case 2:
			l.Doc.Cards[1].Evidence = l.Doc.Cards[0].Evidence
		case 3:
			l.AfterReopen++
		case 4:
			l.TraceID = strings.Repeat("f", 32)
		}
		p, q := l.Captures[0].Receipt.Question.Explore, l.Captures[1].Receipt.Question.Explore
		coherent := p != nil && q != nil && p.Environment == q.Environment && p.FromMS == q.FromMS && p.ToMS == q.ToMS && reflect.DeepEqual(p.Where, l.Q.Where) && len(q.Where.All) == 1 && q.Where.All[0].Field.Key == "trace_id" && q.Where.All[0].Op == "eq" && q.Where.All[0].Value == l.TraceID && reflect.DeepEqual(l.Doc.Cards[0].Evidence, []string{l.Captures[0].Receipt.ID}) && reflect.DeepEqual(l.Doc.Cards[1].Evidence, []string{l.Captures[1].Receipt.ID}) && strings.Contains(l.Doc.Cards[2].Text, l.TraceID) && l.Doc.DocumentInput.Validate() == nil && l.Orphan.Receipt.ID != "" && l.Orphan.Receipt.ID != l.Captures[0].Receipt.ID
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		rec.Record(context.Background(), "failPublish", nil, "obs.reads", 1, "obs.linked", 0, "obs.coherent", coherent)
		rec.Record(context.Background(), "capture", nil, "obs.reads", 2, "obs.linked", 1, "obs.coherent", coherent)
		rec.Record(context.Background(), "capture", nil, "obs.reads", 3, "obs.linked", 2, "obs.coherent", coherent)
		rec.Record(context.Background(), "save", nil, "obs.reads", 3, "obs.linked", 2, "obs.coherent", coherent)
		rec.Record(context.Background(), "reopen", nil, "obs.reads", 3+l.AfterReopen-l.BeforeReopen, "obs.linked", 2, "obs.coherent", coherent)
		checker, e := validate.NewChecker("model/linked_evidence.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		r, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return r.OK()
	}
	if !check(-1) {
		t.Fatal("observed finding rejected")
	}
	for i := 0; i < 5; i++ {
		if check(i) {
			t.Fatal("invalid finding accepted", i)
		}
	}
}

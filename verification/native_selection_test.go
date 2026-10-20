package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestQuintNativeSelectionLedger(t *testing.T) {
	name := os.Getenv("GOO_NATIVE_SELECTION_LEDGER")
	if name == "" {
		t.Skip("real-source opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Initial, Selected, Cleared, Restored, Final                    exploration.DashboardPlan
			BeforeEdit, AfterEdit, AfterApply, BeforeClear, RecipientReads int
			Rendered                                                       []struct {
				Value    string
				Question workbench.Step
			}
			ChosenWidget      exploration.DashboardWidget
			Capture, Reopened exploration.CapturePage
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Selected.Scope.Environment = "other"
		case 1:
			l.Selected.NativeOptions[0].Selected = true
		case 2:
			l.AfterEdit++
		case 3:
			l.Rendered[0].Value = "999 1"
		case 4:
			l.ChosenWidget.Measurement.Definition.Source.Unit = "wrong"
		case 5:
			l.Capture.Receipt.Question.Measurement.Definition.Source.Namespace = &identity.Namespace{Mode: "value", Value: "other"}
		case 6:
			l.Reopened.Receipt.ID = "wrong"
		case 7:
			l.Cleared.Widgets = append(l.Cleared.Widgets, l.ChosenWidget)
		case 8:
			l.Final.Widgets = append(l.Final.Widgets, l.ChosenWidget)
		}
		count := func(p exploration.DashboardPlan) int {
			n := 0
			for _, w := range p.Widgets {
				if w.Measurement != nil {
					n++
				}
			}
			return n
		}
		coherent := func(p exploration.DashboardPlan) bool {
			keys := map[string]bool{}
			for _, o := range p.NativeOptions {
				if o.Selected {
					keys[o.Key] = true
				}
			}
			if p.NativeLimit != 2 || len(keys) > 2 || count(p) != len(keys) || !reflect.DeepEqual(p.Scope, l.Initial.Scope) {
				return false
			}
			for _, w := range p.Widgets {
				if w.Measurement == nil {
					continue
				}
				q := w.Measurement
				if !keys[measurements.Hash(q.Definition)] || q.Validate() != nil || q.Definition.Environment != p.Scope.Environment || q.FromMS != p.Scope.FromMS || q.ToMS != p.Scope.ToMS || !q.Definition.Source.Namespace.Equal(p.Scope.Namespace) || len(q.Definition.Source.Services) != 1 || q.Definition.Source.Services[0] != p.Service {
					return false
				}
			}
			return true
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		record("defaults", coherent(l.Initial) && count(l.Initial) == 2 && len(l.Initial.NativeOptions) == 4)
		record("draft", l.BeforeEdit == l.AfterEdit)
		changed := true
		for _, o := range l.Selected.NativeOptions {
			changed = changed && (o.Selected != o.Default)
		}
		values := len(l.Rendered) == 2
		for _, r := range l.Rendered {
			d := r.Question.Measurement
			expected := ""
			if d.Source.Metric == "queue.depth" {
				expected = "44 {item}"
			} else if d.Source.Version == "2" {
				expected = "22 1"
			} else if strings.HasSuffix(d.Source.SchemaURL, "-alternate") {
				expected = "33 1"
			} else {
				expected = "11 1"
			}
			values = values && r.Value == expected && d.Source.ExactStream != nil && d.Source.ExactStream.Resource["service.namespace"] == "payments"
		}
		record("apply", coherent(l.Selected) && changed && values && l.AfterApply-l.AfterEdit == 2)
		q := l.ChosenWidget.Measurement
		found := false
		for _, w := range l.Selected.Widgets {
			if w.Measurement != nil && reflect.DeepEqual(w.Measurement, q) {
				found = true
			}
		}
		record("capture", found && reflect.DeepEqual(l.Capture.Receipt.Question.Measurement, q) && reflect.DeepEqual(l.Capture.Measurement.Question, *q))
		record("reopen", l.RecipientReads == 0 && reflect.DeepEqual(l.Capture, l.Reopened))
		record("clear", coherent(l.Cleared) && count(l.Cleared) == 0)
		record("restore", coherent(l.Restored) && reflect.DeepEqual(l.Restored.Widgets, l.Initial.Widgets))
		record("supersede", coherent(l.Final) && count(l.Final) == 0)
		checker, e := validate.NewChecker("model/native_selection.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual native selection ledger rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatal("corrupt native selection ledger accepted", i)
		}
	}
}

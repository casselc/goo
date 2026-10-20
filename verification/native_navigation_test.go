package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintNativeNavigationLedger(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_NAVIGATION_LEDGER")
	if path == "" {
		t.Skip("actual browser ledger opt-in")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	type view struct {
		Basis              string
		Question           measurements.Question
		Instances, Reviews []string
		Feedback           string
		Open               bool
	}
	check := func(mutation int) bool {
		var l struct {
			Capture, Reopened                                           exploration.CapturePage
			Inventory, PageTwo, Review, Empty, Found, Cleared, Restored view
			Focused                                                     struct{ Question measurements.Question }
			Reads                                                       int
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Inventory.Instances = l.Inventory.Instances[:24]
		case 1:
			l.PageTwo.Instances = []string{"worker-00"}
		case 2:
			l.Review.Reviews[0] = "false"
		case 3:
			l.Empty.Instances = []string{"worker-27"}
		case 4:
			l.Found.Basis = "other"
		case 5:
			l.Cleared.Instances = l.Cleared.Instances[:1]
		case 6:
			l.Restored.Open = false
		case 7:
			l.Focused.Question.Definition.Source.ExactStream.Resource["service.instance.id"] = "worker-00"
		case 8:
			l.Reads = 1
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(action string, v view, condition bool) {
			same := v.Basis == l.Capture.Measurement.Basis && reflect.DeepEqual(v.Question, l.Capture.Measurement.Question)
			rec.Record(context.Background(), action, nil, "obs.coherent", condition && same && l.Reads == 0)
		}
		record("inventory", l.Inventory, len(l.Inventory.Instances) == 25 && len(l.Capture.Measurement.Counters.Series) == 28)
		record("page", l.PageTwo, reflect.DeepEqual(l.PageTwo.Instances, []string{"worker-25", "worker-26", "worker-27"}))
		record("review", l.Review, reflect.DeepEqual(l.Review.Instances, []string{"worker-26", "worker-27"}) && reflect.DeepEqual(l.Review.Reviews, []string{"true", "true"}))
		record("empty", l.Empty, len(l.Empty.Instances) == 0)
		record("find", l.Found, reflect.DeepEqual(l.Found.Instances, []string{"worker-27"}) && l.Found.Open)
		record("clear", l.Cleared, len(l.Cleared.Instances) == 25)
		record("restore", l.Restored, l.Restored.Open && reflect.DeepEqual(l.Restored.Instances, l.Found.Instances))
		selected := l.Focused.Question.Definition.Source.ExactStream
		record("focus", l.Restored, selected != nil && selected.Resource["service.instance.id"] == "worker-27" && l.Focused.Question.FromNS == l.Capture.Measurement.Question.FromNS && l.Focused.Question.ToNS == l.Capture.Measurement.Question.ToNS && reflect.DeepEqual(l.Reopened, l.Capture))
		checker, e := validate.NewChecker("model/native_navigation.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual native navigation ledger rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatal("corrupt navigation trace accepted", i)
		}
	}
}

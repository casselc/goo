package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type questionApplied struct {
	Question            exploration.Query `json:"question"`
	Basis, Count, State string
}
type questionBounds struct{ FromMS, ToMS int64 }
type questionEvent struct {
	Action     string
	Token      *int
	Generation int
	Phase      string
	Applied    *questionApplied
	Requested  *exploration.Query
	Bounds     *questionBounds
}

func sameRecordQuestion(a, b *exploration.Query) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}
func TestQuintQuestionWorkbenchLive(t *testing.T) {
	name := os.Getenv("GOO_QUESTION_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	type ledger struct {
		Events                   []questionEvent
		Applied, Narrowed, Final exploration.Result
		Inventory                exploration.InventoryResult
		MetricQuestion           struct {
			Service, Environment string
			FromMS, ToMS         int64
		}
		Saved, Reopened                                                                   exploration.CapturePage
		BeforeReturn, AfterReturn, BeforeHandoff, AfterHandoff, BeforeReopen, AfterReopen int
	}
	check := func(mutation int) bool {
		var l ledger
		if json.Unmarshal(data, &l) != nil {
			t.Fatal("invalid ledger")
		}
		switch mutation {
		case 0:
			for i := range l.Events {
				if l.Events[i].Action == "publish" {
					l.Events[i].Generation++
					break
				}
			}
		case 1:
			for i := range l.Events {
				if l.Events[i].Action == "publish" {
					l.Events[i].Applied.Question.Environment = "foreign"
					break
				}
			}
		case 2:
			l.MetricQuestion.FromMS++
		case 3:
			l.AfterReopen++
		case 4:
			l.AfterReturn++
		case 5:
			l.Saved.Explore.Count = "999"
		case 6:
			for i := range l.Events {
				if l.Events[i].Action == "reject" {
					l.Events[i].Action = "publish"
					break
				}
			}
		case 7:
			l.Narrowed.Question.ToMS++
		}
		coherent := len(l.Events) > 0 && l.Applied.Count == "6" && l.Narrowed.Count == "3" && l.Final.Count == "1" && l.Applied.Question.FromMS == l.Narrowed.Question.FromMS && l.Applied.Question.ToMS == l.Narrowed.Question.ToMS && l.Final.Question.FromMS == l.Applied.Question.FromMS && l.Final.Question.ToMS == l.Applied.Question.ToMS && l.MetricQuestion.FromMS == l.Inventory.Question.FromMS && l.MetricQuestion.ToMS == l.Inventory.Question.ToMS && l.MetricQuestion.Environment == l.Inventory.Question.Environment && l.MetricQuestion.Service == "worker" && l.BeforeReturn == l.AfterReturn && l.BeforeHandoff == l.AfterHandoff && l.BeforeReopen == l.AfterReopen && reflect.DeepEqual(l.Saved, l.Reopened) && l.Saved.Explore != nil && l.Saved.Explore.Count == l.Final.Count && sameRecordQuestion(l.Saved.Receipt.Question.Explore, &l.Final.Question)
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		previous := questionEvent{Generation: 0, Phase: "initial"}
		seenReject := false
		publishes := 0
		phases := map[string]int{"initial": 0, "draft": 1, "running": 2, "applied": 3, "cancelled": 4, "failed": 5}
		for _, event := range l.Events {
			valid := coherent
			generation := previous.Generation
			if event.Action == "edit" || event.Action == "start" || event.Action == "cancel" {
				generation++
			}
			valid = valid && event.Generation == generation
			switch event.Action {
			case "start":
				valid = valid && event.Phase == "running" && event.Requested != nil && event.Requested.Validate() == nil && event.Bounds != nil && event.Bounds.FromMS == event.Requested.FromMS && event.Bounds.ToMS == event.Requested.ToMS
			case "publish":
				publishes++
				valid = valid && previous.Phase == "running" && event.Token != nil && *event.Token == generation && event.Phase == "applied" && event.Applied != nil && sameRecordQuestion(previous.Requested, &event.Applied.Question) && event.Applied.Basis != ""
			case "reject":
				seenReject = true
				valid = valid && event.Phase == previous.Phase && event.Token != nil && (*event.Token != generation || previous.Phase != "running") && reflect.DeepEqual(event.Applied, previous.Applied)
			case "edit":
				valid = valid && event.Phase == "draft" && event.Requested == nil && reflect.DeepEqual(event.Applied, previous.Applied) && (event.Bounds == nil || reflect.DeepEqual(event.Bounds, previous.Bounds))
			case "cancel":
				valid = valid && event.Phase == "cancelled" && event.Requested == nil
			case "fail":
				valid = valid && previous.Phase == "running" && event.Phase == "failed" && event.Token != nil && *event.Token == generation && event.Requested == nil
			default:
				valid = false
			}
			action := event.Action
			if action == "fail" {
				action = "failedRead"
			}
			rec.Record(context.Background(), action, nil, "obs.generation", event.Generation, "obs.phase", phases[event.Phase], "obs.coherent", valid)
			previous = event
		}
		if !seenReject || publishes < 3 {
			return false
		}
		checker, e := validate.NewChecker("model/question_session.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("live question lifecycle rejected")
	}
	for i := 0; i < 8; i++ {
		if check(i) {
			t.Fatal("mutated lifecycle accepted", i)
		}
	}
}

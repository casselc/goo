package verification

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

func TestQuintExactStreamLedger(t *testing.T) {
	name := os.Getenv("GOO_EXACT_STREAM_LEDGER")
	if name == "" {
		t.Skip("real source opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Broad                                      measurements.Question
			All, Selected, Restored, Inventory, Absent measurements.Result
			Plan, Cleared                              struct {
				Definition measurements.Definition
				Question   measurements.Question
			}
			Stored                     struct{ Definition measurements.Definition }
			Baseline, Frozen, Reopened exploration.CapturePage
			Fresh                      struct {
				Comparison exploration.EvidenceComparison
				Basis      string
			}
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.All.Counters.Series = l.All.Counters.Series[:1]
		case 1:
			l.Plan.Question.FromNS = "0"
		case 2:
			l.Selected.Question.Definition.Source.ExactStream = nil
		case 3:
			l.Stored.Definition.Source.ExactStream = nil
		case 4:
			l.Fresh.Comparison.NativeWindows[0].Selection.Rate = "350"
		case 5:
			l.Frozen.Measurement.Basis = "other"
		case 6:
			l.Cleared.Definition.Source.ExactStream = l.Plan.Definition.Source.ExactStream
		case 7:
			l.Absent.State = "observed"
		case 8:
			delete(l.Plan.Definition.Source.ExactStream.Resource, "service.namespace")
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		selected := l.Plan.Definition.Source.ExactStream
		record("inventory", len(l.All.Counters.Series) == 2 && l.Broad.Definition.Source.ExactStream == nil && len(l.Inventory.Counters.Series) == 7)
		_, emptyNamespace := selected.Resource["service.namespace"]
		record("focus", selected != nil && emptyNamespace && selected.ResourceSchemaURL != "" && l.Plan.Question.FromNS == l.Broad.FromNS && l.Plan.Question.ToNS == l.Broad.ToNS)
		record("evaluate", reflect.DeepEqual(l.Selected.Question.Definition.Source.ExactStream, selected) && len(l.Selected.Counters.Series) == 1 && reflect.DeepEqual(l.Selected.Counters.Series[0].Stream, *selected))
		record("save", reflect.DeepEqual(l.Stored.Definition.Source.ExactStream, selected) && reflect.DeepEqual(l.Baseline.Receipt.Question.Measurement.Definition.Source.ExactStream, selected))
		record("live", l.Fresh.Comparison.State == "comparable" && l.Fresh.Comparison.NativeWindows[0].Selection.Rate == "5" && reflect.DeepEqual(l.Fresh.Comparison.NativeWindows[0].Stream, *selected) && l.Absent.State == "no_data" && len(l.Absent.Counters.Series) == 0)
		record("freeze", l.Frozen.Measurement.Basis == l.Fresh.Basis && reflect.DeepEqual(l.Frozen, l.Reopened) && reflect.DeepEqual(l.Frozen.Measurement.Question.Definition.Source.ExactStream, selected))
		record("clear", l.Cleared.Definition.Source.ExactStream == nil && len(l.Restored.Counters.Series) == 2)
		checker, e := validate.NewChecker("model/exact_stream.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("actual stream selection trace rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatal("corrupt stream trace accepted", i)
		}
	}
}

func TestHegelExactStreamFocusClearState(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(2, 8))
		d := measurements.Definition{APIVersion: measurements.Version, Name: "requests", Title: "Requests", Kind: "counter", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 10000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "otel.http", Metric: "requests", Unit: "1"}}
		q := measurements.Question{Definition: d, FromMS: 100000, ToMS: 102000}
		rows := []json.RawMessage{}
		for i := 0; i < count; i++ {
			r := map[string]any{"resource": map[string]string{"service.name": "api", "service.instance.id": string(rune('a' + i)), "deployment.environment.name": "prod"}, "scope": "otel.http", "name": "requests", "unit": "1", "attributes": map[string]string{"route": ""}, "start_ns": "100000000000", "end_ns": "102000000000", "value": "10", "temporality": 1, "monotonic": true}
			raw, _ := json.Marshal(r)
			rows = append(rows, raw)
		}
		broad, e := measurements.Evaluate(q, rows, time.Unix(103, 0))
		if e != nil {
			tc.Fatal(e)
		}
		for step := 0; step < 12; step++ {
			action := hegel.Draw(tc, hegel.Integers(0, count))
			var input []json.RawMessage
			var selected *metrics.Stream
			if action == count {
				input = rows
			} else {
				stream := broad.Counters.Series[action].Stream
				selected = &stream
				input = []json.RawMessage{rows[action]}
			}
			q.Definition.Source.ExactStream = selected
			result, e := measurements.Evaluate(q, input, time.Unix(103, 0))
			if e != nil {
				tc.Fatal(e)
			}
			if selected == nil {
				if len(result.Counters.Series) != count {
					tc.Fatal("clear retained hidden filter")
				}
			} else {
				if len(result.Counters.Series) != 1 || !reflect.DeepEqual(result.Counters.Series[0].Stream, *selected) {
					tc.Fatal("focus changed identity")
				}
				if _, e := measurements.Evaluate(q, rows, time.Unix(103, 0)); e == nil {
					tc.Fatal("mixed stream source bypassed selector")
				}
				var changed map[string]any
				json.Unmarshal(input[0], &changed)
				changed["attributes"].(map[string]any)["new.dimension"] = ""
				offstream, _ := json.Marshal(changed)
				if _, e := measurements.Evaluate(q, []json.RawMessage{offstream}, time.Unix(103, 0)); e == nil {
					tc.Fatal("extra empty dimension silently matched")
				}
			}
			if len(broad.Counters.Series) != count || broad.Question.Definition.Source.ExactStream != nil {
				tc.Fatal("focus mutated inventory")
			}
		}
	}, options(100)...)
}

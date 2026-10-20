package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestHegelDashboardPlanning(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		from := hegel.Draw(tc, hegel.Integers(0, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 7200000))
		mask := hegel.Draw(tc, hegel.Integers(0, 7))
		in := exploration.DashboardRequest{Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: int64(from), ToMS: int64(from + width)}, Service: "literal'\\n", Fields: []exploration.Field{{Kind: "attribute", Key: "db.system"}, {Kind: "attribute", Key: "http.route"}}}
		for i, s := range []string{"spans", "logs", "metrics"} {
			if mask&(1<<i) != 0 {
				in.Signals = append(in.Signals, s)
			}
		}
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			in.Scope.Namespace = &identity.Namespace{Mode: "value", Value: "payments"}
		}
		before, _ := json.Marshal(in)
		p, e := exploration.PlanServiceDashboard(in, nil)
		after, _ := json.Marshal(in)
		if e != nil || string(before) != string(after) {
			tc.Fatal(p, e)
		}
		in.Fields[0], in.Fields[1] = in.Fields[1], in.Fields[0]
		for i, j := 0, len(in.Signals)-1; i < j; i, j = i+1, j-1 {
			in.Signals[i], in.Signals[j] = in.Signals[j], in.Signals[i]
		}
		again, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil || !reflect.DeepEqual(p, again) {
			tc.Fatal("order changes plan")
		}
		for _, w := range p.Widgets {
			if w.Query != nil && (w.Query.Validate() != nil || !exploration.Match(w.Query.Where, map[exploration.Field]string{{Kind: "column", Key: "service"}: in.Service, {Kind: "resource", Key: "service.namespace"}: "payments"}) || w.Query.Environment != in.Scope.Environment || w.Query.FromMS != in.Scope.FromMS || w.Query.ToMS != in.Scope.ToMS) {
				tc.Fatal("scope escaped", w)
			}
		}
	}, options(100)...)
}
func TestQuintServiceDashboardLive(t *testing.T) {
	name := os.Getenv("GOO_DASHBOARD_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		var l struct {
			Catalog exploration.ServiceDiscovery
			Events  []struct {
				Action string
				Reads  int
			}
			Plan, DeclarationPlan exploration.DashboardPlan
			Evidence              []struct {
				Widget   exploration.DashboardWidget
				Question json.RawMessage
			}
			Focused, Wider     workbench.Step
			Authored, Captured workbench.Document
			SavedCapture       exploration.CapturePage
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Plan.Scope.ToMS++
		case 1:
			l.Catalog.Dependencies = append(l.Catalog.Dependencies, exploration.Dependency{Kind: "edge", Parent: "invented", Child: "worker", Count: "1"})
		case 2:
			l.Focused.Record.Where.All[1].Value = "wrong"
		case 3:
			for _, e := range l.Evidence {
				if e.Widget.Measurement != nil {
					e.Widget.Measurement.Definition.Source.Unit = "wrong"
				}
			}
		case 4:
			l.Events[2].Reads++
		case 6:
			l.Events[0].Reads++
		case 7:
			l.Events[2].Action = "capture"
		case 5:
			l.SavedCapture.Explore.Count = "777"
		}
		coherent := l.Events[0].Reads == 1 && l.Plan.Scope.Environment == l.Catalog.Question.Environment && l.Plan.Scope.FromMS == l.Catalog.Question.FromMS && l.Plan.Scope.ToMS == l.Catalog.Question.ToMS && l.Plan.Scope.Namespace != nil && l.Plan.Scope.Namespace.Mode == "missing" && l.Plan.Service == "api" && l.Focused.Validate() == nil && l.Focused.Record.Where.All[1].Value == "/items" && l.Wider.Validate() == nil && l.Wider.Record.Where.All[0].Value == "worker" && l.Wider.Record.Where.All[1].Op == "missing" && l.SavedCapture.Explore.Count == "2" && len(l.Captured.Cards[0].Evidence) == 1 && len(l.DeclarationPlan.Declarations) > 0 && len(l.Evidence) == 4
		edges := 0
		for _, d := range l.Catalog.Dependencies {
			if d.Kind == "edge" {
				edges++
				coherent = coherent && d.Parent == "api" && d.Child == "worker" && d.Count == "1"
			}
		}
		coherent = coherent && edges == 1
		for _, ev := range l.Evidence {
			matches := false
			for _, w := range l.Plan.Widgets {
				if w.Key == ev.Widget.Key && reflect.DeepEqual(w, ev.Widget) {
					matches = true
				}
			}
			coherent = coherent && matches
			if ev.Widget.Query != nil {
				var q exploration.Query
				json.Unmarshal(ev.Question, &q)
				coherent = coherent && reflect.DeepEqual(q, *ev.Widget.Query) && q.Environment == l.Catalog.Question.Environment && q.FromMS == l.Catalog.Question.FromMS && q.ToMS == l.Catalog.Question.ToMS && q.Where.All[0].Value == "api" && q.Where.All[1].Op == "missing"
			} else {
				var q measurements.Question
				json.Unmarshal(ev.Question, &q)
				coherent = coherent && reflect.DeepEqual(q, *ev.Widget.Measurement) && q.Definition.Environment == l.Catalog.Question.Environment && q.Definition.Source.Services[0] == "api"
			}
		}
		coherent = coherent && l.Events[1].Reads == l.Events[2].Reads && l.Events[3].Reads == l.Events[2].Reads+1 && l.Events[4].Reads == l.Events[3].Reads
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range l.Events {
			rec.Record(context.Background(), event.Action, nil, "obs.coherent", coherent)
		}
		checker, e := validate.NewChecker("model/service_dashboard.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
		t.Fatal("live trace rejected")
	}
	for i := 0; i < 8; i++ {
		if check(i) {
			t.Fatal("mutation accepted", i)
		}
	}
}

// Generated state sequences exercise edits against a retained inventory. Planning
// must be deterministic, detached and bounded in every state, including clear.
func TestHegelDashboardSelectionSequences(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(3, 8))
		in := exploration.DashboardRequest{Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000, Namespace: &identity.Namespace{Mode: "value", Value: "payments"}}, Service: "api", Signals: []string{"metrics"}}
		for i := 0; i < count; i++ {
			name := fmt.Sprintf("metric-%d", i)
			d := measurements.Definition{APIVersion: measurements.Version, Name: name, Title: name, Kind: "gauge", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Namespace: in.Scope.Namespace.Clone(), Services: []string{"api"}, Scope: "sdk", Metric: name, Unit: "1"}}
			in.Instruments = append(in.Instruments, measurements.Instrument{Kind: "gauge", Metric: name, Unit: "1", Scope: "sdk", Supported: true, Definition: &d})
		}
		original, _ := json.Marshal(in.Instruments)
		base, e := exploration.PlanServiceDashboard(in, nil)
		if e != nil {
			tc.Fatal(e)
		}
		for step := 0; step < 12; step++ {
			mask := hegel.Draw(tc, hegel.Integers(0, (1<<count)-1))
			keys := []string{}
			for i, o := range base.NativeOptions {
				if mask&(1<<i) != 0 {
					keys = append(keys, o.Key)
				}
			}
			in.NativeSelection = &keys
			p, e := exploration.PlanServiceDashboard(in, nil)
			if len(keys) > exploration.DashboardNativeLimit {
				if e == nil {
					tc.Fatal("unbounded selection")
				}
				continue
			}
			if e != nil {
				tc.Fatal(e)
			}
			n := 0
			for _, w := range p.Widgets {
				if w.Measurement != nil {
					n++
					q := w.Measurement
					if q.Validate() != nil || !q.Definition.Source.Namespace.Equal(in.Scope.Namespace) || q.FromMS != in.Scope.FromMS || q.ToMS != in.Scope.ToMS {
						tc.Fatal("selection escaped scope")
					}
				}
			}
			if n != len(keys) {
				tc.Fatal("wrong native count")
			}
			for i, j := 0, len(in.Instruments)-1; i < j; i, j = i+1, j-1 {
				in.Instruments[i], in.Instruments[j] = in.Instruments[j], in.Instruments[i]
			}
			for i, j := 0, len(keys)-1; i < j; i, j = i+1, j-1 {
				keys[i], keys[j] = keys[j], keys[i]
			}
			again, e := exploration.PlanServiceDashboard(in, nil)
			if e != nil || !reflect.DeepEqual(p, again) {
				tc.Fatal("order-dependent selection")
			}
			// Returned metadata must not alias the retained source catalog.
			p.NativeOptions[0].Instrument.Definition.Source.Services[0] = "mutated"
		}
		sort.Slice(in.Instruments, func(i, j int) bool { return in.Instruments[i].Metric < in.Instruments[j].Metric })
		after, _ := json.Marshal(in.Instruments)
		if string(original) != string(after) {
			tc.Fatal("planner mutated source metadata")
		}
	}, options(100)...)
}

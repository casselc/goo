package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type nativeFixture386 struct {
	Page        workbench.NativeViewList
	Source      workbench.NativeView
	Destination measurements.Question
	Reused      workbench.NativeView
	Invalid     []json.RawMessage
}

// The oracle is the closed public sum type and complete destination question,
// not the compiler's SQL or a translated implementation of its validator.
func TestHegelNativeViewReuse386(t *testing.T) {
	fixtures := []nativeFixture386{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		gauge := hegel.Draw(tc, hegel.Booleans())
		if len(fixtures) < 2 {
			gauge = len(fixtures) == 0
		}
		family, selection, kind, category, value, unit := "histogram", "export-interval", "histogram", "bucket", "observation-count", "s"
		if gauge {
			family, selection, kind, category, value, unit = "gauge", "retained-observations", "timeline", "time", "recorded-value", []string{"1", "By", "Cel", "custom/雪", "m/s"}[seed%5]
		}
		if seed%2 == 0 {
			kind = "table"
		}
		ns := &identity.Namespace{Mode: "value", Value: fmt.Sprintf("team-%d'雪", seed)}
		if seed%3 == 0 {
			ns.Mode = "missing"
			ns.Value = ""
		} else if seed%3 == 1 {
			ns.Value = ""
		}
		sourceQ, _ := nrcInputs(nil, family)
		sourceQ.Definition.Source.Unit = unit
		destination, stream := nrcInputs(ns, family)
		d := &destination.Definition
		d.Environment = ""
		d.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("destination-%d", seed)}}
		if seed%4 == 0 {
			d.EnvironmentSelector.Policy = "current-only"
		} else if seed%4 == 1 {
			d.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
		} else if seed%4 == 2 {
			d.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		d.Name = fmt.Sprintf("destination-%d", seed)
		d.Title = "Destination observations"
		d.MaxGapMS = int64(1000 + seed)
		d.Source.Unit = unit
		d.Source.Services = []string{"destination-worker"}
		d.Source.Scope = "destination.instrumentation"
		d.Source.Version = "2"
		d.Source.SchemaURL = "https://opentelemetry.io/schemas/1.45.0"
		d.Source.Metric = "destination.recorded"
		d.Source.Attributes = map[string]string{"only.destination": fmt.Sprintf("literal-%d'雪", seed)}
		stream.Scope, stream.Version, stream.SchemaURL, stream.Name, stream.Unit = d.Source.Scope, d.Source.Version, d.Source.SchemaURL, d.Source.Metric, unit
		stream.Resource = map[string]string{"service.name": "destination-worker", "service.instance.id": fmt.Sprint(seed)}
		if ns.Mode == "value" {
			stream.Resource[identity.NamespaceKey] = ns.Value
		}
		if d.EnvironmentSelector.Scope.Mode != "unreported" {
			stream.Resource[identity.EnvironmentKey] = d.EnvironmentSelector.Scope.Value
		}
		stream.Attributes = map[string]string{"only.destination": d.Source.Attributes["only.destination"]}
		d.Source.ExactStream = &stream
		destination.FromNS = strconv.FormatInt(destination.FromMS*1000000+int64(seed), 10)
		destination.ToNS = strconv.FormatInt(destination.ToMS*1000000-int64(seed), 10)
		declaration := workbench.NativeViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeViewSelection{Kind: selection}, Kind: workbench.ViewKind{Name: kind, Version: 1}, Display: workbench.NativeViewDisplay{Category: category, Value: value, RequiresComplete: !gauge}}
		store := workbench.Store{Dir: t.TempDir()}
		source, e := store.CreateNativeView(workbench.NativeViewInput{Version: 1, Title: fmt.Sprintf("Display %d 雪", seed), Question: sourceQ, Declaration: declaration})
		if e != nil {
			tc.Fatal(e)
		}
		page, e := store.NativeViews("")
		if e != nil || len(page.Items) != 1 {
			tc.Fatal("metadata page absent", e)
		}
		wantSummary := workbench.NativeViewSummary{ID: source.ID, Title: source.Title, Family: family, Selection: selection, Kind: kind, Unit: unit}
		if !reflect.DeepEqual(page.Items[0], wantSummary) {
			tc.Fatal("metadata family/selection lost")
		}
		sourceBefore, destinationBefore := continuationJSON352(source), continuationJSON352(destination)
		reused, e := workbench.ReuseNativeView(*source, destination)
		if e != nil || reused.ID != "" || !reflect.DeepEqual(reused.Question, destination) || !reflect.DeepEqual(reused.Declaration, declaration) {
			tc.Fatal("destination or display authority changed", e)
		}
		// Calculation, including exact stream, nanos and stale policy, is untouched.
		before, e := measurements.Compile(destination)
		if e != nil {
			tc.Fatal(e)
		}
		after, e := measurements.Compile(reused.Question)
		if e != nil || !reflect.DeepEqual(before, after) {
			tc.Fatal("display changed calculation", e)
		}
		for _, mutate := range []func(*measurements.Question){
			func(q *measurements.Question) {
				if q.Definition.Kind == "gauge" {
					q.Definition.Kind = "histogram"
				} else {
					q.Definition.Kind = "gauge"
				}
			},
			func(q *measurements.Question) { q.Definition.Grain = "consumer-group" },
			func(q *measurements.Question) {
				q.Definition.Source.Unit += "different"
				q.Definition.Source.ExactStream.Unit = q.Definition.Source.Unit
			},
			func(q *measurements.Question) {
				q.Definition.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: "host-a"}
			},
		} {
			var bad measurements.Question
			encoded, _ := json.Marshal(destination)
			json.Unmarshal(encoded, &bad)
			mutate(&bad)
			if _, err := workbench.ReuseNativeView(*source, bad); err == nil {
				tc.Fatal("incompatible destination admitted")
			}
		}
		wire, _ := json.Marshal(reused)
		var decoded workbench.NativeViewInput
		if json.Unmarshal(wire, &decoded) != nil || decoded.Validate() != nil {
			tc.Fatal("closed native view round trip failed")
		}
		invalid := []json.RawMessage{}
		// Explicit false is a value; omitted/null/foreign fields may not stand in for it.
		for _, mutate := range []func(map[string]any){
			func(v map[string]any) {
				v["question"].(map[string]any)["definition"].(map[string]any)["source"].(map[string]any)["unit"] = ""
			},
			func(v map[string]any) {
				delete(v["question"].(map[string]any)["definition"].(map[string]any)["source"].(map[string]any), "unit")
			},
			func(v map[string]any) {
				delete(v["declaration"].(map[string]any)["display"].(map[string]any), "requiresComplete")
			},
			func(v map[string]any) {
				v["declaration"].(map[string]any)["display"].(map[string]any)["requiresComplete"] = nil
			},
			func(v map[string]any) {
				v["declaration"].(map[string]any)["display"].(map[string]any)["requiresComplete"] = gauge
			},
			func(v map[string]any) {
				v["declaration"].(map[string]any)["selection"].(map[string]any)["kind"] = "time-bucket"
			},
			func(v map[string]any) {
				v["declaration"].(map[string]any)["display"].(map[string]any)["donorValues"] = []any{}
			},
			func(v map[string]any) {
				v["question"].(map[string]any)["definition"].(map[string]any)["source"].(map[string]any)["attributes"] = []any{}
			},
			func(v map[string]any) {
				v["question"].(map[string]any)["definition"].(map[string]any)["source"].(map[string]any)["exactStream"].(map[string]any)["resource"] = []any{}
			},
			func(v map[string]any) {
				v["question"].(map[string]any)["definition"].(map[string]any)["environmentSelector"] = nil
			},
		} {
			var v map[string]any
			json.Unmarshal(wire, &v)
			mutate(v)
			b, _ := json.Marshal(v)
			var bad workbench.NativeViewInput
			if err := json.Unmarshal(b, &bad); err == nil && bad.Validate() == nil {
				tc.Fatalf("invalid declaration/question admitted %s", b)
			}
			invalid = append(invalid, b)
		}
		var fixture nativeFixture386
		fixtureWire, _ := json.Marshal(nativeFixture386{Page: *page, Source: *source, Destination: destination, Reused: *reused, Invalid: invalid})
		json.Unmarshal(fixtureWire, &fixture)
		fixtures = append(fixtures, fixture)
		// Neither returned nested selectors nor declaration may alias donor/destination.
		reused.Question.Definition.Source.Attributes["only.destination"] = "mutated"
		reused.Question.Definition.Source.ExactStream.Resource["service.instance.id"] = "mutated"
		reused.Question.Definition.Source.Services[0] = "mutated"
		reused.Question.Definition.EnvironmentSelector.Scope.Value = "mutated"
		reused.Declaration.Kind.Name = "mutated"
		if continuationJSON352(source) != sourceBefore || continuationJSON352(destination) != destinationBefore {
			tc.Fatal("returned display aliases an input")
		}
	}, options(100)...)
	encoded, _ := json.Marshal(fixtures)
	cmd := exec.CommandContext(t.Context(), "node", "native_view_386_fixture.mjs")
	cmd.Stdin = bytes.NewReader(encoded)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("cross-language native display admission: %v\n%s", e, out)
	}
	var result struct{ Examples, Refusals int }
	if json.Unmarshal(out, &result) != nil || result.Examples != 100 || result.Refusals != 1800 {
		t.Fatalf("incomplete native view fixture %s", out)
	}
}

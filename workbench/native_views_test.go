package workbench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func nativeViewInput371(t *testing.T) NativeViewInput {
	t.Helper()
	raw, e := os.ReadFile("../config/measurements/checkout-duration.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	return NativeViewInput{Version: 1, Title: "Interval observations", Question: measurements.Question{Definition: d, FromMS: 1000, ToMS: 61000}, Declaration: NativeViewDeclaration{Version: "goo.view/1", Scope: NativeViewScope{"measurement-question", "metrics"}, Selection: NativeViewSelection{"export-interval"}, Kind: ViewKind{"histogram", 1}, Display: NativeViewDisplay{"bucket", "observation-count", true}}}
}
func TestNativeView371SaveReuseIdentity(t *testing.T) {
	in := nativeViewInput371(t)
	store := Store{Dir: t.TempDir()}
	before, e := measurements.Compile(in.Question)
	if e != nil {
		t.Fatal(e)
	}
	saved, e := store.CreateNativeView(in)
	if e != nil {
		t.Fatal(e)
	}
	same, e := store.CreateNativeView(in)
	if e != nil || same.ID != saved.ID {
		t.Fatal("immutable creation unstable", e)
	}
	in.Declaration.Kind.Name = "table"
	table, e := store.CreateNativeView(in)
	if e != nil || table.ID == saved.ID {
		t.Fatal("display absent from view identity", e)
	}
	after, e := measurements.Compile(in.Question)
	if e != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("display changed calculation identity", e)
	}
	dest := in.Question
	dest.FromMS = 70000
	dest.ToMS = 130000
	dest.FromNS = "70000000001"
	dest.ToNS = "129999999999"
	dest.Definition.Environment = ""
	dest.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}}
	dest.Definition.Source.Services = []string{"target-api"}
	dest.Definition.Source.Namespace = &identity.Namespace{Mode: "value", Value: ""}
	dest.Definition.Source.Attributes = map[string]string{}
	dest.Definition.Source.ExactStream = &metrics.Stream{Resource: map[string]string{"service.name": "target-api", "service.instance.id": "instance", "service.namespace": "", "deployment.environment.name": ""}, Scope: dest.Definition.Source.Scope, Version: dest.Definition.Source.Version, SchemaURL: dest.Definition.Source.SchemaURL, Name: dest.Definition.Source.Metric, Unit: dest.Definition.Source.Unit, Attributes: map[string]string{}}
	copied, e := ReuseNativeView(*saved, dest)
	if e != nil {
		t.Fatal(e)
	}
	if copied.ID != "" || !reflect.DeepEqual(copied.Question, dest) || copied.Declaration != saved.Declaration {
		t.Fatal("reuse imported donor scope or changed destination")
	}
	copied.Question.Definition.Source.Services[0] = "mutated"
	copied.Question.Definition.Source.ExactStream.Resource["service.name"] = "mutated"
	copied.Question.Definition.EnvironmentSelector.Scope.Mode = "unreported"
	if dest.Definition.Source.Services[0] != "target-api" || dest.Definition.Source.ExactStream.Resource["service.name"] != "target-api" || dest.Definition.EnvironmentSelector.Scope.Mode != "empty" {
		t.Fatal("destination not detached")
	}
	raw, e := os.ReadFile(filepath.Join(store.Dir, "native-histogram-views", saved.ID+".json"))
	if e != nil {
		t.Fatal(e)
	}
	raw = []byte(strings.Replace(string(raw), "Interval observations", "Forged observations", 1))
	if e = os.WriteFile(filepath.Join(store.Dir, "native-histogram-views", saved.ID+".json"), raw, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = store.LoadNativeView(saved.ID); e == nil {
		t.Fatal("corrupt winner accepted")
	}
}
func TestNativeView371Refusals(t *testing.T) {
	base := nativeViewInput371(t)
	for name, mutate := range map[string]func(*NativeViewInput){"kind": func(v *NativeViewInput) { v.Question.Definition.Kind = "gauge" }, "grain": func(v *NativeViewInput) { v.Question.Definition.Grain = "consumer-group" }, "resource": func(v *NativeViewInput) {
		v.Question.Definition.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: "host"}
	}, "bucket": func(v *NativeViewInput) { n := int64(1000); v.Question.Definition.HistogramBucketMS = &n }, "unit": func(v *NativeViewInput) { v.Question.Definition.Source.Unit = "By" }, "selection": func(v *NativeViewInput) { v.Declaration.Selection.Kind = "time-bucket" }, "complete": func(v *NativeViewInput) { v.Declaration.Display.RequiresComplete = false }, "measure": func(v *NativeViewInput) { v.Declaration.Display.Value = "count" }} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(base)
			var in NativeViewInput
			if e := json.Unmarshal(raw, &in); e != nil {
				t.Fatal(e)
			}
			mutate(&in)
			if _, e := (Store{Dir: t.TempDir()}).CreateNativeView(in); e == nil {
				t.Fatal("invalid view accepted")
			}
		})
	}
	saved, e := (Store{Dir: t.TempDir()}).CreateNativeView(base)
	if e != nil {
		t.Fatal(e)
	}
	dest := base.Question
	dest.Definition.Source.Unit = "ms"
	if _, e = ReuseNativeView(*saved, dest); e == nil {
		t.Fatal("unit conversion silently admitted")
	}
	raw, _ := json.Marshal(base)
	for _, bad := range []string{strings.Replace(string(raw), `"fromMs":1000`, `"fromNS":"","toNS":"","fromMs":1000`, 1), strings.Replace(string(raw), `"source":{`, `"source":{"resource":null,`, 1), strings.TrimSuffix(string(raw), "}") + `,"id":""}`, strings.Replace(string(raw), `"fromMs":1000`, `"fromMs":null`, 1), strings.Replace(string(raw), `"title":"Interval observations"`, `"title":"\ud800"`, 1), strings.Replace(string(raw), `"kind":"histogram"`, `"histogramBucketMs":null,"kind":"histogram"`, 1)} {
		var in NativeViewInput
		if e = json.Unmarshal([]byte(bad), &in); e == nil {
			t.Fatal("malformed wire admitted", bad)
		}
	}
	requestRaw, _ := json.Marshal(NativeViewReuseRequest{ID: saved.ID, Question: base.Question})
	requestRaw = []byte(strings.Replace(string(requestRaw), `"source":{`, `"source":{"resource":null,`, 1))
	emptyRaw, _ := json.Marshal(NativeViewReuseRequest{ID: saved.ID, Question: base.Question})
	emptyRaw = []byte(strings.Replace(string(emptyRaw), `"fromMs":1000`, `"fromNS":"","toNS":"","fromMs":1000`, 1))
	var emptyReuse NativeViewReuseRequest
	if err := json.Unmarshal(emptyRaw, &emptyReuse); err == nil {
		t.Fatal("reuse explicit empty native bounds admitted")
	}
	var reuse NativeViewReuseRequest
	if err := json.Unmarshal(requestRaw, &reuse); err == nil {
		t.Fatal("reuse source resource:null admitted")
	}

}

func gaugeViewInput386(t *testing.T, unit string) NativeViewInput {
	t.Helper()
	in := nativeViewInput371(t)
	in.Question.Definition.Kind = "gauge"
	in.Question.Definition.Source.Unit = unit
	in.Declaration.Selection.Kind = "retained-observations"
	in.Declaration.Kind.Name = "timeline"
	in.Declaration.Display = NativeViewDisplay{"time", "recorded-value", false}
	return in
}

func TestNativeView386GaugeRoundTripAndReuse(t *testing.T) {
	for _, unit := range []string{"1", "By", "custom雪"} {
		t.Run(unit, func(t *testing.T) {
			in := gaugeViewInput386(t, unit)
			store := Store{Dir: t.TempDir()}
			saved, err := store.CreateNativeView(in)
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := store.LoadNativeView(saved.ID)
			if err != nil || !reflect.DeepEqual(saved, loaded) {
				t.Fatal("gauge roundtrip", err)
			}
			page, err := store.NativeViews("")
			if err != nil || len(page.Items) != 1 || page.Items[0].Family != "gauge" || page.Items[0].Selection != "retained-observations" || page.Items[0].Unit != unit {
				t.Fatal("gauge catalog", page, err)
			}
			destination := in.Question
			destination.FromMS, destination.ToMS = 70000, 130000
			destination.Definition.Source.Services = []string{"destination"}
			reused, err := ReuseNativeView(*saved, destination)
			if err != nil || reused.ID != "" || !reflect.DeepEqual(reused.Question, destination) {
				t.Fatal("destination authority", err)
			}
			reused.Question.Definition.Source.Services[0] = "changed"
			if destination.Definition.Source.Services[0] != "destination" {
				t.Fatal("destination aliased")
			}
			destination.Definition.Source.Unit = "another-unit"
			if _, err := ReuseNativeView(*saved, destination); err == nil {
				t.Fatal("different gauge unit accepted")
			}
			destination.Definition.Source.Unit = unit
			destination.Definition.Grain = "service"
			if _, err := ReuseNativeView(*saved, destination); err == nil {
				t.Fatal("different gauge grain accepted")
			}
			destination.Definition.Grain = "per-stream"
			destination.Definition.Kind = "histogram"
			destination.Definition.Source.Unit = "s"
			if _, err := ReuseNativeView(*saved, destination); err == nil {
				t.Fatal("cross-family reuse accepted")
			}
			in.Declaration.Kind.Name = "table"
			if _, err := store.CreateNativeView(in); err != nil {
				t.Fatal("gauge table", err)
			}
		})
	}
}

func TestNativeView386ClosedGaugeWire(t *testing.T) {
	raw, err := json.Marshal(gaugeViewInput386(t, "1"))
	if err != nil {
		t.Fatal(err)
	}
	var accepted NativeViewInput
	if err = json.Unmarshal(raw, &accepted); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(string) string{
		func(s string) string { return strings.Replace(s, `,"requiresComplete":false`, "", 1) },
		func(s string) string {
			return strings.Replace(s, `"requiresComplete":false`, `"requiresComplete":null`, 1)
		},
		func(s string) string {
			return strings.Replace(s, `"requiresComplete":false`, `"RequiresComplete":false`, 1)
		},
		func(s string) string { return strings.Replace(s, `"source":{`, `"lifecycle":null,"source":{`, 1) },
		func(s string) string { return strings.Replace(s, `"source":{`, `"counterDomain":"","source":{`, 1) },
		func(s string) string { return strings.Replace(s, `"source":{`, `"source":{"resource":null,`, 1) },
		func(s string) string { return strings.Replace(s, `"source":{`, `"source":{"Unit":"1",`, 1) },
		func(s string) string { return strings.Replace(s, `"version":1`, `"Version":1`, 1) },
	} {
		if err := json.Unmarshal([]byte(mutate(string(raw))), new(NativeViewInput)); err == nil {
			t.Fatal("invalid gauge wire admitted", mutate(string(raw)))
		}
	}
}

func TestNativeView386RejectsUnitlessGauge(t *testing.T) {
	in := gaugeViewInput386(t, "")
	if err := in.Validate(); err == nil {
		t.Fatal("compiler-rejected empty gauge unit admitted")
	}
}

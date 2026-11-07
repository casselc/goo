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

func nativeViewInput371(t *testing.T) NativeHistogramViewInput {
	t.Helper()
	raw, e := os.ReadFile("../config/measurements/checkout-duration.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	return NativeHistogramViewInput{Version: 1, Title: "Interval observations", Question: measurements.Question{Definition: d, FromMS: 1000, ToMS: 61000}, Declaration: NativeHistogramViewDeclaration{Version: "goo.view/1", Scope: NativeHistogramViewScope{"measurement-question", "metrics"}, Selection: NativeHistogramViewSelection{"export-interval"}, Kind: ViewKind{"histogram", 1}, Display: NativeHistogramViewDisplay{"bucket", "observation-count", true}}}
}
func TestNativeHistogramView371SaveReuseIdentity(t *testing.T) {
	in := nativeViewInput371(t)
	store := Store{Dir: t.TempDir()}
	before, e := measurements.Compile(in.Question)
	if e != nil {
		t.Fatal(e)
	}
	saved, e := store.CreateNativeHistogramView(in)
	if e != nil {
		t.Fatal(e)
	}
	same, e := store.CreateNativeHistogramView(in)
	if e != nil || same.ID != saved.ID {
		t.Fatal("immutable creation unstable", e)
	}
	in.Declaration.Kind.Name = "table"
	table, e := store.CreateNativeHistogramView(in)
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
	copied, e := ReuseNativeHistogramView(*saved, dest)
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
	if _, e = store.LoadNativeHistogramView(saved.ID); e == nil {
		t.Fatal("corrupt winner accepted")
	}
}
func TestNativeHistogramView371Refusals(t *testing.T) {
	base := nativeViewInput371(t)
	for name, mutate := range map[string]func(*NativeHistogramViewInput){"kind": func(v *NativeHistogramViewInput) { v.Question.Definition.Kind = "gauge" }, "grain": func(v *NativeHistogramViewInput) { v.Question.Definition.Grain = "consumer-group" }, "resource": func(v *NativeHistogramViewInput) {
		v.Question.Definition.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: "host"}
	}, "bucket": func(v *NativeHistogramViewInput) { n := int64(1000); v.Question.Definition.HistogramBucketMS = &n }, "unit": func(v *NativeHistogramViewInput) { v.Question.Definition.Source.Unit = "By" }, "selection": func(v *NativeHistogramViewInput) { v.Declaration.Selection.Kind = "time-bucket" }, "complete": func(v *NativeHistogramViewInput) { v.Declaration.Display.RequiresComplete = false }, "measure": func(v *NativeHistogramViewInput) { v.Declaration.Display.Value = "count" }} {
		t.Run(name, func(t *testing.T) {
			raw, _ := json.Marshal(base)
			var in NativeHistogramViewInput
			if e := json.Unmarshal(raw, &in); e != nil {
				t.Fatal(e)
			}
			mutate(&in)
			if _, e := (Store{Dir: t.TempDir()}).CreateNativeHistogramView(in); e == nil {
				t.Fatal("invalid view accepted")
			}
		})
	}
	saved, e := (Store{Dir: t.TempDir()}).CreateNativeHistogramView(base)
	if e != nil {
		t.Fatal(e)
	}
	dest := base.Question
	dest.Definition.Source.Unit = "ms"
	if _, e = ReuseNativeHistogramView(*saved, dest); e == nil {
		t.Fatal("unit conversion silently admitted")
	}
	raw, _ := json.Marshal(base)
	for _, bad := range []string{strings.Replace(string(raw), `"fromMs":1000`, `"fromNS":"","toNS":"","fromMs":1000`, 1), strings.Replace(string(raw), `"source":{`, `"source":{"resource":null,`, 1), strings.TrimSuffix(string(raw), "}") + `,"id":""}`, strings.Replace(string(raw), `"fromMs":1000`, `"fromMs":null`, 1), strings.Replace(string(raw), `"title":"Interval observations"`, `"title":"\ud800"`, 1), strings.Replace(string(raw), `"kind":"histogram"`, `"histogramBucketMs":null,"kind":"histogram"`, 1)} {
		var in NativeHistogramViewInput
		if e = json.Unmarshal([]byte(bad), &in); e == nil {
			t.Fatal("malformed wire admitted", bad)
		}
	}
	requestRaw, _ := json.Marshal(NativeHistogramViewReuseRequest{ID: saved.ID, Question: base.Question})
	requestRaw = []byte(strings.Replace(string(requestRaw), `"source":{`, `"source":{"resource":null,`, 1))
	emptyRaw, _ := json.Marshal(NativeHistogramViewReuseRequest{ID: saved.ID, Question: base.Question})
	emptyRaw = []byte(strings.Replace(string(emptyRaw), `"fromMs":1000`, `"fromNS":"","toNS":"","fromMs":1000`, 1))
	var emptyReuse NativeHistogramViewReuseRequest
	if err := json.Unmarshal(emptyRaw, &emptyReuse); err == nil {
		t.Fatal("reuse explicit empty native bounds admitted")
	}
	var reuse NativeHistogramViewReuseRequest
	if err := json.Unmarshal(requestRaw, &reuse); err == nil {
		t.Fatal("reuse source resource:null admitted")
	}

}

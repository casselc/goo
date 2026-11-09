package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
)

func TestNativeViewsHTTP371(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	mux := http.NewServeMux()
	registerNativeViews(mux, t.TempDir())
	var cases, negatives []planningWireCase
	call := func(name, method, path, pattern, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%s: got%d want%d: %s", name, w.Code, want, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("cacheable metadata", name)
		}
		if want < 400 {
			cases = append(cases, planningWireCase{Name: name, Method: method, Path: pattern, Body: body, Status: w.Code, Headers: map[string]string{"Content-Type": w.Header().Get("Content-Type"), "Cache-Control": w.Header().Get("Cache-Control")}, Response: w.Body.String()})
		}
		if want >= 400 {
			negatives = append(negatives, planningWireCase{Name: name, Method: method, Path: path, Body: body, Status: w.Code, Response: w.Body.String()})
		}
		return w
	}
	encode := func(v any) string {
		t.Helper()
		raw, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return string(raw)
	}
	raw, e := os.ReadFile("../../config/measurements/checkout-duration.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	in := workbench.NativeViewInput{Version: 1, Title: "Native interval view", Question: measurements.Question{Definition: d, FromMS: 1000, ToMS: 61000}, Declaration: workbench.NativeViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeViewSelection{Kind: "export-interval"}, Kind: workbench.ViewKind{Name: "histogram", Version: 1}, Display: workbench.NativeViewDisplay{Category: "bucket", Value: "observation-count", RequiresComplete: true}}}
	w := call("create", "POST", "/native-views", "/native-views", encode(in), 201)
	var saved workbench.NativeView
	if e = json.Unmarshal(w.Body.Bytes(), &saved); e != nil {
		t.Fatal(e)
	}
	call("load", "GET", "/native-views/"+saved.ID, "/native-views/{id}", "", 200)
	in.Declaration.Kind.Name = "table"
	w = call("create-table", "POST", "/native-views", "/native-views", encode(in), 201)
	var table workbench.NativeView
	json.Unmarshal(w.Body.Bytes(), &table)
	if table.ID == saved.ID {
		t.Fatal("view display absent from artifact ID")
	}
	q := in.Question
	q.FromMS = 70000
	q.ToMS = 130000
	q.Definition.Environment = ""
	q.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}}
	q.Definition.Source.Services = []string{"target-api"}
	q.Definition.Source.Namespace = &identity.Namespace{Mode: "missing"}
	reuse := workbench.NativeViewReuseRequest{ID: saved.ID, Question: q}
	w = call("reuse", "POST", "/native-views/reuse", "/native-views/reuse", encode(reuse), 200)
	var output workbench.NativeView
	if e = json.Unmarshal(w.Body.Bytes(), &output); e != nil || output.ID != "" || !reflect.DeepEqual(output.Question, q) || output.Declaration != saved.Declaration {
		t.Fatal("reuse changed destination or imported stored identity", e)
	}
	reuse.Question.Definition.Source.Unit = "ms"
	call("unit-refusal", "POST", "/native-views/reuse", "", encode(reuse), 400)
	body := encode(in)
	for name, bad := range map[string]string{"empty-native-bounds": strings.Replace(body, `"fromMs":1000`, `"fromNS":"","toNS":"","fromMs":1000`, 1), "null-resource": strings.Replace(body, `"source":{`, `"source":{"resource":null,`, 1), "id": strings.TrimSuffix(body, "}") + `,"id":""}`, "unknown": strings.TrimSuffix(body, "}") + `,"observations":[]}`, "null-bounds": strings.Replace(body, `"fromMs":1000`, `"fromMs":null`, 1), "unicode": strings.Replace(body, `"title":"Native interval view"`, `"title":"\ud800"`, 1), "bucket": strings.Replace(body, `"kind":"histogram"`, `"histogramBucketMs":null,"kind":"histogram"`, 1), "trailing": body + ` {}`, "incomplete": strings.Replace(body, `"requiresComplete":true`, `"requiresComplete":false`, 1)} {
		call(name, "POST", "/native-views", "", bad, 400)
	}
	reuse.Question = q
	call("reuse-empty-native-bounds", "POST", "/native-views/reuse", "", strings.Replace(encode(reuse), `"fromMs":70000`, `"fromNS":"","toNS":"","fromMs":70000`, 1), 400)
	call("reuse-null-resource", "POST", "/native-views/reuse", "", strings.Replace(encode(reuse), `"source":{`, `"source":{"resource":null,`, 1), 400)
	call("foreign-source", "POST", "/native-views/reuse", "", `{"id":"`+saved.ID+`","question":null}`, 400)
	call("query-refusal", "GET", "/native-views/"+saved.ID+"?fresh=true", "", "", 400)
	call("bad-id", "GET", "/native-views/bad", "", "", 400)
	call("missing", "GET", "/native-views/"+strings.Repeat("0", 64), "", "", 404)
	call("oversize", "POST", "/native-views", "", `{"padding":"`+strings.Repeat("x", workbench.MaxBytes)+`"}`, 413)

	gauge := in
	gauge.Title = "Native gauge view"
	gauge.Question.Definition.Kind = "gauge"
	gauge.Question.Definition.Source.Unit = "By"
	gauge.Declaration.Selection.Kind = "retained-observations"
	gauge.Declaration.Kind.Name = "timeline"
	gauge.Declaration.Display = workbench.NativeViewDisplay{Category: "time", Value: "recorded-value", RequiresComplete: false}
	w = call("create-gauge", "POST", "/native-views", "/native-views", encode(gauge), 201)
	var savedGauge workbench.NativeView
	if err := json.Unmarshal(w.Body.Bytes(), &savedGauge); err != nil {
		t.Fatal(err)
	}
	call("load-gauge", "GET", "/native-views/"+savedGauge.ID, "/native-views/{id}", "", 200)
	gaugeDestination := gauge.Question
	gaugeDestination.FromMS, gaugeDestination.ToMS = 70000, 130000
	gaugeDestination.Definition.Source.Services = []string{"gauge-target"}
	gaugeReuse := workbench.NativeViewReuseRequest{ID: savedGauge.ID, Question: gaugeDestination}
	w = call("reuse-gauge", "POST", "/native-views/reuse", "/native-views/reuse", encode(gaugeReuse), 200)
	if err := json.Unmarshal(w.Body.Bytes(), &output); err != nil || !reflect.DeepEqual(output.Question, gaugeDestination) || output.ID != "" {
		t.Fatal("gauge destination changed", err)
	}
	gauge.Declaration.Kind.Name = "table"
	call("create-gauge-table", "POST", "/native-views", "/native-views", encode(gauge), 201)
	call("gauge-missing-completeness", "POST", "/native-views", "", strings.Replace(encode(gauge), `,"requiresComplete":false`, "", 1), 400)
	gauge.Question.Definition.Source.Unit = ""
	call("gauge-unitless", "POST", "/native-views", "", encode(gauge), 400)
	gaugeReuse.ID = saved.ID
	call("cross-family-reuse", "POST", "/native-views/reuse", "", encode(gaugeReuse), 400)
	for _, domain := range []string{"", "nonnegative-decimal"} {
		counter := in
		counter.Title = "Native counter " + domain
		counter.Question.Definition.Kind = "counter"
		counter.Question.Definition.CounterDomain = domain
		counter.Question.Definition.Source.Unit = "{request}"
		counter.Declaration.Selection.Kind = "evaluated-intervals"
		counter.Declaration.Kind.Name = "timeline"
		counter.Declaration.Display = workbench.NativeViewDisplay{Category: "time", Value: "counter-rate", RequiresComplete: false}
		w = call("create-counter-"+domain, "POST", "/native-views", "/native-views", encode(counter), 201)
		var savedCounter workbench.NativeView
		if err := json.Unmarshal(w.Body.Bytes(), &savedCounter); err != nil {
			t.Fatal(err)
		}
		call("load-counter-"+domain, "GET", "/native-views/"+savedCounter.ID, "/native-views/{id}", "", 200)
		destination := counter.Question
		destination.FromMS, destination.ToMS = 70000, 130000
		destination.Definition.MaxGapMS = 750
		destination.Definition.Source.Services = []string{"counter-destination"}
		reuse := workbench.NativeViewReuseRequest{ID: savedCounter.ID, Question: destination}
		w = call("reuse-counter-"+domain, "POST", "/native-views/reuse", "/native-views/reuse", encode(reuse), 200)
		if err := json.Unmarshal(w.Body.Bytes(), &output); err != nil || !reflect.DeepEqual(output.Question, destination) || output.ID != "" {
			t.Fatal("counter destination changed", err)
		}
		counter.Declaration.Kind.Name = "table"
		call("create-counter-table-"+domain, "POST", "/native-views", "/native-views", encode(counter), 201)
		reuse.Question.Definition.CounterDomain = "nonnegative-decimal"
		if domain != "" {
			reuse.Question.Definition.CounterDomain = ""
		}
		call("counter-domain-refusal-"+domain, "POST", "/native-views/reuse", "", encode(reuse), 400)
		call("counter-null-domain-"+domain, "POST", "/native-views", "", strings.Replace(encode(counter), `"source":{`, `"counterDomain":null,"source":{`, 1), 400)
		counter.Declaration.Selection.Kind = "retained-observations"
		call("counter-gauge-selection-"+domain, "POST", "/native-views", "", encode(counter), 400)
	}
	w = call("catalog-counter-domains", "GET", "/native-views", "/native-views", "", 200)
	var catalog workbench.NativeViewList
	if err := json.Unmarshal(w.Body.Bytes(), &catalog); err != nil {
		t.Fatal(err)
	}
	counterDomains := map[string]int{}
	for _, item := range catalog.Items {
		if item.Family == "counter" {
			counterDomains[item.CounterDomain]++
		} else if item.CounterDomain != "" {
			t.Fatal("counter domain leaked to another family")
		}
	}
	if counterDomains["nonnegative-integer"] != 2 || counterDomains["nonnegative-decimal"] != 2 {
		t.Fatal("counter catalog policy missing", counterDomains)
	}
	if reads != 0 {
		t.Fatal("metadata accessed telemetry", reads)
	}
	if file := os.Getenv("GOO_NATIVE_HISTOGRAM_VIEW_OBSERVATIONS"); file != "" {
		raw, e := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(file, raw, 0600); e != nil {
			t.Fatal(e)
		}
	}
	if file := os.Getenv("GOO_NATIVE_HISTOGRAM_VIEW_NEGATIVES"); file != "" {
		raw, e := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: negatives}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(file, raw, 0600); e != nil {
			t.Fatal(e)
		}
	}

}

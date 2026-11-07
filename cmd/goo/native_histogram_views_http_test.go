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

func TestNativeHistogramViewsHTTP371(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	mux := http.NewServeMux()
	registerNativeHistogramViews(mux, t.TempDir())
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
	in := workbench.NativeHistogramViewInput{Version: 1, Title: "Native interval view", Question: measurements.Question{Definition: d, FromMS: 1000, ToMS: 61000}, Declaration: workbench.NativeHistogramViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeHistogramViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeHistogramViewSelection{Kind: "export-interval"}, Kind: workbench.ViewKind{Name: "histogram", Version: 1}, Display: workbench.NativeHistogramViewDisplay{Category: "bucket", Value: "observation-count", RequiresComplete: true}}}
	w := call("create", "POST", "/native-histogram-views", "/native-histogram-views", encode(in), 201)
	var saved workbench.NativeHistogramView
	if e = json.Unmarshal(w.Body.Bytes(), &saved); e != nil {
		t.Fatal(e)
	}
	call("load", "GET", "/native-histogram-views/"+saved.ID, "/native-histogram-views/{id}", "", 200)
	in.Declaration.Kind.Name = "table"
	w = call("create-table", "POST", "/native-histogram-views", "/native-histogram-views", encode(in), 201)
	var table workbench.NativeHistogramView
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
	reuse := workbench.NativeHistogramViewReuseRequest{ID: saved.ID, Question: q}
	w = call("reuse", "POST", "/native-histogram-views/reuse", "/native-histogram-views/reuse", encode(reuse), 200)
	var output workbench.NativeHistogramView
	if e = json.Unmarshal(w.Body.Bytes(), &output); e != nil || output.ID != "" || !reflect.DeepEqual(output.Question, q) || output.Declaration != saved.Declaration {
		t.Fatal("reuse changed destination or imported stored identity", e)
	}
	reuse.Question.Definition.Source.Unit = "ms"
	call("unit-refusal", "POST", "/native-histogram-views/reuse", "", encode(reuse), 400)
	body := encode(in)
	for name, bad := range map[string]string{"empty-native-bounds": strings.Replace(body, `"fromMs":1000`, `"fromNS":"","toNS":"","fromMs":1000`, 1), "null-resource": strings.Replace(body, `"source":{`, `"source":{"resource":null,`, 1), "id": strings.TrimSuffix(body, "}") + `,"id":""}`, "unknown": strings.TrimSuffix(body, "}") + `,"observations":[]}`, "null-bounds": strings.Replace(body, `"fromMs":1000`, `"fromMs":null`, 1), "unicode": strings.Replace(body, `"title":"Native interval view"`, `"title":"\ud800"`, 1), "bucket": strings.Replace(body, `"kind":"histogram"`, `"histogramBucketMs":null,"kind":"histogram"`, 1), "trailing": body + ` {}`, "incomplete": strings.Replace(body, `"requiresComplete":true`, `"requiresComplete":false`, 1)} {
		call(name, "POST", "/native-histogram-views", "", bad, 400)
	}
	reuse.Question = q
	call("reuse-empty-native-bounds", "POST", "/native-histogram-views/reuse", "", strings.Replace(encode(reuse), `"fromMs":70000`, `"fromNS":"","toNS":"","fromMs":70000`, 1), 400)
	call("reuse-null-resource", "POST", "/native-histogram-views/reuse", "", strings.Replace(encode(reuse), `"source":{`, `"source":{"resource":null,`, 1), 400)
	call("foreign-source", "POST", "/native-histogram-views/reuse", "", `{"id":"`+saved.ID+`","question":null}`, 400)
	call("query-refusal", "GET", "/native-histogram-views/"+saved.ID+"?fresh=true", "", "", 400)
	call("bad-id", "GET", "/native-histogram-views/bad", "", "", 400)
	call("missing", "GET", "/native-histogram-views/"+strings.Repeat("0", 64), "", "", 404)
	call("oversize", "POST", "/native-histogram-views", "", `{"padding":"`+strings.Repeat("x", workbench.MaxBytes)+`"}`, 413)
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

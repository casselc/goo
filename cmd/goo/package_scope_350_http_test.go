package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
)

func TestTypedPackageMetadataHTTP350(t *testing.T) {
	mux := http.NewServeMux()
	registerTelemetryPackages(mux, t.TempDir())
	call := func(method, path string, body any) *httptest.ResponseRecorder {
		data, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(data)))
		return w
	}
	for _, mode := range []string{"current", "effective", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-'雪"}}
			switch mode {
			case "current":
				selector.Policy = "current-only"
			case "empty":
				selector.Scope = identity.EnvironmentScope{Mode: "empty"}
			case "unreported":
				selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
			q := exploration.Query{Signal: "spans", EnvironmentSelector: selector, FromMS: 1000, ToMS: 61000, Offset: 50, Limit: 10, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Summary: "span-duration", SpanStatus: true, BucketMS: 10000, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: ""}}
			w := call("POST", "/telemetry-packages/draft", map[string]any{"question": q, "name": "typed-package", "title": "Typed package", "lookbackMinutes": 5})
			var plan exploration.PackagePlan
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil {
				t.Fatal(w.Code, w.Body.String())
			}
			expected := q
			expected.Offset = 0
			expected.Limit = 100
			if !reflect.DeepEqual(plan.Question, expected) || plan.Definition.Source.Environment != "" || !reflect.DeepEqual(plan.Definition.Source.EnvironmentSelector, selector) {
				t.Fatal("draft weakened full scope", w.Body.String())
			}
			w = call("POST", "/telemetry-packages/plan", map[string]any{"definition": plan.Definition, "fromMs": 1000, "toMs": 61000})
			if w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			w = call("POST", "/telemetry-packages", map[string]string{"yaml": plan.YAML})
			var saved exploration.SavedPackage
			if w.Code != 201 || json.Unmarshal(w.Body.Bytes(), &saved) != nil {
				t.Fatal(w.Code, w.Body.String())
			}
			w = call("GET", "/telemetry-packages/"+saved.Revision, nil)
			if w.Code != 200 {
				t.Fatal(w.Body.String())
			}
			w = call("POST", "/telemetry-packages/plan", map[string]any{"yaml": plan.YAML, "fromMs": 100000, "toMs": 160000})
			var later exploration.PackagePlan
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &later) != nil || later.Revision != plan.Revision || later.Question.FromMS != 100000 || !reflect.DeepEqual(later.Question.EnvironmentSelector, selector) {
				t.Fatal("new window changed definition/scope", w.Code, w.Body.String())
			}
			w = call("POST", "/telemetry-packages/plan", map[string]any{"yaml": plan.YAML, "fromMs": 1000, "toMs": 61000, "annotation": exploration.AnnotationTarget{Name: "operation.run", Role: "server"}})
			status := 400
			if mode == "current" {
				status = 200
			}
			if w.Code != status {
				t.Fatal("annotation scope admission changed", mode, w.Code, w.Body.String())
			}
			definition, _ := json.Marshal(plan.Definition)
			for _, bad := range []string{
				strings.Replace(string(definition), `"environmentSelector":`, `"environment":"","environmentSelector":`, 1),
				strings.Replace(string(definition), `"policy":"`+selector.Policy+`"`, `"policy":null`, 1),
				strings.Replace(string(definition), `"version":"goo/environment-selector/v1"`, `"version":"goo/environment-selector/v1","unknown":true`, 1),
			} {
				body := `{"definition":` + bad + `,"fromMs":1000,"toMs":61000}`
				badResponse := httptest.NewRecorder()
				mux.ServeHTTP(badResponse, httptest.NewRequest("POST", "/telemetry-packages/plan", strings.NewReader(body)))
				if badResponse.Code != 400 {
					t.Fatal("invalid definition admitted", body, badResponse.Code, badResponse.Body.String())
				}
			}
		})
	}
}

func TestPackageDraftAndDefinitionRawUnicodeHTTP350(t *testing.T) {
	mux := http.NewServeMux()
	registerTelemetryPackages(mux, t.TempDir())
	valid := `{"question":{"signal":"spans","environmentSelector":{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"named","value":"prod"}},"fromMs":1000,"toMs":61000,"limit":100,"offset":0,"where":{"field":{"kind":"attribute","key":"tenant"},"op":"eq","value":"�"}},"name":"unicode","title":"Unicode","lookbackMinutes":1}`
	call := func(path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(body)))
		return w
	}
	response := call("/telemetry-packages/draft", valid)
	if response.Code != 200 {
		t.Fatal(response.Code, response.Body.String())
	}
	var plan exploration.PackagePlan
	if err := json.Unmarshal(response.Body.Bytes(), &plan); err != nil {
		t.Fatal(err)
	}
	// A malformed escape in the outer JSON string carrying YAML must not be
	// repaired before the YAML parser can apply the same definition contract.
	yamlBody, _ := json.Marshal(map[string]any{"yaml": plan.YAML, "fromMs": 1000, "toMs": 61000})
	badYAML := strings.Replace(string(yamlBody), "�", `\ud800`, 1)
	if badYAML == string(yamlBody) {
		t.Fatal("YAML outer-string mutation did not apply")
	}
	if w := call("/telemetry-packages/plan", badYAML); w.Code != 400 {
		t.Fatal("outer YAML string repaired before plan", w.Code, w.Body.String())
	}
	saveBody, _ := json.Marshal(map[string]string{"yaml": plan.YAML})
	badSave := strings.Replace(string(saveBody), "�", `\ud800`, 1)
	if w := call("/telemetry-packages", badSave); w.Code != 400 {
		t.Fatal("outer YAML string repaired before save", w.Code, w.Body.String())
	}
	badDraft := strings.Replace(valid, `"value":"�"`, `"value":"\ud800"`, 1)
	if w := call("/telemetry-packages/draft", badDraft); w.Code != 400 {
		t.Fatal("draft repaired original predicate", w.Code, w.Body.String())
	}
	definition, _ := json.Marshal(plan.Definition)
	badDefinition := strings.Replace(string(definition), `"value":"�"`, `"value":"\ud800"`, 1)
	if w := call("/telemetry-packages/plan", `{"definition":`+badDefinition+`,"fromMs":1000,"toMs":61000}`); w.Code != 400 {
		t.Fatal("nested HTTP definition repaired original predicate", w.Code, w.Body.String())
	}
}

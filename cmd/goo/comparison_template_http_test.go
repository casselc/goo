package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func TestComparisonTemplateHTTPBindsOnlyExplicitReferenceWithoutEvidenceReads(t *testing.T) {
	reads := 0
	previous := http.DefaultTransport
	http.DefaultTransport = templateNoSourceTransport{&reads}
	t.Cleanup(func() { http.DefaultTransport = previous })
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, dir, retentionConnection{URL: "http://source.invalid"}, &exploration.ResultSnapshots{}, nil)
	plan, err := exploration.PlanServiceDashboard(exploration.DashboardRequest{Profile: "http-server", HTTPServerObserved: true, Scope: exploration.InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000}, Service: "api", Signals: []string{"spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	var source workbench.ExplorationView
	for _, w := range plan.Widgets {
		if w.Key == "dimension" {
			out, e := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Template: "otel.http-server-routes", Version: 1, Widget: w})
			if e != nil {
				t.Fatal(e)
			}
			source = out.View
		}
	}
	plainRaw, _ := json.Marshal(source)
	// Comparison and facet controls are deliberately exclusive. This test
	// authors the comparison alternative after explicitly removing the default.
	children := make([]workbench.ViewCompositionChild, 0, len(source.Composition.Children))
	for _, child := range source.Composition.Children {
		if child.Facet == nil {
			children = append(children, child)
		}
	}
	source.Composition.Children = children
	source.Queries[0].SpanStatus = false
	source.Queries[0].BucketMS = 10000
	source.Template.Baseline = "destination"
	original := strings.Repeat("a", 64)
	chosen := strings.Repeat("b", 64)
	source.Composition.Children = append(source.Composition.Children, workbench.ViewCompositionChild{Comparison: &workbench.CountComparisonViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.CountComparisonViewKind{Name: "captured-count-comparison", Version: 1}, BaselineCapture: original, Display: workbench.CountComparisonViewDisplay{Value: "count", View: "groups", Order: "label", Title: "Change"}}, Hidden: true}, workbench.ViewCompositionChild{Timeline: &workbench.TimelineViewDeclaration{Version: "goo.view/1", Scope: workbench.ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: workbench.TimelineViewKind{Name: "time-bar", Version: 1}, Display: workbench.TimelineViewDisplay{Value: "duration.p95", RequiresComplete: true}}})
	target := source.Queries[0]
	target.Environment = "staging"
	target.GroupBy = []exploration.Field{{Kind: "attribute", Key: "http.response.status_code"}}
	raw, _ := json.Marshal(workbench.ViewTemplateExpansionRequest{Source: &source, Query: &target, BaselineCapture: &chosen})
	post := func(raw []byte) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/view-templates/expand", strings.NewReader(string(raw))))
		return w
	}
	got := post(raw)
	if got.Code != 200 {
		t.Fatal(got.Code, got.Body.String())
	}
	var out workbench.ViewTemplateExpansion
	if err = json.Unmarshal(got.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out.View.Queries[0], target) || out.View.Composition.Children[2].Comparison.BaselineCapture != chosen || !out.View.Composition.Children[2].Hidden || out.View.Composition.Children[2].Comparison.Display.View != "groups" || source.Composition.Children[2].Comparison.BaselineCapture != original {
		t.Fatal("scope or presentation transferred", got.Body.String())
	}
	// Neither reference needs to exist for pure expansion. Prepare, not this
	// configuration endpoint, must establish actual stored-evidence compatibility.
	for _, bad := range []string{
		strings.Replace(string(raw), `,"baselineCapture":"`+chosen+`"`, "", 1),
		strings.Replace(string(raw), `"baselineCapture":"`+chosen+`"`, `"baselineCapture":null`, 1),
		strings.Replace(string(raw), `"baselineCapture":"`+chosen+`"`, `"baselineCapture":""`, 1),
		strings.Replace(string(raw), `"baseline":"destination"`, `"baseline":null`, 1),
		strings.Replace(string(raw), `"baseline":"destination"`, `"baseline":""`, 1),
		strings.Replace(string(raw), `"baseline":"destination"`, `"Baseline":"destination"`, 1),
		string(raw[:len(raw)-1]) + `,"widget":null}`,
	} {
		if bad == string(raw) {
			t.Fatal("negative mutation missed")
		}
		if w := post([]byte(bad)); w.Code != 400 {
			t.Fatal("bad comparison expansion accepted", w.Code, w.Body.String())
		}
	}
	var plain workbench.ExplorationView
	json.Unmarshal(plainRaw, &plain)
	q := plain.Queries[0]
	extra, _ := json.Marshal(workbench.ViewTemplateExpansionRequest{Source: &plain, Query: &q, BaselineCapture: &chosen})
	if w := post(extra); w.Code != 400 {
		t.Fatal("ordinary expansion accepted baseline", w.Code, w.Body.String())
	}
	for _, v := range []string{`null`, `""`} {
		bad := strings.Replace(string(extra), `"group":"destination"`, `"group":"destination","baseline":`+v, 1)
		if w := post([]byte(bad)); w.Code != 400 {
			t.Fatal("ordinary null/empty binding accepted", w.Code, w.Body.String())
		}
	}
	if reads != 0 {
		t.Fatal("expansion fetched telemetry or evidence", reads)
	}
	if _, err := os.Stat(filepath.Join(dir, "workbench")); !os.IsNotExist(err) {
		t.Fatal("expansion published configuration", err)
	}
}

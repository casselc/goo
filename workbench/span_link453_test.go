package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func linksTarget453() *exploration.SpanLinkTarget {
	return &exploration.SpanLinkTarget{TraceID: strings.Repeat("a", 32), SpanID: strings.Repeat("b", 16)}
}

func TestLinksTo453FixedViewAndStepRetainCompleteSelector(t *testing.T) {
	for _, grouped := range []bool{false, true} {
		v := compositionView()
		if !grouped {
			v.Composition = nil
			v.Layout = "records"
			v.Columns = []string{"utc", "summary"}
			v.Queries[0].GroupBy = nil
		}
		v.Queries[0].LinksTo = linksTarget453()
		store := Store{Dir: t.TempDir()}
		saved, err := store.CreateExplorationView(v)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := store.LoadExplorationView(saved.ID)
		if err != nil || !reflect.DeepEqual(loaded.Queries, v.Queries) {
			t.Fatal("fixed view changed reported-link scope", err)
		}
		encoded, err := EncodeViewArtifact(ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "referencing-spans", Revision: "v1", View: v})
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := DecodeViewArtifact(encoded.YAML)
		if err != nil || !reflect.DeepEqual(decoded.Artifact.View.Queries, v.Queries) {
			t.Fatal("portable view changed reported-link scope", err)
		}
		q := v.Queries[0]
		step := Step{Version: 1, Kind: "record", Scope: StepScope{Environment: q.Environment, FromMS: q.FromMS, ToMS: q.ToMS}, View: "Records", Record: &q}
		if err := step.Validate(); err != nil {
			t.Fatal(err)
		}
		raw, err := json.Marshal(step)
		if err != nil {
			t.Fatal(err)
		}
		var roundtrip Step
		if err = json.Unmarshal(raw, &roundtrip); err != nil || !reflect.DeepEqual(roundtrip.Record, &q) {
			t.Fatal("step dropped reported-link scope", err)
		}
	}
}

func TestLinksTo453SpecializedViewsRefuseWithoutChangingQuestion(t *testing.T) {
	cases := map[string]ExplorationView{
		"record-only composition": recordOnlyView302("spans", "record-table"),
		"summary":                 summaryView400("spans"),
		"summary template":        summaryTemplate404("spans"),
		"histogram":               histogramView407(histogramQuestion407("spans")),
		"facet":                   facetComposition(true),
	}
	for name, v := range cases {
		t.Run(name, func(t *testing.T) {
			if err := v.Validate(); err != nil {
				t.Fatal("invalid fixture", err)
			}
			v.Queries[0].LinksTo = linksTarget453()
			before, _ := json.Marshal(v)
			if err := v.Validate(); err == nil {
				t.Fatal("specialized view accepted reported-link scope")
			}
			if _, err := (Store{Dir: t.TempDir()}).CreateExplorationView(v); err == nil {
				t.Fatal("unsupported view published")
			}
			after, _ := json.Marshal(v)
			if string(before) != string(after) {
				t.Fatal("refusal changed query")
			}
		})
	}
	in := httpRouteTemplateInput(t)
	in.Widget.Query.LinksTo = linksTarget453()
	before, _ := json.Marshal(in)
	if _, err := ExpandViewTemplate(in); err == nil {
		t.Fatal("shipped profile silently admitted reported-link scope")
	}
	after, _ := json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("template refusal changed query")
	}
}

func TestLinksTo453RollingTailRefusesRatherThanDroppingFixedSelector(t *testing.T) {
	v := ExplorationView{Version: 1, Title: "Referencing spans", Layout: "records", Columns: []string{"utc", "summary"}, Queries: []exploration.Query{{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, LinksTo: linksTarget453()}}}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	v.Tail = &ExplorationTail{Signal: "spans"}
	before, _ := json.Marshal(v)
	if err := v.Validate(); err == nil {
		t.Fatal("rolling tail admitted unsupported link membership")
	}
	after, _ := json.Marshal(v)
	if string(before) != string(after) {
		t.Fatal("tail refusal changed fixed selector")
	}
}

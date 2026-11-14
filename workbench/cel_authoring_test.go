package workbench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
)

func celView(t *testing.T, source string) (ExplorationView, querycel.PlanRequest) {
	t.Helper()
	base := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20, Where: &exploration.Predicate{Any: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "missing"},
		{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"},
	}}}
	request, err := querycel.NewPlanRequest(base, source, "condition", "mount", "9007199254740993")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := querycel.PlanRecord(request)
	if err != nil {
		t.Fatal(err)
	}
	authoring, err := NewRecordAuthoring(request, "parent-1")
	if err != nil {
		t.Fatal(err)
	}
	view := ExplorationView{Version: 1, Title: "Authored logs", Queries: []exploration.Query{plan.Question}, Layout: "records", Columns: []string{"utc", "service", "summary", "context"}, Authoring: map[string]RecordAuthoring{"logs": authoring}}
	if err := view.Validate(); err != nil {
		t.Fatal(err)
	}
	return view, request
}

func TestCELAuthoringStoreAndPortableRoundTrip(t *testing.T) {
	const source = "// Keep this exact source: 雪 < &\nlog.body.contains(\"timeout\") && log.severity_number >= 17u"
	view, request := celView(t, source)
	a := view.Authoring["logs"]
	request.Base.Where.Any[0].Field.Key = "changed"
	if a.Base.Where.Any[0].Field.Key != "team" {
		t.Fatal("constructor aliases request")
	}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(view)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadExplorationView(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	loaded.ID = ""
	if !reflect.DeepEqual(*loaded, view) || loaded.Authoring["logs"].Source != source {
		t.Fatal("store changed exact source or typed meaning")
	}
	artifact := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "authored-logs", Revision: "v1", View: *loaded}
	encoded, err := EncodeViewArtifact(artifact)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(encoded.YAML)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Artifact, artifact) {
		t.Fatal("YAML changed authoring pair")
	}
	imported, err := store.CreateExplorationView(decoded.Artifact.View)
	if err != nil || imported.ID != saved.ID {
		t.Fatal("portable import changed immutable definition", err)
	}
}

func TestCELAuthoringSourceAndLoweringHaveDifferentIdentities(t *testing.T) {
	original, _ := celView(t, `log.body == "timeout"`)
	formatted, _ := celView(t, "// A recoverable comment\n  log.body == \"timeout\"  ")
	changed, _ := celView(t, `log.body == "other"`)
	if original.Authoring["logs"].LoweringDigest != formatted.Authoring["logs"].LoweringDigest {
		t.Fatal("formatting changed lowering")
	}
	if original.Authoring["logs"].LoweringDigest == changed.Authoring["logs"].LoweringDigest {
		t.Fatal("changed meaning retained lowering digest")
	}
	first, _ := ViewContentID(original)
	second, _ := ViewContentID(formatted)
	if first == second {
		t.Fatal("exact-source change did not change definition identity")
	}
	// Revision labels are provenance correlation, not authenticated signatures.
	label := original.Authoring["logs"]
	label.ParentRevision = "parent-2"
	label.Binding.Revision = "another-draft"
	original.Authoring["logs"] = label
	if err := original.Validate(); err != nil {
		t.Fatal("valid imported revision labels refused", err)
	}
}

func TestCELAuthoringRejectsMismatchedPairsBeforeStorage(t *testing.T) {
	edits := map[string]func(*ExplorationView, *RecordAuthoring){
		"source":          func(_ *ExplorationView, a *RecordAuthoring) { a.Source = `log.body == "other"` },
		"digest":          func(_ *ExplorationView, a *RecordAuthoring) { a.LoweringDigest = strings.Repeat("a", 64) },
		"library":         func(_ *ExplorationView, a *RecordAuthoring) { a.Binding.Library = "goo.cel.otel/v0" },
		"compiler":        func(_ *ExplorationView, a *RecordAuthoring) { a.Binding.Compiler = "old" },
		"target":          func(_ *ExplorationView, a *RecordAuthoring) { a.Binding.Target = "databricks" },
		"profile":         func(_ *ExplorationView, a *RecordAuthoring) { a.Binding.Profile = "native" },
		"base binding":    func(_ *ExplorationView, a *RecordAuthoring) { a.Binding.BaseIdentity = strings.Repeat("b", 64) },
		"base":            func(_ *ExplorationView, a *RecordAuthoring) { a.Base.FromMS++ },
		"question policy": func(v *ExplorationView, _ *RecordAuthoring) { v.Queries[0].Limit++ },
		"question filter": func(v *ExplorationView, _ *RecordAuthoring) { v.Queries[0].Where.All[0].Any[1].Value = "other" },
		"parent revision": func(_ *ExplorationView, a *RecordAuthoring) { a.ParentRevision = "bad revision" },
		"representation":  func(_ *ExplorationView, a *RecordAuthoring) { a.Representation = "visual" },
		"version":         func(_ *ExplorationView, a *RecordAuthoring) { a.Version = "old" },
		"invalid source":  func(_ *ExplorationView, a *RecordAuthoring) { a.Source = `false && log.body.matches(".*")` },
	}
	for name, edit := range edits {
		t.Run(name, func(t *testing.T) {
			v, _ := celView(t, `log.body == "timeout"`)
			a := v.Authoring["logs"]
			edit(&v, &a)
			v.Authoring["logs"] = a
			store := Store{Dir: t.TempDir()}
			if _, err := store.CreateExplorationView(v); err == nil {
				t.Fatal("mismatched authoring published")
			}
			if _, err := os.Stat(filepath.Join(store.Dir, "exploration-views")); !os.IsNotExist(err) {
				t.Fatal("refused metadata touched store", err)
			}
		})
	}
	// Composition's early return must not bypass authoring checks.
	v, _ := celView(t, `log.body == "timeout"`)
	v.Layout = ""
	v.Columns = nil
	v.Composition = &ViewComposition{Version: "goo.composition/1", Kind: CompositionKind{Name: "stack", Version: 1}, Children: []ViewCompositionChild{{Records: &RecordViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "logs"}, Kind: RecordViewKind{Name: "record-table", Version: 1}, Display: RecordViewDisplay{Columns: []string{"utc", "service", "summary", "context"}}}}}}
	if err := v.Validate(); err != nil {
		t.Fatal(err)
	}
	a := v.Authoring["logs"]
	a.Binding.Library = "old"
	v.Authoring["logs"] = a
	if err := v.Validate(); err == nil {
		t.Fatal("composition bypassed authoring")
	}
}

func TestCELAuthoringStrictWireAndStoredIntegrity(t *testing.T) {
	v, _ := celView(t, `log.body == "timeout"`)
	a := v.Authoring["logs"]
	raw, _ := json.Marshal(a)
	var obj map[string]any
	json.Unmarshal(raw, &obj)
	for _, name := range []string{"version", "representation", "source", "base", "binding", "parentRevision", "loweringDigest"} {
		copy := map[string]any{}
		for k, value := range obj {
			copy[k] = value
		}
		copy[name] = nil
		data, _ := json.Marshal(copy)
		var out RecordAuthoring
		if json.Unmarshal(data, &out) == nil {
			t.Fatalf("null %s accepted", name)
		}
		delete(copy, name)
		data, _ = json.Marshal(copy)
		if json.Unmarshal(data, &out) == nil {
			t.Fatalf("missing %s accepted", name)
		}
	}
	for _, data := range []string{"null", `{}`, strings.Replace(string(raw), `"source":`, `"Source":`, 1), strings.Replace(string(raw), `"source":"log.body`, `"source":"\ud800log.body`, 1), strings.Replace(string(raw), `"binding":{`, `"binding":{"unknown":"x",`, 1)} {
		var out RecordAuthoring
		if json.Unmarshal([]byte(data), &out) == nil {
			t.Fatal("unsafe authoring JSON accepted", data)
		}
	}
	viewRaw, _ := json.Marshal(v)
	var wire map[string]any
	json.Unmarshal(viewRaw, &wire)
	for _, name := range []string{"authoring", "Authoring"} {
		delete(wire, "authoring")
		wire[name] = nil
		payload, _ := json.Marshal(wire)
		var out ExplorationView
		if json.Unmarshal(payload, &out) == nil {
			t.Fatal("null or case-folded authoring accepted", name)
		}
		delete(wire, name)
	}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(v)
	if err != nil {
		t.Fatal(err)
	}
	savedAuthor := saved.Authoring["logs"]
	savedAuthor.Source = "// changed source only\n" + savedAuthor.Source
	saved.Authoring["logs"] = savedAuthor
	modified, _ := json.Marshal(saved)
	if err := os.WriteFile(filepath.Join(store.Dir, "exploration-views", saved.ID+".json"), modified, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.LoadExplorationView(saved.ID); err == nil {
		t.Fatal("changed provenance under old ID accepted")
	}
}

func TestCELAuthoringDuplicateOuterObjectUsesLastValue(t *testing.T) {
	view, _ := celView(t, `log.body == "timeout"`)
	logs := view.Authoring["logs"]
	base := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 20}
	request, err := querycel.NewPlanRequest(base, `span.name == "work"`, "span-condition", "mount", "1")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := querycel.PlanRecord(request)
	if err != nil {
		t.Fatal(err)
	}
	spans, err := NewRecordAuthoring(request, "parent")
	if err != nil {
		t.Fatal(err)
	}
	view.Queries = append(view.Queries, plan.Question)
	view.Authoring = nil
	raw, _ := json.Marshal(view)
	first, _ := json.Marshal(map[string]RecordAuthoring{"logs": logs})
	last, _ := json.Marshal(map[string]RecordAuthoring{"spans": spans})
	payload := strings.TrimSuffix(string(raw), "}") + `,"authoring":` + string(first) + `,"authoring":` + string(last) + `}`
	var out ExplorationView
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Authoring) != 1 || out.Authoring["spans"].Source != spans.Source {
		t.Fatal("duplicate provenance maps merged rather than replaced")
	}
	if err := out.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestCELAuthoringTemplateExpansionKeepsDestinationAuthority(t *testing.T) {
	source := compositionView()
	request, err := querycel.NewPlanRequest(source.Queries[0], `span.name == "work"`, "condition", "mount", "1")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := querycel.PlanRecord(request)
	if err != nil {
		t.Fatal(err)
	}
	authoring, err := NewRecordAuthoring(request, "parent")
	if err != nil {
		t.Fatal(err)
	}
	source.Queries[0] = plan.Question
	source.Authoring = map[string]RecordAuthoring{"spans": authoring}
	source.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
	if err := source.Validate(); err != nil {
		t.Fatal(err)
	}
	destination := source.Queries[0]
	destination.Environment = "staging"
	destination.FromMS = 2000
	destination.ToMS = 3000
	destination.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "destination"}
	before, _ := json.Marshal(source)
	expanded, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &destination})
	if err != nil {
		t.Fatal(err)
	}
	if expanded.View.Authoring != nil || !reflect.DeepEqual(expanded.View.Queries[0], destination) {
		t.Fatal("template transferred original expression or donor query")
	}
	after, _ := json.Marshal(source)
	if string(before) != string(after) {
		t.Fatal("template mutated source provenance")
	}
}

func TestCELAuthoringOuterWirePreservesSourceUnicodeRefusal(t *testing.T) {
	view, _ := celView(t, `log.body == "timeout"`)
	raw, _ := json.Marshal(view)
	payload := strings.Replace(string(raw), `"source":"log.body`, `"source":"\ud800log.body`, 1)
	var decoded ExplorationView
	if err := json.Unmarshal([]byte(payload), &decoded); err == nil {
		t.Fatal("view repaired malformed exact source Unicode")
	}
}

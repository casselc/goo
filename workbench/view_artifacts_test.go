package workbench

import (
	"os"
	"reflect"
	"strings"
	"testing"
)

func portableArtifact() ViewArtifact {
	return ViewArtifact{Version: ViewArtifactVersion, Namespace: "payments", Name: "http-services", Revision: "v1", View: compositionView()}
}
func TestViewArtifactRoundTrip(t *testing.T) {
	a := portableArtifact()
	encoded, err := EncodeViewArtifact(a)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(encoded.YAML)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Artifact, a) || decoded.YAML != encoded.YAML {
		t.Fatalf("roundtrip changed fixed question: %#v", decoded)
	}
	store := Store{Dir: t.TempDir()}
	saved, err := store.CreateExplorationView(decoded.Artifact.View)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := store.LoadExplorationView(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	loaded.ID = ""
	if !reflect.DeepEqual(*loaded, a.View) {
		t.Fatal("ordinary save changed imported view")
	}
	decoded.Artifact.View.Queries[0].Signal = "mutated"
	if a.View.Queries[0].Signal == "mutated" {
		t.Fatal("codec retained caller memory")
	}
}
func TestViewArtifactRejectsUnsafeYAML(t *testing.T) {
	encoded, err := EncodeViewArtifact(portableArtifact())
	if err != nil {
		t.Fatal(err)
	}
	for name, input := range map[string]string{
		"huge integer": "version: 184467440737095516160\n", "empty": "", "scalar": "hello", "duplicate": "version: a\nversion: b\n", "duplicate nested": strings.Replace(encoded.YAML, "title:", "title: old\n    title:", 1),
		"unknown": encoded.YAML + "unknown: x\n", "unknown view": strings.Replace(encoded.YAML, "    title:", "    unknown: x\n    title:", 1),
		"alias": "a: &x [1]\nb: *x\n", "unused anchor": "a: &x hi\n", "merge": "a: {<<: {x: 1}}", "tag": "a: !custom x", "explicit standard": "a: !!str x",
		"timestamp": "a: 2026-11-01", "nonfinite": "a: .nan", "numeric key": "1: x", "bool key": "true: x", "multidoc": encoded.YAML + "---\n{}", "oversize": strings.Repeat(" ", ViewArtifactBytes+1),
		"version": strings.Replace(encoded.YAML, ViewArtifactVersion, "goo.view-artifact/2", 1), "string number": strings.Replace(encoded.YAML, "revision: v1", "revision: 123", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeViewArtifact(input); err == nil {
				t.Fatal("unsafe artifact accepted")
			}
		})
	}
}
func TestViewArtifactPortableReferences(t *testing.T) {
	a := portableArtifact()
	a.View.ID = strings.Repeat("a", 64)
	if _, err := EncodeViewArtifact(a); err == nil {
		t.Fatal("local view ID accepted")
	}
	a = portableArtifact()
	a.Namespace = "Team/Service"
	if _, err := EncodeViewArtifact(a); err == nil {
		t.Fatal("invalid namespace accepted")
	}
	a = portableArtifact()
	a.View.Composition.Children = append(a.View.Composition.Children, ViewCompositionChild{Comparison: &CountComparisonViewDeclaration{Version: "goo.view/1", Scope: ViewDeclarationScope{Kind: "record-query", Signal: "spans"}, Kind: CountComparisonViewKind{Name: "captured-count-comparison", Version: 1}, BaselineCapture: strings.Repeat("a", 64), Display: CountComparisonViewDisplay{Value: "count"}}})
	if _, err := EncodeViewArtifact(a); err == nil {
		t.Fatal("local comparison accepted")
	}
	a = portableArtifact()
	a.View = authoredView()
	a.View.Tail = &ExplorationTail{Signal: "spans", Baseline: strings.Repeat("a", 64)}
	if _, err := EncodeViewArtifact(a); err == nil {
		t.Fatal("local tail baseline accepted")
	}
	a.View.Tail = nil
	if _, err := EncodeViewArtifact(a); err != nil {
		t.Fatal("flat authored view rejected", err)
	}
}

func TestViewArtifactFixedBaselineAndTemplate(t *testing.T) {
	a := portableArtifact()
	a.View = authoredView()
	baseline := a.View.Queries[0]
	baseline.FromMS = 61000
	baseline.ToMS = 121000
	a.View.Displays = nil
	a.View.Breakdowns = map[string]ExplorationBreakdown{"spans": {Metric: "count", ComparisonMetric: "count", Baseline: &baseline}}
	out, err := EncodeViewArtifact(a)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(out.YAML)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Artifact, a) {
		t.Fatal("portable query baseline changed")
	}
	a = portableArtifact()
	a.View.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
	out, err = EncodeViewArtifact(a)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = DecodeViewArtifact(out.YAML); err != nil {
		t.Fatal(err)
	}
}

func TestViewArtifactCodeManagedExample(t *testing.T) {
	a := portableArtifact()
	a.Namespace = "platform"
	a.Name = "service-requests"
	a.Revision = "v1"
	a.View.Title = "Service span volume"
	a.View.Queries[0].FromMS = 1767225600000
	a.View.Queries[0].ToMS = 1767225900000
	a.View.Queries[0].Limit = 100
	a.View.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
	encoded, err := EncodeViewArtifact(a)
	if err != nil {
		t.Fatal(err)
	}
	const filename = "../config/views/service-requests.yaml"
	if os.Getenv("GOO_UPDATE_VIEW_ARTIFACT_EXAMPLE") == "1" {
		if err = os.WriteFile(filename, []byte(encoded.YAML), 0644); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(string(raw))
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Artifact, a) || decoded.YAML != encoded.YAML {
		t.Fatal("code-managed example differs from checked portable configuration")
	}
}

func TestViewArtifactParserBudgetsAndExactIntegers(t *testing.T) {
	for name, tc := range map[string]struct{ input, want string }{
		"integer overflow": {"version: 184467440737095516160\n", "exact 64-bit"},
		"depth":            {"a: " + strings.Repeat("[", 34) + "0" + strings.Repeat("]", 34), "structure budget"},
		"nodes":            {"a: [" + strings.Repeat("0,", 8193) + "0]", "structure budget"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeViewArtifact(tc.input)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("parser guard not reached: %v", err)
			}
		})
	}
}

func TestViewArtifactRequiresExistingQueryPageContract(t *testing.T) {
	raw, err := os.ReadFile("../config/views/service-requests.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for name, input := range map[string]string{
		"omitted":  strings.Replace(string(raw), "          limit: 100\n", "", 1),
		"zero":     strings.Replace(string(raw), "limit: 100", "limit: 0", 1),
		"negative": strings.Replace(string(raw), "limit: 100", "limit: -1", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := DecodeViewArtifact(input); err == nil {
				t.Fatal("invalid page contract accepted")
			}
		})
	}
}

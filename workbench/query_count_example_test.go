package workbench

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestFixedQueryCountCodeManagedExample308(t *testing.T) {
	view := fixedQueryComparisonView()
	view.Title = "Canary and stable service spans"
	view.Queries[0].FromMS = 1767225600000
	view.Queries[0].ToMS = 1767225900000
	view.Queries[0].Limit = 100
	view.Queries[0].Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "deployment.ring"}, Op: "eq", Value: "canary"}
	baseline := view.Queries[0]
	baseline.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "deployment.ring"}, Op: "eq", Value: "stable"}
	view.Composition.Children[2].Comparison.BaselineQuery = &baseline
	view.Composition.Children[2].Comparison.Display.Title = "Canary vs stable"
	artifact := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "deployment-ring-cohorts", Revision: "v1", View: view}
	encoded, err := EncodeViewArtifact(artifact)
	if err != nil {
		t.Fatal(err)
	}
	const filename = "../config/views/fixed-cohort-comparison.yaml"
	if os.Getenv("GOO_UPDATE_FIXED_COHORT_EXAMPLE") == "1" {
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
	if err = decoded.Artifact.View.Validate(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.Artifact, artifact) || decoded.YAML != encoded.YAML {
		t.Fatal("checked fixed cohort example differs from canonical AST")
	}
	if decoded.Artifact.View.Template != nil || strings.Contains(decoded.YAML, "baselineCapture") {
		t.Fatal("portable fixed pair gained store-local identity or reuse binding")
	}
	// A handwritten baseline must still satisfy the existing explicit page
	// contract; the codec cannot return a zero limit the UI cannot open.
	parts := strings.SplitN(encoded.YAML, "baselineQuery:", 2)
	if len(parts) != 2 {
		t.Fatal("fixed baseline missing from example")
	}
	lines := strings.Split(parts[1], "\n")
	removed := false
	for i, line := range lines {
		if strings.TrimSpace(line) == "limit: 100" {
			lines = append(lines[:i], lines[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		t.Fatal("baseline limit missing from checked example")
	}
	if _, err = DecodeViewArtifact(parts[0] + "baselineQuery:" + strings.Join(lines, "\n")); err == nil {
		t.Fatal("omitted baseline page limit admitted")
	}
}

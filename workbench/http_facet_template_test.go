package workbench

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
)

func facetChild(t *testing.T, v ExplorationView) *FacetViewDeclaration {
	t.Helper()
	for _, child := range v.Composition.Children {
		if child.Facet != nil {
			return child.Facet
		}
	}
	t.Fatal("missing facet child")
	return nil
}
func TestHTTPMethodFacetDefaultsAndSpecializedPreservation311(t *testing.T) {
	for _, numeric := range []bool{true, false, true, false} {
		input := httpRouteTemplateInput(t)
		if numeric {
			input.Widget.Query.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
		}
		before, _ := json.Marshal(input)
		expanded, err := ExpandViewTemplate(input)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expanded.View.Queries[0], *input.Widget.Query) {
			t.Fatal("HTTP default changed the checked query")
		}
		if numeric {
			if expanded.View.Template != nil || len(expanded.View.Composition.Children) != 2 {
				t.Fatal("specialized query acquired unsupported facet/template")
			}
			for _, child := range expanded.View.Composition.Children {
				if child.Facet != nil {
					t.Fatal("specialized query acquired method facet")
				}
			}
		} else {
			facet := facetChild(t, expanded.View)
			if len(expanded.View.Composition.Children) != 3 || expanded.View.Composition.Children[0].Facet == nil || expanded.View.Template == nil || expanded.View.Template.Profile != "otel.http-server" || facet.Field != (exploration.Field{Kind: "attribute", Key: "http.request.method"}) || facet.Display.Title != "HTTP methods" || facet.Display.Basis != "applied" {
				t.Fatal("ordinary HTTP default lost its complementary method facet")
			}
			facet.Field.Key = "mutated"
		}
		after, _ := json.Marshal(input)
		if string(before) != string(after) {
			t.Fatal("shipped default mutated request")
		}
	}
}
func TestOrdinaryFacetTemplateRetainsLiteralFieldAndDestination311(t *testing.T) {
	for _, hidden := range []bool{false, true} {
		source := facetComposition(true)
		source.Composition.Children[2].Hidden = hidden
		source.Template = &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: "ordinary"}
		store := Store{Dir: t.TempDir()}
		saved, err := store.CreateExplorationView(source)
		if err != nil {
			t.Fatal(err)
		}
		target := source.Queries[0]
		target.GroupBy = []exploration.Field{{Kind: "resource", Key: "team"}}
		target.Environment = "staging"
		target.FromMS = 61000
		target.ToMS = 121000
		target.Limit = 70
		target.Offset = 0
		target.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: "payments"}
		expected := target
		result, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: saved, Query: &target})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result.View.Queries[0], expected) || !reflect.DeepEqual(*facetChild(t, result.View), *facetChild(t, source)) || result.View.Composition.Children[2].Hidden != hidden {
			t.Fatal("facet literal configuration or destination query changed")
		}
		facetChild(t, result.View).Field.Key = "mutated"
		loaded, err := store.LoadExplorationView(saved.ID)
		if err != nil || facetChild(t, *loaded).Field.Key != "http.route" {
			t.Fatal("expansion mutated the saved facet", err)
		}
		target.Numeric = &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
		if _, err = ExpandViewTemplate(ViewTemplateExpansionRequest{Source: saved, Query: &target}); err == nil {
			t.Fatal("authored facet template silently lost its control for an invalid destination")
		}
	}
}
func TestHTTPFacetCodeManagedTemplateExample311(t *testing.T) {
	input := httpRouteTemplateInput(t)
	input.Widget.Query.FromMS = 1767225600000
	input.Widget.Query.ToMS = 1767225900000
	expanded, err := ExpandViewTemplate(input)
	if err != nil {
		t.Fatal(err)
	}
	artifact := ViewArtifact{Version: ViewArtifactVersion, Namespace: "platform", Name: "http-server-routes", Revision: "v1", View: expanded.View}
	encoded, err := EncodeViewArtifact(artifact)
	if err != nil {
		t.Fatal(err)
	}
	const file = "../config/views/http-server-routes.yaml"
	if os.Getenv("GOO_UPDATE_HTTP_FACET_EXAMPLE") == "1" {
		if err = os.WriteFile(file, []byte(encoded.YAML), 0644); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeViewArtifact(string(raw))
	if err != nil || !reflect.DeepEqual(decoded.Artifact, artifact) || decoded.YAML != encoded.YAML {
		t.Fatal("HTTP method facet example drift", err)
	}
}

// Explicit titles are user content, even when they match a former shipped default.
func TestHTTPBreakdownTitleIsOptionalAndAuthoredTitlesRemainLiteral313(t *testing.T) {
	aggregate := func(v ExplorationView) *ViewDeclaration {
		t.Helper()
		for _, child := range v.Composition.Children {
			if child.Aggregate != nil {
				return child.Aggregate
			}
		}
		t.Fatal("missing aggregate")
		return nil
	}
	expanded, err := ExpandViewTemplate(httpRouteTemplateInput(t))
	if err != nil {
		t.Fatal(err)
	}
	if expanded.View.Title != "HTTP server routes" || aggregate(expanded.View).Display.Title != "" {
		t.Fatal("shipped view identity or computed-title default changed", expanded.View)
	}
	for _, title := range []string{"HTTP server routes", "Our Ω incident"} {
		source := expanded.View
		aggregate(source).Display.Title = title
		target := source.Queries[0]
		target.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		out, err := ExpandViewTemplate(ViewTemplateExpansionRequest{Source: &source, Query: &target})
		if err != nil {
			t.Fatal(err)
		}
		if aggregate(out.View).Display.Title != title || !reflect.DeepEqual(out.View.Queries[0].GroupBy, target.GroupBy) {
			t.Fatal("authored title or effective destination was rewritten", out.View)
		}
		aggregate(out.View).Display.Title = "changed"
		if aggregate(source).Display.Title != title {
			t.Fatal("authored title aliases expansion")
		}
	}
}

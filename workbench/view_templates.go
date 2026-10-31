package workbench

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"

	"github.com/casselc/goo/exploration"
)

// The shipped descriptor contains bounded presentation and eligibility data,
// never executable expressions or a substitute source-query definition.
//
//go:embed templates/http-server-routes.json
var httpServerRoutesTemplate []byte

type ViewTemplateReference struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Title   string `json:"title"`
}

type ViewTemplateExpansionRequest struct {
	Template string                      `json:"template"`
	Version  int                         `json:"version"`
	Widget   exploration.DashboardWidget `json:"widget"`
}

type ViewTemplateExpansion struct {
	Template ViewTemplateReference `json:"template"`
	View     ExplorationView       `json:"view"`
}

type recordViewTemplate struct {
	ViewTemplateReference
	Family       string                  `json:"family"`
	Grain        string                  `json:"grain"`
	Signal       string                  `json:"signal"`
	Group        exploration.Field       `json:"group"`
	ServiceField exploration.Field       `json:"serviceField"`
	Required     []exploration.Predicate `json:"required"`
	Composition  ViewComposition         `json:"composition"`
}

// ExpandViewTemplate stages existing authored-view configuration. It neither
// reads telemetry nor publishes a saved revision. The supplied widget is
// advisory metadata: its query must independently retain the mandatory HTTP
// population constraints. Extra filters are copied unchanged, not reconstructed.
func ExpandViewTemplate(in ViewTemplateExpansionRequest) (ViewTemplateExpansion, error) {
	var out ViewTemplateExpansion
	var template recordViewTemplate
	decoder := json.NewDecoder(bytes.NewReader(httpServerRoutesTemplate))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&template); err != nil {
		return out, errors.New("invalid shipped view template")
	}
	if in.Template != template.ID || in.Version != template.Version {
		return out, errors.New("choose a supported pinned view template")
	}
	w := in.Widget
	if w.State != "supported" || w.Key != "dimension" || w.Family != template.Family || w.Grain != template.Grain || w.Query == nil || w.Measurement != nil {
		return out, errors.New("template requires a supported HTTP server route record widget")
	}
	q := w.Query
	if _, err := exploration.Compile(*q); err != nil {
		return out, errors.New("template requires a valid checked record query")
	}
	if q.Signal != template.Signal || len(q.GroupBy) != 1 || q.GroupBy[0] != template.Group || q.Offset != 0 || q.ReportedEdge != nil {
		return out, errors.New("template requires a page-zero HTTP route span population")
	}
	service := false
	for _, p := range conjuncts(q.Where) {
		if p.Field != nil && *p.Field == template.ServiceField && p.Op == "eq" && p.Value != "" {
			service = true
		}
	}
	if !service {
		return out, errors.New("template requires an exact nonempty service selector")
	}
	for _, required := range template.Required {
		found := false
		for _, p := range conjuncts(q.Where) {
			if p.Field != nil && required.Field != nil && *p.Field == *required.Field && p.Op == required.Op && p.Value == required.Value {
				found = true
			}
		}
		if !found {
			return out, errors.New("template requires mandatory Server kind and nonempty current HTTP method constraints")
		}
	}
	// JSON cloning detaches nested predicates/bindings/environment selectors while
	// retaining their exact representation. No normalized or weakened query is made.
	raw, err := json.Marshal(q)
	if err != nil {
		return out, err
	}
	var query exploration.Query
	if err := json.Unmarshal(raw, &query); err != nil {
		return out, err
	}
	out.Template = template.ViewTemplateReference
	out.View = ExplorationView{Version: 1, Title: template.Title, Queries: []exploration.Query{query}, Composition: &template.Composition}
	if err := out.View.Validate(); err != nil {
		return ViewTemplateExpansion{}, err
	}
	return out, nil
}

// Only predicates reached through All are mandatory. An occurrence beneath Any
// cannot establish the Server/method/service contract; this is deliberately not
// a general logical implication solver.
func conjuncts(p *exploration.Predicate) []exploration.Predicate {
	if p == nil {
		return nil
	}
	if len(p.Any) > 0 {
		return nil
	}
	if len(p.All) == 0 {
		return []exploration.Predicate{*p}
	}
	var out []exploration.Predicate
	for i := range p.All {
		out = append(out, conjuncts(&p.All[i])...)
	}
	return out
}

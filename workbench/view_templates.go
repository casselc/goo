package workbench

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"reflect"

	"github.com/casselc/goo/exploration"
)

// Shipped descriptors are bounded eligibility and presentation data, not code.
//
//go:embed templates/http-server-routes.json
var httpServerRoutesTemplate []byte

//go:embed templates/db-client-operations.json
var dbClientOperationsTemplate []byte

type ViewTemplateReference struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Title   string `json:"title"`
}

// ViewTemplateBinding opts a saved composition into the destination's existing
// category field. It never supplies a destination query or execution authority.
type ViewTemplateBinding struct {
	Version string `json:"version"`
	Group   string `json:"group"`
	Profile string `json:"profile"`
}

type ViewTemplateExpansionRequest struct {
	Template string                      `json:"template,omitempty"`
	Version  int                         `json:"version,omitempty"`
	Widget   exploration.DashboardWidget `json:"widget,omitzero"`
	Source   *ExplorationView            `json:"source,omitempty"`
	Query    *exploration.Query          `json:"query,omitempty"`
}

// Keep the two request arms exclusive even for explicit empty/null wire fields.
func (r *ViewTemplateExpansionRequest) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	_, source := fields["source"]
	_, query := fields["query"]
	if source || query {
		if !source || !query || len(fields) != 2 || bytes.Equal(bytes.TrimSpace(fields["source"]), []byte("null")) || bytes.Equal(bytes.TrimSpace(fields["query"]), []byte("null")) {
			return errors.New("choose exactly source and query for authored template expansion")
		}
	} else {
		if len(fields) != 3 {
			return errors.New("choose exactly template, version and widget for shipped expansion")
		}
		for _, key := range []string{"template", "version", "widget"} {
			value, ok := fields[key]
			if !ok || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				return errors.New("require template, version and widget")
			}
		}
	}
	type wire ViewTemplateExpansionRequest
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*r = ViewTemplateExpansionRequest(value)
	return nil
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

func shippedTemplate(id string) (recordViewTemplate, error) {
	var data []byte
	switch id {
	case "otel.http-server-routes":
		data = httpServerRoutesTemplate
	case "otel.db-client-operations":
		data = dbClientOperationsTemplate
	default:
		return recordViewTemplate{}, errors.New("choose a supported pinned view template")
	}
	var value recordViewTemplate
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return value, errors.New("invalid shipped view template")
	}
	return value, nil
}
func (b ViewTemplateBinding) validateProfile(q exploration.Query) error {
	if b.Version != "goo.template/1" || b.Group != "destination" {
		return errors.New("choose the pinned destination-group template binding")
	}
	var id string
	switch b.Profile {
	case "ordinary":
		return nil
	case "otel.http-server":
		id = "otel.http-server-routes"
	case "otel.db-client":
		id = "otel.db-client-operations"
	default:
		return errors.New("choose an ordinary, OTel HTTP server or OTel database client profile")
	}
	descriptor, err := shippedTemplate(id)
	if err != nil {
		return err
	}
	return descriptor.validateProfile(q)
}
func (d recordViewTemplate) validateProfile(q exploration.Query) error {
	if q.Signal != d.Signal {
		return errors.New("template profile requires spans")
	}
	service := false
	clauses := conjuncts(q.Where)
	for _, p := range clauses {
		if p.Field != nil && *p.Field == d.ServiceField && p.Op == "eq" && p.Value != "" {
			service = true
		}
	}
	if !service {
		return errors.New("template requires an exact nonempty service selector")
	}
	for _, required := range d.Required {
		found := false
		for _, p := range clauses {
			if p.Field != nil && required.Field != nil && *p.Field == *required.Field && p.Op == required.Op && p.Value == required.Value {
				found = true
				break
			}
		}
		if !found {
			return errors.New("template requires mandatory reported kind and nonempty current protocol attributes")
		}
	}
	return nil
}

func (b ViewTemplateBinding) Validate(q exploration.Query, c ViewComposition) error {
	if _, err := exploration.Compile(q); err != nil {
		return err
	}
	if len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.ReportedEdge != nil || q.GroupTimeSeries {
		return errors.New("reusable templates require one ordinary nonnumeric grouping")
	}
	if err := c.Validate(q); err != nil {
		return err
	}
	if len(c.Children) != 2 {
		return errors.New("reusable templates support aggregate and records only; remove timeline or comparison first")
	}
	for _, child := range c.Children {
		if child.Timeline != nil || child.Comparison != nil {
			return errors.New("templates cannot retain timelines or captured baseline references")
		}
		if child.Aggregate != nil && child.Aggregate.Display.Query != "" {
			return errors.New("clear local group Find before enabling a reusable template")
		}
	}
	return b.validateProfile(q)
}

// ExpandViewTemplate is configuration-only. Neither request arm reads a source,
// publishes a resource, or transfers a source result or completeness witness.
func ExpandViewTemplate(in ViewTemplateExpansionRequest) (ViewTemplateExpansion, error) {
	if in.Source != nil || in.Query != nil {
		if in.Source == nil || in.Query == nil || in.Template != "" || in.Version != 0 || !reflect.DeepEqual(in.Widget, exploration.DashboardWidget{}) {
			return ViewTemplateExpansion{}, errors.New("choose one complete template expansion arm")
		}
		return expandAuthoredTemplate(*in.Source, *in.Query)
	}
	descriptor, err := shippedTemplate(in.Template)
	if err != nil {
		return ViewTemplateExpansion{}, err
	}
	if in.Version != descriptor.Version {
		return ViewTemplateExpansion{}, errors.New("choose a supported pinned view template")
	}
	w := in.Widget
	if w.State != "supported" || w.Key != "dimension" || w.Family != descriptor.Family || w.Grain != descriptor.Grain || w.Query == nil || w.Measurement != nil {
		return ViewTemplateExpansion{}, errors.New("template requires its supported dimension record widget")
	}
	q := *w.Query
	if _, err = exploration.Compile(q); err != nil {
		return ViewTemplateExpansion{}, err
	}
	if q.Signal != descriptor.Signal || len(q.GroupBy) != 1 || q.GroupBy[0] != descriptor.Group || q.Offset != 0 || q.ReportedEdge != nil {
		return ViewTemplateExpansion{}, errors.New("template requires its page-zero qualified record grouping")
	}
	if err = descriptor.validateProfile(q); err != nil {
		return ViewTemplateExpansion{}, err
	}
	profile := "otel.http-server"
	if descriptor.Family == "db-client" {
		profile = "otel.db-client"
	}
	view := ExplorationView{Version: 1, Title: descriptor.Title, Queries: []exploration.Query{q}, Composition: &descriptor.Composition}
	binding := &ViewTemplateBinding{Version: "goo.template/1", Group: "destination", Profile: profile}
	// Some existing shipped widget queries include numeric measures. Preserve
	// their complete query, without falsely advertising them as authorable templates.
	if binding.Validate(q, descriptor.Composition) == nil {
		view.Template = binding
	}
	if err = view.Validate(); err != nil {
		return ViewTemplateExpansion{}, err
	}
	detached, err := cloneTemplateView(view)
	if err != nil {
		return ViewTemplateExpansion{}, err
	}
	return ViewTemplateExpansion{Template: descriptor.ViewTemplateReference, View: detached}, nil
}
func expandAuthoredTemplate(source ExplorationView, q exploration.Query) (ViewTemplateExpansion, error) {
	if err := source.Validate(); err != nil {
		return ViewTemplateExpansion{}, err
	}
	if source.Template == nil || source.Composition == nil {
		return ViewTemplateExpansion{}, errors.New("source must explicitly enable a reusable template")
	}
	originalID := source.ID
	source.ID = ""
	id, err := digest(source)
	if err != nil {
		return ViewTemplateExpansion{}, err
	}
	if originalID != "" && originalID != id {
		return ViewTemplateExpansion{}, errors.New("template source content ID differs from its configuration")
	}
	if _, err := exploration.Compile(q); err != nil {
		return ViewTemplateExpansion{}, err
	}
	q.Offset = 0
	if q.Signal != source.Queries[0].Signal {
		return ViewTemplateExpansion{}, errors.New("template and destination must use the same signal")
	}
	if err = source.Template.Validate(q, *source.Composition); err != nil {
		return ViewTemplateExpansion{}, err
	}
	view := ExplorationView{Version: 1, Title: source.Title, Queries: []exploration.Query{q}, Composition: source.Composition, Template: source.Template}
	detached, err := cloneTemplateView(view)
	if err != nil {
		return ViewTemplateExpansion{}, err
	}
	if err = detached.Validate(); err != nil {
		return ViewTemplateExpansion{}, err
	}
	return ViewTemplateExpansion{Template: ViewTemplateReference{ID: id, Version: 1, Title: source.Title}, View: detached}, nil
}
func cloneTemplateView(value ExplorationView) (ExplorationView, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return ExplorationView{}, err
	}
	var out ExplorationView
	err = json.Unmarshal(raw, &out)
	return out, err
}

// Only leaves reached through All are mandatory; Any cannot establish a profile.
func conjuncts(p *exploration.Predicate) []exploration.Predicate {
	if p == nil || len(p.Any) > 0 {
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

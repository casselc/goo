package exploration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/casselc/goo/measurements"
)

// QuestionScope is derived planning metadata, never part of a stored receipt.
// Selector and Binding retain exact typed constraints, including OTel namespaces.
// Unknown schema/convention metadata is not guessed from names or library versions.
type QuestionScope struct {
	ExplanationFields  []Field               `json:"explanationFields,omitempty"`
	ReportedEdge       *ReportedEdge         `json:"reportedEdge,omitempty"`
	Family             string                `json:"family"`
	Signal             string                `json:"signal"`
	MeasurementKind    string                `json:"measurementKind,omitempty"`
	SourceBinding      string                `json:"sourceBinding"`
	Grain              string                `json:"grain"`
	Environment        string                `json:"environment"`
	FromMS             int64                 `json:"fromMs"`
	ToMS               int64                 `json:"toMs"`
	Selector           *Predicate            `json:"selector,omitempty"`
	Binding            *measurements.Binding `json:"binding,omitempty"`
	DefinitionRevision string                `json:"definitionRevision,omitempty"`
	Convention         string                `json:"convention"`
	SchemaMetadata     string                `json:"schemaMetadata"`
	Members            []QuestionScope       `json:"members,omitempty"`
}

func detached[T any](value T) T {
	b, _ := json.Marshal(value)
	var out T
	json.Unmarshal(b, &out)
	return out
}

// Scope validates and detaches a question without reading telemetry. Presentation
// paging, grouping and chart choices do not define a record population.
func (q CaptureQuestion) Scope() (QuestionScope, error) {
	if err := q.Validate(); err != nil {
		return QuestionScope{}, err
	}
	s := QuestionScope{Family: q.Kind, Convention: "Unbound; no semantic-convention profile inferred", SchemaMetadata: "Unavailable in this record adapter; original resource/scope constraints remain in selector"}
	switch q.Kind {
	case "explanation":
		a, e := (CaptureQuestion{Kind: "explore", Explore: &q.Explanation.Baseline}).Scope()
		if e != nil {
			return s, e
		}
		b, e := (CaptureQuestion{Kind: "explore", Explore: &q.Explanation.Selection}).Scope()
		if e != nil {
			return s, e
		}
		s.Signal = a.Signal
		s.SourceBinding = a.SourceBinding
		s.Grain = "explanation-pair"
		s.Environment = a.Environment
		s.Members = []QuestionScope{a, b}
		s.ExplanationFields = append([]Field{}, q.Explanation.Fields...)
	case "explore", "inventory":
		var v Query
		if q.Explore != nil {
			v = *q.Explore
		} else {
			v = q.Inventory.Query()
		}
		s.Signal = v.Signal
		s.SourceBinding = "clickhouse.otel-records/v1"
		s.Grain = map[string]string{"spans": "span-record", "logs": "log-record"}[v.Signal]
		s.Environment = v.Environment
		s.FromMS = v.FromMS
		s.ToMS = v.ToMS
		s.Selector = detached(v.Where)
		s.ReportedEdge = detached(v.ReportedEdge)
	case "measurement":
		v := q.Measurement
		d := v.Definition
		s.Signal = "metrics"
		s.MeasurementKind = d.Kind
		s.SourceBinding = "clickhouse.otel-measurements/v1"
		s.Grain = d.Grain
		s.Environment = d.Environment
		s.FromMS = v.FromMS
		s.ToMS = v.ToMS
		binding := detached(d.Source)
		s.Binding = &binding
		s.DefinitionRevision = measurements.Hash(d)
		s.SchemaMetadata = "Exact declared scope schema URL (including empty); resource schema URL remains per-stream evidence"
		if d.Kind == "lifecycle" {
			s.Signal = "logs"
			s.Grain = "scoped-subject"
			s.SchemaMetadata = "No convention/schema profile bound; lifecycle identity is application-declared"
		}
	case "comparison":
		a, e := (CaptureQuestion{Kind: "explore", Explore: &q.Comparison.Baseline}).Scope()
		if e != nil {
			return s, e
		}
		b, e := (CaptureQuestion{Kind: "explore", Explore: &q.Comparison.Selection}).Scope()
		if e != nil {
			return s, e
		}
		s.Signal = a.Signal
		s.SourceBinding = a.SourceBinding
		s.Grain = "comparison-pair"
		s.Environment = a.Environment
		s.Members = []QuestionScope{a, b}
	}
	return s, nil
}

type ScopeChange struct {
	Field  string `json:"field"`
	Effect string `json:"effect"` // preserved or replaced; never infer subset containment
}
type QuestionTransform struct {
	MeasurePlan *RecordMeasurePlan `json:"measurePlan,omitempty"`
	Relation    string             `json:"relation"`
	Origin      QuestionScope      `json:"origin"`
	Target      QuestionScope      `json:"target"`
	Changes     []ScopeChange      `json:"changes"`
	Preserves   []string           `json:"preserves"`
	Replaces    []string           `json:"replaces"`
}

// DescribeTransform checks supported relationship claims against typed questions.
// Exact IDs are checked structurally here; capture-owned planners select their
// values from retained observations, never from caller-supplied correlation IDs.
func DescribeTransform(origin, target CaptureQuestion, relation string) (*QuestionTransform, error) {
	a, e := origin.Scope()
	if e != nil {
		return nil, e
	}
	b, e := target.Scope()
	if e != nil {
		return nil, e
	}
	if a.Family == "comparison" || b.Family == "comparison" {
		return nil, fmt.Errorf("comparison requires explicit member transforms")
	}
	if a.Environment != b.Environment || a.FromMS != b.FromMS || a.ToMS != b.ToMS {
		return nil, fmt.Errorf("relationship requires original environment and fixed window")
	}
	switch relation {
	case "same-population":
		if !reflect.DeepEqual(a, b) {
			return nil, fmt.Errorf("same population requires identical semantic scope")
		}
	case "same-question":
		if !reflect.DeepEqual(origin.Normalized(), target.Normalized()) {
			return nil, fmt.Errorf("same question changed")
		}
	case "exact-trace", "exact-span":
		if target.Explore == nil || origin.Explore == nil || target.Explore.ReportedEdge != nil {
			return nil, fmt.Errorf("exact ID relationship requires records")
		}
		p := target.Explore.Where
		trace := func(p *Predicate) bool {
			return p != nil && p.Numeric == nil && p.Field != nil && *p.Field == (Field{Kind: "column", Key: "trace_id"}) && p.Op == "eq" && len(p.All) == 0 && len(p.Any) == 0 && followupTrace.MatchString(p.Value) && p.Value != strings.Repeat("0", 32)
		}
		if relation == "exact-trace" {
			if !trace(p) {
				return nil, fmt.Errorf("exact trace requires one valid trace predicate")
			}
		} else {
			if target.Explore.Signal != "logs" || p == nil || p.Numeric != nil || p.Field != nil || p.Op != "" || p.Value != "" || len(p.Any) != 0 || len(p.All) != 2 || !trace(&p.All[0]) {
				return nil, fmt.Errorf("exact span requires trace and span predicates")
			}
			v := p.All[1]
			if v.Numeric != nil || v.Field == nil || *v.Field != (Field{Kind: "column", Key: "span_id"}) || v.Op != "eq" || len(v.All) != 0 || len(v.Any) != 0 || !followupSpan.MatchString(v.Value) || v.Value == strings.Repeat("0", 16) {
				return nil, fmt.Errorf("invalid exact span predicate")
			}
		}
	default:
		return nil, fmt.Errorf("unsupported relationship")
	}
	return describeScopeChanges(origin, target, a, b, relation), nil
}

func describeScopeChanges(origin, target CaptureQuestion, a, b QuestionScope, relation string) *QuestionTransform {
	out := &QuestionTransform{Relation: relation, Origin: a, Target: b, Changes: []ScopeChange{}, Preserves: []string{}, Replaces: []string{}}
	add := func(field, label string, same bool) {
		effect := "replaced"
		if same {
			effect = "preserved"
			out.Preserves = append(out.Preserves, label)
		} else {
			out.Replaces = append(out.Replaces, label)
		}
		out.Changes = append(out.Changes, ScopeChange{field, effect})
	}
	add("environment", "Original environment", a.Environment == b.Environment)
	add("window", "Original fixed question window", a.FromMS == b.FromMS && a.ToMS == b.ToMS)
	add("sourceBinding", "Source binding", a.SourceBinding == b.SourceBinding)
	add("signal", "Signal", a.Signal == b.Signal)
	add("grain", "Population grain", a.Grain == b.Grain)
	if a.MeasurementKind != "" || b.MeasurementKind != "" {
		add("measurementKind", "Measurement operator", a.MeasurementKind == b.MeasurementKind)
	}
	if a.Family == "explore" || b.Family == "explore" {
		label := "Complete original record predicate"
		sameSelector := reflect.DeepEqual(a.Selector, b.Selector) && a.Family == "explore" && b.Family == "explore"
		if !sameSelector && (relation == "exact-trace" || relation == "exact-span") {
			label += " (replaced by exact reported IDs; allow all services)"
		}
		add("selector", label, sameSelector)
	}
	if a.ReportedEdge != nil || b.ReportedEdge != nil {
		add("reportedEdge", "Reported parent population", reflect.DeepEqual(a.ReportedEdge, b.ReportedEdge))
	}
	if a.Binding != nil || b.Binding != nil {
		add("binding", "Declared measurement binding", reflect.DeepEqual(a.Binding, b.Binding))
	}
	if a.DefinitionRevision != "" || b.DefinitionRevision != "" {
		add("definitionRevision", "Measurement definition", a.DefinitionRevision == b.DefinitionRevision)
	}
	if origin.Explore != nil && target.Explore != nil {
		x, y := origin.Explore, target.Explore
		measures, _ := y.RecordMeasures()
		out.MeasurePlan = &measures
		add("groupTimeSeries", "Split time chart by breakdown", x.GroupTimeSeries == y.GroupTimeSeries)
		names := []string{}
		for _, f := range y.GroupBy {
			names = append(names, f.Kind+"."+f.Key)
		}
		grouping := strings.Join(names, ", ")
		if grouping == "" {
			grouping = "none"
		}
		add("numeric", "Numeric binding and parsing policy", reflect.DeepEqual(origin.Explore.Numeric, target.Explore.Numeric))
		if x.NumericHistogram != nil || y.NumericHistogram != nil {
			add("numericHistogram", "Numeric histogram binning", reflect.DeepEqual(x.NumericHistogram, y.NumericHistogram))
		}
		add("groupBy", "Grouping → "+grouping, reflect.DeepEqual(x.GroupBy, y.GroupBy))
		add("measures", fmt.Sprintf("Summary → %s; recorded-status summary → %t", y.Summary, y.SpanStatus), x.Summary == y.Summary && x.SpanStatus == y.SpanStatus)
		add("bucketMs", fmt.Sprintf("Time chart interval → %d ms (0 disables chart)", y.BucketMS), x.BucketMS == y.BucketMS)
		add("inspectionPage", fmt.Sprintf("Inspection page → offset %d / limit %d", y.Offset, y.Limit), x.Limit == y.Limit && x.Offset == y.Offset)
	}
	return out
}

// DescribeDiscoveryTransform reports service/time context, never record-selector
// equivalence. The capture planner independently selects the service from evidence.
func DescribeDiscoveryTransform(origin CaptureQuestion, d measurements.DiscoveryQuestion) (*QuestionTransform, error) {
	a, e := origin.Scope()
	if e != nil {
		return nil, e
	}
	if e = d.Validate(); e != nil {
		return nil, e
	}
	if a.Family == "comparison" || a.Environment != d.Environment || a.FromMS != d.FromMS || a.ToMS != d.ToMS {
		return nil, fmt.Errorf("discovery must retain original environment/window")
	}
	b := QuestionScope{Family: "discovery", Signal: "metrics", SourceBinding: "clickhouse.otel-metric-discovery/v1", Grain: "observed-instrument-metadata", Environment: d.Environment, FromMS: d.FromMS, ToMS: d.ToMS, Binding: &measurements.Binding{Services: []string{d.Service}}, Convention: "Unbound; no semantic-convention profile inferred", SchemaMetadata: "Observed catalog metadata only; no completeness guarantee"}
	return describeScopeChanges(origin, CaptureQuestion{}, a, b, "service-time"), nil
}

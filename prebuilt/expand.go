package prebuilt

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/casselc/goo/measurement"
	"gopkg.in/yaml.v3"
)

type StatusRange struct {
	Min int `yaml:"min" json:"min"`
	Max int `yaml:"max" json:"max"`
}
type HTTPPolicy struct {
	FailureStatusRanges []StatusRange `yaml:"failureStatusRanges" json:"failureStatusRanges"`
	MissingStatus       string        `yaml:"missingStatus" json:"missingStatus"`
	MissingRoute        string        `yaml:"missingRoute" json:"missingRoute"`
}
type Binding struct {
	APIVersion     string            `yaml:"apiVersion" json:"apiVersion"`
	Name           string            `yaml:"name" json:"name"`
	Profile        string            `yaml:"profile" json:"profile"`
	Scope          string            `yaml:"scope" json:"scope"`
	PreparedTables map[string]string `yaml:"preparedTables" json:"preparedTables"`
	Policy         HTTPPolicy        `yaml:"policy" json:"policy"`
}
type View struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"`
	Measurement string   `json:"measurement"`
	GroupBy     []string `json:"groupBy"`
	Measures    []string `json:"measures"`
}
type Step struct {
	Question string `json:"question"`
	View     string `json:"view"`
}
type Diagnostic struct {
	Capability string `json:"capability"`
	Status     string `json:"status"`
	Reason     string `json:"reason"`
}
type Expansion struct {
	APIVersion        string                       `json:"apiVersion"`
	Revision          string                       `json:"revision"`
	ProfileRevision   string                       `json:"profileRevision"`
	InventoryRevision string                       `json:"inventoryRevision"`
	InventoryBasis    string                       `json:"inventoryBasis"`
	Binding           Binding                      `json:"binding"`
	Resource          map[string]string            `json:"resource"`
	Measurement       *measurement.Definition      `json:"measurement"`
	Views             []View                       `json:"views"`
	Recipe            []Step                       `json:"recipe"`
	Plans             map[string]*measurement.Plan `json:"plans"`
	Diagnostics       []Diagnostic                 `json:"diagnostics"`
}

func digest(value any) string {
	b, _ := json.Marshal(value)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func ReadBinding(r io.Reader) (Binding, error) {
	var b Binding
	d := yaml.NewDecoder(r)
	d.KnownFields(true)
	if err := d.Decode(&b); err != nil {
		return b, err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return b, fmt.Errorf("binding requires one YAML document")
	}
	return b, b.validate()
}
func (b Binding) validate() error {
	if b.APIVersion != "goo/binding/v1alpha1" || b.Name == "" || b.Scope == "" {
		return fmt.Errorf("binding requires version, name and scope")
	}
	if b.Profile != "otel.http.server.spans" {
		return fmt.Errorf("expansion is implemented only for otel.http.server.spans")
	}
	if len(b.PreparedTables) == 0 {
		return fmt.Errorf("explicit prepared-table bindings are required")
	}
	if b.Policy.MissingStatus != "unknown" || b.Policy.MissingRoute != "unknown-bucket" {
		return fmt.Errorf("explicit missingStatus: unknown and missingRoute: unknown-bucket are required")
	}
	if len(b.Policy.FailureStatusRanges) == 0 {
		return fmt.Errorf("explicit failure status ranges are required")
	}
	seen := map[int]bool{}
	for _, r := range b.Policy.FailureStatusRanges {
		if r.Min < 200 || r.Max > 599 || r.Min > r.Max {
			return fmt.Errorf("failure ranges must be within final response statuses 200..599")
		}
		for code := r.Min; code <= r.Max; code++ {
			if seen[code] {
				return fmt.Errorf("overlapping failure ranges")
			}
			seen[code] = true
		}
	}
	return nil
}

// Expand emits reviewable artifacts over a declared prepared source. Recognition
// is a prerequisite, never an assertion that raw telemetry has been prepared.
func Expand(in Inventory, b Binding) (*Expansion, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	proposals, err := Discover(in)
	if err != nil {
		return nil, err
	}
	compatible := false
	for _, p := range proposals {
		if p.Profile == b.Profile && p.Scope == b.Scope && p.Status == "compatible-shape" {
			compatible = true
		}
	}
	if !compatible {
		return nil, fmt.Errorf("selected profile/scope does not have a compatible contract")
	}
	var scope Scope
	for _, s := range in.Scopes {
		if s.ID == b.Scope {
			scope = s
		}
	}
	var profile Profile
	for _, p := range Catalog().Profiles {
		if p.ID == b.Profile {
			profile = p
		}
	}
	b.Policy.FailureStatusRanges = append([]StatusRange(nil), b.Policy.FailureStatusRanges...)
	// Canonical policy order makes equivalent authoring order revision-neutral.
	sort.Slice(b.Policy.FailureStatusRanges, func(i, j int) bool { return b.Policy.FailureStatusRanges[i].Min < b.Policy.FailureStatusRanges[j].Min })
	d := &measurement.Definition{APIVersion: "goo/v1alpha1", Name: b.Name,
		Source:     measurement.Source{Kind: "subject-facts", Fields: map[string]string{"evaluation_id": "string", "scope_id": "string", "trace_id": "string", "span_id": "string", "route": "string", "method": "string", "classification": "string"}, Enums: map[string][]string{"classification": {"success", "failure", "unknown"}}, SubjectKey: []string{"scope_id", "trace_id", "span_id"}, Bindings: b.PreparedTables},
		Population: &measurement.Predicate{Field: "scope_id", Eq: b.Scope}, Dimensions: map[string]string{"route": "route", "method": "method"},
		Measures: map[string]measurement.Measure{"requests": {Op: "count"}, "successful": {Op: "count", Where: &measurement.Predicate{Field: "classification", Eq: "success"}}, "failures": {Op: "count", Where: &measurement.Predicate{Field: "classification", Eq: "failure"}}, "unknown": {Op: "count", Where: &measurement.Predicate{Field: "classification", Eq: "unknown"}}, "success_ratio": {Op: "ratio", Numerator: "successful", Denominator: "requests", UnknownWhen: []string{"unknown"}}}}
	if err := d.Validate(); err != nil {
		return nil, err
	}
	out := &Expansion{APIVersion: "goo/expansion/v1alpha1", ProfileRevision: digest(profile), InventoryBasis: in.Basis, Binding: b, Resource: scope.Resource, Measurement: d,
		Views: []View{{ID: "route-overview", Kind: "aggregate-table", Measurement: b.Name, GroupBy: []string{"route", "method"}, Measures: []string{"requests", "successful", "failures", "unknown", "success_ratio"}}}, Recipe: []Step{{Question: "Which routes have observed failures or unknown outcomes?", View: "route-overview"}}, Plans: map[string]*measurement.Plan{}, Diagnostics: []Diagnostic{{Capability: "raw-span-preparation", Status: "required", Reason: "A source adapter must enforce resource bindings, fixed evaluation selection, scoped span identity and the bound classification policy before these aggregate plans execute."}, {Capability: "duration-heatmap-and-waterfall", Status: "unsupported", Reason: "Duration distributions and relation lookup are not operators in the current aggregate IR."}, {Capability: "objective-and-alert", Status: "not-configured", Reason: "No target, objective window or notification policy is inferred."}}}
	// Only the selected scope/contract contributes to the inventory revision.
	normalized := scope
	normalized.Signals = nil
	for _, signal := range scope.Signals {
		for _, contract := range profile.Requirements[0].AnyOf {
			if matches(signal, contract) {
				normalized.Signals = append(normalized.Signals, contract)
				break
			}
		}
	}
	normalized.Signals = normalized.Signals[:1] // compatible contract checked above; duplicate signatures do not change identity
	out.InventoryRevision = digest(struct {
		Basis string
		Scope Scope
	}{in.Basis, normalized})
	for _, target := range []string{"clickhouse", "databricks"} {
		if b.PreparedTables[target] == "" {
			out.Diagnostics = append(out.Diagnostics, Diagnostic{Capability: target, Status: "unbound", Reason: "No prepared table binding supplied."})
			continue
		}
		plan, e := d.Compile(target, []string{"route", "method"})
		if e != nil {
			return nil, e
		}
		out.Plans[target] = plan
	}
	out.Revision = digest(out)
	return out, nil
}

package exploration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"slices"

	"github.com/casselc/goo/identity"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const PackageVersion = "goo/telemetry-package/v1alpha1"
const PackageBytes = 32 << 10

var packageName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,127}$`)

// TelemetryPackage binds existing OTel data without a source-code manifest.
// A definition is reusable configuration, never a frozen evidence receipt.
type TelemetryPackage struct {
	NumericHistogram *NumericHistogramSpec `json:"numericHistogram,omitempty" yaml:"numericHistogram,omitempty"`
	ReportedEdge     *ReportedEdge         `json:"reportedEdge,omitempty" yaml:"reportedEdge,omitempty"`
	Numeric          *NumericBinding       `json:"numeric,omitempty" yaml:"numeric,omitempty"`
	GroupTimeSeries  bool                  `json:"groupTimeSeries,omitempty" yaml:"groupTimeSeries,omitempty"`
	LogSeverity      bool                  `json:"logSeverity,omitempty" yaml:"logSeverity,omitempty"`
	SpanStatus       bool                  `json:"spanStatus,omitempty" yaml:"spanStatus,omitempty"`
	BucketMS         int64                 `json:"bucketMs,omitempty" yaml:"bucketMs,omitempty"`
	Summary          string                `json:"summary,omitempty" yaml:"summary,omitempty"`
	APIVersion       string                `json:"apiVersion" yaml:"apiVersion"`
	Name             string                `json:"name" yaml:"name"`
	Title            string                `json:"title" yaml:"title"`
	Description      string                `json:"description,omitempty" yaml:"description,omitempty"`
	Source           PackageSource         `json:"source" yaml:"source"`
	Where            *Predicate            `json:"where,omitempty" yaml:"where,omitempty"`
	GroupBy          []Field               `json:"groupBy,omitempty" yaml:"groupBy,omitempty"`
	Measure          string                `json:"measure" yaml:"measure"`
	LookbackMinutes  int                   `json:"lookbackMinutes" yaml:"lookbackMinutes"`
	Views            []string              `json:"views" yaml:"views"`
}

// UnmarshalJSON rejects lossy Unicode before definition fields can be repaired
// into another predicate, profile or revision. Nested source admission stays strict.
func (p *TelemetryPackage) UnmarshalJSON(data []byte) error {
	if !identity.WellFormedJSON(data) {
		return fmt.Errorf("invalid telemetry package Unicode")
	}
	type wire TelemetryPackage
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*p = TelemetryPackage(value)
	return nil
}

type PackageSource struct {
	Adapter             string                        `json:"adapter" yaml:"adapter"`
	Signal              string                        `json:"signal" yaml:"signal"`
	Environment         string                        `json:"environment,omitempty" yaml:"environment,omitempty"`
	EnvironmentSelector *identity.EnvironmentSelector `json:"environmentSelector,omitempty" yaml:"environmentSelector,omitempty"`
}
type AnnotationTarget struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
type InstrumentationHandoff struct {
	Fields    []Field  `json:"fields"`
	GoExample string   `json:"goExample,omitempty"`
	Notes     []string `json:"notes"`
}
type PackagePlan struct {
	MeasurePlan RecordMeasurePlan      `json:"measurePlan"`
	Definition  TelemetryPackage       `json:"definition"`
	Revision    string                 `json:"revision"`
	YAML        string                 `json:"yaml"`
	Question    Query                  `json:"question"`
	Plan        Plan                   `json:"plan"`
	Handoff     InstrumentationHandoff `json:"handoff"`
	Warnings    []string               `json:"warnings"`
}

// PackageSource uses the same checked environment meanings as record questions.
// Wire presence is exclusive: an explicit empty scalar alongside a selector is
// not accepted as a second spelling of the selected scope.
func (s *PackageSource) UnmarshalJSON(data []byte) error {
	if !identity.WellFormedJSON(data) {
		return fmt.Errorf("invalid package source Unicode")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for key := range fields {
		if key != "adapter" && key != "signal" && key != "environment" && key != "environmentSelector" {
			return fmt.Errorf("unknown package source field %q", key)
		}
	}
	_, scalar := fields["environment"]
	_, typed := fields["environmentSelector"]
	if scalar == typed {
		return fmt.Errorf("package source requires exactly one environment or environmentSelector")
	}
	for _, key := range []string{"environment", "environmentSelector"} {
		if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("%s must be omitted or non-null", key)
		}
	}
	type wire PackageSource
	var value wire
	normalized, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(normalized))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	candidate := PackageSource(value)
	if _, err := (Query{Environment: candidate.Environment, EnvironmentSelector: candidate.EnvironmentSelector}).EnvironmentSelection(); err != nil {
		return err
	}
	*s = candidate
	return nil
}
func validatePackageYAMLNode(n *yaml.Node, depth int) error {
	if depth > 16 || n.Kind == yaml.AliasNode || n.Tag == "!!merge" {
		return fmt.Errorf("YAML aliases, merges and excessive depth unsupported")
	}
	for _, child := range n.Content {
		if err := validatePackageYAMLNode(child, depth+1); err != nil {
			return err
		}
	}
	return nil
}
func (s *PackageSource) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("package source requires a mapping")
	}
	if err := validatePackageYAMLNode(n, 0); err != nil {
		return err
	}
	// Binary/custom scalar tags could decode to invalid Go strings which
	// json.Marshal would repair before the shared selector codec sees them.
	var tags func(*yaml.Node) error
	tags = func(node *yaml.Node) error {
		if node.Kind == yaml.ScalarNode && node.Tag != "!!str" && node.Tag != "!!null" && node.Tag != "!!bool" && node.Tag != "!!int" && node.Tag != "!!float" {
			return fmt.Errorf("unsupported package source YAML scalar tag")
		}
		for _, child := range node.Content {
			if err := tags(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := tags(n); err != nil {
		return err
	}
	var fields map[string]any
	if err := n.Decode(&fields); err != nil {
		return err
	}
	data, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, s)
}
func (p TelemetryPackage) rawQuestion(from, to int64) Query {
	return Query{EnvironmentSelector: p.Source.EnvironmentSelector, NumericHistogram: p.NumericHistogram, ReportedEdge: p.ReportedEdge, Numeric: p.Numeric, GroupTimeSeries: p.GroupTimeSeries, SpanStatus: p.SpanStatus, LogSeverity: p.LogSeverity, BucketMS: p.BucketMS, Summary: p.Summary, Signal: p.Source.Signal, Environment: p.Source.Environment, FromMS: from, ToMS: to, Where: p.Where, GroupBy: p.GroupBy, Limit: 100, Offset: 0}
}

// Question detaches every mutable definition field without a JSON round trip
// that could repair an invalid caller string into a different identity.
func (p TelemetryPackage) Question(from, to int64) Query {
	q := p.rawQuestion(from, to)
	q.EnvironmentSelector = copyEnvironmentSelector(q.EnvironmentSelector)
	q.GroupBy = slices.Clone(q.GroupBy)
	if q.Numeric != nil {
		q.Numeric = new(*q.Numeric)
	}
	if q.NumericHistogram != nil {
		q.NumericHistogram = new(*q.NumericHistogram)
	}
	if q.ReportedEdge != nil {
		q.ReportedEdge = new(*q.ReportedEdge)
		q.ReportedEdge.Parent.Namespace = q.ReportedEdge.Parent.Namespace.Clone()
		q.ReportedEdge.Child.Namespace = q.ReportedEdge.Child.Namespace.Clone()
		q.ReportedEdge.DiscoveryNamespace = q.ReportedEdge.DiscoveryNamespace.Clone()
	}
	seen := map[*Predicate]*Predicate{}
	var clone func(*Predicate) *Predicate
	clone = func(in *Predicate) *Predicate {
		if in == nil {
			return nil
		}
		if out, ok := seen[in]; ok {
			return out
		}
		out := new(*in)
		seen[in] = out
		if in.Field != nil {
			out.Field = new(*in.Field)
		}
		if in.Numeric != nil {
			out.Numeric = new(*in.Numeric)
		}
		out.All = slices.Clone(in.All)
		out.Any = slices.Clone(in.Any)
		for i := range in.All {
			out.All[i] = *clone(&in.All[i])
		}
		for i := range in.Any {
			out.Any[i] = *clone(&in.Any[i])
		}
		return out
	}
	q.Where = clone(q.Where)
	return q
}
func (p TelemetryPackage) Validate() error {
	if p.APIVersion != PackageVersion || !packageName.MatchString(p.Name) || p.Title == "" || len(p.Title) > 160 || !utf8.ValidString(p.Title) || strings.ContainsRune(p.Title, 0) || len(p.Description) > 2048 || !utf8.ValidString(p.Description) || strings.ContainsRune(p.Description, 0) {
		return fmt.Errorf("require versioned package, stable name and bounded UTF-8 title/description")
	}
	if p.Source.Adapter != "clickhouse.otel/v1" || p.Measure != "captured-record-count" || p.LookbackMinutes < 1 || p.LookbackMinutes > 120 {
		return fmt.Errorf("require supported OTel adapter, captured-record-count and lookback 1..120 minutes")
	}
	if e := p.rawQuestion(0, int64(p.LookbackMinutes)*60000).Validate(); e != nil {
		return e
	}
	if len(p.Views) == 0 || len(p.Views) > 3 {
		return fmt.Errorf("choose 1..3 supported views")
	}
	seen := map[string]bool{}
	for _, v := range p.Views {
		if seen[v] || v != "records" && v != "distribution" && v != "trace-context" {
			return fmt.Errorf("unsupported or duplicate view %q", v)
		}
		if v == "distribution" && len(p.GroupBy) != 1 {
			return fmt.Errorf("distribution requires exactly one breakdown")
		}
		seen[v] = true
	}
	return nil
}
func ReadTelemetryPackage(text string) (TelemetryPackage, error) {
	var p TelemetryPackage
	if len(text) > PackageBytes {
		return p, fmt.Errorf("package exceeds 32 KiB")
	}
	var root yaml.Node
	if e := yaml.Unmarshal([]byte(text), &root); e != nil {
		return p, e
	}
	if e := validatePackageYAMLNode(&root, 0); e != nil {
		return p, e
	}

	d := yaml.NewDecoder(strings.NewReader(text))
	d.KnownFields(true)
	if e := d.Decode(&p); e != nil {
		return p, e
	}
	if d.Decode(new(any)) != io.EOF {
		return p, fmt.Errorf("require one YAML document")
	}
	return p, p.Validate()
}

// snapshotTelemetryPackage validates before encoding, then freezes one exact
// checked definition for hashing, publication and independent plan outputs.
func snapshotTelemetryPackage(p TelemetryPackage) (TelemetryPackage, []byte, error) {
	if err := p.Validate(); err != nil {
		return TelemetryPackage{}, nil, err
	}
	data, err := json.Marshal(p)
	if err != nil {
		return TelemetryPackage{}, nil, err
	}
	frozen, err := DecodePackageJSON(data)
	return frozen, data, err
}
func packageRevision(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func (p TelemetryPackage) Revision() (string, error) {
	_, data, err := snapshotTelemetryPackage(p)
	if err != nil {
		return "", err
	}
	return packageRevision(data), nil
}
func DraftTelemetryPackage(q Query, name, title string, minutes int) (TelemetryPackage, error) {
	if e := q.Validate(); e != nil {
		return TelemetryPackage{}, e
	}
	// Clone to isolate the reusable definition from mutable caller questions/pages.
	b, _ := json.Marshal(q)
	var clone Query
	json.Unmarshal(b, &clone)
	p := TelemetryPackage{NumericHistogram: clone.NumericHistogram, ReportedEdge: clone.ReportedEdge, Numeric: clone.Numeric, GroupTimeSeries: q.GroupTimeSeries, SpanStatus: q.SpanStatus, LogSeverity: q.LogSeverity, BucketMS: q.BucketMS, Summary: q.Summary, APIVersion: PackageVersion, Name: name, Title: title, Source: PackageSource{Adapter: "clickhouse.otel/v1", Signal: clone.Signal, Environment: clone.Environment, EnvironmentSelector: clone.EnvironmentSelector}, Where: clone.Where, GroupBy: clone.GroupBy, Measure: "captured-record-count", LookbackMinutes: minutes, Views: []string{"records", "trace-context"}}
	if len(p.GroupBy) == 1 {
		p.Views = []string{"records", "distribution", "trace-context"}
	}
	return p, p.Validate()
}
func PlanTelemetryPackage(p TelemetryPackage, from, to int64, target *AnnotationTarget) (*PackagePlan, error) {
	frozen, encoded, e := snapshotTelemetryPackage(p)
	if e != nil {
		return nil, e
	}
	p = frozen
	revision := packageRevision(encoded)
	q := p.Question(from, to)
	plan, e := Compile(q)
	if e != nil {
		return nil, e
	}
	text, e := yaml.Marshal(p)
	if e != nil {
		return nil, e
	}
	if len(text) > PackageBytes {
		return nil, fmt.Errorf("encoded package exceeds 32 KiB")
	}
	handoff := InstrumentationHandoff{Fields: []Field{{Kind: "column", Key: "service"}}, Notes: []string{"Use the package directly over existing OTel telemetry in any language; goo annotations are optional.", "Fields below are referenced by the definition, not proof that every selected record supplies them. Preserve missing/empty distinctions and predicate meaning.", "No source function, event identity, outcome policy, complete coverage, SLI or native metric is inferred from trace samples."}}
	seen := map[Field]bool{handoff.Fields[0]: true}
	add := func(f Field) {
		if !seen[f] {
			seen[f] = true
			handoff.Fields = append(handoff.Fields, f)
		}
	}
	var walk func(*Predicate)
	walk = func(pred *Predicate) {
		if pred == nil {
			return
		}
		if pred.Field != nil {
			add(*pred.Field)
		}
		if pred.Numeric != nil {
			add(pred.Numeric.Field)
		}
		for i := range pred.All {
			walk(&pred.All[i])
		}
		for i := range pred.Any {
			walk(&pred.Any[i])
		}
	}
	walk(p.Where)
	if p.Numeric != nil {
		add(p.Numeric.Field)
	}
	for _, f := range p.GroupBy {
		add(f)
	}
	for _, f := range handoff.Fields {
		if _, _, ok := exceptionFieldSQL(f.Key); f.Kind == "column" && ok {
			handoff.Notes = append(handoff.Notes, "Exception fields derive from span events named exception and their event attributes, not span attributes named after these derived columns. Recorded exception events do not establish failed status or uncaught behavior.")
			break
		}
	}
	if seen[Field{Kind: "column", Key: "severity_number"}] {
		handoff.Notes = append(handoff.Notes, "severity_number is intrinsic OTel log SeverityNumber, not an attribute numeric binding. Preserve the independent raw SeverityText; do not infer numeric severity from its spelling.")
	}
	if p.ReportedEdge != nil {
		add(Field{Kind: "column", Key: "trace_id"})
		add(Field{Kind: "resource", Key: "service.namespace"})
		handoff.Notes = append(handoff.Notes, "Reported-edge queries require standard trace/span/parent-span IDs and original service resource identity. Parent ownership is resolved across records; no single-function annotation expresses this relationship.")
	}
	if target != nil {
		selection, err := q.EnvironmentSelection()
		if err != nil {
			return nil, err
		}
		if !selection.CurrentOnly() || selection.Scope().Mode != "named" {
			return nil, fmt.Errorf("Go annotation templates require a current-key-only named environment; use this telemetry package directly without an annotation target")
		}

		if p.ReportedEdge != nil {
			return nil, fmt.Errorf("reported-edge relationships cannot be generated as a single-function annotation; retain the reusable telemetry package")
		}
		if !packageName.MatchString(target.Name) || !map[string]bool{"internal": true, "server": true, "client": true, "producer": true, "consumer": true}[target.Role] {
			return nil, fmt.Errorf("annotation requires stable operation name and supported role")
		}
		handoff.GoExample = fmt.Sprintf("package example\n\nimport \"context\"\n\n//goo:operation name:%s role:%s\nfunc ObservedOperation(ctx context.Context) (err error) {\n\t// Implement the operation and explicitly attach required attributes.\n\treturn nil\n}\n", target.Name, target.Role)
		handoff.Notes = append(handoff.Notes, "Go example is an opt-in template, not a patch or equivalent source binding. It does not encode the package environment, namespace or predicates. Choose the real function, preserve attributes, validate/weave with goo, and review duplicate instrumentation before changing package selectors. Existing telemetry is unchanged.")
	}
	measures, e := q.RecordMeasures()
	if e != nil {
		return nil, e
	}
	return &PackagePlan{MeasurePlan: measures, Definition: p, Revision: revision, YAML: string(text), Question: q, Plan: plan, Handoff: handoff, Warnings: []string{"Definition revision identifies configuration, not captured evidence. Staging it reads no telemetry; explicit Run performs a fresh live read.", "Counts are captured records, including retries and exporter duplicates; sampling/completeness remain unknown. No business outcome or SLO is inferred.", "Supported record measures include counts, span durations, recorded status, explicit numeric bindings and bounded numeric histograms and split time charts. Native metrics, lifecycle rules, arbitrary numeric expressions and Databricks lowering are separate contracts."}}, nil
}

// DecodePackageJSON is used for stored definitions with strict field validation.
func DecodePackageJSON(data []byte) (TelemetryPackage, error) {
	var p TelemetryPackage
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e := d.Decode(&p); e != nil {
		return p, e
	}
	if d.Decode(new(any)) != io.EOF {
		return p, fmt.Errorf("require one definition")
	}
	return p, p.Validate()
}

package exploration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
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
	GroupTimeSeries bool          `json:"groupTimeSeries,omitempty" yaml:"groupTimeSeries,omitempty"`
	SpanStatus      bool          `json:"spanStatus,omitempty" yaml:"spanStatus,omitempty"`
	BucketMS        int64         `json:"bucketMs,omitempty" yaml:"bucketMs,omitempty"`
	Summary         string        `json:"summary,omitempty" yaml:"summary,omitempty"`
	APIVersion      string        `json:"apiVersion" yaml:"apiVersion"`
	Name            string        `json:"name" yaml:"name"`
	Title           string        `json:"title" yaml:"title"`
	Description     string        `json:"description,omitempty" yaml:"description,omitempty"`
	Source          PackageSource `json:"source" yaml:"source"`
	Where           *Predicate    `json:"where,omitempty" yaml:"where,omitempty"`
	GroupBy         []Field       `json:"groupBy,omitempty" yaml:"groupBy,omitempty"`
	Measure         string        `json:"measure" yaml:"measure"`
	LookbackMinutes int           `json:"lookbackMinutes" yaml:"lookbackMinutes"`
	Views           []string      `json:"views" yaml:"views"`
}
type PackageSource struct {
	Adapter     string `json:"adapter" yaml:"adapter"`
	Signal      string `json:"signal" yaml:"signal"`
	Environment string `json:"environment" yaml:"environment"`
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

func (p TelemetryPackage) Question(from, to int64) Query {
	return Query{GroupTimeSeries: p.GroupTimeSeries, SpanStatus: p.SpanStatus, BucketMS: p.BucketMS, Summary: p.Summary, Signal: p.Source.Signal, Environment: p.Source.Environment, FromMS: from, ToMS: to, Where: p.Where, GroupBy: p.GroupBy, Limit: 100, Offset: 0}
}
func (p TelemetryPackage) Validate() error {
	if p.APIVersion != PackageVersion || !packageName.MatchString(p.Name) || p.Title == "" || len(p.Title) > 160 || !utf8.ValidString(p.Title) || strings.ContainsRune(p.Title, 0) || len(p.Description) > 2048 || !utf8.ValidString(p.Description) || strings.ContainsRune(p.Description, 0) {
		return fmt.Errorf("require versioned package, stable name and bounded UTF-8 title/description")
	}
	if p.Source.Adapter != "clickhouse.otel/v1" || p.Measure != "captured-record-count" || p.LookbackMinutes < 1 || p.LookbackMinutes > 120 {
		return fmt.Errorf("require supported OTel adapter, captured-record-count and lookback 1..120 minutes")
	}
	if e := p.Question(0, int64(p.LookbackMinutes)*60000).Validate(); e != nil {
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
	var check func(*yaml.Node, int) error
	check = func(n *yaml.Node, depth int) error {
		if depth > 16 || n.Kind == yaml.AliasNode || n.Tag == "!!merge" {
			return fmt.Errorf("YAML aliases, merges and excessive depth unsupported")
		}
		for _, child := range n.Content {
			if e := check(child, depth+1); e != nil {
				return e
			}
		}
		return nil
	}
	if e := check(&root, 0); e != nil {
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
func (p TelemetryPackage) Revision() (string, error) {
	if e := p.Validate(); e != nil {
		return "", e
	}
	data, e := json.Marshal(p)
	if e != nil {
		return "", e
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
func DraftTelemetryPackage(q Query, name, title string, minutes int) (TelemetryPackage, error) {
	if e := q.Validate(); e != nil {
		return TelemetryPackage{}, e
	}
	// Clone to isolate the reusable definition from mutable caller questions/pages.
	b, _ := json.Marshal(q)
	var clone Query
	json.Unmarshal(b, &clone)
	p := TelemetryPackage{GroupTimeSeries: q.GroupTimeSeries, SpanStatus: q.SpanStatus, BucketMS: q.BucketMS, Summary: q.Summary, APIVersion: PackageVersion, Name: name, Title: title, Source: PackageSource{"clickhouse.otel/v1", q.Signal, q.Environment}, Where: clone.Where, GroupBy: clone.GroupBy, Measure: "captured-record-count", LookbackMinutes: minutes, Views: []string{"records", "trace-context"}}
	if len(p.GroupBy) == 1 {
		p.Views = []string{"records", "distribution", "trace-context"}
	}
	return p, p.Validate()
}
func PlanTelemetryPackage(p TelemetryPackage, from, to int64, target *AnnotationTarget) (*PackagePlan, error) {
	revision, e := p.Revision()
	if e != nil {
		return nil, e
	}
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
		for i := range pred.All {
			walk(&pred.All[i])
		}
		for i := range pred.Any {
			walk(&pred.Any[i])
		}
	}
	walk(p.Where)
	for _, f := range p.GroupBy {
		add(f)
	}
	if target != nil {
		if !packageName.MatchString(target.Name) || !map[string]bool{"internal": true, "server": true, "client": true, "producer": true, "consumer": true}[target.Role] {
			return nil, fmt.Errorf("annotation requires stable operation name and supported role")
		}
		handoff.GoExample = fmt.Sprintf("package example\n\nimport \"context\"\n\n//goo:operation name:%s role:%s\nfunc ObservedOperation(ctx context.Context) (err error) {\n\t// Implement the operation and explicitly attach required attributes.\n\treturn nil\n}\n", target.Name, target.Role)
		handoff.Notes = append(handoff.Notes, "Go example is an opt-in template, not a patch or equivalent source binding. Choose the real function, preserve attributes, validate/weave with goo, and review duplicate instrumentation before changing package selectors. Existing telemetry is unchanged.")
	}
	measures, e := q.RecordMeasures()
	if e != nil {
		return nil, e
	}
	return &PackagePlan{MeasurePlan: measures, Definition: p, Revision: revision, YAML: string(text), Question: q, Plan: plan, Handoff: handoff, Warnings: []string{"Definition revision identifies configuration, not captured evidence. Applying it performs a fresh live read.", "Counts are captured records, including retries and exporter duplicates; sampling/completeness remain unknown. No business outcome or SLO is inferred.", "Supported record measures include counts, span durations, recorded status and bounded split time charts. Native metrics, lifecycle rules, arbitrary numeric expressions and Databricks lowering are separate contracts."}}, nil
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

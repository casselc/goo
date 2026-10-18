package annotations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/casselc/goo/metrics"
	"gopkg.in/yaml.v3"
)

// OperationPackage declares local observation bindings, not business outcome policy.
type OperationPackage struct {
	Fleet           *FleetContract      `yaml:"fleet,omitempty" json:"fleet,omitempty"`
	Histograms      []HistogramContract `yaml:"histograms,omitempty" json:"histograms,omitempty"`
	APIVersion      string              `yaml:"apiVersion" json:"apiVersion"`
	Name            string              `yaml:"name" json:"name"`
	Title           string              `yaml:"title" json:"title"`
	SourceDirectory string              `yaml:"sourceDirectory" json:"sourceDirectory"`
	Adapter         string              `yaml:"adapter" json:"adapter"`
	Resource        struct {
		Service     string `yaml:"service" json:"service"`
		Environment string `yaml:"environment" json:"environment"`
		Instance    string `yaml:"instance" json:"instance"`
	} `yaml:"resource" json:"resource"`
	Instrumentation struct {
		Name      string `yaml:"name" json:"name"`
		Version   string `yaml:"version" json:"version"`
		SchemaURL string `yaml:"schemaURL" json:"schemaURL"`
	} `yaml:"instrumentation" json:"instrumentation"`
	Ratios     []RatioContract     `yaml:"ratios,omitempty" json:"ratios,omitempty"`
	Rates      []RateContract      `yaml:"rates,omitempty" json:"rates,omitempty"`
	Operations []OperationContract `yaml:"operations" json:"operations"`
	Subjects   *SubjectContract    `yaml:"subjects,omitempty" json:"subjects,omitempty"`
	Views      []string            `yaml:"views" json:"views"`
	Evidence   struct {
		Sampling     string `yaml:"sampling" json:"sampling"`
		Completeness string `yaml:"completeness" json:"completeness"`
		Schema       string `yaml:"schema" json:"schema"`
	} `yaml:"evidence" json:"evidence"`
}
type OperationContract struct {
	Name string `yaml:"name" json:"name"`
	Role string `yaml:"role" json:"role"`
}
type SubjectContract struct {
	Namespace string   `yaml:"namespace" json:"namespace"`
	Events    []string `yaml:"events" json:"events"`
	Identity  []string `yaml:"identity" json:"identity"`
}
type OperationPlan struct {
	APIVersion  string           `json:"apiVersion"`
	Revision    string           `json:"revision"`
	Definition  OperationPackage `json:"definition"`
	Manifest    *Manifest        `json:"manifest"`
	Source      OperationSource  `json:"source"`
	Diagnostics []string         `json:"diagnostics"`
}

func ReadOperationPackage(reader io.Reader) (OperationPackage, error) {
	var p OperationPackage
	data, err := io.ReadAll(io.LimitReader(reader, (1<<20)+1))
	if err != nil {
		return p, err
	}
	if len(data) > 1<<20 {
		return p, fmt.Errorf("operation package exceeds 1 MiB")
	}
	var root yaml.Node
	if err = yaml.Unmarshal(data, &root); err != nil {
		return p, err
	}
	var check func(*yaml.Node) error
	check = func(n *yaml.Node) error {
		if n.Kind == yaml.AliasNode || n.Tag == "!!merge" {
			return fmt.Errorf("YAML aliases and merges are unsupported")
		}
		for _, c := range n.Content {
			if err := check(c); err != nil {
				return err
			}
		}
		return nil
	}
	if err = check(&root); err != nil {
		return p, err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err = dec.Decode(&p); err != nil {
		return p, err
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF {
		return p, fmt.Errorf("require one YAML package document")
	}
	return p, nil
}
func LoadOperationPackage(path string) (*OperationPlan, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	p, err := ReadOperationPackage(f)
	if err != nil {
		return nil, err
	}
	if p.SourceDirectory == "" || filepath.IsAbs(p.SourceDirectory) {
		return nil, fmt.Errorf("sourceDirectory must be a local relative directory")
	}
	m, err := ScanDir(filepath.Join(filepath.Dir(path), p.SourceDirectory))
	if err != nil {
		return nil, fmt.Errorf("source contract: %w", err)
	}
	return p.Bind(m)
}
func (p OperationPackage) Bind(m *Manifest) (*OperationPlan, error) {
	p.Operations = append([]OperationContract(nil), p.Operations...)
	p.Views = append([]string(nil), p.Views...)
	p.Rates = append([]RateContract(nil), p.Rates...)
	for i := range p.Rates {
		p.Rates[i].Outcomes = append([]string(nil), p.Rates[i].Outcomes...)
	}
	if p.Subjects != nil {
		copy := *p.Subjects
		copy.Events = append([]string(nil), copy.Events...)
		copy.Identity = append([]string(nil), copy.Identity...)
		p.Subjects = &copy
	}
	if p.SourceDirectory == "" || filepath.IsAbs(p.SourceDirectory) {
		return nil, fmt.Errorf("sourceDirectory must be a local relative directory")
	}

	if p.APIVersion != "goo/operation-package/v1alpha1" || !stableName.MatchString(p.Name) || p.Title == "" || len(p.Title) > 160 {
		return nil, fmt.Errorf("require versioned operation package with stable name and bounded title")
	}
	if p.Adapter != "clickhouse.otel-traces/v1" {
		return nil, fmt.Errorf("unsupported operation source adapter")
	}
	if m == nil || m.APIVersion != "goo/instrumentation-manifest/v1alpha1" || m.Revision == "" {
		return nil, fmt.Errorf("require source manifest")
	}
	if p.Instrumentation.Name != m.Scope || p.Instrumentation.Version != m.ScopeVersion || p.Instrumentation.SchemaURL != m.SchemaURL {
		return nil, fmt.Errorf("package instrumentation does not match source contract")
	}
	if p.Evidence.Sampling != "unverified" || p.Evidence.Completeness != "unverified" || p.Evidence.Schema != "declared" {
		return nil, fmt.Errorf("only unverified sampling/completeness and declared schema are supported")
	}
	source := OperationSource{Service: p.Resource.Service, Environment: p.Resource.Environment, Instance: p.Resource.Instance, DeclaredSchema: p.Instrumentation.SchemaURL}
	if _, _, err := m.OperationQuery(source, 0, 1); err != nil {
		return nil, err
	}
	expected := map[string]bool{}
	events := map[string]bool{}
	for _, d := range m.Declarations {
		if d.Kind == "operation" {
			expected[d.Name+"/"+d.Role] = true
		}
		if d.Kind == "event" {
			events[d.Name] = true
		}
	}
	if len(p.Operations) == 0 || len(p.Operations) != len(expected) {
		return nil, fmt.Errorf("operations must cover the selected source contract exactly")
	}
	for _, op := range p.Operations {
		k := op.Name + "/" + op.Role
		if !expected[k] {
			return nil, fmt.Errorf("unknown or duplicate operation %s", k)
		}
		delete(expected, k)
	}
	sort.Slice(p.Operations, func(i, j int) bool {
		return p.Operations[i].Name+"/"+p.Operations[i].Role < p.Operations[j].Name+"/"+p.Operations[j].Role
	})
	if p.Subjects != nil {
		if !stableName.MatchString(p.Subjects.Namespace) || len(p.Subjects.Events) == 0 || len(p.Subjects.Events) != len(events) {
			return nil, fmt.Errorf("subject namespace and exact declared event set required")
		}
		for _, name := range p.Subjects.Events {
			if !events[name] {
				return nil, fmt.Errorf("unknown or duplicate subject event %s", name)
			}
			delete(events, name)
		}
		sort.Strings(p.Subjects.Events)
		identity := []string{"environment", "service", "namespace", "tenant", "id"}
		if len(p.Subjects.Identity) != len(identity) {
			return nil, fmt.Errorf("full scoped subject identity required")
		}
		for i, key := range identity {
			if p.Subjects.Identity[i] != key {
				return nil, fmt.Errorf("identity must be environment, service, namespace, tenant, id")
			}
		}
		source.SubjectNamespace = p.Subjects.Namespace
	}
	p.Histograms = append([]HistogramContract(nil), p.Histograms...)
	histNames := map[string]bool{}
	for i, h := range p.Histograms {
		if !stableName.MatchString(h.Name) || histNames[h.Name] || h.Metric != "goo.operation.duration" || h.Unit != "s" || (h.Temporality != "cumulative" && h.Temporality != "delta") {
			return nil, fmt.Errorf("unsupported histogram contract")
		}
		histNames[h.Name] = true
		if e := metrics.ValidateDurationBounds(h.Bounds); e != nil {
			return nil, e
		}
		gap, e := time.ParseDuration(h.MaxGap)
		if e != nil || gap < time.Second || gap > time.Minute {
			return nil, fmt.Errorf("histogram maxGap must be 1s..1m")
		}
		seen := map[string]bool{}
		if len(h.Outcomes) == 0 {
			return nil, fmt.Errorf("histogram outcomes required")
		}
		for _, o := range h.Outcomes {
			if seen[o] || (o != "ok" && o != "error" && o != "panic") {
				return nil, fmt.Errorf("invalid histogram outcomes")
			}
			seen[o] = true
		}
		p.Histograms[i].Bounds = append([]float64{}, h.Bounds...)
		p.Histograms[i].Outcomes = append([]string(nil), h.Outcomes...)
	}
	rateNames := map[string]bool{}
	for _, r := range p.Rates {
		if !stableName.MatchString(r.Name) || rateNames[r.Name] || r.Metric != "goo.operation.calls" || r.Unit != "{invocation}" || (r.Temporality != "cumulative" && r.Temporality != "delta") {
			return nil, fmt.Errorf("unsupported or duplicate counter rate contract")
		}
		rateNames[r.Name] = true
		gap, e := time.ParseDuration(r.MaxGap)
		if e != nil || gap < time.Second || gap > time.Minute {
			return nil, fmt.Errorf("maxGap must be 1s..1m")
		}
		seen := map[string]bool{}
		if len(r.Outcomes) == 0 {
			return nil, fmt.Errorf("rate outcomes required")
		}
		for _, o := range r.Outcomes {
			if seen[o] || (o != "ok" && o != "error" && o != "panic") {
				return nil, fmt.Errorf("unsupported or duplicate rate outcome")
			}
			seen[o] = true
		}
	}
	p.Ratios = append([]RatioContract(nil), p.Ratios...)
	ratioNames := map[string]bool{}
	for i, r := range p.Ratios {
		if !stableName.MatchString(r.Name) || ratioNames[r.Name] {
			return nil, fmt.Errorf("invalid or duplicate ratio name")
		}
		ratioNames[r.Name] = true
		found := false
		for _, rate := range p.Rates {
			if rate.Name == r.TotalRate {
				// The operation outcome domain is exhaustive; partial denominators would mislabel an error ratio.
				if len(rate.Outcomes) != 3 {
					return nil, fmt.Errorf("error ratio requires all operation outcomes")
				}
				if e := metrics.ValidateRatioPolicy(metrics.RatioPolicy{Outcomes: rate.Outcomes, Errors: r.ErrorOutcomes}); e != nil {
					return nil, e
				}
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf("ratio must reference a declared total rate")
		}
		p.Ratios[i].ErrorOutcomes = append([]string{}, r.ErrorOutcomes...)
	}
	if p.Fleet != nil {
		fleet := *p.Fleet
		fleet.Members = append([]string{}, fleet.Members...)
		if fleet.Comparison != nil {
			copy := *fleet.Comparison
			fleet.Comparison = &copy
			skew, e := time.ParseDuration(copy.MaxSkew)
			if e != nil {
				return nil, e
			}
			if e = metrics.ValidateComparisonPolicy(metrics.ComparisonPolicy{Mode: copy.Mode, MaxSkewNS: fmt.Sprint(int64(skew))}); e != nil {
				return nil, e
			}
		}
		p.Fleet = &fleet
		if !stableName.MatchString(fleet.Name) || !ratioNames[fleet.Ratio] || len(fleet.Description) > 500 {
			return nil, fmt.Errorf("fleet requires a stable name and declared ratio")
		}
		if e := metrics.ValidateFleetMembers(fleet.Members); e != nil {
			return nil, e
		}
		selected := false
		for _, instance := range fleet.Members {
			if instance == source.Instance {
				selected = true
			}
			memberSource := source
			memberSource.Instance = instance
			if _, _, e := m.OperationQuery(memberSource, 0, 1); e != nil {
				return nil, e
			}
		}
		if !selected {
			return nil, fmt.Errorf("fleet must include selected resource instance")
		}
		sort.Strings(p.Fleet.Members)
	}
	views := map[string]bool{}
	for _, v := range p.Views {
		if views[v] {
			return nil, fmt.Errorf("duplicate view %s", v)
		}
		switch v {
		case "operation-overview", "duration-heatmap", "trace-waterfall":
		case "histogram-heatmap":
			if len(p.Histograms) == 0 {
				return nil, fmt.Errorf("histogram view requires contract")
			}
		case "fleet-coverage":
			if p.Fleet == nil {
				return nil, fmt.Errorf("fleet-coverage view requires declared fleet")
			}
		case "error-ratios":
			if len(p.Ratios) == 0 {
				return nil, fmt.Errorf("error-ratios view requires ratios")
			}
		case "counter-rates":
			if len(p.Rates) == 0 {
				return nil, fmt.Errorf("counter-rates view requires rates")
			}
		case "subject-events":
			if p.Subjects == nil {
				return nil, fmt.Errorf("subject-events view requires subject binding")
			}
		default:
			return nil, fmt.Errorf("unsupported view %s", v)
		}
		views[v] = true
	}
	if !views["operation-overview"] {
		return nil, fmt.Errorf("operation-overview view required")
	}
	if views["subject-events"] && !views["trace-waterfall"] {
		return nil, fmt.Errorf("subject-events requires trace-waterfall navigation")
	}
	plan := &OperationPlan{APIVersion: "goo/operation-plan/v1alpha1", Definition: p, Manifest: m, Source: source, Diagnostics: []string{"Source declarations are not build-coverage proof.", "Sampling and completeness remain unverified.", "Schema compatibility is declared; this exporter omits scope schema URLs.", "No fleet rate, business outcome, ratio or SLO is inferred."}}
	payload, _ := json.Marshal(struct {
		Definition       OperationPackage
		ManifestRevision string
	}{p, m.Revision})
	plan.Revision = hash(payload)
	return plan, nil
}
func (p *OperationPlan) Fetch(ctx context.Context, client *http.Client, endpoint, user, password string, from, to int64) (*LiveOperations, error) {
	fresh, err := p.Definition.Bind(p.Manifest)
	if err != nil {
		return nil, err
	}
	if fresh.Revision != p.Revision || fresh.Source != p.Source {
		return nil, fmt.Errorf("operation plan changed after binding")
	}
	result, err := FetchOperations(ctx, client, endpoint, user, password, *p.Manifest, p.Source, from, to)
	if err != nil {
		return nil, err
	}
	result.Package = p
	payload, _ := json.Marshal([]string{p.Revision, result.Conformance.Basis})
	result.Conformance.Basis = hash(payload)
	result.Conformance.Warnings = append(result.Conformance.Warnings, p.Diagnostics...)
	return result, nil
}

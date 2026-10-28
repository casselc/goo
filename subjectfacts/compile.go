package subjectfacts

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Plan is an inspectable artifact, not a deployment. Prepared-source constraints
// are part of its contract and must be checked by a future serving adapter.
type Plan struct {
	Target       string            `json:"target"`
	Status       string            `json:"status"`
	Requirements []string          `json:"requirements"`
	Parameters   map[string]string `json:"parameters"`
	Query        string            `json:"query"`
	MetricYAML   string            `json:"metricYaml,omitempty"`
	DDL          string            `json:"ddl,omitempty"`
	Definition   *Definition       `json:"definition"`
}

func quote(s string) string { return "`" + s + "`" }

func table(s string) string {
	p := strings.Split(s, ".")
	for i := range p {
		p[i] = quote(p[i])
	}
	return strings.Join(p, ".")
}

func predicateSQL(p *Predicate) string {
	return predicateQualified(p, "")
}

func predicateQualified(p *Predicate, prefix string) string {
	if p == nil {
		return "true"
	}
	var literal string
	switch v := p.Eq.(type) {
	case bool:
		literal = fmt.Sprint(v)
	case string:
		literal = "'" + strings.ReplaceAll(strings.ReplaceAll(v, "\\", "\\\\"), "'", "''") + "'"
	}
	return prefix + quote(p.Field) + " = " + literal
}

func (d *Definition) Compile(target string, groups []string) (*Plan, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}
	if err := d.validateGroups(groups); err != nil {
		return nil, err
	}
	if target != "clickhouse" && target != "databricks" {
		return nil, fmt.Errorf("unsupported backend %q", target)
	}
	source := d.Source.Bindings[target]
	if source == "" {
		return nil, fmt.Errorf("source has no %s binding", target)
	}
	p := &Plan{
		Target: target, Status: "requires-prepared-source", Definition: d,
		Parameters: map[string]string{"evaluation_id": "string", "evidence_impaired": "bool"},
		Requirements: []string{
			"Prepared, typed subject facts; no raw-event lifecycle inference is performed.",
			"Exactly one evaluation_id pins cohort bounds, definition revision and evidence basis.",
			"Unique scoped subject key within the evaluation; non-null declared fields.",
			"Invalid unkeyed records and empty-scope evidence require a separate evaluation receipt.",
			"Bind evidence_impaired from the complete receipt, including impaired scopes with no rows.",
			"Source binding and declared dimensions must preserve authorization and subject grain.",
		},
	}
	if target == "databricks" {
		if err := d.compileDatabricks(p, groups, source); err != nil {
			return nil, err
		}
	} else {
		d.compileClickHouse(p, groups, source)
	}
	return p, nil
}

func ratioParts(m Measure, ref func(string) string) (string, string) {
	blockers := make([]string, 0, len(m.UnknownWhen))
	for _, b := range m.UnknownWhen {
		blockers = append(blockers, ref(b)+" > 0")
	}
	blocked := "false"
	if len(blockers) > 0 {
		blocked = strings.Join(blockers, " OR ")
	}
	value := "CASE WHEN " + blocked + " THEN NULL ELSE " + ref(m.Numerator) +
		" * 1.0 / NULLIF(" + ref(m.Denominator) + ", 0) END"
	state := "CASE WHEN " + blocked + " THEN 'unknown' WHEN " + ref(m.Denominator) +
		" = 0 THEN 'no_data' ELSE 'ok' END"
	return value, state
}

func (d *Definition) compileClickHouse(p *Plan, groups []string, source string) {
	inner, outer, groupExprs := []string{}, []string{}, []string{}
	for _, g := range groups {
		inner = append(inner, "source."+quote(d.Dimensions[g])+" AS "+quote(g))
		outer = append(outer, quote(g))
		groupExprs = append(groupExprs, "source."+quote(d.Dimensions[g]))
	}
	for _, name := range d.measureOrder() {
		m := d.Measures[name]
		if m.Op == "count" {
			expr := "count()"
			if m.Where != nil {
				expr = "countIf(" + predicateQualified(m.Where, "source.") + ")"
			}
			inner = append(inner, expr+" AS "+quote(name))
			outer = append(outer, quote(name))
		} else {
			value, state := ratioParts(m, quote)
			value, state = guardRatio(value, state, "{evidence_impaired:Bool}")
			outer = append(outer, value+" AS "+quote(name), state+" AS "+quote(name+"_state"))
		}
	}
	q := "WITH aggregate_facts AS (\n  SELECT\n    " + strings.Join(inner, ",\n    ") +
		"\n  FROM " + table(source) + " AS source\n  WHERE source.`evaluation_id` = {evaluation_id:String}" +
		" AND (" + predicateQualified(d.Population, "source.") + ")"
	if len(groupExprs) > 0 {
		q += "\n  GROUP BY " + strings.Join(groupExprs, ", ")
	}
	q += "\n)\nSELECT\n  " + strings.Join(outer, ",\n  ") + "\nFROM aggregate_facts"
	if len(groups) > 0 {
		order := make([]string, len(groups))
		for i, g := range groups {
			order[i] = quote(g)
		}
		q += "\nORDER BY " + strings.Join(order, ", ")
	}
	p.Query = q + "\nFORMAT JSONEachRow"
}

type metricEntry struct {
	Name string `yaml:"name"`
	Expr string `yaml:"expr"`
}

type metricView struct {
	Version    string        `yaml:"version"`
	Source     string        `yaml:"source"`
	Comment    string        `yaml:"comment"`
	Filter     string        `yaml:"filter,omitempty"`
	Dimensions []metricEntry `yaml:"dimensions"`
	Measures   []metricEntry `yaml:"measures"`
}

func (d *Definition) compileDatabricks(p *Plan, groups []string, source string) error {
	ref := func(s string) string { return "MEASURE(" + quote(s) + ")" }
	v := metricView{
		Version: "1.1", Source: source,
		Comment:    "Generated by goo/v1alpha1. Query exactly one evaluation_id; source preparation is required.",
		Dimensions: []metricEntry{{"evaluation_id", quote("evaluation_id")}},
	}
	if d.Population != nil {
		v.Filter = predicateSQL(d.Population)
	}
	for _, name := range names(d.Dimensions) {
		v.Dimensions = append(v.Dimensions, metricEntry{name, quote(d.Dimensions[name])})
	}
	querySelect := []string{}
	for _, g := range groups {
		querySelect = append(querySelect, quote(g))
	}
	for _, name := range d.measureOrder() {
		m := d.Measures[name]
		var expr string
		if m.Op == "count" {
			expr = "COUNT(1)"
			if m.Where != nil {
				expr += " FILTER (WHERE " + predicateSQL(m.Where) + ")"
			}
		} else {
			expr, _ = ratioParts(m, ref)
		}
		v.Measures = append(v.Measures, metricEntry{name, expr})
		queryValue := ref(name)
		if m.Op == "ratio" {
			_, state := ratioParts(m, ref)
			// The metric definition is reusable; the request binds receipt
			// quality outside it so even zero-row scopes suppress the ratio.
			queryValue, state = guardRatio(queryValue, state, ":evidence_impaired")
			querySelect = append(querySelect, state+" AS "+quote(name+"_state"))
		}
		querySelect = append(querySelect, queryValue+" AS "+quote(name))
	}
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	// Dollar quoting is a SQL delimiter, independent of YAML string escaping.
	if strings.Contains(string(b), "$$") {
		return fmt.Errorf("Databricks dollar-quoted DDL cannot contain a $$ literal in this subset")
	}
	p.MetricYAML = string(b)
	parts := strings.Split(source, ".")
	viewName := table(strings.Join(parts[:2], ".") + "." + d.Name)
	if d.Name == parts[2] {
		return fmt.Errorf("metric view name cannot replace its prepared source")
	}
	p.DDL = "CREATE VIEW " + viewName + "\nWITH METRICS\nLANGUAGE YAML\nAS $$\n" + string(b) + "$$;\n"
	p.Query = "SELECT\n  " + strings.Join(querySelect, ",\n  ") + "\nFROM " + viewName +
		"\nWHERE `evaluation_id` = :evaluation_id"
	if len(groups) > 0 {
		p.Query += "\nGROUP BY ALL\nORDER BY "
		order := make([]string, len(groups))
		for i, g := range groups {
			order[i] = quote(g)
		}
		p.Query += strings.Join(order, ", ")
	}
	p.Query += ";\n"
	p.Requirements = append(p.Requirements,
		"Databricks YAML 1.1 profile; warehouse acceptance and execution have not been verified.",
		"DDL is generated only; it is not executed, deployed, or granted automatically.")
	return nil
}

func guardRatio(value, state, parameter string) (string, string) {
	return "CASE WHEN " + parameter + " THEN NULL ELSE (" + value + ") END",
		"CASE WHEN " + parameter + " THEN 'unknown' ELSE (" + state + ") END"
}

// Emit atomic definitions before their composed measures. The portable subset
// only permits count dependencies, so no arbitrary expression DAG is needed.
func (d *Definition) measureOrder() []string {
	ordered := []string{}
	for _, op := range []string{"count", "ratio"} {
		for _, name := range names(d.Measures) {
			if d.Measures[name].Op == op {
				ordered = append(ordered, name)
			}
		}
	}
	return ordered
}

package lifecycle

import (
	"fmt"

	"github.com/casselc/goo/measurement"
)

// Evaluate consumes the complete prepared artifact, not facts alone. A strict
// request-wide receipt guard covers impaired scopes even when they have no rows.
// This is intentionally conservative for cross-scope grouping.
func (p *Prepared) Evaluate(d *measurement.Definition, groups []string) (*Evaluation, error) {
	if _, err := p.Receipt.Policy.validate(); err != nil {
		return nil, err
	}
	if p.Receipt.State != "ok" && p.Receipt.State != "unknown" && p.Receipt.State != "no_data" {
		return nil, fmt.Errorf("invalid preparation receipt state")
	}
	if d.Population == nil || d.Population.Field != "in_cohort" || d.Population.Eq != true {
		return nil, fmt.Errorf("lifecycle evaluation requires population in_cohort = true")
	}
	wantKeys := []string{"environment", "tenant_id", "service_name", p.Receipt.Policy.SubjectField}
	if len(d.Source.SubjectKey) != len(wantKeys) {
		return nil, fmt.Errorf("measurement subject key must preserve environment, tenant, service and subject")
	}
	for _, key := range wantKeys {
		if !has(d.Source.SubjectKey, key) {
			return nil, fmt.Errorf("measurement subject key is missing %s", key)
		}
	}
	for _, f := range p.Facts {
		if f["evaluation_id"] != p.Receipt.Policy.EvaluationID {
			return nil, fmt.Errorf("fact belongs to another evaluation; snapshots must not be mixed")
		}
	}
	rows, err := d.Evaluate(p.Facts, p.Receipt.Policy.EvaluationID, groups)
	if err != nil {
		return nil, err
	}
	if p.Receipt.State == "unknown" {
		for _, row := range rows {
			for name, m := range d.Measures {
				if m.Op == "ratio" {
					row.Measures[name] = measurement.Value{State: "unknown"}
				}
			}
		}
	}
	return &Evaluation{p.Receipt, rows}, nil
}

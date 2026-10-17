package annotations

import (
	"fmt"
	"github.com/casselc/goo/metrics"
)

type RatioContract struct {
	Name          string   `yaml:"name" json:"name"`
	TotalRate     string   `yaml:"totalRate" json:"totalRate"`
	ErrorOutcomes []string `yaml:"errorOutcomes" json:"errorOutcomes"`
}
type LiveRatio struct {
	Contract RatioContract        `json:"contract"`
	Result   *metrics.RatioResult `json:"result"`
}

func (p *OperationPlan) evaluateRatios(rates []LiveRate) ([]LiveRatio, error) {
	result := []LiveRatio{}
	for _, contract := range p.Definition.Ratios {
		found := false
		for _, rate := range rates {
			if rate.Contract.Name == contract.TotalRate {
				r, e := metrics.EvaluateErrorRatios(*rate.Result, metrics.RatioPolicy{Outcomes: rate.Contract.Outcomes, Errors: contract.ErrorOutcomes})
				if e != nil {
					return nil, e
				}
				result = append(result, LiveRatio{contract, r})
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("ratio counter basis unavailable")
		}
	}
	return result, nil
}

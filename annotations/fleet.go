package annotations

import (
	"context"
	"fmt"
	"github.com/casselc/goo/metrics"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"
)

type FleetComparisonContract struct {
	Mode    string `yaml:"mode" json:"mode"`
	MaxSkew string `yaml:"maxSkew" json:"maxSkew"`
}
type FleetContract struct {
	Comparison  *FleetComparisonContract `yaml:"comparison,omitempty" json:"comparison,omitempty"`
	Description string                   `yaml:"description,omitempty" json:"description,omitempty"`
	Name        string                   `yaml:"name" json:"name"`
	Ratio       string                   `yaml:"ratio" json:"ratio"`
	Members     []string                 `yaml:"members" json:"members"`
}
type LiveFleet struct {
	Contract FleetContract        `json:"contract"`
	Result   *metrics.FleetResult `json:"result"`
}

// FetchRates isolates peer source failures from the selected instance result.
// At most four members are queried concurrently; each retains the source guards.
func (p *OperationPlan) FetchRates(ctx context.Context, client *http.Client, endpoint, user, password string, from, to int64) (*LiveRates, error) {
	if p.Definition.Fleet == nil {
		return p.fetchInstanceRates(ctx, client, endpoint, user, password, from, to)
	}
	fresh, e := p.Definition.Bind(p.Manifest)
	if e != nil || fresh.Revision != p.Revision || !reflect.DeepEqual(fresh.Source, p.Source) {
		return nil, fmt.Errorf("invalid operation plan")
	}
	contract := p.Definition.Fleet
	type reply struct {
		result *LiveRates
		err    error
	}
	replies := make([]reply, len(contract.Members))
	var wg sync.WaitGroup
	for index, instance := range contract.Members {
		wg.Add(1)
		go func(index int, instance string) {
			defer wg.Done()
			plan := p
			if instance != p.Source.Instance {
				definition := p.Definition
				definition.Resource.Instance = instance
				var err error
				plan, err = definition.Bind(p.Manifest)
				if err != nil {
					replies[index].err = err
					return
				}
			}
			replies[index].result, replies[index].err = plan.fetchInstanceRates(ctx, client, endpoint, user, password, from, to)
		}(index, instance)
	}
	wg.Wait()
	var result *LiveRates
	for i, instance := range contract.Members {
		if instance == p.Source.Instance {
			if replies[i].err != nil {
				return nil, replies[i].err
			}
			result = replies[i].result
		}
	}
	if result == nil {
		return nil, fmt.Errorf("selected instance absent from fleet")
	}
	var ratioContract RatioContract
	var rateContract RateContract
	for _, r := range p.Definition.Ratios {
		if r.Name == contract.Ratio {
			ratioContract = r
		}
	}
	for _, r := range p.Definition.Rates {
		if r.Name == ratioContract.TotalRate {
			rateContract = r
		}
	}
	request := metrics.FleetRequest{FromNS: strconv.FormatInt(from*1000000, 10), ToNS: strconv.FormatInt(to*1000000, 10), Members: contract.Members, Policy: metrics.RatioPolicy{Outcomes: rateContract.Outcomes, Errors: ratioContract.ErrorOutcomes}}
	if contract.Comparison != nil {
		skew, _ := time.ParseDuration(contract.Comparison.MaxSkew)
		request.Comparison = &metrics.ComparisonPolicy{Mode: contract.Comparison.Mode, MaxSkewNS: strconv.FormatInt(int64(skew), 10)}
	}
	for i, instance := range contract.Members {
		input := metrics.FleetInput{Instance: instance}
		if replies[i].err != nil {
			result.Warnings = append(result.Warnings, "Fleet member "+instance+" source unavailable; fleet coverage remains unknown.")
		} else {
			for _, rate := range replies[i].result.Rates {
				if rate.Contract.Name == rateContract.Name {
					input.Counter = rate.Result
					break
				}
			}
		}
		request.Inputs = append(request.Inputs, input)
	}
	evaluated, e := metrics.EvaluateFleet(request)
	if e != nil {
		return nil, e
	}
	result.Fleet = &LiveFleet{Contract: *contract, Result: evaluated}
	return result, nil
}

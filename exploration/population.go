package exploration

import (
	"encoding/json"
	"fmt"
)

type PopulationRequest struct {
	Question  Query  `json:"question"`
	Companion *Query `json:"companion,omitempty"`
}
type PopulationPlan struct {
	Version string   `json:"version"`
	Queries []Query  `json:"queries"`
	Notices []string `json:"notices"`
}

// PlanPopulation composes existing kernels without reading or broadening selectors.
func PlanPopulation(in PopulationRequest) (*PopulationPlan, error) {
	if e := in.Question.Validate(); e != nil {
		return nil, e
	}
	raw, _ := json.Marshal(in)
	var detached PopulationRequest
	json.Unmarshal(raw, &detached)
	in = detached
	prepare := func(q Query) Query {
		q = q.WithEnvironmentFrom(q)
		q.Offset = 0
		if q.BucketMS == 0 {
			for _, b := range []int64{1000, 10000, 60000, 300000} {
				q.BucketMS = b
				if (q.ToMS-q.FromMS+b-1)/b <= 60 {
					break
				}
			}
		}
		if q.Signal == "spans" {
			q.Summary = "span-duration"
			q.SpanStatus = true
		}
		return q
	}
	q := prepare(in.Question)
	if _, e := Compile(q); e != nil {
		return nil, e
	}
	out := &PopulationPlan{Version: "goo/population/v1", Queries: []Query{q}, Notices: []string{"Independent signal reads in one fixed window. Matching predicates do not establish event identity or completeness."}}
	other := q
	other.Signal = "logs"
	if q.Signal == "logs" {
		other.Signal = "spans"
	}
	other.Summary = ""
	other.SpanStatus = false
	other.GroupBy = nil
	other.GroupTimeSeries = false
	other.Numeric = nil
	other.NumericHistogram = nil
	if in.Companion != nil {
		other = *in.Companion
		if other.Signal == q.Signal || !other.SameEnvironment(q) || other.FromMS != q.FromMS || other.ToMS != q.ToMS {
			return nil, fmt.Errorf("companion requires the other signal and identical environment/window")
		}
		out.Notices = append(out.Notices, "Companion uses an explicitly independent selector; inspect each question before interpreting matching records.")
	}
	other = prepare(other)
	if _, e := Compile(other); e != nil {
		if in.Companion != nil {
			return nil, e
		}
		out.Notices = append(out.Notices, "Companion "+other.Signal+" withheld: "+e.Error()+". No predicates were dropped. Configure an independent companion selector to include it.")
		return out, nil
	}
	out.Queries = append(out.Queries, other)
	return out, nil
}

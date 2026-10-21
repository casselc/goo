package workbench

import (
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
	"reflect"
)

type ExplorationTail struct {
	Signal   string `json:"signal"`
	Baseline string `json:"baseline,omitempty"`
}

type ExplorationBreakdown struct {
	Metric           string             `json:"metric"`
	ComparisonMetric string             `json:"comparisonMetric"`
	Baseline         *exploration.Query `json:"baseline,omitempty"`
}

// ExplorationView is reusable fixed-question configuration, never captured evidence.
type ExplorationView struct {
	Version    int                             `json:"version"`
	ID         string                          `json:"id,omitempty"`
	Title      string                          `json:"title"`
	Queries    []exploration.Query             `json:"queries"`
	Layout     string                          `json:"layout"`
	Columns    []string                        `json:"columns"`
	Tail       *ExplorationTail                `json:"tail,omitempty"`
	Breakdowns map[string]ExplorationBreakdown `json:"breakdowns,omitempty"`
}

func (v ExplorationView) Validate() error {
	if v.Version != 1 || !text(v.Title, 120, true, false) || len(v.Queries) < 1 || len(v.Queries) > 2 || (v.Layout != "charts-records" && v.Layout != "records") || len(v.Columns) < 2 || len(v.Columns) > 4 {
		return errors.New("require bounded title, 1..2 fixed queries, supported layout and 2..4 columns")
	}
	seen := map[string]bool{}
	for _, c := range v.Columns {
		if seen[c] || (c != "utc" && c != "service" && c != "summary" && c != "context") {
			return errors.New("invalid or duplicate view column")
		}
		seen[c] = true
	}
	if !seen["utc"] || !seen["summary"] {
		return errors.New("UTC and summary columns required")
	}
	signals := map[string]bool{}
	first := v.Queries[0]
	for _, q := range v.Queries {
		if signals[q.Signal] || q.Offset != 0 || q.Environment != first.Environment || q.FromMS != first.FromMS || q.ToMS != first.ToMS {
			return errors.New("view queries require distinct signals, page zero and identical environment/window")
		}
		signals[q.Signal] = true
		if _, e := exploration.Compile(q); e != nil {
			return e
		}
	}

	if v.Tail != nil {
		if !signals[v.Tail.Signal] || (v.Tail.Baseline != "" && !ID.MatchString(v.Tail.Baseline)) {
			return errors.New("tail requires a configured signal and valid optional baseline capture ID")
		}
	}
	for signal, config := range v.Breakdowns {
		var current *exploration.Query
		for i := range v.Queries {
			if v.Queries[i].Signal == signal {
				current = &v.Queries[i]
			}
		}
		if current == nil || len(current.GroupBy) != 1 {
			return errors.New("breakdown requires a configured single-group signal")
		}
		if config.Metric != "count" && config.Metric != "error-rate" && config.Metric != "mean" && config.Metric != "p95" && config.Metric != "p99" && config.Metric != "max" && config.Metric != "numeric-mean" && config.Metric != "numeric-sum" && config.Metric != "numeric-min" && config.Metric != "numeric-max" {
			return errors.New("unsupported breakdown ranking")
		}
		if config.ComparisonMetric != "share" && config.ComparisonMetric != "error-rate" && config.ComparisonMetric != "mean" && config.ComparisonMetric != "p95" && config.ComparisonMetric != "p99" {
			return errors.New("unsupported comparison metric")
		}
		if b := config.Baseline; b != nil {
			if b.Signal != signal || b.Environment != current.Environment || b.Offset != 0 || !reflect.DeepEqual(b.GroupBy, current.GroupBy) || b.Numeric != nil || current.Numeric != nil {
				return errors.New("baseline requires the same signal, environment and breakdown without numeric measures")
			}
			if _, e := exploration.Compile(*b); e != nil {
				return e
			}
		}
	}
	return nil
}
func (s Store) CreateExplorationView(v ExplorationView) (*ExplorationView, error) {
	if v.ID != "" {
		return nil, errors.New("new view must omit ID")
	}
	if e := v.Validate(); e != nil {
		return nil, e
	}
	raw, _ := json.Marshal(v)
	var detached ExplorationView
	json.Unmarshal(raw, &detached)
	v = detached
	id, e := digest(v)
	if e != nil {
		return nil, e
	}
	v.ID = id
	if e = s.publish("exploration-views", id, v); e != nil {
		return nil, e
	}
	return s.LoadExplorationView(id)
}
func (s Store) LoadExplorationView(id string) (*ExplorationView, error) {
	var v ExplorationView
	if e := s.read("exploration-views", id, &v); e != nil {
		return nil, e
	}
	actual := v.ID
	v.ID = ""
	hash, e := digest(v)
	v.ID = actual
	if e != nil || hash != id || actual != id || v.Validate() != nil {
		return nil, errors.New("exploration view integrity check failed")
	}
	return &v, nil
}

type ExplorationViewList struct {
	Items      []ExplorationView `json:"items"`
	NextCursor string            `json:"nextCursor,omitempty"`
}

func (s Store) ExplorationViews(after string) (*ExplorationViewList, error) {
	if after != "" && !ID.MatchString(after) {
		return nil, errors.New("invalid view cursor")
	}
	ids, e := s.ids("exploration-views")
	if e != nil {
		return nil, e
	}
	out := &ExplorationViewList{Items: []ExplorationView{}}
	for _, id := range ids {
		if id <= after {
			continue
		}
		if len(out.Items) == 20 {
			out.NextCursor = out.Items[19].ID
			break
		}
		v, e := s.LoadExplorationView(id)
		if e != nil {
			return nil, e
		}
		out.Items = append(out.Items, *v)
	}
	return out, nil
}

package workbench

import (
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
)

// ExplorationView is reusable fixed-question configuration, never captured evidence.
type ExplorationView struct {
	Version int                 `json:"version"`
	ID      string              `json:"id,omitempty"`
	Title   string              `json:"title"`
	Queries []exploration.Query `json:"queries"`
	Layout  string              `json:"layout"`
	Columns []string            `json:"columns"`
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
	json.Unmarshal(raw, &v)
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

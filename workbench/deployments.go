package workbench

import (
	"errors"
	"net/url"
	"sort"
	"time"

	"github.com/casselc/goo/exploration"
)

type DeploymentInput struct {
	Environment  string `json:"environment"`
	Service      string `json:"service,omitempty"`
	Title        string `json:"title"`
	Revision     string `json:"revision"`
	OccurredAtMS int64  `json:"occurredAtMs"`
	URL          string `json:"url,omitempty"`
}
type Deployment struct {
	ID         string `json:"id"`
	Version    int    `json:"version"`
	RecordedAt string `json:"recordedAt"`
	DeploymentInput
}

func (m DeploymentInput) Validate() error {
	if !text(m.Environment, 128, true, false) || !text(m.Service, 128, false, false) || !text(m.Title, 120, true, false) || !text(m.Revision, 160, true, false) || m.OccurredAtMS < 0 || m.OccurredAtMS >= 4102444800000 {
		return errors.New("require bounded environment, deployment title/revision and event time before 2100")
	}
	if m.URL != "" {
		u, e := url.Parse(m.URL)
		if e != nil || len(m.URL) > 2048 || !text(m.URL, 2048, true, false) || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.User != nil {
			return errors.New("deployment URL must be an absolute HTTP(S) URL without credentials")
		}
	}
	return nil
}
func (s Store) CreateDeployment(in DeploymentInput) (*Deployment, error) {
	if e := in.Validate(); e != nil {
		return nil, e
	}
	m := Deployment{Version: 1, RecordedAt: time.Now().UTC().Format(time.RFC3339Nano), DeploymentInput: in}
	id, e := digest(m)
	if e != nil {
		return nil, e
	}
	m.ID = id
	if e = s.publish("deployments", id, m); e != nil {
		return nil, e
	}
	return s.LoadDeployment(id)
}
func (s Store) LoadDeployment(id string) (*Deployment, error) {
	var m Deployment
	if e := s.read("deployments", id, &m); e != nil {
		return nil, e
	}
	actual := m.ID
	m.ID = ""
	hash, e := digest(m)
	m.ID = actual
	if e != nil || actual != id || hash != id || m.Version != 1 || !validTime(m.RecordedAt) || m.DeploymentInput.Validate() != nil {
		return nil, errors.New("deployment integrity check failed")
	}
	return &m, nil
}

type DeploymentScope struct {
	Environment string `json:"environment"`
	FromMS      int64  `json:"fromMs"`
	ToMS        int64  `json:"toMs"`
}

func (q DeploymentScope) Validate() error {
	return (exploration.Query{Signal: "spans", Environment: q.Environment, FromMS: q.FromMS, ToMS: q.ToMS, Limit: 1}).Validate()
}

type DeploymentContext struct {
	Scope  DeploymentScope `json:"scope"`
	ReadAt string          `json:"readAt"`
	Items  []Deployment    `json:"items"`
}

// This is a fresh local-context read, not part of a telemetry capture or causal evidence.
func (s Store) Deployments(q DeploymentScope) (*DeploymentContext, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	ids, e := s.ids("deployments")
	if e != nil {
		return nil, e
	}
	r := &DeploymentContext{Scope: q, ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Items: []Deployment{}}
	cutoff, _ := time.Parse(time.RFC3339Nano, r.ReadAt)
	for _, id := range ids {
		m, e := s.LoadDeployment(id)
		if e != nil {
			return nil, e
		}
		recorded, _ := time.Parse(time.RFC3339Nano, m.RecordedAt)
		if m.Environment == q.Environment && m.OccurredAtMS >= q.FromMS && m.OccurredAtMS < q.ToMS && !recorded.After(cutoff) {
			r.Items = append(r.Items, *m)
			if len(r.Items) > 200 {
				return nil, errors.New("deployment context exceeds 200 markers; narrow the window")
			}
		}
	}
	sort.Slice(r.Items, func(i, j int) bool {
		if r.Items[i].OccurredAtMS == r.Items[j].OccurredAtMS {
			return r.Items[i].ID < r.Items[j].ID
		}
		return r.Items[i].OccurredAtMS < r.Items[j].OccurredAtMS
	})
	return r, nil
}

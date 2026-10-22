package retention

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/casselc/goo/workbench"
	"sort"
	"strings"
	"time"
)

// Catalog is a bounded metadata scan over immutable retained manifests and
// current explicit annotation heads. It never queries the telemetry source.
type CatalogQuery struct {
	Tag         string `json:"tag,omitempty"`
	Service     string `json:"service,omitempty"`
	Environment string `json:"environment,omitempty"`
	From        string `json:"from,omitempty"`
	To          string `json:"to,omitempty"`
	Cursor      string `json:"cursor,omitempty"`
}
type CatalogItem struct {
	ID           string   `json:"id"`
	CapturedAt   string   `json:"capturedAt"`
	Environment  string   `json:"environment"`
	Services     []string `json:"services"`
	Tags         []string `json:"tags"`
	Observations int      `json:"observations"`
	Origin       string   `json:"origin"`
}
type CatalogPage struct {
	Items      []CatalogItem `json:"items"`
	NextCursor string        `json:"nextCursor,omitempty"`
}

func (q CatalogQuery) Validate() error {
	if len(q.Tag) > 64 || len(q.Service) > 256 || len(q.Environment) > 256 || len(q.Cursor) > 1024 {
		return errors.New("catalog filter exceeds bound")
	}
	if q.Tag != "" {
		e := (workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "evidence", ID: strings.Repeat("a", 64)}, Tags: []string{q.Tag}, State: "active"}).Validate()
		if e != nil {
			return errors.New("invalid exact tag")
		}
	}
	var from, to time.Time
	var e error
	if q.From != "" {
		from, e = time.Parse(time.RFC3339Nano, q.From)
		if e != nil {
			return errors.New("invalid captured-from UTC timestamp")
		}
	}
	if q.To != "" {
		to, e = time.Parse(time.RFC3339Nano, q.To)
		if e != nil {
			return errors.New("invalid captured-before UTC timestamp")
		}
	}
	if q.From != "" && q.To != "" && !from.Before(to) {
		return errors.New("capture time window must be half-open and increasing")
	}
	return nil
}
func (s Store) Catalog(ctx context.Context, q CatalogQuery) (*CatalogPage, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	rows, e := c.Query(ctx, "SELECT payload FROM goo_evidence ORDER BY _id LIMIT 1001")
	if e != nil {
		return nil, e
	}
	manifests := []Manifest{}
	bytes := 0
	for rows.Next() {
		var raw string
		if e = rows.Scan(&raw); e != nil {
			rows.Close()
			return nil, e
		}
		bytes += len(raw)
		if bytes > 16<<20 || len(manifests) >= 1000 {
			rows.Close()
			return nil, errors.New("retained catalog exceeds prototype scan bound; no partial catalog returned")
		}
		var m Manifest
		if json.Unmarshal([]byte(raw), &m) != nil || m.Validate() != nil {
			rows.Close()
			return nil, errors.New("invalid catalog manifest")
		}
		manifests = append(manifests, m)
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	rows, e = c.Query(ctx, "SELECT payload FROM goo_annotations ORDER BY _id LIMIT 10001")
	if e != nil {
		return nil, e
	}
	revisions := []workbench.MeaningRevision{}
	parents := map[string]bool{}
	for rows.Next() {
		var raw string
		if e = rows.Scan(&raw); e != nil {
			rows.Close()
			return nil, e
		}
		bytes += len(raw)
		if bytes > 64<<20 || len(revisions) >= 10000 {
			rows.Close()
			return nil, errors.New("annotation catalog exceeds prototype scan bound")
		}
		var r workbench.MeaningRevision
		if json.Unmarshal([]byte(raw), &r) != nil || workbench.ValidateMeaningRevision(r) != nil || r.Target.Kind != "evidence" {
			rows.Close()
			return nil, errors.New("invalid catalog annotation")
		}
		revisions = append(revisions, r)
		if r.Parent != "" {
			parents[r.Parent] = true
		}
	}
	e = rows.Err()
	rows.Close()
	if e != nil {
		return nil, e
	}
	return catalogPage(q, manifests, revisions, parents)
}
func catalogPage(q CatalogQuery, manifests []Manifest, revisions []workbench.MeaningRevision, parents map[string]bool) (*CatalogPage, error) {
	tags := map[string]map[string]bool{}
	for _, r := range revisions {
		if parents[r.ID] || r.State != "active" {
			continue
		}
		if tags[r.Target.ID] == nil {
			tags[r.Target.ID] = map[string]bool{}
		}
		for _, tag := range r.Tags {
			tags[r.Target.ID][tag] = true
		}
	}
	var from, to time.Time
	if q.From != "" {
		from, _ = time.Parse(time.RFC3339Nano, q.From)
	}
	if q.To != "" {
		to, _ = time.Parse(time.RFC3339Nano, q.To)
	}
	items := []CatalogItem{}
	for _, m := range manifests {
		at, e := time.Parse(time.RFC3339Nano, m.Capture.CapturedAt)
		if e != nil {
			return nil, errors.New("invalid capture time in catalog")
		}
		environment := m.Capture.Question.Explore.Environment
		if q.Environment != "" && q.Environment != environment || q.From != "" && at.Before(from) || q.To != "" && !at.Before(to) || q.Tag != "" && !tags[m.ID][q.Tag] {
			continue
		}
		services := map[string]bool{}
		for _, member := range m.Members {
			var row struct {
				Service *string `json:"ServiceName"`
			}
			if e = json.Unmarshal(member.Row, &row); e != nil {
				return nil, e
			}
			if row.Service != nil {
				services[*row.Service] = true
			}
		}
		if q.Service != "" && !services[q.Service] {
			continue
		}
		item := CatalogItem{ID: m.ID, CapturedAt: m.Capture.CapturedAt, Environment: environment, Services: []string{}, Tags: []string{}, Observations: len(m.Members), Origin: m.Origin}
		for name := range services {
			item.Services = append(item.Services, name)
		}
		for tag := range tags[m.ID] {
			item.Tags = append(item.Tags, tag)
		}
		sort.Strings(item.Services)
		sort.Strings(item.Tags)
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	filter := q
	filter.Cursor = ""
	raw, _ := json.Marshal(struct {
		Query CatalogQuery
		Items []CatalogItem
	}{filter, items})
	fingerprint := hash(raw)
	cursor := struct {
		After       string
		Fingerprint string
	}{}
	if q.Cursor != "" {
		raw, e := base64.RawURLEncoding.DecodeString(q.Cursor)
		if e != nil || json.Unmarshal(raw, &cursor) != nil || !workbench.ID.MatchString(cursor.After) || cursor.Fingerprint != fingerprint {
			return nil, errors.New("catalog changed or filter differs; restart browsing")
		}
	}
	page := &CatalogPage{Items: []CatalogItem{}}
	for _, item := range items {
		if item.ID <= cursor.After {
			continue
		}
		if len(page.Items) == 20 {
			cursor.After = page.Items[19].ID
			cursor.Fingerprint = fingerprint
			raw, _ := json.Marshal(cursor)
			page.NextCursor = base64.RawURLEncoding.EncodeToString(raw)
			break
		}
		page.Items = append(page.Items, item)
	}
	return page, nil
}

func (s Store) Annotation(ctx context.Context, id, revision string) (*workbench.MeaningRevision, error) {
	if !workbench.ID.MatchString(id) || !workbench.ID.MatchString(revision) {
		return nil, errors.New("exact evidence and review IDs required")
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	if _, e = load(ctx, c, id); e != nil {
		return nil, e
	}
	var raw string
	if e = c.QueryRow(ctx, "SELECT payload FROM goo_annotations WHERE _id=$1 AND target=$2", revision, id).Scan(&raw); e != nil {
		return nil, e
	}
	var r workbench.MeaningRevision
	if json.Unmarshal([]byte(raw), &r) != nil || r.Target.ID != id || r.Target.Kind != "evidence" || workbench.ValidateMeaningRevision(r) != nil {
		return nil, errors.New("invalid pinned review")
	}
	return &r, nil
}

package retention

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/casselc/goo/workbench"
)

// EvidenceCatalogQuery explicitly selects an origin, or all supported origins.
// From/To apply to each item's named clock: capture time for v1, earliest recorded
// custody timestamp for tail evidence. They never filter telemetry event timestamps.
// This backend catalog is not yet mounted on the legacy HTTP/UI catalog route.
type EvidenceCatalogQuery struct {
	CatalogQuery
	Origin string `json:"origin,omitempty"`
}
type EvidenceCatalogItem struct {
	ID           string   `json:"id"`
	Origin       string   `json:"origin"`
	At           string   `json:"at"`
	Clock        string   `json:"clock"`
	Environment  string   `json:"environment"`
	Services     []string `json:"services"`
	Tags         []string `json:"tags"`
	Observations int      `json:"observations"`
}
type EvidenceCatalogPage struct {
	Items      []EvidenceCatalogItem `json:"items"`
	NextCursor string                `json:"nextCursor,omitempty"`
}

func (q EvidenceCatalogQuery) Validate() error {
	if e := q.CatalogQuery.Validate(); e != nil {
		return e
	}
	if q.Origin != "" && q.Origin != "saved-capture" && q.Origin != "result-page" && q.Origin != "tail-selection" {
		return errors.New("unsupported evidence catalog origin")
	}
	return nil
}
func (s Store) CatalogEvidence(ctx context.Context, q EvidenceCatalogQuery) (*EvidenceCatalogPage, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	evidence := []Evidence{}
	bytes := 0
	// The physical table names are fixed origin adapters, never user identifiers.
	for _, query := range []struct{ sql, origin string }{
		{"SELECT payload FROM goo_evidence ORDER BY _id LIMIT 1001", "capture"},
		{"SELECT payload FROM goo_tail_evidence ORDER BY _id LIMIT 1001", "tail-selection"},
	} {
		rows, e := c.Query(ctx, query.sql)
		if e != nil {
			return nil, e
		}
		for rows.Next() {
			var raw string
			if e = rows.Scan(&raw); e != nil {
				rows.Close()
				return nil, e
			}
			bytes += len(raw)
			if bytes > 16<<20 || len(evidence) >= 1000 {
				rows.Close()
				return nil, errors.New("evidence catalog exceeds scan bound; no partial catalog returned")
			}
			v := Evidence{Origin: query.origin}
			if query.origin == "capture" {
				v.Capture = &Manifest{}
				e = json.Unmarshal([]byte(raw), v.Capture)
				v.Origin = "saved-capture"
				if v.Capture.Origin == "result-page" {
					v.Origin = "result-page"
				}
			} else {
				v.Tail = &TailManifest{}
				e = json.Unmarshal([]byte(raw), v.Tail)
			}
			if e != nil || v.Validate() != nil {
				rows.Close()
				return nil, errors.New("invalid evidence catalog manifest")
			}
			evidence = append(evidence, v)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return nil, e
		}
	}
	receipts, e := tailReceipts(ctx, c, "")
	if e != nil {
		return nil, e
	}
	revisions := []workbench.MeaningRevision{}
	for _, r := range receipts {
		revisions = append(revisions, *r.Annotation)
	}
	rows, e := c.Query(ctx, "SELECT payload FROM goo_annotations ORDER BY _id LIMIT 10001")
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var raw string
		if e = rows.Scan(&raw); e != nil {
			return nil, e
		}
		count++
		bytes += len(raw)
		if bytes > 64<<20 || count > 10000 {
			return nil, errors.New("review catalog exceeds scan bound")
		}
		var r workbench.MeaningRevision
		if json.Unmarshal([]byte(raw), &r) != nil || workbench.ValidateMeaningRevision(r) != nil || r.Target.Kind != "evidence" {
			return nil, errors.New("invalid catalog review")
		}
		revisions = append(revisions, r)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	return evidenceCatalogPage(q, evidence, receipts, revisions)
}
func evidenceCatalogPage(q EvidenceCatalogQuery, evidence []Evidence, receipts []TailReceipt, revisions []workbench.MeaningRevision) (*EvidenceCatalogPage, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	reviews := map[string][]workbench.MeaningRevision{}
	for _, r := range revisions {
		reviews[r.Target.ID] = append(reviews[r.Target.ID], r)
	}
	custody := map[string]time.Time{}
	for _, r := range receipts {
		if r.Manifest.Validate() != nil || r.RetainedAt.IsZero() {
			return nil, errors.New("invalid tail catalog custody")
		}
		id := r.Manifest.ID
		if old, ok := custody[id]; !ok || r.RetainedAt.Before(old) {
			custody[id] = r.RetainedAt
		}
	}
	from, _ := time.Parse(time.RFC3339Nano, q.From)
	to, _ := time.Parse(time.RFC3339Nano, q.To)
	items := []EvidenceCatalogItem{}
	seen := map[string]bool{}
	for _, v := range evidence {
		if e := v.Validate(); e != nil {
			return nil, e
		}
		id := v.ID()
		if seen[id] {
			return nil, errors.New("ambiguous catalog evidence origin")
		}
		seen[id] = true
		item := EvidenceCatalogItem{ID: id, Origin: v.Origin, Services: []string{}, Tags: []string{}}
		records := []json.RawMessage{}
		if v.Capture != nil {
			item.At = v.Capture.Capture.CapturedAt
			item.Clock = "captured-at"
			item.Environment = v.Capture.Capture.Question.Explore.Environment
			for _, m := range v.Capture.Members {
				records = append(records, m.Row)
			}
		} else {
			at, ok := custody[id]
			if !ok {
				return nil, errors.New("tail custody receipt unavailable")
			}
			item.At = at.UTC().Format(time.RFC3339Nano)
			item.Clock = "retained-at"
			item.Environment = v.Tail.Selection.Selector.Environment
			for _, m := range v.Tail.Selection.Members {
				records = append(records, m.Record)
			}
		}
		at, e := time.Parse(time.RFC3339Nano, item.At)
		if e != nil {
			return nil, errors.New("invalid catalog clock")
		}
		page, e := reviewPage(id, reviews[id])
		if e != nil {
			return nil, e
		}
		tags := map[string]bool{}
		services, e := retainedServices(records)
		if e != nil {
			return nil, e
		}
		for _, r := range page.Items {
			if r.Head && r.Revision.State == "active" {
				for _, tag := range r.Revision.Tags {
					tags[tag] = true
				}
			}
		}

		if q.Origin != "" && q.Origin != item.Origin || q.Environment != "" && q.Environment != item.Environment || q.Tag != "" && !tags[q.Tag] || q.Service != "" && !services[q.Service] || q.From != "" && at.Before(from) || q.To != "" && !at.Before(to) {
			continue
		}
		for tag := range tags {
			item.Tags = append(item.Tags, tag)
		}
		sort.Strings(item.Tags)
		for service := range services {
			item.Services = append(item.Services, service)
		}
		sort.Strings(item.Services)
		item.Observations = len(records)
		items = append(items, item)
	}
	filter := q
	filter.Cursor = ""
	page, next, e := paginateCatalog(filter, q.Cursor, items, func(item EvidenceCatalogItem) string { return item.ID })
	if e != nil {
		return nil, e
	}
	return &EvidenceCatalogPage{Items: page, NextCursor: next}, nil
}

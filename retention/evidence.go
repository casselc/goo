package retention

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"

	"github.com/casselc/goo/workbench"
	"github.com/jackc/pgx/v5"
)

// Evidence is an origin-aware read view. It does not rewrite either immutable
// manifest or manufacture a capture/question for tail observations.
type Evidence struct {
	Origin  string        `json:"origin"`
	Capture *Manifest     `json:"capture,omitempty"`
	Tail    *TailManifest `json:"tail,omitempty"`
}

func (v Evidence) Validate() error {
	switch v.Origin {
	case "saved-capture", "result-page":
		if v.Capture == nil || v.Tail != nil {
			break
		}
		origin := "saved-capture"
		if v.Capture.Origin == "result-page" {
			origin = "result-page"
		}
		if origin != v.Origin {
			break
		}
		return v.Capture.Validate()
	case "tail-selection":
		if v.Tail == nil || v.Capture != nil {
			break
		}
		return v.Tail.Validate()
	}
	return errors.New("invalid evidence origin view")
}
func (v Evidence) ID() string {
	if v.Capture != nil {
		return v.Capture.ID
	}
	if v.Tail != nil {
		return v.Tail.ID
	}
	return ""
}
func (s Store) ReadEvidence(ctx context.Context, id string) (*Evidence, error) {
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	return readEvidence(ctx, c, id)
}
func readEvidence(ctx context.Context, c *pgx.Conn, id string) (*Evidence, error) {
	if !workbench.ID.MatchString(id) {
		return nil, errors.New("invalid evidence ID")
	}
	// Only absence permits trying the other origin. Corruption or unavailable
	// storage must never be converted into a successful fallback or a missing ID.
	capture, ce := load(ctx, c, id)
	if ce != nil && !errors.Is(ce, pgx.ErrNoRows) {
		return nil, ce
	}
	tail, te := loadTail(ctx, c, id)
	if te != nil && !errors.Is(te, pgx.ErrNoRows) {
		return nil, te
	}
	if capture != nil && tail != nil {
		return nil, errors.New("ambiguous evidence origin")
	}
	if capture != nil {
		origin := "saved-capture"
		if capture.Origin == "result-page" {
			origin = "result-page"
		}
		return &Evidence{Origin: origin, Capture: capture}, nil
	}
	if tail != nil {
		return &Evidence{Origin: "tail-selection", Tail: tail}, nil
	}
	return nil, pgx.ErrNoRows
}

// Tail receipts published before shared reviews existed are read in place. New
// requests carry a manifest projection; missing projections remain bounded and
// decoded, never silently skipped. No read migrates or rewrites old receipts.
func tailReceipts(ctx context.Context, c *pgx.Conn, id string) ([]TailReceipt, error) {
	query := "SELECT payload FROM goo_tail_requests ORDER BY _id LIMIT 1001"
	args := []any{}
	if id != "" {
		query = "SELECT payload FROM goo_tail_requests WHERE manifest_id=$1 OR manifest_id IS NULL ORDER BY _id LIMIT 1001"
		args = append(args, id)
	}
	rows, e := c.Query(ctx, query, args...)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := []TailReceipt{}
	count, bytes := 0, 0
	for rows.Next() {
		var raw string
		if e = rows.Scan(&raw); e != nil {
			return nil, e
		}
		count++
		bytes += len(raw)
		if count > 1000 || bytes > 16<<20 {
			return nil, errors.New("tail review receipts exceed scan bound; no partial history returned")
		}
		r, e := decodeTailReceipt(raw)
		if e != nil {
			return nil, e
		}
		if id == "" || r.Manifest.ID == id {
			out = append(out, *r)
		}
	}
	return out, rows.Err()
}
func tailRoots(ctx context.Context, c *pgx.Conn, id string) ([]workbench.MeaningRevision, error) {
	receipts, e := tailReceipts(ctx, c, id)
	if e != nil {
		return nil, e
	}
	roots := make([]workbench.MeaningRevision, 0, len(receipts))
	for _, r := range receipts {
		roots = append(roots, *r.Annotation)
	}
	return roots, nil
}
func reviewPage(id string, all []workbench.MeaningRevision) (*workbench.MeaningPage, error) {
	unique := map[string]workbench.MeaningRevision{}
	parents := map[string]bool{}
	for _, r := range all {
		if r.Target.Kind != "evidence" || r.Target.ID != id || workbench.ValidateMeaningRevision(r) != nil {
			return nil, errors.New("invalid evidence review")
		}
		if previous, ok := unique[r.ID]; ok {
			a, _ := json.Marshal(previous)
			b, _ := json.Marshal(r)
			if string(a) != string(b) {
				return nil, errors.New("conflicting review identity")
			}
			continue
		}
		unique[r.ID] = r
	}
	if len(unique) > 1000 {
		return nil, errors.New("annotation history exceeds prototype bound")
	}
	for _, r := range unique {
		if r.Parent != "" {
			if _, ok := unique[r.Parent]; !ok {
				return nil, errors.New("review parent unavailable")
			}
			parents[r.Parent] = true
		}
	}
	// Keep stable ID order for shared readers and reproducible pinned documents.
	ids := make([]string, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	page := &workbench.MeaningPage{Items: []workbench.MeaningItem{}}
	for _, key := range ids {
		r := unique[key]
		page.Items = append(page.Items, workbench.MeaningItem{Revision: &r, Head: !parents[key]})
	}
	return page, nil
}
func evidenceReviews(ctx context.Context, c *pgx.Conn, id string) (*workbench.MeaningPage, error) {
	evidence, e := readEvidence(ctx, c, id)
	if e != nil {
		return nil, e
	}
	all := []workbench.MeaningRevision{}
	if evidence.Tail != nil {
		all, e = tailRoots(ctx, c, id)
		if e == nil && len(all) == 0 {
			return nil, errors.New("tail custody reviews unavailable")
		}
		if e != nil {
			return nil, e
		}
	}
	rows, e := c.Query(ctx, "SELECT payload FROM goo_annotations WHERE target=$1 ORDER BY _id LIMIT 1001", id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	count, bytes := 0, 0
	for rows.Next() {
		var raw string
		if e = rows.Scan(&raw); e != nil {
			return nil, e
		}
		count++
		bytes += len(raw)
		if count > 1000 || bytes > 16<<20 {
			return nil, errors.New("annotation history exceeds prototype bound")
		}
		var r workbench.MeaningRevision
		if json.Unmarshal([]byte(raw), &r) != nil {
			return nil, errors.New("invalid evidence review")
		}
		all = append(all, r)
	}
	if e = rows.Err(); e != nil {
		return nil, e
	}
	return reviewPage(id, all)
}

func evidenceAnnotation(ctx context.Context, c *pgx.Conn, id, revision string) (*workbench.MeaningRevision, error) {
	evidence, e := readEvidence(ctx, c, id)
	if e != nil {
		return nil, e
	}
	var raw string
	e = c.QueryRow(ctx, "SELECT payload FROM goo_annotations WHERE _id=$1 AND target=$2", revision, id).Scan(&raw)
	if e == nil {
		var r workbench.MeaningRevision
		if json.Unmarshal([]byte(raw), &r) != nil || r.ID != revision || r.Target.Kind != "evidence" || r.Target.ID != id || workbench.ValidateMeaningRevision(r) != nil {
			return nil, errors.New("invalid pinned review")
		}
		return &r, nil
	}
	if !errors.Is(e, pgx.ErrNoRows) {
		return nil, e
	}
	if evidence.Tail != nil {
		roots, e := tailRoots(ctx, c, id)
		if e != nil {
			return nil, e
		}
		for _, r := range roots {
			if r.ID == revision {
				return &r, nil
			}
		}
	}
	return nil, errors.New("pinned review not found for retained evidence")
}

// EvidenceContent exposes typed immutable content and its exact JSON encoding.
// Recorded tail custody is the earliest receipt timestamp, not ingestion time.
type EvidenceContent struct {
	Evidence
	ManifestJSON string `json:"manifestJSON"`
	RetainedAt   string `json:"retainedAt,omitempty"`
}

func (s Store) Content(ctx context.Context, id string) (*EvidenceContent, error) {
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	v, e := readEvidence(ctx, c, id)
	if e != nil {
		return nil, e
	}
	out := &EvidenceContent{Evidence: *v}
	if v.Capture != nil {
		raw, _ := json.Marshal(v.Capture)
		out.ManifestJSON = string(raw)
	} else {
		raw, _ := json.Marshal(v.Tail)
		out.ManifestJSON = string(raw)
		receipts, e := tailReceipts(ctx, c, id)
		if e != nil {
			return nil, e
		}
		if len(receipts) == 0 {
			return nil, errors.New("tail custody receipt unavailable")
		}
		at := receipts[0].RetainedAt
		for _, r := range receipts {
			if r.RetainedAt.Before(at) {
				at = r.RetainedAt
			}
		}
		out.RetainedAt = at.UTC().Format(time.RFC3339Nano)
	}
	return out, nil
}

// Package retention is the optional XTDB retained-evidence experiment.
package retention

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/jackc/pgx/v5"
)

const MaxBytes = 2 << 20

// The JSON payload is retained as text to preserve exact numeric encodings and
// missing/empty distinctions. Indexed columns are independent projections.
type Member struct {
	Index int             `json:"index"`
	Hash  string          `json:"hash"`
	Row   json.RawMessage `json:"row"`
}
type Manifest struct {
	Version       int                        `json:"version"`
	ID            string                     `json:"id"`
	Capture       exploration.CaptureReceipt `json:"capture"`
	Basis         string                     `json:"basis"`
	Count         string                     `json:"sourceCount"`
	StoredRecords int                        `json:"storedRecords"`
	Coverage      string                     `json:"coverage"`
	Members       []Member                   `json:"members"`
}
type Receipt struct {
	Manifest     Manifest                   `json:"manifest"`
	Annotation   *workbench.MeaningRevision `json:"annotation"`
	ManifestJSON string                     `json:"manifestJSON"`
	AwaitToken   string                     `json:"awaitToken"`
}
type Store struct {
	URL          string
	beforeCommit func() error
}

func hash(b []byte) string { v := sha256.Sum256(b); return hex.EncodeToString(v[:]) }
func Select(c *exploration.SavedCapture, indexes []int) (Manifest, error) {
	m := Manifest{Version: 1}
	if c == nil || c.Explore == nil || len(indexes) < 1 || len(indexes) > 20 {
		return m, errors.New("select 1..20 rows from a saved record capture")
	}
	indexes = append([]int{}, indexes...)
	sort.Ints(indexes)
	receipt, _ := json.Marshal(c.CaptureReceipt)
	if e := json.Unmarshal(receipt, &m.Capture); e != nil {
		return Manifest{}, e
	}
	m.Basis = c.Explore.Basis
	m.Count = c.Explore.Count
	m.StoredRecords = len(c.Explore.Rows)
	m.Coverage = "selected stored observations; population completeness unknown"
	for i, index := range indexes {
		if index < 0 || index >= len(c.Explore.Rows) || i > 0 && indexes[i-1] == index {
			return m, errors.New("duplicate or unavailable capture row")
		}
		var compact bytes.Buffer
		if e := json.Compact(&compact, c.Explore.Rows[index]); e != nil {
			return Manifest{}, e
		}
		b := append(json.RawMessage{}, compact.Bytes()...)
		m.Members = append(m.Members, Member{index, hash(b), b})
	}
	raw, e := json.Marshal(m)
	if e != nil || len(raw) > MaxBytes {
		return Manifest{}, errors.New("selected evidence exceeds 2 MiB")
	}
	m.ID = hash(raw)
	return m, nil
}
func (m Manifest) Validate() error {
	id := m.ID
	m.ID = ""
	raw, e := json.Marshal(m)
	if e != nil || len(raw) > MaxBytes || hash(raw) != id || m.Version != 1 || !workbench.ID.MatchString(m.Capture.ID) || m.Capture.Question.Validate() != nil || len(m.Members) < 1 || len(m.Members) > 20 {
		return errors.New("invalid retained manifest")
	}
	last := -1
	for _, v := range m.Members {
		if v.Index <= last || v.Index >= m.StoredRecords || hash(v.Row) != v.Hash || !json.Valid(v.Row) {
			return errors.New("invalid retained member")
		}
		last = v.Index
	}
	return nil
}
func (s Store) connect(ctx context.Context) (*pgx.Conn, error) {
	config, e := pgx.ParseConfig(s.URL)
	if e != nil {
		return nil, errors.New("invalid XTDB connection configuration")
	}
	config.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	c, e := pgx.ConnectConfig(ctx, config)
	if e != nil {
		return nil, errors.New("XTDB unavailable")
	}
	return c, nil
}
func (s Store) Publish(ctx context.Context, m Manifest, tags []string, note string, requestID string) (*Receipt, error) {
	if e := m.Validate(); e != nil {
		return nil, e
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	if !workbench.ID.MatchString(requestID) {
		return nil, errors.New("request ID must be 64 lowercase hex characters")
	}
	for _, ddl := range []string{"CREATE TABLE goo_evidence (_id,payload,retained_at)", "CREATE TABLE goo_annotations (_id,target,parent,state,payload)", "CREATE TABLE goo_evidence_members (_id,manifest_id,ordinal,payload,service,trace_id,span_id,event_time_ns)", "CREATE TABLE goo_capture_requests (_id,intent,payload)"} {
		if _, e = c.Exec(ctx, ddl); e != nil {
			return nil, e
		}
	}
	intentBytes, _ := json.Marshal(struct {
		Manifest Manifest
		Tags     []string
		Note     string
	}{m, tags, note})
	intent := hash(intentBytes)
	existing := func() (*Receipt, error) {
		var previous, raw string
		e := c.QueryRow(ctx, "SELECT intent,payload FROM goo_capture_requests WHERE _id=$1", requestID).Scan(&previous, &raw)
		if e != nil {
			return nil, e
		}
		if previous != intent {
			return nil, errors.New("request ID already used for another capture intent")
		}
		var r Receipt
		if json.Unmarshal([]byte(raw), &r) != nil {
			return nil, errors.New("invalid request receipt")
		}
		if r.ManifestJSON == "" {
			encoded, _ := json.Marshal(r.Manifest)
			r.ManifestJSON = string(encoded)
		}
		if e = c.QueryRow(ctx, "SHOW AWAIT_TOKEN").Scan(&r.AwaitToken); e != nil {
			return nil, e
		}
		return &r, nil
	}
	if previous, err := existing(); err == nil {
		return previous, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	r, e := workbench.PrepareMeaning(workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "evidence", ID: m.ID}, Tags: tags, Note: note, State: "active"})
	if e != nil {
		return nil, e
	}
	payload, _ := json.Marshal(m)
	assertion, _ := json.Marshal(r)
	tx, e := c.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	// The request and manifest checks are evaluated inside the serial transaction.
	if _, e = tx.Exec(ctx, "ASSERT NOT EXISTS (SELECT 1 FROM goo_capture_requests WHERE _id=$1)", requestID); e != nil {
		return nil, e
	}
	// ASSERT makes same-ID retries immutable rather than overwriting earlier evidence.
	if _, e = tx.Exec(ctx, `ASSERT NOT EXISTS (SELECT 1 FROM goo_evidence WHERE _id=$1 AND payload<>$2)`, m.ID, string(payload)); e != nil {
		return nil, e
	}
	if _, e = tx.Exec(ctx, `INSERT INTO goo_evidence (_id,payload,retained_at) SELECT $1,$2,$3 WHERE NOT EXISTS (SELECT 1 FROM goo_evidence WHERE _id=$1)`, m.ID, string(payload), time.Now().UTC()); e != nil {
		return nil, e
	}
	for _, member := range m.Members {
		var record struct {
			Service string `json:"ServiceName"`
			Trace   string `json:"TraceId"`
			Span    string `json:"SpanId"`
			Time    string `json:"timestamp_ns"`
		}
		if e = json.Unmarshal(member.Row, &record); e != nil {
			return nil, e
		}
		memberID := hash([]byte(m.ID + "/" + strconv.Itoa(member.Index)))
		if _, e = tx.Exec(ctx, `INSERT INTO goo_evidence_members (_id,manifest_id,ordinal,payload,service,trace_id,span_id,event_time_ns) SELECT $1,$2,$3,$4,$5,$6,$7,$8 WHERE NOT EXISTS (SELECT 1 FROM goo_evidence_members WHERE _id=$1)`, memberID, m.ID, member.Index, string(member.Row), record.Service, record.Trace, record.Span, record.Time); e != nil {
			return nil, e
		}
	}
	if _, e = tx.Exec(ctx, `INSERT INTO goo_annotations (_id,target,parent,state,payload) VALUES ($1,$2,$3,$4,$5)`, r.ID, m.ID, r.Parent, r.State, string(assertion)); e != nil {
		return nil, e
	}
	receipt := Receipt{Manifest: m, Annotation: r, ManifestJSON: string(payload)}
	receiptJSON, _ := json.Marshal(receipt)
	if _, e = tx.Exec(ctx, "INSERT INTO goo_capture_requests (_id,intent,payload) VALUES ($1,$2,$3)", requestID, intent, string(receiptJSON)); e != nil {
		return nil, e
	}
	if s.beforeCommit != nil {
		if e = s.beforeCommit(); e != nil {
			return nil, e
		}
	}
	if e = tx.Commit(ctx); e != nil {
		if previous, err := existing(); err == nil {
			return previous, nil
		}
		return nil, e
	}
	var token string
	if e = c.QueryRow(ctx, "SHOW AWAIT_TOKEN").Scan(&token); e != nil {
		return nil, e
	}
	return &Receipt{Manifest: m, Annotation: r, ManifestJSON: string(payload), AwaitToken: token}, nil
}
func (s Store) Load(ctx context.Context, id string) (*Manifest, error) {
	if !workbench.ID.MatchString(id) {
		return nil, errors.New("invalid evidence ID")
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	return load(ctx, c, id)
}
func load(ctx context.Context, c *pgx.Conn, id string) (*Manifest, error) {
	var raw string
	if e := c.QueryRow(ctx, `SELECT payload FROM goo_evidence WHERE _id=$1`, id).Scan(&raw); e != nil {
		return nil, e
	}
	var m Manifest
	if len(raw) > MaxBytes || json.Unmarshal([]byte(raw), &m) != nil || m.ID != id {
		return nil, errors.New("invalid retained manifest")
	}
	if e := m.Validate(); e != nil {
		return nil, e
	}
	return &m, nil
}
func (s Store) Annotations(ctx context.Context, id string) (*workbench.MeaningPage, error) {
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	if _, e = load(ctx, c, id); e != nil {
		return nil, e
	}
	rows, e := c.Query(ctx, `SELECT payload FROM goo_annotations WHERE target=$1 ORDER BY _id LIMIT 1001`, id)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	all := []*workbench.MeaningRevision{}
	parents := map[string]bool{}
	for rows.Next() {
		var raw string
		if e = rows.Scan(&raw); e != nil {
			return nil, e
		}
		var r workbench.MeaningRevision
		if json.Unmarshal([]byte(raw), &r) != nil || r.Target.ID != id || r.Target.Kind != "evidence" || workbench.ValidateMeaningRevision(r) != nil {
			return nil, errors.New("invalid assertion")
		}
		all = append(all, &r)
		parents[r.Parent] = true
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(all) > 1000 {
		return nil, errors.New("annotation history exceeds prototype bound")
	}
	p := &workbench.MeaningPage{Items: []workbench.MeaningItem{}}
	for _, r := range all {
		p.Items = append(p.Items, workbench.MeaningItem{Revision: r, Head: !parents[r.ID]})
	}
	return p, nil
}
func (s Store) Correct(ctx context.Context, id string, in workbench.MeaningInput) (*workbench.MeaningRevision, error) {
	if in.Target.Kind != "evidence" || in.Target.ID != id {
		return nil, errors.New("annotation requires exact retained-evidence target")
	}
	p, e := s.Annotations(ctx, id)
	if e != nil {
		return nil, e
	}
	found := false
	for _, v := range p.Items {
		found = found || v.Revision.ID == in.Parent
	}
	if in.Parent != "" && !found {
		return nil, errors.New("parent not found for retained evidence")
	}
	r, e := workbench.PrepareMeaning(in)
	if e != nil {
		return nil, e
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	raw, _ := json.Marshal(r)
	_, e = c.Exec(ctx, `INSERT INTO goo_annotations (_id,target,parent,state,payload) VALUES ($1,$2,$3,$4,$5)`, r.ID, id, r.Parent, r.State, string(raw))
	if e != nil {
		return nil, fmt.Errorf("publish correction: %w", e)
	}
	return r, nil
}

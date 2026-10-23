package retention

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/jackc/pgx/v5"
)

const TailEvidenceVersion = "goo/tail-evidence/v1alpha1"

var ErrTailIntentConflict = errors.New("request ID already used for another tail intent")

// TailIntent contains references only. PublishTail never accepts client records.
// This library is deliberately not yet exposed through HTTP or SSE.
type TailIntent struct {
	Version    string                          `json:"version"`
	RequestID  string                          `json:"requestID"`
	Connection string                          `json:"connection"`
	Selector   exploration.TailSelector        `json:"selector"`
	Refs       []exploration.TailMembershipRef `json:"refs"`
	Tags       []string                        `json:"tags"`
	Note       string                          `json:"note"`
}
type TailManifest struct {
	Version   string                         `json:"version"`
	ID        string                         `json:"id"`
	Selection exploration.TailSelectionDraft `json:"selection"`
}
type TailReceipt struct {
	Manifest   TailManifest               `json:"manifest"`
	Annotation *workbench.MeaningRevision `json:"annotation"`
	RetainedAt time.Time                  `json:"retainedAt"`
	AwaitToken string                     `json:"awaitToken,omitempty"`
}

func (m TailManifest) Validate() error {
	id := m.ID
	m.ID = ""
	raw, e := json.Marshal(m)
	if e != nil || len(raw) > MaxBytes || m.Version != TailEvidenceVersion || hash(raw) != id {
		return errors.New("invalid tail manifest")
	}
	return m.Selection.Validate()
}
func normalizeTailIntent(in TailIntent) (TailIntent, string, error) {
	raw, e := json.Marshal(in)
	if e != nil || len(raw) > 16384 {
		return in, "", errors.New("tail intent exceeds bound")
	}
	// Own all nested request data before sorting or holding it across database IO.
	var owned TailIntent
	if e = json.Unmarshal(raw, &owned); e != nil {
		return in, "", e
	}
	in = owned
	if in.Selector.Signal == "" {
		in.Selector.Signal = "logs"
	}
	if in.Version != TailEvidenceVersion || !workbench.ID.MatchString(in.RequestID) || !workbench.ID.MatchString(in.Connection) || len(in.Refs) < 1 || len(in.Refs) > 20 {
		return in, "", errors.New("invalid tail intent")
	}
	if e = in.Selector.Query(60000).Validate(); e != nil {
		return in, "", e
	}
	if e = (workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "evidence", ID: in.RequestID}, Tags: in.Tags, Note: in.Note, State: "active"}).Validate(); e != nil {
		return in, "", e
	}
	sort.Slice(in.Refs, func(i, j int) bool {
		if in.Refs[i].Token != in.Refs[j].Token {
			return in.Refs[i].Token < in.Refs[j].Token
		}
		return in.Refs[i].Key < in.Refs[j].Key
	})
	seen := map[string]bool{}
	for _, r := range in.Refs {
		prefix, suffix, found := strings.Cut(r.Key, ":")
		ordinal, parseErr := strconv.Atoi(suffix)
		if !found || !workbench.ID.MatchString(prefix) || parseErr != nil || ordinal < 1 || ordinal > exploration.TailReadLimit || suffix != strconv.Itoa(ordinal) {
			return in, "", errors.New("invalid tail occurrence handle")
		}
		if !workbench.ID.MatchString(r.Token) || len(r.Key) > 72 || len(r.Key) < 66 || seen[r.Key] {
			return in, "", errors.New("invalid tail reference")
		}
		seen[r.Key] = true
	}
	raw, _ = json.Marshal(in)
	return in, hash(raw), nil
}
func decodeTailReceipt(raw string) (*TailReceipt, error) {
	var r TailReceipt
	if len(raw) > MaxBytes+16384 || json.Unmarshal([]byte(raw), &r) != nil || r.Manifest.Validate() != nil || r.Annotation == nil || workbench.ValidateMeaningRevision(*r.Annotation) != nil || r.Annotation.Target.Kind != "evidence" || r.Annotation.Target.ID != r.Manifest.ID || r.RetainedAt.IsZero() {
		return nil, errors.New("invalid tail receipt")
	}
	for _, member := range r.Manifest.Selection.Members {
		if member.ObservedAt.After(r.RetainedAt) {
			return nil, errors.New("tail read clock exceeds custody clock")
		}
	}
	return &r, nil
}

// PublishTail recovers a committed request BEFORE consulting ephemeral proof.
// Dedicated tables prevent v1 catalog/readers from decoding a tail as a capture.
func (s Store) PublishTail(ctx context.Context, bank *exploration.TailMembershipSnapshots, input TailIntent) (*TailReceipt, error) {
	in, intent, e := normalizeTailIntent(input)
	if e != nil {
		return nil, e
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	for _, ddl := range []string{"CREATE TABLE goo_tail_evidence (_id,payload)", "CREATE TABLE goo_tail_requests (_id,intent,payload)"} {
		if _, e = c.Exec(ctx, ddl); e != nil {
			return nil, e
		}
	}
	existing := func() (*TailReceipt, error) {
		var previous, raw string
		if e := c.QueryRow(ctx, "SELECT intent,payload FROM goo_tail_requests WHERE _id=$1", in.RequestID).Scan(&previous, &raw); e != nil {
			return nil, e
		}
		if previous != intent {
			return nil, ErrTailIntentConflict
		}
		r, e := decodeTailReceipt(raw)
		if e != nil {
			return nil, e
		}
		if e = c.QueryRow(ctx, "SHOW AWAIT_TOKEN").Scan(&r.AwaitToken); e != nil {
			return nil, e
		}
		return r, nil
	}
	if r, e := existing(); e == nil {
		return r, nil
	} else if !errors.Is(e, pgx.ErrNoRows) {
		return nil, e
	}
	if bank == nil {
		return nil, exploration.ErrTailMembershipUnavailable
	}
	draft, e := bank.Resolve(in.Connection, in.Selector, in.Refs)
	if e != nil {
		return nil, e
	}
	m := TailManifest{Version: TailEvidenceVersion, Selection: *draft}
	raw, _ := json.Marshal(m)
	m.ID = hash(raw)
	if e = m.Validate(); e != nil {
		return nil, e
	}
	annotation, e := workbench.PrepareMeaning(workbench.MeaningInput{Target: workbench.MeaningTarget{Kind: "evidence", ID: m.ID}, Tags: in.Tags, Note: in.Note, State: "active"})
	if e != nil {
		return nil, e
	}
	r := &TailReceipt{Manifest: m, Annotation: annotation, RetainedAt: time.Now().UTC()}
	raw, _ = json.Marshal(r)
	if _, e = decodeTailReceipt(string(raw)); e != nil {
		return nil, e
	}
	manifest, _ := json.Marshal(m)
	tx, e := c.Begin(ctx)
	if e != nil {
		return nil, e
	}
	defer tx.Rollback(ctx)
	for _, statement := range []struct {
		sql  string
		args []any
	}{
		{"ASSERT NOT EXISTS (SELECT 1 FROM goo_tail_requests WHERE _id=$1)", []any{in.RequestID}},
		{"ASSERT NOT EXISTS (SELECT 1 FROM goo_tail_evidence WHERE _id=$1 AND payload<>$2)", []any{m.ID, string(manifest)}},
		{"INSERT INTO goo_tail_evidence (_id,payload) SELECT $1,$2 WHERE NOT EXISTS (SELECT 1 FROM goo_tail_evidence WHERE _id=$1)", []any{m.ID, string(manifest)}},
		{"INSERT INTO goo_tail_requests (_id,intent,payload) VALUES ($1,$2,$3)", []any{in.RequestID, intent, string(raw)}},
	} {
		if _, e = tx.Exec(ctx, statement.sql, statement.args...); e != nil {
			return nil, e
		}
	}
	if s.beforeCommit != nil {
		if e = s.beforeCommit(); e != nil {
			return nil, e
		}
	}
	if e = tx.Commit(ctx); e != nil {
		if r, recovery := existing(); recovery == nil {
			return r, nil
		} else if errors.Is(recovery, ErrTailIntentConflict) {
			return nil, recovery
		}
		return nil, e
	}
	if s.afterTailCommit != nil {
		if e = s.afterTailCommit(); e != nil {
			return nil, e
		}
	}
	if e = c.QueryRow(ctx, "SHOW AWAIT_TOKEN").Scan(&r.AwaitToken); e != nil {
		return nil, e
	}
	return r, nil
}

// LoadTail resolves only the explicit tail origin; it cannot enter v1 readers.
func (s Store) LoadTail(ctx context.Context, id string) (*TailManifest, error) {
	if !workbench.ID.MatchString(id) {
		return nil, errors.New("invalid tail evidence ID")
	}
	c, e := s.connect(ctx)
	if e != nil {
		return nil, e
	}
	defer c.Close(ctx)
	var raw string
	if e = c.QueryRow(ctx, "SELECT payload FROM goo_tail_evidence WHERE _id=$1", id).Scan(&raw); e != nil {
		return nil, e
	}
	var m TailManifest
	if len(raw) > MaxBytes || json.Unmarshal([]byte(raw), &m) != nil || m.ID != id {
		return nil, errors.New("invalid tail manifest")
	}
	if e = m.Validate(); e != nil {
		return nil, e
	}
	return &m, nil
}

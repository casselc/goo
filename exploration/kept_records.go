package exploration

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"time"
)

const KeepRecordsVersion = "goo/keep-records/v1"
const KeptRecordsVersion = "goo/kept-records/v1"

var ErrKeepPrecondition = errors.New("original record scope differs from keep precondition")

var ErrKeptRecordsUnavailable = errors.New("original record selection expired or unavailable; no telemetry was reread")

// Indexes are local to a result-token page or absolute in a saved capture's row bank.
// Browser record contents are never accepted. Optional expected scope/basis is
// an optimistic precondition, not authority for the retained content.
type KeepRecordExpectation struct {
	Question Query  `json:"question"`
	Basis    string `json:"basis"`
}
type KeepRecordsIntent struct {
	Expected    *KeepRecordExpectation `json:"expected,omitempty"`
	Version     string                 `json:"version"`
	RequestID   string                 `json:"requestID"`
	ResultToken string                 `json:"resultToken,omitempty"`
	CaptureID   string                 `json:"captureID,omitempty"`
	Indexes     []int                  `json:"indexes"`
}

func (in KeepRecordsIntent) normalized() (KeepRecordsIntent, error) {
	if in.Version != KeepRecordsVersion || !validTailConnection(in.RequestID) || (in.ResultToken == "") == (in.CaptureID == "") || in.ResultToken != "" && !validTailConnection(in.ResultToken) || in.CaptureID != "" && !validTailConnection(in.CaptureID) || len(in.Indexes) < 1 || len(in.Indexes) > 20 {
		return in, errors.New("require one result token or saved capture ID and 1..20 row indexes")
	}
	if in.Expected != nil {
		if e := in.Expected.Question.Validate(); e != nil {
			return in, e
		}
		if !validTailConnection(in.Expected.Basis) {
			return in, errors.New("invalid expected record basis")
		}
		in.Expected = detached(in.Expected)
	}
	in.Indexes = append([]int(nil), in.Indexes...)
	sort.Ints(in.Indexes)
	for i, n := range in.Indexes {
		if n < 0 || n >= 10000 || i > 0 && n == in.Indexes[i-1] {
			return in, errors.New("invalid or duplicate record index")
		}
	}
	return in, nil
}
func (in KeepRecordsIntent) Validate() error { _, e := in.normalized(); return e }

type KeptRecord struct {
	Index  int             `json:"index"`
	Digest string          `json:"digest"`
	Record json.RawMessage `json:"record"`
}
type KeptRecords struct {
	Version         string       `json:"version"`
	ID              string       `json:"id"`
	IntentDigest    string       `json:"intentDigest"`
	Digest          string       `json:"digest"`
	KeptAt          time.Time    `json:"keptAt"`
	Origin          string       `json:"origin"`
	CaptureID       string       `json:"captureID,omitempty"`
	CaptureAt       string       `json:"captureAt,omitempty"`
	Question        Query        `json:"question"`
	ObservedAt      string       `json:"observedAt"`
	Basis           string       `json:"basis"`
	State           string       `json:"state"`
	ReturnedRecords int          `json:"returnedRecords"`
	RecordsLimited  bool         `json:"recordsLimited"`
	Members         []KeptRecord `json:"members"`
}

func (v *KeptRecords) keepIdentity() (string, string) { return v.ID, v.IntentDigest }
func (v KeptRecords) Validate() error {
	if v.Version != KeptRecordsVersion || !validTailConnection(v.ID) || !validTailConnection(v.IntentDigest) || v.KeptAt.IsZero() || len(v.Members) < 1 || len(v.Members) > 20 || v.ReturnedRecords < len(v.Members) || v.ReturnedRecords > 10000 || len(v.Basis) == 0 || len(v.Basis) > 4096 {
		return errors.New("invalid kept records receipt")
	}
	if v.Origin == "saved-capture" {
		if v.ReturnedRecords > CaptureRowLimit {
			return errors.New("saved record bank exceeds bound")
		}
		if !validTailConnection(v.CaptureID) {
			return errors.New("invalid original capture")
		}
		if _, e := time.Parse(time.RFC3339Nano, v.CaptureAt); e != nil {
			return e
		}
	} else if v.Origin != "result-page" || v.CaptureID != "" || v.CaptureAt != "" || v.ReturnedRecords > 100 {
		return errors.New("invalid kept records origin")
	}
	if _, e := time.Parse(time.RFC3339Nano, v.ObservedAt); e != nil {
		return e
	}
	if e := v.Question.Validate(); e != nil {
		return e
	}
	if v.State != "observed" && v.State != "no_data" && v.State != "changed_between_reads" {
		return errors.New("invalid original result state")
	}
	selection, e := v.Question.EnvironmentSelection()
	if e != nil {
		return e
	}
	for i, m := range v.Members {
		if m.Index < 0 || m.Index >= v.ReturnedRecords || i > 0 && m.Index <= v.Members[i-1].Index || m.Digest != tailMembershipHash(json.RawMessage(m.Record)) {
			return errors.New("invalid kept record index/content")
		}
		if e = validateSelectedRecord(selection, v.Question.Signal, v.Question.FromMS, v.Question.ToMS, m.Record); e != nil {
			return e
		}
	}
	digest := v.Digest
	v.Digest = ""
	if !validTailConnection(digest) || tailMembershipHash(v) != digest {
		return errors.New("kept records digest mismatch")
	}
	return nil
}
func (s KeptObservationStore) KeepRecords(ctx context.Context, bank *ResultSnapshots, captures CaptureStore, input KeepRecordsIntent) (*KeptRecords, error) {
	in, e := input.normalized()
	if e != nil {
		return nil, e
	}
	digest := tailMembershipHash(in)
	result, e := s.publish(ctx, in.RequestID, digest, func() (keptResource, error) {
		var c *SavedCapture
		var err error
		origin := "saved-capture"
		if in.ResultToken != "" {
			origin = "result-page"
			if bank == nil {
				return nil, ErrKeptRecordsUnavailable
			}
			c, err = bank.Load(in.ResultToken)
		} else {
			c, err = captures.Load(in.CaptureID)
		}
		if err != nil || c == nil || c.Explore == nil {
			return nil, ErrKeptRecordsUnavailable
		}
		r := c.Explore
		if in.Expected != nil && (in.Expected.Basis != r.Basis || tailMembershipHash(in.Expected.Question) != tailMembershipHash(r.Question)) {
			return nil, ErrKeepPrecondition
		}

		now := time.Now()
		if s.now != nil {
			now = s.now()
		}
		v := &KeptRecords{Version: KeptRecordsVersion, ID: in.RequestID, IntentDigest: digest, KeptAt: now.UTC(), Origin: origin, Question: r.Question, ObservedAt: r.CapturedAt, Basis: r.Basis, State: r.State, ReturnedRecords: len(r.Rows), RecordsLimited: r.HasMore, Members: make([]KeptRecord, 0, len(in.Indexes))}
		if origin == "saved-capture" {
			v.CaptureID = in.CaptureID
			v.CaptureAt = c.CapturedAt
		}
		for _, index := range in.Indexes {
			if index >= len(r.Rows) {
				return nil, ErrKeptRecordsUnavailable
			}
			raw := append(json.RawMessage(nil), r.Rows[index]...)
			v.Members = append(v.Members, KeptRecord{Index: index, Digest: tailMembershipHash(raw), Record: raw})
		}
		v.Digest = tailMembershipHash(v)
		return v, nil
	})
	if e != nil {
		return nil, e
	}
	v, ok := result.(*KeptRecords)
	if !ok {
		return nil, ErrKeepConflict
	}
	return v, nil
}

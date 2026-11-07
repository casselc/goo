package exploration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// CaptureRecordRef identifies an occurrence in the full retained row bank, not
// its position on a displayed page or a physical event identified by span ID.
type CaptureRecordRef struct {
	CaptureID string `json:"captureId"`
	Index     int    `json:"index"`
}

// CapturedRecord is server-resolved evidence, not a user judgment or training label.
// Record preserves stored JSON number tokens; callers own the returned data.
type CapturedRecord struct {
	Reference        CaptureRecordRef `json:"reference"`
	Receipt          CaptureReceipt   `json:"receipt"`
	Record           json.RawMessage  `json:"record"`
	RecordDigest     string           `json:"recordDigest"`
	SourceReadAt     string           `json:"sourceReadAt"`
	Basis            string           `json:"basis"`
	ObservationState string           `json:"observationState"`
	MatchingRecords  string           `json:"matchingRecords"`
	RetainedRecords  int              `json:"retainedRecords"`
	RecordsLimited   bool             `json:"recordsLimited"`
}

// ResolveRecord verifies the immutable capture before selecting one absolute
// occurrence. It reads only the bounded capture store, never telemetry.
func (s CaptureStore) ResolveRecord(ref CaptureRecordRef) (*CapturedRecord, error) {
	if ref.Index < 0 || ref.Index >= CaptureRowLimit {
		return nil, errors.New("record index outside retained capture limit")
	}
	capture, err := s.Load(ref.CaptureID)
	if err != nil {
		return nil, fmt.Errorf("resolve captured record: %w", err)
	}
	if capture.Question.Kind != "explore" || capture.Explore == nil || (capture.Explore.Question.Signal != "logs" && capture.Explore.Question.Signal != "spans") {
		return nil, errors.New("record reference requires an ordinary logs or spans capture")
	}
	result := capture.Explore
	if ref.Index >= len(result.Rows) {
		return nil, errors.New("record index outside retained capture rows")
	}
	raw := bytes.Clone(result.Rows[ref.Index])
	sum := sha256.Sum256(raw)
	return &CapturedRecord{Reference: ref, Receipt: capture.CaptureReceipt, Record: raw,
		RecordDigest: hex.EncodeToString(sum[:]), SourceReadAt: result.CapturedAt,
		Basis: result.Basis, ObservationState: result.State, MatchingRecords: result.Count,
		RetainedRecords: len(result.Rows), RecordsLimited: result.HasMore}, nil
}

package exploration

import (
	"context"
	"fmt"
	"time"
)

// LiveRecordComparison reads a rolling full-population record question derived
// exclusively from the frozen baseline. Inspection rows never define totals.
func (s Source) LiveRecordComparison(ctx context.Context, baseline *SavedCapture, now time.Time) (*SavedCapture, *EvidenceComparison, error) {
	if baseline == nil || baseline.Question.Kind != "explore" || baseline.Question.Explore == nil {
		return nil, nil, fmt.Errorf("live comparison requires a record capture")
	}
	q := *baseline.Question.Explore
	duration := q.ToMS - q.FromMS
	if duration <= 0 || duration > 7200000 {
		return nil, nil, fmt.Errorf("live window must be at most two hours")
	}
	q.ToMS = now.UnixMilli()
	q.FromMS = q.ToMS - duration
	result, err := s.captureRows(ctx, q)
	if err != nil {
		return nil, nil, err
	}
	current := &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Frozen live record comparison", CapturedAt: now.UTC().Format(time.RFC3339Nano), Question: CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: result}
	comparison, err := CompareEvidence(baseline, current)
	return current, comparison, err
}

// FreezeSnapshot publishes the server-owned snapshot without another source read.
func (s CaptureStore) FreezeSnapshot(snapshot *SavedCapture) (*SavedCapture, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("snapshot required")
	}
	c := *snapshot
	c.CapturedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return s.publish(c)
}

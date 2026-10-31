package exploration

import (
	"context"
	"encoding/json"
)

// RecordPageResult is a fresh bounded inspection page and an independently read
// matching count. It grants no aggregate, field-discovery or completeness claim.
type RecordPageResult struct {
	Question    Query             `json:"question"`
	CapturedAt  string            `json:"capturedAt"`
	Basis       string            `json:"basis"`
	Count       string            `json:"count"`
	State       string            `json:"state"`
	Rows        []json.RawMessage `json:"rows"`
	HasMore     bool              `json:"hasMore"`
	ResultToken string            `json:"resultToken,omitempty"`
	Warnings    []string          `json:"warnings"`
}

// FetchRecordPage preserves the validated question but reads only plain count
// and bounded rows. The Result is retained internally for existing selection
// custody; callers expose the narrow RecordPageResult rather than this envelope.
func (s Source) FetchRecordPage(ctx context.Context, q Query) (*Result, error) {
	return s.fetchProfile(ctx, q, q.Limit, true)
}

// RecordPageResponse exposes only the acquired page aspects.
func RecordPageResponse(r *Result) *RecordPageResult {
	return &RecordPageResult{Question: r.Question, CapturedAt: r.CapturedAt, Basis: r.Basis, Count: r.Count, State: r.State, Rows: r.Rows, HasMore: r.HasMore, ResultToken: r.ResultToken, Warnings: r.Warnings}
}

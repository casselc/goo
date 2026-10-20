package exploration

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/casselc/goo/measurements"
)

type LiveCounterInfo struct {
	Policy         string                `json:"policy"`
	EvaluationAt   string                `json:"evaluationAt"`
	SearchQuestion measurements.Question `json:"searchQuestion"`
	SearchReadAt   string                `json:"searchReadAt"`
	SearchBasis    string                `json:"searchBasis"`
	EndpointNS     string                `json:"endpointNS,omitempty"`
	LagNS          string                `json:"lagNS"`
	Reasons        []string              `json:"reasons"`
}

// LiveCounterComparison performs bounded endpoint discovery followed by one exact
// window read. The baseline must identify one complete exact OTel stream. New
// members are not dropped; existing window comparison refuses identity changes.
func (s Source) LiveCounterComparison(ctx context.Context, baseline *SavedCapture, now time.Time) (*SavedCapture, *EvidenceComparison, *LiveCounterInfo, error) {
	if baseline == nil || baseline.Measurement == nil || baseline.Question.Measurement == nil || baseline.Measurement.Counters == nil || len(baseline.Measurement.Counters.Series) != 1 {
		return nil, nil, nil, fmt.Errorf("live counter baseline must contain one exact stream")
	}
	checked, e := CompareEvidence(baseline, baseline)
	if e != nil || checked.State != "comparable" {
		return nil, nil, nil, fmt.Errorf("live counter baseline requires complete validated coverage")
	}
	q := *baseline.Question.Measurement
	from, to := q.BoundsNS()
	a, _ := strconv.ParseInt(from, 10, 64)
	b, _ := strconv.ParseInt(to, 10, 64)
	width := b - a
	search := q
	search.FromNS = ""
	search.ToNS = ""
	search.ToMS = now.UnixMilli()
	horizon := (width+999999)/1000000 + 2*q.Definition.MaxGapMS
	if horizon > 7200000 {
		horizon = 7200000
	}
	search.FromMS = search.ToMS - horizon
	adapter := measurements.Source{Endpoint: s.Endpoint, User: s.User, Password: s.Password, Client: s.Client}
	observations, e := adapter.Run(ctx, search, nil)
	if e != nil {
		return nil, nil, nil, e
	}
	end, e := measurements.LatestCounterEndpoint(observations)
	if e != nil {
		return nil, nil, nil, e
	}
	missingEndpoint := end == 0
	if missingEndpoint {
		end = now.UnixMilli() * 1000000
	}
	q.FromNS = strconv.FormatInt(end-width, 10)
	q.ToNS = strconv.FormatInt(end, 10)
	q.FromMS = (end - width) / 1000000
	q.ToMS = (end + 999999) / 1000000
	if e = q.Validate(); e != nil {
		return nil, nil, nil, e
	}
	result, e := adapter.Run(ctx, q, nil)
	if e != nil {
		return nil, nil, nil, e
	}
	snapshot := &SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Frozen live counter comparison", CapturedAt: now.UTC().Format(time.RFC3339Nano), Question: CaptureQuestion{Kind: "measurement", Measurement: &q}}, Measurement: result}
	comparison, e := CompareEvidence(baseline, snapshot)
	if e != nil {
		return nil, nil, nil, e
	}
	lag := now.UnixNano() - end
	if lag < 0 {
		return nil, nil, nil, fmt.Errorf("counter endpoint exceeds evaluation clock")
	}
	info := &LiveCounterInfo{Policy: "latest-observed-endpoint/v1", EvaluationAt: now.UTC().Format(time.RFC3339Nano), SearchQuestion: search, SearchReadAt: observations.ReadAt, SearchBasis: observations.Basis, LagNS: strconv.FormatInt(lag, 10), Reasons: []string{}}
	if missingEndpoint {
		info.Reasons = append(info.Reasons, "no-observed-endpoint")
		info.LagNS = ""
	} else {
		info.EndpointNS = strconv.FormatInt(end, 10)
	}
	if missingEndpoint || lag > q.Definition.MaxGapMS*1000000 {
		reason := "no-observed-endpoint"
		if !missingEndpoint {
			reason = "stale-export-endpoint"
			info.Reasons = append(info.Reasons, reason)
		}
		comparison.State = "unavailable"
		comparison.CountDifference = ""
		comparison.Warnings = append(comparison.Warnings, "Live rate withheld: endpoint discovery did not establish a fresh export endpoint. Frozen windows remain historical evidence.")
		for i := range comparison.NativeWindows {
			comparison.NativeWindows[i].State = "unavailable"
			comparison.NativeWindows[i].Reasons = append(comparison.NativeWindows[i].Reasons, reason)
			comparison.NativeWindows[i].RateDifference = ""
		}
	}
	return snapshot, comparison, info, nil
}

package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/casselc/goo/metrics"
)

// EvidenceComparison compares validated stored results. It never executes a source
// query or infers aggregates from retained inspection rows.
type EvidenceSide struct {
	Receipt CaptureReceipt `json:"receipt"`
	ReadAt  string         `json:"readAt"`
	State   string         `json:"state"`
	Count   string         `json:"count"`
}
type GaugeChange struct {
	Stream     metrics.Stream       `json:"stream"`
	Baseline   *metrics.GaugeSample `json:"baseline,omitempty"`
	Selection  *metrics.GaugeSample `json:"selection,omitempty"`
	Difference string               `json:"difference,omitempty"`
	State      string               `json:"state"`
}
type EvidenceComparison struct {
	Baseline        EvidenceSide       `json:"baseline"`
	Selection       EvidenceSide       `json:"selection"`
	State           string             `json:"state"`
	Differences     []string           `json:"differences"`
	Warnings        []string           `json:"warnings"`
	CountDifference string             `json:"countDifference,omitempty"`
	Measures        *MeasureComparison `json:"measures,omitempty"`
	Gauges          []GaugeChange      `json:"gauges,omitempty"`
}

func evidenceSide(c *SavedCapture) EvidenceSide {
	s := EvidenceSide{Receipt: c.CaptureReceipt}
	if r := c.Explore; r != nil {
		s.ReadAt = r.CapturedAt
		s.State = r.State
		s.Count = r.Count
	}
	if r := c.Measurement; r != nil {
		s.ReadAt = r.ReadAt
		s.State = r.State
		s.Count = r.Count
	}
	return s
}
func CompareEvidence(a, b *SavedCapture) (*EvidenceComparison, error) {
	if a == nil || b == nil {
		return nil, fmt.Errorf("two stored captures required")
	}
	if err := validateCapture(a); err != nil {
		return nil, err
	}
	if err := validateCapture(b); err != nil {
		return nil, err
	}
	r := &EvidenceComparison{Baseline: evidenceSide(a), Selection: evidenceSide(b), State: "comparable", Differences: []string{}, Warnings: []string{"Stored results from separate reads; late arrivals, sampling and source changes can affect differences. Coverage remains unknown.", "Differences are descriptive, not causal attribution or statistical significance. Counts describe source records, not unique requests or native metric values."}}
	qa, qb := a.Question.Normalized(), b.Question.Normalized()
	if qa.Kind != qb.Kind {
		r.Differences = append(r.Differences, "Signal family differs")
		r.State = "incompatible"
		return r, nil
	}
	if qa.Kind != "explore" && qa.Kind != "measurement" {
		r.State = "unsupported"
		r.Differences = append(r.Differences, "Choose two record or measurement captures")
		return r, nil
	}
	var ea, eb string
	var fa, ta, fb, tb int64
	if qa.Explore != nil {
		x, y := *qa.Explore, *qb.Explore
		ea, eb = x.Environment, y.Environment
		fa, ta, fb, tb = x.FromMS, x.ToMS, y.FromMS, y.ToMS
		if x.Signal != y.Signal {
			r.Differences = append(r.Differences, "Signal family differs")
			r.State = "incompatible"
		}
		x.Environment = ""
		y.Environment = ""
		x.FromMS = 0
		y.FromMS = 0
		x.ToMS = 0
		y.ToMS = 0
		if !reflect.DeepEqual(x, y) {
			r.Differences = append(r.Differences, "Question definition differs")
			r.State = "incompatible"
		}
	} else {
		x, y := *qa.Measurement, *qb.Measurement
		ea, eb = x.Definition.Environment, y.Definition.Environment
		fa, ta, fb, tb = x.FromMS, x.ToMS, y.FromMS, y.ToMS
		if x.Definition.Source.Unit != y.Definition.Source.Unit {
			r.Differences = append(r.Differences, "Unit differs")
			r.State = "incompatible"
		}
		x.Definition.Environment = ""
		y.Definition.Environment = ""
		if !reflect.DeepEqual(x.Definition, y.Definition) {
			r.Differences = append(r.Differences, "Question definition differs")
			r.State = "incompatible"
		}
	}
	if ea != eb {
		r.Differences = append(r.Differences, "Environment differs")
		r.State = "incompatible"
	}
	if fa != fb || ta != tb {
		r.Differences = append(r.Differences, "Fixed window differs")
	}
	if ta-fa != tb-fb {
		r.Warnings = append(r.Warnings, "Window lengths differ; raw count differences are not normalized rates.")
	}
	if r.State != "comparable" {
		return r, nil
	}
	if r.Baseline.State != "observed" || r.Selection.State != "observed" {
		r.State = "unavailable"
		r.Warnings = append(r.Warnings, "No numeric differences: one result is empty, unknown or changed between reads.")
		return r, nil
	}
	if _, e := decimal(r.Baseline.Count); e != nil {
		return nil, e
	}
	if _, e := decimal(r.Selection.Count); e != nil {
		return nil, e
	}
	r.CountDifference = signedDifference(r.Baseline.Count, r.Selection.Count)
	if a.Explore != nil {
		r.Measures = compareMeasures(a.Explore.Duration, b.Explore.Duration, a.Explore.SpanStatus, b.Explore.SpanStatus)
	} else if a.Measurement.Gauges != nil && b.Measurement.Gauges != nil {
		type endpoints struct {
			stream metrics.Stream
			a, b   *metrics.GaugeSample
		}
		m := map[string]*endpoints{}
		for side, result := range []*metrics.GaugeResult{a.Measurement.Gauges, b.Measurement.Gauges} {
			for _, s := range result.Series {
				raw, _ := json.Marshal(s.Stream)
				key := string(raw)
				p := m[key]
				if p == nil {
					p = &endpoints{stream: s.Stream}
					m[key] = p
				}
				// Keep the endpoint even when stale/unknown; never fall back to an earlier healthy sample.
				var last *metrics.GaugeSample
				for _, sample := range s.Samples {
					v := sample
					if last == nil || nsLess(last.AtNS, v.AtNS) {
						last = &v
					}
				}
				if side == 0 {
					p.a = last
				} else {
					p.b = last
				}
			}
		}
		keys := []string{}
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			p := m[k]
			g := GaugeChange{Stream: p.stream, Baseline: p.a, Selection: p.b, State: "unavailable"}
			if p.a != nil && p.b != nil && p.a.State == "observed" && p.b.State == "observed" && p.a.Value != nil && p.b.Value != nil {
				// Exact decimal subtraction avoids floating overflow and preserves wire values.
				x, _ := new(big.Rat).SetString(strconv.FormatFloat(*p.a.Value, 'g', -1, 64))
				y, _ := new(big.Rat).SetString(strconv.FormatFloat(*p.b.Value, 'g', -1, 64))
				g.Difference = strings.TrimRight(strings.TrimRight(new(big.Rat).Sub(y, x).FloatString(324), "0"), ".")
				g.State = "comparable"
			}
			r.Gauges = append(r.Gauges, g)
		}
		r.Warnings = append(r.Warnings, "Gauge differences compare latest stored endpoints per exact OTel stream, at the displayed times; they are not window averages.")
	} else {
		r.Warnings = append(r.Warnings, "Native interval and lifecycle results remain side by side. No fleet scalar, percentile subtraction or interval realignment is inferred.")
	}
	return r, nil
}
func nsLess(a, b string) bool {
	x, _ := new(big.Int).SetString(a, 10)
	y, _ := new(big.Int).SetString(b, 10)
	return x != nil && y != nil && x.Cmp(y) < 0
}

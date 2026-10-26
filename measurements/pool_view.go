package measurements

import (
	"encoding/json"
	"fmt"
	"maps"
	"math/big"
	"sort"

	"github.com/casselc/goo/metrics"
)

const PoolViewVersion = "goo/pool-view/v1"

type PoolViewRequest struct {
	Count    Result         `json:"count"`
	Stream   metrics.Stream `json:"stream"`
	Capacity *Result        `json:"capacity,omitempty"`
	Wait     *Result        `json:"wait,omitempty"`
}
type PoolViewEvidence struct {
	Basis  string `json:"basis"`
	ReadAt string `json:"readAt"`
}
type PoolViewCountState struct {
	Name   string               `json:"name"`
	Stream metrics.Stream       `json:"stream"`
	Latest *metrics.StockSample `json:"latest,omitempty"`
}
type PoolViewCapacity struct {
	Stream   metrics.Stream       `json:"stream"`
	Latest   *metrics.StockSample `json:"latest,omitempty"`
	Evidence PoolViewEvidence     `json:"evidence"`
}
type PoolViewWait struct {
	Stream   metrics.Stream             `json:"stream"`
	Latest   *metrics.HistogramInterval `json:"latest,omitempty"`
	Evidence PoolViewEvidence           `json:"evidence"`
}
type PoolViewUtilization struct {
	State       string `json:"state"`
	Reason      string `json:"reason"`
	Numerator   string `json:"numerator,omitempty"`
	Denominator string `json:"denominator,omitempty"`
}
type PoolView struct {
	Version          string               `json:"version"`
	Stream           metrics.Stream       `json:"stream"`
	CountQuestion    Question             `json:"countQuestion"`
	CapacityQuestion *Question            `json:"capacityQuestion,omitempty"`
	WaitQuestion     *Question            `json:"waitQuestion,omitempty"`
	States           []PoolViewCountState `json:"states"`
	CountEvidence    PoolViewEvidence     `json:"countEvidence"`
	Capacity         *PoolViewCapacity    `json:"capacity,omitempty"`
	Wait             *PoolViewWait        `json:"wait,omitempty"`
	Utilization      PoolViewUtilization  `json:"utilization"`
	Warnings         []string             `json:"warnings"`
}

func poolViewIdentity(a, b metrics.Stream) bool {
	return a.ResourceSchemaURL == b.ResourceSchemaURL && maps.Equal(a.Resource, b.Resource) && a.Scope == b.Scope && a.Version == b.Version && a.SchemaURL == b.SchemaURL && maps.Equal(a.Attributes, b.Attributes)
}
func poolViewWithoutState(s metrics.Stream) metrics.Stream {
	s.Attributes = maps.Clone(s.Attributes)
	delete(s.Attributes, "db.client.connection.state")
	return s
}
func poolViewLatest(samples []metrics.StockSample) *metrics.StockSample {
	if len(samples) == 0 {
		return nil
	}
	copy := samples[len(samples)-1]
	return &copy
}

// ComposePoolView certifies only reported identity and clocks. It makes no source
// reads and creates no paired receipt or transactional observation claim.
func ComposePoolView(in PoolViewRequest) (*PoolView, error) {
	if e := ValidateResult(&in.Count); e != nil {
		return nil, e
	}
	d := in.Count.Question.Definition
	if d.Kind != "stock" || d.StockDomain != "nonnegative-integer" || d.Source.Metric != "db.client.connection.count" || d.Source.Unit != "{connection}" || d.Source.SchemaURL != PoolConnectionsSchema || in.Count.Stocks == nil {
		return nil, fmt.Errorf("pool view requires certified connection count stocks")
	}
	if in.Stream.Resource["service.instance.id"] == "" || in.Stream.Attributes["db.client.connection.pool.name"] == "" || in.Stream.Attributes["db.client.connection.state"] == "" {
		return nil, fmt.Errorf("pool view requires an observed instance, pool and state")
	}
	anchorFound := false
	for _, series := range in.Count.Stocks.Series {
		if sameExactStream(series.Stream, in.Stream) {
			anchorFound = true
			break
		}
	}
	if !anchorFound {
		return nil, fmt.Errorf("pool view anchor is not in the held count result")
	}
	for _, optional := range []*Result{in.Capacity, in.Wait} {
		if optional != nil {
			if e := ValidateResult(optional); e != nil {
				return nil, e
			}
			from, to := optional.Question.BoundsNS()
			a, b := in.Count.Question.BoundsNS()
			if from != a || to != b {
				return nil, fmt.Errorf("pool view results require identical fixed evaluation bounds")
			}
		}
	}
	encoded, e := json.Marshal(in)
	if e != nil {
		return nil, e
	}
	var req PoolViewRequest
	if e = json.Unmarshal(encoded, &req); e != nil {
		return nil, e
	}
	result := &PoolView{Version: PoolViewVersion, Stream: req.Stream, CountQuestion: req.Count.Question, CountEvidence: PoolViewEvidence{Basis: req.Count.Basis, ReadAt: req.Count.ReadAt}, States: []PoolViewCountState{}, Utilization: PoolViewUtilization{State: "unavailable", Reason: "capacity-not-supplied"}, Warnings: []string{"Each signal retains its independent source read time and evidence basis. Matching reported observation clocks and epochs is not a transactional snapshot or proof of pool health.", "Only the literal used state can form reported used / reported maximum. Idle and custom states remain independent; no states are summed or interpolated.", "Acquisition wait is an independent histogram population and interval, never the denominator or an aligned connection-count population."}}
	expected := poolViewWithoutState(req.Stream)
	for _, series := range req.Count.Stocks.Series {
		if poolViewIdentity(poolViewWithoutState(series.Stream), expected) {
			name := series.Stream.Attributes["db.client.connection.state"]
			if name == "" {
				continue
			}
			result.States = append(result.States, PoolViewCountState{Name: name, Stream: series.Stream, Latest: poolViewLatest(series.Samples)})
		}
	}
	sort.Slice(result.States, func(i, j int) bool { return result.States[i].Name < result.States[j].Name })
	check := func(r *Result, kind, metric, unit string) (metrics.Stream, error) {
		def := r.Question.Definition
		if def.Kind != kind || def.Source.Metric != metric || def.Source.Unit != unit || def.Source.SchemaURL != PoolConnectionsSchema || def.Source.ExactStream == nil {
			return metrics.Stream{}, fmt.Errorf("pool view optional signal requires the exact pinned kind and stream")
		}
		stream := *def.Source.ExactStream
		if !poolViewIdentity(stream, expected) || stream.Name != metric || stream.Unit != unit {
			return metrics.Stream{}, fmt.Errorf("pool view optional signal identity differs from held pool")
		}
		return stream, nil
	}
	if req.Capacity != nil {
		capacity := req.Capacity
		stream, e := check(capacity, "stock", "db.client.connection.max", "{connection}")
		if e != nil {
			return nil, e
		}
		if capacity.Question.Definition.StockDomain != "nonnegative-integer" || capacity.Stocks == nil {
			return nil, fmt.Errorf("pool capacity requires certified nonnegative integer stocks")
		}
		result.CapacityQuestion = &capacity.Question
		result.Capacity = &PoolViewCapacity{Stream: stream, Evidence: PoolViewEvidence{Basis: capacity.Basis, ReadAt: capacity.ReadAt}}
		for _, series := range capacity.Stocks.Series {
			if !sameExactStream(series.Stream, stream) {
				return nil, fmt.Errorf("pool capacity includes an unexpected stream")
			}
			result.Capacity.Latest = poolViewLatest(series.Samples)
		}
	}
	if req.Wait != nil {
		wait := req.Wait
		stream, e := check(wait, "histogram", "db.client.connection.wait_time", "s")
		if e != nil {
			return nil, e
		}
		if wait.Histograms == nil {
			return nil, fmt.Errorf("pool wait requires histogram evidence")
		}
		result.WaitQuestion = &wait.Question
		result.Wait = &PoolViewWait{Stream: stream, Evidence: PoolViewEvidence{Basis: wait.Basis, ReadAt: wait.ReadAt}}
		for _, series := range wait.Histograms.Series {
			if !sameExactStream(series.Stream, stream) {
				return nil, fmt.Errorf("pool wait includes an unexpected stream")
			}
			if len(series.Intervals) > 0 {
				latest := series.Intervals[len(series.Intervals)-1]
				if latest.State != "ok" {
					latest.Distribution = nil
				}
				result.Wait.Latest = &latest
			}
		}
	}
	if result.Capacity == nil {
		return result, nil
	}
	var used *metrics.StockSample
	for _, state := range result.States {
		if state.Name == "used" {
			used = state.Latest
		}
	}
	unavailable := func(reason string) (*PoolView, error) {
		result.Utilization = PoolViewUtilization{State: "unavailable", Reason: reason}
		return result, nil
	}
	if used == nil {
		return unavailable("used-not-reported")
	}
	if used.State != "observed" || used.Value == nil {
		return unavailable("used-unavailable")
	}
	capacity := result.Capacity.Latest
	if capacity == nil {
		return unavailable("capacity-no-data")
	}
	if capacity.State != "observed" || capacity.Value == nil {
		return unavailable("capacity-unavailable")
	}
	if used.AtNS != capacity.AtNS {
		return unavailable("observation-time-mismatch")
	}
	if used.StartNS == "0" || capacity.StartNS == "0" {
		return unavailable("epoch-not-reported")
	}
	if used.StartNS != capacity.StartNS {
		return unavailable("epoch-mismatch")
	}
	if *capacity.Value == 0 {
		return unavailable("zero-capacity")
	}
	if *used.Value > *capacity.Value {
		result.Utilization = PoolViewUtilization{State: "inconsistent", Reason: "used-exceeds-capacity"}
		return result, nil
	}
	ratio := new(big.Rat).SetFrac(big.NewInt(int64(*used.Value)), big.NewInt(int64(*capacity.Value)))
	result.Utilization = PoolViewUtilization{State: "observed", Reason: "matching-reported-observation-and-epoch", Numerator: ratio.Num().String(), Denominator: ratio.Denom().String()}
	return result, nil
}

package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
)

// DurationSummary describes captured span records, never unique requests or traces.
// Mean is the exact rational SumNS/Count. Values remain decimal integers on wire.
type DurationSummary struct {
	State string `json:"state"`
	Count string `json:"count"`
	SumNS string `json:"sumNs"`
	MinNS string `json:"minNs,omitempty"`
	MaxNS string `json:"maxNs,omitempty"`
	P50NS string `json:"p50Ns,omitempty"`
	P95NS string `json:"p95Ns,omitempty"`
	P99NS string `json:"p99Ns,omitempty"`
}

func durationDecimal(v string, bits int) (*big.Int, error) {
	n, ok := new(big.Int).SetString(v, 10)
	if !ok || n.Sign() < 0 || n.BitLen() > bits || n.String() != v {
		return nil, fmt.Errorf("invalid unsigned decimal summary")
	}
	return n, nil
}
func (d DurationSummary) Validate() error {
	n, e := durationDecimal(d.Count, 64)
	if e != nil {
		return e
	}
	sum, e := durationDecimal(d.SumNS, 128)
	if e != nil {
		return e
	}
	values := []string{d.MinNS, d.P50NS, d.P95NS, d.P99NS, d.MaxNS}
	if n.Sign() == 0 {
		if d.State != "no_data" || sum.Sign() != 0 {
			return fmt.Errorf("invalid empty duration summary")
		}
		for _, v := range values {
			if v != "" {
				return fmt.Errorf("empty summary has values")
			}
		}
		return nil
	}
	if d.State != "observed" {
		return fmt.Errorf("invalid duration state")
	}
	var previous *big.Int
	for _, v := range values {
		current, e := durationDecimal(v, 64)
		if e != nil {
			return e
		}
		if previous != nil && current.Cmp(previous) < 0 {
			return fmt.Errorf("unordered duration quantiles")
		}
		previous = current
	}
	lo, _ := durationDecimal(d.MinNS, 64)
	hi, _ := durationDecimal(d.MaxNS, 64)
	if sum.Cmp(new(big.Int).Mul(lo, n)) < 0 || sum.Cmp(new(big.Int).Mul(hi, n)) > 0 {
		return fmt.Errorf("duration sum outside bounds")
	}
	return nil
}
func DecodeDurationSummary(raw json.RawMessage) (*DurationSummary, error) {
	var row struct {
		Count string   `json:"count"`
		Sum   string   `json:"sum_ns"`
		Min   string   `json:"min_ns"`
		Max   string   `json:"max_ns"`
		Q     []string `json:"quantiles_ns"`
	}
	if json.Unmarshal(raw, &row) != nil || len(row.Q) != 3 {
		return nil, fmt.Errorf("invalid duration summary source row")
	}
	d := &DurationSummary{State: "observed", Count: row.Count, SumNS: row.Sum, MinNS: row.Min, MaxNS: row.Max, P50NS: row.Q[0], P95NS: row.Q[1], P99NS: row.Q[2]}
	if row.Count == "0" { // ClickHouse's empty UInt64 aggregate defaults are not observations.
		if row.Min != "0" || row.Max != "0" || row.Q[0] != "0" || row.Q[1] != "0" || row.Q[2] != "0" {
			return nil, fmt.Errorf("invalid empty duration source row")
		}
		d.State = "no_data"
		d.MinNS = ""
		d.MaxNS = ""
		d.P50NS = ""
		d.P95NS = ""
		d.P99NS = ""
	}
	return d, d.Validate()
}

// ReferenceDurationSummary uses exact upper order statistics at floor(p*N),
// clamped to N-1, matching quantilesExactHigh. It does not interpolate.
func ReferenceDurationSummary(values []uint64) DurationSummary {
	a := append([]uint64(nil), values...)
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	d := DurationSummary{State: "no_data", Count: fmt.Sprint(len(a)), SumNS: "0"}
	if len(a) == 0 {
		return d
	}
	sum := new(big.Int)
	for _, v := range a {
		sum.Add(sum, new(big.Int).SetUint64(v))
	}
	at := func(p int) string {
		i := (len(a)/100)*p + (len(a)%100)*p/100
		if i >= len(a) {
			i = len(a) - 1
		}
		return fmt.Sprint(a[i])
	}
	d.State = "observed"
	d.SumNS = sum.String()
	d.MinNS = fmt.Sprint(a[0])
	d.MaxNS = fmt.Sprint(a[len(a)-1])
	d.P50NS = at(50)
	d.P95NS = at(95)
	d.P99NS = at(99)
	return d
}

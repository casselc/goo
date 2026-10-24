package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// NumericHistogramSpec requests a full-population integer distribution from the
// same aggregate read as numeric coverage. Binning is fixed, never sample-derived.
type NumericHistogramSpec struct {
	Binning string `json:"binning" yaml:"binning"`
}
type NumericHistogramBin struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Count string `json:"count"`
}
type NumericHistogram struct {
	Binning string                `json:"binning"`
	Bins    []NumericHistogramBin `json:"bins"`
}

func (s NumericHistogramSpec) Validate(b *NumericBinding, signal string) error {
	if s.Binning != "powers-of-two" || b == nil {
		return fmt.Errorf("powers-of-two numeric histogram requires numeric binding")
	}
	return b.Validate(signal)
}

// NumericHistogramEdges includes the exclusive max+1 sentinel. That sentinel is
// histogram geometry only, never a valid numeric predicate threshold or value.
func NumericHistogramEdges(typ string) ([]string, error) {
	edges := []string{}
	switch typ {
	case "uint64":
		edges = append(edges, "0")
		for n := 0; n <= 64; n++ {
			edges = append(edges, new(big.Int).Lsh(big.NewInt(1), uint(n)).String())
		}
	case "int64":
		for n := 63; n >= 0; n-- {
			edges = append(edges, new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(n))).String())
		}
		edges = append(edges, "0")
		for n := 0; n <= 63; n++ {
			edges = append(edges, new(big.Int).Lsh(big.NewInt(1), uint(n)).String())
		}
	default:
		return nil, fmt.Errorf("histogram requires int64 or uint64")
	}
	return edges, nil
}

func EvaluateNumericHistogram(b NumericBinding, values []*string) (NumericHistogram, error) {
	if e := b.Validate("spans"); e != nil {
		return NumericHistogram{}, e
	}
	edges, _ := NumericHistogramEdges(b.Type)
	h := NumericHistogram{Binning: "powers-of-two", Bins: make([]NumericHistogramBin, len(edges)-1)}
	thresholds := make([]*big.Int, len(edges))
	for i, v := range edges {
		thresholds[i], _ = new(big.Int).SetString(v, 10)
		if i < len(h.Bins) {
			h.Bins[i] = NumericHistogramBin{From: v, To: edges[i+1], Count: "0"}
		}
	}
	counts := make([]uint64, len(h.Bins))
	for _, raw := range values {
		if raw == nil {
			continue
		}
		v, ok := parseCanonicalInteger(*raw, b.Type)
		if !ok {
			continue
		}
		i := sort.Search(len(thresholds), func(i int) bool { return thresholds[i].Cmp(v) > 0 }) - 1
		counts[i]++
	}
	for i, n := range counts {
		h.Bins[i].Count = fmt.Sprint(n)
	}
	return h, nil
}

func (h NumericHistogram) Validate(b NumericBinding, summary NumericSummary) error {
	if e := b.Validate("spans"); e != nil {
		return e
	}
	if e := summary.Validate(b); e != nil {
		return e
	}
	edges, _ := NumericHistogramEdges(b.Type)
	if h.Binning != "powers-of-two" || len(h.Bins) != len(edges)-1 {
		return fmt.Errorf("numeric histogram requires complete fixed bin schema")
	}
	sum := new(big.Int)
	for i, bin := range h.Bins {
		if bin.From != edges[i] || bin.To != edges[i+1] {
			return fmt.Errorf("numeric histogram boundaries differ from fixed schema")
		}
		n, e := decimal(bin.Count)
		if e != nil {
			return fmt.Errorf("invalid numeric histogram count")
		}
		sum.Add(sum, n)
	}
	if sum.String() != summary.Valid {
		return fmt.Errorf("numeric histogram counts disagree with numeric valid coverage")
	}
	return nil
}

func validateOptionalNumericHistogram(q Query, h *NumericHistogram, summary *NumericSummary) error {
	if q.NumericHistogram == nil {
		if h != nil {
			return fmt.Errorf("unexpected numeric histogram")
		}
		return nil
	}
	if e := q.NumericHistogram.Validate(q.Numeric, q.Signal); e != nil {
		return e
	}
	if h == nil || summary == nil {
		return fmt.Errorf("missing numeric histogram or numeric coverage")
	}
	if h.Binning != q.NumericHistogram.Binning {
		return fmt.Errorf("numeric histogram binning mismatch")
	}
	return h.Validate(*q.Numeric, *summary)
}

// Only trusted, generated edges become SQL literals; exporter keys remain typed
// parameters. Int256 comparisons cover max+1 without overflow or float conversion.
func (c *compiler) numericHistogramSQL(b NumericBinding) string {
	_, _, value := c.numericValueSQL(b)
	edges, _ := NumericHistogramEdges(b.Type)
	bins := make([]string, 0, len(edges)-1)
	for i := 0; i < len(edges)-1; i++ {
		condition := "toInt256(" + value + ") >= toInt256('" + edges[i] + "') AND toInt256(" + value + ") < toInt256('" + edges[i+1] + "')"
		bins = append(bins, "map('from', '"+edges[i]+"', 'to', '"+edges[i+1]+"', 'count', toString(countIf(ifNull(("+condition+"), 0))))")
	}
	return "[" + strings.Join(bins, ", ") + "] AS numeric_histogram"
}

func decodeNumericHistogram(raw json.RawMessage, q Query, summary *NumericSummary) (*NumericHistogram, error) {
	var row struct {
		Bins []NumericHistogramBin `json:"numeric_histogram"`
	}
	if e := json.Unmarshal(raw, &row); e != nil {
		return nil, fmt.Errorf("invalid numeric histogram source row")
	}
	if q.NumericHistogram == nil {
		if row.Bins != nil {
			return nil, fmt.Errorf("unexpected numeric histogram source row")
		}
		return nil, nil
	}
	h := &NumericHistogram{Binning: q.NumericHistogram.Binning, Bins: row.Bins}
	if e := validateOptionalNumericHistogram(q, h, summary); e != nil {
		return nil, e
	}
	return h, nil
}

type NumericHistogramComparisonQueries struct {
	Baseline  Query `json:"baseline"`
	Selection Query `json:"selection"`
}
type NumericHistogramSelection struct {
	NumericComparison *NumericHistogramComparisonQueries `json:"numericComparison,omitempty"`
	Parent            Query                              `json:"parent"`
	Selected          Query                              `json:"selected"`
	Rest              Query                              `json:"rest"`
	Comparison        NumericHistogramComparisonQueries  `json:"comparison"`
	FromBin           int                                `json:"fromBin"`
	ToBin             int                                `json:"toBin"`
	Binning           string                             `json:"binning"`
	From              string                             `json:"from"`
	To                string                             `json:"to"`
}

// PlanNumericHistogramSelection is source-free. Indexes are half-open, like bin
// geometry. Rest is explicit coverage states plus outside valid values, not NOT
// over a nullable comparison. All output questions own detached configuration.
func PlanNumericHistogramSelection(parent Query, fromBin, toBin int) (*NumericHistogramSelection, error) {
	if e := parent.Validate(); e != nil {
		return nil, e
	}
	if parent.NumericHistogram == nil {
		return nil, fmt.Errorf("numeric histogram required for bin selection")
	}
	edges, _ := NumericHistogramEdges(parent.Numeric.Type)
	if fromBin < 0 || toBin <= fromBin || toBin >= len(edges) {
		return nil, fmt.Errorf("require nonempty half-open histogram bin indexes")
	}
	leaf := func(op, value string) Predicate {
		b := *parent.Numeric
		return Predicate{Numeric: &b, Op: op, Value: value}
	}
	selectedParts := []Predicate{}
	if fromBin > 0 {
		selectedParts = append(selectedParts, leaf("gte", edges[fromBin]))
	}
	if toBin < len(edges)-1 {
		selectedParts = append(selectedParts, leaf("lt", edges[toBin]))
	}
	selected := leaf("valid", "")
	if len(selectedParts) == 1 {
		selected = selectedParts[0]
	} else if len(selectedParts) == 2 {
		selected = Predicate{All: selectedParts}
	}
	restParts := []Predicate{leaf("missing", ""), leaf("empty", ""), leaf("invalid", "")}
	if fromBin > 0 {
		restParts = append(restParts, leaf("lt", edges[fromBin]))
	}
	if toBin < len(edges)-1 {
		restParts = append(restParts, leaf("gte", edges[toBin]))
	}
	compose := func(p Predicate) Query {
		q := detached(parent)
		q.Offset = 0
		if q.Where == nil {
			q.Where = &p
		} else {
			q.Where = &Predicate{All: []Predicate{*q.Where, p}}
		}
		return detached(q)
	}
	out := &NumericHistogramSelection{Parent: detached(parent), Selected: compose(selected), Rest: compose(Predicate{Any: restParts}), FromBin: fromBin, ToBin: toBin, Binning: parent.NumericHistogram.Binning, From: edges[fromBin], To: edges[toBin]}
	project := func(q Query) Query {
		v := detached(q)
		v.Numeric = nil
		v.NumericHistogram = nil
		v.Summary = ""
		v.SpanStatus = false
		// Existing cohort comparison requires one shared distribution dimension.
		// Keep the first declared dimension, or use the service-name distribution.
		if len(v.GroupBy) == 0 {
			v.GroupBy = []Field{{Kind: "column", Key: "service"}}
		} else {
			v.GroupBy = v.GroupBy[:1]
		}
		return v
	}
	out.Comparison = NumericHistogramComparisonQueries{Baseline: project(out.Rest), Selection: project(out.Selected)}
	numericProject := func(q Query) Query {
		v := detached(q)
		v.NumericHistogram = nil
		v.Summary = ""
		v.SpanStatus = false
		if len(v.GroupBy) == 0 {
			v.GroupBy = []Field{{Kind: "column", Key: "service"}}
		} else {
			v.GroupBy = v.GroupBy[:1]
		}
		return v
	}
	out.NumericComparison = &NumericHistogramComparisonQueries{Baseline: numericProject(out.Rest), Selection: numericProject(out.Selected)}
	for _, q := range []Query{out.Selected, out.Rest, out.Comparison.Baseline, out.Comparison.Selection, out.NumericComparison.Baseline, out.NumericComparison.Selection} {
		if e := q.Validate(); e != nil {
			return nil, fmt.Errorf("histogram selection exceeds supported question: %w", e)
		}
	}
	return out, nil
}

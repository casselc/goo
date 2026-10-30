package exploration

import "math/big"

// ViewAggregation describes the exact operation behind a supplied measure. It
// is not a client-side reducer; quantiles cannot be combined by averaging.
type ViewAggregation struct {
	Family      string `json:"family"`
	Kind        string `json:"kind"`
	Method      string `json:"method"`
	Quantile    string `json:"quantile,omitempty"`
	Numerator   string `json:"numerator,omitempty"`
	Denominator string `json:"denominator,omitempty"`
}
type ViewCell struct {
	State   string       `json:"state"`
	Support *ViewSupport `json:"support,omitempty"`
}
type ViewSupport struct {
	Total   string `json:"total"`
	Valid   string `json:"valid"`
	Missing string `json:"missing"`
	Empty   string `json:"empty"`
	Invalid string `json:"invalid"`
}

func viewMeasureFields(q Query) []ViewField {
	fields := []ViewField{}
	add := func(family, key, label, kind, method, unit, quantile string) {
		fields = append(fields, ViewField{Key: family + "." + key, Label: label, Type: "rational", Unit: unit, Nullable: true, Aggregation: &ViewAggregation{Family: family, Kind: kind, Method: method, Quantile: quantile}})
	}
	if q.Numeric != nil {
		add("numeric", "mean", "Numeric mean", "mean", "exact-rational", q.Numeric.Unit, "")
		for _, kind := range []string{"sum", "min", "max"} {
			add("numeric", kind, "Numeric "+kind, kind, "exact", q.Numeric.Unit, "")
		}
	}
	if q.Summary == "span-duration" {
		add("duration", "mean", "Mean duration", "mean", "exact-rational", "ns", "")
		add("duration", "min", "Min duration", "min", "exact", "ns", "")
		add("duration", "max", "Max duration", "max", "exact", "ns", "")
		for _, p := range []struct{ key, value string }{{"p50", "0.5"}, {"p95", "0.95"}, {"p99", "0.99"}} {
			add("duration", p.key, p.key+" duration", "quantile", "quantilesExactHigh", "ns", p.value)
		}
	}
	if q.SpanStatus {
		for _, status := range []struct{ key, label string }{{"error", "Recorded errors"}, {"ok", "Explicit OK"}, {"unset", "Unset status"}, {"unknown", "Unknown status"}} {
			fields = append(fields, ViewField{Key: "status." + status.key, Label: status.label, Type: "number", Unit: "records", Nullable: true, Aggregation: &ViewAggregation{Family: "span-status", Kind: "count", Method: "exact"}})
		}
		fields = append(fields, ViewField{Key: "status.error_fraction", Label: "Recorded error fraction", Type: "rational", Unit: "1", Nullable: true, Aggregation: &ViewAggregation{Family: "ratio", Kind: "fraction", Method: "exact-rational", Numerator: "status.error", Denominator: "count"}})
	}
	return fields
}
func viewRational(coefficient string, denominator *big.Int) *string {
	n, _ := new(big.Int).SetString(coefficient, 10)
	return viewText(new(big.Rat).SetFrac(n, denominator).RatString())
}

// Summaries are validated against their binding and row count before converting
// coefficients into exact human units. Unavailable sums stay null rather than
// inheriting the arithmetic identity zero from an empty source aggregate.
func viewMeasureCells(q Query, count string, numeric *NumericSummary, duration *DurationSummary, status *SpanStatusSummary, unknown bool, values ViewValues) (map[string]ViewCell, bool) {
	fields := viewMeasureFields(q)
	if len(fields) == 0 {
		return nil, numeric == nil && duration == nil && status == nil
	}
	cells := map[string]ViewCell{}
	if unknown {
		if count != "" || numeric != nil || duration != nil || status != nil {
			return nil, false
		}
		for _, f := range fields {
			values[f.Key] = nil
			cells[f.Key] = ViewCell{State: "unknown"}
		}
		return cells, true
	}
	if validateOptionalNumeric(q.Numeric, numeric, count) != nil {
		return nil, false
	}
	if q.Summary == "span-duration" {
		if duration == nil || duration.Count != count || duration.Validate() != nil {
			return nil, false
		}
	} else if duration != nil {
		return nil, false
	}
	if q.SpanStatus {
		if status == nil || status.Count != count || status.Validate() != nil {
			return nil, false
		}
	} else if status != nil {
		return nil, false
	}
	if numeric != nil {
		support := &ViewSupport{Total: numeric.Count, Valid: numeric.Valid, Missing: numeric.Missing, Empty: numeric.Empty, Invalid: numeric.Invalid}
		state := numeric.State
		if state == "observed" {
			state = "derived"
		}
		for _, kind := range []string{"mean", "sum", "min", "max"} {
			key := "numeric." + kind
			values[key] = nil
			cells[key] = ViewCell{State: state, Support: support}
		}
		if numeric.Valid != "0" {
			scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(q.Numeric.Scale)), nil)
			valid, _ := new(big.Int).SetString(numeric.Valid, 10)
			values["numeric.mean"] = viewRational(numeric.Sum, new(big.Int).Mul(valid, scale))
			values["numeric.sum"] = viewRational(numeric.Sum, scale)
			values["numeric.min"] = viewRational(numeric.Min, scale)
			values["numeric.max"] = viewRational(numeric.Max, scale)
		}
	}
	if duration != nil {
		support := &ViewSupport{Total: duration.Count, Valid: duration.Count, Missing: "0", Empty: "0", Invalid: "0"}
		state := duration.State
		if state == "observed" {
			state = "derived"
		}
		measures := map[string]string{"mean": duration.SumNS, "min": duration.MinNS, "max": duration.MaxNS, "p50": duration.P50NS, "p95": duration.P95NS, "p99": duration.P99NS}
		for kind, value := range measures {
			key := "duration." + kind
			values[key] = nil
			cells[key] = ViewCell{State: state, Support: support}
			if duration.Count != "0" {
				denominator := big.NewInt(1)
				if kind == "mean" {
					denominator, _ = new(big.Int).SetString(duration.Count, 10)
				}
				values[key] = viewRational(value, denominator)
			}
		}
	}
	if status != nil {
		for key, value := range map[string]string{"error": status.Error, "ok": status.OK, "unset": status.Unset, "unknown": status.Unknown} {
			values["status."+key] = viewText(value)
			cells["status."+key] = ViewCell{State: "derived"}
		}
		values["status.error_fraction"] = nil
		cells["status.error_fraction"] = ViewCell{State: "no_data"}
		if count != "0" {
			denominator, _ := new(big.Int).SetString(count, 10)
			values["status.error_fraction"] = viewRational(status.Error, denominator)
			cells["status.error_fraction"] = ViewCell{State: "derived"}
		}
	}
	return cells, true
}

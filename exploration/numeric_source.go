package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
)

func (b NumericBinding) Validate(signal string) error {
	if _, e := b.Field.Type(signal); e != nil {
		return e
	}
	if b.Field.Kind != "attribute" && b.Field.Kind != "resource" {
		return fmt.Errorf("numeric binding requires attribute or resource field")
	}
	if b.Type != "int64" && b.Type != "uint64" {
		return fmt.Errorf("numeric type must be int64 or uint64")
	}
	if b.Parse != "canonical-decimal" {
		return fmt.Errorf("numeric parse must be canonical-decimal; no implicit conversions")
	}
	if !validNumericUnit(b.Unit) {
		return fmt.Errorf("numeric binding requires a bounded unit label without whitespace or control characters")
	}
	return nil
}

// Roundtrip equality detects ClickHouse's permissive parsing and overflowing
// integer conversions. Only canonical, in-range inputs become non-null values.
// Count <= UInt64 and input <= UInt64 imply |sum| < 2^128; Int256 cannot overflow.
func (c *compiler) numericValueSQL(b NumericBinding) (raw, present, value string) {
	raw, present = c.field(b.Field)
	cast := "toInt64OrNull"
	if b.Type == "uint64" {
		cast = "toUInt64OrNull"
	}
	parsed := cast + "(" + raw + ")"
	value = "if((" + present + ") AND length(" + raw + ") <= 20 AND toString(" + parsed + ") = " + raw + ", " + parsed + ", NULL)"
	return raw, present, value
}

func (c *compiler) numericSQL(b NumericBinding) string {
	raw, present, value := c.numericValueSQL(b)
	valid := "count(" + value + ")"
	missing := "countIf(NOT (" + present + "))"
	empty := "countIf((" + present + ") AND " + raw + " = '')"
	unit := c.param(b.Unit, "String")
	return "map('count', toString(count()), 'valid', toString(" + valid + "), 'missing', toString(" + missing + "), 'empty', toString(" + empty + "), 'invalid', toString(count() - " + valid + " - " + missing + " - " + empty + "), 'sum', toString(sum(ifNull(toInt256(" + value + "), toInt256(0)))), 'min', if(" + valid + " > 0, toString(min(" + value + ")), ''), 'max', if(" + valid + " > 0, toString(max(" + value + ")), ''), 'state', if(count() = 0, 'no_data', if(" + valid + " = 0, 'no_valid_values', 'observed')), 'unit', " + unit + ") AS numeric"
}

func decodeNumeric(raw json.RawMessage, b NumericBinding, count string) (*NumericSummary, error) {
	var v struct {
		Numeric *NumericSummary `json:"numeric"`
	}
	if json.Unmarshal(raw, &v) != nil {
		return nil, fmt.Errorf("invalid numeric source row")
	}
	if e := validateOptionalNumeric(&b, v.Numeric, count); e != nil {
		return nil, e
	}
	return v.Numeric, nil
}
func validateOptionalNumeric(b *NumericBinding, s *NumericSummary, count string) error {
	if b == nil {
		if s != nil {
			return fmt.Errorf("unexpected numeric summary")
		}
		return nil
	}
	if s == nil || s.Count != count || s.Unit != b.Unit {
		return fmt.Errorf("missing or mismatched numeric summary")
	}
	return s.Validate(*b)
}
func (s NumericSummary) Validate(b NumericBinding) error {
	if e := b.Validate("spans"); e != nil {
		return e
	}
	if s.Unit != b.Unit {
		return fmt.Errorf("numeric unit mismatch")
	}
	counts := []string{s.Count, s.Valid, s.Missing, s.Empty, s.Invalid}
	nums := make([]*big.Int, len(counts))
	for i, v := range counts {
		n, e := durationDecimal(v, 64)
		if e != nil {
			return fmt.Errorf("invalid numeric coverage count")
		}
		nums[i] = n
	}
	n := new(big.Int)
	for _, v := range nums[1:] {
		n.Add(n, v)
	}
	if n.Cmp(nums[0]) != 0 {
		return fmt.Errorf("numeric coverage does not partition record count")
	}
	sum, ok := new(big.Int).SetString(s.Sum, 10)
	if !ok || sum.String() != s.Sum || sum.BitLen() > 128 {
		return fmt.Errorf("invalid numeric sum")
	}
	if nums[1].Sign() == 0 {
		state := "no_valid_values"
		if nums[0].Sign() == 0 {
			state = "no_data"
		}
		if s.State != state || s.Sum != "0" || s.Min != "" || s.Max != "" {
			return fmt.Errorf("numeric unavailable values must retain coverage and unavailable extrema")
		}
		return nil
	}
	if s.State != "observed" {
		return fmt.Errorf("invalid numeric state")
	}
	lo, ok := new(big.Int).SetString(s.Min, 10)
	if !ok || lo.String() != s.Min {
		return fmt.Errorf("invalid numeric min")
	}
	hi, ok := new(big.Int).SetString(s.Max, 10)
	if !ok || hi.String() != s.Max {
		return fmt.Errorf("invalid numeric max")
	}
	for _, v := range []*big.Int{lo, hi} {
		if b.Type == "int64" && !v.IsInt64() || b.Type == "uint64" && !v.IsUint64() {
			return fmt.Errorf("numeric extrema exceed declared type")
		}
	}
	if lo.Cmp(hi) > 0 || sum.Cmp(new(big.Int).Add(new(big.Int).Mul(lo, new(big.Int).Sub(nums[1], big.NewInt(1))), hi)) < 0 || sum.Cmp(new(big.Int).Add(new(big.Int).Mul(hi, new(big.Int).Sub(nums[1], big.NewInt(1))), lo)) > 0 {
		return fmt.Errorf("numeric sum outside extrema/count bounds")
	}
	return nil
}

// ValidateNumericResult is also used on immutable capture load; never trust a
// captured numeric value simply because its outer result has a basis string.
func ValidateNumericResult(r *Result) error {
	if r == nil {
		return fmt.Errorf("missing result")
	}
	b := r.Question.Numeric
	if e := validateOptionalNumeric(b, r.Numeric, r.Count); e != nil {
		return e
	}
	for _, raw := range r.Groups {
		var v struct {
			Count   string          `json:"count"`
			Numeric *NumericSummary `json:"numeric"`
		}
		if json.Unmarshal(raw, &v) != nil {
			return fmt.Errorf("invalid numeric group")
		}
		if e := validateOptionalNumeric(b, v.Numeric, v.Count); e != nil {
			return e
		}
	}
	for _, p := range r.TimeSeries {
		if e := validateOptionalNumeric(b, p.Numeric, p.Count); e != nil {
			return e
		}
	}
	for _, g := range r.GroupedTimeSeries {
		for _, p := range g.Points {
			if e := validateOptionalNumeric(b, p.Numeric, p.Count); e != nil {
				return e
			}
		}
	}
	return nil
}

func mergeNumeric(b NumericBinding, parts []*NumericSummary) *NumericSummary {
	out, _ := EvaluateNumericValues(b, nil)
	add := func(a, b string) string {
		x, _ := new(big.Int).SetString(a, 10)
		y, _ := new(big.Int).SetString(b, 10)
		return x.Add(x, y).String()
	}
	for _, s := range parts {
		out.Count = add(out.Count, s.Count)
		out.Valid = add(out.Valid, s.Valid)
		out.Missing = add(out.Missing, s.Missing)
		out.Empty = add(out.Empty, s.Empty)
		out.Invalid = add(out.Invalid, s.Invalid)
		out.Sum = add(out.Sum, s.Sum)
		if s.Valid != "0" {
			lo, _ := new(big.Int).SetString(s.Min, 10)
			hi, _ := new(big.Int).SetString(s.Max, 10)
			oldLo, _ := new(big.Int).SetString(out.Min, 10)
			oldHi, _ := new(big.Int).SetString(out.Max, 10)
			if oldLo == nil || lo.Cmp(oldLo) < 0 {
				out.Min = s.Min
			}
			if oldHi == nil || hi.Cmp(oldHi) > 0 {
				out.Max = s.Max
			}
		}
	}
	if out.Count != "0" {
		out.State = "no_valid_values"
	}
	if out.Valid != "0" {
		out.State = "observed"
	}
	return &out
}
func numericReadsAgree(r *Result) bool {
	b := r.Question.Numeric
	if b == nil {
		return true
	}
	if ValidateNumericResult(r) != nil || ValidateTimeSeries(r.Question, r.TimeSeries) != nil || ValidateGroupedTimeSeries(r.Question, r.GroupedTimeSeries) != nil {
		return false
	}
	same := func(want *NumericSummary, parts []*NumericSummary) bool {
		return want != nil && *want == *mergeNumeric(*b, parts)
	}
	groups := map[string]*NumericSummary{}
	parts := []*NumericSummary{}
	for _, raw := range r.Groups {
		v, _ := groupValues(r.Question, raw)
		var row struct {
			Numeric *NumericSummary `json:"numeric"`
		}
		json.Unmarshal(raw, &row)
		groups[groupKey(v)] = row.Numeric
		parts = append(parts, row.Numeric)
	}
	if len(r.Question.GroupBy) > 0 && !same(r.Numeric, parts) {
		return false
	}
	if r.Question.BucketMS != 0 {
		parts = nil
		for _, p := range r.TimeSeries {
			parts = append(parts, p.Numeric)
		}
		if !same(r.Numeric, parts) {
			return false
		}
	}
	if r.Question.GroupTimeSeries {
		if len(groups) != len(r.GroupedTimeSeries) {
			return false
		}
		for _, g := range r.GroupedTimeSeries {
			parts = nil
			for _, p := range g.Points {
				parts = append(parts, p.Numeric)
			}
			if !same(groups[groupKey(g.Values)], parts) {
				return false
			}
		}
		for i, p := range r.TimeSeries {
			parts = nil
			for _, g := range r.GroupedTimeSeries {
				parts = append(parts, g.Points[i].Numeric)
			}
			if !same(p.Numeric, parts) {
				return false
			}
		}
	}
	return true
}

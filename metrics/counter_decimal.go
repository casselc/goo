package metrics

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

const DecimalCounterDomain = "nonnegative-decimal"

var counterDecimal = regexp.MustCompile(`^(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$`)

// ParseCounterDecimal accepts a bounded canonical base-ten value, never binary
// rounding or exponent expansion. Source adapters canonicalize exported decimals.
func ParseCounterDecimal(s string) (*big.Rat, error) {
	if len(s) > 512 || !counterDecimal.MatchString(s) {
		return nil, fmt.Errorf("require canonical bounded nonnegative decimal counter value")
	}
	value, ok := new(big.Rat).SetString(s)
	if !ok {
		return nil, fmt.Errorf("invalid decimal counter")
	}
	return value, nil
}

// CounterDecimal formats finite base-ten differences without rounding.
func CounterDecimal(r *big.Rat) string {
	d := new(big.Int).Set(r.Denom())
	two, five := 0, 0
	for new(big.Int).Mod(d, big.NewInt(2)).Sign() == 0 {
		d.Quo(d, big.NewInt(2))
		two++
	}
	for new(big.Int).Mod(d, big.NewInt(5)).Sign() == 0 {
		d.Quo(d, big.NewInt(5))
		five++
	}
	if d.Cmp(big.NewInt(1)) != 0 {
		return r.RatString()
	}
	out := r.FloatString(max(two, five))
	if strings.Contains(out, ".") {
		out = strings.TrimRight(strings.TrimRight(out, "0"), ".")
	}
	return out
}

var exportedCounterDecimal = regexp.MustCompile(`^[+-]?([0-9]+(\.[0-9]*)?|\.[0-9]+)([eE][+-]?[0-9]+)?$`)

// ExportedNonnegativeDecimal certifies a bounded finite exported decimal string,
// including scientific notation, without substituting its rounded Float64 value.
func ExportedNonnegativeDecimal(raw string) (*big.Rat, bool) {
	if len(raw) > 128 || !exportedCounterDecimal.MatchString(raw) {
		return nil, false
	}
	if at := strings.IndexAny(raw, "eE"); at >= 0 {
		e, ok := new(big.Int).SetString(raw[at+1:], 10)
		if !ok || !e.IsInt64() || e.Int64() < -400 || e.Int64() > 400 {
			return nil, false
		}
	}
	r, ok := new(big.Rat).SetString(raw)
	if !ok || r.Sign() < 0 {
		return nil, false
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil || math.IsInf(f, 0) || math.IsNaN(f) || f == 0 && r.Sign() != 0 {
		return nil, false
	}
	if len(CounterDecimal(r)) > 512 {
		return nil, false
	}
	return r, true
}

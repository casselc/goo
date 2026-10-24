package exploration

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// Decimal source strings have exactly Scale fractional digits. Arithmetic and
// summaries use signed int64 coefficients; scale never changes record counts.
var fixedDecimalGrammar = func() [10]*regexp.Regexp {
	var patterns [10]*regexp.Regexp
	for scale := 1; scale <= 9; scale++ {
		patterns[scale] = regexp.MustCompile(fmt.Sprintf(`^-?(0|[1-9][0-9]*)[.][0-9]{%d}$`, scale))
	}
	return patterns
}()

func parseNumericCoefficient(raw string, b NumericBinding) (*big.Int, bool) {
	if b.Type != "decimal64" {
		if b.Scale != 0 {
			return nil, false
		}
		return parseCanonicalInteger(raw, b.Type)
	}
	if b.Scale < 1 || b.Scale > 9 || len(raw) > 21 || !fixedDecimalGrammar[b.Scale].MatchString(raw) {
		return nil, false
	}
	coefficient, ok := new(big.Int).SetString(strings.Replace(raw, ".", "", 1), 10)
	if !ok || !coefficient.IsInt64() || formatFixedCoefficient(coefficient, b.Scale) != raw {
		return nil, false
	}
	return coefficient, true
}
func formatFixedCoefficient(coefficient *big.Int, scale int) string {
	negative := coefficient.Sign() < 0
	digits := new(big.Int).Abs(coefficient).String()
	if len(digits) <= scale {
		digits = strings.Repeat("0", scale+1-len(digits)) + digits
	}
	value := digits[:len(digits)-scale] + "." + digits[len(digits)-scale:]
	if negative {
		value = "-" + value
	}
	return value
}
func decimalCoefficientSQL(raw, present string, scale int) (string, string, string) {
	bounded := "if(length(" + raw + ") <= 21, " + raw + ", '')"
	// Normalize zeros in coefficient text only after preserving the separate strict
	// source grammar. A negative zero becomes -0 and fails integer roundtrip.
	coefficient := "replaceRegexpOne(replaceOne(" + bounded + ", '.', ''), '^(-?)0*([0-9])', '\\\\1\\\\2')"
	parsed := "toInt64OrNull(" + coefficient + ")"
	// Use explicit end-of-text rather than $, whose final-newline semantics can
	// differ between regular-expression engines. The bounded regex sees no long input.
	grammar := fmt.Sprintf("'^-?(0|[1-9][0-9]*)[.][0-9]{%d}\\\\z'", scale)
	value := "if((" + present + ") AND length(" + raw + ") <= 21 AND match(" + bounded + ", " + grammar + ") AND toString(" + parsed + ") = " + coefficient + ", " + parsed + ", NULL)"
	return raw, present, value
}

package exploration

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const LogPatternMaxBodyBytes = 256
const LogPatternVersion = "lp1"

// LogPattern describes one body shape, never severity, cause or event identity.
// Unclassified states deliberately carry no truncated body or inferred shape.
type LogPattern struct {
	Identity string `json:"identity"`
	Label    string `json:"label"`
	State    string `json:"state"`
}

var logPatternDigits = regexp.MustCompile(`[0-9]+`)

// ReferenceLogPattern replaces maximal ASCII digit runs anywhere. Signs,
// decimal points, UUID punctuation, whitespace and non-ASCII digits stay literal.
func ReferenceLogPattern(body string) LogPattern {
	if len(body) > LogPatternMaxBodyBytes {
		return LogPattern{"lp1:o", "Body exceeds 256-byte bound · unclassified", "oversize"}
	}
	if !utf8.ValidString(body) {
		return LogPattern{"lp1:u", "Invalid UTF-8 body · unclassified", "invalid-utf8"}
	}
	if body == "" {
		return LogPattern{"lp1:e", "Empty body · unclassified", "empty"}
	}
	parts := logPatternDigits.Split(body, -1)
	encoded := make([]string, len(parts))
	labels := make([]string, len(parts))
	for i, p := range parts {
		encoded[i] = hex.EncodeToString([]byte(p))
		labels[i] = escapeLogPatternLiteral(p)
	}
	return LogPattern{"lp1:p:" + strings.Join(encoded, "n"), strings.Join(labels, "⟪digits⟫"), "pattern"}
}
func escapeLogPatternLiteral(s string) string {
	var out strings.Builder
	for _, r := range s {
		switch r {
		case '\\':
			out.WriteString(`\\`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		case '⟪':
			out.WriteString(`\⟪`)
		default:
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == '\u2028' || r == '\u2029' {
				if r <= 0xffff {
					fmt.Fprintf(&out, `\u%04x`, r)
				} else {
					fmt.Fprintf(&out, `\U%08x`, r)
				}
			} else {
				out.WriteRune(r)
			}
		}
	}
	return out.String()
}

// DecodeLogPattern validates the complete canonical identity rather than an
// encoded substring. A representative digit run reconstructs exactly its shape.
func DecodeLogPattern(identity string) (LogPattern, error) {
	switch identity {
	case "lp1:e":
		return ReferenceLogPattern(""), nil
	case "lp1:o":
		return ReferenceLogPattern(strings.Repeat("x", 257)), nil
	case "lp1:u":
		return ReferenceLogPattern(string([]byte{0xff})), nil
	}
	if len(identity) > 518 || !strings.HasPrefix(identity, "lp1:p:") {
		return LogPattern{}, fmt.Errorf("invalid log pattern identity")
	}
	segments := strings.Split(strings.TrimPrefix(identity, "lp1:p:"), "n")
	var body strings.Builder
	for i, segment := range segments {
		b, err := hex.DecodeString(segment)
		if err != nil {
			return LogPattern{}, fmt.Errorf("invalid log pattern identity")
		}
		if i > 0 {
			body.WriteByte('0')
		}
		body.Write(b)
	}
	value := body.String()
	if len(value) > LogPatternMaxBodyBytes || !utf8.ValidString(value) {
		return LogPattern{}, fmt.Errorf("invalid log pattern identity")
	}
	p := ReferenceLogPattern(value)
	if p.Identity != identity {
		return LogPattern{}, fmt.Errorf("noncanonical log pattern identity")
	}
	return p, nil
}
func logPatternSQL() string {
	// All branches remain bounded even when ClickHouse does not short-circuit.
	capped := "if(length(Body) <= 256, Body, '')"
	bounded := "if(isValidUTF8(" + capped + "), " + capped + ", '')"
	shape := "concat('lp1:p:', arrayStringConcat(arrayMap(p -> lower(hex(p)), splitByRegexp('[0-9]+', " + bounded + ")), 'n'))"
	return "multiIf(length(Body) > 256, 'lp1:o', NOT isValidUTF8(" + capped + "), 'lp1:u', Body = '', 'lp1:e', " + shape + ")"
}

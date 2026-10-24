package verification

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Scan bytes without regexp, production pattern encoders or numeric parsing.
// A literal chunk is encoded losslessly; maximal ASCII digit runs delimit chunks.
func lpOracle(body string) exploration.LogPattern {
	if len(body) > 256 {
		return exploration.LogPattern{Identity: "lp1:o", Label: "Body exceeds 256-byte bound · unclassified", State: "oversize"}
	}
	if !utf8.ValidString(body) {
		return exploration.LogPattern{Identity: "lp1:u", Label: "Invalid UTF-8 body · unclassified", State: "invalid-utf8"}
	}
	if body == "" {
		return exploration.LogPattern{Identity: "lp1:e", Label: "Empty body · unclassified", State: "empty"}
	}
	chunks := []string{}
	start := 0
	for i := 0; i < len(body); {
		if body[i] < '0' || body[i] > '9' {
			i++
			continue
		}
		chunks = append(chunks, body[start:i])
		for i < len(body) && body[i] >= '0' && body[i] <= '9' {
			i++
		}
		start = i
	}
	chunks = append(chunks, body[start:])
	encoded, labels := []string{}, []string{}
	for _, chunk := range chunks {
		encoded = append(encoded, hex.EncodeToString([]byte(chunk)))
		var label strings.Builder

		for _, r := range chunk {
			if r == '\n' || r == '\r' || r == '\t' {
				label.WriteByte('\\')
				label.WriteByte(map[rune]byte{'\n': 'n', '\r': 'r', '\t': 't'}[r])
				continue
			}
			if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == '\u2028' || r == '\u2029' {
				if r <= 0xffff {
					fmt.Fprintf(&label, `\u%04x`, r)
				} else {
					fmt.Fprintf(&label, `\U%08x`, r)
				}
				continue
			}
			if r == '\\' || r == '⟪' {
				label.WriteByte('\\')
			}
			label.WriteRune(r)
		}
		labels = append(labels, label.String())
	}
	return exploration.LogPattern{Identity: "lp1:p:" + strings.Join(encoded, "n"), Label: strings.Join(labels, "⟪digits⟫"), State: "pattern"}
}

// Validate the ID grammar directly: lowercase hex literal chunks containing no
// digit bytes, nonempty interior chunks, representative UTF-8 length <=256.
func lpCanonical(id string) bool {
	if id == "lp1:e" || id == "lp1:o" || id == "lp1:u" {
		return true
	}
	if len(id) > 518 || !strings.HasPrefix(id, "lp1:p:") {
		return false
	}
	chunks := strings.Split(id[6:], "n")
	var body strings.Builder
	for i, chunk := range chunks {
		if len(chunk)%2 != 0 || i > 0 && i < len(chunks)-1 && chunk == "" {
			return false
		}
		for _, c := range []byte(chunk) {
			if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
				return false
			}
		}
		raw, e := hex.DecodeString(chunk)
		if e != nil {
			return false
		}
		for _, c := range raw {
			if c >= '0' && c <= '9' {
				return false
			}
		}
		if i > 0 {
			body.WriteByte('0')
		}
		body.Write(raw)
	}
	return body.Len() > 0 && body.Len() <= 256 && utf8.ValidString(body.String())
}
func lpCorpus(seed int) []string {
	return []string{"", fmt.Sprintf("order %d ready", seed), "order 999999999 ready", "order ⟪digits⟫ ready", "order <num> ready", "n", "lp1:p:n", "12", "0", "-12.34e5", "１２ ١٢ 1", "雪12\n\t\x00\\⟪digits⟫", "\uFEFF\u202E\U000E0001\u2028\u2029", strings.Repeat("雪", 85) + "1", strings.Repeat("x", 256), strings.Repeat("x", 257), strings.Repeat("a1", 128), string([]byte{0xff}), strings.Repeat("x", 256) + string([]byte{0xff})}
}
func TestHegelLogPatternIndependentProjection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		repeat := hegel.Draw(tc, hegel.Integers(1, 50))
		literal := hegel.Draw(tc, hegel.SampledFrom([]string{"雪", "⟪digits⟫", "\\", "n", "<num>", "１２", "a\x00\t\n"}))
		bodies := append(lpCorpus(seed), strings.Repeat(literal, repeat)+fmt.Sprint(seed))
		for _, body := range bodies {
			before := body
			actual := exploration.ReferenceLogPattern(body)
			expected := lpOracle(body)
			if actual != expected || body != before || len(actual.Identity) > 518 {
				tc.Errorf("byte projection %q: %+v != %+v", body, actual, expected)
			}
			decoded, e := exploration.DecodeLogPattern(expected.Identity)
			if e != nil || decoded != expected || !lpCanonical(expected.Identity) {
				tc.Errorf("canonical representative %q: %+v %v", expected.Identity, decoded, e)
			}
		}
		a, b := fmt.Sprintf("order %d ready", seed), "order 000999 ready"
		if exploration.ReferenceLogPattern(a) != exploration.ReferenceLogPattern(b) {
			tc.Errorf("digit run magnitude inferred")
		}
		if exploration.ReferenceLogPattern(a).Identity == exploration.ReferenceLogPattern("order ⟪digits⟫ ready").Identity {
			tc.Errorf("display marker collided with literal")
		}
		// Distinct long strings share only the explicitly unclassified sentinel.
		for _, body := range []string{strings.Repeat("x", 257), strings.Repeat("雪", 100)} {
			if p := exploration.ReferenceLogPattern(body); p.Identity != "lp1:o" || p.State == "pattern" {
				tc.Errorf("oversize inferred shape")
			}
		}
	}, options(80)...)
}
func TestHegelLogPatternCanonicalPredicateBoundary(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		field := exploration.Field{Kind: "column", Key: "body_pattern_v1"}
		q := exploration.Query{Signal: "logs", Environment: "lp-property", FromMS: 1000, ToMS: 61000, Limit: 100, GroupBy: []exploration.Field{field}, BucketMS: 10000}
		id := lpOracle(fmt.Sprintf("order %d done", seed)).Identity
		ids := []string{id, "lp1:e", "lp1:o", "lp1:u", "lp1:p:n", "lp1:p:", "lp1:p:nn", "lp1:p:31", "lp1:p:6A", "lp1:p:a", "lp1:p:ff", "lp1:p:" + strings.Repeat("61", 257), "lp1:eextra", "lp2:p:n", strings.Replace(id, "6f", "6F", 1), id + "n"}
		for _, candidate := range ids {
			_, e := exploration.DecodeLogPattern(candidate)
			valid := lpCanonical(candidate)
			if (e == nil) != valid {
				tc.Errorf("decoder canonical grammar %q valid=%v error=%v", candidate, valid, e)
			}
			for _, op := range []string{"eq", "ne"} {
				q.Where = &exploration.Predicate{Field: &field, Op: op, Value: candidate}
				if (q.Validate() == nil) != valid {
					tc.Errorf("predicate validation %s %q", op, candidate)
				}
			}
		}
		q.Where = &exploration.Predicate{Field: &field, Op: "contains", Value: "lp1"}
		if q.Validate() == nil {
			tc.Errorf("encoded substring predicate accepted")
		}
		q.Where = &exploration.Predicate{Field: &field, Op: "eq", Value: id}
		plan, e := exploration.Compile(q)
		if e != nil || plan.GroupedTimeSeries != "" || plan.TimeSeries == "" || strings.Contains(strings.ToUpper(plan.Count), " LIMIT ") {
			tc.Errorf("record-grain grouping/trend boundary %v", e)
		}
		for _, body := range lpCorpus(seed) {
			expected := lpOracle(body).Identity == id
			values := map[exploration.Field]string{field: lpOracle(body).Identity}
			if exploration.Match(q.Where, values) != expected {
				tc.Errorf("exact pattern record membership")
			}
			q.Where.Op = "ne"
			if exploration.Match(q.Where, values) == expected {
				tc.Errorf("pattern complement not all parent rest")
			}
			q.Where.Op = "eq"
		}
		q.Signal = "spans"
		if q.Validate() == nil {
			tc.Errorf("log-only field accepted for spans")
		}
	}, options(70)...)
}

type lpFrozen struct {
	Bodies []string
	Groups map[string]int
	Wire   string
}
type lpMachine struct {
	Bodies []string
	Frozen []lpFrozen
}

func (m *lpMachine) RuleAppend(tc hegel.TestCase) {
	c := lpCorpus(hegel.Draw(tc, hegel.Integers(1, 1000000)))
	m.Bodies = append(m.Bodies, c[hegel.Draw(tc, hegel.Integers(0, len(c)-1))])
}
func (m *lpMachine) RuleReplace(tc hegel.TestCase) {
	if len(m.Bodies) == 0 {
		return
	}
	c := lpCorpus(77)
	m.Bodies[hegel.Draw(tc, hegel.Integers(0, len(m.Bodies)-1))] = c[hegel.Draw(tc, hegel.Integers(0, len(c)-1))]
}
func lpObservedGroups(bodies []string) map[string]int {
	out := map[string]int{}
	for _, b := range bodies {
		out[exploration.ReferenceLogPattern(b).Identity]++
	}
	return out
}
func (m *lpMachine) RuleFreeze(tc hegel.TestCase) {
	g := lpObservedGroups(m.Bodies)
	wire, _ := json.Marshal(g)
	m.Frozen = append(m.Frozen, lpFrozen{append([]string(nil), m.Bodies...), g, string(wire)})
}
func (m *lpMachine) InvariantRecordCountsAndCustody(tc hegel.TestCase) {
	want := map[string]int{}
	for _, b := range m.Bodies {
		want[lpOracle(b).Identity]++
	}
	if !reflect.DeepEqual(want, lpObservedGroups(m.Bodies)) {
		tc.Errorf("record grouping changed after edit")
	}
	for _, f := range m.Frozen {
		expected := map[string]int{}
		for _, b := range f.Bodies {
			expected[lpOracle(b).Identity]++
		}
		wire, _ := json.Marshal(f.Groups)
		if !reflect.DeepEqual(f.Groups, expected) || string(wire) != f.Wire {
			tc.Errorf("frozen group snapshot reinterpreted")
		}
	}
}
func TestHegelLogPatternRecordStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		hegel.RunStateful(tc, &lpMachine{}, hegel.WithStatefulStepCount(20), hegel.WithAlwaysCheckInvariants("InvariantRecordCountsAndCustody"))
	}, options(70)...)
}

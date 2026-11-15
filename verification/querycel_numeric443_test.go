package verification

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
	"hegel.dev/go/hegel"
)

type numericCase443 struct {
	typ       string
	scale     int
	threshold string
}

var numericCases443 = []numericCase443{{"int64", 0, "-7"}, {"uint64", 0, "9007199254740993"}, {"decimal64", 2, "1.20"}}
var numericOperations443 = []string{"eq", "ne", "gt", "gte", "lt", "lte", "valid", "missing", "empty", "invalid"}
var numericValues443 = []string{"", "0", "-0", "+1", "01", " 7", "7 ", "1e2", "NaN", "Infinity", "-7", "7", "9007199254740992", "9007199254740993", "18446744073709551615", "18446744073709551616", "9223372036854775807", "9223372036854775808", "-9223372036854775808", "-9223372036854775809", "0.00", "-0.00", "-0.01", "1.2", "1.20", "1.200", "01.20", "92233720368547758.07", "92233720368547758.08", "-92233720368547758.08", "-92233720368547758.09", "雪"}

// Deliberately independent of Goo's parser: classify the reported text using
// decimal digits, sign, scale and arbitrary-precision range comparisons.
func numericCoefficient443(raw string, tc numericCase443) (*big.Int, bool) {
	negative := strings.HasPrefix(raw, "-")
	unsigned := raw
	if negative {
		unsigned = raw[1:]
	}
	if unsigned == "" {
		return nil, false
	}
	parts := strings.Split(unsigned, ".")
	if tc.typ == "decimal64" {
		if len(parts) != 2 || len(parts[1]) != tc.scale {
			return nil, false
		}
	} else if len(parts) != 1 {
		return nil, false
	}
	if parts[0] == "" || len(parts[0]) > 1 && parts[0][0] == '0' {
		return nil, false
	}
	for _, part := range parts {
		for _, digit := range part {
			if digit < '0' || digit > '9' {
				return nil, false
			}
		}
	}
	coefficient, ok := new(big.Int).SetString(strings.Join(parts, ""), 10)
	if !ok || negative && coefficient.Sign() == 0 {
		return nil, false
	}
	if negative {
		coefficient.Neg(coefficient)
	}
	low, high := new(big.Int), new(big.Int)
	if tc.typ == "uint64" {
		high.SetString("18446744073709551615", 10)
	} else {
		low.SetString("-9223372036854775808", 10)
		high.SetString("9223372036854775807", 10)
	}
	return coefficient, coefficient.Cmp(low) >= 0 && coefficient.Cmp(high) <= 0
}
func numericMeaning443(tc numericCase443, op, raw string, present bool) bool {
	actual, valid := numericCoefficient443(raw, tc)
	switch op {
	case "missing":
		return !present
	case "empty":
		return present && raw == ""
	case "invalid":
		return present && raw != "" && !valid
	case "valid":
		return present && valid
	}
	if !present || !valid {
		return false
	}
	threshold, ok := numericCoefficient443(tc.threshold, tc)
	if !ok {
		panic("invalid test threshold")
	}
	order := actual.Cmp(threshold)
	switch op {
	case "eq":
		return order == 0
	case "ne":
		return order != 0
	case "gt":
		return order > 0
	case "gte":
		return order >= 0
	case "lt":
		return order < 0
	case "lte":
		return order <= 0
	}
	panic("unknown test operation")
}
func numericExpression443(kind, key, unit, op string, tc numericCase443) string {
	scope := "record.attributes"
	if kind == "resource" {
		scope = "resource.attributes"
	}
	call := scope + "." + tc.typ + "(" + celLiteral(key) + ", " + celLiteral(unit)
	if tc.typ == "decimal64" {
		call += ", " + strconv.Itoa(tc.scale)
	}
	call += ")." + op + "("
	if op != "valid" && op != "missing" && op != "empty" && op != "invalid" {
		call += celLiteral(tc.threshold)
	}
	return call + ")"
}

func TestHegelCELNumeric443Parity(t *testing.T) {
	hegel.Test(t, func(h *hegel.T) {
		signal := hegel.Draw(h, hegel.SampledFrom([]string{"logs", "spans"}))
		kind := hegel.Draw(h, hegel.SampledFrom([]string{"resource", "attribute"}))
		key := hegel.Draw(h, hegel.SampledFrom([]string{"http.response.status_code", "message.body.size", "quote\"'\\_雪"}))
		tc := hegel.Draw(h, hegel.SampledFrom(numericCases443))
		op := hegel.Draw(h, hegel.SampledFrom(numericOperations443))
		raw := hegel.Draw(h, hegel.SampledFrom(numericValues443))
		present := hegel.Draw(h, hegel.Booleans())
		if hegel.Draw(h, hegel.Booleans()) {
			n := hegel.Draw(h, hegel.Integers(-10000, 10000))
			raw = strconv.Itoa(n)
			if tc.typ == "decimal64" {
				raw += ".20"
			}
		}
		expression := numericExpression443(kind, key, "1", op, tc)
		program, err := querycel.Compile(signal, expression)
		if err != nil {
			h.Fatal(expression, err)
		}
		record := map[exploration.Field]string{}
		if present {
			record[exploration.Field{Kind: kind, Key: key}] = raw
		}
		other := "resource"
		if kind == "resource" {
			other = "attribute"
		}
		record[exploration.Field{Kind: other, Key: key}] = tc.threshold
		got, err := program.Eval(record)
		want := numericMeaning443(tc, op, raw, present)
		if err != nil || got != want || exploration.Match(program.Predicate(), record) != want {
			h.Fatal("numeric CEL/AST/independent oracle differ", expression, raw, present, got, want, err)
		}
		printed, err := querycel.PrintPredicate(signal, program.Predicate())
		if err != nil {
			h.Fatal(err)
		}
		round, err := querycel.Compile(signal, printed)
		if err != nil || !reflect.DeepEqual(program.Predicate(), round.Predicate()) {
			h.Fatal("numeric inverse changed binding, threshold or branch", printed, err)
		}
		parent := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 7, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}}
		before, _ := json.Marshal(parent)
		refined, err := program.Apply(parent)
		if err != nil {
			h.Fatal(err)
		}
		if refined.Environment != parent.Environment || refined.FromMS != parent.FromMS || refined.ToMS != parent.ToMS || refined.Limit != parent.Limit || !reflect.DeepEqual(refined.Where.All[0], *parent.Where) {
			h.Fatal("typed numeric condition changed parent scope")
		}
		refined.Where.All[0].Value = "changed"
		after, _ := json.Marshal(parent)
		if string(before) != string(after) {
			h.Fatal("numeric apply aliases parent")
		}
	}, options(240)...)
}

// Exhaust all exceptional text boundaries, so random selection cannot miss an
// invalid/missing ne result or a one-coefficient overflow at either signed edge.
func TestCELNumeric443CoveragePartitions(t *testing.T) {
	for _, signal := range []string{"logs", "spans"} {
		for _, kind := range []string{"resource", "attribute"} {
			for _, tc := range numericCases443 {
				programs := map[string]*querycel.Program{}
				for _, op := range numericOperations443 {
					p, err := querycel.Compile(signal, numericExpression443(kind, "k", "1", op, tc))
					if err != nil {
						t.Fatal(err)
					}
					programs[op] = p
				}
				for _, raw := range numericValues443 {
					for _, present := range []bool{false, true} {
						record := map[exploration.Field]string{}
						if present {
							record[exploration.Field{Kind: kind, Key: "k"}] = raw
						}
						states := 0
						for _, op := range numericOperations443 {
							p := programs[op]
							got, err := p.Eval(record)
							want := numericMeaning443(tc, op, raw, present)
							if err != nil || got != want || exploration.Match(p.Predicate(), record) != want {
								t.Fatalf("%s %s %s %s value=%q present=%v got=%v want=%v err=%v", signal, kind, tc.typ, op, raw, present, got, want, err)
							}
							if got && (op == "valid" || op == "missing" || op == "empty" || op == "invalid") {
								states++
							}
						}
						if states != 1 {
							t.Fatalf("numeric states do not partition text: %s raw=%q present=%v states=%d", tc.typ, raw, present, states)
						}
					}
				}
			}
		}
	}
}

// The real adapter supplies typed parameters and canonical parse SQL. Only the
// physical relation is replaced by a read-only generated map fixture.
func TestCELNumeric443ClickHouseParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CEL_CLICKHOUSE_URL")
	if endpoint == "" {
		t.Skip("requires GOO_CEL_CLICKHOUSE_URL for live numeric parity")
	}
	const key = "otel.test.numeric443"
	encoded := make([]string, len(numericValues443))
	for i, v := range numericValues443 {
		encoded[i] = "unhex('" + hex.EncodeToString([]byte(v)) + "')"
	}
	for _, signal := range []string{"logs", "spans"} {
		for _, kind := range []string{"resource", "attribute"} {
			column, table := "LogAttributes", "otel_logs"
			if signal == "spans" {
				column, table = "SpanAttributes", "otel_traces"
			}
			reported := "if(number=0,map(),map('" + key + "',arrayElement([" + strings.Join(encoded, ",") + "],number)))"
			resource, attribute := "map('deployment.environment.name','cel-numeric443')", "map()"
			if kind == "resource" {
				resource = "mapConcat(" + resource + "," + reported + ")"
			} else {
				attribute = reported
			}
			relation := "(SELECT fromUnixTimestamp64Milli(1000) AS Timestamp," + resource + " AS ResourceAttributes," + attribute + " AS " + column + " FROM numbers(" + strconv.Itoa(len(numericValues443)+1) + "))"
			for _, tc := range numericCases443 {
				for _, op := range numericOperations443 {
					t.Run(signal+"/"+kind+"/"+tc.typ+"/"+op, func(t *testing.T) {
						p, err := querycel.Compile(signal, numericExpression443(kind, key, "1", op, tc))
						if err != nil {
							t.Fatal(err)
						}
						q, err := p.Apply(exploration.Query{Signal: signal, Environment: "cel-numeric443", FromMS: 1000, ToMS: 61000, Limit: 1})
						if err != nil {
							t.Fatal(err)
						}
						plan, err := exploration.Compile(q)
						if err != nil {
							t.Fatal(err)
						}
						boundary := " FROM " + table + " WHERE "
						if strings.Count(plan.Count, boundary) != 1 {
							t.Fatal("physical boundary changed")
						}
						rows, err := reportedCH353(t.Context(), endpoint, strings.Replace(plan.Count, boundary, " FROM "+relation+" WHERE ", 1), plan.Parameters)
						if err != nil || len(rows) != 1 {
							t.Fatal("live numeric count failed", err, len(rows))
						}
						var count string
						if err = json.Unmarshal(rows[0]["count"], &count); err != nil {
							t.Fatal(err)
						}
						want := 0
						if numericMeaning443(tc, op, "", false) {
							want++
						}
						for _, raw := range numericValues443 {
							if numericMeaning443(tc, op, raw, true) {
								want++
							}
						}
						if count != strconv.Itoa(want) {
							t.Fatalf("source count=%s independent oracle=%d", count, want)
						}
					})
				}
			}
		}
	}
}

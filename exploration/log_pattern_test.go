package exploration

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func logPatternBodies() []string {
	return []string{"", "order 12 ready", "order 999 ready", "order <num> ready", "literal ⟪digits⟫ \\ 4", "-12.34e5", "v1 127.0.0.1 a12b", "１２ ١٢", "雪12\n\t\x00", "1", "00", strings.Repeat("x", 256), strings.Repeat("x", 257), string([]byte{0xff}), string([]byte{0xff}) + strings.Repeat("x", 256)}
}
func TestLogPatternCanonicalIdentityAndLiteralSafety(t *testing.T) {
	for _, body := range logPatternBodies() {
		p := ReferenceLogPattern(body)
		got, e := DecodeLogPattern(p.Identity)
		if e != nil || got != p {
			t.Fatalf("roundtrip %q: %#v %#v %v", body, p, got, e)
		}
		if len(p.Identity) > 518 {
			t.Fatal("identity budget")
		}
	}
	a, b := ReferenceLogPattern("order 12 ready"), ReferenceLogPattern("order 999 ready")
	if a != b || a.Label != "order ⟪digits⟫ ready" {
		t.Fatal(a, b)
	}
	if ReferenceLogPattern("order <num> ready").Identity == a.Identity {
		t.Fatal("literal collision")
	}
	if ReferenceLogPattern("literal ⟪digits⟫ \\ 4").Label != "literal \\⟪digits⟫ \\\\ ⟪digits⟫" {
		t.Fatal("ambiguous literal display")
	}
	if ReferenceLogPattern("１２ ١٢").Label != "１２ ١٢" {
		t.Fatal("non ASCII inference")
	}
	for _, id := range []string{"", "lp2:e", "lp1:p:", "lp1:p:nn", "lp1:p:31", "lp1:p:6A", "lp1:p:a", "lp1:p:ff", "lp1:p:" + strings.Repeat("61", 257), "lp1:eextra"} {
		if _, e := DecodeLogPattern(id); e == nil {
			t.Fatal("accepted noncanonical", id)
		}
	}
}
func TestLogPatternLabelsExposeControlAndPlaceholderLiterals(t *testing.T) {
	for _, tc := range []struct{ body, label string }{
		{"a\nb", `a\nb`}, {`a\nb`, `a\\nb`},
		{"\x00\t\r", `\u0000\t\r`},
		{"雪\u200b\u2028\u2029", `雪\u200b\u2028\u2029`},
		{"\U000e0001⟪digits⟫12", `\U000e0001\⟪digits⟫⟪digits⟫`},
	} {
		if got := ReferenceLogPattern(tc.body).Label; got != tc.label {
			t.Fatalf("label %q: %q != %q", tc.body, got, tc.label)
		}
	}
	if ReferenceLogPattern("a\nb").Identity == ReferenceLogPattern(`a\nb`).Identity {
		t.Fatal("raw backslash literal collided")
	}
}

func TestLogPatternFieldSignalPredicatesAndOptionalRows(t *testing.T) {
	f := Field{Kind: "column", Key: "body_pattern_v1"}
	if _, e := f.Type("logs"); e != nil {
		t.Fatal(e)
	}
	if _, e := f.Type("spans"); e == nil {
		t.Fatal("span field accepted")
	}
	q := Query{Signal: "logs", Environment: "lab", FromMS: 1000, ToMS: 61000, Limit: 10}
	legacy, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(legacy.Rows, "lp1:") {
		t.Fatal("legacy row SQL changed")
	}
	q.GroupBy = []Field{f}
	q.BucketMS = 10000
	p, e := Compile(q)
	if e != nil {
		t.Fatal(e)
	}
	if !strings.Contains(p.Groups, "splitByRegexp('[0-9]+'") || !strings.Contains(p.Rows, "AS body_pattern_v1") || p.GroupedTimeSeries != "" || p.TimeSeries == "" {
		t.Fatal("missing bounded record-grain aggregation")
	}
	for _, op := range []string{"eq", "ne", "exists", "missing"} {
		value := ""
		if op == "eq" || op == "ne" {
			value = ReferenceLogPattern("x1").Identity
		}
		q.Where = &Predicate{Field: &f, Op: op, Value: value}
		if e := q.Validate(); e != nil {
			t.Fatal(op, e)
		}
	}
	q.Where = &Predicate{Field: &f, Op: "eq", Value: "lp1:p:31"}
	if q.Validate() == nil {
		t.Fatal("noncanonical filter accepted")
	}
	q.Where = &Predicate{Field: &f, Op: "contains", Value: "lp1"}
	if q.Validate() == nil {
		t.Fatal("encoded substring accepted")
	}
}
func TestClickHouseLogPatternParity(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL not configured")
	}
	source := Source{Endpoint: endpoint, User: os.Getenv("GOO_CLICKHOUSE_TEST_USER"), Password: os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD")}
	bodies := logPatternBodies()
	bodies = append(bodies, "abc123def456", "abc9def0", "2026-10-24T01:02:03Z", "550e8400-e29b-41d4-a716-446655440000", "12\n", "n", "<num>", "⟪digits⟫", strings.Repeat("a1", 128), strings.Repeat("雪", 85)+"1")
	encoded := []string{}
	for _, body := range bodies {
		encoded = append(encoded, hex.EncodeToString([]byte(body)))
	}
	data, _ := json.Marshal(encoded)
	inline := " FROM (SELECT number AS n, unhex(JSONExtract({bodies:String}, 'Array(String)')[number+1]) AS Body FROM numbers(" + fmt.Sprint(len(bodies)) + "))"
	params := map[string]string{"bodies": string(data)}
	rows, e := source.read(context.Background(), "SELECT toString(n) AS n, "+logPatternSQL()+" AS identity"+inline+" ORDER BY toUInt64(n) FORMAT JSONEachRow", params)
	if e != nil || len(rows) != len(bodies) {
		t.Fatal(e, len(rows))
	}
	counts := map[string]int{}
	for i, row := range rows {
		var v struct{ Identity string }
		if e := json.Unmarshal(row, &v); e != nil {
			t.Fatal(e)
		}
		want := ReferenceLogPattern(bodies[i])
		if v.Identity != want.Identity {
			t.Fatalf("parity %q: %s != %s", bodies[i], v.Identity, want.Identity)
		}
		counts[want.Identity]++
	}
	grouped, e := source.read(context.Background(), "SELECT "+logPatternSQL()+" AS identity, toString(count()) AS count"+inline+" GROUP BY identity FORMAT JSONEachRow", params)
	if e != nil {
		t.Fatal(e)
	}
	got := map[string]int{}
	for _, row := range grouped {
		var v struct{ Identity, Count string }
		json.Unmarshal(row, &v)
		var n int
		fmt.Sscan(v.Count, &n)
		got[v.Identity] = n
	}
	if !reflect.DeepEqual(got, counts) {
		t.Fatal("full population group counts", got, counts)
	}
	target := ReferenceLogPattern("order 12 ready").Identity
	for _, op := range []string{"eq", "ne"} {
		c := compiler{signal: "logs", params: map[string]string{"bodies": string(data)}}
		where := c.predicate(&Predicate{Field: &Field{Kind: "column", Key: "body_pattern_v1"}, Op: op, Value: target})
		rows, e := source.read(context.Background(), "SELECT toString(n) AS n"+inline+" WHERE "+where+" ORDER BY toUInt64(n) FORMAT JSONEachRow", c.params)
		if e != nil {
			t.Fatal(e)
		}
		expected := 0
		for _, body := range bodies {
			match := ReferenceLogPattern(body).Identity == target
			if match == (op == "eq") {
				expected++
			}
		}
		if len(rows) != expected {
			t.Fatal("exact partition", op, len(rows), expected)
		}
	}
}

func TestLogPatternSQLCapsBeforeUTF8Validation(t *testing.T) {
	sql := logPatternSQL()
	if strings.Contains(sql, "isValidUTF8(Body)") || !strings.Contains(sql, "isValidUTF8(if(length(Body) <= 256, Body, ''))") {
		t.Fatal("UTF-8 validation must receive only capped input", sql)
	}
}
func TestClickHouseLogPatternLongBodiesWithoutShortCircuit(t *testing.T) {
	endpoint := os.Getenv("GOO_CLICKHOUSE_TEST_URL")
	if endpoint == "" {
		t.Skip("GOO_CLICKHOUSE_TEST_URL not configured")
	}
	bodies := []string{strings.Repeat("雪1", 16384), string([]byte{0xff}) + strings.Repeat("x", 65536), "ok12", string([]byte{0xff})}
	inline := " FROM (SELECT number AS n, [repeat('雪1', 16384), concat(unhex('ff'), repeat('x', 65536)), 'ok12', unhex('ff')][number+1] AS Body FROM numbers(4))"
	// readonly=2 forbids writes while allowing this query-local setting. The
	// production adapter's readonly=1 correctly refuses any settings mutation.
	endpointURL, e := url.Parse(endpoint)
	if e != nil {
		t.Fatal(e)
	}
	settings := endpointURL.Query()
	settings.Set("readonly", "2")
	endpointURL.RawQuery = settings.Encode()
	sql := "SELECT toString(n) AS n, " + logPatternSQL() + " AS identity" + inline + " ORDER BY toUInt64(n) SETTINGS short_circuit_function_evaluation='disable' FORMAT JSONEachRow"
	request, e := http.NewRequest("POST", endpointURL.String(), strings.NewReader(sql))
	if e != nil {
		t.Fatal(e)
	}
	request.SetBasicAuth(os.Getenv("GOO_CLICKHOUSE_TEST_USER"), os.Getenv("GOO_CLICKHOUSE_TEST_PASSWORD"))
	response, e := (&http.Client{Timeout: 3 * time.Second}).Do(request)
	if e != nil {
		t.Fatal(e)
	}
	defer response.Body.Close()
	raw, e := io.ReadAll(io.LimitReader(response.Body, 4096))
	if e != nil || response.StatusCode != 200 {
		t.Fatalf("read-only fixture: status=%d error=%v response=%s", response.StatusCode, e, raw)
	}
	rows := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if e != nil || len(rows) != len(bodies) {
		t.Fatal(e, len(rows))
	}
	for i, row := range rows {
		var got struct{ Identity string }
		if e := json.Unmarshal([]byte(row), &got); e != nil {
			t.Fatal(e)
		}
		if want := ReferenceLogPattern(bodies[i]).Identity; got.Identity != want {
			t.Fatalf("long body without short circuit %d: %q != %q", i, got.Identity, want)
		}
	}
}

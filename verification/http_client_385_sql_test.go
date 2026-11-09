package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

// Read-only virtual-source execution complements the generated interpreter and
// the real journey; no source table or successful response is replaced.
func TestHTTPClientCompiledSQL385(t *testing.T) {
	endpoint := os.Getenv("GOO_HTTP_CLIENT_385_CHURL")
	if endpoint == "" {
		t.Skip("explicit read-only ClickHouse fixture endpoint required")
	}
	const cte = "WITH fixture AS (SELECT fromUnixTimestamp64Milli(toInt64(JSONExtractString(row,'timestamp_ms'))) AS Timestamp, JSONExtractString(row,'ServiceName') AS ServiceName, JSONExtractString(row,'SpanId') AS SpanId, JSONExtractString(row,'SpanKind') AS SpanKind, JSONExtract(row,'ResourceAttributes','Map(String,String)') AS ResourceAttributes, JSONExtract(row,'SpanAttributes','Map(String,String)') AS SpanAttributes, toUInt64(JSONExtractString(row,'duration_ns')) AS Duration, JSONExtractString(row,'StatusCode') AS StatusCode FROM (SELECT arrayJoin(JSONExtractArrayRaw({fixture:String})) AS row)) "
	for index := range 20 {
		f := httpFixtureFor385(3850+index, index%5, index/5)
		plan, e := httpPlan385(f, "http-client", nil)
		if e != nil {
			t.Fatal(e)
		}
		for _, widget := range plan.Widgets {
			key := map[string]string{"dimension": "server.address", "http-client-methods": "http.request.method", "http-client-ports": "server.port", "http-client-responses": "http.response.status_code"}[widget.Key]
			if key == "" {
				continue
			}
			q := widget.Query
			if q == nil {
				t.Fatal("missing current Client group question")
			}
			compiled, e := exploration.Compile(*q)
			if e != nil {
				t.Fatal(e)
			}
			// Bounded parameter chunks avoid the HTTP form-field limit without changing
			// source limits, fixture records, or the compiled population predicate.
			params := map[string]string{}
			wire := continuationJSON352(f.Rows)
			chunks := []string{}
			for offset := 0; offset < len(wire); offset += 16000 {
				key := fmt.Sprintf("fixture%d", len(chunks))
				params[key] = wire[offset:min(offset+16000, len(wire))]
				chunks = append(chunks, "{"+key+":String}")
			}
			fixtureCTE := strings.ReplaceAll(cte, "{fixture:String}", "concat("+strings.Join(chunks, ",")+")")
			for k, v := range compiled.Parameters {
				params[k] = v
			}
			const physical = " FROM otel_traces WHERE "
			start := strings.Index(compiled.Count, physical)
			if start < 0 || strings.Count(compiled.Count, "FROM otel_traces") != 1 {
				t.Fatal("unexpected public count source shape")
			}
			where := strings.TrimSuffix(compiled.Count[start:], " FORMAT JSONEachRow")
			counts, e := reportedCH353(t.Context(), endpoint, fixtureCTE+strings.ReplaceAll(compiled.Count, "FROM otel_traces", "FROM fixture"), params)
			if e != nil {
				t.Fatal(e)
			}
			wantSum := new(big.Int)
			wantStatus := map[string]int{"status_error": 0, "status_ok": 0, "status_unset": 0, "status_unknown": 0}
			for _, row := range f.Rows {
				if !httpMember385(f, row) {
					continue
				}
				n, ok := new(big.Int).SetString(row.DurationNS, 10)
				if !ok {
					t.Fatal("invalid oracle duration")
				}
				wantSum.Add(wantSum, n)
				key := "status_unknown"
				switch row.Status {
				case "Error":
					key = "status_error"
				case "Ok":
					key = "status_ok"
				case "Unset":
					key = "status_unset"
				}
				wantStatus[key]++
			}
			if q.Summary == "span-duration" {
				var sum string
				if len(counts) != 1 || json.Unmarshal(counts[0]["sum_ns"], &sum) != nil || sum != wantSum.String() {
					t.Fatal("full duration sum differs from independent Client records")
				}
			}
			if q.SpanStatus {
				for key, want := range wantStatus {
					var n string
					if len(counts) != 1 || json.Unmarshal(counts[0][key], &n) != nil || n != strconv.Itoa(want) {
						t.Fatal("recorded span status inferred from HTTP response or changed scope", key)
					}
				}
			}
			var count string
			if len(counts) != 1 || json.Unmarshal(counts[0]["count"], &count) != nil || count != strconv.Itoa(len(httpIDs385(f))) {
				t.Fatalf("fixture%d/%s full count differs from raw oracle", index, key)
			}
			ids, e := reportedCH353(t.Context(), endpoint, fixtureCTE+"SELECT SpanId AS id"+strings.ReplaceAll(where, "FROM otel_traces", "FROM fixture")+" ORDER BY id FORMAT JSONEachRow", params)
			if e != nil {
				t.Fatal(e)
			}
			got := []string{}
			for _, r := range ids {
				var id string
				if json.Unmarshal(r["id"], &id) != nil {
					t.Fatal("invalid actual SQL identity")
				}
				got = append(got, id)
			}
			slices.Sort(got)
			if !slices.Equal(got, httpIDs385(f)) {
				t.Fatal("actual SQL role/method/scope membership differs")
			}
			groups, e := reportedCH353(t.Context(), endpoint, fixtureCTE+strings.ReplaceAll(compiled.Groups, "FROM otel_traces", "FROM fixture"), params)
			if e != nil {
				t.Fatal(e)
			}
			gotGroups := map[string]int{}
			for _, r := range groups {
				var value *string
				var n string
				if json.Unmarshal(r["g0"], &value) != nil || json.Unmarshal(r["count"], &n) != nil {
					t.Fatal("malformed literal group")
				}
				number, e := strconv.Atoi(n)
				if e != nil {
					t.Fatal(e)
				}
				identity := continuationJSON352(value)
				if _, exists := gotGroups[identity]; exists {
					t.Fatal("duplicate qualified group")
				}
				gotGroups[identity] = number
			}
			if !reflect.DeepEqual(gotGroups, httpGroupCounts385(f, key)) {
				t.Fatalf("actual SQL collapsed missing/empty/literal values: %v/%v", gotGroups, httpGroupCounts385(f, key))
			}
		}
	}
}

package verification

import (
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
)

func TestCurrentRPCCompiledSQL357(t *testing.T) {
	endpoint := os.Getenv("GOO_RPC_357_CHURL")
	if endpoint == "" {
		t.Skip("explicit read-only ClickHouse fixture endpoint required")
	}
	const cte = "WITH fixture AS (SELECT fromUnixTimestamp64Milli(JSONExtractInt(row,'timestamp_ms')) AS Timestamp, JSONExtractString(row,'ServiceName') AS ServiceName, JSONExtractString(row,'SpanId') AS SpanId, JSONExtractString(row,'SpanKind') AS SpanKind, JSONExtract(row,'ResourceAttributes','Map(String,String)') AS ResourceAttributes, JSONExtract(row,'SpanAttributes','Map(String,String)') AS SpanAttributes, toUInt64(1000000) AS Duration, 'Ok' AS StatusCode FROM (SELECT arrayJoin(JSONExtractArrayRaw({fixture:String})) AS row)) "
	for index := range 20 {
		f := rpcFixtureFor357(3570+index, index%5, index/5)
		for _, profile := range []string{"rpc-server-current", "rpc-server"} {
			dashboard, err := rpcPlan357(f, profile, true, true, nil)
			if err != nil {
				t.Fatal(err)
			}
			var question *exploration.Query
			for _, w := range dashboard.Widgets {
				if w.Key == "dimension" {
					question = w.Query
				}
			}
			if question == nil {
				t.Fatal("missing actual RPC dimension query")
			}
			compiled, err := exploration.Compile(*question)
			if err != nil {
				t.Fatal(err)
			}
			parameters := map[string]string{"fixture": continuationJSON352(f.Rows)}
			for k, v := range compiled.Parameters {
				parameters[k] = v
			}
			const physical = " FROM otel_traces WHERE "
			start := strings.Index(compiled.Count, physical)
			if start < 0 || strings.Count(compiled.Count, "FROM otel_traces") != 1 {
				t.Fatal("unexpected compiled count source shape")
			}
			where := strings.TrimSuffix(compiled.Count[start:], " FORMAT JSONEachRow")
			counts, err := reportedCH353(t.Context(), endpoint, cte+strings.ReplaceAll(compiled.Count, "FROM otel_traces", "FROM fixture"), parameters)
			if err != nil {
				t.Fatalf("fixture%d/%s count: %v", index, profile, err)
			}
			ids, err := reportedCH353(t.Context(), endpoint, cte+"SELECT SpanId AS id"+strings.ReplaceAll(where, "FROM otel_traces", "FROM fixture")+" ORDER BY id FORMAT JSONEachRow", parameters)
			if err != nil {
				t.Fatal(err)
			}
			groups, err := reportedCH353(t.Context(), endpoint, cte+strings.ReplaceAll(compiled.Groups, "FROM otel_traces", "FROM fixture"), parameters)
			if err != nil {
				t.Fatal(err)
			}
			current := profile == "rpc-server-current"
			var count string
			if len(counts) != 1 || json.Unmarshal(counts[0]["count"], &count) != nil || count != rpcCount357(f, current) {
				t.Fatalf("fixture%d/%s independent count %s expected%s", index, profile, count, rpcCount357(f, current))
			}
			gotIDs := []string{}
			for _, r := range ids {
				var id string
				if json.Unmarshal(r["id"], &id) != nil {
					t.Fatal("invalid projected ID")
				}
				gotIDs = append(gotIDs, id)
			}
			slices.Sort(gotIDs)
			if !slices.Equal(gotIDs, rpcIDs357(f, current)) {
				t.Fatal("actual SQL membership differs from independent raw OTel current/legacy oracle")
			}
			expectedGroups := map[string]int{}
			for _, r := range f.Rows {
				if !rpcMember357(f, r, current) {
					continue
				}
				var method *string
				if value, ok := r.Attributes["rpc.method"]; ok {
					copy := value
					method = &copy
				}
				expectedGroups[continuationJSON352(method)]++
			}
			actualGroups := map[string]int{}
			for _, r := range groups {
				var method *string
				var n string
				if json.Unmarshal(r["g0"], &method) != nil || json.Unmarshal(r["count"], &n) != nil {
					t.Fatal("invalid real SQL method group")
				}
				value, err := strconv.Atoi(n)
				if err != nil {
					t.Fatal(err)
				}
				key := continuationJSON352(method)
				if _, duplicate := actualGroups[key]; duplicate {
					t.Fatal("duplicate method group")
				}
				actualGroups[key] = value
			}
			if continuationJSON352(actualGroups) != continuationJSON352(expectedGroups) {
				t.Fatalf("fixture%d/%s null/empty/_OTHER groups changed: actual%v expected%v", index, profile, actualGroups, expectedGroups)
			}
		}
	}
}

package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

// ClickHouse HTTP typed String parameters use Escaped input, independently of
// URL encoding. This test adapter mirrors that physical transport boundary only;
// it never substitutes values into SQL or changes generated predicates.
func reportedEscaped353(value string) string {
	var out strings.Builder
	for _, r := range value {
		if r == '\\' {
			out.WriteString(`\\`)
		} else if r < 32 || r == 127 {
			fmt.Fprintf(&out, `\x%02x`, r)
		} else {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func reportedCH353(ctx context.Context, endpoint, sql string, parameters map[string]string) ([]map[string]json.RawMessage, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid fixture endpoint")
	}
	values := u.Query()
	for key, value := range map[string]string{"readonly": "1", "max_execution_time": "5", "max_result_rows": "2048", "result_overflow_mode": "throw", "max_bytes_to_read": "16777216", "read_overflow_mode": "throw", "max_memory_usage": "134217728"} {
		values.Set(key, value)
	}
	for key, value := range parameters {
		values.Set("param_"+key, reportedEscaped353(value))
	}
	user, password := os.Getenv("GOO_TEST_CLICKHOUSE_USER"), os.Getenv("GOO_TEST_CLICKHOUSE_PASSWORD")
	if user == "" {
		user = "lab"
	}
	if password == "" {
		password = "lab-local"
	}
	if u.User != nil {
		user = u.User.Username()
		password, _ = u.User.Password()
		u.User = nil
	}
	u.RawQuery = values.Encode()
	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), strings.NewReader(sql))
	if err != nil {
		return nil, fmt.Errorf("cannot construct fixture request")
	}
	req.SetBasicAuth(user, password)
	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fixture request failed (%T)", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("fixture response read failed")
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("fixture HTTP %d: %.512s", response.StatusCode, body)
	}
	rows := []map[string]json.RawMessage{}
	for _, line := range strings.Split(strings.TrimSpace(string(body)), "\n") {
		if line == "" {
			continue
		}
		var row map[string]json.RawMessage
		if json.Unmarshal([]byte(line), &row) != nil {
			return nil, fmt.Errorf("invalid fixture response")
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// Twenty stratified cases execute the actual generated ownership predicate over
// a read-only table-function CTE. Only physical table references are redirected;
// no predicate, parameter, aggregate or owner comparison is reconstructed here.
// No table is created and no genuine lab data is inserted, deleted or changed.
func TestReportedEdgeCompiledSQL353(t *testing.T) {
	endpoint := os.Getenv("GOO_REPORTED_EDGE_353_CHURL")
	if endpoint == "" {
		t.Skip("read-only generated ClickHouse fixture opt-in")
	}
	for index := range 20 {
		f := reportedFixtureFor353(1000+index, index%5, index%3, 1+index%3, index%2 == 0)
		q, _, err := reportedUI353(f)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := exploration.Compile(q)
		if err != nil {
			t.Fatal(err)
		}
		parameters := map[string]string{"fixture": continuationJSON352(f.Rows)}
		for key, value := range plan.Parameters {
			parameters[key] = value
		}
		const physical = " FROM otel_traces WHERE "
		start := strings.Index(plan.Count, physical)
		if start < 0 || strings.Count(plan.Count, "FROM otel_traces") != 2 {
			t.Fatal("unexpected physical source shape")
		}
		where := strings.TrimSuffix(plan.Count[start:], " FORMAT JSONEachRow")
		cte := "WITH fixture AS (SELECT fromUnixTimestamp64Milli(JSONExtractInt(row,'timestamp_ms')) AS Timestamp, JSONExtractString(row,'ServiceName') AS ServiceName, JSONExtractString(row,'TraceId') AS TraceId, JSONExtractString(row,'SpanId') AS SpanId, JSONExtractString(row,'ParentSpanId') AS ParentSpanId, JSONExtract(row,'ResourceAttributes','Map(String,String)') AS ResourceAttributes, JSONExtract(row,'SpanAttributes','Map(String,String)') AS SpanAttributes, toUInt64(1000000) AS Duration, 'Ok' AS StatusCode FROM (SELECT arrayJoin(JSONExtractArrayRaw({fixture:String})) AS row)) "
		countSQL := cte + strings.ReplaceAll(plan.Count, "FROM otel_traces", "FROM fixture")
		idsSQL := cte + "SELECT SpanId AS id" + strings.ReplaceAll(where, "FROM otel_traces", "FROM fixture") + " ORDER BY id FORMAT JSONEachRow"
		countRows, err := reportedCH353(t.Context(), endpoint, countSQL, parameters)
		if err != nil {
			t.Fatalf("case %d count: %v", index, err)
		}
		idRows, err := reportedCH353(t.Context(), endpoint, idsSQL, parameters)
		if err != nil {
			t.Fatalf("case %d IDs: %v", index, err)
		}
		wantRows := reportedOracle353(f)
		var count string
		if len(countRows) != 1 || json.Unmarshal(countRows[0]["count"], &count) != nil || count != strconv.Itoa(len(wantRows)) {
			t.Fatalf("case %d count differs from raw ownership oracle: %s vs%d", index, count, len(wantRows))
		}
		want, got := []string{}, []string{}
		for _, row := range wantRows {
			want = append(want, row.ID)
		}
		for _, row := range idRows {
			var id string
			if json.Unmarshal(row["id"], &id) != nil {
				t.Fatal("invalid fixture ID")
			}
			got = append(got, id)
		}
		slices.Sort(want)
		slices.Sort(got)
		if !slices.Equal(want, got) {
			t.Fatalf("case %d child record multiset differs from all-owner oracle: %v vs%v", index, got, want)
		}
	}
}

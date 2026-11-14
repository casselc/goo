package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
)

type consistencyTransport312 func(*http.Request) (*http.Response, error)

func (f consistencyTransport312) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func consistencySource312(count, rows int, during func()) (Source, *int) {
	calls := new(int)
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: consistencyTransport312(func(r *http.Request) (*http.Response, error) {
		*calls++
		if during != nil {
			during()
		}
		sql, _ := io.ReadAll(r.Body)
		body := ""
		switch {
		case strings.Contains(string(sql), "SELECT DISTINCT"):
		case strings.Contains(string(sql), "GROUP BY"):
			if count > 0 {
				body = fmt.Sprintf(`{"g0":"api","count":"%d"}`, count)
			}
		case strings.Contains(string(sql), "toString(count()) AS count"):
			body = fmt.Sprintf(`{"count":"%d"}`, count)
		default:
			for range rows {
				body += "{\"timestamp_ns\":\"1000000001\",\"duration_ns\":\"0\",\"Body\":\"log\",\"SpanName\":\"span\",\"ResourceAttributes\":{\"deployment.environment.name\":\"prod\"}}\n"
			}
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	return source, calls
}
func TestSourcePageConsistencyExhaustiveBoundedStates312(t *testing.T) {
	for count := 0; count <= 8; count++ {
		for offset := 0; offset <= 5; offset++ {
			for limit := 1; limit <= 4; limit++ {
				for rawRows := 0; rawRows <= limit+1; rawRows++ {
					q := Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: limit, Offset: offset}
					source, calls := consistencySource312(count, rawRows, nil)
					result, err := source.FetchRecordPage(context.Background(), q)
					if err != nil {
						t.Fatal(err)
					}
					available := max(count-offset, 0)
					expectedRows := min(limit, available)
					expectedMore := available > limit
					want := "observed"
					if count == 0 {
						want = "no_data"
					}
					if len(result.Rows) != expectedRows || result.HasMore != expectedMore {
						want = "changed_between_reads"
					}
					if result.State != want || *calls != 2 {
						t.Fatalf("count=%d offset=%d limit=%d rawRows=%d: state=%s want=%s SQL=%d", count, offset, limit, rawRows, result.State, want, *calls)
					}
				}
			}
		}
	}
}
func TestSourceShortTerminalPageRetainsDiagnosticsWithoutAttestation312(t *testing.T) {
	for _, full := range []bool{false, true} {
		q := Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
		source, calls := consistencySource312(5, 1, nil)
		var result *Result
		var err error
		if full {
			result, err = source.Fetch(context.Background(), q)
		} else {
			result, err = source.FetchRecordPage(context.Background(), q)
		}
		if err != nil {
			t.Fatal(err)
		}
		if result.State != "changed_between_reads" || result.Count != "5" || len(result.Rows) != 1 || result.HasMore {
			t.Fatal("terminal page disagreement lost", result)
		}
		if full {
			if *calls != 3 || result.SummaryDataset == nil || result.SummaryDataset.Completeness != nil {
				t.Fatal("disagreeing full read gained attestation", result.SummaryDataset)
			}
		} else if *calls != 2 {
			t.Fatal("narrow page gained source work", *calls)
		}
	}
}
func TestSourceQueryDetachedBeforeIOAndAfterReturn312(t *testing.T) {
	for _, full := range []bool{false, true} {
		q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 1, GroupBy: []Field{{Kind: "column", Key: "service"}}, Where: &Predicate{All: []Predicate{{Field: &Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "core"}, {Any: []Predicate{{Field: &Field{Kind: "resource", Key: "zone"}, Op: "eq", Value: "west"}, {Field: &Field{Kind: "resource", Key: "zone"}, Op: "missing"}}}}}}
		q.Environment = ""
		q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		before := detached(q)
		source, _ := consistencySource312(1, 1, func() { q.Where.All[0].Value = "during-io" })
		var result *Result
		var err error
		if full {
			result, err = source.Fetch(context.Background(), q)
		} else {
			result, err = source.FetchRecordPage(context.Background(), q)
		}
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(result.Question, before) {
			t.Fatal("caller changed the question during source I/O")
		}
		receipt, _ := json.Marshal(result)
		q.Where.All[0].Field.Key = "changed"
		q.Where.All[1].Any[0].Value = "east"
		q.EnvironmentSelector.Scope.Value = "changed"
		q.GroupBy[0].Key = "name"
		q.Where.All = append(q.Where.All, Predicate{Op: "invented"})
		after, _ := json.Marshal(result)
		if string(receipt) != string(after) {
			t.Fatal("returned receipt aliases caller query")
		}
	}
}
func TestInventoryAndDiscoveryDetachNamespaceBeforeIO312(t *testing.T) {
	for _, discovery := range []bool{false, true} {
		q := InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 2000, Namespace: &identity.Namespace{Mode: "value", Value: "payments"}}
		calls := 0
		source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: consistencyTransport312(func(r *http.Request) (*http.Response, error) {
			calls++
			q.Namespace.Value = "during-io"
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(serviceProbeResponse379(r)))}, nil
		})}}
		var returned InventoryQuestion
		if discovery {
			result, err := source.DiscoverServices(context.Background(), q)
			if err != nil {
				t.Fatal(err)
			}
			returned = result.Question
		} else {
			result, err := source.Inventory(context.Background(), q)
			if err != nil {
				t.Fatal(err)
			}
			returned = result.Question
		}
		q.Namespace.Value = "after-return"
		if returned.Namespace.Value != "payments" || calls != map[bool]int{true: 14, false: 1}[discovery] {
			t.Fatal("discovery receipt aliases caller namespace", returned, calls)
		}
	}
}

func TestSourceQueryInvalidBeforeAnySQL312(t *testing.T) {
	for _, full := range []bool{false, true} {
		for _, mutate := range []func(*Query){func(q *Query) { q.ToMS = q.FromMS }, func(q *Query) { q.Offset = 1001 }, func(q *Query) {
			q.Where = &Predicate{Field: &Field{Kind: "column", Key: "invented"}, Op: "eq", Value: "x"}
		}} {
			q := Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
			mutate(&q)
			source, calls := consistencySource312(0, 0, nil)
			var err error
			if full {
				_, err = source.Fetch(context.Background(), q)
			} else {
				_, err = source.FetchRecordPage(context.Background(), q)
			}
			if err == nil || *calls != 0 {
				t.Fatal("invalid query reached SQL", err, *calls)
			}
		}
	}
}

// Positive aggregate fixtures must honor the record query's lookahead/offset;
// otherwise the stricter adapter correctly treats their old partial rows as drift.
var fixturePage312 = regexp.MustCompile(` LIMIT ([0-9]+) OFFSET ([0-9]+) FORMAT JSONEachRow$`)

func writeRecordFixtureRows312(t *testing.T, w io.Writer, sql string, total int, row string) {
	t.Helper()
	parts := fixturePage312.FindStringSubmatch(sql)
	if len(parts) != 3 {
		t.Errorf("expected bounded record fixture SQL: %s", sql)
		return
	}
	limit, _ := strconv.Atoi(parts[1])
	offset, _ := strconv.Atoi(parts[2])
	for range min(limit, max(total-offset, 0)) {
		fmt.Fprintln(w, row)
	}
}

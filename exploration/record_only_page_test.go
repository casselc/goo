package exploration

import (
	"context"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRecordsOnlyInitialReadUsesExistingTwoStatementProfile302(t *testing.T) {
	for _, name := range []string{"ungrouped logs", "exact trace", "multiple groups", "latent time and duration"} {
		t.Run(name, func(t *testing.T) {
			q := basic()
			q.Offset = 0
			switch name {
			case "ungrouped logs":
				q.Signal = "logs"
				q.GroupBy = nil
			case "exact trace":
				q.GroupBy = nil
				q.Where = &Predicate{Field: &Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: strings.Repeat("a", 32)}
			case "multiple groups":
				q.GroupBy = []Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}
			case "latent time and duration":
				q.GroupBy = []Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}
				q.BucketMS = 10000
				q.Summary = "span-duration"
				q.SpanStatus = true
			}
			expected := q.WithEnvironmentFrom(q)
			recorder := &pageReadRecorder{count: "3"}
			source := Source{Endpoint: "http://recording.invalid", Client: &http.Client{Transport: recorder}}
			result, err := source.FetchRecordPage(context.Background(), q)
			if err != nil {
				t.Fatal(err)
			}
			if len(recorder.sql) != 2 || !strings.Contains(recorder.sql[0], "toString(count()) AS count") || !strings.Contains(recorder.sql[1], " AS timestamp_ns") {
				t.Fatalf("initial source read shape differs: %+v", recorder.sql)
			}
			for _, sql := range recorder.sql {
				if strings.Contains(sql, "GROUP BY") {
					t.Fatal("record-only profile acquired aggregate", sql)
				}
			}
			if !reflect.DeepEqual(result.Question, expected) {
				t.Fatal("profile rewrote grouping, time or measure directives", result.Question, expected)
			}
			if result.ViewDataset != nil || result.TimeDataset != nil || result.SummaryDataset != nil || len(result.Groups) != 0 || len(result.Fields) != 0 {
				t.Fatal("profile invented unacquired aspects")
			}
			if result.Count != "3" || len(result.Rows) != 2 || !result.HasMore {
				t.Fatal("bounded rows replaced full count", result.Count, len(result.Rows), result.HasMore)
			}
			t.Logf("initial records-only statements=%d (independent count + bounded rows)", len(recorder.sql))
		})
	}
}

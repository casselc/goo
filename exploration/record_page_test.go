package exploration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type pageReadRecorder struct {
	params []map[string]string
	sql    []string
	count  string
	fail   string
}

func (p *pageReadRecorder) RoundTrip(req *http.Request) (*http.Response, error) {
	b, _ := io.ReadAll(req.Body)
	sql := string(b)
	p.sql = append(p.sql, sql)
	parameters := map[string]string{}
	for key, values := range req.URL.Query() {
		if strings.HasPrefix(key, "param_") {
			parameters[strings.TrimPrefix(key, "param_")] = values[0]
		}
	}
	p.params = append(p.params, parameters)
	body := ""
	status := 200
	if strings.Contains(sql, "toString(count()) AS count") && !strings.Contains(sql, "GROUP BY") {
		body = `{"count":"` + p.count + `"}`
	}
	if strings.Contains(sql, " AS timestamp_ns") {
		body = `{"timestamp_ns":"1","duration_ns":"0","SpanName":"sample","Body":"sample","ResourceAttributes":{"deployment.environment.name":"local","deployment.environment":"local"}}` + "\n" + `{"timestamp_ns":"2","duration_ns":"0","SpanName":"sample","Body":"sample","ResourceAttributes":{"deployment.environment.name":"local","deployment.environment":"local"}}` + "\n" + `{"timestamp_ns":"3","duration_ns":"0","SpanName":"sample","Body":"sample","ResourceAttributes":{"deployment.environment.name":"local","deployment.environment":"local"}}`
	}
	if p.fail != "" && strings.Contains(sql, p.fail) {
		status = 503
	}
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
}

func TestRecordPagingSourceWorkMeasurement289(t *testing.T) {
	for _, tc := range []struct {
		name          string
		group, bucket bool
		reads         int
	}{{"records", false, false, 3}, {"grouped", true, false, 4}, {"grouped-timeline", true, true, 5}} {
		t.Run(tc.name, func(t *testing.T) {
			q := basic()
			if tc.group {
				q.GroupBy = []Field{{Kind: "column", Key: "service"}}
			}
			if tc.bucket {
				q.BucketMS = 10000
			}
			// Deterministic source output permits exact read-shape measurement with
			// no latency claim and without coupling fixtures to aggregate decoder details.
			recorder := &pageReadRecorder{count: "0"}
			source := Source{Endpoint: "http://recording.invalid", Client: &http.Client{Transport: recorder}}
			if _, err := source.Fetch(context.Background(), q); err != nil {
				t.Fatal(err)
			}
			if len(recorder.sql) != tc.reads {
				t.Fatalf("initial reads %d, want %d", len(recorder.sql), tc.reads)
			}
			first := append([]string(nil), recorder.sql...)
			recorder.sql = nil
			q.Offset = 2
			if _, err := source.Fetch(context.Background(), q); err != nil {
				t.Fatal(err)
			}
			if len(recorder.sql) != tc.reads {
				t.Fatalf("full paging reads %d, want %d", len(recorder.sql), tc.reads)
			}
			if first[0] != recorder.sql[0] {
				t.Fatal("offset changed full count SQL")
			}
			recorder.sql = nil
			page, err := source.FetchRecordPage(context.Background(), q)
			if err != nil {
				t.Fatal(err)
			}
			if len(recorder.sql) != 2 {
				t.Fatalf("page reads %d, want2", len(recorder.sql))
			}
			if len(page.Fields) != 0 || len(page.Groups) != 0 || page.TimeDataset != nil || page.SummaryDataset != nil || page.ViewDataset != nil {
				t.Fatal("page invented aggregate/discovery authority")
			}
			t.Logf("initial=%d full-page=%d page-profile=%d", tc.reads, tc.reads, len(recorder.sql))
		})
	}
}

func TestRecordPageRetainsQuestionAndSelectionCustody(t *testing.T) {
	q := basic()
	q.Offset = 2
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	q.BucketMS = 10000
	q.Summary = "span-duration"
	q.SpanStatus = true
	recorder := &pageReadRecorder{count: "5"}
	source := Source{Endpoint: "http://recording.invalid", Client: &http.Client{Transport: recorder}}
	result, err := source.FetchRecordPage(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Question, q.WithEnvironmentFrom(q)) || result.Count != "5" || len(result.Rows) != 2 || !result.HasMore {
		t.Fatalf("wrong original question or bounded rows: %+v", result)
	}
	for _, sql := range recorder.sql {
		if strings.Contains(sql, "quantile") || strings.Contains(sql, "GROUP BY") || strings.Contains(sql, "SELECT DISTINCT") || strings.Contains(sql, "sum(") {
			t.Fatalf("aggregate work leaked: %s", sql)
		}
	}
	bank := &ResultSnapshots{}
	token, err := bank.Freeze(result)
	if err != nil {
		t.Fatal(err)
	}
	result.ResultToken = token
	held, err := bank.Load(token)
	if err != nil {
		t.Fatal(err)
	}
	if held.Explore.Basis != result.Basis || !reflect.DeepEqual(held.Explore.Rows, result.Rows) || held.Explore.Question.Offset != 2 {
		t.Fatal("page token lost exact record custody")
	}
	kept, err := (KeptObservationStore{Dir: t.TempDir()}).KeepRecords(context.Background(), bank, CaptureStore{}, KeepRecordsIntent{Version: KeepRecordsVersion, RequestID: strings.Repeat("d", 64), ResultToken: token, Indexes: []int{1}})
	if err != nil || kept.Question.Offset != 2 || len(kept.Members) != 1 || string(kept.Members[0].Record) != string(result.Rows[1]) {
		t.Fatal("page selection custody failed", err)
	}
	raw, err := json.Marshal(RecordPageResponse(result))
	if err != nil {
		t.Fatal(err)
	}
	var slots map[string]json.RawMessage
	json.Unmarshal(raw, &slots)
	if len(slots) != 9 {
		t.Fatalf("unexpected wire slots: %s", raw)
	}
	for _, name := range []string{"plan", "groups", "fields", "duration", "viewDataset", "timeDataset", "summaryDataset", "traceExamples"} {
		if _, ok := slots[name]; ok {
			t.Fatalf("page exposes %s", name)
		}
	}
}

func TestRecordPageValidationAndFailureNeverBecomeHealthy(t *testing.T) {
	recorder := &pageReadRecorder{count: "0"}
	source := Source{Endpoint: "http://recording.invalid", Client: &http.Client{Transport: recorder}}
	q := basic()
	q.Summary = "invalid"
	if r, err := source.FetchRecordPage(context.Background(), q); err == nil || r != nil || len(recorder.sql) != 0 {
		t.Fatal("stripping allowed invalid original question")
	}
	recorder.fail = " AS timestamp_ns"
	if r, err := source.FetchRecordPage(context.Background(), basic()); err == nil || r != nil {
		t.Fatal("partial source failure became valid page")
	}
}

func TestRecordPagePreservesOriginalDerivedProjection289(t *testing.T) {
	for _, tc := range []struct{ signal, key string }{{"spans", "exception_type"}, {"logs", "body_pattern_v1"}, {"logs", "severity_number"}} {
		t.Run(tc.key, func(t *testing.T) {
			q := basic()
			q.Signal = tc.signal
			q.GroupBy = []Field{{Kind: "column", Key: tc.key}}
			plan, err := Compile(q)
			if err != nil {
				t.Fatal(err)
			}
			recorder := &pageReadRecorder{count: "0"}
			source := Source{Endpoint: "http://recording.invalid", Client: &http.Client{Transport: recorder}}
			if _, err = source.FetchRecordPage(context.Background(), q); err != nil {
				t.Fatal(err)
			}
			if len(recorder.sql) != 2 || recorder.sql[1] != plan.Rows {
				t.Fatal("original derived row projection changed")
			}
			wantParams := map[string]string{}
			for key, value := range plan.Parameters {
				wantParams[key] = encodeParameter(value)
			}
			if !reflect.DeepEqual(recorder.params[1], wantParams) {
				t.Fatal("original row query parameters changed", recorder.params[1], wantParams)
			}

		})
	}
}

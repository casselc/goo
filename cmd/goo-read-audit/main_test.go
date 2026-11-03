package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestSourceSummaryCannotPublishSourceContent(t *testing.T) {
	got := sourceSummary(`{"read_rows":"123","read_bytes":"18446744073709551615","elapsed_ns":"0","query":"secret record","result_bytes":"record contents","written_rows":"01","written_bytes":"18446744073709551616","result_rows":-1,"total_rows_to_read":123}`)
	want := map[string]string{"read_rows": "123", "read_bytes": "18446744073709551615", "elapsed_ns": "0"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unsafe summary: %#v", got)
	}
	for _, raw := range []string{`null`, `"source content"`, `[]`, `{"read_rows":{}}`, `invalid`} {
		if len(sourceSummary(raw)) != 0 {
			t.Fatalf("accepted malformed summary %s", raw)
		}
	}
}

type auditFixtureTransport func(*http.Request) (*http.Response, error)

func (f auditFixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFixedCountOperandAggregateAuditKeepsMeaningAndReducesReadSet314(t *testing.T) {
	q := exploration.Query{Signal: "spans", Environment: "audit-fixture", FromMS: 1000, ToMS: 61000, Limit: 1, Summary: "span-duration", GroupBy: []exploration.Field{{Kind: "attribute", Key: "http.route"}}, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "http.request.method"}, Op: "ne", Value: "POST"}}
	plan, err := exploration.Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	makeAudit := func(rowCount int, groupDrift bool) *audit {
		return &audit{kinds: map[string]string{plan.Count: "count", plan.Rows: "rows", plan.Groups: "groups", plan.Fields: "fields"}, reads: []*read{}, base: auditFixtureTransport(func(request *http.Request) (*http.Response, error) {
			raw, _ := io.ReadAll(request.Body)
			body := ""
			measures := `"count":"4","sum_ns":"40","min_ns":"10","max_ns":"10","quantiles_ns":["10","10","10"]`
			switch string(raw) {
			case plan.Count:
				body = "{" + measures + "}"
			case plan.Groups:
				body = `{"g0":"/synthetic",` + measures + "}"
				if groupDrift {
					body = strings.Replace(body, `"count":"4","sum_ns":"40"`, `"count":"3","sum_ns":"30"`, 1)
				}
			case plan.Rows:
				body = strings.Repeat(`{"timestamp_ns":"1000000001","duration_ns":"10"}`+"\n", rowCount)
			case plan.Fields:
				body = `{"field":"attribute:http.route"}`
			default:
				t.Error("unexpected source statement")
			}
			return &http.Response{StatusCode: 200, Header: http.Header{"X-Clickhouse-Summary": []string{`{"read_rows":"4","read_bytes":"128"}`}}, Body: io.NopCloser(strings.NewReader(body))}, nil
		})}
	}
	fullAudit, narrowAudit := makeAudit(2, false), makeAudit(2, false)
	full, err := (exploration.Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: fullAudit}}).Fetch(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	narrow, err := (exploration.Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: narrowAudit}}).FetchRecordAggregates(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	if len(fullAudit.reads) != 4 || len(narrowAudit.reads) != 2 || narrowAudit.reads[0].Kind != "count" || narrowAudit.reads[1].Kind != "groups" {
		t.Fatal("operand source statement budget changed")
	}
	a := aggregateFingerprint(full.Question, full.Count, full.State, full.Groups, full.Duration, full.SpanStatus, full.TimeSeries)
	b := aggregateFingerprint(narrow.Question, narrow.Count, narrow.State, narrow.Groups, narrow.Duration, narrow.SpanStatus, narrow.TimeSeries)
	if a != b || full.State != "observed" || narrow.Count != "4" || narrow.Duration == nil {
		t.Fatal("aggregate-only acquisition changed admitted meaning")
	}
	if narrow.Reads.Count.ReadAt == "" || narrow.Reads.Groups.ReadAt == "" || narrow.Reads.Count.Basis == "" || narrow.Reads.Groups.Basis == "" {
		t.Fatal("narrow read lost independent receipts")
	}
	if fullAudit.reads[1].ResponseBytes == 0 || fullAudit.reads[3].ResponseBytes == 0 {
		t.Fatal("audit missed the avoidable row/field payload")
	}
	// Fingerprints must fail equality rather than hiding source diagnostic state
	// or measures. They deliberately exclude separate read clocks/DTO shape.
	changed := aggregateFingerprint(narrow.Question, narrow.Count, "changed_between_reads", narrow.Groups, narrow.Duration, narrow.SpanStatus, narrow.TimeSeries)
	if changed == b {
		t.Fatal("fingerprint discarded certainty")
	}
	// Row contradictions are observable only to the full reader. A narrow
	// result can be coherent for its own aggregates, but must not masquerade as
	// an inspection result or acquire a full snapshot/witness by adaptation.
	shortAudit := makeAudit(0, false)
	short, err := (exploration.Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: shortAudit}}).Fetch(context.Background(), q)
	if err != nil || short.State != "changed_between_reads" || short.SummaryDataset.Completeness != nil {
		t.Fatal("full row contradiction lost", err)
	}
	if narrow.State != "observed" {
		t.Fatal("aggregate certainty depends on unacquired rows")
	}
	narrowWire, _ := json.Marshal(narrow)
	var fields map[string]json.RawMessage
	json.Unmarshal(narrowWire, &fields)
	for _, key := range []string{"rows", "viewDataset", "summaryDataset", "timeDataset", "resultToken"} {
		if _, ok := fields[key]; ok {
			t.Fatal("aggregate result fabricated inspection authority", key)
		}
	}
	driftAudit := makeAudit(2, true)
	drift, err := (exploration.Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: driftAudit}}).FetchRecordAggregates(context.Background(), q)
	if err != nil || drift.State != "changed_between_reads" || len(driftAudit.reads) != 2 {
		t.Fatal("narrow aggregate disagreement hidden", err)
	}
	raw, err := json.Marshal(narrowAudit.reads)
	if err != nil {
		t.Fatal(err)
	}
	for _, private := range []string{"/synthetic", "POST", "fixture.invalid", "SELECT ", "param_"} {
		if bytes.Contains(raw, []byte(private)) {
			t.Fatal("audit exported source content")
		}
	}
}

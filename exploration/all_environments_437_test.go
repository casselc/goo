package exploration

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/identity"
	"gopkg.in/yaml.v3"
)

func allQuestion437() Query {
	wire, _ := identity.AllEnvironments().Selector()
	return Query{Signal: "logs", EnvironmentSelector: &wire, FromMS: 100000, ToMS: 160000, Limit: 100, Where: &Predicate{Field: &Field{Kind: "column", Key: "body"}, Op: "contains", Value: "test' OR 1=1 --"}}
}

func TestAllEnvironments437TelemetryPackageRoundTrip(t *testing.T) {
	q := allQuestion437()
	p, err := DraftTelemetryPackage(q, "all-logs", "All environment logs", 5)
	if err != nil {
		t.Fatal(err)
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := ReadTelemetryPackage(string(data))
	if err != nil || !reflect.DeepEqual(loaded.Question(q.FromMS, q.ToMS), q) {
		t.Fatal("package YAML lost the complete All log question", err)
	}
	loaded.Source.Signal = "spans"
	if loaded.Validate() == nil {
		t.Fatal("package validation bypassed the logs-only gate")
	}
}

func TestAllEnvironments437CompilePreservesBoundsFiltersAndBudgets(t *testing.T) {
	q := allQuestion437()
	p, err := Compile(q)
	if err != nil {
		t.Fatal(err)
	}
	if p.Parameters["from"] != "100000" || p.Parameters["to"] != "160000" || strings.Contains(p.Rows, q.Where.Value) || !strings.Contains(p.Rows, " AND 1 AND ") || !strings.Contains(p.Rows, "LIMIT 101 OFFSET 0") {
		t.Fatal("All changed the source budget or filter", p)
	}
	if _, exists := p.Parameters["environment"]; exists {
		t.Fatal("All retained a named environment binding")
	}
	found := false
	for _, value := range p.Parameters {
		found = found || value == q.Where.Value
	}
	if !found {
		t.Fatal("All lost the ordinary predicate binding")
	}
	for _, changed := range []Query{func() Query { v := q; v.Signal = "spans"; return v }(), func() Query { v := q; v.Environment = "prod"; return v }(), func() Query { v := q; v.ToMS = v.FromMS; return v }()} {
		if _, err := Compile(changed); err == nil {
			t.Fatal("unsupported or malformed All question compiled", changed)
		}
	}
	if _, _, err := (&compiler{}).source(Query{Signal: "spans"}, identity.AllEnvironments()); err == nil {
		t.Fatal("direct source lowering bypassed the logs-only gate")
	}
	if _, _, err := (&compiler{params: map[string]string{"environment": "prod"}}).source(q, identity.AllEnvironments()); err == nil {
		t.Fatal("All inherited a stale named binding")
	}
	if _, err := SurroundingQuestion(q, "130000000000", "api", 1000, 0); err == nil {
		t.Fatal("All gained unsupported row context")
	}
	if _, err := (Query{Signal: "logs"}).EnvironmentSelection(); err == nil {
		t.Fatal("blank environment became All")
	}
}

func TestAllEnvironments437DoesNotBypassRecordAdmission(t *testing.T) {
	q := allQuestion437()
	for _, record := range []json.RawMessage{
		json.RawMessage(`{"timestamp_ns":"130000000000","Body":"missing resource map"}`),
		json.RawMessage(`{"timestamp_ns":"130000000000","Body":"null map","ResourceAttributes":null}`),
		json.RawMessage(`{"timestamp_ns":"130000000000","Body":"null attribute","ResourceAttributes":{"deployment.environment.name":null}}`),
		json.RawMessage(`{"timestamp_ns":"130000000000","SpanName":"wrong family","ResourceAttributes":{}}`),
		json.RawMessage(`{"timestamp_ns":"160000000000","Body":"exclusive end","ResourceAttributes":{}}`),
		json.RawMessage(`{"timestamp_ns":"0130000000000","Body":"noncanonical time","ResourceAttributes":{}}`),
	} {
		if err := validateSelectedRows(q, []json.RawMessage{record}); err == nil {
			t.Fatal("All erased a non-environment source guard", string(record))
		}
	}
}

func TestAllEnvironments437ReadsAndReopensMixedLabels(t *testing.T) {
	resources := []map[string]string{{identity.EnvironmentKey: "prod"}, {identity.LegacyEnvironmentKey: "legacy"}, {identity.EnvironmentKey: "", identity.LegacyEnvironmentKey: "prod"}, {}}
	rows := make([]json.RawMessage, len(resources))
	for i, resource := range resources {
		rows[i], _ = json.Marshal(map[string]any{"timestamp_ns": fmt.Sprint(130000000000 + i), "Body": "test' OR 1=1 --", "ServiceName": "api", "TraceId": strings.Repeat("a", 32), "ResourceAttributes": resource})
	}
	var reads atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		body, _ := io.ReadAll(r.Body)
		sql := string(body)
		if r.URL.Query().Has("param_environment") || !strings.Contains(sql, "Timestamp >= fromUnixTimestamp64Milli({from:Int64})") || !strings.Contains(sql, "Timestamp < fromUnixTimestamp64Milli({to:Int64})") {
			t.Error("All broadened beyond the fixed window or introduced a dummy binding", sql)
		}
		switch {
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "count()"):
			fmt.Fprintln(w, `{"count":"4"}`)
		default:
			for _, row := range rows {
				fmt.Fprintln(w, string(row))
			}
		}
	}))
	defer server.Close()
	source := Source{Endpoint: server.URL, Client: server.Client()}
	q := allQuestion437()
	tail, limited, err := source.PollTail(t.Context(), TailSelector{EnvironmentSelector: q.EnvironmentSelector}, 160000)
	if err != nil || limited || len(tail) != 4 {
		t.Fatal("tail did not retain every observed label category", len(tail), limited, err)
	}
	before := reads.Load()
	if _, _, err := source.PollTail(t.Context(), TailSelector{Signal: "spans", EnvironmentSelector: q.EnvironmentSelector}, 160000); err == nil || reads.Load() != before {
		t.Fatal("unsupported All tail read the source", err)
	}
	store := CaptureStore{Dir: t.TempDir()}
	capture, err := store.Create(t.Context(), source, "Across reported labels", CaptureQuestion{Kind: "explore", Explore: &q}, "")
	if err != nil || capture.Explore.Count != "4" || len(capture.Explore.Rows) != 4 || !capture.Question.Explore.SameEnvironment(q) {
		t.Fatal("capture lost the complete All question", capture, err)
	}
	server.Close()
	before = reads.Load()
	loaded, err := store.Load(capture.ID)
	if err != nil {
		t.Fatal(err)
	}
	page, err := loaded.Page("")
	if err != nil || page.Receipt.Question.Explore.EnvironmentSelector == nil || page.Receipt.Question.Explore.EnvironmentSelector.Scope.Mode != "all" || len(page.Explore.Rows) != 4 {
		t.Fatal("stored page changed All", page, err)
	}
	followups, err := loaded.Followups(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range followups.Actions {
		if action.ID == "trace" || action.ID == "trace-logs" || action.ID == "span-logs" || action.Discovery != nil {
			t.Fatal("All offered unsupported row context", action)
		}
	}
	if reads.Load() != before {
		t.Fatal("stored reopening read telemetry")
	}
}

package exploration

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func exampleRow(trace, span, duration, status string) json.RawMessage {
	raw, _ := json.Marshal(map[string]string{"TraceId": trace, "SpanId": span, "ServiceName": "api", "SpanName": "GET /generic", "timestamp_ns": "1500000000", "duration_ns": duration, "StatusCode": status})
	return raw
}
func TestTraceExamplesExactPagePolicy(t *testing.T) {
	q := Query{Signal: "spans", Environment: "test", FromMS: 1000, ToMS: 2000, Limit: 100}
	a, b, c := "11111111111111111111111111111111", "22222222222222222222222222222222", "33333333333333333333333333333333"
	r := Result{Question: q, State: "observed", Count: "9000", HasMore: true, Rows: []json.RawMessage{exampleRow(a, "1111111111111111", "9007199254740993", "Error"), exampleRow(b, "2222222222222222", "9007199254740994", "Ok"), exampleRow(c, "3333333333333333", "0", "Unset"), exampleRow("00000000000000000000000000000000", "4444444444444444", "1", "Error")}}
	result := SelectTraceExamples(r)
	if result.State != "observed" || result.Considered != 4 || result.Excluded != 1 || len(result.Items) != 3 || result.Items[0].TraceID != a || result.Items[1].TraceID != b || result.Items[2].DurationNS != "0" {
		t.Fatal(result)
	}
	r.Rows[0], r.Rows[2] = r.Rows[2], r.Rows[0]
	if !reflect.DeepEqual(result, SelectTraceExamples(r)) {
		t.Fatal("page order changes examples")
	}
	r.State = "changed_between_reads"
	if len(SelectTraceExamples(r).Items) != 0 {
		t.Fatal("inconsistent reads offered examples")
	}
	r.State = "observed"
	r.Rows = append(r.Rows, exampleRow(a, "5555555555555555", "18446744073709551615", "Error"))
	if SelectTraceExamples(r).Items[0].DurationNS != "18446744073709551615" {
		t.Fatal("uint64 precision lost")
	}
	r.Question.Signal = "logs"
	if SelectTraceExamples(r) != nil {
		t.Fatal("log examples inferred")
	}
}
func TestTraceExamplesDeduplicateAndRejectInvalidAnchor(t *testing.T) {
	q := Query{Signal: "spans", Environment: "test", FromMS: 1000, ToMS: 2000, Limit: 100}
	r := Result{Question: q, State: "observed", Rows: []json.RawMessage{exampleRow("11111111111111111111111111111111", "1111111111111111", "1", "Error"), exampleRow("11111111111111111111111111111111", "2222222222222222", "2", "Ok")}}
	if len(SelectTraceExamples(r).Items) != 1 {
		t.Fatal("duplicate trace")
	}
	r.Rows = []json.RawMessage{json.RawMessage(`{"TraceId":"11111111111111111111111111111111","SpanId":"1111111111111111","timestamp_ns":"2000000000","duration_ns":"1"}`)}
	if SelectTraceExamples(r).State != "no_linked_spans" {
		t.Fatal("outside window anchor")
	}
}

func TestStoredTraceExamplesFollowCapturePage(t *testing.T) {
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
	rows := []json.RawMessage{}
	for i := 0; i < 201; i++ {
		raw, _ := json.Marshal(map[string]string{"TraceId": fmt.Sprintf("%032x", i+1), "SpanId": fmt.Sprintf("%016x", i+1), "duration_ns": strconv.Itoa(i), "timestamp_ns": "1500000000", "StatusCode": "Ok"})
		rows = append(rows, raw)
	}
	r := Result{Question: q, State: "observed", Rows: rows}
	r.TraceExamples = SelectTraceExamples(r)
	c := SavedCapture{CaptureReceipt: CaptureReceipt{ID: strings.Repeat("a", 64), Question: CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &r}
	first, e := c.Page("")
	if e != nil {
		t.Fatal(e)
	}
	second, e := c.Page(first.NextCursor)
	if e != nil {
		t.Fatal(e)
	}
	if first.Explore.TraceExamples.Considered != 100 || second.Explore.TraceExamples.Considered != 100 || first.Explore.TraceExamples.Items[0].DurationNS != "99" || second.Explore.TraceExamples.Items[0].DurationNS != "199" {
		t.Fatal("examples escaped frozen page")
	}
	if r.TraceExamples.Considered != 201 {
		t.Fatal("paging mutated stored evidence")
	}
	r.TraceExamples = nil
	old, e := c.Page("")
	if e != nil || old.Explore.TraceExamples != nil {
		t.Fatal("legacy capture synthesized new metadata")
	}
}

func TestTraceExampleTieFieldsDoNotCollide(t *testing.T) {
	row := func(service, name string) json.RawMessage {
		b, _ := json.Marshal(map[string]string{"TraceId": strings.Repeat("1", 32), "SpanId": strings.Repeat("2", 16), "timestamp_ns": "1500000000", "duration_ns": "1", "ServiceName": service, "SpanName": name})
		return b
	}
	q := Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
	a, b := row("a/b", "c"), row("a", "b/c")
	one := SelectTraceExamples(Result{Question: q, State: "observed", Rows: []json.RawMessage{a, b}})
	two := SelectTraceExamples(Result{Question: q, State: "observed", Rows: []json.RawMessage{b, a}})
	if !reflect.DeepEqual(one, two) {
		t.Fatal("field separators changed deterministic tie")
	}
}

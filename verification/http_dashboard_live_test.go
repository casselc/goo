package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
)

// This opt-in check derives populations from ordinary OTLP fixture facts, not
// the planner/evaluator. Browser ownership and effect assertions remain in 187.
type hdLiveRecord struct {
	ID, TraceID, Environment, Service, Namespace, Name, DurationNS string
	Kind, Status                                                   int
	TimeMS                                                         int64
	Attributes                                                     map[string]string
}
type hdLiveLedger struct {
	Records   []hdLiveRecord
	Responses []struct {
		Endpoint string
		Status   int
		Body     json.RawMessage
	}
}

func hdLivePopulation(records []hdLiveRecord, q exploration.Query) []hdLiveRecord {
	var out []hdLiveRecord
	for _, r := range records {
		fields := map[exploration.Field]string{
			{Kind: "column", Key: "trace_id"}:            r.TraceID,
			{Kind: "column", Key: "span_id"}:             r.ID,
			{Kind: "column", Key: "service"}:             r.Service,
			{Kind: "column", Key: "name"}:                r.Name,
			{Kind: "column", Key: "kind"}:                map[int]string{1: "Internal", 2: "Server", 3: "Client", 4: "Producer", 5: "Consumer"}[r.Kind],
			{Kind: "resource", Key: "service.namespace"}: r.Namespace,
		}
		for k, v := range r.Attributes {
			fields[exploration.Field{Kind: "attribute", Key: k}] = v
		}
		if r.Environment == q.Environment && r.TimeMS >= q.FromMS && r.TimeMS < q.ToMS && hdPredicate(q.Where, fields) {
			out = append(out, r)
		}
	}
	return out
}
func hdLiveResult(records []hdLiveRecord, r exploration.Result) error {
	pop := hdLivePopulation(records, r.Question)
	if r.Count != strconv.Itoa(len(pop)) {
		return fmt.Errorf("population count %s != %d", r.Count, len(pop))
	}
	byID := map[string]hdLiveRecord{}
	sum := new(big.Int)
	errors := 0
	expectedGroups := map[string]int{}
	for _, row := range pop {
		byID[row.ID] = row
		n, ok := new(big.Int).SetString(row.DurationNS, 10)
		if !ok {
			return fmt.Errorf("fixture duration")
		}
		sum.Add(sum, n)
		if row.Status == 2 {
			errors++
		}
		group := map[string]any{}
		for i, f := range r.Question.GroupBy {
			var v any
			switch f.Kind {
			case "column":
				switch f.Key {
				case "name":
					v = row.Name
				case "service":
					v = row.Service
				default:
					return fmt.Errorf("unsupported group %v", f)
				}
			case "attribute":
				if s, ok := row.Attributes[f.Key]; ok {
					v = s
				}
			default:
				return fmt.Errorf("unsupported group %v", f)
			}
			group[fmt.Sprintf("g%d", i)] = v
		}
		key, _ := json.Marshal(group)
		expectedGroups[string(key)]++
	}
	if r.Duration != nil && (r.Duration.Count != r.Count || r.Duration.SumNS != sum.String()) {
		return fmt.Errorf("duration aggregate")
	}
	if r.SpanStatus != nil && (r.SpanStatus.Count != r.Count || r.SpanStatus.Error != strconv.Itoa(errors) || r.SpanStatus.OK != strconv.Itoa(len(pop)-errors)) {
		return fmt.Errorf("reported status aggregate")
	}
	if len(r.Question.GroupBy) > 0 {
		got := map[string]int{}
		for _, raw := range r.Groups {
			var g map[string]json.RawMessage
			if json.Unmarshal(raw, &g) != nil {
				return fmt.Errorf("group JSON")
			}
			var count string
			if json.Unmarshal(g["count"], &count) != nil {
				return fmt.Errorf("group count")
			}
			n, e := strconv.Atoi(count)
			if e != nil {
				return e
			}
			keys := map[string]any{}
			for i := range r.Question.GroupBy {
				k := fmt.Sprintf("g%d", i)
				var v any
				if json.Unmarshal(g[k], &v) != nil {
					return fmt.Errorf("group value")
				}
				keys[k] = v
			}
			key, _ := json.Marshal(keys)
			if _, ok := got[string(key)]; ok {
				return fmt.Errorf("duplicate group")
			}
			got[string(key)] = n
		}
		if !reflect.DeepEqual(got, expectedGroups) {
			return fmt.Errorf("full group population")
		}
	}
	want := len(pop) - r.Question.Offset
	if want < 0 {
		want = 0
	}
	if want > r.Question.Limit {
		want = r.Question.Limit
	}
	if len(r.Rows) != want {
		return fmt.Errorf("page length")
	}
	seen := map[string]bool{}
	for _, raw := range r.Rows {
		var row struct {
			SpanID     string            `json:"SpanId"`
			TraceID    string            `json:"TraceId"`
			Duration   string            `json:"duration_ns"`
			Name       string            `json:"SpanName"`
			Attributes map[string]string `json:"SpanAttributes"`
			Resource   map[string]string `json:"ResourceAttributes"`
			Service    string            `json:"ServiceName"`
			Timestamp  string            `json:"timestamp_ns"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return fmt.Errorf("row JSON")
		}
		fact, ok := byID[row.SpanID]
		if !ok || seen[row.SpanID] {
			return fmt.Errorf("row membership/unique IDs")
		}
		seen[row.SpanID] = true
		if row.Service != fact.Service || row.Timestamp != new(big.Int).Mul(big.NewInt(fact.TimeMS), big.NewInt(1000000)).String() || row.Resource["service.namespace"] != fact.Namespace || row.Resource["deployment.environment.name"] != fact.Environment {
			return fmt.Errorf("raw row scope/time changed")
		}
		if row.TraceID != fact.TraceID || row.Duration != fact.DurationNS || row.Name != fact.Name || !reflect.DeepEqual(row.Attributes, fact.Attributes) {
			return fmt.Errorf("raw row changed")
		}
	}
	return nil
}
func TestHTTPDashboardObservedPopulations(t *testing.T) {
	path := os.Getenv("GOO_HTTP_DASHBOARD_LEDGER")
	if path == "" {
		t.Skip("set GOO_HTTP_DASHBOARD_LEDGER to an actual 187 ledger")
	}
	hdCheckObservedLedger(t, path, 225)
}

func hdCheckObservedLedger(t *testing.T, path string, wantRecords int) {
	t.Helper()
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var ledger hdLiveLedger
	if e = json.Unmarshal(raw, &ledger); e != nil {
		t.Fatal(e)
	}
	if len(ledger.Records) != wantRecords {
		t.Fatal("ordinary fixture population missing")
	}
	checked := 0
	for _, response := range ledger.Responses {
		if response.Endpoint != "/api/explore" || response.Status != 200 {
			continue
		}
		var r exploration.Result
		if json.Unmarshal(response.Body, &r) != nil || r.Question.Signal != "spans" || r.Count == "" {
			continue
		}
		// Metadata discovery can use a different projection. Verify only executed
		// duration/status overview questions and their exact focused continuations.
		if r.Duration == nil && r.SpanStatus == nil {
			continue
		}
		if e := hdLiveResult(ledger.Records, r); e != nil {
			t.Fatalf("source response %d: %v", checked, e)
		}
		// A guaranteed changed count checks that the independent guard is live.
		changed := r
		changed.Count = new(big.Int).Add(func() *big.Int { n, _ := new(big.Int).SetString(r.Count, 10); return n }(), big.NewInt(1)).String()
		if hdLiveResult(ledger.Records, changed) == nil {
			t.Fatal("changed count accepted")
		}
		checked++
	}
	if checked < 6 {
		t.Fatalf("only %d full-population responses", checked)
	}
	t.Logf("independently checked %d actual full population/page/group/status/duration responses; count mutations rejected", checked)
}

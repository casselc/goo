package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type nrcFact struct {
	ID, TraceID, Environment, Service, Instance, TimestampNS, DurationNS, Name string
	Namespace                                                                  *string
	Attributes                                                                 map[string]string
}
type nrcFacts struct{ Records []nrcFact }
type nrcNative struct {
	Question             measurements.Question
	Basis, ReadAt        string
	Counters, Histograms *struct {
		Series []struct{ Stream metrics.Stream }
	}
}
type nrcView struct {
	ParentEvidence string
	Bounds         struct {
		FromNS string `json:"fromNs"`
		ToNS   string `json:"toNs"`
	}
	Timeline json.RawMessage
	Selected json.RawMessage
	Text     string
	Open     bool
}
type nrcEvent struct {
	Action                   string
	Source                   int
	Visible                  bool
	CancelledView, AfterLate json.RawMessage
}
type nrcLedger struct {
	Facts                                  nrcFacts
	Native                                 nrcNative
	Before, After                          nrcView
	Staged                                 struct{ Question exploration.Query }
	Records, Logs, Trace, LinkedLogs, Held exploration.Result
	TraceOrigin                            struct {
		Origin      exploration.Query
		OriginBasis string
	}
	Events []nrcEvent
	Errors []json.RawMessage
	Cases  []struct {
		Name                                 string
		Native                               nrcNative
		View                                 nrcView
		SelectedInterval                     struct{ FromNS, ToNS string }
		Staged                               struct{ Question exploration.Query }
		Result                               exploration.Result
		SourceBeforeStage, SourceAfterReturn int
	}
	Requests []struct {
		Endpoint, Method, Effect string
		Payload                  json.RawMessage
	}
}

func nrcExpected(q exploration.Query, stream metrics.Stream, from, to string) bool {
	a, ok := new(big.Int).SetString(from, 10)
	if !ok {
		return false
	}
	b, ok := new(big.Int).SetString(to, 10)
	if !ok || a.Cmp(b) >= 0 {
		return false
	}
	div := big.NewInt(1000000)
	lo := new(big.Int).Div(new(big.Int).Set(a), div)
	hi, rem := new(big.Int).QuoRem(new(big.Int).Set(b), div, new(big.Int))
	if rem.Sign() != 0 {
		hi.Add(hi, big.NewInt(1))
	}
	ns, hasNS := stream.Resource["service.namespace"]
	namespace := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "missing"}
	if hasNS {
		namespace.Op = "eq"
		namespace.Value = ns
	}
	want := exploration.Query{Signal: q.Signal, Environment: stream.Resource["deployment.environment.name"], FromMS: lo.Int64(), ToMS: hi.Int64(), Limit: 100, Where: &exploration.Predicate{All: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: stream.Resource["service.name"]}, namespace,
		{Field: &exploration.Field{Kind: "resource", Key: "service.instance.id"}, Op: "eq", Value: stream.Resource["service.instance.id"]},
		{Field: &exploration.Field{Kind: "column", Key: "timestamp_ns"}, Op: "gte", Value: from},
		{Field: &exploration.Field{Kind: "column", Key: "timestamp_ns"}, Op: "lte", Value: new(big.Int).Sub(b, big.NewInt(1)).String()},
	}}}
	return (q.Signal == "logs" || q.Signal == "spans") && nhLiveSameQuery(q, want)
}
func nrcStream(n nrcNative, views ...nrcView) (metrics.Stream, bool) {
	var streams []metrics.Stream
	if n.Counters != nil {
		for _, s := range n.Counters.Series {
			streams = append(streams, s.Stream)
		}
	}
	if n.Histograms != nil {
		for _, s := range n.Histograms.Series {
			streams = append(streams, s.Stream)
		}
	}
	if len(streams) == 1 {
		return streams[0], true
	}
	if len(views) != 1 {
		return metrics.Stream{}, false
	}
	// Read the actually selected stream's retained Raw record disclosure, rather
	// than assuming discovery returned a single stream or deriving it from child filters.
	text := views[0].Text
	i := strings.Index(text, "Raw record{")
	if i < 0 {
		return metrics.Stream{}, false
	}
	var selected struct {
		Name, Unit, Scope, Version, Schema, ResourceSchema string
		Resource                                           map[string]string `json:"ResourceAttributes"`
		Attributes                                         map[string]string
	}
	if json.NewDecoder(strings.NewReader(text[i+len("Raw record"):])).Decode(&selected) != nil {
		return metrics.Stream{}, false
	}
	var found metrics.Stream
	count := 0
	for _, s := range streams {
		if s.Name == selected.Name && s.Unit == selected.Unit && s.Scope == selected.Scope && s.Version == selected.Version && s.SchemaURL == selected.Schema && s.ResourceSchemaURL == selected.ResourceSchema && reflect.DeepEqual(s.Resource, selected.Resource) && reflect.DeepEqual(s.Attributes, selected.Attributes) {
			found = s
			count++
		}
	}
	return found, count == 1
}
func nrcFactsOK(facts nrcFacts, result exploration.Result) bool {
	byID := map[string]nrcFact{}
	for _, f := range facts.Records {
		fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: f.Service, {Kind: "column", Key: "trace_id"}: f.TraceID, {Kind: "column", Key: "span_id"}: f.ID, {Kind: "column", Key: "timestamp_ns"}: f.TimestampNS, {Kind: "resource", Key: "service.instance.id"}: f.Instance}
		if f.Namespace != nil {
			fields[exploration.Field{Kind: "resource", Key: "service.namespace"}] = *f.Namespace
		}
		ns, ok := new(big.Int).SetString(f.TimestampNS, 10)
		if !ok {
			return false
		}
		lo := new(big.Int).Mul(big.NewInt(result.Question.FromMS), big.NewInt(1000000))
		hi := new(big.Int).Mul(big.NewInt(result.Question.ToMS), big.NewInt(1000000))
		if f.Environment == result.Question.Environment && ns.Cmp(lo) >= 0 && ns.Cmp(hi) < 0 && nrcPredicate(result.Question.Where, fields) {
			byID[f.ID] = f
		}
	}
	if result.Count != strconv.Itoa(len(byID)) {
		return false
	}
	want := len(byID) - result.Question.Offset
	if want < 0 {
		want = 0
	}
	if want > result.Question.Limit {
		want = result.Question.Limit
	}
	if len(result.Rows) != want {
		return false
	}
	seen := map[string]bool{}
	for _, raw := range result.Rows {
		var row struct {
			ID          string            `json:"SpanId"`
			TraceID     string            `json:"TraceId"`
			TimestampNS string            `json:"timestamp_ns"`
			Service     string            `json:"ServiceName"`
			Resource    map[string]string `json:"ResourceAttributes"`
			Attributes  map[string]string `json:"Attributes"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		f, ok := byID[row.ID]
		if !ok || seen[row.ID] {
			return false
		}
		seen[row.ID] = true
		if row.TraceID != f.TraceID || row.TimestampNS != f.TimestampNS || row.Service != f.Service || row.Resource["service.instance.id"] != f.Instance || !reflect.DeepEqual(row.Attributes, f.Attributes) {
			return false
		}
		ns, present := row.Resource["service.namespace"]
		if present != (f.Namespace != nil) || present && ns != *f.Namespace {
			return false
		}
	}
	return true
}
func nrcMainOK(l nrcLedger) bool {
	stream, ok := nrcStream(l.Native, l.Before)
	if !ok {
		return false
	}
	if !nrcExpected(l.Staged.Question, stream, l.Before.Bounds.FromNS, l.Before.Bounds.ToNS) || !nhLiveSameQuery(l.Staged.Question, l.Records.Question) || !nrcFactsOK(l.Facts, l.Records) || !nrcFactsOK(l.Facts, l.Logs) || !nrcFactsOK(l.Facts, l.Trace) || !nrcFactsOK(l.Facts, l.LinkedLogs) {
		return false
	}
	if !nhLiveSameQuery(l.TraceOrigin.Origin, l.Records.Question) || l.TraceOrigin.OriginBasis != l.Records.Basis || !reflect.DeepEqual(l.Before, l.After) {
		return false
	}
	// Trace expansion deliberately uses the older coarse window/trace-ID scope.
	if l.Trace.Question.FromMS != l.Records.Question.FromMS || l.Trace.Question.ToMS != l.Records.Question.ToMS || len(l.Trace.Question.Where.All) != 1 || l.Trace.Question.Where.All[0].Field == nil || l.Trace.Question.Where.All[0].Field.Key != "trace_id" {
		return false
	}
	return len(l.Errors) == 0
}
func nrcReplay(t *testing.T, events []nrcEvent, badCounter bool, badOrder bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	reads, depth := 0, 0
	for i, e := range events {
		action := "history"
		target := e.Source
		switch e.Action {
		case "records-staged", "logs-staged":
			action = "stage"
			depth++
		case "native-return", "logs-return":
			action = "back"
			depth--
		}
		for reads < target {
			reads++
			rec.Record(context.Background(), "read", nil, "obs.reads", reads, "obs.depth", depth, "obs.coherent", true)
		}
		if e.Action == "native-run" || e.Action == "records-run" || e.Action == "trace-expanded" || e.Action == "logs-run" {
			continue
		}
		observedReads := reads
		if badCounter && i == 2 {
			observedReads++
		}
		if badOrder && i == 2 {
			action = "back"
		}
		rec.Record(context.Background(), action, nil, "obs.reads", observedReads, "obs.depth", depth, "obs.coherent", true)
	}
	c, e := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	r, e := c.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return r.OK()
}
func TestQuintNativeRecordContextObserved(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_RECORD_CONTEXT_LEDGER")
	if path == "" {
		t.Skip("actual192 ledger opt-in")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l nrcLedger
	if e = json.Unmarshal(raw, &l); e != nil {
		t.Fatal(e)
	}
	if !nrcMainOK(l) {
		s, known := nrcStream(l.Native, l.Before)
		t.Fatalf("actual mapping rejected: stream=%v plan=%v staged=%v records=%v logs=%v trace=%v linkedlogs=%v parent=%v origin=%v basis=%v", known, nrcExpected(l.Staged.Question, s, l.Before.Bounds.FromNS, l.Before.Bounds.ToNS), nhLiveSameQuery(l.Staged.Question, l.Records.Question), nrcFactsOK(l.Facts, l.Records), nrcFactsOK(l.Facts, l.Logs), nrcFactsOK(l.Facts, l.Trace), nrcFactsOK(l.Facts, l.LinkedLogs), reflect.DeepEqual(l.Before, l.After), nhLiveSameQuery(l.TraceOrigin.Origin, l.Records.Question), l.TraceOrigin.OriginBasis == l.Records.Basis)
	}
	wantActions := []string{"native-run", "interval-selected", "records-staged", "records-run", "trace-expanded", "native-return", "logs-staged", "logs-run", "logs-return"}
	wantSources := []int{1, 1, 1, 2, 4, 4, 4, 5, 5}
	if len(l.Events) != len(wantActions) {
		t.Fatal("event count")
	}
	for i, e := range l.Events {
		if e.Action != wantActions[i] || e.Source != wantSources[i] {
			t.Fatal("observed source/order", i)
		}
	}
	for _, mutation := range []string{"count", "upper-edge", "instance", "metric-filter", "origin", "parent-custody"} {
		var m nrcLedger
		json.Unmarshal(raw, &m)
		switch mutation {
		case "count":
			m.Records.Count = "0"
		case "upper-edge":
			m.Staged.Question.Where.All[4].Value = m.Before.Bounds.ToNS
		case "instance":
			m.Staged.Question.Where.All[2].Value = "other"
		case "metric-filter":
			m.Staged.Question.Where.All = append(m.Staged.Question.Where.All, exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "metric.only"}, Op: "exists"})
		case "origin":
			m.TraceOrigin.OriginBasis = "changed"
		case "parent-custody":
			m.After.ParentEvidence += "changed"
		}
		if nrcMainOK(m) {
			t.Fatal("changed semantic control accepted", mutation)
		}
	}
	if !nrcReplay(t, l.Events, false, false) || nrcReplay(t, l.Events, true, false) || nrcReplay(t, l.Events, false, true) {
		t.Fatal("finite observed navigation/counter controls")
	}
	t.Log("actual mapping/raw membership/retained parent accepted;6semantic+2finite controls rejected using existing population_drilldown model")
}
func TestNativeRecordContextCasesAndOwnership(t *testing.T) {
	for _, v := range []string{"GOO_NATIVE_RECORD_CONTEXT_CASES", "GOO_NATIVE_RECORD_CONTEXT_OWNERSHIP"} {
		path := os.Getenv(v)
		if path == "" {
			t.Skip("actual192 cases+ownership opt-in")
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var l nrcLedger
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if strings.HasSuffix(v, "CASES") {
			if len(l.Cases) != 4 {
				t.Fatal("actual case count")
			}
			for _, c := range l.Cases {
				s, ok := nrcStream(c.Native, c.View)
				if !ok || !nrcExpected(c.Staged.Question, s, c.SelectedInterval.FromNS, c.SelectedInterval.ToNS) || !nhLiveSameQuery(c.Staged.Question, c.Result.Question) || !nrcFactsOK(l.Facts, c.Result) || c.SourceAfterReturn != c.SourceBeforeStage+1 {
					t.Fatal("actual case scope/read", c.Name)
				}
				changed := c.Result
				changed.Count = "999999"
				if nrcFactsOK(l.Facts, changed) {
					t.Fatal("case count control")
				}
			}
		} else {
			if !nrcFactsOK(l.Facts, l.Held) || len(l.Events) != 6 {
				t.Fatal("held source evidence")
			}
			want := []int{1, 2, 3, 3, 3, 3}
			actions := []string{"hidden-action-rejected", "record-read-held", "return-during-read", "late-response-rejected", "closed-action-rejected", "browser-back-retains-native"}
			for i, e := range l.Events {
				if e.Action != actions[i] || e.Source != want[i] {
					t.Fatal("owner observed order/counter")
				}
			}
			late := l.Events[3]
			if !late.Visible || !reflect.DeepEqual(late.CancelledView, late.AfterLate) {
				t.Fatal("late response published")
			}
			sources, cancels := 0, 0
			for _, r := range l.Requests {
				if r.Method == "POST" && strings.HasSuffix(r.Endpoint, "/cancel") {
					cancels++
				} else if r.Method == "POST" && (r.Endpoint == "/api/explore" || r.Endpoint == "/api/measurements") {
					sources++
				}
			}
			if sources != 2 || cancels != 1 {
				t.Fatal("owner telemetry vs cancellation counts")
			}
		}
		if len(l.Errors) != 0 {
			t.Fatal("browser errors")
		}
	}
	t.Log(fmt.Sprintf("4actual cases+held-response ownership checked;cancel classified separately from2telemetry reads"))
}

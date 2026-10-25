package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

// No page/fields are synthesized for narrow responses. Full Explore remains
// separately validated by hdLiveResult with its original raw-row checks.
func raLiveFacts(records []hdLiveRecord) []raFact {
	out := []raFact{}
	for _, r := range records {
		n, e := strconv.ParseUint(r.DurationNS, 10, 64)
		if e != nil {
			return nil
		}
		status := "unknown"
		switch r.Status {
		case 0:
			status = "Unset"
		case 1:
			status = "Ok"
		case 2:
			status = "Error"
		}
		out = append(out, raFact{Duration: n, Status: status, TimeMS: r.TimeMS})
	}
	return out
}
func raLiveResult(records []hdLiveRecord, raw json.RawMessage) error {
	var r exploration.RecordAggregatesResult
	if e := json.Unmarshal(raw, &r); e != nil {
		return e
	}
	if r.Version != "goo/record-aggregates/v1" {
		return fmt.Errorf("narrow version")
	}
	var wire map[string]json.RawMessage
	json.Unmarshal(raw, &wire)
	for _, key := range []string{"rows", "fields", "plan", "resultToken", "token", "hasMore", "traceExamples"} {
		if _, ok := wire[key]; ok {
			return fmt.Errorf("narrow response leaked %s", key)
		}
	}
	q := r.Question
	if len(q.GroupBy) != 1 || q.Signal != "spans" {
		return fmt.Errorf("span aggregate shape")
	}
	full := exploration.Result{Question: q, Count: r.Count, Groups: r.Groups, Duration: r.Duration, SpanStatus: r.SpanStatus}
	if e := hdLiveResultMode(records, full, false); e != nil {
		return e
	}
	population := hdLivePopulation(records, q)
	facts := raLiveFacts(population)
	if len(facts) != len(population) {
		return fmt.Errorf("raw duration")
	}
	_, duration, status := raSummary(facts)
	state := "observed"
	if len(population) == 0 {
		state = "no_data"
	}
	if r.State != state {
		return fmt.Errorf("unavailable aggregate")
	}
	if q.Summary != "" && !reflect.DeepEqual(r.Duration, &duration) {
		return fmt.Errorf("global duration/order statistics")
	}
	if q.SpanStatus && !reflect.DeepEqual(r.SpanStatus, &status) {
		return fmt.Errorf("global status")
	}
	for _, rawGroup := range r.Groups {
		var group struct {
			Value    *string                        `json:"g0"`
			Duration *exploration.DurationSummary   `json:"duration"`
			Status   *exploration.SpanStatusSummary `json:"spanStatus"`
		}
		if json.Unmarshal(rawGroup, &group) != nil {
			return fmt.Errorf("group measures")
		}
		members := []hdLiveRecord{}
		for _, record := range population {
			label, known := raLiveLabel(record, q.GroupBy[0])
			if !known {
				return fmt.Errorf("unsupported live group")
			}
			if (label == nil) == (group.Value == nil) && (label == nil || *label == *group.Value) {
				members = append(members, record)
			}
		}
		_, d, s := raSummary(raLiveFacts(members))
		if q.Summary != "" && !reflect.DeepEqual(group.Duration, &d) {
			return fmt.Errorf("group duration/orderstatistics")
		}
		if q.SpanStatus && !reflect.DeepEqual(group.Status, &s) {
			return fmt.Errorf("group status")
		}
	}
	if !cvLiveReceipt(r.Reads.Count) || !cvLiveReceipt(r.Reads.Groups) {
		return fmt.Errorf("aggregate receipts")
	}
	a, _ := time.Parse(time.RFC3339Nano, r.Reads.Count.ReadAt)
	b, _ := time.Parse(time.RFC3339Nano, r.Reads.Groups.ReadAt)
	if b.Before(a) {
		return fmt.Errorf("aggregate read order")
	}
	if q.BucketMS == 0 {
		if len(r.TimeSeries) != 0 || r.Reads.TimeSeries != nil {
			return fmt.Errorf("unrequested series")
		}
	} else {
		if r.Reads.TimeSeries == nil || !cvLiveReceipt(*r.Reads.TimeSeries) {
			return fmt.Errorf("series receipt")
		}
		expected := []exploration.TimePoint{}
		for start := q.FromMS; start < q.ToMS; start += q.BucketMS {
			end := start + q.BucketMS
			if end > q.ToMS {
				end = q.ToMS
			}
			chunk := []raFact{}
			for _, f := range facts {
				if f.TimeMS >= start && f.TimeMS < end {
					chunk = append(chunk, f)
				}
			}
			_, d, s := raSummary(chunk)
			point := exploration.TimePoint{FromMS: start, ToMS: end, Count: strconv.Itoa(len(chunk)), State: "observed"}
			if len(chunk) == 0 {
				point.State = "no_data"
			}
			if q.Summary != "" {
				point.Duration = &d
			}
			if q.SpanStatus {
				point.SpanStatus = &s
			}
			expected = append(expected, point)
		}
		if !reflect.DeepEqual(r.TimeSeries, expected) {
			return fmt.Errorf("complete series population")
		}
	}
	return nil
}

func raLiveLabel(r hdLiveRecord, f exploration.Field) (*string, bool) {
	switch f.Kind {
	case "attribute":
		v, ok := r.Attributes[f.Key]
		if !ok {
			return nil, true
		}
		return &v, true
	case "column":
		switch f.Key {
		case "service":
			return &r.Service, true
		case "name":
			return &r.Name, true
		}
	case "resource":
		if f.Key == "service.namespace" {
			return &r.Namespace, true
		}
	}
	return nil, false
}

type raLogFact struct {
	ID, Environment, Service, Namespace, Severity, Body string
	TimeMS                                              int64
	Attributes                                          map[string]string
}

func raLiveJournalResult(spans []hdLiveRecord, logs []raLogFact, raw json.RawMessage) error {
	var header struct{ Question exploration.Query }
	if json.Unmarshal(raw, &header) != nil {
		return fmt.Errorf("aggregate JSON")
	}
	if header.Question.Signal == "spans" {
		return raLiveResult(spans, raw)
	}
	if header.Question.Signal != "logs" {
		return fmt.Errorf("unsupported aggregate signal")
	}
	return raLiveLogResult(logs, raw)
}
func raLiveLogResult(facts []raLogFact, raw json.RawMessage) error {
	var r exploration.RecordAggregatesResult
	if json.Unmarshal(raw, &r) != nil || r.Version != "goo/record-aggregates/v1" {
		return fmt.Errorf("log aggregate version")
	}
	var wire map[string]json.RawMessage
	json.Unmarshal(raw, &wire)
	for _, key := range []string{"rows", "fields", "plan", "resultToken", "token", "hasMore", "traceExamples"} {
		if _, ok := wire[key]; ok {
			return fmt.Errorf("log inspection leak")
		}
	}
	q := r.Question
	if len(q.GroupBy) != 1 || q.GroupBy[0] != (exploration.Field{Kind: "column", Key: "severity"}) || r.Duration != nil || r.SpanStatus != nil {
		return fmt.Errorf("log grain/projections")
	}
	population := []raLogFact{}
	groups := map[string]int{}
	for _, f := range facts {
		values := map[exploration.Field]string{{Kind: "column", Key: "service"}: f.Service, {Kind: "column", Key: "severity"}: f.Severity, {Kind: "resource", Key: "service.namespace"}: f.Namespace}
		for k, v := range f.Attributes {
			values[exploration.Field{Kind: "attribute", Key: k}] = v
		}
		if f.Environment == q.Environment && f.TimeMS >= q.FromMS && f.TimeMS < q.ToMS && (q.Where == nil || tsPredicate(*q.Where, values)) {
			population = append(population, f)
			groups[f.Severity]++
		}
	}
	if r.Count != strconv.Itoa(len(population)) {
		return fmt.Errorf("independent log count")
	}
	state := "observed"
	if len(population) == 0 {
		state = "no_data"
	}
	if r.State != state {
		return fmt.Errorf("log state")
	}
	got := map[string]int{}
	for _, rawGroup := range r.Groups {
		var g exploration.CountedValue
		if json.Unmarshal(rawGroup, &g) != nil || g.Value == nil {
			return fmt.Errorf("log group")
		}
		n, e := strconv.Atoi(g.Count)
		if e != nil || n < 1 {
			return fmt.Errorf("log group count")
		}
		if _, duplicate := got[*g.Value]; duplicate {
			return fmt.Errorf("duplicate log group")
		}
		got[*g.Value] = n
	}
	if !reflect.DeepEqual(got, groups) {
		return fmt.Errorf("log group population")
	}
	if !cvLiveReceipt(r.Reads.Count) || !cvLiveReceipt(r.Reads.Groups) {
		return fmt.Errorf("log read receipt")
	}
	if q.BucketMS == 0 {
		if len(r.TimeSeries) != 0 || r.Reads.TimeSeries != nil {
			return fmt.Errorf("unrequested log series")
		}
	} else {
		if r.Reads.TimeSeries == nil || !cvLiveReceipt(*r.Reads.TimeSeries) {
			return fmt.Errorf("log series receipt")
		}
		expected := []exploration.TimePoint{}
		for start := q.FromMS; start < q.ToMS; start += q.BucketMS {
			end := start + q.BucketMS
			if end > q.ToMS {
				end = q.ToMS
			}
			n := 0
			for _, f := range population {
				if f.TimeMS >= start && f.TimeMS < end {
					n++
				}
			}
			s := "observed"
			if n == 0 {
				s = "no_data"
			}
			expected = append(expected, exploration.TimePoint{FromMS: start, ToMS: end, Count: strconv.Itoa(n), State: s})
		}
		if !reflect.DeepEqual(r.TimeSeries, expected) {
			return fmt.Errorf("log series population")
		}
	}
	return nil
}

func raLiveLogFull(facts []raLogFact, r exploration.Result) bool {
	population := map[string]raLogFact{}
	for _, f := range facts {
		values := map[exploration.Field]string{{Kind: "column", Key: "service"}: f.Service, {Kind: "column", Key: "severity"}: f.Severity, {Kind: "resource", Key: "service.namespace"}: f.Namespace}
		for k, v := range f.Attributes {
			values[exploration.Field{Kind: "attribute", Key: k}] = v
		}
		if f.Environment == r.Question.Environment && f.TimeMS >= r.Question.FromMS && f.TimeMS < r.Question.ToMS && (r.Question.Where == nil || tsPredicate(*r.Question.Where, values)) {
			population[f.Body] = f
		}
	}
	if r.Count != strconv.Itoa(len(population)) {
		return false
	}
	want := len(population)
	if want > r.Question.Limit {
		want = r.Question.Limit
	}
	if len(r.Rows) != want {
		return false
	}
	seen := map[string]bool{}
	for _, raw := range r.Rows {
		var row struct {
			Body       string
			Timestamp  string `json:"timestamp_ns"`
			Service    string `json:"ServiceName"`
			Severity   string `json:"SeverityText"`
			Attributes map[string]string
			Resource   map[string]string `json:"ResourceAttributes"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		f, ok := population[row.Body]
		if !ok || seen[row.Body] {
			return false
		}
		seen[row.Body] = true
		if row.Timestamp != strconv.FormatInt(f.TimeMS*1000000, 10) || row.Service != f.Service || row.Severity != f.Severity || row.Resource["service.namespace"] != f.Namespace || !reflect.DeepEqual(row.Attributes, f.Attributes) {
			return false
		}
	}
	return true
}

type raObservedResponse struct {
	Endpoint                                             string
	Status                                               int
	Body                                                 json.RawMessage
	Aborted, BodyUnavailable, Cancelled, CancelRequested bool
	Control                                              string
	BodyObserved                                         *bool
}

func raResponseComplete(r raObservedResponse) (bool, error) {
	if r.Control != "" {
		if r.Control == "cancel" && strings.HasSuffix(r.Endpoint, "/cancel") && r.BodyObserved != nil && !*r.BodyObserved {
			return false, nil
		}
		return false, fmt.Errorf("invalid control observation")
	}
	if r.Aborted || r.BodyUnavailable {
		if r.Aborted && r.BodyUnavailable && (r.Cancelled || r.CancelRequested) {
			return false, nil
		}
		return false, fmt.Errorf("partial positive body without cancellation/failure evidence")
	}
	return r.Status == 200, nil
}

type raServiceLedger struct {
	Facts struct {
		Records []hdLiveRecord
		Logs    []raLogFact
	}
	Initial struct {
		Plan       exploration.DashboardPlan
		Results    map[string]json.RawMessage
		Supporting exploration.Result
	}
	Records, Operation, Trace, TraceLogs exploration.Result
	Saved, Reopened                      exploration.CapturePage
	StoredBefore, StoredAfter            struct{ SHA256 string }
	Events                               []struct {
		Action        string
		Source, Plans int
	}
	Responses []raObservedResponse
	Errors    []json.RawMessage
}

func raServiceOK(l raServiceLedger) bool {
	if len(l.Errors) != 0 || len(l.Facts.Records) != 225 || len(l.Facts.Logs) != 160 {
		return false
	}
	if hdLiveResult(l.Facts.Records, l.Initial.Supporting) != nil || !raLiveLogFull(l.Facts.Logs, l.Records) || hdLiveResult(l.Facts.Records, l.Operation) != nil || hdLiveResult(l.Facts.Records, l.Trace) != nil || !raLiveLogFull(l.Facts.Logs, l.TraceLogs) {
		return false
	}
	for _, w := range l.Initial.Plan.Widgets {
		if w.Query == nil {
			continue
		}
		raw, ok := l.Initial.Results[w.Key]
		if !ok || raLiveJournalResult(l.Facts.Records, l.Facts.Logs, raw) != nil {
			return false
		}
		var header struct{ Question exploration.Query }
		json.Unmarshal(raw, &header)
		if !nhLiveSameQuery(*w.Query, header.Question) {
			return false
		}
	}
	if l.Saved.Explore == nil || hdLiveResult(l.Facts.Records, *l.Saved.Explore) != nil || len(l.Saved.Explore.Fields) == 0 || !reflect.DeepEqual(l.Saved, l.Reopened) {
		return false
	}
	for i, e := range l.Events {
		switch e.Action {
		case "logs-staged", "logs-return", "frozen-api-reopen", "capture-return", "operation-return", "frozen-ui-reopen":
			if i == 0 || e.Source != l.Events[i-1].Source {
				return false
			}
		}
	}
	for _, r := range l.Responses {
		complete, e := raResponseComplete(r)
		if e != nil {
			return false
		}
		if !complete {
			continue
		}
		if r.Endpoint == "/api/explore" {
			var full exploration.Result
			if json.Unmarshal(r.Body, &full) != nil || full.Count == "" || (full.Question.Signal != "spans" && full.Question.Signal != "logs") {
				return false
			}
			if full.Question.Signal == "spans" && hdLiveResult(l.Facts.Records, full) != nil {
				return false
			}
			if full.Question.Signal == "logs" && !raLiveLogFull(l.Facts.Logs, full) {
				return false
			}
		}
		if r.Endpoint == "/api/record-aggregates" && raLiveJournalResult(l.Facts.Records, l.Facts.Logs, r.Body) != nil {
			return false
		}
	}
	return true
}
func TestServiceRecordAggregatesObserved(t *testing.T) {
	path := os.Getenv("GOO_SERVICE_AGGREGATES_LEDGER")
	if path == "" {
		t.Skip("actual194 service aggregate ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l raServiceLedger
	if e = json.Unmarshal(raw, &l); e != nil {
		t.Fatal(e)
	}
	if !raServiceOK(l) {
		t.Fatal("actual service aggregate/full-record/return ledger rejected")
	}
	before, e := os.ReadFile(filepath.Join(filepath.Dir(path), "capture-before.json"))
	if e != nil {
		t.Fatal(e)
	}
	after, e := os.ReadFile(filepath.Join(filepath.Dir(path), "capture-after.json"))
	if e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(before)
	if !reflect.DeepEqual(before, after) || hex.EncodeToString(sum[:]) != l.StoredBefore.SHA256 || l.StoredAfter.SHA256 != l.StoredBefore.SHA256 {
		t.Fatal("full capture custody bytes")
	}
	var stored exploration.SavedCapture
	if json.Unmarshal(before, &stored) != nil || stored.Explore == nil || stored.ID != l.Saved.Receipt.ID || hdLiveResultMode(l.Facts.Records, *stored.Explore, true, 1000) != nil || len(stored.Explore.Rows) != 137 || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Explore: stored.Explore}) {
		t.Fatal("full retained capture hash/membership")
	}
	for _, name := range []string{"count", "group-quantile", "series", "scope", "receipt", "inspection-leak", "full-records", "return-read"} {
		var m raServiceLedger
		json.Unmarshal(raw, &m)
		original, _ := json.Marshal(m)
		switch name {
		case "full-records":
			m.Saved.Explore.Rows = nil
		case "return-read":
			for i := range m.Events {
				if m.Events[i].Action == "logs-return" {
					m.Events[i].Source++
				}
			}
		default:
			var value map[string]json.RawMessage
			json.Unmarshal(m.Initial.Results["spans"], &value)
			switch name {
			case "count":
				value["count"] = json.RawMessage(`"0"`)
			case "inspection-leak":
				value["rows"] = json.RawMessage(`[]`)
			case "scope":
				var q exploration.Query
				json.Unmarshal(value["question"], &q)
				q.Environment += "-other"
				value["question"], _ = json.Marshal(q)
			case "receipt":
				var reads exploration.RecordAggregatesReads
				json.Unmarshal(value["reads"], &reads)
				reads.Count.Basis = ""
				value["reads"], _ = json.Marshal(reads)
			case "group-quantile":
				var groups []map[string]json.RawMessage
				json.Unmarshal(value["groups"], &groups)
				var d exploration.DurationSummary
				json.Unmarshal(groups[0]["duration"], &d)
				d.P95NS = "1"
				groups[0]["duration"], _ = json.Marshal(d)
				value["groups"], _ = json.Marshal(groups)
			case "series":
				var points []exploration.TimePoint
				json.Unmarshal(value["timeSeries"], &points)
				points[0].Count = "999"
				value["timeSeries"], _ = json.Marshal(points)
			}
			m.Initial.Results["spans"], _ = json.Marshal(value)
		}
		changed, _ := json.Marshal(m)
		if string(original) == string(changed) {
			t.Fatal("unchanged control", name)
		}
		if raServiceOK(m) {
			t.Fatal("changed service control accepted", name)
		}
	}
	t.Log("actual194 narrow span/log aggregates, full supporting/drilldown pages, retained137-row capture hash/bytes and source-free returns passed;8changed controls rejected")
}

type raOwnerLedger struct {
	Facts struct {
		Records []hdLiveRecord
		Logs    []raLogFact
	}
	OriginalPlan, FreshPlan exploration.DashboardPlan
	Held                    struct {
		Question exploration.Query
		Result   json.RawMessage
	}
	Fresh         json.RawMessage
	BeforeRelease int
	AfterRelease  struct{ Hidden, Visible bool }
	Responses     []raObservedResponse
	Errors        []json.RawMessage
}

func raOwnerOK(l raOwnerLedger) bool {
	if len(l.Errors) != 0 || !l.AfterRelease.Hidden || l.AfterRelease.Visible || l.BeforeRelease < 1 {
		return false
	}
	var original, fresh *exploration.Query
	for _, w := range l.OriginalPlan.Widgets {
		if w.Key == "spans" {
			original = w.Query
		}
	}
	for _, w := range l.FreshPlan.Widgets {
		if w.Key == "spans" {
			fresh = w.Query
		}
	}
	if original == nil || fresh == nil || !nhLiveSameQuery(*original, *fresh) {
		return false
	}
	// The real held response belongs to the explicit All spans switch, not
	// the original HTTP overview. Bind it to its actual pure-plan receipt.
	heldPlanned := false
	for _, r := range l.Responses {
		if r.Endpoint != "/api/dashboards/plan" || r.Status != 200 {
			continue
		}
		var p exploration.DashboardPlan
		if json.Unmarshal(r.Body, &p) != nil {
			continue
		}
		if p.Profile != "all-spans" || p.Service != l.OriginalPlan.Service || !reflect.DeepEqual(p.Scope, l.OriginalPlan.Scope) {
			continue
		}
		for _, w := range p.Widgets {
			if w.Key == "spans" && w.Query != nil && nhLiveSameQuery(*w.Query, l.Held.Question) {
				heldPlanned = true
			}
		}
	}
	if !heldPlanned {
		return false
	}
	for i, raw := range []json.RawMessage{l.Held.Result, l.Fresh} {
		if raLiveJournalResult(l.Facts.Records, l.Facts.Logs, raw) != nil {
			return false
		}
		var header struct{ Question exploration.Query }
		json.Unmarshal(raw, &header)
		want := l.Held.Question
		if i == 1 {
			want = *fresh
		}
		if !nhLiveSameQuery(header.Question, want) {
			return false
		}
	}
	for _, r := range l.Responses {
		complete, e := raResponseComplete(r)
		if e != nil {
			return false
		}
		if complete && r.Endpoint == "/api/record-aggregates" && raLiveJournalResult(l.Facts.Records, l.Facts.Logs, r.Body) != nil {
			return false
		}
	}
	return true
}
func TestServiceRecordAggregatesOwnership(t *testing.T) {
	path := os.Getenv("GOO_SERVICE_AGGREGATES_OWNERSHIP")
	if path == "" {
		t.Skip("actual194 held-response ownership ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l raOwnerLedger
	if json.Unmarshal(raw, &l) != nil || !raOwnerOK(l) {
		t.Fatal("actual held aggregate/stale owner/fresh recovery rejected")
	}
	for _, name := range []string{"held-count", "held-binding", "stale-visible", "fresh-receipt"} {
		var m raOwnerLedger
		json.Unmarshal(raw, &m)
		before, _ := json.Marshal(m)
		switch name {
		case "held-binding":
			m.Held.Question.Environment += "-wrong"
		case "stale-visible":
			m.AfterRelease.Visible = true
		default:
			var result exploration.RecordAggregatesResult
			if name == "held-count" {
				json.Unmarshal(m.Held.Result, &result)
				result.Count = "0"
				m.Held.Result, _ = json.Marshal(result)
			} else {
				json.Unmarshal(m.Fresh, &result)
				result.Reads.Count.Basis = ""
				m.Fresh, _ = json.Marshal(result)
			}
		}
		after, _ := json.Marshal(m)
		if string(before) == string(after) {
			t.Fatal("unchanged owner control", name)
		}
		if raOwnerOK(m) {
			t.Fatal("changed owner control accepted", name)
		}
	}
	t.Log("actual complete held DTO and fresh recovery validated;stale owner remains hidden;4changed controls rejected;cancel headers not treated as telemetry bodies")
}

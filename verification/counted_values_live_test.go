package verification

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

type cvLiveFact struct {
	ID, Environment, Service, Namespace, Body string
	TimeMS                                    int64
	Attributes                                map[string]string
}
type cvLiveLedger struct {
	Facts              struct{ Records []cvLiveFact }
	Original, Selected exploration.Result
	Large              struct {
		Request exploration.Query
		Data    exploration.CountedValuesResult
	}
	SmallRequest              exploration.Query
	Small, HeldData           exploration.CountedValuesResult
	BeforeBrowse, BeforeLocal int
	StalePublished            bool
	Errors                    []json.RawMessage
	Requests                  []struct {
		Method, Endpoint string
		Payload          json.RawMessage
	}
}

func cvLiveMembers(facts []cvLiveFact, q exploration.Query) []cvLiveFact {
	out := []cvLiveFact{}
	for _, f := range facts {
		fields := map[exploration.Field]string{{Kind: "column", Key: "service"}: f.Service, {Kind: "column", Key: "body"}: f.Body, {Kind: "resource", Key: "service.namespace"}: f.Namespace}
		for k, v := range f.Attributes {
			fields[exploration.Field{Kind: "attribute", Key: k}] = v
		}
		if f.Environment == q.Environment && f.TimeMS >= q.FromMS && f.TimeMS < q.ToMS && (q.Where == nil || tsPredicate(*q.Where, fields)) {
			out = append(out, f)
		}
	}
	return out
}
func cvLiveReceipt(r exploration.CountedValuesRead) bool {
	b, e := hex.DecodeString(r.Basis)
	_, clock := time.Parse(time.RFC3339Nano, r.ReadAt)
	return e == nil && len(b) == 32 && clock == nil
}
func cvLiveData(facts []cvLiveFact, q exploration.Query, r exploration.CountedValuesResult) bool {
	if !nhLiveSameQuery(q, r.Question) || len(q.GroupBy) != 1 || r.Version != "goo/counted-values/v1" || !cvLiveReceipt(r.Reads.Count) || !cvLiveReceipt(r.Reads.Groups) {
		return false
	}
	if q.GroupBy[0].Kind != "attribute" {
		return false
	}
	members := cvLiveMembers(facts, q)
	if r.Count != strconv.Itoa(len(members)) {
		return false
	}
	state := "observed"
	if len(members) == 0 {
		state = "no_data"
	}
	if r.State != state {
		return false
	}
	expected := map[string]int{}
	for _, f := range members {
		v, ok := f.Attributes[q.GroupBy[0].Key]
		key := "missing"
		if ok {
			key = "present:" + v
		}
		expected[key]++
	}
	got := map[string]int{}
	for _, g := range r.Groups {
		key := "missing"
		if g.Value != nil {
			key = "present:" + *g.Value
		}
		if _, ok := got[key]; ok {
			return false
		}
		n, e := strconv.Atoi(g.Count)
		if e != nil || n <= 0 || strconv.Itoa(n) != g.Count {
			return false
		}
		got[key] = n
	}
	if !reflect.DeepEqual(got, expected) {
		return false
	}
	countClock, _ := time.Parse(time.RFC3339Nano, r.Reads.Count.ReadAt)
	groupClock, _ := time.Parse(time.RFC3339Nano, r.Reads.Groups.ReadAt)
	captured, e := time.Parse(time.RFC3339Nano, r.CapturedAt)
	hash, e2 := hex.DecodeString(r.Basis)
	return e == nil && e2 == nil && len(hash) == 32 && !groupClock.Before(countClock) && !captured.Before(groupClock)
}
func cvLiveFull(facts []cvLiveFact, r exploration.Result) bool {
	members := cvLiveMembers(facts, r.Question)
	if r.Count != strconv.Itoa(len(members)) {
		return false
	}
	byBody := map[string]cvLiveFact{}
	for _, f := range members {
		byBody[f.Body] = f
	}
	want := len(members)
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
			Service    string            `json:"ServiceName"`
			Timestamp  string            `json:"timestamp_ns"`
			Resource   map[string]string `json:"ResourceAttributes"`
			Attributes map[string]string
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		f, ok := byBody[row.Body]
		if !ok || seen[row.Body] {
			return false
		}
		seen[row.Body] = true
		if row.Service != f.Service || row.Timestamp != strconv.FormatInt(f.TimeMS*1000000, 10) || row.Resource["service.namespace"] != f.Namespace || !reflect.DeepEqual(row.Attributes, f.Attributes) {
			return false
		}
	}
	return true
}
func cvLiveOK(l cvLiveLedger) bool {
	if len(l.Errors) != 0 || l.StalePublished || l.BeforeBrowse < 1 || l.BeforeLocal != l.BeforeBrowse+1 {
		return false
	}
	if !cvLiveData(l.Facts.Records, l.Large.Request, l.Large.Data) || !cvLiveData(l.Facts.Records, l.SmallRequest, l.Small) || !cvLiveData(l.Facts.Records, l.Large.Request, l.HeldData) || !cvLiveFull(l.Facts.Records, l.Original) || !cvLiveFull(l.Facts.Records, l.Selected) {
		return false
	}
	// The counted projection changes only grouping/presentation; all parent filters remain.
	want := l.Original.Question
	want.Limit = 1
	want.Offset = 0
	want.GroupBy = []exploration.Field{{Kind: "attribute", Key: "release"}}
	if !nhLiveSameQuery(want, l.Large.Request) {
		return false
	}
	seq := []string{}
	for _, r := range l.Requests {
		if r.Method == "POST" && (r.Endpoint == "/api/explore" || r.Endpoint == "/api/explore/counted-values") {
			seq = append(seq, r.Endpoint)
		}
	}
	return reflect.DeepEqual(seq, []string{"/api/explore", "/api/explore/counted-values", "/api/explore", "/api/explore", "/api/explore/counted-values"})
}
func TestCountedValuesObserved(t *testing.T) {
	path := os.Getenv("GOO_COUNTED_VALUES_LEDGER")
	if path == "" {
		t.Skip("actual193 ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var l cvLiveLedger
	if e = json.Unmarshal(raw, &l); e != nil {
		t.Fatal(e)
	}
	if !cvLiveOK(l) {
		t.Fatalf("actual counted values rejected: large=%v small=%v held=%v original=%v selected=%v browse=%d local=%d", cvLiveData(l.Facts.Records, l.Large.Request, l.Large.Data), cvLiveData(l.Facts.Records, l.SmallRequest, l.Small), cvLiveData(l.Facts.Records, l.Large.Request, l.HeldData), cvLiveFull(l.Facts.Records, l.Original), cvLiveFull(l.Facts.Records, l.Selected), l.BeforeBrowse, l.BeforeLocal)
	}
	for _, name := range []string{"total", "group-count", "missing-to-empty", "read-receipt", "request-binding", "numeric-rounded", "state", "stale-publish"} {
		var m cvLiveLedger
		json.Unmarshal(raw, &m)
		before, _ := json.Marshal(m)
		switch name {
		case "total":
			m.Large.Data.Count = "0"
		case "group-count":
			m.Large.Data.Groups[0].Count = "999"
		case "missing-to-empty":
			for i, g := range m.Large.Data.Groups {
				if g.Value == nil {
					m.Large.Data.Groups[i].Value = lpcPointer("")
				}
			}
		case "read-receipt":
			m.Large.Data.Reads.Count.Basis = ""
		case "request-binding":
			m.Large.Data.Question.Environment += "-wrong"
		case "numeric-rounded":
			m.Large.Request.Where.All[3].Value = "9007199254740992"
			m.Large.Data.Question.Where.All[3].Value = "9007199254740992"
		case "state":
			m.Large.Data.State = "changed_between_reads"
		case "stale-publish":
			m.StalePublished = true
		}
		after, _ := json.Marshal(m)
		if string(before) == string(after) {
			t.Fatal("unchanged control", name)
		}
		if cvLiveOK(m) {
			t.Fatal("changed control accepted", name)
		}
	}
	t.Log("real193 raw membership/count/group/receipt/binding/ownership checks passed;8guaranteed changed controls rejected;no new model")
}

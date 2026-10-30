package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func viewFixture() *Result {
	q := basic()
	q.GroupBy = []Field{{Kind: "attribute", Key: "team"}}
	return &Result{Question: q, Basis: strings.Repeat("a", 64), Count: "18446744073709551615", State: "observed", HasMore: true, CapturedAt: "2026-10-30T00:00:00Z", Groups: []json.RawMessage{json.RawMessage(`{"g0":null,"count":"18446744073709551614"}`), json.RawMessage(`{"g0":"","count":"1"}`)}}
}
func TestViewDatasetPreservesExactSupportAndScope(t *testing.T) {
	r := viewFixture()
	d := groupViewDataset(r, true)
	if d == nil || d.Version != "goo.dataset/1" || d.Quality.Extent != "complete" || d.Quality.Sampling != "unknown" || d.Completeness == nil || !reflect.DeepEqual(d.Completeness.Aspects, []string{"groups"}) {
		t.Fatal(d)
	}
	if d.Rows[0].Values.Category != nil || d.Rows[1].Values.Category == nil || *d.Rows[1].Values.Category != "" || d.Rows[0].Key == d.Rows[1].Key || d.Rows[0].Values.Count != "18446744073709551614" {
		t.Fatal(d.Rows)
	}
	var q Query
	if json.Unmarshal(d.Scope.Selector, &q) != nil || !reflect.DeepEqual(q, r.Question) {
		t.Fatal("lost selector", string(d.Scope.Selector))
	}
	before := string(d.Scope.Selector)
	r.Question.GroupBy[0].Key = "mutated"
	if string(d.Scope.Selector) != before {
		t.Fatal("caller mutated snapshot")
	}
	if d.Revision != r.Basis || d.Scope.Generation != r.Basis || d.Completeness.ScopeID != d.Scope.ID || d.Completeness.Revision != d.Revision {
		t.Fatal("unbound attestation", d)
	}
}
func TestViewDatasetCompletenessComesFromReadNotCount(t *testing.T) {
	r := viewFixture()
	// Even perfect count agreement is insufficient without an exhaustive read.
	d := groupViewDataset(r, false)
	if d.Completeness != nil || d.Quality.Extent != "unknown" {
		t.Fatal(d)
	}
	// A read can be complete independently of a bounded/hasMore record page.
	d = groupViewDataset(r, true)
	if d.Completeness == nil || !r.HasMore {
		t.Fatal(d)
	}
	r.State = "changed_between_reads"
	d = groupViewDataset(r, true)
	if d.Completeness != nil || d.Quality.Extent != "unknown" {
		t.Fatal(d)
	}
	r.State = "no_data"
	r.Groups = nil
	r.Count = "0"
	d = groupViewDataset(r, true)
	if d.Completeness == nil || len(d.Rows) != 0 {
		t.Fatal(d)
	}
}
func TestViewDatasetRejectsUnsupportedOrMalformedGroups(t *testing.T) {
	cases := []func(*Result){func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":1,"count":"1"}`)} }, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"g0":"x","count":"01"}`)} }, func(r *Result) { r.Groups = []json.RawMessage{json.RawMessage(`{"count":"1"}`)} }, func(r *Result) { r.Groups = append(r.Groups, r.Groups[0]) }, func(r *Result) { r.Question.GroupBy = append(r.Question.GroupBy, Field{"column", "service"}) }, func(r *Result) { r.Question.Summary = "span-duration" }, func(r *Result) { r.Basis = "unbound" }}
	for i, edit := range cases {
		r := viewFixture()
		edit(r)
		if groupViewDataset(r, true) != nil {
			t.Fatal("admitted", i)
		}
	}
}
func TestSourceSuppliesDatasetWithoutAdditionalReads(t *testing.T) {
	reads := 0
	source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(req *http.Request) (*http.Response, error) {
		reads++
		body, _ := io.ReadAll(req.Body)
		sql := string(body)
		text := ""
		switch {
		case strings.Contains(sql, "GROUP BY"):
			if !strings.Contains(sql, "LIMIT 1001") || req.URL.Query().Get("result_overflow_mode") != "throw" {
				t.Fatal("no exhaustive group budget")
			}
			text = `{"g0":"alpha","count":"3"}`
		case strings.Contains(sql, "SELECT DISTINCT"):
			text = ""
		case strings.Contains(sql, "count()"):
			text = `{"count":"3"}`
		default:
			text = "{\"timestamp_ns\":\"1\",\"duration_ns\":\"0\"}\n{\"timestamp_ns\":\"2\",\"duration_ns\":\"0\"}\n{\"timestamp_ns\":\"3\",\"duration_ns\":\"0\"}"
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
	})}}
	q := basic()
	q.GroupBy = []Field{{Kind: "column", Key: "service"}}
	result, err := source.Fetch(context.Background(), q)
	if err != nil || reads != 4 || result.ViewDataset == nil || result.ViewDataset.Completeness == nil {
		t.Fatal(result, err, reads)
	}
	if result.ViewDataset.Rows[0].Values.Count != "3" || !result.HasMore {
		t.Fatal(result)
	}
}
func TestViewDatasetStableKeysAndIdentity(t *testing.T) {
	for n := 0; n < 100; n++ {
		r := viewFixture()
		r.Groups = nil
		for i := 0; i < n; i++ {
			raw, _ := json.Marshal(map[string]string{"g0": fmt.Sprintf("category-%d", i), "count": fmt.Sprint(i + 1)})
			r.Groups = append(r.Groups, raw)
		}
		a := groupViewDataset(r, true)
		b := groupViewDataset(r, true)
		if !reflect.DeepEqual(a, b) || len(a.Rows) != n {
			t.Fatal("unstable", n)
		}
		r.Question.ToMS++
		c := groupViewDataset(r, true)
		if c.Scope.ID == a.Scope.ID {
			t.Fatal("scope identity reused", n)
		}
	}
}

func TestViewDatasetCategoryByteBudgetAndHashedKeys(t *testing.T) {
	r := viewFixture()
	// JSON escaping may exceed the key budget although the actual category fits.
	value := strings.Repeat("\"", 1024)
	raw, _ := json.Marshal(map[string]string{"g0": value, "count": "1"})
	r.Groups = []json.RawMessage{raw}
	d := groupViewDataset(r, true)
	if d == nil || len(d.Rows[0].Key) != len("category:")+64 || *d.Rows[0].Values.Category != value {
		t.Fatal("valid escaped category lost", d)
	}
	// Do not truncate identities or silently drop an over-budget group.
	for _, value := range []string{strings.Repeat("x", 1025), strings.Repeat("界", 342)} {
		raw, _ := json.Marshal(map[string]string{"g0": value, "count": "1"})
		r.Groups = []json.RawMessage{raw}
		if groupViewDataset(r, true) != nil {
			t.Fatal("over-byte-budget snapshot admitted")
		}
	}
}

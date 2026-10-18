package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGroupedDurationReadAndDisagreement(t *testing.T) {
	q := basic()
	q.Summary = "span-duration"
	q.GroupBy = []Field{{"attribute", "route"}}
	p, e := Compile(q)
	if e != nil || !strings.Contains(p.Groups, "quantilesExactHigh") || strings.Contains(p.Count, "LIMIT") {
		t.Fatal(p, e)
	}
	mode := "valid"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		switch {
		case strings.Contains(sql, "GROUP BY g0"):
			fmt.Fprintln(w, `{"g0":null,"count":"1","sum_ns":"100","min_ns":"100","max_ns":"100","quantiles_ns":["100","100","100"]}`)
			if mode == "duplicate" {
				fmt.Fprintln(w, `{"g0":null,"count":"1","sum_ns":"200","min_ns":"200","max_ns":"200","quantiles_ns":["200","200","200"]}`)
				return
			}
			if mode == "missing" {
				fmt.Fprintln(w, `{"g0":"","count":"1"}`)
				return
			}
			value := "200"
			if mode == "changed" {
				value = "201"
			}
			fmt.Fprintf(w, "{\"g0\":\"\",\"count\":\"1\",\"sum_ns\":\"%s\",\"min_ns\":\"%s\",\"max_ns\":\"%s\",\"quantiles_ns\":[\"%s\",\"%s\",\"%s\"]}\n", value, value, value, value, value, value)
		case strings.Contains(sql, "SELECT toString(count())"):
			fmt.Fprintln(w, `{"count":"2","sum_ns":"300","min_ns":"100","max_ns":"200","quantiles_ns":["200","200","200"]}`)
		case strings.Contains(sql, " AS timestamp_ns"):
			fmt.Fprintln(w, `{"timestamp_ns":"1000","duration_ns":"100"}`)
		}
	}))
	defer srv.Close()
	source := Source{Endpoint: srv.URL}
	r, e := source.Fetch(context.Background(), q)
	if e != nil || r.State != "observed" || len(r.Groups) != 2 {
		t.Fatal(r, e)
	}
	var g struct {
		Value    *string          `json:"g0"`
		Duration *DurationSummary `json:"duration"`
	}
	if json.Unmarshal(r.Groups[0], &g) != nil || g.Value != nil || g.Duration == nil || g.Duration.P95NS != "100" || strings.Contains(string(r.Groups[0]), "quantiles_ns") {
		t.Fatal(string(r.Groups[0]))
	}
	mode = "changed"
	r, e = source.Fetch(context.Background(), q)
	if e != nil || r.State != "changed_between_reads" || r.Count != "2" {
		t.Fatal(r, e)
	}
	for _, m := range []string{"duplicate", "missing"} {
		mode = m
		if r, e = source.Fetch(context.Background(), q); e == nil || r != nil {
			t.Fatal("invalid group accepted", m)
		}
	}
}

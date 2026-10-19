package exploration

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServiceDiscoveryScopeAndValidation(t *testing.T) {
	q := InventoryQuestion{Environment: "prod'; DROP", FromMS: 1000, ToMS: 3000}
	p, e := CompileServiceDiscovery(q)
	if e != nil || strings.Contains(p.Count, q.Environment) || !strings.Contains(p.Groups, "length(p.services)>1") || !strings.Contains(p.Groups, "c.TraceId=p.TraceId") {
		t.Fatal(p, e)
	}
	bad := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "WITH captured") {
			if bad {
				fmt.Fprintln(w, `{"child":"worker","parent":null,"kind":"edge","count":"1"}`)
			} else {
				fmt.Fprintln(w, `{"child":"worker","parent":"api","kind":"edge","count":"2"}`)
				fmt.Fprintln(w, `{"child":"worker","parent":"","kind":"ambiguous","count":"1"}`)
			}
		} else {
			fmt.Fprintln(w, `{"service":"api","signal":"spans","count":"18446744073709551615","first_ms":"1000","last_ms":"2999"}`)
			fmt.Fprintln(w, `{"service":"logger","signal":"logs","count":"1","first_ms":"1100","last_ms":"1100"}`)
		}
	}))
	defer srv.Close()
	source := Source{Endpoint: srv.URL}
	r, e := source.DiscoverServices(context.Background(), q)
	if e != nil || len(r.Signals) != 2 || len(r.Dependencies) != 2 || r.Signals[0].Count != "18446744073709551615" {
		t.Fatal(r, e)
	}
	bad = true
	if _, e = source.DiscoverServices(context.Background(), q); e == nil {
		t.Fatal("missing dimension accepted")
	}
}

func TestMetricServiceDiscoveryValidation(t *testing.T) {
	q := InventoryQuestion{Environment: "prod'\n", FromMS: 1000, ToMS: 3000}
	plan, err := CompileServiceDiscovery(q)
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range []string{"gauge", "sum", "histogram"} {
		if !strings.Contains(plan.Count, "otel_metrics_"+family) {
			t.Fatal("missing metric family", family)
		}
	}
	if strings.Contains(plan.Count, q.Environment) || !strings.Contains(plan.Count, "toString(sum(n))") || strings.Contains(plan.Count, "sum(Value)") || strings.Contains(plan.Groups, "otel_metrics_") {
		t.Fatal("metric presence scope or grain changed")
	}
	rows := `{"service":"metrics-only","signal":"metrics","count":"18446744073709551615","first_ms":"1000","last_ms":"2999"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(b), "WITH captured") {
			fmt.Fprintln(w, rows)
		}
	}))
	defer server.Close()
	source := Source{Endpoint: server.URL}
	result, err := source.DiscoverServices(context.Background(), q)
	if err != nil || len(result.Signals) != 1 || len(result.Dependencies) != 0 || result.Signals[0].Signal != "metrics" || result.Signals[0].Count != "18446744073709551615" {
		t.Fatal(result, err)
	}
	original := rows
	for _, bad := range []string{strings.Replace(original, "metrics\"", "unknown\"", 1), strings.Replace(original, "18446744073709551615", "0", 1), strings.Replace(original, "2999", "3000", 1), strings.Replace(original, "1000", "999", 1), original + "\n" + original} {
		rows = bad
		if _, e := source.DiscoverServices(context.Background(), q); e == nil {
			t.Fatal("invalid metric service row accepted", bad)
		}
	}
	rows = strings.Repeat(original+"\n", 1001)
	if _, e := source.DiscoverServices(context.Background(), q); e == nil {
		t.Fatal("partial over-budget discovery published")
	}
}

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

func TestSourcePagesDoNotDefineTotals(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		sql := string(b)
		if r.URL.Query().Get("readonly") != "1" || r.URL.Query().Get("max_bytes_to_read") == "" {
			t.Error("unbounded source query")
		}
		switch {
		case strings.Contains(sql, "toString(count()) AS count") && !strings.Contains(sql, "GROUP BY"):
			fmt.Fprintln(w, `{"count":"100000"}`)
		case strings.Contains(sql, "SELECT DISTINCT"):
			fmt.Fprintln(w, `{"field":"attribute:empty"}`)
		case strings.Contains(sql, "GROUP BY"):
			fmt.Fprintln(w, `{"g0":null,"count":"50000"}`)
			fmt.Fprintln(w, `{"g0":"","count":"50000"}`)
		default:
			for range 3 {
				fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"0","TraceId":"a"}`)
			}
		}
	}))
	defer server.Close()
	q := basic()
	q.GroupBy = []Field{{"attribute", "empty"}}
	source := Source{Endpoint: server.URL}
	a, e := source.Fetch(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	if a.Count != "100000" || len(a.Rows) != 2 || !a.HasMore || len(a.Groups) != 2 {
		t.Fatal("partial row page became population", a)
	}
	b, e := source.Fetch(context.Background(), q)
	if e != nil || a.Basis != b.Basis {
		t.Fatal("receipt time changes evidence basis", e)
	}
}
func TestSourceRejectsGroupOverflowAndFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if strings.Contains(string(b), "GROUP BY") {
			for range 1001 {
				fmt.Fprintln(w, `{"g0":"x","count":"1"}`)
			}
		} else if strings.Contains(string(b), "count()") {
			fmt.Fprintln(w, `{"count":"1001"}`)
		}
	}))
	q := basic()
	q.GroupBy = []Field{{"column", "service"}}
	s := Source{Endpoint: server.URL}
	if _, e := s.Fetch(context.Background(), q); e == nil {
		t.Fatal("partial groups accepted")
	}
	server.Close()
	if r, e := s.Fetch(context.Background(), q); e == nil || r != nil {
		t.Fatal("source failure became healthy zero")
	}
}
func TestChangedReadsDoNotClaimNoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		switch {
		case strings.Contains(string(b), "count()"):
			fmt.Fprintln(w, `{"count":"0"}`)
		case strings.Contains(string(b), "DISTINCT"):
		default:
			fmt.Fprintln(w, `{"timestamp_ns":"1","duration_ns":"0"}`)
		}
	}))
	defer server.Close()
	r, e := (Source{Endpoint: server.URL}).Fetch(context.Background(), basic())
	if e != nil || r.State != "changed_between_reads" || r.Count != "0" || len(r.Rows) != 1 {
		t.Fatal("source disagreement hidden", r, e)
	}
}

func TestHTTPParameterEscaping(t *testing.T) {
	value := "a\\b\\n\t\n雪"
	expected := `a\\b\\n\x09\x0a雪`
	if got := encodeParameter(value); got != expected {
		t.Fatalf("%q != %q", got, expected)
	}
}

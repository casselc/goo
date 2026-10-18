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

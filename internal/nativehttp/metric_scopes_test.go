package nativehttp_test

import (
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/measurements"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricScopeBoundary432(t *testing.T) {
	reads := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		w.Header().Set("X-ClickHouse-Exception-Code", "497")
		w.WriteHeader(403)
		w.Write([]byte("private secret SELECT source address"))
	}))
	defer upstream.Close()
	mux := http.NewServeMux()
	nativehttp.RegisterMeasurements(mux, measurements.Source{Endpoint: upstream.URL})
	for _, body := range []string{`{}`, `null`, `[]`, `{"fromMs":1,"toMs":2,"environment":"local"}`, `{"fromMs":1,"toMs":2} {}`, `{"fromMs":1,"toMs":7200002}`, `{"fromMs":1,"toMs":2,"service":"known"}`} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/metric-scopes", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatalf("accepted %s: %d", body, w.Code)
		}
	}
	if reads != 0 {
		t.Fatal("invalid identity bootstrap read source")
	}
	policy, ok := nativehttp.PolicyFor("POST", "/metric-scopes")
	if !ok || policy.BodyLimit != 4096 || policy.Timeout.Seconds() != 30 {
		t.Fatal("missing policy", policy)
	}
	budget := queryhttp.NewBudget(mux, 0, func(r *http.Request) bool {
		p, ok := nativehttp.PolicyFor(r.Method, r.URL.Path)
		return ok && p.Effect != queryhttp.Pure
	})
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, httptest.NewRequest("POST", "/metric-scopes", strings.NewReader(`{"fromMs":1,"toMs":2}`)))
	if w.Code != 429 || reads != 0 {
		t.Fatal("source admission bypass", w.Code, reads)
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/metric-scopes", strings.NewReader(`{"fromMs":1,"toMs":2}`)))
	if w.Code != 502 || reads != 4 || strings.Contains(w.Body.String(), "secret") {
		t.Fatal("denial became an empty catalog or raw failure", w.Code, reads, w.Body.String())
	}
}

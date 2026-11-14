package recordhttp_test

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/recordhttp"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSourceCapabilitiesBoundary431(t *testing.T) {
	reads := 0
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads++
		w.Header().Set("X-ClickHouse-Exception-Code", "497")
		w.WriteHeader(403)
		w.Write([]byte("private secret SQL address"))
	}))
	defer source.Close()
	mux := http.NewServeMux()
	recordhttp.Register(mux, exploration.Source{Endpoint: source.URL}, &exploration.ResultSnapshots{})
	for _, body := range []string{`{}`, `null`, `[]`, `{"fromMs":1,"toMs":2,"environment":"local"}`, `{"fromMs":1,"toMs":2} {}`, `{"fromMs":1,"toMs":7200002}`, `{"fromMs":1,"toMs":2,"table":"fake"}`} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/source-capabilities", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatalf("accepted %s: %d", body, w.Code)
		}
	}
	if reads != 0 {
		t.Fatal("invalid request reached source")
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/source-capabilities", strings.NewReader(`{"fromMs":1,"toMs":2}`)))
	var result exploration.SourceCapabilitiesResult
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || len(result.Families) != 6 || reads != 6 {
		t.Fatal(w.Code, w.Body.String(), reads)
	}
	for _, family := range result.Families {
		if family.State != "unreadable" || family.Observation.State != "unknown" {
			t.Fatal(family)
		}
	}
	if strings.Contains(w.Body.String(), "secret") {
		t.Fatal("raw source error leaked")
	}
	policy, ok := recordhttp.PolicyFor("POST", "/source-capabilities")
	if !ok || policy.BodyLimit != 4096 || policy.Timeout.Seconds() != 30 {
		t.Fatal("missing bounded policy", policy)
	}
}

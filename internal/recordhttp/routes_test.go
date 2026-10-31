package recordhttp_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/internal/recordhttp"
)

func TestAdmissionDistinguishesPureComparisonFromRecordReads(t *testing.T) {
	mux := http.NewServeMux()
	recordhttp.Register(mux, exploration.Source{}, &exploration.ResultSnapshots{})
	budget := queryhttp.NewBudget(mux, 0, func(r *http.Request) bool {
		policy, ok := recordhttp.PolicyFor(r.Method, r.URL.Path)
		return ok && policy.Effect != queryhttp.Pure
	})
	for _, path := range []string{"/environment-scopes", "/environment-scopes/records", "/inventory", "/discovery", "/explore", "/explore/page", "/record-aggregates", "/explore/counted-values", "/trace-search"} {
		t.Run(path, func(t *testing.T) {
			w := httptest.NewRecorder()
			budget.ServeHTTP(w, httptest.NewRequest("POST", path, strings.NewReader(`{}`)))
			if w.Code != 429 {
				t.Fatalf("source route bypassed full admission: %d %s", w.Code, w.Body.String())
			}
		})
	}
	cohort := `{"question":{"signal":"spans","environment":"local","fromMs":1,"toMs":60001,"limit":100,"groupBy":[{"kind":"attribute","key":"region"}]},"basis":"client-capture","count":"2","state":"observed","groups":[{"g0":"east","count":"2"}]}`
	r := httptest.NewRequest("POST", "/compare", strings.NewReader(`{"baseline":`+cohort+`,"selection":`+cohort+`}`))
	r.Header.Set("X-Goo-Query-ID", "irrelevant-to-pure-computation")
	w := httptest.NewRecorder()
	budget.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("pure comparison inherited query admission: %d %s", w.Code, w.Body.String())
	}
	for _, route := range []struct{ method, path string }{{"GET", "/explore"}, {"POST", "/explore/other"}, {"POST", "/captures"}, {"POST", "/tail"}} {
		if _, ok := recordhttp.PolicyFor(route.method, route.path); ok {
			t.Fatalf("record policy claimed an unrelated effect: %+v", route)
		}
	}
}

func TestRegistrationsKeepSuppliedSnapshotBanksIndependent(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		if strings.Contains(string(body), "count()") {
			fmt.Fprintln(w, `{"count":"0"}`)
		}
	}))
	defer source.Close()
	banks := [2]*exploration.ResultSnapshots{{}, {}}
	for i, bank := range banks {
		mux := http.NewServeMux()
		recordhttp.Register(mux, exploration.Source{Endpoint: source.URL, Client: source.Client()}, bank)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/explore", strings.NewReader(`{"signal":"spans","environment":"prod","fromMs":1,"toMs":60001,"limit":100}`)))
		var result exploration.Result
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &result) != nil || result.ResultToken == "" {
			t.Fatalf("missing independently held result: %d %s", w.Code, w.Body.String())
		}
		if _, err := bank.Load(result.ResultToken); err != nil {
			t.Fatal(err)
		}
		if _, err := banks[1-i].Load(result.ResultToken); err == nil {
			t.Fatal("result token leaked across supplied snapshot banks")
		}
	}
}

type recordPageTransport struct{ reads int }

func (s *recordPageTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	s.reads++
	body, _ := io.ReadAll(r.Body)
	out := ""
	if strings.Contains(string(body), "count()") {
		out = `{"count":"0"}`
	}
	return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(out))}, nil
}
func TestRecordPageRouteNarrowResponseAndIndependentSnapshot(t *testing.T) {
	transport := &recordPageTransport{}
	bank := &exploration.ResultSnapshots{}
	mux := http.NewServeMux()
	recordhttp.Register(mux, exploration.Source{Endpoint: "http://recording.invalid", Client: &http.Client{Transport: transport}}, bank)
	body := `{"signal":"spans","environment":"prod","fromMs":1,"toMs":60001,"limit":100,"offset":100,"groupBy":[{"kind":"column","key":"service"}],"summary":"span-duration","bucketMs":10000}`
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/explore/page", strings.NewReader(body)))
	var wire map[string]json.RawMessage
	var page exploration.RecordPageResult
	if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &wire) != nil || json.Unmarshal(w.Body.Bytes(), &page) != nil || transport.reads != 2 {
		t.Fatalf("page read failed %d %s reads%d", w.Code, w.Body.String(), transport.reads)
	}
	if len(wire) != 9 || page.ResultToken == "" || page.Question.Offset != 100 || page.Question.Summary != "span-duration" {
		t.Fatal("page contract lost", w.Body.String())
	}
	held, err := bank.Load(page.ResultToken)
	if err != nil || held.Explore.Basis != page.Basis || held.Explore.Question.BucketMS != 10000 {
		t.Fatal("exact page not held", err)
	}
	invalid := httptest.NewRecorder()
	mux.ServeHTTP(invalid, httptest.NewRequest("POST", "/explore/page", strings.NewReader(strings.Replace(body, "span-duration", "invalid", 1))))
	if invalid.Code != 400 || transport.reads != 2 {
		t.Fatal("invalid original question reached source")
	}
}

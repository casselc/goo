package recordhttp_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/recordhttp"
)

func TestAllEnvironments437UnsupportedRoutesRefuseBeforeSource(t *testing.T) {
	const all = `{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"all"}}`
	var reads atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reads.Add(1) }))
	defer server.Close()
	mux := http.NewServeMux()
	recordhttp.Register(mux, exploration.Source{Endpoint: server.URL, Client: server.Client()}, &exploration.ResultSnapshots{})
	for _, tc := range []struct{ path, body string }{
		{"/explore", `{"signal":"spans","environmentSelector":` + all + `,"fromMs":1000,"toMs":2000,"limit":100}`},
		{"/explore/page", `{"signal":"spans","environmentSelector":` + all + `,"fromMs":1000,"toMs":2000,"limit":100}`},
		{"/record-aggregates", `{"signal":"spans","environmentSelector":` + all + `,"fromMs":1000,"toMs":2000,"limit":100}`},
		{"/inventory", `{"environmentSelector":` + all + `,"fromMs":1000,"toMs":2000}`},
		{"/discovery", `{"environmentSelector":` + all + `,"fromMs":1000,"toMs":2000}`},
		{"/trace-search", `{"version":"goo/trace-search/v1","environmentSelector":` + all + `,"fromMs":1000,"toMs":2000,"conditions":[{"field":{"kind":"column","key":"service"},"op":"eq","value":"worker"}],"limit":10}`},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body)))
			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "all-environments") {
				t.Fatal("valid unsupported-family input was not refused by its All gate", response.Code, response.Body.String())
			}
		})
	}
	if reads.Load() != 0 {
		t.Fatal("unsupported All routes performed source work", reads.Load())
	}
}

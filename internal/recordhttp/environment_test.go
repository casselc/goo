package recordhttp_test

import (
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/recordhttp"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdoptionValidationBeforeSourceAndNoLegacyCoercion(t *testing.T) {
	reads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reads++ }))
	defer server.Close()
	mux := http.NewServeMux()
	recordhttp.Register(mux, exploration.Source{Endpoint: server.URL}, &exploration.ResultSnapshots{})
	for _, body := range []string{`{"signal":"logs","fromMs":1000,"toMs":2000,"limit":100}`, `{"signal":"logs","fromMs":1000,"toMs":2000,"scope":{"mode":"any"},"limit":100}`, `{"signal":"logs","environment":"prod","fromMs":1000,"toMs":2000,"scope":{"mode":"unreported"},"limit":100}`} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/environment-scopes/records", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatal(w.Code, w.Body.String())
		}
	}
	for _, body := range []string{
		`{"signal":"logs","environment":"prod","fromMs":1000,"toMs":2000,"limit":100,"selector":{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported"}}}`,
		`{"signal":"logs","environment":"prod","fromMs":1000,"toMs":2000,"limit":100,"environmentScope":{"mode":"unreported"}}`,
	} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/explore", strings.NewReader(body)))
		if w.Code != 400 {
			t.Fatal("unsupported selector accepted by old Query", w.Code, w.Body.String())
		}
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("POST", "/explore", strings.NewReader(`{"signal":"logs","environment":"","fromMs":1000,"toMs":2000,"limit":100}`)))
	if w.Code != 400 {
		t.Fatal("legacy empty environment changed")
	}
	if reads != 0 {
		t.Fatal("invalid scope performed source work", reads)
	}
}

package main

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
)

func TestResponseSelectorDoesNotExtendLegacyInputs249(t *testing.T) {
	var reads atomic.Int32
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reads.Add(1); http.Error(w, "unexpected source", 500) }))
	defer source.Close()
	dir := t.TempDir()
	mux := http.NewServeMux()
	s := exploration.Source{Endpoint: source.URL, Client: source.Client()}
	registerTail(mux, s, exploration.NewTailMembershipSnapshots(nil))
	capturehttp.Register(mux, s, dir, &exploration.ResultSnapshots{})
	registerActivityViews(mux, dir)
	const selector = `{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported"}}`
	cases := []struct{ path, body string }{
		{"/activity-views", `{"version":1,"title":"Must refuse selector","selector":{"signal":"logs","environment":"","environmentSelector":` + selector + `},"columns":["utc","service","summary"]}`},
		{"/tail", `{"signal":"logs","environment":"prod","selector":` + selector + `}`},
		{"/captures", `{"title":"Must refuse","question":{"kind":"explore","explore":{"signal":"logs","environment":"prod","fromMs":1000,"toMs":2000,"limit":100,"selector":` + selector + `}}}`},
		{"/activity-views", `{"version":1,"title":"Must refuse","selector":{"signal":"logs","environment":"prod","selector":` + selector + `},"columns":["utc","service","summary"]}`},
	}
	for _, tc := range cases {
		if !json.Valid([]byte(tc.body)) {
			t.Fatal("bad fixture", tc.body)
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", tc.path, strings.NewReader(tc.body)))
		if w.Code != 400 {
			t.Fatal("response selector accepted as legacy input", tc.path, w.Code, w.Body.String())
		}
	}
	if reads.Load() != 0 {
		t.Fatal("unsupported continuation read source")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 0 {
		t.Fatal("unsupported continuation published files", entries, err)
	}
}

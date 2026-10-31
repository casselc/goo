package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestAuthoredViewTailEnvironmentHTTP(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		if strings.Contains(string(raw), "toString(count()) AS count") {
			fmt.Fprintln(w, `{"count":"0"}`)
		}
	}))
	defer source.Close()
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerWorkbench(mux, dir, retentionConnection{}, &exploration.ResultSnapshots{}, nil)
	for _, mode := range []string{"named", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			scope := identity.EnvironmentScope{Mode: mode}
			if mode == "named" {
				scope.Value = "prod"
			}
			selection, _ := identity.EffectiveEnvironment(scope)
			wire, _ := selection.Selector()
			q := exploration.Query{Signal: "spans", EnvironmentSelector: &wire, FromMS: 1000, ToMS: 2000, Limit: 100}
			capture, err := (exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}).Create(context.Background(), exploration.Source{Endpoint: source.URL}, "Baseline", exploration.CaptureQuestion{Kind: "explore", Explore: &q}, "")
			if err != nil {
				t.Fatal(err)
			}
			q.FromMS = 2000
			q.ToMS = 3000
			view := workbench.ExplorationView{Version: 1, Title: "Live", Queries: []exploration.Query{q}, Layout: "records", Columns: []string{"utc", "summary"}, Tail: &workbench.ExplorationTail{Signal: "spans", Baseline: capture.ID}}
			post := func() *httptest.ResponseRecorder {
				raw, _ := json.Marshal(view)
				w := httptest.NewRecorder()
				mux.ServeHTTP(w, httptest.NewRequest("POST", "/exploration-views", strings.NewReader(string(raw))))
				return w
			}
			if w := post(); w.Code != 201 {
				t.Fatal("matching typed selector refused", w.Code, w.Body.String())
			}
			view.Queries[0].EnvironmentSelector = nil
			view.Queries[0].Environment = "prod"
			if w := post(); w.Code != 400 {
				t.Fatal("different environment policy accepted", w.Code, w.Body.String())
			}
			view.Queries[0] = q
			view.Tail.Baseline = strings.Repeat("f", 64)
			if w := post(); w.Code != 400 {
				t.Fatal("missing baseline accepted", w.Code)
			}
		})
	}
}

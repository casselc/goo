package capturehttp_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureSearchHTTPContract(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "exploration")
	if e := os.Mkdir(root, 0700); e != nil {
		t.Fatal(e)
	}
	wire, e := os.ReadFile("../../testdata/storage/exploration-capture.json")
	if e != nil {
		t.Fatal(e)
	}
	var c exploration.SavedCapture
	if e = json.Unmarshal(wire, &c); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(root, c.ID+".json"), wire, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = (exploration.CaptureStore{Dir: root}).RebuildCatalog(t.Context()); e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	capturehttp.Register(mux, exploration.Source{Endpoint: "http://127.0.0.1:1"}, dir, &exploration.ResultSnapshots{})
	for _, tt := range []struct {
		path    string
		code    int
		matches int
	}{{"/captures/search", 200, 1}, {"/captures/search?title=API&signal=spans&environment=local&kind=explore&order=oldest", 200, 1}, {"/captures/search?environment=", 200, 0}, {"/captures/search?cursor=" + c.ID, 400, 0}, {"/captures/search?kind=wat", 400, 0}, {"/captures/search?order=newest&order=oldest", 400, 0}, {"/captures/search?unknown=x", 400, 0}, {"/captures/search?title=%zz", 400, 0}} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", tt.path, nil))
		if w.Code != tt.code {
			t.Fatal(tt.path, w.Code, w.Body.String())
		}
		if w.Code == 200 {
			var p exploration.CaptureSearchPage
			if e = json.Unmarshal(w.Body.Bytes(), &p); e != nil || p.Matched != tt.matches {
				t.Fatal(p, e)
			}
		}
	}
	metadata := filepath.Join(root, ".catalog-v2", c.ID+".json")
	backup, e := os.ReadFile(metadata)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.Remove(metadata); e != nil {
		t.Fatal(e)
	}
	unavailable := httptest.NewRecorder()
	mux.ServeHTTP(unavailable, httptest.NewRequest("GET", "/captures/search", nil))
	if unavailable.Code != 503 {
		t.Fatal("missing index not refused", unavailable.Code)
	}
	reopened := httptest.NewRecorder()
	mux.ServeHTTP(reopened, httptest.NewRequest("GET", "/captures/"+c.ID, nil))
	if reopened.Code != 200 {
		t.Fatal("metadata absence blocked source-free reopen", reopened.Code, reopened.Body.String())
	}
	if e = os.WriteFile(metadata, backup, 0600); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(root, c.ID+".json"), []byte("bad"), 0600); e != nil {
		t.Fatal(e)
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/captures/search?title=excluded", nil))
	if w.Code != 500 {
		t.Fatal(w.Code)
	}
}

// Crosses the actual admission, immutable writer, index and recovery boundaries.
// The browser does not HTML-escape '<'; Go's stored receipt does.
func TestAcceptedLargeQuestionRemainsDiscoverableAndRepairable(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sql, _ := io.ReadAll(r.Body)
		if strings.Contains(string(sql), "toString(count()) AS count") {
			fmt.Fprintln(w, `{"count":"0"}`)
		}
	}))
	defer source.Close()
	dir := t.TempDir()
	mux := http.NewServeMux()
	capturehttp.Register(mux, exploration.Source{Endpoint: source.URL}, dir, &exploration.ResultSnapshots{})
	post := func(body []byte) exploration.CapturePage {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("POST", "/captures", bytes.NewReader(body)))
		if w.Code != 201 {
			t.Fatal(w.Code, w.Body.String())
		}
		var page exploration.CapturePage
		if e := json.Unmarshal(w.Body.Bytes(), &page); e != nil {
			t.Fatal(e)
		}
		return page
	}
	small := post([]byte(`{"title":"Earlier evidence","question":{"kind":"explore","explore":{"signal":"logs","environment":"prod","fromMs":1000,"toMs":61000,"limit":100}}}`))
	leaves := make([]exploration.Predicate, 31)
	for i := range leaves {
		leaves[i] = exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: strings.Repeat("<", 340)}
	}
	q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Where: &exploration.Predicate{All: leaves}}
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(false)
	if e := encoder.Encode(map[string]any{"title": "Large <雪> question", "question": exploration.CaptureQuestion{Kind: "explore", Explore: &q}}); e != nil {
		t.Fatal(e)
	}
	if body.Len() > 16<<10 {
		t.Fatal("fixture does not reach accepted HTTP boundary", body.Len())
	}
	large := post(body.Bytes())
	receipt, _ := json.Marshal(large.Receipt)
	if len(receipt) <= 64<<10 {
		t.Fatal("fixture no longer reproduces old index overflow", len(receipt))
	}
	root := filepath.Join(dir, "exploration")
	store := exploration.CaptureStore{Dir: root}
	check := func() {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", "/captures/search?environment=prod&signal=logs", nil))
		var page exploration.CaptureSearchPage
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &page) != nil || page.Total != 2 || page.Matched != 2 || page.Version != "goo/capture-search/v2" {
			t.Fatal(w.Code, w.Body.String())
		}
		if bytes.Contains(w.Body.Bytes(), []byte(`"question"`)) {
			t.Fatal("search leaked full question")
		}
		for _, id := range []string{small.Receipt.ID, large.Receipt.ID} {
			found := false
			for _, s := range page.Items {
				found = found || s.ID == id
			}
			if !found {
				t.Fatal("accepted capture omitted", id)
			}
		}
	}
	check()
	payload := filepath.Join(root, large.Receipt.ID+".json")
	original, e := os.ReadFile(payload)
	if e != nil {
		t.Fatal(e)
	}
	// Simulate a pre-v2 store, including the receipt that v1 could not index.
	if e = os.RemoveAll(filepath.Join(root, ".catalog-v2")); e != nil {
		t.Fatal(e)
	}
	if e = os.Mkdir(filepath.Join(root, ".catalog-v1"), 0700); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(root, ".catalog-v1", "legacy"), []byte("old derived data"), 0600); e != nil {
		t.Fatal(e)
	}
	unavailable := httptest.NewRecorder()
	mux.ServeHTTP(unavailable, httptest.NewRequest("GET", "/captures/search", nil))
	if unavailable.Code != 503 {
		t.Fatal("migration was not explicit", unavailable.Code)
	}
	reopened := httptest.NewRecorder()
	mux.ServeHTTP(reopened, httptest.NewRequest("GET", "/captures/"+large.Receipt.ID, nil))
	if reopened.Code != 200 {
		t.Fatal("migration broke exact open", reopened.Code)
	}
	rebuilt, e := store.RebuildCatalog(t.Context())
	if e != nil || rebuilt.Indexed != 2 {
		t.Fatal(rebuilt, e)
	}
	check()
	after, e := os.ReadFile(payload)
	if e != nil || !bytes.Equal(original, after) {
		t.Fatal("repair altered evidence", e)
	}
	legacy, e := os.ReadFile(filepath.Join(root, ".catalog-v1", "legacy"))
	if e != nil || string(legacy) != "old derived data" {
		t.Fatal("migration touched old index", e)
	}
	list := httptest.NewRecorder()
	mux.ServeHTTP(list, httptest.NewRequest("GET", "/captures", nil))
	if list.Code != 200 {
		t.Fatal("legacy list changed", list.Code)
	}
}

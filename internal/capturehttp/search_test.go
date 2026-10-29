package capturehttp_test

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/capturehttp"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	metadata := filepath.Join(root, ".catalog-v1", c.ID+".json")
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

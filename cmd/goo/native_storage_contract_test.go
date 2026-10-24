package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/casselc/goo/measurements"
)

func TestNativeYAMLStorageContract(t *testing.T) {
	t.Chdir("../..")
	wire, e := os.ReadFile("testdata/storage/native-measurement.yaml")
	if e != nil {
		t.Fatal(e)
	}
	want, e := os.ReadFile("testdata/storage/native-measurement-response.json")
	if e != nil {
		t.Fatal(e)
	}
	var ids map[string]string
	b, e := os.ReadFile("testdata/storage/identities.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &ids); e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(wire))
	if e != nil || measurements.Hash(d) != ids["nativeMeasurementRevision"] {
		t.Fatal("native identity changed", e)
	}
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerMeasurements(mux, measurements.Source{Client: http.DefaultClient}, dir)
	payload, _ := json.Marshal(map[string]string{"yaml": string(wire)})
	request := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		return w
	}
	file := filepath.Join(dir, "measurement-packages", ids["nativeMeasurementRevision"]+".yaml")
	for i := 0; i < 2; i++ {
		r := request("POST", "/measurement-packages", string(payload))
		if r.Code != 201 || !bytes.Equal(r.Body.Bytes(), want) || r.Header().Get("Content-Type") != "application/json" || r.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("native response changed", r.Code, r.Body.String())
		}
		stored, e := os.ReadFile(file)
		if e != nil || !bytes.Equal(stored, wire) {
			t.Fatal("native YAML bytes changed", e)
		}
	}
	reopened := http.NewServeMux()
	registerMeasurements(reopened, measurements.Source{Client: http.DefaultClient}, dir)
	w := httptest.NewRecorder()
	reopened.ServeHTTP(w, httptest.NewRequest("GET", "/measurement-packages", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), ids["nativeMeasurementRevision"]) {
		t.Fatal("fresh registry reopen", w.Code, w.Body.String())
	}
	for _, bad := range []string{`{"yaml":"invalid"}`, string(payload) + ` {}`, `{"yaml":"invalid","foreign":1}`} {
		if r := request("POST", "/measurement-packages", bad); r.Code != 400 {
			t.Fatal("invalid publication accepted", r.Code)
		}
	}
	// A corrupt immutable winner is refused without repair, rather than reporting
	// publication success for a package the registry cannot reopen.
	if e = os.WriteFile(file, []byte("corrupt YAML"), 0600); e != nil {
		t.Fatal(e)
	}
	r := request("POST", "/measurement-packages", string(payload))
	if r.Code != 400 || !strings.Contains(r.Body.String(), "existing measurement package") {
		t.Fatal("corrupt duplicate reported success", r.Code, r.Body.String())
	}
	stored, e := os.ReadFile(file)
	if e != nil || string(stored) != "corrupt YAML" {
		t.Fatal("winner overwritten", e)
	}
	if r = request("GET", "/measurement-packages", ""); r.Code != 400 {
		t.Fatal("corrupt registry hidden", r.Code)
	}
	entries, e := os.ReadDir(filepath.Dir(file))
	if e != nil || len(entries) != 1 {
		t.Fatal("scratch leaked", e)
	}
}

func TestNativeYAMLExistingWinnerValidation(t *testing.T) {
	t.Chdir("../..")
	wire, e := os.ReadFile("testdata/storage/native-measurement.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(wire))
	if e != nil {
		t.Fatal(e)
	}
	other := strings.Replace(string(wire), "title: API utilization", "title: Different valid package", 1)
	padded := "#" + strings.Repeat("x", 32768-len(wire)-2) + "\n" + string(wire)
	if changed, e := measurements.Read(other); e != nil || measurements.Hash(changed) == measurements.Hash(d) {
		t.Fatal("wrong-identity test fixture invalid", e)
	}
	for _, tc := range []struct {
		name, stored string
		status       int
		directory    bool
	}{
		{"valid noncanonical winner", "# preserve original representation\n" + string(wire), 201, false},
		{"exact limit valid winner", padded, 201, false},
		{"wrong identity", other, 400, false},
		{"one byte over limit", padded + "\n", 400, false},
		{"empty", "", 400, false},
		{"unreadable directory", "", 400, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			packages := filepath.Join(dir, "measurement-packages")
			if e := os.MkdirAll(packages, 0700); e != nil {
				t.Fatal(e)
			}
			file := filepath.Join(packages, measurements.Hash(d)+".yaml")
			if tc.directory {
				e = os.Mkdir(file, 0700)
			} else {
				e = os.WriteFile(file, []byte(tc.stored), 0600)
			}
			if e != nil {
				t.Fatal(e)
			}
			mux := http.NewServeMux()
			registerMeasurements(mux, measurements.Source{}, dir)
			payload, _ := json.Marshal(map[string]string{"yaml": string(wire)})
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", "/measurement-packages", bytes.NewReader(payload)))
			if w.Code != tc.status {
				t.Fatal("winner policy", w.Code, w.Body.String())
			}
			if tc.status != 201 && !strings.Contains(w.Body.String(), "existing measurement package") {
				t.Fatal("unhelpful error", w.Body.String())
			}
			if tc.directory {
				if info, e := os.Stat(file); e != nil || !info.IsDir() {
					t.Fatal("winner replaced", e)
				}
			} else if stored, e := os.ReadFile(file); e != nil || string(stored) != tc.stored {
				t.Fatal("winner changed", e)
			}
			entries, e := os.ReadDir(packages)
			if e != nil || len(entries) != 1 {
				t.Fatal("scratch leaked", e)
			}
		})
	}
}

func TestNativeYAMLConcurrentDuplicatePublication(t *testing.T) {
	t.Chdir("../..")
	wire, e := os.ReadFile("testdata/storage/native-measurement.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(wire))
	if e != nil {
		t.Fatal(e)
	}
	payload, _ := json.Marshal(map[string]string{"yaml": string(wire)})
	dir := t.TempDir()
	mux := http.NewServeMux()
	registerMeasurements(mux, measurements.Source{}, dir)
	const requests = 16
	results := make(chan *httptest.ResponseRecorder, requests)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("POST", "/measurement-packages", bytes.NewReader(payload)))
			results <- w
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	for w := range results {
		if w.Code != 201 || !strings.Contains(w.Body.String(), measurements.Hash(d)) {
			t.Fatal("concurrent valid duplicate refused", w.Code, w.Body.String())
		}
	}
	packages := filepath.Join(dir, "measurement-packages")
	stored, e := os.ReadFile(filepath.Join(packages, measurements.Hash(d)+".yaml"))
	if e != nil || !bytes.Equal(stored, wire) {
		t.Fatal("concurrent winner changed", e)
	}
	entries, e := os.ReadDir(packages)
	if e != nil || len(entries) != 1 {
		t.Fatal("scratch leaked", e)
	}
	reopened := http.NewServeMux()
	registerMeasurements(reopened, measurements.Source{}, dir)
	w := httptest.NewRecorder()
	reopened.ServeHTTP(w, httptest.NewRequest("GET", "/measurement-packages", nil))
	if w.Code != 200 || !strings.Contains(w.Body.String(), measurements.Hash(d)) {
		t.Fatal("concurrent publication not reopenable", w.Code, w.Body.String())
	}
}

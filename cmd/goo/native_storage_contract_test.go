package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
	// Existing contract: POST accepts an existing name without reading or repairing it.
	if e = os.WriteFile(file, []byte("corrupt YAML"), 0600); e != nil {
		t.Fatal(e)
	}
	r := request("POST", "/measurement-packages", string(payload))
	if r.Code != 201 || !bytes.Equal(r.Body.Bytes(), want) {
		t.Fatal("duplicate response policy changed", r.Code)
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

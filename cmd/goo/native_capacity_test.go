package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"gopkg.in/yaml.v3"
)

func TestNativeCapacityHTTPAdmissionAndExactRecovery(t *testing.T) {
	wire, e := os.ReadFile("../../config/measurements/worker-active.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(wire))
	if e != nil {
		t.Fatal(e)
	}
	root := t.TempDir()
	t.Chdir(root)
	mux := http.NewServeMux()
	registerMeasurements(mux, measurements.Source{}, root)
	request := func(method, path string, body []byte) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, bytes.NewReader(body)))
		return w
	}
	var acceptedBody []byte
	var acceptedRevision string
	for i := range 257 {
		d.Name = fmt.Sprintf("native_%d", i)
		text, e := yaml.Marshal(d)
		if e != nil {
			t.Fatal(e)
		}
		body, _ := json.Marshal(map[string]string{"yaml": string(text)})
		w := request("POST", "/measurement-packages", body)
		if i < 256 {
			if w.Code != 201 {
				t.Fatal(i, w.Code, w.Body.String())
			}
			acceptedBody = body
			acceptedRevision = measurements.Hash(d)
		} else if w.Code != 400 || !strings.Contains(w.Body.String(), "capacity") {
			t.Fatal("257th publication accepted", w.Code, w.Body.String())
		}
	}
	listed := request("GET", "/measurement-packages", nil)
	var entries []json.RawMessage
	if listed.Code != 200 || json.Unmarshal(listed.Body.Bytes(), &entries) != nil || len(entries) != 256 {
		t.Fatal("admission disabled catalog", listed.Code, len(entries))
	}
	if w := request("POST", "/measurement-packages", acceptedBody); w.Code != 201 {
		t.Fatal("retry at capacity refused", w.Code)
	}
	if w := request("GET", "/measurement-packages/revisions/"+acceptedRevision, nil); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	// Already over-limit stores preserve valid known IDs without whole-catalog reads.
	d.Name = "legacy_extra"
	text, _ := yaml.Marshal(d)
	extra := measurements.Hash(d)
	if e = os.WriteFile(filepath.Join(root, "measurement-packages", extra+".yaml"), text, 0600); e != nil {
		t.Fatal(e)
	}
	if w := request("GET", "/measurement-packages", nil); w.Code != 400 {
		t.Fatal("over-limit store truncated", w.Code)
	}
	if w := request("GET", "/measurement-packages/revisions/"+extra, nil); w.Code != 200 {
		t.Fatal("known legacy revision inaccessible", w.Code, w.Body.String())
	}
	for _, path := range []string{"/measurement-packages/revisions/bad", "/measurement-packages/revisions/" + extra + "?foo=1"} {
		if w := request("GET", path, nil); w.Code != 400 {
			t.Fatal(path, w.Code)
		}
	}
	if w := request("GET", "/measurement-packages/revisions/"+strings.Repeat("0", 64), nil); w.Code != 404 {
		t.Fatal(w.Code)
	}
}

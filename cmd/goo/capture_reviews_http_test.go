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

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
)

func TestCaptureReviewsHTTPSourceFreeContracts(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile("../../testdata/storage/exploration-capture.json")
	if err != nil {
		t.Fatal(err)
	}
	var capture struct {
		ID string `json:"id"`
	}
	if err = json.Unmarshal(raw, &capture); err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(filepath.Join(dir, "exploration"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "exploration", capture.ID+".json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	registerCaptureReviews(mux, dir)
	var observations []apiObservation
	call := func(method, path, pattern, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
		if w.Code != want {
			t.Fatalf("%s %s: got %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		if w.Header().Get("Cache-Control") != "no-store" {
			t.Fatal("review response may be cached", method, path)
		}
		if want < 400 {
			observations = append(observations, apiObservation{Method: method, Path: pattern, Status: w.Code, Request: json.RawMessage(body), Response: append(json.RawMessage(nil), bytes.TrimSpace(w.Body.Bytes())...)})
		}
		return w
	}
	encode := func(v any) string {
		t.Helper()
		b, e := json.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return string(b)
	}
	id := func(w *httptest.ResponseRecorder) string {
		t.Helper()
		var v struct {
			ID string `json:"id"`
		}
		if e := json.Unmarshal(w.Body.Bytes(), &v); e != nil || v.ID == "" {
			t.Fatal(e, w.Body.String())
		}
		return v.ID
	}
	recordPath := "/record-reviews/records/" + capture.ID + "/0"
	member := call("GET", recordPath, "/record-reviews/records/{captureId}/{index}", "", 200)
	if !bytes.Contains(member.Body.Bytes(), []byte("18446744073709551615")) {
		t.Fatal("resolver changed raw numeric token")
	}
	taskBody := `{"requestID":"` + strings.Repeat("a", 64) + `","version":1,"name":"timeout","title":"Timeout review","labels":[{"key":"timeout","title":"Timeout"},{"key":"not-timeout","title":"Not timeout"}]}`
	taskID := id(call("POST", "/record-reviews/tasks", "/record-reviews/tasks", taskBody, 201))
	if id(call("POST", "/record-reviews/tasks", "/record-reviews/tasks", taskBody, 201)) != taskID {
		t.Fatal("lost acknowledgment changed task")
	}
	call("POST", "/record-reviews/tasks", "", strings.Replace(taskBody, "Timeout review", "Changed", 1), 409)
	call("GET", "/record-reviews/tasks/"+taskID, "/record-reviews/tasks/{id}", "", 200)
	decision := map[string]any{"requestID": strings.Repeat("b", 64), "taskID": taskID, "target": exploration.CaptureRecordRef{CaptureID: capture.ID, Index: 0}, "parents": []string{}, "judgment": map[string]string{"kind": "label", "value": "timeout"}, "review": "unreviewed", "state": "active", "note": "Observed by a local user"}
	first := id(call("POST", "/record-reviews/decisions", "/record-reviews/decisions", encode(decision), 201))
	if id(call("POST", "/record-reviews/decisions", "/record-reviews/decisions", encode(decision), 201)) != first {
		t.Fatal("decision retry forked")
	}
	manifest := map[string]any{"requestID": strings.Repeat("c", 64), "taskID": taskID, "selected": []string{first}, "knownHeads": []string{first}, "historical": true}
	call("POST", "/record-reviews/manifests", "", encode(manifest), 400)
	decision["requestID"] = strings.Repeat("d", 64)
	decision["review"] = "reviewed"
	decision["parents"] = []string{first}
	reviewed := id(call("POST", "/record-reviews/decisions", "/record-reviews/decisions", encode(decision), 201))
	call("GET", "/record-reviews/decisions/"+reviewed, "/record-reviews/decisions/{id}", "", 200)
	manifest["requestID"] = strings.Repeat("e", 64)
	manifest["selected"] = []string{reviewed}
	manifest["knownHeads"] = []string{reviewed}
	manifestID := id(call("POST", "/record-reviews/manifests", "/record-reviews/manifests", encode(manifest), 201))
	call("GET", "/record-reviews/manifests/"+manifestID, "/record-reviews/manifests/{id}", "", 200)
	exportPath := "/record-reviews/manifests/" + manifestID + "/export"
	exported := call("GET", exportPath, "/record-reviews/manifests/{id}/export", "", 200)
	if exported.Header().Get("Content-Type") != "application/x-ndjson" || !bytes.Contains(exported.Body.Bytes(), []byte("1700000000000000123")) || !bytes.Contains(exported.Body.Bytes(), []byte("18446744073709551615")) {
		t.Fatal("export lost type or exact numeric tokens")
	}
	decision["requestID"] = strings.Repeat("f", 64)
	decision["state"] = "withdrawn"
	decision["parents"] = []string{reviewed}
	call("POST", "/record-reviews/decisions", "/record-reviews/decisions", encode(decision), 201)
	if !bytes.Equal(exported.Body.Bytes(), call("GET", exportPath, "/record-reviews/manifests/{id}/export", "", 200).Body.Bytes()) {
		t.Fatal("withdrawal changed historical export")
	}
	// Strict boundaries must fail before any alternate target or authority is inferred.
	base := encode(decision)
	for _, body := range []string{
		strings.Replace(base, `"index":0`, `"index":null`, 1),
		strings.Replace(base, `,"index":0`, ``, 1),
		strings.Replace(base, `"kind":"label","value":"timeout"`, `"kind":"uncertain","value":null`, 1),
		strings.Replace(base, `"kind":"label","value":"timeout"`, `"kind":"abstain","value":""`, 1),
		strings.TrimSuffix(base, "}") + `,"foreign":true}`,
		base + ` {}`,
	} {
		call("POST", "/record-reviews/decisions", "", body, 400)
	}
	decision["requestID"] = strings.Repeat("1", 64)
	decision["parents"] = []string{strings.Repeat("0", 64)}
	call("POST", "/record-reviews/decisions", "", encode(decision), 404)
	call("GET", recordPath+"?fresh=true", "", "", 400)
	call("GET", "/record-reviews/records/bad/0", "", "", 400)
	call("GET", "/record-reviews/records/"+capture.ID+"/00", "", "", 400)
	call("GET", "/record-reviews/tasks/"+strings.Repeat("0", 64), "", "", 404)
	call("POST", "/record-reviews/tasks", "", `{"padding":"`+strings.Repeat("x", workbench.MaxBytes)+`"}`, 413)
	if file := os.Getenv("GOO_CAPTURE_REVIEW_OBSERVATIONS"); file != "" {
		b, e := json.Marshal(observations)
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(file, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
}

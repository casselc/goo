package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/workbench"
)

// These routes resolve existing evidence and publish local metadata only. They
// have no telemetry source or optional evidence-store dependency.
func registerCaptureReviews(mux *http.ServeMux, dir string) {
	store := workbench.Store{Dir: filepath.Join(dir, "workbench")}
	captures := exploration.CaptureStore{Dir: filepath.Join(dir, "exploration")}
	fail := func(w http.ResponseWriter, err error) {
		code := http.StatusBadRequest
		switch {
		case errors.Is(err, workbench.ErrCaptureReviewIntentConflict):
			code = http.StatusConflict
		case errors.Is(err, os.ErrNotExist):
			code = http.StatusNotFound
		default:
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				code = http.StatusRequestEntityTooLarge
			} else if _, ok := errors.AsType[*os.PathError](err); ok {
				code = http.StatusInternalServerError
			}
		}
		httpjson.Write(w, code, map[string]string{"error": err.Error()})
	}
	query := func(w http.ResponseWriter, r *http.Request) bool {
		if r.URL.RawQuery != "" {
			fail(w, errors.New("record review routes have no query parameters"))
			return false
		}
		return true
	}
	decode := func(w http.ResponseWriter, r *http.Request, out any, target bool) bool {
		if !query(w, r) {
			return false
		}
		var raw json.RawMessage
		if err := httpjson.Decode(w, r, &raw, workbench.MaxBytes, "expected one review object"); err != nil {
			fail(w, err)
			return false
		}
		if !identity.WellFormedJSON(raw) {
			fail(w, errors.New("invalid review JSON or Unicode"))
			return false
		}
		d := json.NewDecoder(bytes.NewReader(raw))
		d.DisallowUnknownFields()
		if err := d.Decode(out); err != nil {
			fail(w, err)
			return false
		}
		if target {
			var fields map[string]json.RawMessage
			var ref map[string]json.RawMessage
			if json.Unmarshal(raw, &fields) != nil || json.Unmarshal(fields["target"], &ref) != nil {
				fail(w, errors.New("explicit captured target required"))
				return false
			}
			var judgment map[string]json.RawMessage
			if json.Unmarshal(fields["judgment"], &judgment) == nil {
				var kind string
				json.Unmarshal(judgment["kind"], &kind)
				if _, present := judgment["value"]; present && (kind == "uncertain" || kind == "abstain") {
					fail(w, errors.New("nonlabel judgment must omit value"))
					return false
				}
			}
			index, ok := ref["index"]
			if !ok || bytes.Equal(bytes.TrimSpace(index), []byte("null")) {
				fail(w, errors.New("explicit absolute record index required"))
				return false
			}
		}
		return true
	}
	mux.HandleFunc("GET /record-reviews/records/{captureId}/{index}", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		index, err := strconv.Atoi(r.PathValue("index"))
		if err != nil || strconv.Itoa(index) != r.PathValue("index") {
			fail(w, errors.New("canonical absolute index required"))
			return
		}
		v, err := captures.ResolveRecord(exploration.CaptureRecordRef{CaptureID: r.PathValue("captureId"), Index: index})
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, 200, v)
	})
	mux.HandleFunc("POST /record-reviews/tasks", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			workbench.CaptureReviewTaskInput
			RequestID string `json:"requestID"`
		}
		if !decode(w, r, &in, false) {
			return
		}
		v, err := store.CreateCaptureReviewTaskOnce(in.RequestID, in.CaptureReviewTaskInput)
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, 201, v)
	})
	mux.HandleFunc("GET /record-reviews/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		v, err := store.LoadCaptureReviewTask(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, 200, v)
	})
	mux.HandleFunc("POST /record-reviews/decisions", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			workbench.CaptureReviewDecisionInput
			RequestID string `json:"requestID"`
		}
		if !decode(w, r, &in, true) {
			return
		}
		v, err := store.CreateCaptureReviewDecisionOnce(in.RequestID, in.CaptureReviewDecisionInput, captures.ResolveRecord)
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, 201, v)
	})
	mux.HandleFunc("GET /record-reviews/decisions/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		v, err := store.LoadCaptureReviewDecision(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, 200, v)
	})
	mux.HandleFunc("POST /record-reviews/manifests", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			workbench.CaptureReviewManifestInput
			RequestID string `json:"requestID"`
		}
		if !decode(w, r, &in, false) {
			return
		}
		v, err := store.CreateCaptureReviewManifestOnce(in.RequestID, in.CaptureReviewManifestInput, captures.ResolveRecord)
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, 201, v)
	})
	mux.HandleFunc("GET /record-reviews/manifests/{id}", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		v, err := store.LoadCaptureReviewManifest(r.PathValue("id"))
		if err != nil {
			fail(w, err)
			return
		}
		httpjson.Write(w, 200, v)
	})
	mux.HandleFunc("GET /record-reviews/manifests/{id}/export", func(w http.ResponseWriter, r *http.Request) {
		if !query(w, r) {
			return
		}
		data, err := store.ExportCaptureReviewManifest(r.PathValue("id"), captures.ResolveRecord)
		if err != nil {
			fail(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="goo-reviewed-%s.jsonl"`, r.PathValue("id")))
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Write(data)
	})
}

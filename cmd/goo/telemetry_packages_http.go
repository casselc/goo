package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/casselc/goo/exploration"
)

type telemetryPlanRequest struct {
	YAML       string                        `json:"yaml"`
	Definition *exploration.TelemetryPackage `json:"definition,omitempty"`
	FromMS     int64                         `json:"fromMs,omitempty"`
	ToMS       int64                         `json:"toMs,omitempty"`
	Annotation *exploration.AnnotationTarget `json:"annotation,omitempty"`
}

func registerTelemetryPackages(mux *http.ServeMux, dir string) {
	store := exploration.PackageStore{Dir: dir}
	send := func(w http.ResponseWriter, code int, value any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(value)
	}
	fail := func(w http.ResponseWriter, code int, e error) { send(w, code, map[string]string{"error": e.Error()}) }
	decode := func(w http.ResponseWriter, r *http.Request, v any) bool {
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
		d.DisallowUnknownFields()
		if e := d.Decode(v); e != nil {
			fail(w, 400, e)
			return false
		}
		if d.Decode(new(any)) != io.EOF {
			fail(w, 400, fmt.Errorf("require one JSON object"))
			return false
		}
		return true
	}
	mux.HandleFunc("POST /telemetry-packages/draft", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Question exploration.Query `json:"question"`
			Name     string            `json:"name"`
			Title    string            `json:"title"`
			Lookback int               `json:"lookbackMinutes"`
		}
		if !decode(w, r, &input) {
			return
		}
		p, e := exploration.DraftTelemetryPackage(input.Question, input.Name, input.Title, input.Lookback)
		if e != nil {
			fail(w, 400, e)
			return
		}
		plan, e := exploration.PlanTelemetryPackage(p, input.Question.FromMS, input.Question.ToMS, nil)
		if e != nil {
			fail(w, 400, e)
			return
		}
		send(w, 200, plan)
	})
	mux.HandleFunc("POST /telemetry-packages/plan", func(w http.ResponseWriter, r *http.Request) {
		var input telemetryPlanRequest
		if !decode(w, r, &input) {
			return
		}
		var p exploration.TelemetryPackage
		var e error
		if (input.YAML == "") == (input.Definition == nil) {
			fail(w, 400, fmt.Errorf("provide YAML or definition, not both"))
			return
		}
		if input.Definition != nil {
			p = *input.Definition
		} else {
			p, e = exploration.ReadTelemetryPackage(input.YAML)
		}
		if e != nil {
			fail(w, 400, e)
			return
		}
		if input.FromMS == 0 && input.ToMS == 0 {
			input.ToMS = time.Now().UnixMilli()
			input.FromMS = input.ToMS - int64(p.LookbackMinutes)*60000
		}
		plan, e := exploration.PlanTelemetryPackage(p, input.FromMS, input.ToMS, input.Annotation)
		if e != nil {
			fail(w, 400, e)
			return
		}
		send(w, 200, plan)
	})
	mux.HandleFunc("POST /telemetry-packages", func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			YAML string `json:"yaml"`
		}
		if !decode(w, r, &input) {
			return
		}
		p, e := exploration.ReadTelemetryPackage(input.YAML)
		if e != nil {
			fail(w, 400, e)
			return
		}
		saved, e := store.Save(p)
		if e != nil {
			fail(w, 500, e)
			return
		}
		send(w, 201, saved)
	})
	mux.HandleFunc("GET /telemetry-packages", func(w http.ResponseWriter, r *http.Request) {
		list, e := store.List()
		if e != nil {
			fail(w, 500, e)
			return
		}
		send(w, 200, list)
	})
	mux.HandleFunc("GET /telemetry-packages/{revision}", func(w http.ResponseWriter, r *http.Request) {
		p, e := store.Load(r.PathValue("revision"))
		if e != nil {
			code := 400
			if os.IsNotExist(e) {
				code = 404
			}
			fail(w, code, e)
			return
		}
		send(w, 200, p)
	})
}

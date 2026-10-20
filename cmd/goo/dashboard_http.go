package main

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/exploration"
	"io"
	"net/http"
)

func registerDashboards(mux *http.ServeMux, packages []*annotations.OperationPlan) {
	declarations := []exploration.DashboardDeclaration{}
	for _, p := range packages {
		m := ""
		if p.Manifest != nil {
			m = p.Manifest.Revision
		}
		declarations = append(declarations, exploration.DashboardDeclaration{Name: p.Definition.Name, Title: p.Definition.Title, Revision: p.Revision, ManifestRevision: m, Service: p.Definition.Resource.Service, Environment: p.Definition.Resource.Environment})
	}
	mux.HandleFunc("POST /dashboards/plan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		fail := func(e error) { w.WriteHeader(400); json.NewEncoder(w).Encode(map[string]string{"error": e.Error()}) }
		var in exploration.DashboardRequest
		d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10))
		d.DisallowUnknownFields()
		if err := d.Decode(&in); err != nil {
			fail(err)
			return
		}
		if d.Decode(new(any)) != io.EOF {
			fail(fmt.Errorf("expected one JSON object"))
			return
		}
		p, err := exploration.PlanServiceDashboard(in, declarations)
		if err != nil {
			fail(err)
			return
		}
		json.NewEncoder(w).Encode(p)
	})
}

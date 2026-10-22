package main

import (
	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"net/http"
)

func registerDashboards(mux *http.ServeMux, packages []*annotations.OperationPlan) {
	declarations := []exploration.DashboardDeclaration{}
	for _, p := range packages {
		m := ""
		if p.Manifest != nil {
			m = p.Manifest.Revision
		}
		ops := []exploration.DashboardOperation{}
		for _, op := range p.Definition.Operations {
			ops = append(ops, exploration.DashboardOperation{Name: op.Name, Role: op.Role})
		}
		declarations = append(declarations, exploration.DashboardDeclaration{Namespace: p.Definition.Resource.Namespace.Clone(), Instance: p.Definition.Resource.Instance, Scope: p.Definition.Instrumentation.Name, Version: p.Definition.Instrumentation.Version, SchemaURL: p.Definition.Instrumentation.SchemaURL, Operations: ops, Name: p.Definition.Name, Title: p.Definition.Title, Revision: p.Revision, ManifestRevision: m, Service: p.Definition.Resource.Service, Environment: p.Definition.Resource.Environment})
	}
	mux.HandleFunc("POST /dashboards/plan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		fail := func(e error) { httpjson.Write(w, 400, map[string]string{"error": e.Error()}) }
		var in exploration.DashboardRequest
		if err := httpjson.Decode(w, r, &in, 256<<10, "expected one JSON object"); err != nil {
			fail(err)
			return
		}
		p, err := exploration.PlanServiceDashboard(in, declarations)
		if err != nil {
			fail(err)
			return
		}
		httpjson.Write(w, 200, p)
	})
}

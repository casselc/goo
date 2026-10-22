package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/casselc/goo/annotations"
	"github.com/casselc/goo/prebuilt"
)

func TestOperationGETValidationWireContract(t *testing.T) {
	plan, e := annotations.LoadOperationPackage("../../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	mux := http.NewServeMux()
	registerOperationReads(mux, map[string]*annotations.OperationPlan{"checkout": plan}, []*annotations.OperationPlan{plan}, nil, prebuilt.HTTPSource{}, clickHouseConnection{Client: http.DefaultClient})
	for _, c := range []struct{ path, message string }{{"/http-package?minutes=0", "minutes must be 1..120"}, {"/http-package?minutes=5&minutes=6", "unsupported query parameter"}, {"/http-package?foreign=1", "unsupported query parameter"}, {"/operation-packages?foreign=1", "no query parameters supported"}, {"/operations?package=missing", "unknown operation package"}, {"/operations?minutes=121", "minutes must be 1..120"}, {"/operations?minutes=1&minutes=2", "unsupported query parameter"}, {"/counter-rates?foreign=1", "unsupported rate query parameter"}, {"/counter-rates?package=checkout&fromMs=1&toMs=60001", "require package and bounded fromMs/toMs"}, {"/histograms?foreign=1", "unsupported histogram query parameter"}, {"/histograms?package=missing&fromMs=60000&toMs=120000", "require package and bounded fromMs/toMs"}} {
		t.Run(c.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest("GET", c.path, nil))
			if w.Code != 400 || w.Body.String() != `{"error":"`+c.message+`"}`+"\n" || w.Header().Get("Content-Type") != "text/plain; charset=utf-8" || w.Header().Get("Cache-Control") != "no-store" || w.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Fatal("GET validation wire changed", w.Code, w.Header(), w.Body.String())
			}
		})
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/operation-packages", nil))
	if w.Code != 200 || w.Header().Get("Content-Type") != "application/json" || !strings.Contains(w.Body.String(), plan.Revision) {
		t.Fatal("declaration route changed", w.Code)
	}
}

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRawInstrument448HTTPAliasesRefuseBeforeSource(t *testing.T) {
	for _, tc := range []struct {
		name, path, body string
		request          func(*testing.T, planningWireCase) (*httptest.ResponseRecorder, int)
	}{
		{"gauge", "/metric-observations", rawGauge438HTTPBody, rawGauge438HTTPResponse},
		{"sum", "/sum-observations", rawSum441HTTPBody, rawSum441HTTPResponse},
		{"histogram", "/histogram-observations", rawHistogram447HTTPBody, rawHistogram447HTTPResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, alias := range []string{`"Scope":"overridden"`, `"schemaUrl":"overridden"`, `"Unit":"overridden"`} {
				body := strings.Replace(tc.body, `"scope":""`, `"scope":"",`+alias, 1)
				response, reads := tc.request(t, planningWireCase{Name: "exact-case-refusal", Method: http.MethodPost, Path: tc.path, Body: body, Status: http.StatusBadRequest})
				if response.Code != http.StatusBadRequest || reads != 0 {
					t.Fatalf("alias %s: status=%d source reads=%d body=%s", alias, response.Code, reads, response.Body.String())
				}
			}
			response, reads := tc.request(t, planningWireCase{Name: "observed", Method: http.MethodPost, Path: tc.path, Body: tc.body, Status: http.StatusOK})
			if response.Code != http.StatusOK || reads == 0 {
				t.Fatalf("exact tuple refused: status=%d source reads=%d body=%s", response.Code, reads, response.Body.String())
			}
		})
	}
}

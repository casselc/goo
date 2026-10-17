package prebuilt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func sourceFixture(t *testing.T) (*Expansion, HTTPSpanRow) {
	t.Helper()
	in, b, _ := packageFixture(t)
	e, err := Expand(in, b)
	if err != nil {
		t.Fatal(err)
	}
	return e, HTTPSpanRow{TimestampNS: "100000000", TraceID: strings.Repeat("1", 32), SpanID: strings.Repeat("2", 16), Resource: in.Scopes[0].Resource, Attributes: map[string]string{"http.route": "/checkout", "http.request.method": "POST", "http.response.status_code": "503"}}
}
func TestSourceDuplicatesConflictBoundsAndMissingOutcome(t *testing.T) {
	e, row := sourceFixture(t)
	p, result, duplicates, err := e.EvaluateHTTPRows([]HTTPSpanRow{row, row}, "e")
	if err != nil || duplicates != 1 || len(p.Facts) != 1 || *result[0].Measures["failures"].Count != 1 {
		t.Fatalf("dedup mismatch: %v", err)
	}
	conflict := row
	conflict.TimestampNS = "100000001"
	if _, _, _, err = e.EvaluateHTTPRows([]HTTPSpanRow{row, conflict}, "e"); err == nil {
		t.Fatal("conflict accepted")
	}
	if _, _, _, err = e.EvaluateHTTPRows(make([]HTTPSpanRow, 1001), "e"); err == nil {
		t.Fatal("partial population")
	}
	delete(row.Attributes, "http.response.status_code")
	_, result, _, err = e.EvaluateHTTPRows([]HTTPSpanRow{row}, "e")
	if err != nil {
		t.Fatal(err)
	}
	if result[0].Measures["success_ratio"].State != "unknown" {
		t.Fatal("absent outcome became healthy")
	}
	row.TraceID = "bad"
	if _, _, _, err = e.EvaluateHTTPRows([]HTTPSpanRow{row}, "e"); err == nil {
		t.Fatal("invalid trace identity")
	}
}
func TestLiveSourceUsesParametersAndRejectsOutOfWindow(t *testing.T) {
	e, row := sourceFixture(t)
	source := HTTPSource{ScopeName: "library'example", ScopeVersion: "1", SchemaURL: "https://opentelemetry.io/schemas/1.37.0"}
	sql, params, err := e.HTTPSourceQuery(source, 100, 200)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sql, source.ScopeName) || params["scope"] != source.ScopeName {
		t.Fatal("unbound value")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("param_scope") != source.ScopeName || r.URL.Query().Get("readonly") != "1" {
			t.Error("missing typed source parameters")
		}
		json.NewEncoder(w).Encode(row)
	}))
	defer server.Close()
	result, err := FetchHTTP(context.Background(), server.Client(), server.URL, "user", "pass", e, source, 100, 200)
	if err != nil || len(result.Results) != 1 {
		t.Fatalf("fetch: %v", err)
	}
	row.TimestampNS = "200000000"
	if _, err = FetchHTTP(context.Background(), server.Client(), server.URL, "user", "pass", e, source, 100, 200); err == nil {
		t.Fatal("exclusive upper bound violated")
	}
	if _, _, err = e.HTTPSourceQuery(source, 200, 100); err == nil {
		t.Fatal("invalid bounds")
	}
}

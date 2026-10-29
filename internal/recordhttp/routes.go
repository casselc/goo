package recordhttp

import (
	"context"
	"net/http"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/httpjson"
	"github.com/casselc/goo/internal/queryhttp"
)

// Register binds one configured source and the application's existing ephemeral
// snapshot bank. It does not construct a source, open files or publish captures.
func Register(mux *http.ServeMux, source exploration.Source, snapshots *exploration.ResultSnapshots) {
	mux.HandleFunc("POST /compare", func(w http.ResponseWriter, r *http.Request) {
		policy, _ := PolicyFor(r.Method, r.URL.Path)
		fail := func(err error) { httpjson.Write(w, 400, map[string]string{"error": err.Error()}) }
		var request exploration.ComparisonRequest
		if err := httpjson.Decode(w, r, &request, policy.BodyLimit, "expected one JSON object"); err != nil {
			fail(err)
			return
		}
		result, err := exploration.CompareCohorts(request)
		if err != nil {
			fail(err)
			return
		}
		httpjson.Write(w, 200, result)
	})
	registerRead(mux, "/environment-scopes", exploration.AdoptionWindow.Validate, source.DiscoverEnvironmentScopes)
	registerRead(mux, "/environment-scopes/records", exploration.AdoptionRecordsQuestion.Validate, source.PreviewAdoptionRecords)
	registerRead(mux, "/inventory", exploration.InventoryQuestion.Validate, source.Inventory)
	registerRead(mux, "/discovery", exploration.InventoryQuestion.Validate, source.DiscoverServices)
	registerRead(mux, "/record-aggregates", exploration.ValidateRecordAggregatesQuery, source.FetchRecordAggregates)
	registerRead(mux, "/explore/counted-values", exploration.ValidateCountedValuesQuery, source.FetchCountedValues)
	registerRead(mux, "/trace-search", func(request exploration.TraceSearchRequest) error {
		_, err := exploration.CompileTraceSearch(request)
		return err
	}, source.SearchTraces)
	registerRead(mux, "/explore", exploration.Query.Validate, func(ctx context.Context, q exploration.Query) (*exploration.Result, error) {
		result, err := source.Fetch(ctx, q)
		if err != nil {
			return nil, err
		}
		if token, err := snapshots.Freeze(result); err == nil {
			result.ResultToken = token
		} else {
			result.Warnings = append(result.Warnings, "Live selection retention unavailable: "+err.Error())
		}
		return result, nil
	})
}

// Type inference keeps route wiring ordinary while sharing the validation-before-
// execution boundary. Family-specific validators and source methods remain typed.
func registerRead[Q, R any](mux *http.ServeMux, path string, validate func(Q) error, read func(context.Context, Q) (R, error)) {
	policy, ok := PolicyFor("POST", path)
	if !ok {
		panic("recordhttp: missing source-read policy for " + path)
	}
	mux.HandleFunc("POST "+path, queryhttp.ReadHandler(policy, validate, read))
}

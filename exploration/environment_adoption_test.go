package exploration

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnvironmentPredicateIsBoundedAndParameterized(t *testing.T) {
	for _, scope := range []identity.EnvironmentScope{{Mode: "unreported"}, {Mode: "empty"}, {Mode: "named", Value: "prod' OR 1=1 --"}} {
		sql, params, e := EnvironmentPredicate(scope)
		if e != nil {
			t.Fatal(e)
		}
		if !strings.Contains(sql, "mapContains") || scope.Mode == "named" && (strings.Contains(sql, scope.Value) || params["environment"] != scope.Value) {
			t.Fatal(sql, params)
		}
	}
	if _, _, e := EnvironmentPredicate(identity.EnvironmentScope{}); e == nil {
		t.Fatal("zero scope compiled")
	}
}
func TestAdoptionSourceScopeAndWindowGuards(t *testing.T) {
	q := AdoptionRecordsQuestion{AdoptionWindow: AdoptionWindow{Signal: "logs", FromMS: 1000, ToMS: 2000}, Scope: identity.EnvironmentScope{Mode: "unreported"}, Limit: 1}
	for _, tc := range []struct {
		name, wire string
		ok         bool
	}{
		{"unreported", `{"timestamp_ns":"1000000000","ResourceAttributes":{},"Body":"hello"}`, true},
		{"empty is not unreported", `{"timestamp_ns":"1000000000","ResourceAttributes":{"deployment.environment.name":""},"Body":"hello"}`, false},
		{"legacy is not unreported", `{"timestamp_ns":"1000000000","ResourceAttributes":{"deployment.environment":"prod"},"Body":"hello"}`, false},
		{"missing resource map", `{"timestamp_ns":"1000000000","Body":"hello"}`, false},
		{"outside window", `{"timestamp_ns":"2000000000","ResourceAttributes":{},"Body":"hello"}`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, tc.wire) }))
			defer server.Close()
			got, e := (Source{Endpoint: server.URL}).PreviewAdoptionRecords(t.Context(), q)
			if (e == nil) != tc.ok {
				t.Fatal(got, e)
			}
		})
	}
}
func TestEnvironmentDiagnosticsPreservePresenceConflictAndLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, row := range []map[string]any{{"current_present": 0, "legacy_present": 0, "current_value": "", "legacy_value": "", "count": "1"}, {"current_present": 1, "legacy_present": 1, "current_value": "", "legacy_value": "prod", "count": "9007199254740993"}} {
			json.NewEncoder(w).Encode(row)
		}
	}))
	defer server.Close()
	q := AdoptionWindow{Signal: "spans", FromMS: 1000, ToMS: 2000}
	r, e := (Source{Endpoint: server.URL}).DiscoverEnvironmentScopes(t.Context(), q)
	if e != nil || len(r.Groups) != 2 || r.Groups[0].Observation.Scope.Mode != "unreported" || !r.Groups[1].Observation.Conflict || r.Groups[1].Observation.Scope.Mode != "empty" || r.Groups[1].Count != "9007199254740993" {
		t.Fatal(r, e)
	}
	for _, q := range []AdoptionWindow{{}, {Signal: "metrics", FromMS: 1, ToMS: 2}, {Signal: "logs", FromMS: 2, ToMS: 1}} {
		if q.Validate() == nil {
			t.Fatal("invalid adoption window accepted", q)
		}
	}
}

func TestEnvironmentDiagnosticsMalformedRowsAndOverflow(t *testing.T) {
	for _, wire := range []string{`{"current_value":"","legacy_value":"","count":"1"}`, `{"current_present":0,"legacy_present":0,"current_value":"hidden","legacy_value":"","count":"1"}`, `{"current_present":1,"legacy_present":0,"current_value":"prod","legacy_value":"","count":"01"}`, `{"current_present":1,"legacy_present":0,"current_value":"prod","legacy_value":"","count":"0"}`} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, wire) }))
		_, e := (Source{Endpoint: server.URL}).DiscoverEnvironmentScopes(t.Context(), AdoptionWindow{Signal: "logs", FromMS: 1000, ToMS: 2000})
		server.Close()
		if e == nil {
			t.Fatal("malformed scope row accepted", wire)
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := range 65 {
			json.NewEncoder(w).Encode(map[string]any{"current_present": 1, "legacy_present": 0, "current_value": fmt.Sprint("env-", i), "legacy_value": "", "count": "1"})
		}
	}))
	defer server.Close()
	result, e := (Source{Endpoint: server.URL}).DiscoverEnvironmentScopes(t.Context(), AdoptionWindow{Signal: "logs", FromMS: 1000, ToMS: 2000})
	if e != nil || len(result.Groups) != 64 || !result.HasMore {
		t.Fatal(result, e)
	}
}

func TestAdoptionRejectsWrongOverflowAndBoundsRetainedBytes(t *testing.T) {
	q := AdoptionRecordsQuestion{AdoptionWindow: AdoptionWindow{Signal: "logs", FromMS: 1000, ToMS: 2000}, Scope: identity.EnvironmentScope{Mode: "unreported"}, Limit: 1}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"timestamp_ns":"1000000000","ResourceAttributes":{},"Body":"good"}`)
		fmt.Fprintln(w, `{"timestamp_ns":"1000000001","ResourceAttributes":{"deployment.environment.name":"prod"},"Body":"wrong overflow"}`)
	}))
	_, e := (Source{Endpoint: server.URL}).PreviewAdoptionRecords(t.Context(), q)
	server.Close()
	if e == nil {
		t.Fatal("wrong-scope overflow row accepted")
	}
	q.Scope = identity.EnvironmentScope{Mode: "empty"}
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"timestamp_ns":"1000000000","ResourceAttributes":{"deployment.environment.name":null},"Body":"bad"}`)
	}))
	_, e = (Source{Endpoint: server.URL}).PreviewAdoptionRecords(t.Context(), q)
	server.Close()
	if e == nil {
		t.Fatal("null treated as empty environment")
	}
	q.Scope = identity.EnvironmentScope{Mode: "unreported"}
	q.Limit = 100
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := range 100 {
			fmt.Fprintf(w, "{\"timestamp_ns\":\"%d\",\"ResourceAttributes\":{},\"Body\":\"%s\"}\n", 1000000000+i, strings.Repeat("x", 20000))
		}
	}))
	defer server.Close()
	out, e := (Source{Endpoint: server.URL}).PreviewAdoptionRecords(t.Context(), q)
	if e != nil || !out.Limited || len(out.Rows) == 0 || len(out.Rows) >= 100 {
		t.Fatal(out, e)
	}
	total := 0
	for _, row := range out.Rows {
		total += len(row)
	}
	if total > 1<<20 {
		t.Fatal("retained byte limit bypassed", total)
	}
}

func TestAdoptionSelectorMetadataMatchesSourcePolicy249(t *testing.T) {
	for _, scope := range []identity.EnvironmentScope{{Mode: "unreported"}, {Mode: "empty"}, {Mode: "named", Value: "prod"}} {
		resource := map[string]string{}
		if scope.Mode == "empty" {
			resource[identity.EnvironmentKey] = ""
		}
		if scope.Mode == "named" {
			resource[identity.LegacyEnvironmentKey] = scope.Value
		}
		wire, _ := json.Marshal(map[string]any{"timestamp_ns": "1000000000", "ResourceAttributes": resource, "Body": "hello"})
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if scope.Mode == "named" && r.URL.Query().Get("param_environment") != scope.Value {
				t.Error("selection binding changed")
			}
			w.Write(wire)
		}))
		q := AdoptionRecordsQuestion{AdoptionWindow: AdoptionWindow{Signal: "logs", FromMS: 1000, ToMS: 2000}, Scope: scope, Limit: 1}
		result, err := (Source{Endpoint: server.URL, Client: server.Client()}).PreviewAdoptionRecords(t.Context(), q)
		server.Close()
		if err != nil {
			t.Fatal(err)
		}
		selection, err := result.Selector.Selection()
		expected, _ := identity.EffectiveEnvironment(scope)
		if err != nil || !selection.Equal(expected) || !selection.Matches(resource) {
			t.Fatal("selector metadata is not actual effective source policy", err, result.Selector)
		}
	}
}

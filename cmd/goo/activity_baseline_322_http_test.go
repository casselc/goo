package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/workbench"
)

func TestActivityTypedSavedDurationBaselineHTTP322(t *testing.T) {
	for _, mode := range []string{"named", "empty", "unreported"} {
		t.Run(mode, func(t *testing.T) {
			selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: mode}}
			if mode == "named" {
				selector.Scope.Value = "prod"
			}
			type read struct {
				sql    string
				params url.Values
			}
			reads := []read{}
			source := exploration.Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: nativeDependencyTransport(func(request *http.Request) (*http.Response, error) {
				sql, _ := io.ReadAll(request.Body)
				text := string(sql)
				params := request.URL.Query()
				reads = append(reads, read{text, params})
				body := ""
				summary := `"count":"2","sum_ns":"200","min_ns":"100","max_ns":"100","quantiles_ns":["100","100","100"]`
				switch {
				case strings.Contains(text, "SELECT DISTINCT"):
				case strings.Contains(text, "GROUP BY"):
					value := "api"
					if !strings.Contains(text, "sum_ns") {
						value = "stable"
						for _, values := range params {
							for _, v := range values {
								if v == "canary" {
									value = "canary"
								}
							}
						}
					}
					if strings.Contains(text, "sum_ns") {
						body = `{"g0":"api",` + summary + `}`
					} else {
						body = fmt.Sprintf(`{"g0":%q,"count":"2"}`, value)
					}
				case strings.Contains(text, "toString(count()) AS count"):
					body = `{` + summary + `}`
				default:
					from, _ := strconv.ParseInt(params.Get("param_from"), 10, 64)
					resource := map[string]string{"service.namespace": "team"}
					if mode == "named" {
						resource["deployment.environment"] = "prod"
					} else if mode == "empty" {
						resource["deployment.environment.name"] = ""
						resource["deployment.environment"] = "prod"
					}
					for index := range 2 {
						row := map[string]any{"timestamp_ns": strconv.FormatInt(from*1000000+int64(index+1), 10), "duration_ns": "100", "ServiceName": "api", "SpanName": "request", "ResourceAttributes": resource, "Attributes": map[string]string{"app.release": "stable"}}
						raw, _ := json.Marshal(row)
						body += string(raw) + "\n"
					}
				}
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})}}
			snapshots := &exploration.ResultSnapshots{}
			mux := http.NewServeMux()
			registerExploration(mux, source, snapshots, nil)
			capturehttp.Register(mux, source, t.TempDir(), snapshots)
			call := func(path string, body any, want int) []byte {
				t.Helper()
				raw, _ := json.Marshal(body)
				response := httptest.NewRecorder()
				mux.ServeHTTP(response, httptest.NewRequest("POST", path, bytes.NewReader(raw)))
				if response.Code != want {
					t.Fatal(path, response.Code, response.Body.String())
				}
				return response.Body.Bytes()
			}
			namespace := exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: "team"}
			release := exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "app.release"}, Op: "eq", Value: "stable"}
			baseline := exploration.Query{Signal: "spans", EnvironmentSelector: selector, FromMS: 1000, ToMS: 3000, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, Summary: "span-duration", Where: &exploration.Predicate{All: []exploration.Predicate{namespace, release}}}
			raw := call("/captures", map[string]any{"title": "Typed duration baseline", "question": exploration.CaptureQuestion{Kind: "explore", Explore: &baseline}}, 201)
			var saved exploration.CapturePage
			if err := json.Unmarshal(raw, &saved); err != nil {
				t.Fatal(err)
			}
			if len(reads) != 4 || saved.Explore == nil || saved.Explore.Duration == nil || saved.Explore.Duration.SumNS != "200" || !saved.Receipt.Question.Explore.SameEnvironment(baseline) {
				t.Fatal("typed duration capture or SQL budget changed")
			}
			get := httptest.NewRecorder()
			mux.ServeHTTP(get, httptest.NewRequest("GET", "/captures/"+saved.Receipt.ID, nil))
			if get.Code != 200 || get.Body.String() != string(raw) || len(reads) != 4 {
				t.Fatal("reopening a frozen baseline read telemetry or changed its receipt")
			}
			current := baseline
			current.FromMS = 5000
			current.ToMS = 7000
			current.Where = &exploration.Predicate{All: []exploration.Predicate{namespace, {Field: &exploration.Field{Kind: "attribute", Key: "app.release"}, Op: "eq", Value: "canary"}}}
			groups := []exploration.ValueCount{}
			for _, rawGroup := range saved.Explore.Groups {
				var group exploration.ValueCount
				if err := json.Unmarshal(rawGroup, &group); err != nil {
					t.Fatal(err)
				}
				if group.Duration == nil || group.Duration.Validate() != nil {
					t.Fatal("saved duration group missing")
				}
				groups = append(groups, group)
			}
			a := exploration.Cohort{Question: saved.Explore.Question, Count: saved.Explore.Count, State: saved.Explore.State, Basis: saved.Explore.Basis, CapturedAt: saved.Explore.CapturedAt, Duration: saved.Explore.Duration, Groups: groups}
			duration := &exploration.DurationSummary{State: "observed", Count: "2", SumNS: "400", MinNS: "200", MaxNS: "200", P50NS: "200", P95NS: "200", P99NS: "200"}
			group := "api"
			b := exploration.Cohort{Question: current, Count: "2", State: "observed", Basis: "held-current", Duration: duration, Groups: []exploration.ValueCount{{Value: &group, Count: "2", Duration: duration}}}
			var comparison exploration.CohortComparison
			if err := json.Unmarshal(call("/compare", exploration.ComparisonRequest{Baseline: a, Selection: b}, 200), &comparison); err != nil {
				t.Fatal(err)
			}
			if len(reads) != 4 || comparison.Measures == nil || comparison.Measures.BaselineDuration == nil || comparison.Measures.MeanDifferenceNS != "100.000" {
				t.Fatal("pure saved duration comparison lost exact aggregates or reread source")
			}
			mismatch := b
			mismatch.Question = current
			if mode == "named" {
				mismatch.Question.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
			} else {
				other := "empty"
				if mode == "empty" {
					other = "unreported"
				}
				mismatch.Question.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: other}}
			}
			call("/compare", exploration.ComparisonRequest{Baseline: a, Selection: mismatch}, 400)
			if len(reads) != 4 {
				t.Fatal("policy mismatch reached source")
			}
			field := exploration.Field{Kind: "attribute", Key: "app.release"}
			request := exploration.ExplanationRequest{Version: exploration.ExplanationVersion, Baseline: saved.Explore.Question, Selection: current, Fields: []exploration.Field{field}}
			call("/population/explain/plan", request, 200)
			if len(reads) != 4 {
				t.Fatal("explanation planning read telemetry")
			}
			var explained exploration.ExplanationResult
			if json.Unmarshal(call("/population/explain", request, 200), &explained) != nil || !reflect.DeepEqual(explained.Request, request) || explained.State != "complete" || len(reads) != 8 {
				t.Fatal("typed explanation did not reread its exact two scopes")
			}
			for _, read := range reads[4:] {
				from, to := read.params.Get("param_from"), read.params.Get("param_to")
				releaseValue := ""
				hasNamespace := false
				for _, values := range read.params {
					for _, value := range values {
						hasNamespace = hasNamespace || value == "team"
						if value == "stable" || value == "canary" {
							releaseValue = value
						}
					}
				}
				if !hasNamespace || from == "1000" && (to != "3000" || releaseValue != "stable") || from == "5000" && (to != "7000" || releaseValue != "canary") || from != "1000" && from != "5000" || !strings.Contains(read.sql, "'deployment.environment'") {
					t.Fatal("explanation dropped typed policy/namespace/predicate/window")
				}
				_, named := read.params["param_environment"]
				if named != (mode == "named") {
					t.Fatal("empty/unreported read acquired a fallback name")
				}
			}
			for _, side := range []string{"baseline", "selection"} {
				var filtered struct {
					Question exploration.Query `json:"question"`
				}
				if json.Unmarshal(call("/population/explain/filter", map[string]any{"request": request, "field": field, "value": "interesting", "side": side, "mode": "include"}, 200), &filtered) != nil {
					t.Fatal("invalid planned filter")
				}
				expected := request.Baseline
				if side == "selection" {
					expected = request.Selection
				}
				if !filtered.Question.SameEnvironment(expected) || filtered.Question.FromMS != expected.FromMS || filtered.Question.ToMS != expected.ToMS || filtered.Question.Where == nil || len(filtered.Question.Where.All) != 2 || !reflect.DeepEqual(filtered.Question.Where.All[0], *expected.Where) || len(reads) != 8 {
					t.Fatal("filter changed its selected side or performed IO")
				}
				step := workbench.Step{Version: 1, Kind: "record", Scope: workbench.StepScope{EnvironmentSelector: filtered.Question.EnvironmentSelector, FromMS: filtered.Question.FromMS, ToMS: filtered.Question.ToMS}, View: "Records", Record: &filtered.Question}
				if err := step.Validate(); err != nil {
					t.Fatal(err)
				}
				wire, _ := json.Marshal(step)
				link := "#step=" + url.PathEscape(string(wire))
				decoded, _ := url.PathUnescape(strings.TrimPrefix(link, "#step="))
				var reopened workbench.Step
				if json.Unmarshal([]byte(decoded), &reopened) != nil || reopened.Validate() != nil || reopened.Record.Environment != "" || !reflect.DeepEqual(reopened.Record.EnvironmentSelector, selector) {
					t.Fatal("encoded record-step payload lost its typed environment")
				}
			}
			bad := request
			bad.Selection = mismatch.Question
			call("/population/explain", bad, 400)
			call("/population/explain/filter", map[string]any{"request": bad, "field": field, "value": "interesting", "side": "selection", "mode": "include"}, 400)
			if len(reads) != 8 {
				t.Fatal("invalid explanation/filter policy reached source")
			}
		})
	}
}

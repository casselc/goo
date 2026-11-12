package httpresponse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// These generated configuration/membership checks are not source observations.
func TestIncomingResponsePopulation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for _, mode := range []string{"named", "empty", "unreported"} {
			env := identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: mode}}
			if mode == "named" {
				env.Scope.Value = fmt.Sprintf("prod'雪-%d", seed)
				if seed%2 == 0 {
					env.Policy = "current-only"
				}
			}
			for _, ns := range []*identity.Namespace{{Mode: "missing"}, {Mode: "value", Value: ""}, {Mode: "value", Value: "team'雪"}} {
				for _, profile := range []string{"auto", "http-server", "all-spans", "http-client"} {
					in := exploration.DashboardRequest{Profile: profile, HTTPServerObserved: true, Scope: exploration.InventoryQuestion{EnvironmentSelector: &env, Namespace: ns, FromMS: int64(1000 + seed), ToMS: int64(61000 + seed)}, Service: fmt.Sprintf("api'雪-%d", seed), Signals: []string{"logs", "spans"}, Fields: []exploration.Field{{Kind: "attribute", Key: "http.request.method"}, {Kind: "attribute", Key: "http.status_code"}}}
					plan, err := exploration.PlanServiceDashboard(in, nil)
					if err != nil {
						tc.Fatal(err)
					}
					raw, _ := json.Marshal(plan)
					var decoded exploration.DashboardPlan
					if json.Unmarshal(raw, &decoded) != nil {
						tc.Fatal("whole dashboard carrier cannot decode")
					}
					again, _ := json.Marshal(decoded)
					if !bytes.Equal(raw, again) || !reflect.DeepEqual(plan.Scope, in.Scope) {
						tc.Fatal("whole dashboard scope/carrier changed")
					}
					var response, activity *exploration.Query
					for _, w := range plan.Widgets {
						if w.Query != nil {
							compiled, e := exploration.Compile(*w.Query)
							if e != nil {
								tc.Fatal(e)
							}
							if strings.Contains(compiled.Rows, in.Service) {
								tc.Fatal("literal service interpolated into SQL")
							}
						}
						if w.Key == "spans" {
							activity = w.Query
						}
						if w.Key == "http-server-responses" {
							if response != nil || w.Family != "http-server" || w.Grain != "server-span-record" {
								tc.Fatal("duplicate or foreign response widget")
							}
							response = w.Query
						}
					}
					if plan.Profile != "http-server" {
						if response != nil {
							tc.Fatal("incoming response widget leaked into another profile")
						}
						continue
					}
					if response == nil || activity == nil || !response.SpanStatus || response.Summary != "span-duration" || !reflect.DeepEqual(response.GroupBy, []exploration.Field{{Kind: "attribute", Key: "http.response.status_code"}}) {
						tc.Fatal("response grouping or independent measures changed")
					}
					// Activity requests its existing 1s timeline for this 60s window.
					// The context needs no timeline read; neither projection selects rows.
					if response.BucketMS != 0 || activity.BucketMS != 1000 || response.Limit != 100 || response.Offset != 0 {
						tc.Fatal("response context changed its bounded non-temporal projection")
					}
					same := *response
					same.GroupBy = activity.GroupBy
					same.BucketMS = activity.BucketMS
					if !reflect.DeepEqual(same, *activity) {
						tc.Fatal("response grouping changed the existing Server population")
					}
					for _, kind := range []string{"Server", "Client", "Internal", "SERVER", ""} {
						for _, method := range []*string{nil, ptr(""), ptr("GET"), ptr("_OTHER"), ptr(" "), ptr("custom'雪")} {
							values := map[exploration.Field]string{{Kind: "column", Key: "service"}: in.Service, {Kind: "column", Key: "kind"}: kind, {Kind: "attribute", Key: "http.method"}: "GET"}
							if ns.Mode == "value" {
								values[exploration.Field{Kind: "resource", Key: "service.namespace"}] = ns.Value
							}
							if method != nil {
								values[exploration.Field{Kind: "attribute", Key: "http.request.method"}] = *method
							}
							want := kind == "Server" && method != nil && *method != ""
							// Response strings and span Error vary independently; neither selects the population.
							for _, code := range []*string{nil, ptr(""), ptr("200"), ptr("503"), ptr("0200"), ptr("unexpected'雪")} {
								for _, status := range []string{"Error", "Ok", "Unset"} {
									row := maps.Clone(values)
									row[exploration.Field{Kind: "column", Key: "status"}] = status
									if code != nil {
										row[exploration.Field{Kind: "attribute", Key: "http.response.status_code"}] = *code
									}
									if exploration.Match(response.Where, row) != want {
										tc.Fatal("response/status/legacy method changed Server membership")
									}
									row[exploration.Field{Kind: "column", Key: "service"}] = "foreign"
									if exploration.Match(response.Where, row) {
										tc.Fatal("foreign service admitted")
									}
									row = maps.Clone(values)
									row[exploration.Field{Kind: "resource", Key: "service.namespace"}] = "foreign"
									if exploration.Match(response.Where, row) {
										tc.Fatal("foreign namespace admitted")
									}
								}
							}
						}
					}
				}
			}
		}
	}, hegel.WithTestCases(8), hegel.WithSeed(410), hegel.WithDatabase(filepath.Join(t.TempDir(), "hegel")))
}

func ptr(s string) *string { return &s }

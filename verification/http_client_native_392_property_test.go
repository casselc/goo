package verification

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type clientNativeVector392 struct {
	Instrument  measurements.Instrument         `json:"instrument"`
	Plan        *exploration.DashboardPlan      `json:"plan"`
	Shape       bool                            `json:"shape"`
	Eligible    bool                            `json:"eligible"`
	Source      measurements.Question           `json:"source"`
	Destination measurements.Question           `json:"destination"`
	Declaration workbench.NativeViewDeclaration `json:"declaration"`
	View        *workbench.NativeView           `json:"view,omitempty"`
}

// This checks discovered-profile adoption and exact source ownership, not a
// second implementation of histogram evaluation. The latter keeps its existing
// delta/cumulative/reset/gap/flag kernel and observed-source regressions.
func TestHegelHTTPClientNativeProfile392(t *testing.T) {
	vectors := []clientNativeVector392{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for variant := 0; variant < 8; variant++ {
			scope := exploration.InventoryQuestion{Environment: "prod", FromMS: 1790000000000, ToMS: 1790000060000, Namespace: &identity.Namespace{Mode: "value", Value: fmt.Sprintf("team-%c", 'a'+seed%5)}}
			service := fmt.Sprintf("caller-%c", 'a'+seed%7)
			i := hdInstrument("outgoing-duration", "http.client.request.duration", "histogram", "s", scope, service)
			i.SchemaURL = "https://opentelemetry.io/schemas/1.45.0"
			attrs := map[string]string{"http.request.method": []string{"GET", "POST", "_OTHER"}[seed%3], "server.address": fmt.Sprintf("peer-%c.example", 'a'+seed%11), "server.port": fmt.Sprint(8000 + seed%5), "literal.owner": fmt.Sprintf("owner-%c", 'a'+seed%7)}
			switch variant {
			case 1:
				i.Temporality = 2
				i.SchemaURL = ""
			case 2:
				i.SchemaURL = "https://example.test/custom-schema"
				delete(attrs, "http.request.method")
				attrs["server.address"] = ""
				delete(attrs, "server.port")
			case 3:
				i.Unit = "ms"
			case 4:
				i.Kind = "gauge"
				i.Temporality = 0
			case 5:
				i.Temporality = 0
				i.Supported = false
			case 6:
				i.Metric = "custom.request.duration"
			case 7:
				i.Kind = "exponential-histogram"
				i.Supported = false
			}
			if i.Definition != nil {
				d := i.Definition
				d.Kind = i.Kind
				d.Source.Metric = i.Metric
				d.Source.Unit = i.Unit
				d.Source.SchemaURL = i.SchemaURL
			}
			resource := map[string]string{"service.name": service, "service.namespace": scope.Namespace.Value, "service.instance.id": fmt.Sprintf("instance-%c", 'a'+seed%3), "deployment.environment.name": "prod", "literal.resource": "kept"}
			stream := metrics.Stream{Resource: resource, Attributes: attrs, Scope: i.Scope, Version: i.Version, SchemaURL: i.SchemaURL, ResourceSchemaURL: "https://example.test/resource-schema", Name: i.Metric, Unit: i.Unit}
			if i.Supported {
				i.Definition.Source.ExactStream = &stream
			} else {
				i.Definition = nil
			}
			shape := variant <= 2 || variant == 5
			capability := measurements.PlanInstrument(i, []measurements.DiscoveryObservation{{Resource: resource, Attributes: attrs, ResourceSchemaURL: stream.ResourceSchemaURL}})
			if capability.ShapeMatches != shape || capability.Eligible != i.Supported || capability.Verification != "not-run" {
				tc.Fatalf("variant%d shape/eligibility adopted extra authority: %+v", variant, capability)
			}
			if shape && capability.Profile != "http-client-request-duration" {
				tc.Fatal("exact client name did not retain diagnostic profile")
			}
			if len(capability.Coverage.Streams) != 1 || !reflect.DeepEqual(capability.Coverage.Streams[0].Stream, stream) {
				tc.Fatal("coverage changed full stream identity")
			}
			i.CapabilityPlan = &capability
			plan, e := exploration.PlanServiceDashboard(exploration.DashboardRequest{Scope: scope, Service: service, Profile: "auto", Signals: []string{"metrics"}, Instruments: []measurements.Instrument{i}}, nil)
			if e != nil || plan.Profile == "http-client" {
				tc.Fatal("metric profile promoted Client-span evidence", e)
			}
			qualified := 0
			for _, w := range plan.Widgets {
				if w.Measurement == nil || w.Family != "http-client" {
					continue
				}
				qualified++
				if variant > 2 || w.Grain != "per-stream" || !reflect.DeepEqual(w.Measurement.Definition, *i.Definition) || w.Measurement.FromMS != scope.FromMS || w.Measurement.ToMS != scope.ToMS {
					tc.Fatal("qualified widget changed exact native source")
				}
			}
			if (qualified == 1) != (variant <= 2) {
				tc.Fatalf("variant%d native qualification count%d", variant, qualified)
			}
			if variant > 2 {
				vectors = append(vectors, clientNativeVector392{Instrument: i, Plan: plan, Shape: shape, Eligible: i.Supported})
				continue
			}
			source := measurements.Question{Definition: *i.Definition, FromMS: scope.FromMS, ToMS: scope.ToMS}
			raw, _ := json.Marshal(source)
			var destination measurements.Question
			json.Unmarshal(raw, &destination)
			destination.FromMS += 120000
			destination.ToMS += 120000
			destination.Definition.Name = "destination-duration"
			destination.Definition.Environment = "destination"
			destination.Definition.Source.Services = []string{"destination-worker"}
			destination.Definition.Source.Attributes = map[string]string{"only.destination": "yes"}
			destination.Definition.Source.ExactStream = nil
			declaration := workbench.NativeViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeViewSelection{Kind: "export-interval"}, Kind: workbench.ViewKind{Name: []string{"histogram", "table"}[seed%2], Version: 1}, Display: workbench.NativeViewDisplay{Category: "bucket", Value: "observation-count", RequiresComplete: true}}
			saved, e := (workbench.Store{Dir: t.TempDir()}).CreateNativeView(workbench.NativeViewInput{Version: 1, Title: "Native outgoing observations", Question: source, Declaration: declaration})
			if e != nil {
				tc.Fatal(e)
			}
			reused, e := workbench.ReuseNativeView(*saved, destination)
			if e != nil || !reflect.DeepEqual(reused.Declaration, declaration) || !reflect.DeepEqual(reused.Question, destination) {
				tc.Fatal("compatible native display reuse failed", e)
			}
			vectors = append(vectors, clientNativeVector392{i, plan, shape, i.Supported, source, destination, declaration, saved})
		}
	}, options(16)...)
	if err := runNativeHistogramJS(map[string]any{"cases": vectors}, clientNativeJS392); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d shared profile/source vectors", len(vectors))
}

const clientNativeJS392 = `import fs from 'node:fs';import assert from 'node:assert/strict';import * as D from '../ui/service-dashboard.mjs';import * as N from '../ui/native-view-declaration.mjs';
globalThis.fetch=()=>{throw Error('Local native reuse attempted a telemetry read')};for(const f of JSON.parse(fs.readFileSync(0,'utf8')).cases){const before=JSON.stringify(f);const input={scope:f.plan.scope,service:f.plan.service??'caller',profile:'auto',signals:['metrics'],instruments:[f.instrument]};const staged=D.planningInput(input);assert.equal(staged.profile,'auto');assert.equal(staged.instruments[0].unit,f.instrument.unit);if(!f.shape||!f.eligible){assert.equal(f.plan.widgets.some(w=>w.measurement&&w.family==='http-client'),false);assert.equal(JSON.stringify(f),before);continue;}for(const w of f.plan.widgets.filter(w=>w.measurement&&w.family==='http-client')){assert.equal(w.grain,'per-stream');assert.match(D.widgetQualifier(w),/HTTP.client/i);assert.deepEqual(w.measurement,f.source);}const view=f.view;const reused=N.reuse(view,f.destination);assert.deepEqual(reused.declaration,f.declaration);assert.deepEqual(reused.question,f.destination);assert.equal(JSON.stringify(f),before,'reuse/planning changed source or destination');const wrong=structuredClone(f.destination);wrong.definition.source.unit='ms';assert.throws(()=>N.reuse(view,wrong));}`

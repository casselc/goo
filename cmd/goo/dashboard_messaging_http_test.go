package main

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func dashboardMessagingBody(profile any, fields []exploration.Field, hint any) string {
	var body map[string]any
	_ = json.Unmarshal([]byte(dashboardProfileBody(profile, []string{"spans", "logs"}, fields)), &body)
	body["messagingConsumerObserved"] = hint
	raw, _ := json.Marshal(body)
	return string(raw)
}

func dashboardMessagingNativeBody(unit string) string {
	var body map[string]any
	_ = json.Unmarshal([]byte(dashboardProfileBody("auto", []string{"metrics"}, []exploration.Field{})), &body)
	d := measurements.Definition{APIVersion: measurements.Version, Name: "messaging-consumer-duration", Title: "Messaging processing duration", Kind: "histogram", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Namespace: &identity.Namespace{Mode: "value", Value: "commerce"}, Services: []string{"checkout"}, Scope: "ordinary-sdk", Metric: "messaging.process.duration", Unit: unit}}
	body["instruments"] = []measurements.Instrument{{Kind: "histogram", Scope: "ordinary-sdk", Metric: d.Source.Metric, Unit: unit, Temporality: 2, Supported: true, Definition: &d}}
	raw, _ := json.Marshal(body)
	return string(raw)
}

func TestDashboardMessagingProfileHTTPWireFixtures(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "messaging.system"}, {Kind: "attribute", Key: "messaging.destination.name"}}
	partial := []exploration.Field{{Kind: "attribute", Key: "messaging.destination"}}
	mixed := append(append([]exploration.Field{}, fields...), exploration.Field{Kind: "attribute", Key: "http.request.method"})
	both := dashboardMessagingBody("auto", mixed, true)
	var body map[string]any
	_ = json.Unmarshal([]byte(both), &body)
	body["httpServerObserved"] = true
	raw, _ := json.Marshal(body)
	rpcBody := map[string]any{}
	_ = json.Unmarshal([]byte(dashboardMessagingBody("auto", append(append([]exploration.Field{}, fields...), exploration.Field{Kind: "attribute", Key: "rpc.system"}, exploration.Field{Kind: "attribute", Key: "rpc.method"}), true)), &rpcBody)
	rpcBody["rpcServerObserved"] = true
	rpcRaw, _ := json.Marshal(rpcBody)
	var producerBody map[string]any
	_ = json.Unmarshal([]byte(dashboardMessagingBody("auto", fields, false)), &producerBody)
	producerBody["messagingProducerObserved"] = true
	producerRaw, _ := json.Marshal(producerBody)
	var explicitProducer map[string]any
	_ = json.Unmarshal([]byte(dashboardMessagingBody("messaging-producer", fields, false)), &explicitProducer)
	explicitRaw, _ := json.Marshal(explicitProducer)
	var dbBody map[string]any
	_ = json.Unmarshal([]byte(dashboardMessagingBody("auto", append(append([]exploration.Field{}, fields...), exploration.Field{Kind: "attribute", Key: "db.system.name"}), true)), &dbBody)
	dbBody["dbClientObserved"] = true
	dbRaw, _ := json.Marshal(dbBody)
	producerBody["messagingConsumerObserved"] = true
	bothRolesRaw, _ := json.Marshal(producerBody)
	producerBody["messagingProducerObserved"] = "yes"
	wrongProducerRaw, _ := json.Marshal(producerBody)
	var nativeProducer map[string]any
	_ = json.Unmarshal([]byte(dashboardMessagingNativeBody("s")), &nativeProducer)
	nativeProducer["profile"] = "messaging-producer"
	nativeProducerRaw, _ := json.Marshal(nativeProducer)
	cases := []planningWireCase{
		{Name: "Messaging-flat-keys-candidate-only", Body: dashboardMessagingBody("auto", fields, false), Status: 200},
		{Name: "Messaging-positive-consumer-hint", Body: dashboardMessagingBody("auto", fields, true), Status: 200},
		{Name: "explicit-Messaging-without-positive-hint", Body: dashboardMessagingBody("messaging-consumer", fields, false), Status: 200},
		{Name: "explicit-all-spans-Messaging-candidate", Body: dashboardMessagingBody("all-spans", fields, true), Status: 200},
		{Name: "Messaging-destination-only-not-admitted", Body: dashboardMessagingBody("auto", partial, true), Status: 200},
		{Name: "explicit-Messaging-destination-only-refused", Body: dashboardMessagingBody("messaging-consumer", partial, true), Status: 400},
		{Name: "both-positive-hints-HTTP-precedence", Body: string(raw), Status: 200},
		{Name: "RPC-positive-hint-before-Messaging", Body: string(rpcRaw), Status: 200},
		{Name: "Messaging-native-seconds-auto", Body: dashboardMessagingNativeBody("s"), Status: 200},
		{Name: "Messaging-native-wrong-unit-no-admission", Body: dashboardMessagingNativeBody("ms"), Status: 200},
		{Name: "Messaging-positive-producer-hint", Body: string(producerRaw), Status: 200},
		{Name: "Messaging-explicit-producer", Body: string(explicitRaw), Status: 200},
		{Name: "DB-before-Messaging", Body: string(dbRaw), Status: 200},
		{Name: "Consumer-before-Producer", Body: string(bothRolesRaw), Status: 200},
		{Name: "Wrong-producer-hint-type", Body: string(wrongProducerRaw), Status: 400},
		{Name: "Producer-without-system-refused", Body: dashboardMessagingBody("messaging-producer", partial, false), Status: 400},
		{Name: "Native-processing-cannot-admit-producer", Body: string(nativeProducerRaw), Status: 400},
		{Name: "Messaging-wrong-hint-type", Body: dashboardMessagingBody("auto", fields, "yes"), Status: 400},
	}
	for i := range cases {
		cases[i].Method = "POST"
		cases[i].Path = "/dashboards/plan"
	}
	const filename = "../../testdata/http/dashboard-messaging-profiles.json"
	if os.Getenv("GOO_UPDATE_DASHBOARD_MESSAGING_HTTP_FIXTURES") == "1" {
		for i := range cases {
			w := planningWireResponse(t, cases[i])
			if w.Code != cases[i].Status {
				t.Fatal(cases[i].Name, w.Code, w.Body.String())
			}
			cases[i].Response = w.Body.String()
			cases[i].Headers = map[string]string{}
			for _, k := range []string{"Content-Type", "Cache-Control", "Allow", "X-Content-Type-Options"} {
				cases[i].Headers[k] = w.Header().Get(k)
			}
		}
		out, _ := json.MarshalIndent(planningWireFixtures{Version: 1, Cases: cases}, "", "  ")
		out = append(out, '\n')
		if err := os.WriteFile(filename, out, 0644); err != nil {
			t.Fatal(err)
		}
	}
	fixture := dashboardProfileFixture(t, filename, len(cases))
	for _, c := range fixture.Cases {
		t.Run(c.Name, func(t *testing.T) {
			w := planningWireResponse(t, c)
			if w.Code != c.Status || w.Body.String() != c.Response {
				t.Fatal(w.Code, w.Body.String())
			}

			if c.Status == 200 {
				var plan exploration.DashboardPlan
				if err := json.Unmarshal(w.Body.Bytes(), &plan); err != nil {
					t.Fatal(err)
				}
				expected := map[string]string{"Messaging-flat-keys-candidate-only": "all-spans", "Messaging-positive-consumer-hint": "messaging-consumer", "explicit-Messaging-without-positive-hint": "messaging-consumer", "explicit-all-spans-Messaging-candidate": "all-spans", "Messaging-destination-only-not-admitted": "all-spans", "both-positive-hints-HTTP-precedence": "http-server", "RPC-positive-hint-before-Messaging": "rpc-server", "Messaging-native-seconds-auto": "messaging-consumer", "Messaging-native-wrong-unit-no-admission": "all-spans", "Messaging-positive-producer-hint": "messaging-producer", "Messaging-explicit-producer": "messaging-producer", "DB-before-Messaging": "db-client", "Consumer-before-Producer": "messaging-consumer"}[c.Name]
				if expected == "" || plan.Profile != expected {
					t.Fatal("profile admission changed", c.Name, plan.Profile, expected)
				}
			}
			for k, v := range c.Headers {
				if w.Header().Get(k) != v {
					t.Fatal("header changed", k)
				}
			}
		})
	}
}

func TestDashboardMessagingProfileExactScopedRolePredicate(t *testing.T) {
	for _, role := range []struct{ profile, kind, opposite string }{{"messaging-consumer", "Consumer", "Producer"}, {"messaging-producer", "Producer", "Consumer"}} {
		fields := []exploration.Field{{Kind: "attribute", Key: "messaging.system"}}
		w := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardMessagingBody(role.profile, fields, true)})
		var plan exploration.DashboardPlan
		if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil || plan.Profile != role.profile {
			t.Fatal(w.Code, w.Body.String())
		}
		queries := 0
		for _, widget := range plan.Widgets {
			q := widget.Query
			if q == nil || q.Signal != "spans" {
				continue
			}
			queries++
			if q.Environment != "prod" || q.FromMS != 1000 || q.ToMS != 61000 {
				t.Fatal("scope changed", q)
			}
			for _, sample := range []struct {
				kind, system, namespace string
				present, want           bool
			}{
				{role.kind, "kafka", "commerce", true, true}, {role.opposite, "kafka", "commerce", true, false}, {"Server", "kafka", "commerce", true, false},
				{"Internal", "kafka", "commerce", true, false}, {role.kind, "", "commerce", true, false},
				{role.kind, "", "commerce", false, false}, {role.kind, "kafka", "billing", true, false},
			} {
				values := map[exploration.Field]string{{Kind: "column", Key: "service"}: "checkout", {Kind: "column", Key: "kind"}: sample.kind, {Kind: "resource", Key: "service.namespace"}: sample.namespace}
				if sample.present {
					values[exploration.Field{Kind: "attribute", Key: "messaging.system"}] = sample.system
				}
				if got := exploration.Match(q.Where, values); got != sample.want {
					t.Fatal("Messaging population mismatch", widget.Key, sample, got)
				}
				if !sample.want {
					values[exploration.Field{Kind: "attribute", Key: "messaging.destination"}] = "kafka"
					if exploration.Match(q.Where, values) {
						t.Fatal("destination alone admitted", sample)
					}
				}
			}
			if _, err := exploration.Compile(*q); err != nil {
				t.Fatal(err)
			}
		}
		if queries == 0 {
			t.Fatal("no executable Messaging span question")
		}
	}

}

func TestDashboardMessagingHintCannotActivateOmittedProfile(t *testing.T) {
	fields := []exploration.Field{{Kind: "attribute", Key: "messaging.system"}}
	var baseline string
	for _, profile := range []any{"", nil} {
		for _, hint := range []any{false, true} {
			w := planningWireResponse(t, planningWireCase{Method: "POST", Path: "/dashboards/plan", Body: dashboardMessagingBody(profile, fields, hint)})
			var plan exploration.DashboardPlan
			if w.Code != 200 || json.Unmarshal(w.Body.Bytes(), &plan) != nil || plan.Profile != "" || len(plan.Families) != 0 {
				t.Fatal("hint activated legacy", w.Code, w.Body.String())
			}
			if baseline == "" {
				baseline = w.Body.String()
			} else if baseline != w.Body.String() {
				t.Fatal("hint changed legacy bytes")
			}
		}
	}
}

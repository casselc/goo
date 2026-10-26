package exploration

import (
	"github.com/casselc/goo/measurements"
	"testing"
)

func TestDashboardPoolDomainDerivedIdentityAndPriority(t *testing.T) {
	in := dashboardHTTPInput()
	in.Signals = append(in.Signals, "metrics")
	pool := dashboardTestInstrument("z-pool", "db.client.connection.count", "stock", "{connection}")
	pool.Temporality = 2
	pool.SchemaURL = measurements.PoolConnectionsSchema
	pool.Definition.Source.SchemaURL = pool.SchemaURL
	observation := measurements.DiscoveryObservation{Resource: map[string]string{"service.name": "api", "service.instance.id": "one", "deployment.environment.name": "prod"}, Attributes: map[string]string{"db.client.connection.pool.name": "primary", "db.client.connection.state": "custom"}}
	p := measurements.PlanInstrument(pool, []measurements.DiscoveryObservation{observation})
	pool.CapabilityPlan = &p
	http := dashboardTestInstrument("a-http", "http.server.request.duration", "histogram", "s")
	in.Instruments = []measurements.Instrument{dashboardTestInstrument("b-gauge", "load", "gauge", "1"), pool, http}
	plan, e := PlanServiceDashboard(in, nil)
	if e != nil {
		t.Fatal(e)
	}
	if plan.NativeOptions[0].Instrument.Metric != http.Metric || plan.NativeOptions[1].Instrument.Metric != pool.Metric {
		t.Fatal("role priority or pool priority lost", plan.NativeOptions)
	}
	option := plan.NativeOptions[1]
	if option.Instrument.Definition.StockDomain != "nonnegative-integer" || option.Key != measurements.Hash(*option.Instrument.Definition) || pool.Definition.StockDomain != "" {
		t.Fatal("derived identity inconsistent or input mutated")
	}
	found := false
	for _, w := range plan.Widgets {
		if w.Family == "database-pool-connections" {
			found = true
			if w.Measurement == nil || w.Measurement.Definition.StockDomain != "nonnegative-integer" {
				t.Fatal(w)
			}
		}
	}
	if !found {
		t.Fatal("pool widget missing")
	}
	selected := []string{option.Key}
	in.NativeSelection = &selected
	again, e := PlanServiceDashboard(in, nil)
	if e != nil || !again.NativeOptions[1].Selected {
		t.Fatal("derived key cannot reselect", e)
	}
}

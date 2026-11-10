package exploration

import (
	"github.com/casselc/goo/measurements"
	"reflect"
	"testing"
)

func TestNativeHTTPClient392IndependentDefaultsAndSelection(t *testing.T) {
	for _, profile := range []string{"", "auto", "all-spans"} {
		for _, tempo := range []int{1, 2} {
			for _, schema := range []string{measurements.CurrentHTTPClientSchema, "", "https://custom.invalid/schema"} {
				in := dashboardHTTPInput()
				in.Profile = profile
				in.Signals = []string{"metrics"}
				in.Fields = nil
				in.HTTPServerObserved = false
				client := dashboardTestInstrument("client", measurements.CurrentHTTPClientDurationMetric, "histogram", "s")
				client.Temporality = tempo
				client.SchemaURL = schema
				client.Definition.Source.SchemaURL = schema
				in.Instruments = []measurements.Instrument{dashboardTestInstrument("aaa", "aaa.duration", "histogram", "s"), client, dashboardTestInstrument("zzz", "zzz.duration", "histogram", "s")}
				original := *client.Definition
				p, err := PlanServiceDashboard(in, nil)
				if err != nil {
					t.Fatal(err)
				}
				if len(p.NativeOptions) != 3 || nativeWidgets392(p) != 2 || p.NativeOptions[0].Instrument.Metric != client.Metric || !p.NativeOptions[0].Default {
					t.Fatal("client not preferred within bounded defaults", p)
				}
				found := false
				for _, widget := range p.Widgets {
					if widget.Measurement != nil && widget.Measurement.Definition.Source.Metric == client.Metric {
						found = true
						if widget.Family != "http-client" || widget.Grain != "per-stream" || !reflect.DeepEqual(widget.Measurement.Definition, original) {
							t.Fatal(widget)
						}
					}
					if widget.Query != nil {
						t.Fatal("metrics granted record query", widget)
					}
				}
				if !found {
					t.Fatal("missing native client")
				}
				for _, family := range p.Families {
					if family.Key == "http-client" {
						t.Fatal("metric granted outgoing record family", family)
					}
				}
				empty := []string{}
				in.NativeSelection = &empty
				p, err = PlanServiceDashboard(in, nil)
				if err != nil || nativeWidgets392(p) != 0 {
					t.Fatal("explicit no-read selection ignored", p, err)
				}
				only := []string{p.NativeOptions[0].Key}
				in.NativeSelection = &only
				p, err = PlanServiceDashboard(in, nil)
				if err != nil || nativeWidgets392(p) != 1 {
					t.Fatal(p, err)
				}
			}
		}
	}
}
func TestNativeHTTPClient392SecondsOnlyAndServerPriority(t *testing.T) {
	in := dashboardHTTPInput()
	in.Signals = []string{"metrics"}
	in.Fields = nil
	in.HTTPServerObserved = false
	server := dashboardTestInstrument("server", "http.server.request.duration", "histogram", "s")
	client := dashboardTestInstrument("client", measurements.CurrentHTTPClientDurationMetric, "histogram", "s")
	generic := dashboardTestInstrument("generic", "aaa.duration", "histogram", "s")
	in.Instruments = []measurements.Instrument{generic, client, server}
	p, err := PlanServiceDashboard(in, nil)
	if err != nil || p.NativeOptions[0].Instrument.Metric != server.Metric || p.NativeOptions[1].Instrument.Metric != client.Metric {
		t.Fatal(p, err)
	}
	client.Unit = "ms"
	client.Definition.Source.Unit = "ms"
	in.Instruments = []measurements.Instrument{client}
	p, err = PlanServiceDashboard(in, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range p.Widgets {
		if w.Family == "http-client" {
			t.Fatal("milliseconds acquired client profile", w)
		}
	}
	in.Profile = "http-client"
	if _, err = PlanServiceDashboard(in, nil); err == nil {
		t.Fatal("metrics-only acquired outgoing span lens")
	}
}

func TestNativeHTTPClient392ExplicitClientLensPriority(t *testing.T) {
	client := dashboardTestInstrument("zzz", measurements.CurrentHTTPClientDurationMetric, "histogram", "s")
	for _, others := range [][]measurements.Instrument{
		{dashboardTestInstrument("aaa", "aaa.duration", "histogram", "s"), dashboardTestInstrument("bbb", "bbb.duration", "histogram", "s")},
		{dashboardRuntimeInstrument("go.memory.used", "one", "heap"), dashboardRuntimeInstrument("go.memory.gc.goal", "one", "")},
	} {
		in := dashboardHTTPClient385()
		in.Signals = []string{"spans", "metrics"}
		in.Instruments = append(others, client)
		p, err := PlanServiceDashboard(in, nil)
		if err != nil {
			t.Fatal(err)
		}
		if p.Profile != "http-client" || len(p.NativeOptions) != 3 || p.NativeOptions[0].Instrument.Metric != client.Metric || !p.NativeOptions[0].Default || !p.NativeOptions[0].Selected {
			t.Fatal("active client profile did not prioritize its native observation", p)
		}
	}
}

func nativeWidgets392(p *DashboardPlan) int {
	n := 0
	for _, widget := range p.Widgets {
		if widget.Measurement != nil {
			n++
		}
	}
	return n
}

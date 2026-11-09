package workbench

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

func counterViewInput389(t *testing.T, domain, kind string) NativeViewInput {
	t.Helper()
	in := nativeViewInput371(t)
	in.Title = "Evaluated counter rates"
	in.Question.Definition.Kind = "counter"
	in.Question.Definition.CounterDomain = domain
	in.Question.Definition.Source.Unit = "{request}"
	in.Declaration.Selection.Kind = "evaluated-intervals"
	in.Declaration.Kind.Name = kind
	in.Declaration.Display = NativeViewDisplay{"time", "counter-rate", false}
	return in
}

func TestNativeCounterView389RoundTripAndDestinationAuthority(t *testing.T) {
	for _, domain := range []string{"", metrics.DecimalCounterDomain} {
		for _, kind := range []string{"timeline", "table"} {
			t.Run(domain+"/"+kind, func(t *testing.T) {
				in := counterViewInput389(t, domain, kind)
				store := Store{Dir: t.TempDir()}
				beforePlan, err := measurements.Compile(in.Question)
				if err != nil {
					t.Fatal(err)
				}
				saved, err := store.CreateNativeView(in)
				if err != nil {
					t.Fatal(err)
				}
				file := filepath.Join(store.Dir, "native-histogram-views", saved.ID+".json")
				beforeBytes, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				loaded, err := store.LoadNativeView(saved.ID)
				if err != nil || !reflect.DeepEqual(saved, loaded) {
					t.Fatal("counter roundtrip", err)
				}
				repeated, err := store.CreateNativeView(in)
				if err != nil || repeated.ID != saved.ID {
					t.Fatal("unstable immutable configuration", err)
				}
				page, err := store.NativeViews("")
				wantDomain := domain
				if wantDomain == "" {
					wantDomain = "nonnegative-integer"
				}
				if err != nil || len(page.Items) != 1 || page.Items[0] != (NativeViewSummary{ID: saved.ID, Title: in.Title, Kind: kind, Unit: "{request}", Family: "counter", Selection: "evaluated-intervals", CounterDomain: wantDomain}) {
					t.Fatal("counter catalog policy", page, err)
				}
				destination := in.Question
				destination.FromMS, destination.ToMS = 70000, 130000
				destination.FromNS, destination.ToNS = "70000000001", "129999999999"
				destination.Definition.Name, destination.Definition.Title = "another-counter", "Destination counter"
				destination.Definition.MaxGapMS = 1250
				destination.Definition.Environment = ""
				destination.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}}
				destination.Definition.Source.Services = []string{"destination"}
				destination.Definition.Source.Namespace = &identity.Namespace{Mode: "value", Value: ""}
				destination.Definition.Source.Attributes = map[string]string{"route": "destination"}
				destination.Definition.Source.ExactStream = &metrics.Stream{Resource: map[string]string{"service.name": "destination", "service.instance.id": "worker", "service.namespace": "", "deployment.environment.name": ""}, Scope: destination.Definition.Source.Scope, Version: destination.Definition.Source.Version, SchemaURL: destination.Definition.Source.SchemaURL, Name: destination.Definition.Source.Metric, Unit: destination.Definition.Source.Unit, Attributes: map[string]string{"route": "destination"}}
				destinationPlan, err := measurements.Compile(destination)
				if err != nil {
					t.Fatal(err)
				}
				reused, err := ReuseNativeView(*saved, destination)
				if err != nil || reused.ID != "" || !reflect.DeepEqual(reused.Question, destination) || reused.Declaration != in.Declaration {
					t.Fatal("donor policy or scope transferred", err)
				}
				afterPlan, err := measurements.Compile(reused.Question)
				if err != nil || !reflect.DeepEqual(destinationPlan, afterPlan) {
					t.Fatal("display changed destination calculation", err)
				}
				reused.Question.Definition.Source.Services[0] = "mutated"
				reused.Question.Definition.Source.Attributes["route"] = "mutated"
				reused.Question.Definition.Source.ExactStream.Resource["service.instance.id"] = "mutated"
				reused.Question.Definition.Source.ExactStream.Attributes["route"] = "mutated"
				reused.Question.Definition.EnvironmentSelector.Scope.Mode = "unreported"
				if destination.Definition.Source.Services[0] != "destination" || destination.Definition.Source.Attributes["route"] != "destination" || destination.Definition.Source.ExactStream.Resource["service.instance.id"] != "worker" || destination.Definition.EnvironmentSelector.Scope.Mode != "empty" {
					t.Fatal("destination aliases reused configuration")
				}
				afterBytes, err := os.ReadFile(file)
				if err != nil || !bytes.Equal(beforeBytes, afterBytes) {
					t.Fatal("saved configuration bytes changed", err)
				}
				unchangedPlan, err := measurements.Compile(in.Question)
				if err != nil || !reflect.DeepEqual(beforePlan, unchangedPlan) {
					t.Fatal("donor calculation changed", err)
				}
			})
		}
	}
}

func TestNativeCounterView389Refusals(t *testing.T) {
	for _, domain := range []string{"", metrics.DecimalCounterDomain} {
		in := counterViewInput389(t, domain, "timeline")
		store := Store{Dir: t.TempDir()}
		saved, err := store.CreateNativeView(in)
		if err != nil {
			t.Fatal(err)
		}
		for name, change := range map[string]func(*NativeViewInput){
			"grain":  func(v *NativeViewInput) { v.Question.Definition.Grain = "service" },
			"domain": func(v *NativeViewInput) { v.Question.Definition.CounterDomain = "nonnegative-integer" },
			"resource": func(v *NativeViewInput) {
				v.Question.Definition.Source.Resource = &measurements.ResourceBinding{Kind: "host", ID: "h"}
			},
			"bucket":              func(v *NativeViewInput) { v.Question.Definition.HistogramBucketMS = new(int64(1000)) },
			"missing unit":        func(v *NativeViewInput) { v.Question.Definition.Source.Unit = "" },
			"gauge selection":     func(v *NativeViewInput) { v.Declaration.Selection.Kind = "retained-observations" },
			"histogram selection": func(v *NativeViewInput) { v.Declaration.Selection.Kind = "export-interval" },
			"complete":            func(v *NativeViewInput) { v.Declaration.Display.RequiresComplete = true },
			"value":               func(v *NativeViewInput) { v.Declaration.Display.Value = "recorded-value" },
			"category":            func(v *NativeViewInput) { v.Declaration.Display.Category = "bucket" },
			"kind":                func(v *NativeViewInput) { v.Declaration.Kind.Name = "histogram" },
		} {
			t.Run(domain+"/"+name, func(t *testing.T) {
				v := counterViewInput389(t, domain, "timeline")
				change(&v)
				if v.Validate() == nil {
					t.Fatal("invalid counter declaration admitted")
				}
			})
		}
		for _, destinationDomain := range []string{"", metrics.DecimalCounterDomain} {
			q := in.Question
			q.Definition.CounterDomain = destinationDomain
			_, err := ReuseNativeView(*saved, q)
			if (err == nil) != (destinationDomain == domain) {
				t.Fatal("domain compatibility", domain, destinationDomain, err)
			}
		}
		q := in.Question
		q.Definition.Source.Unit = "requests"
		if _, err = ReuseNativeView(*saved, q); err == nil {
			t.Fatal("unit spelling converted")
		}
		q = in.Question
		q.Definition.Kind = "gauge"
		q.Definition.CounterDomain = ""
		if _, err = ReuseNativeView(*saved, q); err == nil {
			t.Fatal("counter intervals coerced into gauge samples")
		}
	}
}

func TestNativeCounterView389ClosedWire(t *testing.T) {
	for _, domain := range []string{"", metrics.DecimalCounterDomain} {
		raw, _ := json.Marshal(counterViewInput389(t, domain, "table"))
		for _, valid := range []string{string(raw), strings.Replace(string(raw), `"source":{`, `"counterDomain":"`+domain+`","source":{`, 1)} {
			if err := json.Unmarshal([]byte(valid), new(NativeViewInput)); err != nil {
				t.Fatal("counter wire", err)
			}
		}
		for _, bad := range []string{
			strings.Replace(string(raw), `"source":{`, `"counterDomain":null,"source":{`, 1),
			strings.Replace(string(raw), `"source":{`, `"CounterDomain":"","source":{`, 1),
			strings.Replace(string(raw), `,"requiresComplete":false`, "", 1),
			strings.Replace(string(raw), `"requiresComplete":false`, `"requiresComplete":null`, 1),
			strings.Replace(string(raw), `"source":{`, `"stockDomain":null,"source":{`, 1),
			strings.Replace(string(raw), `"source":{`, `"source":{"resource":null,`, 1),
		} {
			if json.Unmarshal([]byte(bad), new(NativeViewInput)) == nil {
				t.Fatal("invalid counter wire admitted", bad)
			}
		}
	}
	for _, family := range []string{"gauge", "histogram"} {
		in := nativeViewInput371(t)
		if family == "gauge" {
			in = gaugeViewInput386(t, "By")
		}
		raw, _ := json.Marshal(in)
		for _, domain := range []string{`""`, `null`, `"nonnegative-decimal"`} {
			bad := strings.Replace(string(raw), `"source":{`, `"counterDomain":`+domain+`,"source":{`, 1)
			if json.Unmarshal([]byte(bad), new(NativeViewInput)) == nil {
				t.Fatal("foreign counter field presence admitted", family, domain)
			}
		}
	}
}

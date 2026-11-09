package exploration

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

type observationFixture388 struct {
	Name     string                    `json:"name"`
	Result   *measurements.Result      `json:"result"`
	Stream   metrics.Stream            `json:"stream"`
	Signal   string                    `json:"signal"`
	AtNS     string                    `json:"atNS"`
	Expected *NativeObservationContext `json:"expected,omitempty"`
	Refused  bool                      `json:"refused,omitempty"`
}

func observationResult388(t *testing.T, namespace *identity.Namespace, from, to uint64, timestamps ...uint64) (*measurements.Result, metrics.Stream) {
	t.Helper()
	d := *dashboardTestInstrument("observation", "queue.depth", "gauge", "{item}").Definition
	d.Source.Namespace = namespace
	d.Source.Attributes = map[string]string{"metric.only": "not-record-identity"}
	d.MaxGapMS = 10_000
	stream := metrics.Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": "instance-one", "deployment.environment.name": "prod", "host.name": "not-a-record-predicate"}, ResourceSchemaURL: "https://opentelemetry.io/schemas/1.37.0", Scope: d.Source.Scope, Version: "", Name: d.Source.Metric, Unit: d.Source.Unit, Attributes: map[string]string{"metric.only": "not-record-identity"}}
	if namespace != nil && namespace.Mode == "value" {
		stream.Resource[identity.NamespaceKey] = namespace.Value
	}
	question := measurements.Question{Definition: d, FromMS: int64(from / 1_000_000), ToMS: int64((to + 999_999) / 1_000_000), FromNS: strconv.FormatUint(from, 10), ToNS: strconv.FormatUint(to, 10)}
	rows := make([]json.RawMessage, len(timestamps))
	for i, timestamp := range timestamps {
		rows[i], _ = json.Marshal(map[string]any{"resource": stream.Resource, "resource_schema_url": stream.ResourceSchemaURL, "scope": stream.Scope, "version": stream.Version, "schema_url": stream.SchemaURL, "name": stream.Name, "unit": stream.Unit, "attributes": stream.Attributes, "end_ns": strconv.FormatUint(timestamp, 10), "value": strconv.Itoa(i), "flags": 0})
	}
	result, err := measurements.Evaluate(question, rows, time.UnixMilli(int64(to/1_000_000)))
	if err != nil {
		t.Fatal(err)
	}
	return result, stream
}

func TestNativeObservationContext388(t *testing.T) {
	const lo, hi uint64 = 100_000_000_123, 300_000_000_877
	for _, namespace := range []*identity.Namespace{nil, {Mode: "value", Value: ""}, {Mode: "value", Value: "team"}} {
		for _, at := range []uint64{lo, lo + 1, lo + NativeObservationRadiusNS, 200_000_000_000, hi - NativeObservationRadiusNS, hi - 1} {
			result, stream := observationResult388(t, namespace, lo, hi, at)
			for _, signal := range []string{"logs", "spans"} {
				before, _ := json.Marshal(result)
				context, err := PlanNativeObservationContext(result, stream, signal, strconv.FormatUint(at, 10))
				if err != nil {
					t.Fatal(err)
				}
				after, _ := json.Marshal(result)
				if string(before) != string(after) {
					t.Fatal("projection changed retained evidence")
				}
				// The oracle uses signed bounded arithmetic independently of planner
				// distance comparisons. Source membership is half-open at nanoseconds.
				from, to := max(int64(lo), int64(at)-60_000_000_000), min(int64(hi), int64(at)+60_000_000_000)
				if context.FromNS != strconv.FormatInt(from, 10) || context.ToNS != strconv.FormatInt(to, 10) || context.AtNS != strconv.FormatUint(at, 10) || context.RadiusNS != "60000000000" || context.Grain != "observation-context" || context.Relation != "observation-context" {
					t.Fatal("observation radius or clipped bounds changed", context)
				}
				if context.Question.FromMS != from/1_000_000 || context.Question.ToMS != (to+999_999)/1_000_000 || len(context.Question.Where.All) != 5 {
					t.Fatal("record projection widened the resource or timestamp scope")
				}
				compiled, err := Compile(context.Question)
				if err != nil || strings.Contains(compiled.Rows, "metric.only") || strings.Contains(compiled.Rows, "host.name") {
					t.Fatal("metric-specific selector escaped into ordinary records", err)
				}
				fields := map[Field]string{{Kind: "column", Key: "service"}: "api", {Kind: "resource", Key: "service.instance.id"}: "instance-one"}
				if namespace != nil {
					fields[Field{Kind: "resource", Key: identity.NamespaceKey}] = namespace.Value
				}
				for _, timestamp := range []int64{from - 1, from, int64(at), to - 1, to} {
					fields[Field{Kind: "column", Key: "timestamp_ns"}] = strconv.FormatInt(timestamp, 10)
					if got, want := Match(context.Question.Where, fields), timestamp >= from && timestamp < to; got != want {
						t.Fatalf("timestamp %d matched %v, want %v", timestamp, got, want)
					}
				}
				fields[Field{Kind: "column", Key: "timestamp_ns"}] = strconv.FormatUint(at, 10)
				fields[Field{Kind: "resource", Key: "service.instance.id"}] = "foreign-instance"
				if Match(context.Question.Where, fields) {
					t.Fatal("foreign instance matched")
				}
			}
		}
	}
	// Near-epoch clipping must not underflow unsigned arithmetic.
	result, stream := observationResult388(t, nil, 0, 20_000_000_000, 1)
	context, err := PlanNativeObservationContext(result, stream, "logs", "1")
	if err != nil || context.FromNS != "0" || context.ToNS != "20000000000" {
		t.Fatal("epoch clipping failed", context, err)
	}
}

func TestNativeObservationMembership388(t *testing.T) {
	result, stream := observationResult388(t, nil, 100_000_000_123, 300_000_000_877, 100_000_000_123, 290_000_000_001)
	copyResult := func() *measurements.Result {
		bytes, _ := json.Marshal(result)
		var copy measurements.Result
		if err := json.Unmarshal(bytes, &copy); err != nil {
			t.Fatal(err)
		}
		return &copy
	}
	for name, change := range map[string]func(*measurements.Result){
		"missing-series":   func(r *measurements.Result) { r.Gauges.Series = nil },
		"duplicate-series": func(r *measurements.Result) { r.Gauges.Series = append(r.Gauges.Series, r.Gauges.Series[0]) },
		"duplicate-sample": func(r *measurements.Result) {
			r.Gauges.Series[0].Samples = append(r.Gauges.Series[0].Samples, r.Gauges.Series[0].Samples[0])
		},
		"invented-time":    func(r *measurements.Result) { r.Gauges.Series[0].Samples[0].AtNS = "150000000000" },
		"changed-basis":    func(r *measurements.Result) { r.Basis = strings.Repeat("a", 64) },
		"changed-question": func(r *measurements.Result) { r.Question.Definition.Environment = "other" },
		"changed-state":    func(r *measurements.Result) { r.Gauges.Series[0].Samples[0].State = "observed" },
		"fabricated-value": func(r *measurements.Result) { r.Gauges.Series[0].Samples[0].Value = new(float64(99)) },
	} {
		t.Run(name, func(t *testing.T) {
			r := copyResult()
			change(r)
			if _, err := PlanNativeObservationContext(r, stream, "logs", "100000000123"); err == nil {
				t.Fatal("corrupt retained evidence authorized context")
			}
		})
	}
	for _, at := range []string{"", "0100000000123", "100000000124", "300000000877", "-1", "18446744073709551616"} {
		if _, err := PlanNativeObservationContext(result, stream, "logs", at); err == nil {
			t.Fatalf("unreported/invalid timestamp %q accepted", at)
		}
	}
	for _, signal := range []string{"", "metrics", "Logs"} {
		if _, err := PlanNativeObservationContext(result, stream, signal, "100000000123"); err == nil {
			t.Fatal("unsupported signal accepted")
		}
	}
	for _, key := range []string{"service.name", "service.instance.id", "deployment.environment.name", identity.NamespaceKey, "host.name"} {
		foreign := stream
		foreign.Resource = map[string]string{}
		for k, v := range stream.Resource {
			foreign.Resource[k] = v
		}
		foreign.Resource[key] = "foreign"
		if _, err := PlanNativeObservationContext(result, foreign, "logs", "100000000123"); err == nil {
			t.Fatalf("foreign full stream identity %q accepted", key)
		}
	}
	for name, mutate := range map[string]func(*metrics.Stream){
		"scope":               func(s *metrics.Stream) { s.Scope = "foreign" },
		"version":             func(s *metrics.Stream) { s.Version = "foreign" },
		"resource-schema-url": func(s *metrics.Stream) { s.ResourceSchemaURL = "foreign" },
		"schema-url":          func(s *metrics.Stream) { s.SchemaURL = "foreign" },
		"name":                func(s *metrics.Stream) { s.Name = "foreign" },
		"unit":                func(s *metrics.Stream) { s.Unit = "foreign" },
		"metric-attributes":   func(s *metrics.Stream) { s.Attributes = map[string]string{"metric.only": "foreign"} },
	} {
		foreign := stream
		mutate(&foreign)
		if _, err := PlanNativeObservationContext(result, foreign, "logs", "100000000123"); err == nil {
			t.Fatalf("foreign stream %s accepted merely because record resource projection matched", name)
		}
	}
	if _, err := PlanNativeObservationContext(nil, stream, "logs", "100000000123"); err == nil {
		t.Fatal("missing retained result accepted")
	}
	if _, err := PlanNativeRecordContext(result.Question, stream, "logs", "100000000123", "160000000123"); err == nil {
		t.Fatal("interval planner's strict gauge point semantics relaxed")
	}
}

func TestNativeObservationUnknownAndEnvironment388(t *testing.T) {
	base, stream := observationResult388(t, nil, 100_000_000_123, 300_000_000_877, 200_000_000_000)
	for _, policy := range []string{"current-only", "effective"} {
		for _, absent := range []bool{false, true} {
			parent := base.Question
			parent.Definition.Environment = ""
			parent.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: policy, Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
			var row map[string]any
			if err := json.Unmarshal(base.Rows[0], &row); err != nil {
				t.Fatal(err)
			}
			if absent {
				row["flags"] = 1
			}
			raw, _ := json.Marshal(row)
			result, err := measurements.Evaluate(parent, []json.RawMessage{raw}, time.UnixMilli(parent.ToMS))
			if err != nil {
				t.Fatal(err)
			}
			got, err := PlanNativeObservationContext(result, stream, "spans", "200000000000")
			if err != nil || got.Question.Environment != "" || got.Question.EnvironmentSelector.Policy != policy {
				t.Fatal("typed environment policy was not preserved", got, err)
			}
			if absent && (result.Gauges.Series[0].Samples[0].Value != nil || result.Gauges.Series[0].Samples[0].State != "unknown") {
				t.Fatal("no-recorded-value observation gained a magnitude")
			}
			got.Question.EnvironmentSelector.Scope.Value = "changed"
			if result.Question.Definition.EnvironmentSelector.Scope.Value != "prod" {
				t.Fatal("projected environment selector aliases retained evidence")
			}
		}
	}
}

func TestNativeObservationFixtures388(t *testing.T) {
	fixtures := []observationFixture388{}
	for i, namespace := range []*identity.Namespace{nil, {Mode: "value", Value: ""}, {Mode: "value", Value: "team"}} {
		result, stream := observationResult388(t, namespace, 100_000_000_123, 300_000_000_877, 100_000_000_123, 200_000_000_000, 290_000_000_001)
		for _, signal := range []string{"logs", "spans"} {
			for _, at := range []string{"100000000123", "200000000000", "290000000001", "100000000124"} {
				context, err := PlanNativeObservationContext(result, stream, signal, at)
				fixtures = append(fixtures, observationFixture388{Name: strconv.Itoa(i) + "-" + signal + "-" + at, Result: result, Stream: stream, Signal: signal, AtNS: at, Expected: context, Refused: err != nil})
			}
		}
	}
	path := "../tests/fixtures/native-observation-context-388.json"
	if os.Getenv("GOO_UPDATE_NATIVE_OBSERVATION_388_FIXTURES") == "1" {
		bytes, _ := json.MarshalIndent(fixtures, "", "  ")
		if err := os.WriteFile(path, append(bytes, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	frozenBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var frozen []observationFixture388
	if err := json.Unmarshal(frozenBytes, &frozen); err != nil {
		t.Fatal("invalid shared observation fixtures", err)
	}
	expectedJSON, _ := json.Marshal(fixtures)
	frozenJSON, _ := json.Marshal(frozen)
	if !bytes.Equal(expectedJSON, frozenJSON) {
		t.Fatal("shared observation fixtures drifted")
	}
}

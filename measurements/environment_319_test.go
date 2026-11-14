package measurements

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/metrics"
	"gopkg.in/yaml.v3"
)

func selectors319() []identity.EnvironmentSelector {
	return []identity.EnvironmentSelector{
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	}
}
func gauge319(selector identity.EnvironmentSelector) Definition {
	return Definition{APIVersion: Version, Name: "typed-gauge", Title: "Typed gauge", Kind: "gauge", EnvironmentSelector: &selector, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: Binding{Services: []string{"api"}, Namespace: &identity.Namespace{Mode: "value", Value: "team"}, Scope: "sdk", Metric: "load", Unit: "items"}}
}
func rows319(d Definition) []metricRow {
	resources := []map[string]string{{"deployment.environment.name": "prod"}, {"deployment.environment": "prod"}, {"deployment.environment.name": "other", "deployment.environment": "prod"}, {"deployment.environment.name": "", "deployment.environment": "prod"}, {}}
	values := []string{"123", "456", "789", "234", "567"}
	rows := []metricRow{}
	for index, environment := range resources {
		row := gaugeRow(d, "shared-instance", 150000, values[index])
		delete(row.Resource, "deployment.environment.name")
		for key, value := range environment {
			row.Resource[key] = value
		}
		row.Resource[identity.NamespaceKey] = "team"
		rows = append(rows, row)
	}
	return rows
}
func TestTypedGaugeMembershipAndRawStreamIdentity319(t *testing.T) {
	expected := [][]int{{0}, {0, 1}, {3}, {4}}
	for index, selector := range selectors319() {
		t.Run(selector.Policy+selector.Scope.Mode, func(t *testing.T) {
			d := gauge319(selector)
			q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
			plan, err := Compile(q)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(plan.SQL, "'deployment.environment'") != (selector.Policy == "effective") {
				t.Fatal("SQL policy mismatch")
			}
			all := rows319(d)
			var admitted []json.RawMessage
			wantValues := []float64{}
			for _, which := range expected[index] {
				admitted = append(admitted, raw(all[which]))
				value := []float64{123, 456, 789, 234, 567}[which]
				wantValues = append(wantValues, value)
			}
			result, err := Evaluate(q, admitted, time.UnixMilli(170000))
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Gauges.Series) != len(admitted) {
				t.Fatal("raw resources were merged into an effective-label stream")
			}
			got := []float64{}
			for _, series := range result.Gauges.Series {
				if len(series.Samples) != 1 || series.Samples[0].Value == nil {
					t.Fatal("gauge value unavailable")
				}
				got = append(got, *series.Samples[0].Value)
			}
			sort.Float64s(got)
			sort.Float64s(wantValues)
			if !reflect.DeepEqual(got, wantValues) {
				t.Fatal("native values changed", got, wantValues)
			}
			selection, _ := d.EnvironmentSelection()
			for _, row := range all {
				if !selection.Matches(row.Resource) {
					if _, err := Evaluate(q, []json.RawMessage{raw(row)}, time.UnixMilli(170000)); err == nil {
						t.Fatal("off-scope row published")
					}
				}
			}
			good := all[expected[index][0]]
			exact := d
			exact.Source.ExactStream = &metrics.Stream{Resource: good.Resource, Scope: good.Scope, Name: good.Name, Unit: good.Unit}
			if err := exact.Validate(); err != nil {
				t.Fatal("matching exact stream refused", err)
			}
			for _, row := range all {
				if !selection.Matches(row.Resource) {
					bad := exact
					copy := *exact.Source.ExactStream
					copy.Resource = row.Resource
					bad.Source.ExactStream = &copy
					if bad.Validate() == nil {
						t.Fatal("off-scope exact stream admitted")
					}
					break
				}
			}
		})
	}
}
func TestTypedGaugeClosedProfileAndWire319(t *testing.T) {
	for _, kind := range []string{"stock", "state", "lifecycle"} {
		d := gauge319(selectors319()[1])
		d.Kind = kind
		if _, err := Compile(Question{Definition: d, FromMS: 100000, ToMS: 160000}); err == nil {
			t.Fatal("typed non-gauge compiled", kind)
		}
	}
	d := gauge319(selectors319()[1])
	d.Source.Resource = &ResourceBinding{Kind: "host", ID: "host"}
	if d.Validate() == nil {
		t.Fatal("typed resource authority admitted")
	}
	d = gauge319(selectors319()[1])
	encoded, _ := json.Marshal(d)
	for _, bad := range []string{strings.Replace(string(encoded), `"environmentSelector":`, `"environment":"prod","environmentSelector":`, 1), strings.Replace(string(encoded), `"environmentSelector":`, `"Environment":null,"environmentSelector":`, 1), strings.Replace(string(encoded), `"environmentSelector":{`, `"environmentSelector":null,"unknown":{`, 1)} {
		var parsed Definition
		err := json.Unmarshal([]byte(bad), &parsed)
		if err == nil {
			err = parsed.Validate()
		}
		if err == nil {
			t.Fatal("invalid native wire admitted")
		}
	}
	for _, selector := range selectors319() {
		d = gauge319(selector)
		text, err := yaml.Marshal(d)
		if err != nil {
			t.Fatal(err)
		}
		parsed, err := Read(string(text))
		if err != nil {
			t.Fatal(err)
		}
		before, _ := d.Revision()
		after, _ := parsed.Revision()
		if before != after || !d.SameEnvironment(parsed) {
			t.Fatal("YAML lost native policy identity")
		}
	}
	text, _ := yaml.Marshal(gauge319(selectors319()[1]))
	for _, replacement := range []string{"value: null", "value: 42", "value: true", "value: !custom prod"} {
		if _, err := Read(strings.Replace(string(text), "value: prod", replacement, 1)); err == nil {
			t.Fatal("YAML selector coercion admitted", replacement)
		}
	}
	legacy := gauge319(selectors319()[0])
	legacy.Environment = "prod"
	legacy.EnvironmentSelector = nil
	text, _ = yaml.Marshal(legacy)
	if _, err := Read(string(text) + "environmentSelector: null\n"); err == nil {
		t.Fatal("YAML null selector silently became legacy scope")
	}
}

type environmentTransport319 func(*http.Request) (*http.Response, error)

func (f environmentTransport319) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestTypedGaugeRunDetachesBeforeIO319(t *testing.T) {
	d := gauge319(selectors319()[1])
	q := Question{Definition: d, FromMS: 100000, ToMS: 160000}
	expected, err := q.detached()
	if err != nil {
		t.Fatal(err)
	}
	rows := rows319(d)
	payload := string(raw(rows[1])) + "\n"
	calls := 0
	source := Source{Endpoint: "http://source", Client: &http.Client{Transport: environmentTransport319(func(r *http.Request) (*http.Response, error) {
		calls++
		q.Definition.EnvironmentSelector.Scope.Value = "changed"
		q.Definition.Source.Services[0] = "changed"
		q.Definition.Source.Namespace.Value = "changed"
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload))}, nil
	})}}
	result, err := source.Run(t.Context(), q, nil)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || !reflect.DeepEqual(result.Question, expected) || *result.Gauges.Series[0].Samples[0].Value != 456 {
		t.Fatal("native read/evaluation scope changed during IO")
	}
	q.Definition.EnvironmentSelector.Policy = "current-only"
	if !reflect.DeepEqual(result.Question, expected) {
		t.Fatal("result scope aliases caller")
	}
}
func TestTypedNativeDiscovery319(t *testing.T) {
	d := gauge319(selectors319()[1])
	q := DiscoveryQuestion{Service: "api", Namespace: d.Source.Namespace, FromMS: 100000, ToMS: 160000}.WithEnvironmentFrom(d)
	resource := rows319(d)[1].Resource
	observations := []json.RawMessage{}
	for _, kind := range []string{"gauge", "counter", "histogram", "stock"} {
		unit := "items"
		if kind == "histogram" {
			unit = "ms"
		}
		tempo := 1
		monotonic := true
		if kind == "stock" {
			tempo = 2
			monotonic = false
		}
		observations = append(observations, raw(map[string]any{"kind": kind, "scope": "sdk", "version": "", "schemaURL": "", "metric": "load-" + kind, "unit": unit, "temporality": tempo, "monotonic": monotonic, "dimensions": []string{}, "resource": resource, "attributes": map[string]string{}}))
	}
	result, err := discoverRows(q, observations)
	if err != nil {
		t.Fatal(err)
	}
	for _, instrument := range result.Instruments {
		if typedServiceKind(instrument.Kind) {
			if !instrument.Supported || instrument.Definition == nil || !instrument.Definition.SameEnvironment(d) {
				t.Fatal("typed native prototype lost scope")
			}
		} else if instrument.Supported || instrument.Definition != nil || instrument.UnsupportedReason == "" {
			t.Fatal("typed unsupported family became executable")
		}
	}
	bad := append([]json.RawMessage(nil), observations...)
	other := rows319(d)[2].Resource
	bad[0] = raw(map[string]any{"kind": "gauge", "scope": "sdk", "metric": "load", "unit": "items", "resource": other})
	if _, err := discoverRows(q, bad); err == nil {
		t.Fatal("discovery published off-scope native metadata")
	}
}

func TestTypedDiscoveryDetachesBeforeIO319(t *testing.T) {
	definition := gauge319(selectors319()[1])
	question := DiscoveryQuestion{Service: "api", Namespace: definition.Source.Namespace, FromMS: 100000, ToMS: 160000}.WithEnvironmentFrom(definition)
	expected := question.clone()
	resource := rows319(definition)[1].Resource
	payload := string(raw(map[string]any{"kind": "gauge", "scope": "sdk", "version": "", "schemaURL": "", "metric": "load", "unit": "items", "temporality": 0, "monotonic": false, "dimensions": []string{}, "resource": resource, "attributes": map[string]string{}})) + "\n"
	source := Source{Endpoint: "http://source", Client: &http.Client{Transport: environmentTransport319(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if response, handled := sourceCapabilityResponse431(string(body)); handled {
			return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(response))}, nil
		}
		question.EnvironmentSelector.Policy = "current-only"
		question.Namespace.Value = "changed"
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(payload))}, nil
	})}}
	result, err := source.Discover(t.Context(), question)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(result.Question, expected) || len(result.Instruments) != 1 || result.Instruments[0].Definition == nil || !result.Instruments[0].Definition.SameEnvironment(definition) {
		t.Fatal("discovery scope/prototype changed during IO")
	}
	result.Question.EnvironmentSelector.Policy = "current-only"
	if result.Instruments[0].Definition.EnvironmentSelector.Policy != "effective" {
		t.Fatal("prototype aliases returned discovery question")
	}
}

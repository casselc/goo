package exploration

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
)

func serviceSelectors318() []identity.EnvironmentSelector {
	return []identity.EnvironmentSelector{
		{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}},
		{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}},
	}
}
func TestServiceScopeSQLAndNativeBoundary318(t *testing.T) {
	for _, selector := range serviceSelectors318() {
		t.Run(selector.Policy+selector.Scope.Mode, func(t *testing.T) {
			q := InventoryQuestion{EnvironmentSelector: &selector, FromMS: 100000, ToMS: 160000, Namespace: &identity.Namespace{Mode: "value"}}
			selected, err := q.EnvironmentSelection()
			if err != nil {
				t.Fatal(err)
			}
			predicate, bindings, err := environmentSelectionPredicate(selected)
			if err != nil {
				t.Fatal(err)
			}
			inventory, err := CompileInventory(q)
			if err != nil {
				t.Fatal(err)
			}
			discovery, err := CompileServiceDiscovery(q)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(inventory.Groups, predicate) || strings.Count(discovery.Count, predicate) != 5 || !strings.Contains(discovery.Groups, predicate) {
				t.Fatal("record/metric-presence scope lowering diverged")
			}
			for _, plan := range []Plan{inventory, discovery} {
				if plan.Parameters["environment"] != bindings["environment"] {
					t.Fatal("environment binding changed")
				}
				if _, named := plan.Parameters["environment"]; named != (selector.Scope.Mode == "named") {
					t.Fatal("nameless selector acquired a named binding")
				}
			}
			input := DashboardRequest{Scope: q, Service: "api", Signals: []string{"spans", "logs", "metrics"}, Profile: "auto", HTTPServerObserved: true, Fields: []Field{{Kind: "attribute", Key: "http.request.method"}}}
			declarations := []DashboardDeclaration{{Environment: "prod", Service: "api", Name: "declared", Title: "Declared", Revision: "rev"}}
			plan, err := PlanServiceDashboard(input, declarations)
			if err != nil {
				t.Fatal(err)
			}
			current := selector.Policy == "current-only"
			if (plan.NativeCapability == nil) != current || !current && plan.NativeCapability.State != "gauge-only" {
				t.Fatal("native capability does not preserve policy")
			}
			if (len(plan.Declarations) > 0) != current {
				t.Fatal("declarations matched a different environment policy")
			}
			if plan.Profile != "http-server" || !reflect.DeepEqual(plan.Scope, q) {
				t.Fatal("typed record profile lost its checked scope")
			}
			count := 0
			for _, widget := range plan.Widgets {
				if widget.Query != nil {
					count++
					if widget.Query.Environment != "" || !reflect.DeepEqual(widget.Query.EnvironmentSelector, &selector) {
						t.Fatal("record widget lost typed selection")
					}
					if _, err := Compile(*widget.Query); err != nil {
						t.Fatal(err)
					}
				}
				if !current && widget.Measurement != nil {
					t.Fatal("unsupported native evaluator was acquired")
				}
			}
			if count == 0 {
				t.Fatal("typed record profile generated no questions")
			}
			if !current {
				input.Instruments = []measurements.Instrument{{Metric: "untrusted", Supported: true}}
				if _, err := PlanServiceDashboard(input, nil); err == nil {
					t.Fatal("unsupported native instrument admitted")
				}
				input.Instruments = nil
				selection := []string{"native"}
				input.NativeSelection = &selection
				if _, err := PlanServiceDashboard(input, nil); err == nil {
					t.Fatal("unsupported native selection admitted")
				}
			}
		})
	}
}

type serviceScopeTransport318 func(*http.Request) (*http.Response, error)

func (f serviceScopeTransport318) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestServiceScopeQuestionOwnership318(t *testing.T) {
	for _, discover := range []bool{false, true} {
		t.Run(map[bool]string{false: "inventory", true: "discovery"}[discover], func(t *testing.T) {
			selector := serviceSelectors318()[1]
			q := InventoryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "value"}, FromMS: 100000, ToMS: 160000}
			want := q.clone()
			reads := 0
			source := Source{Endpoint: "http://unused", Client: &http.Client{Transport: serviceScopeTransport318(func(r *http.Request) (*http.Response, error) {
				reads++
				q.EnvironmentSelector.Scope.Value = "changed-during-read"
				q.Namespace.Value = "changed"
				return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
			})}}
			var actual InventoryQuestion
			if discover {
				result, err := source.DiscoverServices(t.Context(), q)
				if err != nil {
					t.Fatal(err)
				}
				actual = result.Question
			} else {
				result, err := source.Inventory(t.Context(), q)
				if err != nil {
					t.Fatal(err)
				}
				actual = result.Question
			}
			if !reflect.DeepEqual(actual, want) || reads != map[bool]int{false: 1, true: 2}[discover] {
				t.Fatal("returned scope aliases caller or read budget changed")
			}
			q.EnvironmentSelector.Scope.Value = "changed-after-return"
			if !reflect.DeepEqual(actual, want) {
				t.Fatal("returned selector changed after return")
			}
		})
	}
	selector := serviceSelectors318()[0]
	q := InventoryQuestion{EnvironmentSelector: &selector, FromMS: 100000, ToMS: 160000}
	plan, err := PlanServiceDashboard(DashboardRequest{Scope: q, Service: "api", Signals: []string{"spans"}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	selector.Scope.Value = "changed"
	if plan.Scope.EnvironmentSelector.Scope.Value != "prod" {
		t.Fatal("plan scope aliases caller")
	}
	plan.Scope.EnvironmentSelector.Scope.Value = "changed-plan"
	for _, widget := range plan.Widgets {
		if widget.Query != nil && widget.Query.EnvironmentSelector.Scope.Value != "prod" {
			t.Fatal("widget aliases plan scope")
		}
	}
}
func TestServiceScopeStrictWire318(t *testing.T) {
	for _, raw := range []string{`{"environment":null}`, `{"environment":"prod","environmentSelector":null}`, `{"environment":"prod","unknown":1}`, `{"environmentSelector":{"version":"goo/environment-selector/v1","policy":"effective","scope":{"mode":"unreported","unknown":1}}}`} {
		var q InventoryQuestion
		if json.Unmarshal([]byte(raw), &q) == nil {
			t.Fatal("invalid closed wire accepted", raw)
		}
	}
}

func TestTypedInventoryCaptureNormalization318(t *testing.T) {
	selector := serviceSelectors318()[1]
	question := CaptureQuestion{Kind: "inventory", Inventory: &InventoryQuestion{EnvironmentSelector: &selector, Namespace: &identity.Namespace{Mode: "missing"}, FromMS: 100000, ToMS: 160000}}
	normalized := question.Normalized()
	selector.Scope.Value = "changed"
	question.Inventory.Namespace.Mode = "value"
	if normalized.Inventory.EnvironmentSelector.Scope.Value != "prod" || normalized.Inventory.Namespace.Mode != "missing" {
		t.Fatal("inventory capture normalization aliases typed caller scope")
	}
	scope, err := normalized.Scope()
	if err != nil || scope.EnvironmentSelector == nil || scope.EnvironmentSelector.Policy != "effective" {
		t.Fatal("inventory capture scope lost policy", err)
	}
}

func TestTypedCurrentScopeNativeProjection318(t *testing.T) {
	selector := serviceSelectors318()[0]
	definition := measurements.Definition{APIVersion: measurements.Version, Name: "queue", Title: "Queue observations", Kind: "gauge", Environment: "prod", LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "sdk", Metric: "queue.depth", Unit: "1"}}
	typedScope := InventoryQuestion{EnvironmentSelector: &selector, FromMS: 100000, ToMS: 160000}
	legacyScope := typedScope
	legacyScope.EnvironmentSelector = nil
	legacyScope.Environment = "prod"
	for _, compile := range []func(InventoryQuestion) (Plan, error){CompileInventory, CompileServiceDiscovery} {
		a, e := compile(typedScope)
		if e != nil {
			t.Fatal(e)
		}
		b, e := compile(legacyScope)
		if e != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("typed-current and legacy service plans differ", e)
		}
	}
	input := DashboardRequest{Scope: typedScope, Service: "api", Signals: []string{"metrics"}, Instruments: []measurements.Instrument{{Kind: "gauge", Scope: "sdk", Metric: "queue.depth", Unit: "1", Supported: true, Definition: &definition}}}
	plan, err := PlanServiceDashboard(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan.NativeCapability != nil || len(plan.NativeOptions) != 1 || plan.Scope.Environment != "" || !reflect.DeepEqual(plan.Scope.EnvironmentSelector, &selector) {
		t.Fatal("equivalent native projection changed the original typed scope")
	}
	found := false
	for _, widget := range plan.Widgets {
		if widget.Measurement != nil {
			found = true
			if widget.Measurement.Definition.Environment != "prod" {
				t.Fatal("native projection did not retain current-only name")
			}
		}
	}
	if !found {
		t.Fatal("equivalent typed-current scope lost native support")
	}
	input.Scope.EnvironmentSelector = &serviceSelectors318()[1]
	if _, err := PlanServiceDashboard(input, nil); err == nil {
		t.Fatal("effective name converted to native current-only name")
	}
}

func TestTypedGaugeDashboardNativePolicy319(t *testing.T) {
	selector := serviceSelectors318()[1]
	definition := measurements.Definition{APIVersion: measurements.Version, Name: "typed-gauge", Title: "Typed gauge", Kind: "gauge", EnvironmentSelector: &selector, LookbackMinutes: 1, MaxGapMS: 60000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"api"}, Scope: "sdk", Metric: "load", Unit: "items"}}
	input := DashboardRequest{Scope: InventoryQuestion{EnvironmentSelector: &selector, FromMS: 100000, ToMS: 160000}, Service: "api", Signals: []string{"metrics"}, Instruments: []measurements.Instrument{{Kind: "gauge", Scope: "sdk", Metric: "load", Unit: "items", Supported: true, Definition: &definition}}}
	plan, err := PlanServiceDashboard(input, nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan.NativeCapability == nil || plan.NativeCapability.State != "gauge-only" || len(plan.NativeOptions) != 1 {
		t.Fatal("typed gauge native option unavailable")
	}
	found := false
	for _, widget := range plan.Widgets {
		if widget.Measurement != nil {
			found = true
			if widget.Measurement.Definition.Environment != "" || !reflect.DeepEqual(widget.Measurement.Definition.EnvironmentSelector, &selector) {
				t.Fatal("dashboard gauge lost checked policy")
			}
			if _, err := measurements.Compile(*widget.Measurement); err != nil {
				t.Fatal(err)
			}
		}
	}
	if !found {
		t.Fatal("typed native value question missing")
	}
	// The equal label does not admit a current-only definition in an effective scope.
	definition.EnvironmentSelector = nil
	definition.Environment = "prod"
	if _, err := PlanServiceDashboard(input, nil); err == nil {
		t.Fatal("current-only gauge silently substituted for effective policy")
	}
}

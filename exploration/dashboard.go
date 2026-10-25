package exploration

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"sort"
	"strings"
	"unicode/utf8"
)

const DashboardVersion = "goo/service-dashboard/v2"
const DashboardNativeLimit = 2

// Inputs are supplied observations, not authenticated source evidence or health.
type DashboardRequest struct {
	HTTPServerObserved bool                      `json:"httpServerObserved,omitempty"`
	Profile            string                    `json:"profile,omitempty"`
	Scope              InventoryQuestion         `json:"scope"`
	Service            string                    `json:"service"`
	Signals            []string                  `json:"signals"`
	Fields             []Field                   `json:"fields"`
	Instruments        []measurements.Instrument `json:"instruments"`
	// Nil uses defaults; an explicit empty list selects no native reads.
	NativeSelection *[]string `json:"nativeSelection,omitempty"`
}
type DashboardNativeOption struct {
	Key        string                  `json:"key"`
	Instrument measurements.Instrument `json:"instrument"`
	Selected   bool                    `json:"selected"`
	Default    bool                    `json:"default"`
}
type DashboardDeclaration struct {
	Namespace        *identity.Namespace  `json:"namespace,omitempty"`
	Instance         string               `json:"instance,omitempty"`
	Scope            string               `json:"scope,omitempty"`
	Version          string               `json:"version,omitempty"`
	SchemaURL        string               `json:"schemaURL,omitempty"`
	Operations       []DashboardOperation `json:"operations,omitempty"`
	Match            string               `json:"match,omitempty"`
	Name             string               `json:"name"`
	Title            string               `json:"title"`
	Revision         string               `json:"revision"`
	ManifestRevision string               `json:"manifestRevision"`
	Service          string               `json:"service"`
	Environment      string               `json:"environment"`
}
type DashboardOperation struct {
	Name string `json:"name"`
	Role string `json:"role"`
}
type DashboardWidget struct {
	Family      string                 `json:"family,omitempty"`
	Key         string                 `json:"key"`
	Title       string                 `json:"title"`
	State       string                 `json:"state"`
	Reason      string                 `json:"reason,omitempty"`
	Grain       string                 `json:"grain"`
	Query       *Query                 `json:"query,omitempty"`
	Measurement *measurements.Question `json:"measurement,omitempty"`
}
type DashboardFamily struct {
	Key      string `json:"key"`
	Title    string `json:"title"`
	State    string `json:"state"`
	Grain    string `json:"grain"`
	Reason   string `json:"reason,omitempty"`
	RuleSet  string `json:"ruleSet,omitempty"`
	Selected bool   `json:"selected"`
}
type DashboardPlan struct {
	Profile       string                  `json:"profile,omitempty"`
	Families      []DashboardFamily       `json:"families,omitempty"`
	Version       string                  `json:"version"`
	Scope         InventoryQuestion       `json:"scope"`
	Service       string                  `json:"service"`
	Widgets       []DashboardWidget       `json:"widgets"`
	Declarations  []DashboardDeclaration  `json:"declarations"`
	Warnings      []string                `json:"warnings"`
	NativeOptions []DashboardNativeOption `json:"nativeOptions"`
	NativeLimit   int                     `json:"nativeLimit"`
}

func DashboardRecordQuestion(scope InventoryQuestion, service, signal string) Query {
	q := Query{Signal: signal, Environment: scope.Environment, FromMS: scope.FromMS, ToMS: scope.ToMS, Limit: 100, Where: &Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: service}}
	if scope.Namespace != nil {
		q.Where = &Predicate{All: []Predicate{*q.Where, *NamespacePredicate(scope.Namespace)}}
	}
	return q
}

// PlanServiceDashboard performs no source reads. Every executable widget uses an
// existing typed evaluator; no SQL, health policy or fleet aggregation is inferred.
func PlanServiceDashboard(in DashboardRequest, declarations []DashboardDeclaration) (*DashboardPlan, error) {
	if in.Profile != "" && in.Profile != "auto" && in.Profile != "all-spans" && in.Profile != "http-server" {
		return nil, fmt.Errorf("unsupported dashboard profile")
	}
	if err := in.Scope.Validate(); err != nil {
		return nil, err
	}
	if !utf8.ValidString(in.Service) || len(in.Service) > 128 || strings.ContainsRune(in.Service, 0) {
		return nil, fmt.Errorf("invalid service")
	}
	if len(in.Fields) > 128 || len(in.Instruments) > 128 || len(in.Signals) > 3 || len(declarations) > 128 {
		return nil, fmt.Errorf("dashboard metadata budget exceeded")
	}
	signals := map[string]bool{}
	for _, s := range in.Signals {
		if s != "spans" && s != "logs" && s != "metrics" || signals[s] {
			return nil, fmt.Errorf("invalid or duplicate signal")
		}
		signals[s] = true
	}
	for _, f := range in.Fields {
		if _, err := f.Type("spans"); err != nil {
			return nil, err
		}
	}
	scope := in.Scope
	scope.Namespace = in.Scope.Namespace.Clone()
	p := &DashboardPlan{Version: DashboardVersion, Scope: scope, Service: in.Service, Widgets: []DashboardWidget{}, Declarations: []DashboardDeclaration{}, NativeOptions: []DashboardNativeOption{}, NativeLimit: DashboardNativeLimit, Warnings: []string{"Metadata suggests widgets, not service health or telemetry completeness. Each widget is a separate bounded source read; no transactional snapshot is claimed.", "Counts are captured records, not request throughput. Generic spans include retries and mixed span kinds; unset status is not verified success.", "Native evaluations preserve full OTel identity, units and temporality. At most two supported instruments are loaded; selection is local to this service visit, not saved dashboard configuration."}}
	httpCandidate := false
	httpAuto := false
	if signals["spans"] {
		for _, f := range in.Fields {
			if f == (Field{Kind: "attribute", Key: "http.request.method"}) {
				httpCandidate = true
				httpAuto = in.HTTPServerObserved
			}
		}
	}
	if signals["metrics"] {
		for _, instrument := range in.Instruments {
			if dashboardNativeHTTP(in.Scope, in.Service, instrument) {
				httpCandidate = true
				httpAuto = true
			}
		}
	}
	httpActive := false
	if in.Profile != "" {
		p.Profile = in.Profile
		if p.Profile == "auto" {
			p.Profile = "all-spans"
			if httpAuto {
				p.Profile = "http-server"
			}
		}
		if p.Profile == "http-server" && !httpCandidate {
			return nil, fmt.Errorf("HTTP server profile requires supplied current HTTP metadata or an eligible duration instrument")
		}
		httpActive = p.Profile == "http-server"
		genericState := "not_observed"
		if signals["spans"] {
			genericState = "supported"
		}
		httpState := "not_observed"
		httpReason := "No current HTTP method key or eligible HTTP duration instrument supplied; coverage unknown"
		if httpCandidate {
			httpState = "candidate"
			httpReason = "Metadata candidate only. Incoming span observations require exact Server kind and a nonempty current http.request.method in the actual read; conformance is unverified"
		}
		p.Families = []DashboardFamily{{Key: "http-server", Title: "Incoming HTTP", State: httpState, Grain: "server-span-record", Reason: httpReason, RuleSet: "goo/http-server-span-observations/v1", Selected: httpActive}, {Key: "all-spans", Title: "All spans", State: genericState, Grain: "span-record", Reason: "Mixed reported span kinds and retries; records are not request throughput", Selected: !httpActive}}
	}
	bucket := int64(1000)
	for _, b := range []int64{1000, 10000, 60000, 300000} {
		bucket = b
		if (in.Scope.ToMS-in.Scope.FromMS+b-1)/b <= 60 {
			break
		}
	}
	for _, signal := range []string{"spans", "logs"} {
		w := DashboardWidget{Key: signal, Title: map[string]string{"spans": "Span activity and duration", "logs": "Log activity and severity"}[signal], Grain: map[string]string{"spans": "span-record", "logs": "log-record"}[signal], State: "not_observed"}
		if signals[signal] {
			q := DashboardRecordQuestion(in.Scope, in.Service, signal)
			if httpActive && signal == "spans" {
				q = dashboardHTTPQuestion(q)
				w.Title = "HTTP server span observations"
				w.Grain = "server-span-record"
				w.Reason = "Exact Server kind and nonempty current HTTP method; observed span status is not a verified HTTP outcome policy"
			}
			q.BucketMS = bucket
			q.GroupBy = []Field{{Kind: "column", Key: map[string]string{"spans": "name", "logs": "severity"}[signal]}}
			if signal == "spans" {
				q.Summary = "span-duration"
				q.SpanStatus = true
			}
			w.State = "supported"
			w.Query = &q
		} else {
			w.Reason = "Not observed in supplied inventory; coverage unknown"
		}
		if in.Profile != "" {
			w.Family = map[string]string{"spans": p.Profile, "logs": "logs"}[signal]
		}
		p.Widgets = append(p.Widgets, w)
	}
	if signals["spans"] {
		keys := []string{"http.route", "rpc.method", "db.system.name", "db.system", "messaging.destination.name", "messaging.system"}
		if httpActive {
			keys = []string{"http.route"}
		}
		for _, key := range keys {
			found := httpActive
			for _, f := range in.Fields {
				if f.Kind == "attribute" && f.Key == key {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			q := DashboardRecordQuestion(in.Scope, in.Service, "spans")
			if httpActive {
				q = dashboardHTTPQuestion(q)
			}
			q.GroupBy = []Field{{Kind: "attribute", Key: key}}
			q.Summary = "span-duration"
			q.SpanStatus = true
			w := DashboardWidget{Key: "dimension", Title: "Observed " + key, State: "supported", Grain: "span-record", Query: &q}
			if in.Profile != "" {
				w.Family = p.Profile
			}
			if httpActive {
				w.Title = "HTTP server routes"
				w.Grain = "server-span-record"
				w.Reason = "Missing and empty routes remain separate; only matching server span observations are counted"
			}
			p.Widgets = append(p.Widgets, w)
			break
		}
	}
	// Detach and canonically sort descriptors. Never pick by discovery row order.
	raw, _ := json.Marshal(in.Instruments)
	var instruments []measurements.Instrument
	json.Unmarshal(raw, &instruments)
	sort.Slice(instruments, func(i, j int) bool {
		if in.Profile != "" {
			if httpActive {
				a, b := dashboardNativeHTTP(in.Scope, in.Service, instruments[i]), dashboardNativeHTTP(in.Scope, in.Service, instruments[j])
				if a != b {
					return a
				}
			}
			a, _ := json.Marshal(instruments[i].Definition)
			b, _ := json.Marshal(instruments[j].Definition)
			if string(a) != string(b) {
				return string(a) < string(b)
			}
		}
		a, _ := json.Marshal(instruments[i])
		b, _ := json.Marshal(instruments[j])
		return string(a) < string(b)
	})
	seen := map[string]bool{}
	selected := map[string]bool{}
	if in.NativeSelection != nil {
		if len(*in.NativeSelection) > DashboardNativeLimit {
			return nil, fmt.Errorf("select at most %d native instruments", DashboardNativeLimit)
		}
		for _, key := range *in.NativeSelection {
			if len(key) != 64 || selected[key] {
				return nil, fmt.Errorf("invalid or duplicate native selection")
			}
			selected[key] = true
		}
	}
	native := 0
	for _, instrument := range instruments {
		if !signals["metrics"] || !instrument.Supported || instrument.Definition == nil {
			continue
		}
		if instrument.Kind != "gauge" && instrument.Kind != "counter" && instrument.Kind != "histogram" {
			return nil, fmt.Errorf("unsupported native instrument kind")
		}
		if instrument.Kind == "gauge" && instrument.Temporality != 0 || instrument.Kind != "gauge" && instrument.Temporality != 1 && instrument.Temporality != 2 || instrument.Kind == "counter" && !instrument.Monotonic {
			return nil, fmt.Errorf("unsupported native temporality or monotonicity")
		}
		d := *instrument.Definition
		q := measurements.Question{Definition: d, FromMS: in.Scope.FromMS, ToMS: in.Scope.ToMS}
		if !d.Source.Namespace.Equal(in.Scope.Namespace) || d.Environment != in.Scope.Environment || len(d.Source.Services) != 1 || d.Source.Services[0] != in.Service || d.Kind != instrument.Kind || d.Source.Metric != instrument.Metric || d.Source.Unit != instrument.Unit || d.Source.Scope != instrument.Scope || d.Source.Version != instrument.Version || d.Source.SchemaURL != instrument.SchemaURL || q.Validate() != nil {
			return nil, fmt.Errorf("instrument definition does not match dashboard scope and identity")
		}
		encoded, _ := json.Marshal(d)
		key := string(encoded)
		if seen[key] {
			continue
		}
		seen[key] = true
		optionKey := measurements.Hash(d)
		isDefault := len(p.NativeOptions) < DashboardNativeLimit
		chosen := isDefault
		if in.NativeSelection != nil {
			chosen = selected[optionKey]
		}
		p.NativeOptions = append(p.NativeOptions, DashboardNativeOption{Key: optionKey, Instrument: instrument, Default: isDefault, Selected: chosen})
		delete(selected, optionKey)
		if !chosen {
			continue
		}
		native++
		w := DashboardWidget{Key: fmt.Sprintf("native-%d", native), Title: d.Title, State: "supported", Grain: "per-stream", Measurement: &q}
		if in.Profile != "" {
			w.Family = "native"
			if dashboardNativeHTTP(in.Scope, in.Service, instrument) {
				w.Family = "http-server"
				w.Reason = "Independent native HTTP duration binding; units, temporality and interval coverage require evaluation, not inferred span counts or conformance"
			}
		}
		p.Widgets = append(p.Widgets, w)
	}
	if len(selected) != 0 {
		return nil, fmt.Errorf("native selection is absent from supported instrument inventory")
	}
	for _, d := range declarations {
		if d.Environment == in.Scope.Environment && d.Service == in.Service {
			d.Namespace = d.Namespace.Clone()
			d.Operations = append([]DashboardOperation(nil), d.Operations...)
			d.Match = "exact"
			if d.Namespace == nil && in.Scope.Namespace != nil {
				d.Match = "namespace-unspecified"
			} else if !d.Namespace.Equal(in.Scope.Namespace) {
				d.Match = "namespace-mismatch"
			}
			p.Declarations = append(p.Declarations, d)
		}
	}
	sort.Slice(p.Declarations, func(i, j int) bool { return p.Declarations[i].Revision < p.Declarations[j].Revision })
	for i, d := range p.Declarations {
		if i >= 2 {
			break
		}
		w := DashboardWidget{Key: fmt.Sprintf("contract-%d", i+1), Title: d.Title + " · declared operation spans", Grain: "span-record", State: "unsupported", Reason: "Requires a complete bounded instrumentation binding and at most six declared operations"}
		if d.Match == "namespace-mismatch" {
			w.Reason = "Declared namespace differs; no contract question generated"
		} else if !signals["spans"] {
			w.State = "not_observed"
			w.Reason = "No spans observed in supplied inventory; declaration is not runtime evidence"
		} else if d.Scope != "" && d.Instance != "" && len(d.Operations) > 0 && len(d.Operations) <= 6 {
			q := DashboardRecordQuestion(in.Scope, in.Service, "spans")
			all := []Predicate{*q.Where}
			if q.Where.All != nil {
				all = q.Where.All
			}
			for _, f := range []struct{ kind, key, value string }{{"resource", "service.instance.id", d.Instance}, {"column", "scope", d.Scope}, {"column", "scope_version", d.Version}} {
				all = append(all, Predicate{Field: &Field{Kind: f.kind, Key: f.key}, Op: "eq", Value: f.value})
			}
			ops := []Predicate{}
			for _, op := range d.Operations {
				ops = append(ops, Predicate{All: []Predicate{{Field: &Field{Kind: "attribute", Key: "goo.operation.name"}, Op: "eq", Value: op.Name}, {Field: &Field{Kind: "attribute", Key: "goo.operation.role"}, Op: "eq", Value: op.Role}}})
			}
			all = append(all, Predicate{Any: ops})
			q.Where = &Predicate{All: all}
			q.GroupBy = []Field{{Kind: "attribute", Key: "goo.operation.name"}}
			q.Summary = "span-duration"
			q.SpanStatus = true
			w.State = "supported"
			w.Reason = "Captured spans matching declared operation names/roles and source binding; not a business-outcome or conformance evaluation. Schema URL is declared metadata; this exporter does not store it for verification"
			if d.Match == "namespace-unspecified" {
				w.Reason += ". Question explicitly narrows a namespace-unspecified declaration"
			}
			w.Query = &q
		}
		p.Widgets = append(p.Widgets, w)
	}
	for _, w := range p.Widgets {
		if w.Query != nil {
			if err := w.Query.Validate(); err != nil {
				return nil, err
			}
		}
	}
	return p, nil
}

// dashboardHTTPQuestion narrows the existing service/namespace selector. The
// current-key method policy is explicit: missing and reported empty are excluded;
// custom nonempty method strings remain reported values, not verified conformance.
func dashboardHTTPQuestion(q Query) Query {
	parts := []Predicate{}
	if q.Where != nil {
		if len(q.Where.All) > 0 {
			parts = append(parts, q.Where.All...)
		} else {
			parts = append(parts, *q.Where)
		}
	}
	parts = append(parts, Predicate{Field: &Field{Kind: "column", Key: "kind"}, Op: "eq", Value: "Server"}, Predicate{Field: &Field{Kind: "attribute", Key: "http.request.method"}, Op: "ne"})
	q.Where = &Predicate{All: parts}
	return q
}
func dashboardNativeHTTP(scope InventoryQuestion, service string, i measurements.Instrument) bool {
	if !i.Supported || i.Definition == nil || i.Kind != "histogram" || i.Metric != "http.server.request.duration" || (i.Unit != "s" && i.Unit != "ms") || (i.Temporality != 1 && i.Temporality != 2) {
		return false
	}
	d := i.Definition
	q := measurements.Question{Definition: *d, FromMS: scope.FromMS, ToMS: scope.ToMS}
	return d.Kind == i.Kind && d.Source.Metric == i.Metric && d.Source.Unit == i.Unit && d.Source.Scope == i.Scope && d.Source.Version == i.Version && d.Source.SchemaURL == i.SchemaURL && d.Source.Namespace.Equal(scope.Namespace) && d.Environment == scope.Environment && len(d.Source.Services) == 1 && d.Source.Services[0] == service && q.Validate() == nil
}

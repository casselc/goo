package exploration

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/measurements"
	"sort"
	"strings"
	"unicode/utf8"
)

const DashboardVersion = "goo/service-dashboard/v1"
const DashboardNativeLimit = 2

// Inputs are supplied observations, not authenticated source evidence or health.
type DashboardRequest struct {
	Scope       InventoryQuestion         `json:"scope"`
	Service     string                    `json:"service"`
	Signals     []string                  `json:"signals"`
	Fields      []Field                   `json:"fields"`
	Instruments []measurements.Instrument `json:"instruments"`
}
type DashboardDeclaration struct {
	Name             string `json:"name"`
	Title            string `json:"title"`
	Revision         string `json:"revision"`
	ManifestRevision string `json:"manifestRevision"`
	Service          string `json:"service"`
	Environment      string `json:"environment"`
}
type DashboardWidget struct {
	Key         string                 `json:"key"`
	Title       string                 `json:"title"`
	State       string                 `json:"state"`
	Reason      string                 `json:"reason,omitempty"`
	Grain       string                 `json:"grain"`
	Query       *Query                 `json:"query,omitempty"`
	Measurement *measurements.Question `json:"measurement,omitempty"`
}
type DashboardPlan struct {
	Version      string                 `json:"version"`
	Scope        InventoryQuestion      `json:"scope"`
	Service      string                 `json:"service"`
	Widgets      []DashboardWidget      `json:"widgets"`
	Declarations []DashboardDeclaration `json:"declarations"`
	Warnings     []string               `json:"warnings"`
}

func DashboardRecordQuestion(scope InventoryQuestion, service, signal string) Query {
	return Query{Signal: signal, Environment: scope.Environment, FromMS: scope.FromMS, ToMS: scope.ToMS, Limit: 100, Where: &Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: service}}
}

// PlanServiceDashboard performs no source reads. Every executable widget uses an
// existing typed evaluator; no SQL, health policy or fleet aggregation is inferred.
func PlanServiceDashboard(in DashboardRequest, declarations []DashboardDeclaration) (*DashboardPlan, error) {
	if err := in.Scope.Query().Validate(); err != nil {
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
	p := &DashboardPlan{Version: DashboardVersion, Scope: in.Scope, Service: in.Service, Widgets: []DashboardWidget{}, Declarations: []DashboardDeclaration{}, Warnings: []string{"Metadata suggests widgets, not service health or telemetry completeness. Each widget is a separate bounded source read; no transactional snapshot is claimed.", "Counts are captured records, not request throughput. Generic spans include retries and mixed span kinds; unset status is not verified success.", "Native evaluations preserve full OTel identity, units and temporality. At most two supported instruments are loaded automatically; other instruments remain discoverable."}}
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
		p.Widgets = append(p.Widgets, w)
	}
	if signals["spans"] {
		for _, key := range []string{"http.route", "rpc.method", "db.system.name", "db.system", "messaging.destination.name", "messaging.system"} {
			found := false
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
			q.GroupBy = []Field{{Kind: "attribute", Key: key}}
			q.Summary = "span-duration"
			q.SpanStatus = true
			p.Widgets = append(p.Widgets, DashboardWidget{Key: "dimension", Title: "Observed " + key, State: "supported", Grain: "span-record", Query: &q})
			break
		}
	}
	// Detach and canonically sort descriptors. Never pick by discovery row order.
	raw, _ := json.Marshal(in.Instruments)
	var instruments []measurements.Instrument
	json.Unmarshal(raw, &instruments)
	sort.Slice(instruments, func(i, j int) bool {
		a, _ := json.Marshal(instruments[i])
		b, _ := json.Marshal(instruments[j])
		return string(a) < string(b)
	})
	seen := map[string]bool{}
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
		if d.Environment != in.Scope.Environment || len(d.Source.Services) != 1 || d.Source.Services[0] != in.Service || d.Kind != instrument.Kind || d.Source.Metric != instrument.Metric || d.Source.Unit != instrument.Unit || d.Source.Scope != instrument.Scope || d.Source.Version != instrument.Version || d.Source.SchemaURL != instrument.SchemaURL || q.Validate() != nil {
			return nil, fmt.Errorf("instrument definition does not match dashboard scope and identity")
		}
		encoded, _ := json.Marshal(d)
		key := string(encoded)
		if seen[key] {
			continue
		}
		seen[key] = true
		if native >= DashboardNativeLimit {
			continue
		}
		native++
		p.Widgets = append(p.Widgets, DashboardWidget{Key: fmt.Sprintf("native-%d", native), Title: d.Title, State: "supported", Grain: "per-stream", Measurement: &q})
	}
	for _, d := range declarations {
		if d.Environment == in.Scope.Environment && d.Service == in.Service {
			p.Declarations = append(p.Declarations, d)
		}
	}
	sort.Slice(p.Declarations, func(i, j int) bool { return p.Declarations[i].Revision < p.Declarations[j].Revision })
	for _, w := range p.Widgets {
		if w.Query != nil {
			if err := w.Query.Validate(); err != nil {
				return nil, err
			}
		}
	}
	return p, nil
}

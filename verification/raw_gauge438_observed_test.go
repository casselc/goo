package verification

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strconv"
	"time"

	"regexp"
	"strings"
	"testing"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type gauge438Counters struct{ Source, Writes, SQL int }
type gauge438Step struct {
	Action             string
	Before, After      gauge438Counters
	Calls              []pair428Call
	SQL                []pair428SQL
	Focus, FocusBefore struct{ ID, Role, Text string }
}
type gauge438Report struct {
	Version, Status          string
	Cleanup, DatabaseCleanup bool
	Errors                   []string
	Fixture                  struct {
		Token, Database, Environment, Service, Namespace, Scope string
		Clock                                                   int64
	}
	Setup []struct {
		Kind   string
		Body   json.RawMessage
		Status int
	}
	Raw           []json.RawMessage
	Calls         []pair428Call
	SQL, ActorSQL []pair428SQL
	Steps         []gauge438Step
	Acquisitions  []struct {
		Name              string
		Request, Response json.RawMessage
		Status            int
	}
	Rejected []struct {
		Name, Path        string
		Request, Response json.RawMessage
		Status, SQL       int
	}
	RowBank []struct {
		Name   string
		ReadAt string
		Rows   []struct {
			Key    string
			Record json.RawMessage
		}
		SelectedKey string
		Detail      json.RawMessage
	}
	Bombadil struct {
		Snapshots int
		Completed json.RawMessage
	}
}

var gauge438Database = regexp.MustCompile(`^goo438_[a-f0-9]{12}$`)

func gauge438JSONEqual(a, b json.RawMessage) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}

// This gate separates source facts from the deliberately small acquisition model.
// The model never interprets these observations as evaluated gauge samples.
func TestQuintRawGauge438Observed(t *testing.T) {
	file := os.Getenv("GOO_RAW_GAUGE_438_REPORT")
	if file == "" {
		t.Skip("requires completed actual raw-gauge browser report")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report gauge438Report
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if reason := gauge438Admission(report); reason != "" {
		t.Fatal("actual raw-gauge admission:", reason)
	}
	checker, err := validate.NewChecker("model/population_drilldown.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	check := func(r gauge438Report) bool {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		reads := 0
		coherent := gauge438Admission(r) == ""
		for _, step := range r.Steps {
			if len(step.SQL) == 0 {
				rec.Record(t.Context(), "history", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			}
			for range step.SQL {
				reads++
				rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
			}
		}
		for range r.ActorSQL {
			reads++
			rec.Record(t.Context(), "read", nil, "obs.reads", reads, "obs.depth", 0, "obs.coherent", coherent)
		}
		verdict, e := checker.Check(t.Context(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return verdict.OK()
	}
	if !check(report) {
		t.Fatal("actual acquisition/local replay rejected")
	}
	for mutant := range 8 {
		encoded, _ := json.Marshal(report)
		var bad gauge438Report
		_ = json.Unmarshal(encoded, &bad)
		switch mutant {
		case 0:
			bad.Cleanup = false
		case 1:
			bad.Raw = bad.Raw[1:]
		case 2:
			bad.Acquisitions[0].Response = json.RawMessage(`{}`)
		case 3:
			bad.Rejected[0].SQL++
		case 4:
			bad.RowBank[0].Rows[0].Key = "invented-key"
		case 5:
			bad.Steps[0].After.Writes++
		case 6:
			bad.RowBank[len(bad.RowBank)-1].ReadAt = "2000-01-01T00:00:00Z"
		case 7:
			for i := range bad.Steps {
				if bad.Steps[i].Action == "late-response-ignored" {
					bad.Steps[i].Focus.Text = "stolen-focus"
				}
			}
		}
		if check(bad) {
			t.Fatalf("raw-gauge negative control %d accepted", mutant)
		}
	}
	t.Logf("admitted %d actual steps, %d source SQL, %d exporter rows and eight rejecting controls", len(report.Steps), len(report.SQL), len(report.Raw))
}

func gauge438Admission(r gauge438Report) string {
	if r.Version != "goo.rawGauge438/1" || r.Status != "passed" || !r.Cleanup || !r.DatabaseCleanup || len(r.Errors) != 0 {
		return "unfinished run"
	}
	if !gauge438Database.MatchString(r.Fixture.Database) || r.Fixture.Token == "" || r.Fixture.Service == "" || len(r.Setup) == 0 || len(r.Raw) == 0 {
		return "missing actual source bank"
	}
	if len(r.Acquisitions) == 0 || len(r.Rejected) == 0 || len(r.RowBank) == 0 || len(r.RowBank[0].Rows) == 0 || len(r.Steps) == 0 {
		return "missing acceptance banks"
	}
	for _, setup := range r.Setup {
		if setup.Kind != "metrics" || setup.Status < 200 || setup.Status >= 300 || !strings.Contains(string(setup.Body), "resourceMetrics") {
			return "OTLP setup not witnessed"
		}
	}
	for _, reject := range r.Rejected {
		var refused measurements.RawGaugeQuestion
		if json.Unmarshal(reject.Request, &refused) != nil || refused.Validate() == nil || reject.Path != "/api/metric-observations" {
			return "unsupported request is not a typed refusal"
		}
		switch reject.Name {
		case "unsupported-kind":
			if refused.Instrument.Kind != "counter" {
				return "non-gauge refusal missing"
			}
			refused.Instrument.Kind = "gauge"
		case "unsupported-all":
			if refused.Question.Environment != "" || refused.Question.EnvironmentSelector == nil || refused.Question.EnvironmentSelector.Scope.Mode != "all" {
				return "All refusal has mixed or absent scope"
			}
			refused.Question.EnvironmentSelector = nil
			refused.Question.Environment = r.Fixture.Environment
		default:
			return "undeclared unsupported control"
		}
		if refused.Validate() != nil {
			return "refusal control is otherwise malformed"
		}
		if reject.Status < 400 || reject.SQL < 0 || (!strings.Contains(reject.Name, "source") && !strings.Contains(reject.Name, "budget") && reject.SQL != 0) {
			return "rejected request performed unauthorized source work"
		}
	}
	for _, step := range r.Steps {
		if step.Before.Writes != 0 || step.After.Writes != 0 {
			return "raw observation gained store authority"
		}
		if step.After.Source < step.Before.Source || step.After.SQL-step.Before.SQL != len(step.SQL) {
			return "source counter discontinuity"
		}
	}
	return gauge438DataAdmission(r)
}

func gauge438DataAdmission(r gauge438Report) string {

	type exporter struct {
		ServiceName, MetricName, MetricUnit, ScopeName, ScopeVersion, ScopeSchemaUrl, ResourceSchemaUrl, TimeUnixNS, StartTimeUnixNS, Value string
		ResourceAttributes, Attributes                                                                                                      map[string]string
		Flags                                                                                                                               uint32
	}
	if reason := gauge438OTLPAdmission(r); reason != "" {
		return reason
	}
	raw := []exporter{}
	for _, encoded := range r.Raw {
		var row exporter
		if json.Unmarshal(encoded, &row) != nil || row.ServiceName != r.Fixture.Service || row.ResourceAttributes["fixture.bank"] != r.Fixture.Token {
			return "exporter bank mismatch"
		}
		raw = append(raw, row)
	}
	admitted := map[string]measurements.RawGaugeObservation{}
	readsByName := map[string]string{}
	for _, acq := range r.Acquisitions {
		var q measurements.RawGaugeQuestion
		var result measurements.RawGaugeResult
		if acq.Status != 200 {
			if (acq.Name != "source-refusal" && acq.Name != "source-budget-refusal" && acq.Name != "actual-source-budget") || acq.Status != 502 {
				return "unexpected failed acquisition"
			}
			var refusal struct {
				Refusal struct{ Kind string }
				Rows    json.RawMessage
			}
			if json.Unmarshal(acq.Response, &refusal) != nil || len(refusal.Rows) != 0 || refusal.Refusal.Kind == "" {
				return "failed acquisition published rows or omitted refusal"
			}
			continue
		}
		if json.Unmarshal(acq.Request, &q) != nil || q.Validate() != nil || strict438(acq.Response, &result) != nil {
			return "invalid acquisition wire"
		}
		if result.Version != "goo/raw-gauge-observations/v1" || result.Limit != 10000 || !reflect.DeepEqual(result.Question, q.Question) || result.Instrument != q.Instrument {
			return "raw profile echo changed"
		}
		witnessed := false
		for _, call := range r.Calls {
			if call.Method == "POST" && call.Path == "/api/metric-observations" && call.Status == acq.Status && gauge438JSONEqual(call.Body, acq.Request) && gauge438JSONEqual(call.Response, acq.Response) {
				witnessed = true
				break
			}
		}
		if !witnessed {
			return "acquisition absent from complete HTTP ledger"
		}
		readAt, e := time.Parse(time.RFC3339Nano, result.ReadAt)
		if e != nil || readAt.UnixMilli() <= q.Question.ToMS {
			return "readAt confused with observation time"
		}
		readsByName[acq.Name] = result.ReadAt
		selection, _ := q.Question.EnvironmentSelection()
		expected := []exporter{}
		for _, row := range raw {
			ns, e := strconv.ParseInt(row.TimeUnixNS, 10, 64)
			if e == nil && ns >= q.Question.FromMS*1000000 && ns < q.Question.ToMS*1000000 && row.ServiceName == q.Question.Service && selection.Matches(row.ResourceAttributes) && q.Question.Namespace.Matches(row.ResourceAttributes) && row.ScopeName == q.Instrument.Scope && row.ScopeVersion == q.Instrument.Version && row.ScopeSchemaUrl == q.Instrument.SchemaURL && row.MetricName == q.Instrument.Metric && row.MetricUnit == q.Instrument.Unit {
				expected = append(expected, row)
			}
		}
		if len(expected) != len(result.Rows) {
			return "raw occurrence multiplicity changed"
		}
		keys := map[string]bool{}
		for _, row := range result.Rows {
			if row.OccurrenceKey == "" || keys[row.OccurrenceKey] {
				return "duplicate local occurrence key"
			}
			keys[row.OccurrenceKey] = true
			match := slices.IndexFunc(expected, func(source exporter) bool {
				return source.TimeUnixNS == row.EndNS && source.StartTimeUnixNS == value438(row.StartNS) && source.ScopeName == row.Stream.Scope && source.ScopeVersion == row.Stream.Version && source.ScopeSchemaUrl == row.Stream.SchemaURL && source.MetricName == row.Stream.Name && source.MetricUnit == row.Stream.Unit && source.ResourceSchemaUrl == row.Stream.ResourceSchemaURL && reflect.DeepEqual(source.ResourceAttributes, row.Stream.Resource) && reflect.DeepEqual(source.Attributes, row.Stream.Attributes) && row.Flags != nil && source.Flags == *row.Flags && gauge438JSONEqual(row.RawValue, mustJSON438(source.Value))
			})
			if match < 0 {
				return "typed observation lost actual exporter facts"
			}
			source := expected[match]
			expected = append(expected[:match], expected[match+1:]...)
			var projected struct {
				Resource, Attributes       map[string]string
				Scope, Version, Name, Unit string
				SchemaURL                  string `json:"schema_url"`
				ResourceSchemaURL          string `json:"resource_schema_url"`
				StartNS                    string `json:"start_ns"`
				EndNS                      string `json:"end_ns"`
				Value                      string
				Flags                      uint32
			}
			if json.Unmarshal(row.RawRow, &projected) != nil || !reflect.DeepEqual(projected.Resource, source.ResourceAttributes) || !reflect.DeepEqual(projected.Attributes, source.Attributes) || projected.Value != source.Value || projected.EndNS != source.TimeUnixNS || projected.StartNS != source.StartTimeUnixNS || projected.Flags != source.Flags || projected.Name != source.MetricName || projected.Unit != source.MetricUnit || projected.Scope != source.ScopeName || projected.Version != source.ScopeVersion || projected.SchemaURL != source.ScopeSchemaUrl || projected.ResourceSchemaURL != source.ResourceSchemaUrl {
				return "rawRow differs from independent exporter"
			}
			value, e := strconv.ParseFloat(source.Value, 64)
			wantState := "finite"
			if source.Flags&1 != 0 {
				wantState = "absent"
			} else if source.Value == "nan" || source.Value == "NaN" || strings.Contains(strings.ToLower(source.Value), "inf") {
				wantState = "nonfinite"
			} else if e != nil {
				wantState = "invalid"
			}
			if row.ValueState != wantState || wantState == "finite" && (row.Value == nil || *row.Value != value) || wantState != "finite" && row.Value != nil {
				return "reported scalar state changed"
			}
			encoded, _ := json.Marshal(row)
			admitted[string(encoded)] = row
		}
	}
	owners := map[string]string{"all-held": "main", "find-standard": "main", "other-unit": "other-unit", "nonfinite-otlp": "nonfinite-otlp", "revisit": "revisit", "after-late-response": "revisit"}
	for _, bank := range r.RowBank {
		if owners[bank.Name] == "" || bank.ReadAt == "" || bank.ReadAt != readsByName[owners[bank.Name]] {
			return "held source-read receipt changed"
		}
		for _, held := range bank.Rows {
			var row measurements.RawGaugeObservation
			if json.Unmarshal(held.Record, &row) != nil || row.OccurrenceKey != held.Key {
				return "held occurrence key changed"
			}
			encoded, _ := json.Marshal(row)
			if _, ok := admitted[string(encoded)]; !ok {
				return "held row absent from actual acquired bank"
			}
		}
	}
	local := map[string]bool{"metadata-open": true, "find-held": true, "inspect-observation": true, "close-details": true, "retired-control": true, "late-response-ignored": true}
	seen := map[string]bool{}
	indices := map[int]bool{}
	for _, step := range r.Steps {
		seen[step.Action] = true
		if step.Action == "late-response-ignored" && step.Focus != step.FocusBefore {
			return "late response stole focus"
		}
		if local[step.Action] && (step.Before != step.After || len(step.SQL) != 0) {
			return "local action gained source authority"
		}
		for _, query := range step.SQL {
			if indices[query.Index] {
				return "SQL assigned to multiple steps"
			}
			indices[query.Index] = true
		}
	}
	for _, query := range r.ActorSQL {
		if indices[query.Index] {
			return "actor duplicated main SQL"
		}
		indices[query.Index] = true
	}
	for _, query := range r.SQL {
		if !indices[query.Index] || query.Parameters["database"] != r.Fixture.Database || query.Parameters["readonly"] != "1" {
			return "configured-source SQL ownership changed"
		}
	}
	for _, action := range []string{"metadata-open", "read-observations", "find-held", "inspect-observation", "read-other-unit", "pending-departure", "retired-control", "explicit-revisit-read", "late-response-ignored"} {
		if !seen[action] {
			return "missing required journey action: " + action
		}
	}
	var actor struct {
		Version                                               string
		Phase, Source, Forbidden, Rows, Requests, BeforeLocal int
		Overflow, Detail                                      bool
	}
	if r.Bombadil.Snapshots <= 0 || json.Unmarshal(r.Bombadil.Completed, &actor) != nil || actor.Version != "goo.rawGauge438Actor/1" || actor.Phase != 6 || actor.Source != 1 || actor.Forbidden != 0 || actor.Overflow || !actor.Detail || actor.Rows != 2 || actor.Requests != actor.BeforeLocal {
		return "actor did not complete bounded source/local campaign"
	}
	return ""
}

func value438(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
func mustJSON438(value any) json.RawMessage { raw, _ := json.Marshal(value); return raw }

// Independently flatten the actual OTLP request rather than deriving the source
// bank from product results. Only the fixture's scalar gauge profile is admitted.
func gauge438OTLPAdmission(r gauge438Report) string {
	type attribute struct {
		Key   string
		Value struct{ StringValue string }
	}
	type point struct {
		TimeUnixNano, StartTimeUnixNano string
		AsDouble                        json.RawMessage
		Attributes                      []attribute
		Flags                           uint32
	}
	type metric struct {
		Name, Unit string
		Gauge      *struct{ DataPoints []point }
	}
	type scope struct {
		Scope     struct{ Name, Version string }
		SchemaURL string `json:"schemaUrl"`
		Metrics   []metric
	}
	type resource struct {
		Resource     struct{ Attributes []attribute }
		SchemaURL    string `json:"schemaUrl"`
		ScopeMetrics []scope
	}
	attrMap := func(attrs []attribute) map[string]string {
		out := map[string]string{}
		for _, a := range attrs {
			out[a.Key] = a.Value.StringValue
		}
		return out
	}
	expected := []json.RawMessage{}
	for _, setup := range r.Setup {
		var body struct{ ResourceMetrics []resource }
		if json.Unmarshal(setup.Body, &body) != nil {
			return "malformed actual OTLP setup"
		}
		for _, resource := range body.ResourceMetrics {
			for _, scope := range resource.ScopeMetrics {
				for _, metric := range scope.Metrics {
					if metric.Gauge == nil {
						continue
					}
					for _, point := range metric.Gauge.DataPoints {
						value := string(point.AsDouble)
						if strings.HasPrefix(value, "\"") {
							if json.Unmarshal(point.AsDouble, &value) != nil {
								return "bad OTLP scalar"
							}
						}
						start := point.StartTimeUnixNano
						if start == "" {
							start = "0"
						}
						attrs := attrMap(resource.Resource.Attributes)
						row := map[string]any{"ServiceName": attrs["service.name"], "ResourceAttributes": attrs, "ResourceSchemaUrl": resource.SchemaURL, "Attributes": attrMap(point.Attributes), "MetricName": metric.Name, "MetricUnit": metric.Unit, "ScopeName": scope.Scope.Name, "ScopeVersion": scope.Scope.Version, "ScopeSchemaUrl": scope.SchemaURL, "TimeUnixNS": point.TimeUnixNano, "StartTimeUnixNS": start, "Value": strings.ToLower(value), "Flags": point.Flags}
						expected = append(expected, mustJSON438(row))
					}
				}
			}
		}
	}
	if len(expected) != len(r.Raw) {
		return "OTLP/exporter observation count differs"
	}
	for _, raw := range r.Raw {
		var row map[string]any
		if json.Unmarshal(raw, &row) != nil {
			return "malformed exporter oracle"
		}
		if value, ok := row["Value"].(string); ok {
			row["Value"] = strings.ToLower(value)
		}
		canonical := mustJSON438(row)
		i := slices.IndexFunc(expected, func(candidate json.RawMessage) bool { return gauge438JSONEqual(candidate, canonical) })
		if i < 0 {
			return "exporter row absent from actual OTLP request"
		}
		expected = append(expected[:i], expected[i+1:]...)
	}
	return ""
}

func strict438(raw json.RawMessage, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

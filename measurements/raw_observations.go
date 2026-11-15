package measurements

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/casselc/goo/internal/sourcefailure"
)

type rawObservationInstrument struct{ Scope, Version, SchemaURL, Metric, Unit string }

// RawObservationStream is reported metadata, never an inferred unique producer.
type RawObservationStream struct {
	ResourceSchemaURL string            `json:"resourceSchemaURL"`
	Resource          map[string]string `json:"resource"`
	Scope             string            `json:"scope"`
	Version           string            `json:"version"`
	SchemaURL         string            `json:"schemaURL"`
	Name              string            `json:"name"`
	Unit              string            `json:"unit"`
	Attributes        map[string]string `json:"attributes"`
}
type RawObservationDiagnostic struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}
type RawObservation struct {
	OccurrenceKey string                     `json:"occurrenceKey"`
	Stream        RawObservationStream       `json:"stream"`
	StartNS       *string                    `json:"startNS"`
	EndNS         string                     `json:"endNS"`
	StartState    string                     `json:"startState"`
	RawValue      json.RawMessage            `json:"rawValue"`
	Value         *float64                   `json:"value"`
	ValueState    string                     `json:"valueState"`
	Flags         *uint32                    `json:"flags"`
	FlagsState    string                     `json:"flagsState"`
	Diagnostics   []RawObservationDiagnostic `json:"diagnostics"`
	RawRow        json.RawMessage            `json:"rawRow"`
}

func rawObservationRows(q DiscoveryQuestion, i rawObservationInstrument, rows []json.RawMessage) ([]RawObservation, error) {
	return rawObservationRowsForFamily(q, i, rows, true)
}

// The histogram adapter reuses identity, clock and flag admission without
// manufacturing a scalar value. Existing scalar profiles retain their own facts.
func rawObservationRowsForFamily(q DiscoveryQuestion, i rawObservationInstrument, rows []json.RawMessage, scalar bool) ([]RawObservation, error) {
	if len(rows) > RowLimit {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("raw observation rows exceed limit"))
	}
	out := []RawObservation{}
	selection, _ := q.EnvironmentSelection()
	for index, raw := range rows {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || fields == nil || !utf8.Valid(raw) {
			return nil, fmt.Errorf("invalid raw observation object")
		}
		var wire struct {
			Resource          map[string]string `json:"resource"`
			Attributes        map[string]string `json:"attributes"`
			ResourceSchemaURL string            `json:"resource_schema_url"`
			Scope             string            `json:"scope"`
			Version           string            `json:"version"`
			SchemaURL         string            `json:"schema_url"`
			Name              string            `json:"name"`
			Unit              string            `json:"unit"`
		}
		if err := json.Unmarshal(raw, &wire); err != nil {
			return nil, fmt.Errorf("invalid raw observation identity")
		}
		for _, key := range []string{"resource", "attributes", "resource_schema_url", "scope", "version", "schema_url", "name", "unit"} {
			field, ok := fields[key]
			if !ok || bytes.Equal(bytes.TrimSpace(field), []byte("null")) {
				return nil, fmt.Errorf("missing raw observation identity field")
			}
		}
		for _, key := range []string{"resource", "attributes"} {
			var values map[string]json.RawMessage
			if json.Unmarshal(fields[key], &values) != nil || values == nil {
				return nil, fmt.Errorf("invalid raw observation attributes")
			}
			for _, value := range values {
				var text string
				if len(value) == 0 || value[0] != '"' || json.Unmarshal(value, &text) != nil {
					return nil, fmt.Errorf("invalid raw observation attribute value")
				}
			}
		}
		if wire.Resource == nil || wire.Attributes == nil || wire.Resource["service.name"] != q.Service || !selection.Matches(wire.Resource) || !q.Namespace.Matches(wire.Resource) || wire.Scope != i.Scope || wire.Version != i.Version || wire.SchemaURL != i.SchemaURL || wire.Name != i.Metric || wire.Unit != i.Unit {
			return nil, fmt.Errorf("raw observation identity differs from selected scope")
		}
		end, state := rawObservationTimestamp(fields["end_ns"])
		if state != "valid" {
			return nil, fmt.Errorf("raw observation requires observation time")
		}
		endNS, _ := strconv.ParseInt(*end, 10, 64)
		if endNS < q.FromMS*1000000 || endNS >= q.ToMS*1000000 {
			return nil, fmt.Errorf("raw observation observation outside selected bounds")
		}
		row := RawObservation{OccurrenceKey: strconv.Itoa(index + 1), Stream: RawObservationStream{ResourceSchemaURL: wire.ResourceSchemaURL, Resource: wire.Resource, Attributes: wire.Attributes, Scope: wire.Scope, Version: wire.Version, SchemaURL: wire.SchemaURL, Name: wire.Name, Unit: wire.Unit}, EndNS: *end, RawRow: bytes.Clone(raw), Diagnostics: []RawObservationDiagnostic{}}
		row.StartNS, row.StartState = rawObservationTimestamp(fields["start_ns"])
		diagnostic := func(field, code string) {
			row.Diagnostics = append(row.Diagnostics, RawObservationDiagnostic{Field: field, Code: code})
		}
		if row.StartState != "valid" {
			diagnostic("startNS", row.StartState)
		} else {
			start, _ := strconv.ParseInt(*row.StartNS, 10, 64)
			if start > endNS {
				row.StartState = "invalid"
				diagnostic("startNS", "after-observation")
			}
		}
		if scalar {
			row.RawValue = bytes.Clone(fields["value"])
			row.Value, row.ValueState = rawObservationValue(row.RawValue)
			if row.ValueState != "finite" {
				diagnostic("value", row.ValueState)
			}
		}
		row.FlagsState = "absent"
		if flag, exists := fields["flags"]; exists && !bytes.Equal(bytes.TrimSpace(flag), []byte("null")) {
			flags, valid := rawObservationFlags(flag)
			if !valid {
				row.FlagsState = "invalid"
			} else {
				row.Flags = &flags
				row.FlagsState = "valid"
				if flags & ^uint32(1) != 0 {
					diagnostic("flags", "unknown-bits")
				}
				if scalar && flags&1 != 0 {
					row.Value = nil
					row.ValueState = "absent"
					diagnostic("value", "no-recorded-value")
				}
			}
		}
		if row.FlagsState != "valid" {
			diagnostic("flags", row.FlagsState)
		}
		out = append(out, row)
	}
	return out, nil
}
func rawObservationTimestamp(raw json.RawMessage) (*string, string) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, "absent"
	}
	var value string
	if json.Unmarshal(raw, &value) != nil {
		return nil, "invalid"
	}
	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil || n < 0 || strconv.FormatInt(n, 10) != value {
		return &value, "invalid"
	}
	return &value, "valid"
}
func rawObservationValue(raw json.RawMessage) (*float64, string) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, "absent"
	}
	text := string(raw)
	if len(raw) > 0 && raw[0] == '"' {
		if json.Unmarshal(raw, &text) != nil {
			return nil, "invalid"
		}
	} else if len(raw) == 0 || !strings.ContainsAny(string(raw[:1]), "-0123456789") {
		return nil, "invalid"
	}
	n, err := strconv.ParseFloat(text, 64)
	if math.IsNaN(n) || math.IsInf(n, 0) {
		return nil, "nonfinite"
	}
	if err != nil {
		return nil, "invalid"
	}
	return &n, "finite"
}

// Flags are semantic unsigned integers, including equivalent JSON decimals or
// exponents. Exact rational parsing avoids rounding a fractional/overflow value
// into an admitted bitmask at the uint32 boundary.
func rawObservationFlags(raw json.RawMessage) (uint32, bool) {
	if len(raw) == 0 || len(raw) > 256 || !strings.ContainsAny(string(raw[:1]), "0123456789-") {
		return 0, false
	}
	var number json.Number
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&number) != nil {
		return 0, false
	}
	text := string(number)
	if e := strings.IndexAny(text, "eE"); e >= 0 {
		exponent, err := strconv.ParseInt(text[e+1:], 10, 32)
		if err != nil || exponent < -4096 || exponent > 4096 {
			return 0, false
		}
	}
	n, ok := new(big.Rat).SetString(text)
	if !ok || !n.IsInt() || n.Sign() < 0 || n.Num().BitLen() > 32 {
		return 0, false
	}
	return uint32(n.Num().Uint64()), true
}

// validateRawObservation selects source scope directly; it does not construct an evaluator definition.
func validateRawObservation(q DiscoveryQuestion, i rawObservationInstrument) error {
	if _, err := q.EnvironmentSelection(); err != nil {
		return err
	}
	if err := q.Namespace.Validate(); err != nil {
		return err
	}
	d := q
	if d.Service == "" || !text(d.Service, 128) {
		return fmt.Errorf("require one bounded service")
	}
	if d.FromMS < 0 || d.ToMS <= d.FromMS || d.ToMS-d.FromMS > 7200000 || d.ToMS >= 4102444800000 {
		return fmt.Errorf("choose a half-open window of at most 120 minutes before 2100")
	}
	if i.Metric == "" || !text(i.Metric, 256) || !text(i.Scope, 256) || !text(i.Version, 128) || !text(i.SchemaURL, 1024) || !text(i.Unit, 64) {
		return fmt.Errorf("require a bounded exact instrument")
	}
	return nil
}

// The optional sum tuple selects only the fixed sum source and its exact reported metadata.
func rawObservationPlan(q DiscoveryQuestion, i rawObservationInstrument) (Plan, string, error) {
	p := Plan{Parameters: map[string]string{}}
	param := func(value, typ string) string {
		key := fmt.Sprintf("p%d", len(p.Parameters))
		p.Parameters[key] = value
		return "{" + key + ":" + typ + "}"
	}
	d := q
	where := "TimeUnix>=fromUnixTimestamp64Milli(" + param(strconv.FormatInt(d.FromMS, 10), "Int64") + ") AND TimeUnix<fromUnixTimestamp64Milli(" + param(strconv.FormatInt(d.ToMS, 10), "Int64") + ")"
	selection, err := d.EnvironmentSelection()
	if err != nil {
		return Plan{}, "", err
	}
	environment, err := metricEnvironmentPredicate(selection, param)
	if err != nil {
		return Plan{}, "", err
	}
	where += " AND " + environment + " AND ServiceName=" + param(d.Service, "String")
	if n := d.Namespace; n != nil {
		if n.Mode == "missing" {
			where += " AND NOT mapContains(ResourceAttributes,'service.namespace')"
		} else {
			where += " AND mapContains(ResourceAttributes,'service.namespace') AND ResourceAttributes['service.namespace']=" + param(n.Value, "String")
		}
	}
	for _, pair := range [][2]string{{"ScopeName", i.Scope}, {"ScopeVersion", i.Version}, {"ScopeSchemaUrl", i.SchemaURL}, {"MetricName", i.Metric}, {"MetricUnit", i.Unit}} {
		where += " AND " + pair[0] + "=" + param(pair[1], "String")
	}
	return p, where, nil
}

func compileRawObservation(q DiscoveryQuestion, i rawObservationInstrument, sum *RawSumInstrument) (Plan, error) {
	p, where, err := rawObservationPlan(q, i)
	if err != nil {
		return Plan{}, err
	}
	param := func(value, typ string) string {
		key := fmt.Sprintf("p%d", len(p.Parameters))
		p.Parameters[key] = value
		return "{" + key + ":" + typ + "}"
	}

	fields := "SELECT mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, ResourceSchemaUrl AS resource_schema_url, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schema_url, MetricName AS name, MetricUnit AS unit, Attributes AS attributes, toString(toUnixTimestamp64Nano(StartTimeUnix)) AS start_ns, toString(toUnixTimestamp64Nano(TimeUnix)) AS end_ns, toString(Value) AS value, Flags AS flags"
	table := "otel_metrics_gauge"
	if sum != nil {
		table = "otel_metrics_sum"
		fields += ", AggregationTemporality AS temporality, IsMonotonic AS monotonic"
		where += " AND AggregationTemporality=" + param(strconv.FormatInt(int64(sum.Temporality), 10), "Int32") + " AND IsMonotonic=" + param(strconv.FormatBool(sum.Monotonic), "Bool")
	}
	p.SQL = fields + " FROM " + table + " WHERE " + where + " ORDER BY TimeUnix LIMIT 10001 FORMAT JSONEachRow"
	return p, nil
}

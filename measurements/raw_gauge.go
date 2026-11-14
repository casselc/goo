package measurements

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcefailure"
)

// RawGaugeInstrument selects reported metadata exactly, including empty labels.
// It makes no claim about evaluator readiness or producer identity.
type RawGaugeInstrument struct {
	Kind      string `json:"kind"`
	Scope     string `json:"scope"`
	Version   string `json:"version"`
	SchemaURL string `json:"schemaURL"`
	Metric    string `json:"metric"`
	Unit      string `json:"unit"`
}
type RawGaugeQuestion struct {
	Question   DiscoveryQuestion  `json:"question"`
	Instrument RawGaugeInstrument `json:"instrument"`
}

func (i *RawGaugeInstrument) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return fmt.Errorf("instrument requires an object")
	}
	for _, key := range []string{"kind", "scope", "version", "schemaURL", "metric", "unit"} {
		raw, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("instrument requires explicit non-null %s", key)
		}
	}
	type wire RawGaugeInstrument
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*i = RawGaugeInstrument(value)
	return nil
}
func (q RawGaugeQuestion) Validate() error {
	// Raw inspection validates source selection directly; evaluator binding and
	// producer readiness must never become prerequisites for reading observations.
	if _, err := q.Question.EnvironmentSelection(); err != nil {
		return err
	}
	if err := q.Question.Namespace.Validate(); err != nil {
		return err
	}
	d := q.Question
	if d.Service == "" || !text(d.Service, 128) {
		return fmt.Errorf("require one bounded service")
	}
	if d.FromMS < 0 || d.ToMS <= d.FromMS || d.ToMS-d.FromMS > 7200000 || d.ToMS >= 4102444800000 {
		return fmt.Errorf("choose a half-open window of at most 120 minutes before 2100")
	}
	i := q.Instrument
	if i.Kind != "gauge" || i.Metric == "" || !text(i.Metric, 256) || !text(i.Scope, 256) || !text(i.Version, 128) || !text(i.SchemaURL, 1024) || !text(i.Unit, 64) {
		return fmt.Errorf("require a bounded exact gauge instrument")
	}
	return nil
}

// RawGaugeStream is reported metadata, never an inferred unique producer.
type RawGaugeStream struct {
	ResourceSchemaURL string            `json:"resourceSchemaURL"`
	Resource          map[string]string `json:"resource"`
	Scope             string            `json:"scope"`
	Version           string            `json:"version"`
	SchemaURL         string            `json:"schemaURL"`
	Name              string            `json:"name"`
	Unit              string            `json:"unit"`
	Attributes        map[string]string `json:"attributes"`
}
type RawGaugeDiagnostic struct {
	Field string `json:"field"`
	Code  string `json:"code"`
}
type RawGaugeObservation struct {
	OccurrenceKey string               `json:"occurrenceKey"`
	Stream        RawGaugeStream       `json:"stream"`
	StartNS       *string              `json:"startNS"`
	EndNS         string               `json:"endNS"`
	StartState    string               `json:"startState"`
	RawValue      json.RawMessage      `json:"rawValue"`
	Value         *float64             `json:"value"`
	ValueState    string               `json:"valueState"`
	Flags         *uint32              `json:"flags"`
	FlagsState    string               `json:"flagsState"`
	Diagnostics   []RawGaugeDiagnostic `json:"diagnostics"`
	RawRow        json.RawMessage      `json:"rawRow"`
}
type RawGaugeResult struct {
	Version    string                `json:"version"`
	Question   DiscoveryQuestion     `json:"question"`
	Instrument RawGaugeInstrument    `json:"instrument"`
	ReadAt     string                `json:"readAt"`
	Rows       []RawGaugeObservation `json:"rows"`
	Limit      int                   `json:"limit"`
}

// CompileRawGaugeObservations binds a half-open source read, not a measurement.
func CompileRawGaugeObservations(q RawGaugeQuestion) (Plan, error) {
	if err := q.Validate(); err != nil {
		return Plan{}, err
	}
	p := Plan{Parameters: map[string]string{}}
	param := func(value, typ string) string {
		key := fmt.Sprintf("p%d", len(p.Parameters))
		p.Parameters[key] = value
		return "{" + key + ":" + typ + "}"
	}
	d := q.Question
	i := q.Instrument
	where := "TimeUnix>=fromUnixTimestamp64Milli(" + param(strconv.FormatInt(d.FromMS, 10), "Int64") + ") AND TimeUnix<fromUnixTimestamp64Milli(" + param(strconv.FormatInt(d.ToMS, 10), "Int64") + ")"
	selection, err := d.EnvironmentSelection()
	if err != nil {
		return Plan{}, err
	}
	environment, err := metricEnvironmentPredicate(selection, param)
	if err != nil {
		return Plan{}, err
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
	p.SQL = "SELECT mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, ResourceSchemaUrl AS resource_schema_url, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schema_url, MetricName AS name, MetricUnit AS unit, Attributes AS attributes, toString(toUnixTimestamp64Nano(StartTimeUnix)) AS start_ns, toString(toUnixTimestamp64Nano(TimeUnix)) AS end_ns, toString(Value) AS value, Flags AS flags FROM otel_metrics_gauge WHERE " + where + " ORDER BY TimeUnix LIMIT 10001 FORMAT JSONEachRow"
	return p, nil
}

// ReadRawGaugeObservations checks metadata availability then explicitly reads
// bounded raw points. It creates no custody, capture, evaluator or stream IDs.
func (s Source) ReadRawGaugeObservations(ctx context.Context, q RawGaugeQuestion) (out *RawGaugeResult, resultErr error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	q.Question = q.Question.clone()
	defer func() { resultErr = sourcefailure.Protocol(resultErr) }()
	families, err := sourcecap.Inspect(ctx, func(ctx context.Context, sql string, params map[string]string) ([]json.RawMessage, error) {
		return s.read(ctx, Plan{SQL: sql, Parameters: params})
	}, sourcecap.Gauge)
	if err != nil {
		return nil, err
	}
	if len(families) != 1 || !families[0].Compatible() {
		return nil, sourcefailure.Wrap(sourcefailure.Unavailable, fmt.Errorf("gauge metadata source unavailable"))
	}
	plan, err := CompileRawGaugeObservations(q)
	if err != nil {
		return nil, err
	}
	rows, err := s.read(ctx, plan)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out, err = rawGaugeRows(q, rows, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	// Raw rows are retained alongside typed displays; enforce the response budget
	// as well as the bounded acquisition before publishing anything.
	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	if len(encoded) > 8<<20 {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("raw gauge result exceeds 8 MiB"))
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func rawGaugeRows(q RawGaugeQuestion, rows []json.RawMessage, at time.Time) (*RawGaugeResult, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	if len(rows) > RowLimit {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("raw gauge rows exceed limit"))
	}
	out := &RawGaugeResult{Version: "goo/raw-gauge-observations/v1", Question: q.Question.clone(), Instrument: q.Instrument, ReadAt: at.UTC().Format(time.RFC3339Nano), Rows: []RawGaugeObservation{}, Limit: RowLimit}
	selection, _ := q.Question.EnvironmentSelection()
	for index, raw := range rows {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil || fields == nil || !utf8.Valid(raw) {
			return nil, fmt.Errorf("invalid raw gauge object")
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
			return nil, fmt.Errorf("invalid raw gauge identity")
		}
		for _, key := range []string{"resource", "attributes", "resource_schema_url", "scope", "version", "schema_url", "name", "unit"} {
			field, ok := fields[key]
			if !ok || bytes.Equal(bytes.TrimSpace(field), []byte("null")) {
				return nil, fmt.Errorf("missing raw gauge identity field")
			}
		}
		for _, key := range []string{"resource", "attributes"} {
			var values map[string]json.RawMessage
			if json.Unmarshal(fields[key], &values) != nil || values == nil {
				return nil, fmt.Errorf("invalid raw gauge attributes")
			}
			for _, value := range values {
				var text string
				if len(value) == 0 || value[0] != '"' || json.Unmarshal(value, &text) != nil {
					return nil, fmt.Errorf("invalid raw gauge attribute value")
				}
			}
		}
		i := q.Instrument
		if wire.Resource == nil || wire.Attributes == nil || wire.Resource["service.name"] != q.Question.Service || !selection.Matches(wire.Resource) || !q.Question.Namespace.Matches(wire.Resource) || wire.Scope != i.Scope || wire.Version != i.Version || wire.SchemaURL != i.SchemaURL || wire.Name != i.Metric || wire.Unit != i.Unit {
			return nil, fmt.Errorf("raw gauge identity differs from selected scope")
		}
		end, state := rawGaugeTimestamp(fields["end_ns"])
		if state != "valid" {
			return nil, fmt.Errorf("raw gauge requires observation time")
		}
		endNS, _ := strconv.ParseInt(*end, 10, 64)
		if endNS < q.Question.FromMS*1000000 || endNS >= q.Question.ToMS*1000000 {
			return nil, fmt.Errorf("raw gauge observation outside selected bounds")
		}
		row := RawGaugeObservation{OccurrenceKey: strconv.Itoa(index + 1), Stream: RawGaugeStream{ResourceSchemaURL: wire.ResourceSchemaURL, Resource: wire.Resource, Attributes: wire.Attributes, Scope: wire.Scope, Version: wire.Version, SchemaURL: wire.SchemaURL, Name: wire.Name, Unit: wire.Unit}, EndNS: *end, RawRow: bytes.Clone(raw), Diagnostics: []RawGaugeDiagnostic{}}
		row.StartNS, row.StartState = rawGaugeTimestamp(fields["start_ns"])
		diagnostic := func(field, code string) {
			row.Diagnostics = append(row.Diagnostics, RawGaugeDiagnostic{Field: field, Code: code})
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
		row.RawValue = bytes.Clone(fields["value"])
		row.Value, row.ValueState = rawGaugeValue(row.RawValue)
		if row.ValueState != "finite" {
			diagnostic("value", row.ValueState)
		}
		row.FlagsState = "absent"
		if flag, exists := fields["flags"]; exists && !bytes.Equal(bytes.TrimSpace(flag), []byte("null")) {
			flags, valid := rawGaugeFlags(flag)
			if !valid {
				row.FlagsState = "invalid"
			} else {
				row.Flags = &flags
				row.FlagsState = "valid"
				if flags & ^uint32(1) != 0 {
					diagnostic("flags", "unknown-bits")
				}
				if flags&1 != 0 {
					row.Value = nil
					row.ValueState = "absent"
					diagnostic("value", "no-recorded-value")
				}
			}
		}
		if row.FlagsState != "valid" {
			diagnostic("flags", row.FlagsState)
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}
func rawGaugeTimestamp(raw json.RawMessage) (*string, string) {
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
func rawGaugeValue(raw json.RawMessage) (*float64, string) {
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
func rawGaugeFlags(raw json.RawMessage) (uint32, bool) {
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

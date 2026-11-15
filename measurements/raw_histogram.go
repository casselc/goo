package measurements

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"

	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcefailure"
)

// RawHistogramInstrument selects an exact explicit-histogram exporter tuple.
// Unknown temporality is inspectable; it does not become evaluator-ready.
type RawHistogramInstrument struct {
	Kind        string `json:"kind"`
	Scope       string `json:"scope"`
	Version     string `json:"version"`
	SchemaURL   string `json:"schemaURL"`
	Metric      string `json:"metric"`
	Unit        string `json:"unit"`
	Temporality int32  `json:"temporality"`
}
type RawHistogramQuestion struct {
	Question   DiscoveryQuestion      `json:"question"`
	Instrument RawHistogramInstrument `json:"instrument"`
}

func (i *RawHistogramInstrument) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return fmt.Errorf("instrument requires an object")
	}
	for _, key := range []string{"kind", "scope", "version", "schemaURL", "metric", "unit", "temporality"} {
		raw, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("instrument requires explicit non-null %s", key)
		}
	}
	for key := range fields {
		switch key {
		case "kind", "scope", "version", "schemaURL", "metric", "unit", "temporality":
		default:
			return fmt.Errorf("unsupported instrument field %s", key)
		}
	}
	type wire RawHistogramInstrument
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*i = RawHistogramInstrument(value)
	return nil
}
func (q RawHistogramQuestion) Validate() error {
	if q.Instrument.Kind != "histogram" {
		return fmt.Errorf("require one exact explicit histogram instrument")
	}
	i := q.Instrument
	return validateRawObservation(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit})
}

// RawHistogramScalar retains a source column, not a claim that the producer set
// an optional OTel sum/min/max. The fixed exporter flattens absent values to zero.
type RawHistogramScalar struct {
	RawValue      json.RawMessage `json:"rawValue"`
	Value         *float64        `json:"value"`
	ValueState    string          `json:"valueState"`
	PresenceState string          `json:"presenceState"`
}
type RawHistogramObservation struct {
	RecordingState    string                     `json:"recordingState"`
	OccurrenceKey     string                     `json:"occurrenceKey"`
	Stream            RawObservationStream       `json:"stream"`
	StartNS           *string                    `json:"startNS"`
	EndNS             string                     `json:"endNS"`
	StartState        string                     `json:"startState"`
	Flags             *uint32                    `json:"flags"`
	FlagsState        string                     `json:"flagsState"`
	Temporality       int32                      `json:"temporality"`
	Count             *string                    `json:"count"`
	CountState        string                     `json:"countState"`
	Bounds            []float64                  `json:"bounds"`
	BoundsState       string                     `json:"boundsState"`
	BucketCounts      []string                   `json:"bucketCounts"`
	BucketCountsState string                     `json:"bucketCountsState"`
	BucketsState      string                     `json:"bucketsState"`
	Sum               RawHistogramScalar         `json:"sum"`
	Min               RawHistogramScalar         `json:"min"`
	Max               RawHistogramScalar         `json:"max"`
	Diagnostics       []RawObservationDiagnostic `json:"diagnostics"`
	RawRow            json.RawMessage            `json:"rawRow"`
}
type RawHistogramResult struct {
	Version    string                    `json:"version"`
	Question   DiscoveryQuestion         `json:"question"`
	Instrument RawHistogramInstrument    `json:"instrument"`
	ReadAt     string                    `json:"readAt"`
	Rows       []RawHistogramObservation `json:"rows"`
	Limit      int                       `json:"limit"`
}

func CompileRawHistogramObservations(q RawHistogramQuestion) (Plan, error) {
	return compileRawHistogramObservations(q, map[string]bool{"Sum": true, "Min": true, "Max": true})
}
func compileRawHistogramObservations(q RawHistogramQuestion, statistics map[string]bool) (Plan, error) {
	if err := q.Validate(); err != nil {
		return Plan{}, err
	}
	i := q.Instrument
	p, where, err := rawObservationPlan(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit})
	if err != nil {
		return Plan{}, err
	}
	key := fmt.Sprintf("p%d", len(p.Parameters))
	p.Parameters[key] = strconv.FormatInt(int64(i.Temporality), 10)
	where += " AND AggregationTemporality={" + key + ":Int32}"
	statistic := func(column, alias string) string {
		if statistics[column] {
			return "toString(" + column + ") AS " + alias
		}
		return "NULL AS " + alias
	}
	p.SQL = "SELECT mapUpdate(ResourceAttributes,map('service.name',toString(ServiceName))) AS resource, ResourceSchemaUrl AS resource_schema_url, ScopeName AS scope, ScopeVersion AS version, ScopeSchemaUrl AS schema_url, MetricName AS name, MetricUnit AS unit, Attributes AS attributes, toString(toUnixTimestamp64Nano(StartTimeUnix)) AS start_ns, toString(toUnixTimestamp64Nano(TimeUnix)) AS end_ns, Flags AS flags, AggregationTemporality AS temporality, toString(Count) AS count, arrayMap(c -> toString(c),BucketCounts) AS bucket_counts, arrayMap(b -> toString(b),ExplicitBounds) AS bounds, " + statistic("Sum", "sum") + ", " + statistic("Min", "min") + ", " + statistic("Max", "max") + " FROM otel_metrics_histogram WHERE " + where + " ORDER BY TimeUnix LIMIT 10001 FORMAT JSONEachRow"
	return p, nil
}
func (s Source) ReadRawHistogramObservations(ctx context.Context, q RawHistogramQuestion) (out *RawHistogramResult, resultErr error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	q.Question = q.Question.clone()
	defer func() { resultErr = sourcefailure.Protocol(resultErr) }()
	statistics := map[string]bool{}
	families, err := sourcecap.Inspect(ctx, func(ctx context.Context, sql string, params map[string]string) ([]json.RawMessage, error) {
		rows, err := s.read(ctx, Plan{SQL: sql, Parameters: params})
		if err == nil && sql == "DESCRIBE TABLE otel_metrics_histogram FORMAT JSONEachRow" {
			for _, raw := range rows {
				var column struct {
					Name string `json:"name"`
					Type string `json:"type"`
				}
				if json.Unmarshal(raw, &column) == nil && (column.Type == "Float64" || column.Type == "Nullable(Float64)") {
					switch column.Name {
					case "Sum", "Min", "Max":
						statistics[column.Name] = true
					}
				}
			}
		}
		return rows, err
	}, sourcecap.Histogram)
	if err != nil {
		return nil, err
	}
	if len(families) != 1 || !families[0].Compatible() {
		return nil, sourcefailure.Wrap(sourcefailure.Unavailable, fmt.Errorf("histogram metadata source unavailable"))
	}
	plan, err := compileRawHistogramObservations(q, statistics)
	if err != nil {
		return nil, err
	}
	rows, err := s.read(ctx, plan)
	if err != nil {
		return nil, err
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	out, err = rawHistogramRows(q, rows, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	if len(encoded) > 8<<20 {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("raw histogram result exceeds 8 MiB"))
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
func rawHistogramUnsigned(raw json.RawMessage) (*string, string) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, "absent"
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return nil, "invalid"
	}
	n, err := strconv.ParseUint(text, 10, 64)
	if err != nil || strconv.FormatUint(n, 10) != text {
		return nil, "invalid"
	}
	return &text, "valid"
}
func rawHistogramBounds(raw json.RawMessage) ([]float64, string) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, "absent"
	}
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, "invalid"
	}
	if len(values) > 10000 {
		return nil, "invalid"
	}
	out := make([]float64, 0, len(values))
	for _, raw := range values {
		value, state := rawObservationValue(raw)
		if state != "finite" || value == nil || math.IsNaN(*value) || len(out) > 0 && *value <= out[len(out)-1] {
			return nil, "invalid"
		}
		out = append(out, *value)
	}
	return out, "valid"
}
func rawHistogramCounts(raw json.RawMessage) ([]string, string) {
	if len(raw) == 0 || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, "absent"
	}
	var values []json.RawMessage
	if json.Unmarshal(raw, &values) != nil || values == nil || len(values) > 10001 {
		return nil, "invalid"
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		text, state := rawHistogramUnsigned(value)
		if state != "valid" {
			return nil, "invalid"
		}
		out = append(out, *text)
	}
	return out, "valid"
}
func rawHistogramScalar(raw json.RawMessage) RawHistogramScalar {
	if len(raw) == 0 {
		raw = json.RawMessage("null")
	}
	value, state := rawObservationValue(raw)
	return RawHistogramScalar{RawValue: bytes.Clone(raw), Value: value, ValueState: state, PresenceState: "unavailable"}
}
func rawHistogramRows(q RawHistogramQuestion, rows []json.RawMessage, at time.Time) (*RawHistogramResult, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	i := q.Instrument
	common, err := rawObservationRowsForFamily(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit}, rows, false)
	if err != nil {
		return nil, err
	}
	out := &RawHistogramResult{Version: "goo/raw-histogram-observations/v1", Question: q.Question.clone(), Instrument: i, ReadAt: at.UTC().Format(time.RFC3339Nano), Rows: []RawHistogramObservation{}, Limit: RowLimit}
	for _, base := range common {
		var fields map[string]json.RawMessage
		json.Unmarshal(base.RawRow, &fields)
		var tempo *int32
		if json.Unmarshal(fields["temporality"], &tempo) != nil || tempo == nil || *tempo != i.Temporality {
			return nil, fmt.Errorf("raw histogram temporality differs from selected instrument")
		}
		row := RawHistogramObservation{OccurrenceKey: base.OccurrenceKey, Stream: base.Stream, StartNS: base.StartNS, EndNS: base.EndNS, StartState: base.StartState, Flags: base.Flags, FlagsState: base.FlagsState, Temporality: *tempo, Diagnostics: base.Diagnostics, RawRow: base.RawRow}
		add := func(field, code string) {
			row.Diagnostics = append(row.Diagnostics, RawObservationDiagnostic{Field: field, Code: code})
		}
		if i.Temporality == 0 {
			add("temporality", "unspecified")
		} else if i.Temporality != 1 && i.Temporality != 2 {
			add("temporality", "unrecognized")
		}
		for _, budget := range []struct {
			key   string
			limit int
		}{{"bounds", 10000}, {"bucket_counts", 10001}} {
			var values []json.RawMessage
			if json.Unmarshal(fields[budget.key], &values) == nil && len(values) > budget.limit {
				return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("raw histogram %s exceeds per-point limit", budget.key))
			}
		}
		row.Count, row.CountState = rawHistogramUnsigned(fields["count"])
		row.Bounds, row.BoundsState = rawHistogramBounds(fields["bounds"])
		row.BucketCounts, row.BucketCountsState = rawHistogramCounts(fields["bucket_counts"])
		for _, fact := range []struct{ field, state string }{{"count", row.CountState}, {"bounds", row.BoundsState}, {"bucketCounts", row.BucketCountsState}} {
			if fact.state != "valid" {
				add(fact.field, fact.state)
			}
		}
		row.BucketsState = "valid"
		if row.CountState != "valid" && row.CountState != "absent" || row.BoundsState != "valid" && row.BoundsState != "absent" || row.BucketCountsState != "valid" && row.BucketCountsState != "absent" {
			row.BucketsState = "invalid"
			add("buckets", "invalid")
		} else if row.CountState == "absent" || row.BoundsState == "absent" || row.BucketCountsState == "absent" || len(row.Bounds) == 0 && len(row.BucketCounts) == 0 && row.BoundsState == "valid" && row.BucketCountsState == "valid" {
			row.BucketsState = "absent"
			add("buckets", "absent")

		} else if len(row.BucketCounts) != len(row.Bounds)+1 {
			row.BucketsState = "invalid"
			add("buckets", "shape")
		} else {
			sum := new(big.Int)
			for _, count := range row.BucketCounts {
				n, _ := new(big.Int).SetString(count, 10)
				sum.Add(sum, n)
			}
			if sum.String() != *row.Count {
				row.BucketsState = "invalid"
				add("buckets", "count-mismatch")
			}
		}
		row.RecordingState = "unknown"
		if row.FlagsState == "valid" && row.Flags != nil {
			if *row.Flags&1 != 0 {
				row.RecordingState = "no-recorded-value"
				row.Count = nil
				row.CountState = "absent"
				row.Bounds = nil
				row.BoundsState = "absent"
				row.BucketCounts = nil
				row.BucketCountsState = "absent"
				row.BucketsState = "absent"
				add("count", "no-recorded-value")
				add("buckets", "no-recorded-value")
			} else if *row.Flags == 0 {
				row.RecordingState = "recorded"
			}
		}
		row.Sum = rawHistogramScalar(fields["sum"])
		row.Min = rawHistogramScalar(fields["min"])
		row.Max = rawHistogramScalar(fields["max"])
		for _, fact := range []struct {
			field string
			value *RawHistogramScalar
		}{{"sum", &row.Sum}, {"min", &row.Min}, {"max", &row.Max}} {
			if fact.value.ValueState != "finite" {
				add(fact.field, fact.value.ValueState)
			}
			if row.RecordingState == "no-recorded-value" {
				fact.value.Value = nil
				fact.value.ValueState = "absent"
				add(fact.field, "no-recorded-value")
			}
			add(fact.field, "presence-unavailable")
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}

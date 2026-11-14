package measurements

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcefailure"
)

// RawSumInstrument selects one exact reported sum tuple. Discovery labels
// cumulative nonmonotonic sums stock; all other reported tuples remain counter.
type RawSumInstrument struct {
	Kind        string `json:"kind"`
	Scope       string `json:"scope"`
	Version     string `json:"version"`
	SchemaURL   string `json:"schemaURL"`
	Metric      string `json:"metric"`
	Unit        string `json:"unit"`
	Temporality int32  `json:"temporality"`
	Monotonic   bool   `json:"monotonic"`
}
type RawSumQuestion struct {
	Question   DiscoveryQuestion `json:"question"`
	Instrument RawSumInstrument  `json:"instrument"`
}

func (i *RawSumInstrument) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(data, &fields) != nil || fields == nil {
		return fmt.Errorf("instrument requires an object")
	}
	for _, key := range []string{"kind", "scope", "version", "schemaURL", "metric", "unit", "temporality", "monotonic"} {
		raw, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("instrument requires explicit non-null %s", key)
		}
	}
	type wire RawSumInstrument
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*i = RawSumInstrument(value)
	return nil
}
func (q RawSumQuestion) Validate() error {
	i := q.Instrument
	expected := "counter"
	if i.Temporality == 2 && !i.Monotonic {
		expected = "stock"
	}
	if i.Kind != expected {
		return fmt.Errorf("sum kind differs from reported temporality and monotonicity")
	}
	return validateRawObservation(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit})
}

// RawSumObservation is a reported occurrence, not a calculated interval or rate.
type RawSumObservation struct {
	RawObservation
	Temporality int32 `json:"temporality"`
	Monotonic   bool  `json:"monotonic"`
}
type RawSumResult struct {
	Version    string              `json:"version"`
	Question   DiscoveryQuestion   `json:"question"`
	Instrument RawSumInstrument    `json:"instrument"`
	ReadAt     string              `json:"readAt"`
	Rows       []RawSumObservation `json:"rows"`
	Limit      int                 `json:"limit"`
}

func CompileRawSumObservations(q RawSumQuestion) (Plan, error) {
	if err := q.Validate(); err != nil {
		return Plan{}, err
	}
	i := q.Instrument
	return compileRawObservation(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit}, &i)
}

// ReadRawSumObservations reads only the reported sum source; evaluator preflight,
// custody and stream identity are independent and never relaxed here.
func (s Source) ReadRawSumObservations(ctx context.Context, q RawSumQuestion) (out *RawSumResult, resultErr error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	q.Question = q.Question.clone()
	defer func() { resultErr = sourcefailure.Protocol(resultErr) }()
	families, err := sourcecap.Inspect(ctx, func(ctx context.Context, sql string, params map[string]string) ([]json.RawMessage, error) {
		return s.read(ctx, Plan{SQL: sql, Parameters: params})
	}, sourcecap.Counter)
	if err != nil {
		return nil, err
	}
	if len(families) != 1 || !families[0].Compatible() {
		return nil, sourcefailure.Wrap(sourcefailure.Unavailable, fmt.Errorf("sum metadata source unavailable"))
	}
	plan, err := CompileRawSumObservations(q)
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
	out, err = rawSumRows(q, rows, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, err
	}
	if len(encoded) > 8<<20 {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("raw sum result exceeds 8 MiB"))
	}
	if err = ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
func rawSumRows(q RawSumQuestion, rows []json.RawMessage, at time.Time) (*RawSumResult, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	i := q.Instrument
	common, err := rawObservationRows(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit}, rows)
	if err != nil {
		return nil, err
	}
	out := &RawSumResult{Version: "goo/raw-sum-observations/v1", Question: q.Question.clone(), Instrument: i, ReadAt: at.UTC().Format(time.RFC3339Nano), Rows: []RawSumObservation{}, Limit: RowLimit}
	for _, row := range common {
		var metadata struct {
			Temporality *int32 `json:"temporality"`
			Monotonic   *bool  `json:"monotonic"`
		}
		if json.Unmarshal(row.RawRow, &metadata) != nil || metadata.Temporality == nil || metadata.Monotonic == nil || *metadata.Temporality != i.Temporality || *metadata.Monotonic != i.Monotonic {
			return nil, fmt.Errorf("raw sum metadata differs from selected instrument")
		}
		if i.Temporality == 0 {
			row.Diagnostics = append(row.Diagnostics, RawObservationDiagnostic{Field: "temporality", Code: "unspecified"})
		} else if i.Temporality != 1 && i.Temporality != 2 {
			row.Diagnostics = append(row.Diagnostics, RawObservationDiagnostic{Field: "temporality", Code: "unrecognized"})
		}
		out.Rows = append(out.Rows, RawSumObservation{RawObservation: row, Temporality: *metadata.Temporality, Monotonic: *metadata.Monotonic})
	}
	return out, nil
}

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
	if q.Instrument.Kind != "gauge" {
		return fmt.Errorf("require a bounded exact gauge instrument")
	}
	i := q.Instrument
	return validateRawObservation(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit})
}

type RawGaugeResult struct {
	Version    string             `json:"version"`
	Question   DiscoveryQuestion  `json:"question"`
	Instrument RawGaugeInstrument `json:"instrument"`
	ReadAt     string             `json:"readAt"`
	Rows       []RawObservation   `json:"rows"`
	Limit      int                `json:"limit"`
}

// CompileRawGaugeObservations binds a half-open source read, not a measurement.
func CompileRawGaugeObservations(q RawGaugeQuestion) (Plan, error) {
	if err := q.Validate(); err != nil {
		return Plan{}, err
	}
	i := q.Instrument
	return compileRawObservation(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit}, nil)
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
	i := q.Instrument
	admitted, err := rawObservationRows(q.Question, rawObservationInstrument{i.Scope, i.Version, i.SchemaURL, i.Metric, i.Unit}, rows)
	if err != nil {
		return nil, err
	}
	return &RawGaugeResult{Version: "goo/raw-gauge-observations/v1", Question: q.Question.clone(), Instrument: i, ReadAt: at.UTC().Format(time.RFC3339Nano), Rows: admitted, Limit: RowLimit}, nil
}

package exploration

import (
	"fmt"
	"reflect"
	"strings"
)

// RecordMeasure is a typed operator produced by the compatibility adapter over
// existing question controls. It is not arbitrary SQL or a numeric cast policy.
type RecordMeasure struct {
	ID          string `json:"id"`
	Operator    string `json:"operator"`
	Input       *Field `json:"input,omitempty"`
	InputType   string `json:"inputType,omitempty"`
	OutputType  string `json:"outputType"`
	Unit        string `json:"unit"`
	Probability string `json:"probability,omitempty"`
	Merge       string `json:"merge"`
}
type RecordMeasurePlan struct {
	Scope            QuestionScope   `json:"scope"`
	Measures         []RecordMeasure `json:"measures"`
	Dimensions       []Field         `json:"dimensions"`
	BucketMS         int64           `json:"bucketMs"`
	GroupTimeSeries  bool            `json:"groupTimeSeries"`
	PercentilePolicy string          `json:"percentilePolicy,omitempty"`
}

func (q Query) RecordMeasures() (RecordMeasurePlan, error) {
	scope, e := (CaptureQuestion{Kind: "explore", Explore: &q}).Scope()
	if e != nil {
		return RecordMeasurePlan{}, e
	}
	return RecordMeasurePlan{Scope: scope, Measures: recordMeasures(q), Dimensions: append([]Field{}, q.GroupBy...), BucketMS: q.BucketMS, GroupTimeSeries: q.GroupTimeSeries, PercentilePolicy: func() string {
		if q.Summary != "" {
			return "exact-high: floor(p*N), clamped to N-1; no interpolation"
		}
		return ""
	}()}, nil
}
func recordMeasures(q Query) []RecordMeasure {
	out := []RecordMeasure{{ID: "count", Operator: "count", OutputType: "uint64", Unit: "{record}", Merge: "sum"}}
	if q.Summary == "span-duration" {
		for _, v := range []struct{ id, op, typ, merge, p string }{{"sum_ns", "sum", "uint128", "sum", ""}, {"min_ns", "min", "uint64", "min", ""}, {"max_ns", "max", "uint64", "max", ""}, {"mean_ns", "mean", "rational", "sum-and-count", ""}, {"p50_ns", "quantile", "uint64", "raw-values-required", "0.50"}, {"p95_ns", "quantile", "uint64", "raw-values-required", "0.95"}, {"p99_ns", "quantile", "uint64", "raw-values-required", "0.99"}} {
			field := Field{Kind: "column", Key: "duration_ns"}
			out = append(out, RecordMeasure{ID: v.id, Operator: v.op, Input: &field, InputType: "uint64", OutputType: v.typ, Unit: "ns", Probability: v.p, Merge: v.merge})
		}
	}
	if q.SpanStatus {
		field := Field{Kind: "column", Key: "status_class"}
		out = append(out, RecordMeasure{ID: "recorded_status", Operator: "partition-count", Input: &field, InputType: "string", OutputType: "partitioned-uint64", Unit: "{record}", Merge: "sum-aligned-partitions"})
	}
	return out
}

// Lower only planner-produced operators. The existing public controls expand to
// this bounded typed model; user-defined operators/casts are intentionally absent.
func recordAggregateSQL(q Query) (string, error) { return lowerRecordMeasures(recordMeasures(q)) }
func lowerRecordMeasures(measures []RecordMeasure) (string, error) {
	parts := []string{}
	quantiles := []string{}
	status := false
	for _, m := range measures {
		if m.Operator == "count" {
			if m.Input != nil || m.Unit != "{record}" {
				return "", fmt.Errorf("record count requires record grain")
			}
		} else if m.Operator == "partition-count" {
			if m.Input == nil || *m.Input != (Field{Kind: "column", Key: "status_class"}) || m.InputType != "string" || m.Unit != "{record}" {
				return "", fmt.Errorf("unsupported status binding")
			}
		} else if m.Operator == "sum" || m.Operator == "min" || m.Operator == "max" || m.Operator == "mean" || m.Operator == "quantile" {
			if m.Input == nil || *m.Input != (Field{Kind: "column", Key: "duration_ns"}) || m.InputType != "uint64" || m.Unit != "ns" {
				return "", fmt.Errorf("unsupported record numeric binding or unit; no implicit casts")
			}
		}
		switch m.Operator {
		case "count":
			parts = append(parts, "toString(count()) AS count")
		case "sum":
			parts = append(parts, "toString(sum(toUInt128(Duration))) AS sum_ns")
		case "min":
			parts = append(parts, "toString(min(Duration)) AS min_ns")
		case "max":
			parts = append(parts, "toString(max(Duration)) AS max_ns")
		case "mean": // Derived from exact sum/count in the result adapter.
		case "quantile":
			quantiles = append(quantiles, m.Probability)
		case "partition-count":
			status = true
		default:
			return "", fmt.Errorf("unsupported record operator %q", m.Operator)
		}
	}
	if len(quantiles) > 0 && !reflect.DeepEqual(quantiles, []string{"0.50", "0.95", "0.99"}) {
		return "", fmt.Errorf("unsupported percentile profile")
	}
	if len(quantiles) > 0 {
		parts = append(parts, "arrayMap(x -> toString(x), quantilesExactHigh("+strings.Join(quantiles, ", ")+")(Duration)) AS quantiles_ns")
	}
	sql := strings.Join(parts, ", ")
	if status {
		sql += statusAggregateSQL()
	}
	return sql, nil
}

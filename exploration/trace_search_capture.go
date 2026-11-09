package exploration

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/casselc/goo/identity"
)

// ValidateTraceSearchResult checks retained evidence without recompiling its
// historical plan or reading telemetry. Refresh separately checks current-plan
// equality. Qualified missing values remain absent, rather than empty strings.
func ValidateTraceSearchResult(result *TraceSearchResult) error {
	if result == nil {
		return errors.New("trace-search result required")
	}
	r := result.Request
	if err := r.Validate(); err != nil {
		return err
	}
	if _, err := time.Parse(time.RFC3339Nano, result.ReadAt); err != nil {
		return errors.New("invalid trace-search read clock")
	}
	if !regexpRevision.MatchString(result.Basis) || result.Items == nil || result.Warnings == nil || len(result.Items) > r.Limit {
		return errors.New("invalid trace-search evidence envelope")
	}
	total, err := decimal(result.MatchedTraces)
	if err != nil {
		return err
	}
	n := big.NewInt(int64(len(result.Items)))
	if total.Cmp(n) < 0 || result.HasMore != (total.Cmp(n) > 0) || result.HasMore && len(result.Items) != r.Limit {
		return errors.New("trace-search list and full count disagree")
	}
	if (total.Sign() == 0 && result.State != "no_data") || (total.Sign() > 0 && result.State != "observed") {
		return errors.New("trace-search state and full count disagree")
	}
	if strings.TrimSpace(result.Plan.SQL) == "" || len(result.Plan.SQL) > 256<<10 || !utf8.ValidString(result.Plan.SQL) || result.Plan.Parameters == nil || len(result.Plan.Parameters) > 1024 {
		return errors.New("invalid retained trace-search plan")
	}
	for k, v := range result.Plan.Parameters {
		if k == "" || len(k) > 256 || len(v) > 65536 || !utf8.ValidString(k) || !utf8.ValidString(v) {
			return errors.New("invalid retained trace-search parameters")
		}
	}
	if len(result.Warnings) == 0 || len(result.Warnings) > 16 {
		return errors.New("trace-search limitations required")
	}
	for _, warning := range result.Warnings {
		if warning == "" || len(warning) > 4096 || !utf8.ValidString(warning) {
			return errors.New("invalid trace-search limitation")
		}
	}
	selection, _ := r.EnvironmentSelection()
	last := ""
	for _, item := range result.Items {
		if !exampleID(item.TraceID, 32) || item.TraceID <= last || len(item.Witnesses) != len(r.Conditions) {
			return errors.New("invalid retained trace identity or witness count")
		}
		last = item.TraceID
		for i, w := range item.Witnesses {
			count, err := decimal(w.MatchingRecords)
			if err != nil || count.Sign() == 0 || w.ConditionIndex != i {
				return errors.New("invalid retained witness count or condition")
			}
			timestamp, ok := canonicalNS(w.TimestampNS)
			if !ok || timestamp < uint64(r.FromMS)*1000000 || timestamp >= uint64(r.ToMS)*1000000 {
				return errors.New("invalid retained witness time")
			}
			if _, ok := canonicalNS(w.DurationNS); !ok {
				return errors.New("invalid retained witness duration")
			}
			for _, v := range []string{w.SpanID, w.ParentSpanID, w.Service, w.Operation, w.Kind, w.TimestampNS, w.DurationNS} {
				if len(v) > 4096 || !utf8.ValidString(v) {
					return errors.New("invalid retained witness value")
				}
			}
			fields := traceSearchFields(r.Conditions[i])
			if len(w.FieldValues) != len(fields) {
				return errors.New("incomplete qualified witness fields")
			}
			values := map[Field]string{}
			resources := map[string]string{}
			expected := map[string]string{"service": w.Service, "name": w.Operation, "kind": w.Kind, "duration_ns": w.DurationNS, "timestamp_ns": w.TimestampNS, "span_id": w.SpanID, "trace_id": item.TraceID}
			for j, field := range fields {
				fv := w.FieldValues[j]
				if fv.Field != field || fv.Value == nil && !field.CanBeMissing() {
					return errors.New("invalid retained qualified witness field")
				}
				if fv.Value != nil {
					if len(*fv.Value) > 4096 || !utf8.ValidString(*fv.Value) {
						return errors.New("invalid qualified witness value")
					}
					values[field] = *fv.Value
					if field.Kind == "resource" {
						resources[field.Key] = *fv.Value
					}
				}
				if field.Kind == "column" {
					if want, known := expected[field.Key]; known && (fv.Value == nil || *fv.Value != want) {
						return errors.New("retained witness column disagrees with record")
					}
				}
			}
			if !Match(&r.Conditions[i], values) || !selection.Matches(resources) {
				return errors.New("retained witness does not satisfy condition or environment")
			}
		}
	}
	encoded, err := json.Marshal(result)
	if err != nil || len(encoded) > 8<<20 {
		return errors.New("trace-search evidence exceeds response budget")
	}
	return nil
}

// UnmarshalJSON makes new search requests agree with JSON.parse: later object
// values replace earlier ones, and null/unknown/case-aliased fields cannot become
// omitted defaults. This same request is used by live search and saved intent.
func (r *TraceSearchRequest) UnmarshalJSON(data []byte) error {
	*r = TraceSearchRequest{}
	if !identity.WellFormedJSON(data) {
		return errors.New("trace-search contains malformed Unicode")
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	if decoder.Decode(new(any)) != io.EOF {
		return errors.New("expected one trace-search request")
	}
	type wire TraceSearchRequest
	if !strictTraceSearchValue(value, reflect.TypeFor[wire](), 0) {
		return errors.New("trace-search requires known non-null fields")
	}
	object := value.(map[string]any)
	for _, key := range []string{"version", "fromMs", "toMs", "conditions", "limit"} {
		if _, ok := object[key]; !ok {
			return fmt.Errorf("trace-search requires %s", key)
		}
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var out wire
	if err = json.Unmarshal(canonical, &out); err != nil {
		return err
	}
	*r = TraceSearchRequest(out)
	return nil
}
func strictTraceSearchValue(value any, t reflect.Type, depth int) bool {
	if value == nil || depth > 32 {
		return false
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Struct:
		object, ok := value.(map[string]any)
		if !ok {
			return false
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name != "" && name != "-" {
				fields[name] = f.Type
			}
		}
		for key, v := range object {
			ft, ok := fields[key]
			if !ok || !strictTraceSearchValue(v, ft, depth+1) {
				return false
			}
		}
	case reflect.Slice:
		array, ok := value.([]any)
		if !ok {
			return false
		}
		for _, v := range array {
			if !strictTraceSearchValue(v, t.Elem(), depth+1) {
				return false
			}
		}
	}
	return true
}

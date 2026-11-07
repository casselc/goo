package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

func TestHegelHistogramExemplarContext377(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		environment := hegel.Draw(tc, hegel.Integers(0, 4))
		namespaceMode := hegel.Draw(tc, hegel.Integers(0, 2))
		bindNamespace := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
		spanMode := hegel.Draw(tc, hegel.Integers(0, 3))
		ordinal := hegel.Draw(tc, hegel.Integers(0, 2))
		var namespace *identity.Namespace
		if namespaceMode > 0 {
			namespace = &identity.Namespace{Mode: "value"}
			if namespaceMode == 2 {
				namespace.Value = "team-雪"
			}
		}
		parent, stream := nrcInputs(namespace, "histogram")
		parent.Definition.Source.Unit = "s"
		stream.Unit = "s"
		if !bindNamespace {
			parent.Definition.Source.Namespace = nil
		}
		if environment > 0 {
			parent.Definition.Environment = ""
			parent.Definition.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: fmt.Sprintf("prod-%d", seed)}}
			delete(stream.Resource, "deployment.environment.name")
			switch environment {
			case 1:
				parent.Definition.EnvironmentSelector.Policy = "current-only"
			case 3:
				parent.Definition.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
			case 4:
				parent.Definition.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
			}
			if environment != 4 {
				stream.Resource["deployment.environment.name"] = parent.Definition.EnvironmentSelector.Scope.Value
			}
		}
		parent.FromNS = strconv.FormatInt(parent.FromMS*1000000+int64(seed%1000000), 10)
		parent.ToNS = strconv.FormatInt(parent.ToMS*1000000-int64(seed%1000000), 10)
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			pinned := stream
			parent.Definition.Source.ExactStream = &pinned
		}
		trace := fmt.Sprintf("%032x", seed)
		span := fmt.Sprintf("%016x", seed)
		switch spanMode {
		case 1:
			span = ""
		case 2:
			span = strings.Repeat("0", 16)
		case 3:
			span = "invalid-reported-span"
		}
		start := strconv.FormatInt(parent.FromMS*1000000+1000000, 10)
		end := strconv.FormatInt(parent.FromMS*1000000+3000000000, 10)
		raw, err := json.Marshal(map[string]any{"resource_schema_url": stream.ResourceSchemaURL, "resource": stream.Resource, "scope": stream.Scope, "version": stream.Version, "schema_url": stream.SchemaURL, "name": stream.Name, "unit": stream.Unit, "attributes": stream.Attributes, "start_ns": start, "end_ns": end, "temporality": 1, "count": "3", "bucket_counts": []string{"0", "3", "0", "0"}, "bounds": []float64{0, .1, 1}, "flags": 0, "exemplar_total": "3", "exemplar_span_total": "3", "exemplar_time_total": "3", "exemplar_value_total": "3", "exemplar_attribute_total": "3", "exemplar_times": []string{"0", parent.FromNS, strconv.FormatInt(parent.ToMS*1000000+1, 10)}, "exemplar_values": []string{"0.05", "NaN", "Inf"}, "exemplar_trace_ids": []string{trace, trace, trace}, "exemplar_span_ids": []string{span, span, span}})
		if err != nil {
			tc.Fatal(err)
		}
		result, err := measurements.Evaluate(parent, []json.RawMessage{raw}, time.Date(2026, 11, 7, 0, 0, 0, 123, time.UTC))
		if err != nil {
			tc.Fatal(err)
		}
		before, _ := json.Marshal(result)
		context, err := exploration.PlanHistogramExemplarContext(result, stream, end, ordinal)
		if err != nil {
			tc.Fatal(err)
		}
		after, _ := json.Marshal(result)
		if string(before) != string(after) {
			tc.Fatal("exemplar projection mutated retained source")
		}
		if context.FromNS != parent.FromNS || context.ToNS != parent.ToNS || context.Question.FromMS != parent.FromMS || context.Question.ToMS != parent.ToMS || context.TraceID != trace || context.SourceRow != 0 || context.ExemplarIndex != ordinal || context.ExportStartNS != start || context.ExportEndNS != end || context.Question.Offset != 0 || context.Question.Limit != 100 || !reflect.DeepEqual(context.Question.EnvironmentSelector, parent.Definition.EnvironmentSelector) || context.Question.Environment != parent.Definition.Environment {
			tc.Fatal("reference/export/original scope identity changed")
		}
		if (context.SpanID != "") != (spanMode == 0) {
			tc.Fatal("invalid span hint became a trace predicate or valid hint")
		}
		if _, err = exploration.Compile(context.Question); err != nil {
			tc.Fatal(err)
		}
		from, _ := strconv.ParseInt(parent.FromNS, 10, 64)
		to, _ := strconv.ParseInt(parent.ToNS, 10, 64)
		for _, at := range []int64{from - 1, from, from + 1, to - 1, to, to + 1} {
			for _, sameTrace := range []bool{false, true} {
				for _, recordNamespace := range []*string{nil, new(string), func() *string { s := "team-雪"; return &s }(), func() *string { s := "other"; return &s }()} {
					record := map[exploration.Field]string{{Kind: "column", Key: "timestamp_ns"}: strconv.FormatInt(at, 10), {Kind: "column", Key: "trace_id"}: trace, {Kind: "column", Key: "service"}: "other-service", {Kind: "resource", Key: "service.instance.id"}: "other-instance", {Kind: "attribute", Key: "metric.only"}: "other-metric-value"}
					if !sameTrace {
						record[exploration.Field{Kind: "column", Key: "trace_id"}] = strings.Repeat("f", 32)
					}
					if recordNamespace != nil {
						record[exploration.Field{Kind: "resource", Key: identity.NamespaceKey}] = *recordNamespace
					}
					nsMatches := namespaceMode == 0 && recordNamespace == nil || namespaceMode > 0 && recordNamespace != nil && *recordNamespace == namespace.Value
					want := sameTrace && nsMatches && at >= from && at < to
					if nrcPredicate(context.Question.Where, record) != want {
						tc.Fatal("trace context lost exact membership or copied metric selectors", seed, at, sameTrace, recordNamespace)
					}
				}
			}
		}
		input, err := json.Marshal(map[string]any{"result": result, "stream": stream, "endNS": end, "ordinal": ordinal, "expected": context})
		if err != nil {
			tc.Fatal(err)
		}
		command := exec.CommandContext(t.Context(), "node", "histogram_exemplar_377_fixture.mjs")
		command.Stdin = bytes.NewReader(input)
		out, err := command.CombinedOutput()
		if err != nil {
			tc.Fatalf("actual JS exemplar context: %v\n%s", err, out)
		}
		var actual exploration.HistogramExemplarContext
		if err = json.Unmarshal(out, &actual); err != nil || !reflect.DeepEqual(actual, *context) {
			tc.Fatalf("Go/JS reported reference context differs: %v\n%s", err, out)
		}
	}, options(40)...)
}

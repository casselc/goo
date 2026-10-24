package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"gopkg.in/yaml.v3"
	"hegel.dev/go/hegel"
)

type etEvent struct {
	Name       string
	Attributes map[string]string
}
type etRecord struct {
	Environment, Signal, ID, Service, Status, Severity, SeverityNumber, Case, TraceID, Phase string
	TimeMS                                                                                   int64 `json:"timeMs"`
	Attributes, Resource                                                                     map[string]string
	Events                                                                                   []etEvent
}
type etProjection struct {
	Event string
	State string
	Type  *string
}

// Independent logical projection from ordinary event facts: sort reported values
// and compare neighbors, rather than invoking the production projection/parser.
func etProject(events []etEvent) etProjection {
	out := etProjection{Event: "no", State: "unreported"}
	values := []string{}
	for _, event := range events {
		if event.Name != "exception" {
			continue
		}
		out.Event = "yes"
		if value, ok := event.Attributes["exception.type"]; ok {
			values = append(values, value)
		}
	}
	sort.Strings(values)
	distinct := []string{}
	for _, value := range values {
		if len(distinct) == 0 || value != distinct[len(distinct)-1] {
			distinct = append(distinct, value)
		}
	}
	if len(distinct) > 1 {
		out.State = "multiple"
	} else if len(distinct) == 1 {
		out.State = "single"
		value := distinct[0]
		out.Type = &value
		if value == "" {
			out.State = "empty"
		}
	}
	return out
}
func etProjectedMap(p etProjection) map[exploration.Field]string {
	out := map[exploration.Field]string{{Kind: "column", Key: "exception_event"}: p.Event, {Kind: "column", Key: "exception_type_state"}: p.State}
	if p.Type != nil {
		out[exploration.Field{Kind: "column", Key: "exception_type"}] = *p.Type
	}
	return out
}
func etReported(row etRecord, rule string) bool {
	status := row.Status == "Error"
	typed := row.Attributes["error.type"] != ""
	event := etProject(row.Events).Event == "yes"
	if row.Signal == "logs" {
		n, e := strconv.ParseUint(row.SeverityNumber, 10, 64)
		status = e == nil && strconv.FormatUint(n, 10) == row.SeverityNumber && n >= 17 && n <= 24
		for _, text := range []string{"ERROR", "ERROR2", "ERROR3", "ERROR4", "FATAL", "FATAL2", "FATAL3", "FATAL4"} {
			status = status || row.Severity == text
		}
		for _, key := range []string{"exception.type", "exception.message", "exception.stacktrace"} {
			typed = typed || row.Attributes[key] != ""
		}
	}
	switch rule {
	case "status":
		return status
	case "type":
		return typed
	case "event":
		return event
	default:
		return status || typed || (row.Signal == "spans" && event)
	}
}
func etValues(row etRecord) map[exploration.Field]string {
	out := map[exploration.Field]string{{Kind: "column", Key: "service"}: row.Service}
	for key, value := range row.Attributes {
		out[exploration.Field{Kind: "attribute", Key: key}] = value
	}
	for key, value := range row.Resource {
		out[exploration.Field{Kind: "resource", Key: key}] = value
	}
	if row.Signal == "spans" {
		out[exploration.Field{Kind: "column", Key: "status"}] = row.Status
		state := "unknown"
		if row.Status == "Unset" {
			state = "unset"
		}
		if row.Status == "Error" {
			state = "error"
		} else if row.Status == "Ok" {
			state = "ok"
		}
		out[exploration.Field{Kind: "column", Key: "status_class"}] = state
		for f, v := range etProjectedMap(etProject(row.Events)) {
			out[f] = v
		}
	} else {
		out[exploration.Field{Kind: "column", Key: "severity"}] = row.Severity
		out[exploration.Field{Kind: "column", Key: "severity_number"}] = row.SeverityNumber
	}
	return out
}
func etEventCorpus(value string) [][]etEvent {
	e := func(name string, v *string) etEvent {
		attrs := map[string]string{}
		if v != nil {
			attrs["exception.type"] = *v
		}
		return etEvent{name, attrs}
	}
	empty := ""
	other := value + "-other"
	return [][]etEvent{nil, {}, {e("exception", nil)}, {e("exception", &empty)}, {e("exception", &value)}, {e("exception", &value), e("exception", &value)}, {e("exception", &value), e("exception", &other)}, {e("exception", &value), e("exception", &empty)}, {e("Exception", &value)}, {e("note", &other), e("exception", &value)}, {e("exception", nil), e("exception", &value)}, {e("note", &value), e("note", &empty)}}
}

func TestHegelErrorTriageProjectionAndRules(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		value := fmt.Sprintf("Domain-'\\雪-%d", seed)
		corpus := etEventCorpus(value)
		for _, events := range corpus {
			names := []string{}
			attrs := []map[string]string{}
			for _, event := range events {
				names = append(names, event.Name)
				attrs = append(attrs, event.Attributes)
			}
			if exploration.ExceptionEventPresent(names) != (etProject(events).Event == "yes") {
				tc.Fatal("names-only exception presence mismatch")
			}
			if !exploration.ExceptionEventPresent(append(append([]string{}, names...), "exception")) {
				tc.Fatal("named occurrence was erased by unavailable attributes")
			}
			before, _ := json.Marshal([]any{names, attrs})
			actual, e := exploration.ReferenceExceptionFields(names, attrs)
			if e != nil || !reflect.DeepEqual(actual, etProjectedMap(etProject(events))) {
				tc.Fatal("reported event projection mismatch", e, events, actual)
			}
			if after, _ := json.Marshal([]any{names, attrs}); string(after) != string(before) {
				tc.Fatal("projection mutated input")
			}
			// Repetition and reversal cannot multiply records or invent another type.
			reversed := append([]etEvent{}, events...)
			for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
				reversed[i], reversed[j] = reversed[j], reversed[i]
			}
			repeated := append(append([]etEvent{}, events...), reversed...)
			nn := []string{}
			aa := []map[string]string{}
			for _, event := range repeated {
				nn = append(nn, event.Name)
				aa = append(aa, event.Attributes)
			}
			again, e := exploration.ReferenceExceptionFields(nn, aa)
			if e != nil || !reflect.DeepEqual(actual, again) {
				tc.Fatal("event repetition changed record projection")
			}
			broken, e := exploration.ReferenceExceptionFields(append(names, "exception"), attrs)
			if e == nil || broken != nil {
				tc.Fatal("misaligned arrays did not fail closed")
			}
			broken, e = exploration.ReferenceExceptionFields(names, append(append([]map[string]string{}, attrs...), map[string]string{"exception.type": "extra-unmatched"}))
			if e == nil || broken != nil {
				tc.Fatal("extra attribute array entry did not fail closed")
			}

		}
		parent := exploration.Query{Signal: "spans", Environment: "triage-property", FromMS: 1700000000000, ToMS: 1700000060000, Limit: 1, Offset: seed % 100, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "resource", Key: "team"}, Op: "eq", Value: "core"}}
		raw, _ := json.Marshal(parent)
		script := `const t=require('../ui/error-triage');const q=JSON.parse(process.argv[1]);const out=[];for(const signal of ['spans','logs'])for(const rule of signal==='spans'?['reported','status','type','event']:['reported','status','type'])out.push({signal,rule,question:t.triageQuery({...q,signal},{rule})});process.stdout.write(JSON.stringify(out))`
		output, e := exec.Command("node", "-e", script, string(raw)).CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(output))
		}
		var queries []struct {
			Signal, Rule string
			Question     exploration.Query
		}
		if e = json.Unmarshal(output, &queries); e != nil || len(queries) != 7 {
			tc.Fatal("triage template wire invalid", e)
		}
		severities := []string{"INFO", "ERROR", "error", "FATAL4", "WARN", "custom", "ERROR2"}
		numbers := []string{"0", "16", "17", "24", "25", "18446744073709551615", "1"}
		for _, item := range queries {
			q := item.Question
			wantedType := exploration.Field{Kind: "column", Key: "exception_type"}
			if item.Signal == "logs" {
				wantedType = exploration.Field{Kind: "attribute", Key: "exception.type"}
			}
			if q.Signal != item.Signal || q.Environment != parent.Environment || q.FromMS != parent.FromMS || q.ToMS != parent.ToMS || q.Limit != 100 || q.Offset != 0 || q.BucketMS != 1000 || len(q.GroupBy) != 2 || q.GroupBy[0] != (exploration.Field{Kind: "column", Key: "service"}) {
				tc.Fatal("triage changed parent scope or bounded grouping")
			}
			if q.GroupBy[1] != wantedType {
				tc.Fatal("triage default type grouping changed")
			}
			if _, e := exploration.Compile(q); e != nil {
				tc.Fatal(e)
			}
			for i, events := range corpus {
				for _, team := range []string{"core", "outside"} {
					attrs := map[string]string{}
					switch (i + seed) % 6 {
					case 0:
						attrs["error.type"] = value
					case 1:
						attrs["error.type"] = ""
					case 2:
						attrs["exception.message"] = "caught"
					case 3:
						attrs["exception.type"] = ""
					case 4:
						attrs["exception.stacktrace"] = "stack"
					}
					row := etRecord{Signal: item.Signal, Status: []string{"Unset", "Ok", "Error"}[(i+seed)%3], Severity: severities[(i+seed)%len(severities)], SeverityNumber: numbers[(i+seed)%len(numbers)], Attributes: attrs, Resource: map[string]string{"team": team}, Events: events}
					if actual := exploration.Match(q.Where, etValues(row)); actual != (team == "core" && etReported(row, item.Rule)) {
						tc.Fatal("error evidence rule or parent scope mismatch", item.Signal, item.Rule, row, actual)
					}
				}
			}
			compiled, _ := exploration.Compile(q)
			if !etNoExplode(compiled.Count) {
				tc.Fatal("error counts exploded span events")
			}
		}
	}, options(70)...)
}

func etField(field exploration.Field, row etRecord) (*string, error) {
	if field.Kind == "resource" && field.Key == "service.name" {
		return &row.Service, nil
	}
	if field.Kind == "resource" && field.Key == "deployment.environment.name" {
		return &row.Environment, nil
	}
	if field.Kind == "attribute" || field.Kind == "resource" {
		values := row.Attributes
		if field.Kind == "resource" {
			values = row.Resource
		}
		value, ok := values[field.Key]
		if !ok {
			return nil, nil
		}
		return &value, nil
	}
	if field.Kind != "column" {
		return nil, fmt.Errorf("unsupported field kind %s", field.Kind)
	}
	if field.Key == "name" {
		v := "ordinary " + row.Case
		return &v, nil
	}
	if field.Key == "span_id" {
		return &row.ID, nil
	}
	if field.Key == "trace_id" {
		return &row.TraceID, nil
	}
	if field.Key == "timestamp_ns" {
		v := strconv.FormatInt(row.TimeMS*1000000, 10)
		return &v, nil
	}
	if field.Key == "exception_type" {
		return etProject(row.Events).Type, nil
	}
	values := etValues(row)
	value, ok := values[field]
	if !ok {
		return nil, fmt.Errorf("fixture lacks column %s", field.Key)
	}
	return &value, nil
}
func etMatch(p *exploration.Predicate, row etRecord) (bool, error) {
	if p == nil {
		return true, nil
	}
	if len(p.All) > 0 {
		for i := range p.All {
			yes, e := etMatch(&p.All[i], row)
			if e != nil || !yes {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			yes, e := etMatch(&p.Any[i], row)
			if e != nil {
				return false, e
			}
			if yes {
				return true, nil
			}
		}
		return false, nil
	}
	if p.Numeric != nil {
		plain := peLiveRecord{Environment: row.Environment, Signal: row.Signal, ID: row.ID, TimeMS: row.TimeMS, Service: row.Service, Status: row.Status, Severity: row.Severity, Attributes: row.Attributes, Resource: row.Resource}
		return peLiveMatch(p, plain)
	}
	if p.Field == nil {
		return false, fmt.Errorf("predicate lacks field")
	}
	value, e := etField(*p.Field, row)
	if e != nil {
		return false, e
	}
	switch p.Op {
	case "missing":
		return value == nil, nil
	case "exists":
		return value != nil, nil
	}
	if value == nil {
		return false, nil
	}
	if p.Field.Kind == "column" && p.Field.Key == "severity_number" {
		a, e := strconv.ParseUint(*value, 10, 64)
		if e != nil || strconv.FormatUint(a, 10) != *value {
			return false, fmt.Errorf("noncanonical severity number")
		}
		z, e := strconv.ParseUint(p.Value, 10, 64)
		if e != nil || strconv.FormatUint(z, 10) != p.Value {
			return false, fmt.Errorf("noncanonical numeric predicate")
		}
		switch p.Op {
		case "eq":
			return a == z, nil
		case "ne":
			return a != z, nil
		case "lt":
			return a < z, nil
		case "lte":
			return a <= z, nil
		case "gt":
			return a > z, nil
		case "gte":
			return a >= z, nil
		}
	}
	switch p.Op {
	case "eq":
		return *value == p.Value, nil
	case "ne":
		return *value != p.Value, nil
	case "contains":
		return strings.Contains(*value, p.Value), nil
	case "starts_with":
		return strings.HasPrefix(*value, p.Value), nil
	}
	return false, fmt.Errorf("unsupported independent predicate %s", p.Op)
}
func etRows(q exploration.Query, environment string, records []etRecord) ([]etRecord, error) {
	if q.Environment != environment || q.Signal != "spans" && q.Signal != "logs" || q.ToMS <= q.FromMS {
		return nil, fmt.Errorf("invalid source scope")
	}
	rows := []etRecord{}
	for _, row := range records {
		if row.Environment != environment || row.Signal != q.Signal || row.TimeMS < q.FromMS || row.TimeMS >= q.ToMS {
			continue
		}
		yes, e := etMatch(q.Where, row)
		if e != nil {
			return nil, e
		}
		if yes {
			rows = append(rows, row)
		}
	}
	return rows, nil
}
func etGroupKey(fields []exploration.Field, row etRecord) (string, error) {
	values := []*string{}
	for _, field := range fields {
		value, e := etField(field, row)
		if e != nil {
			return "", e
		}
		values = append(values, value)
	}
	raw, e := json.Marshal(values)
	return string(raw), e
}
func etSource(result exploration.Result, environment string, records []etRecord) bool {
	if result.Numeric != nil || result.NumericHistogram != nil || len(result.GroupedTimeSeries) > 0 {
		return false
	}
	rows, e := etRows(result.Question, environment, records)
	if e != nil || !etMeasures(result.Question, result.Duration, result.SpanStatus, rows) {
		return false
	}
	plain := make([]peLiveRecord, 0, len(rows))
	byID := map[string]etRecord{}
	for _, r := range rows {
		plain = append(plain, peLiveRecord{Environment: r.Environment, Signal: r.Signal, ID: r.ID, TimeMS: r.TimeMS, Service: r.Service, Status: r.Status, Severity: r.Severity, Attributes: r.Attributes, Resource: r.Resource})
		byID[r.ID] = r
	}
	// Membership was independently selected above. Reuse the ordinary OTLP raw-row
	// identity/clock/paging oracle with the predicate removed only from its copy.
	page := result
	page.Question.Where = nil
	if result.Question.Signal == "logs" {
		page.Rows = make([]json.RawMessage, len(result.Rows))
		for i, raw := range result.Rows {
			var record map[string]json.RawMessage
			var body string
			if json.Unmarshal(raw, &record) != nil || json.Unmarshal(record["Body"], &body) != nil {
				return false
			}
			id := etBodyID(body)
			fact, ok := byID[id]
			if !ok || body != "triage "+fact.Case+" "+fact.ID {
				return false
			}
			record["Body"], _ = json.Marshal("work observed " + id)
			page.Rows[i], _ = json.Marshal(record)
		}
	}
	if !peLiveSource(page, environment, plain) {
		return false
	}
	for _, raw := range result.Rows {
		var r map[string]json.RawMessage
		if json.Unmarshal(raw, &r) != nil {
			return false
		}
		var id string
		if result.Question.Signal == "logs" {
			var body string
			if json.Unmarshal(r["Body"], &body) != nil {
				return false
			}
			id = etBodyID(body)
		} else if json.Unmarshal(r["SpanId"], &id) != nil {
			return false
		}
		fact, ok := byID[id]
		if !ok {
			return false
		}
		for _, column := range []string{"exception_event", "exception_type", "exception_type_state", "severity_number"} {
			raw, exists := r[column]
			if !exists {
				continue
			}
			value, e := etField(exploration.Field{Kind: "column", Key: column}, fact)
			if e != nil {
				return false
			}
			var actual *string
			if json.Unmarshal(raw, &actual) != nil || !reflect.DeepEqual(value, actual) {
				return false
			}
		}
		if fact.Signal == "spans" {
			var duration, name string
			if json.Unmarshal(r["duration_ns"], &duration) != nil || duration != "1000000" || json.Unmarshal(r["SpanName"], &name) != nil || name != "ordinary "+fact.Case {
				return false
			}
			var trace string
			var times []string
			if json.Unmarshal(r["TraceId"], &trace) != nil || trace != fact.TraceID || json.Unmarshal(r["event_times_ns"], &times) != nil || len(times) != len(fact.Events) {
				return false
			}
			for i, at := range times {
				if at != strconv.FormatInt(fact.TimeMS*1000000+int64(i+1)*1000, 10) {
					return false
				}
			}
			var names []string
			var attrs []map[string]string
			if json.Unmarshal(r["event_names"], &names) != nil || json.Unmarshal(r["event_attributes"], &attrs) != nil || len(names) != len(fact.Events) || len(attrs) != len(fact.Events) {
				return false
			}
			for i, event := range fact.Events {
				if names[i] != event.Name || !reflect.DeepEqual(attrs[i], event.Attributes) {
					return false
				}
			}
		}
	}
	expected := map[string]int{}
	if len(result.Question.GroupBy) > 0 {
		for _, row := range rows {
			key, e := etGroupKey(result.Question.GroupBy, row)
			if e != nil {
				return false
			}
			expected[key]++
		}
	}
	actual := map[string]int{}
	for _, raw := range result.Groups {
		var group map[string]json.RawMessage
		if json.Unmarshal(raw, &group) != nil {
			return false
		}
		values := []*string{}
		for i := range result.Question.GroupBy {
			field, ok := group[fmt.Sprintf("g%d", i)]
			if !ok {
				return false
			}
			var value *string
			if json.Unmarshal(field, &value) != nil {
				return false
			}
			values = append(values, value)
		}
		var count string
		if json.Unmarshal(group["count"], &count) != nil {
			return false
		}
		n, e := strconv.Atoi(count)
		if e != nil || strconv.Itoa(n) != count || n <= 0 {
			return false
		}
		key, _ := json.Marshal(values)
		if _, exists := actual[string(key)]; exists {
			return false
		}
		actual[string(key)] = n
		selected := []etRecord{}
		for _, record := range rows {
			groupKey, e := etGroupKey(result.Question.GroupBy, record)
			if e != nil {
				return false
			}
			if groupKey == string(key) {
				selected = append(selected, record)
			}
		}
		var duration *exploration.DurationSummary
		var status *exploration.SpanStatusSummary
		if d, ok := group["duration"]; ok {
			if json.Unmarshal(d, &duration) != nil {
				return false
			}
		}
		if s, ok := group["spanStatus"]; ok {
			if json.Unmarshal(s, &status) != nil {
				return false
			}
		}
		if !etMeasures(result.Question, duration, status, selected) {
			return false
		}
	}
	if !reflect.DeepEqual(expected, actual) {
		return false
	}
	if result.Question.BucketMS > 0 {
		n := (result.Question.ToMS - result.Question.FromMS + result.Question.BucketMS - 1) / result.Question.BucketMS
		if int64(len(result.TimeSeries)) != n {
			return false
		}
		for i, point := range result.TimeSeries {
			from := result.Question.FromMS + int64(i)*result.Question.BucketMS
			to := from + result.Question.BucketMS
			if to > result.Question.ToMS {
				to = result.Question.ToMS
			}
			count := 0
			for _, row := range rows {
				if row.TimeMS >= from && row.TimeMS < to {
					count++
				}
			}
			state := "observed"
			if count == 0 {
				state = "no_data"
			}
			if point.FromMS != from || point.ToMS != to || point.Count != strconv.Itoa(count) || point.State != state || point.Numeric != nil || point.Duration != nil || point.SpanStatus != nil {
				return false
			}
		}
	} else if len(result.TimeSeries) > 0 {
		return false
	}
	return etNoExplode(result.Plan.Count)
}

func etBodyID(body string) string {
	at := strings.LastIndexByte(body, ' ')
	if at < 0 {
		return ""
	}
	return body[at+1:]
}

func etNoExplode(sql string) bool {
	upper := strings.ToUpper(sql)
	return !strings.Contains(upper, "ARRAY JOIN") && !strings.Contains(upper, "ARRAYJOIN(")
}

func etCohort(result exploration.Result) (exploration.Cohort, error) {
	c := exploration.Cohort{Question: result.Question, Basis: result.Basis, Count: result.Count, State: result.State, Groups: []exploration.ValueCount{}}
	if len(result.Question.GroupBy) != 1 {
		return c, fmt.Errorf("one field required")
	}
	for _, raw := range result.Groups {
		var g exploration.ValueCount
		if e := json.Unmarshal(raw, &g); e != nil {
			return c, e
		}
		c.Groups = append(c.Groups, g)
	}
	return c, nil
}
func etComparison(request exploration.ComparisonRequest, response exploration.CohortComparison, baseline, selection exploration.Result, environment string, records []etRecord) bool {
	if request.Numeric || request.Distribution || !etSource(baseline, environment, records) || !etSource(selection, environment, records) {
		return false
	}
	a, e := etCohort(baseline)
	if e != nil {
		return false
	}
	z, e := etCohort(selection)
	if e != nil {
		return false
	}
	if !reflect.DeepEqual(a, request.Baseline) || !reflect.DeepEqual(z, request.Selection) || a.Question.Signal != z.Question.Signal || a.Question.Environment != z.Question.Environment || !reflect.DeepEqual(a.Question.GroupBy, z.Question.GroupBy) || response.Measures != nil || response.AggregateOnly || response.Field != a.Question.GroupBy[0] || response.BaselineCount != a.Count || response.SelectionCount != z.Count || response.BaselineBasis != a.Basis || response.SelectionBasis != z.Basis {
		return false
	}
	canonical := request
	canonical.Baseline.Groups = append([]exploration.ValueCount{}, a.Groups...)
	canonical.Selection.Groups = append([]exploration.ValueCount{}, z.Groups...)
	for _, groups := range [][]exploration.ValueCount{canonical.Baseline.Groups, canonical.Selection.Groups} {
		sort.Slice(groups, func(i, j int) bool {
			return explanationValueKey(groups[i].Value) < explanationValueKey(groups[j].Value)
		})
	}
	raw, _ := json.Marshal(canonical)
	hash := sha256.Sum256(raw)
	if response.Basis != hex.EncodeToString(hash[:]) {
		return false
	}
	if a.State != "observed" || z.State != "observed" || a.Count == "0" || z.Count == "0" {
		return false
	}
	if response.State != "ok" || response.Reason != "" {
		return false
	}
	type pair struct {
		Value *string
		A, Z  int64
		Score *big.Rat
		Row   exploration.ValueComparison
	}
	expected := map[string]*pair{}
	for _, side := range []struct {
		groups    []exploration.ValueCount
		selection bool
	}{{a.Groups, false}, {z.Groups, true}} {
		for _, g := range side.groups {
			key := explanationValueKey(g.Value)
			if expected[key] == nil {
				expected[key] = &pair{Value: g.Value}
			}
			n, e := strconv.ParseInt(g.Count, 10, 64)
			if e != nil || n < 0 {
				return false
			}
			if side.selection {
				expected[key].Z = n
			} else {
				expected[key].A = n
			}
		}
	}
	ac, e := strconv.ParseInt(a.Count, 10, 64)
	if e != nil || ac <= 0 {
		return false
	}
	zc, e := strconv.ParseInt(z.Count, 10, 64)
	if e != nil || zc <= 0 {
		return false
	}
	keys := []string{}
	for key, p := range expected {
		ap := new(big.Rat).SetFrac(big.NewInt(p.A*100), big.NewInt(ac))
		zp := new(big.Rat).SetFrac(big.NewInt(p.Z*100), big.NewInt(zc))
		delta := new(big.Rat).Sub(zp, ap)
		p.Score = new(big.Rat).Abs(delta)
		display := delta.FloatString(2)
		if display == "-0.00" {
			display = "0.00"
		}
		p.Row = exploration.ValueComparison{Value: p.Value, BaselineCount: strconv.FormatInt(p.A, 10), SelectionCount: strconv.FormatInt(p.Z, 10), BaselinePercent: ap.FloatString(2), SelectionPercent: zp.FloatString(2), DifferencePP: display}
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		x, y := expected[keys[i]], expected[keys[j]]
		if n := x.Score.Cmp(y.Score); n != 0 {
			return n > 0
		}
		return keys[i] < keys[j]
	})
	if len(response.Rows) != len(keys) {
		return false
	}
	for i, key := range keys {
		if !reflect.DeepEqual(response.Rows[i], expected[key].Row) {
			return false
		}
	}
	return true
}

func etStoredCapture(stored exploration.SavedCapture, page exploration.CapturePage, environment string, records []etRecord) bool {
	if stored.Explore == nil || stored.Question.Kind != "explore" || stored.Question.Explore == nil || page.Explore == nil || stored.Explanation != nil || stored.Comparison != nil || stored.Measurement != nil || stored.Inventory != nil || !reflect.DeepEqual(stored.CaptureReceipt, page.Receipt) || !nhLiveSameQuery(stored.Explore.Question, *stored.Question.Explore) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Explore: stored.Explore}) {
		return false
	}
	rows, e := etRows(stored.Explore.Question, environment, records)
	if e != nil || len(rows) < 1 || len(rows) > 1000 || len(stored.Explore.Rows) != len(rows) || stored.Explore.HasMore || page.RetainedRecords != len(rows) || page.RecordsLimited {
		return false
	}
	n := len(rows)
	if n > 100 {
		n = 100
	}
	if !aeJSONEqual(page.Explore.Rows, stored.Explore.Rows[:n]) || page.Explore.HasMore != (len(rows) > n) || page.Explore.Count != stored.Explore.Count || page.Explore.Basis != stored.Explore.Basis || page.Explore.CapturedAt != stored.Explore.CapturedAt {
		return false
	}
	retained := *stored.Explore
	retained.Question.Limit = len(rows)
	return etSource(retained, environment, records) && etSource(*page.Explore, environment, records)
}

type etDetail struct {
	State, Title            string
	Types, Messages, Stacks []string
}
type etLiveEvent struct {
	Action                     string
	SourceOps, Streams, Writes int
}
type etLiveLedger struct {
	Version             int
	Environment, Signal string
	Records             []etRecord
	Selector            exploration.TailSelector
	View                workbench.ActivityView
	Applied             struct {
		Selector   exploration.TailSelector
		ReadWindow aeLiveWindowBounds
	}
	ActivityDetail, RecordDetail etDetail
	Trace                        struct {
		Spans, Logs, SelectedLogs exploration.Result
		Detail                    etDetail
	}
	Result, Rerun, Inspected, ParentAgain, WithinService, Previous exploration.Result
	States                                                         *exploration.Result
	Rules                                                          []struct {
		Rule   string
		Result exploration.Result
	}
	Saved, Reopened exploration.CapturePage
	Stored          exploration.SavedCapture
	Group           json.RawMessage
	Comparison      struct {
		Request  exploration.ComparisonRequest
		Response exploration.CohortComparison
	}
	Draft    exploration.PackagePlan
	Reusable exploration.SavedPackage
	Document workbench.Document
	Events   []etLiveEvent
	Requests []struct {
		Effect, URL string
		Body        json.RawMessage
	}
}

func etDetailCheck(detail etDetail, signal string, records []etRecord) bool {
	wanted := "handled-exception"
	if signal == "logs" {
		wanted = "multiple-error-signals"
	}
	var attrs map[string]string
	for _, record := range records {
		if record.Signal == signal && record.Case == wanted {
			if signal == "spans" {
				for _, event := range record.Events {
					if event.Name == "exception" {
						attrs = event.Attributes
						break
					}
				}
			} else {
				attrs = record.Attributes
			}
			break
		}
	}
	if attrs == nil || detail.State != "available" || !strings.Contains(detail.Title, "1 supplied occurrence") || !reflect.DeepEqual(detail.Types, []string{attrs["exception.type"]}) || !reflect.DeepEqual(detail.Stacks, []string{attrs["exception.stacktrace"]}) {
		return false
	}
	for _, message := range detail.Messages {
		if message == attrs["exception.message"] {
			return true
		}
	}
	return false
}
func etAppendWhere(parent *exploration.Predicate, children ...exploration.Predicate) *exploration.Predicate {
	parts := []exploration.Predicate{}
	if parent != nil {
		if len(parent.All) > 0 {
			parts = append(parts, parent.All...)
		} else {
			parts = append(parts, *parent)
		}
	}
	parts = append(parts, children...)
	return &exploration.Predicate{All: parts}
}
func etLiveScope(l etLiveLedger) bool {
	if l.Version != 1 || l.Saved.Receipt.Question.Explore == nil || l.Selector.Where == nil || l.Signal != "spans" && l.Signal != "logs" || l.Selector.Signal != l.Signal || l.Selector.Environment != l.Environment || !reflect.DeepEqual(l.Selector, l.Applied.Selector) || l.Applied.ReadWindow.ToMS-l.Applied.ReadWindow.FromMS != 60000 || !reflect.DeepEqual(l.View.Selector, l.Selector) {
		return false
	}
	q := l.Result.Question
	typ := exploration.Field{Kind: "column", Key: "exception_type"}
	if l.Signal == "logs" {
		typ = exploration.Field{Kind: "attribute", Key: "exception.type"}
	}
	if q.Signal != l.Signal || q.Environment != l.Environment || q.FromMS != l.Applied.ReadWindow.FromMS || q.ToMS != l.Applied.ReadWindow.ToMS || q.Limit != 100 || q.Offset != 0 || q.BucketMS != 1000 || q.Numeric != nil || q.NumericHistogram != nil || q.Summary != "" || q.SpanStatus || q.GroupTimeSeries || !reflect.DeepEqual(q.GroupBy, []exploration.Field{{Kind: "column", Key: "service"}, typ}) || q.Where == nil || len(q.Where.All) != 2 || !reflect.DeepEqual(q.Where.All[0], *l.Selector.Where) {
		return false
	}
	for _, row := range l.Records {
		parent, e := etMatch(l.Selector.Where, row)
		if e != nil {
			return false
		}
		actual, e := etMatch(q.Where, row)
		if e != nil || actual != (parent && etReported(row, "reported")) {
			return false
		}
	}
	for _, result := range []exploration.Result{l.Result, l.Rerun, l.Inspected, l.ParentAgain, l.WithinService, l.Previous} {
		if !etSource(result, l.Environment, l.Records) {
			return false
		}
	}
	if !nhLiveSameQuery(l.Rerun.Question, q) || !nhLiveSameQuery(l.ParentAgain.Question, q) || !nhLiveSameQuery(*l.Saved.Receipt.Question.Explore, q) || !etStoredCapture(l.Stored, l.Saved, l.Environment, l.Records) || !aeJSONEqual(l.Saved, l.Reopened) {
		return false
	}
	var group struct {
		Service, Type *string
		Count         string
	}
	var rawGroup map[string]json.RawMessage
	if json.Unmarshal(l.Group, &rawGroup) != nil || json.Unmarshal(rawGroup["g0"], &group.Service) != nil || json.Unmarshal(rawGroup["g1"], &group.Type) != nil || json.Unmarshal(rawGroup["count"], &group.Count) != nil || group.Service == nil || group.Type == nil || *group.Type != "IOError" {
		return false
	}
	found := false
	for _, raw := range l.Rerun.Groups {
		found = found || aeJSONEqual(raw, l.Group)
	}
	if !found {
		return false
	}
	service := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: *group.Service}
	typeFilter := exploration.Predicate{Field: &typ, Op: "eq", Value: *group.Type}
	within := comparisonClone(q)
	within.Where = etAppendWhere(q.Where, service)
	within.GroupBy = []exploration.Field{typ}
	inspect := comparisonClone(within)
	inspect.Where = etAppendWhere(q.Where, service, typeFilter)
	if !nhLiveSameQuery(l.WithinService.Question, within) || !nhLiveSameQuery(l.Inspected.Question, inspect) || l.Inspected.Count != group.Count {
		return false
	}
	previous := comparisonClone(within)
	previous.FromMS = within.FromMS - (within.ToMS - within.FromMS)
	previous.ToMS = within.FromMS
	if !nhLiveSameQuery(l.Previous.Question, previous) || !etComparison(l.Comparison.Request, l.Comparison.Response, l.Previous, l.WithinService, l.Environment, l.Records) {
		return false
	}
	if !etDetailCheck(l.ActivityDetail, l.Signal, l.Records) || !etDetailCheck(l.RecordDetail, l.Signal, l.Records) {
		return false
	}
	if l.Signal == "spans" {
		if l.States == nil || !etSource(*l.States, l.Environment, l.Records) {
			return false
		}
		statesQ := comparisonClone(q)
		statesQ.GroupBy = []exploration.Field{{Kind: "column", Key: "exception_type_state"}}
		if !nhLiveSameQuery(l.States.Question, statesQ) {
			return false
		}
		if !etTraceScopes(l) || !etSource(l.Trace.Spans, l.Environment, l.Records) || !etSource(l.Trace.Logs, l.Environment, l.Records) || !etSource(l.Trace.SelectedLogs, l.Environment, l.Records) || l.Trace.Spans.Count != "1" || l.Trace.Logs.Count != "0" || l.Trace.SelectedLogs.Count != "0" || !etDetailCheck(l.Trace.Detail, l.Signal, l.Records) {
			return false
		}
	} else if l.States != nil || l.Trace.Spans.Count != "" || l.Trace.Logs.Count != "" || l.Trace.SelectedLogs.Count != "" {
		return false
	}
	expectedRules := []string{"status", "type"}
	if l.Signal == "spans" {
		expectedRules = append(expectedRules, "event")
	}
	if len(l.Rules) != len(expectedRules) {
		return false
	}
	for i, item := range l.Rules {
		if item.Rule != expectedRules[i] || !etSource(item.Result, l.Environment, l.Records) {
			return false
		}
		want := comparisonClone(q)
		want.Where = item.Result.Question.Where
		if !nhLiveSameQuery(want, item.Result.Question) {
			return false
		}
		for _, r := range l.Records {
			parent, e := etMatch(l.Selector.Where, r)
			if e != nil {
				return false
			}
			yes, e := etMatch(item.Result.Question.Where, r)
			if e != nil || yes != (parent && etReported(r, item.Rule)) {
				return false
			}
		}
	}
	return etReusable(l)
}
func etReusable(l etLiveLedger) bool {
	if !nhLiveSameQuery(l.Draft.Question, l.WithinService.Question) || !reflect.DeepEqual(l.Draft.Definition, l.Reusable.Definition) || l.Draft.Revision != l.Reusable.Revision {
		return false
	}
	p := l.Reusable.Definition
	q := l.WithinService.Question
	if p.APIVersion != "goo/telemetry-package/v1alpha1" || p.Source.Adapter != "clickhouse.otel/v1" || p.Source.Signal != q.Signal || p.Source.Environment != q.Environment || p.Measure != "captured-record-count" || !reflect.DeepEqual(p.Where, q.Where) || !reflect.DeepEqual(p.GroupBy, q.GroupBy) || p.BucketMS != q.BucketMS || p.Numeric != nil || p.Summary != "" || p.SpanStatus || p.GroupTimeSeries || p.LookbackMinutes < 1 {
		return false
	}
	var decoded exploration.TelemetryPackage
	if yaml.Unmarshal([]byte(l.Draft.YAML), &decoded) != nil || !reflect.DeepEqual(decoded, p) {
		return false
	}
	raw, _ := json.Marshal(p)
	digest := sha256.Sum256(raw)
	if hex.EncodeToString(digest[:]) != l.Reusable.Revision {
		return false
	}
	capture, note := false, false
	for _, card := range l.Document.Cards {
		capture = capture || card.Kind == "capture" && card.Reference == l.Saved.Receipt.ID
		note = note || card.Kind == "note" && strings.Contains(card.Text, "Handled exception events") && strings.Contains(card.Text, "do not establish an incident or failure rate")
	}
	return capture && note && peLiveClock(l.Document.RecordedAt) && peLiveHex(l.Document.ID)
}

func etLiveRequests(l etLiveLedger) bool {
	expected := []exploration.Result{l.Result}
	if l.Signal == "spans" {
		expected = append(expected, l.Trace.Spans, l.Trace.Logs, l.Trace.SelectedLogs)
	}
	expected = append(expected, l.Rerun, l.Inspected, l.ParentAgain, l.WithinService, l.Previous)
	source, explore, streams, writes := 0, 0, 0, 0
	for _, r := range l.Requests {
		switch r.Effect {
		case "stream":
			var s exploration.TailSelector
			if r.URL != "/api/tail" || json.Unmarshal(r.Body, &s) != nil {
				return false
			}
			if s.Signal == "" {
				s.Signal = "logs"
			}
			if !reflect.DeepEqual(s, l.Selector) {
				return false
			}
			streams++
		case "source":
			source++
			if r.URL == "/api/captures" {
				var c struct {
					Title    string
					Question exploration.CaptureQuestion
				}
				if json.Unmarshal(r.Body, &c) != nil || c.Title != l.Saved.Receipt.Title || c.Question.Kind != "explore" || c.Question.Explore == nil || !nhLiveSameQuery(*c.Question.Explore, l.Result.Question) {
					return false
				}
				continue
			}
			var q exploration.Query
			if r.URL != "/api/explore" || explore >= len(expected) || json.Unmarshal(r.Body, &q) != nil || !nhLiveSameQuery(q, expected[explore].Question) {
				return false
			}
			explore++
		case "publication":
			if writes == 0 {
				var p struct{ YAML string }
				if r.URL != "/api/telemetry-packages" || json.Unmarshal(r.Body, &p) != nil || p.YAML != l.Draft.YAML {
					return false
				}
			} else if writes == 1 {
				var input struct {
					Parent   string
					Document workbench.DocumentInput
				}
				if r.URL != "/api/documents" || json.Unmarshal(r.Body, &input) != nil || input.Parent != l.Document.Parent || !reflect.DeepEqual(input.Document, l.Document.DocumentInput) {
					return false
				}
			} else {
				return false
			}
			writes++
		default:
			return false
		}
	}
	want := 7
	if l.Signal == "spans" {
		want = 10
	}
	return source == want && explore == len(expected) && streams == 1 && writes == 2
}
func etLiveReplay(t *testing.T, l etLiveLedger) bool {
	t.Helper()
	want := 21
	if l.Signal == "spans" {
		want = 22
	}
	if len(l.Events) != want {
		return false
	}
	sink := &qobs.MemorySink{}
	r := qobs.NewRecorder(sink)
	for _, event := range l.Events {
		r.Record(context.Background(), strings.ReplaceAll(event.Action, "-", "_"), nil, "obs.coherent", true, "obs.sources", event.SourceOps, "obs.streams", event.Streams, "obs.writes", event.Writes)
	}
	checker, e := validate.NewChecker("model/error_triage_"+l.Signal+".yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintErrorTriageLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_ERROR_TRIAGE_LEDGER")
	if file == "" {
		t.Skip("requires actual ordinary OTLP error triage journey ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	stored, e := os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.json"))
	if e != nil {
		t.Fatal(e)
	}
	check := func(control int) bool {
		var l etLiveLedger
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		if e = json.Unmarshal(stored, &l.Stored); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(l)
		switch control {
		case 0:
			l.Applied.ReadWindow.FromMS++
		case 1:
			l.Result.Question.Where = l.Selector.Where
		case 2:
			l.Result.Count += "1"
		case 3:
			var g map[string]any
			json.Unmarshal(l.Result.Groups[0], &g)
			g["count"] = "999999"
			l.Result.Groups[0], _ = json.Marshal(g)
		case 4:
			l.Result.TimeSeries[0].Count += "1"
		case 5:
			var row map[string]any
			json.Unmarshal(l.Result.Rows[0], &row)
			row["timestamp_ns"] = "0"
			l.Result.Rows[0], _ = json.Marshal(row)
		case 6:
			i := etDisplayedFactIndex(l)
			if i < 0 {
				t.Fatal("displayed source identity missing")
			}
			if l.Signal == "spans" {
				l.Records[i].Events = append(l.Records[i].Events, etEvent{Name: "exception", Attributes: map[string]string{"exception.type": "MutatedType"}})
			} else {
				l.Records[i].Attributes["exception.type"] = "MutatedType"
			}
		case 7:
			l.Stored.ID = "changed"
		case 8:
			l.Stored.Explore.Rows[0] = json.RawMessage(`{"SpanId":"changed"}`)
		case 9:
			l.Reopened.Receipt.Title += "changed"
		case 10:
			l.Inspected.Question.Where = l.Result.Question.Where
		case 11:
			l.WithinService.Question.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
		case 12:
			l.Previous.Question.Where = nil
		case 13:
			l.Previous.Question.FromMS++
		case 14:
			l.Comparison.Response.Rows[0].DifferencePP += "1"
		case 15:
			l.Comparison.Response.Basis = "changed"
		case 16:
			l.RecordDetail.Types = []string{"InventedType"}
		case 17:
			l.RecordDetail.Stacks = []string{"Invented stack"}
		case 18:
			l.ActivityDetail.Title = "Failure rate"
		case 19:
			for i := range l.Events {
				if l.Events[i].Action == "held-inspect" {
					l.Events[i].SourceOps++
				}
			}
		case 20:
			for i := range l.Events {
				if l.Events[i].Action == "reopen" {
					l.Events[i].SourceOps++
				}
			}
		case 21:
			for i := range l.Events {
				if l.Events[i].Action == "group-return" {
					l.Events[i].Streams++
				}
			}
		case 22:
			for i := range l.Events {
				if l.Events[i].Action == "presets" {
					l.Events[i].Writes++
				}
			}
		case 23:
			i := etDisplayedFactIndex(l)
			if i < 0 {
				t.Fatal("displayed source identity missing")
			}
			if l.Signal == "logs" {
				l.Records[i].SeverityNumber = "18446744073709551615"
			} else {
				if l.Records[i].Status == "Error" {
					l.Records[i].Status = "Ok"
				} else {
					l.Records[i].Status = "Error"
				}
			}
		case 24:
			l.Rules[0].Result.Question.Where = nil
		case 25:
			l.Draft.Definition.Where = nil
		case 26:
			l.Reusable.Revision = "changed"
		case 27:
			l.Document.Cards = nil
		case 28:
			l.Requests[0].URL = "/api/explore"
		case 29:
			l.Events[0], l.Events[1] = l.Events[1], l.Events[0]
		case 30:
			l.Inspected.CapturedAt = "invalid"
		case 31:
			l.ActivityDetail.Messages = nil
		case 32:
			l.ParentAgain.Question.ToMS++
		}
		if control >= 0 {
			after, _ := json.Marshal(l)
			if string(after) == string(before) {
				t.Fatalf("error triage control %d did not change evidence", control)
			}
		}
		semantic, requests := etLiveScope(l), etLiveRequests(l)
		if control < 0 {
			t.Logf("independent source/scope evidence=%v request discipline=%v result=%v capture=%v comparison=%v reusable=%v", semantic, requests, etSource(l.Result, l.Environment, l.Records), etStoredCapture(l.Stored, l.Saved, l.Environment, l.Records), etComparison(l.Comparison.Request, l.Comparison.Response, l.Previous, l.WithinService, l.Environment, l.Records), etReusable(l))
		}
		// Arithmetic/scope corruption is rejected by the independent oracle. Replay
		// the real action trace, including every lifecycle corruption that passes that
		// oracle, rather than repeatedly invoking Quint for already-invalid facts.
		return semantic && requests && etLiveReplay(t, l)
	}
	if !check(-1) {
		t.Fatal("actual error triage source/lifecycle evidence rejected")
	}
	for i := 0; i < 33; i++ {
		if check(i) {
			t.Fatalf("error triage control %d accepted", i)
		}
		t.Logf("error triage control %d rejected", i)
	}
}

type etOwnershipLedger struct {
	Version             int
	Environment, Signal string
	Selector            exploration.TailSelector
	View                workbench.ActivityView
	Applied             struct {
		Selector   exploration.TailSelector
		ReadWindow aeLiveWindowBounds
	}
	CurrentWindow int64
	Observations  []struct {
		Action, URL, AfterURL string
		PanelHidden           bool
		SourceOps, Streams    int
	}
	Events   []etLiveEvent
	Requests []struct {
		Effect, URL string
		Body        json.RawMessage
	}
}

func etOwnershipEvidence(l etOwnershipLedger) bool {
	if l.Version != 1 || l.Signal != "spans" || l.Selector.Signal != l.Signal || l.Selector.Environment != l.Environment || !reflect.DeepEqual(l.Selector, l.View.Selector) || !reflect.DeepEqual(l.Selector, l.Applied.Selector) || l.Applied.ReadWindow.ToMS-l.Applied.ReadWindow.FromMS != 60000 || l.CurrentWindow < l.Applied.ReadWindow.ToMS || len(l.Observations) != 4 || len(l.Requests) != 3 {
		return false
	}
	names := []string{"dirty-open", "applied-old-open", "hidden-open", "resumed-old-open"}
	streams := []int{1, 2, 2, 3}
	for i, o := range l.Observations {
		if o.Action != names[i] || o.SourceOps != 0 || o.Streams != streams[i] || !o.PanelHidden || o.URL == "" || o.AfterURL != o.URL {
			return false
		}
		if o.Action == "hidden-open" && !strings.HasSuffix(o.URL, "#definitions") {
			return false
		}
	}
	for _, r := range l.Requests {
		var s exploration.TailSelector
		if r.Effect != "stream" || r.URL != "/api/tail" || json.Unmarshal(r.Body, &s) != nil || !reflect.DeepEqual(s, l.Selector) {
			return false
		}
	}
	return true
}
func etOwnershipReplay(t *testing.T, l etOwnershipLedger) bool {
	t.Helper()
	if len(l.Events) != 8 {
		return false
	}
	sink := &qobs.MemorySink{}
	r := qobs.NewRecorder(sink)
	for _, event := range l.Events {
		r.Record(context.Background(), strings.ReplaceAll(event.Action, "-", "_"), nil, "obs.coherent", true, "obs.sources", event.SourceOps, "obs.streams", event.Streams, "obs.writes", event.Writes)
	}
	checker, e := validate.NewChecker("model/error_triage_ownership.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintErrorTriageOwnershipLedger(t *testing.T) {
	file := os.Getenv("GOO_ERROR_TRIAGE_OWNERSHIP_LEDGER")
	if file == "" {
		t.Skip("requires actual stale Activity triage owner ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	check := func(control int) bool {
		var l etOwnershipLedger
		if e = json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		before, _ := json.Marshal(l)
		switch control {
		case 0:
			l.Observations[0].AfterURL += "changed"
		case 1:
			l.Observations[1].PanelHidden = false
		case 2:
			l.Observations[2].SourceOps++
		case 3:
			l.Events[2].SourceOps++
		case 4:
			l.Events[6].Streams++
		case 5:
			l.Events[4].Writes++
		case 6:
			l.Applied.ReadWindow.ToMS++
		case 7:
			l.Applied.Selector.Where = nil
		case 8:
			l.CurrentWindow = l.Applied.ReadWindow.ToMS - 1
		case 9:
			l.Requests[0].URL = "/api/explore"
		case 10:
			l.Events[0], l.Events[1] = l.Events[1], l.Events[0]
		case 11:
			l.Observations[3].Action = "stale-owner-opened"
		case 12:
			l.Requests[1].Body = json.RawMessage(`{"signal":"spans","environment":"outside"}`)
		case 13:
			l.View.Selector.Environment += "outside"
		}
		if control >= 0 {
			after, _ := json.Marshal(l)
			if string(after) == string(before) {
				t.Fatalf("triage ownership control %d did not change evidence", control)
			}
		}
		evidence := etOwnershipEvidence(l)
		if control < 0 {
			t.Logf("stale-owner scope/navigation evidence=%v", evidence)
		}
		return evidence && etOwnershipReplay(t, l)
	}
	if !check(-1) {
		t.Fatal("actual triage stale-owner lifecycle rejected")
	}
	for i := 0; i < 14; i++ {
		if check(i) {
			t.Fatalf("triage ownership control %d accepted", i)
		}
		t.Logf("triage ownership control %d rejected", i)
	}
}

func etDisplayedFactIndex(l etLiveLedger) int {
	if len(l.Result.Rows) == 0 {
		return -1
	}
	var row map[string]json.RawMessage
	if json.Unmarshal(l.Result.Rows[0], &row) != nil {
		return -1
	}
	var id string
	if l.Signal == "logs" {
		var body string
		if json.Unmarshal(row["Body"], &body) != nil {
			return -1
		}
		id = etBodyID(body)
	} else {
		if json.Unmarshal(row["SpanId"], &id) != nil {
			return -1
		}
	}
	for i, record := range l.Records {
		if record.ID == id {
			return i
		}
	}
	return -1
}

// The ordinary OTLP fixture emits every span with a one-millisecond duration.
// These expectations use those raw fixture facts, never production summaries.
func etMeasures(q exploration.Query, duration *exploration.DurationSummary, status *exploration.SpanStatusSummary, rows []etRecord) bool {
	if q.Summary == "" {
		if duration != nil {
			return false
		}
	} else if q.Summary == "span-duration" {
		d := exploration.DurationSummary{State: "no_data", Count: strconv.Itoa(len(rows)), SumNS: "0"}
		if len(rows) > 0 {
			d.State = "observed"
			d.SumNS = strconv.FormatInt(int64(len(rows))*1000000, 10)
			d.MinNS = "1000000"
			d.MaxNS = "1000000"
			d.P50NS = "1000000"
			d.P95NS = "1000000"
			d.P99NS = "1000000"
		}
		if duration == nil || *duration != d {
			return false
		}
	} else {
		return false
	}
	if !q.SpanStatus {
		return status == nil
	}
	counts := map[string]int{}
	for _, r := range rows {
		switch r.Status {
		case "Error", "Ok", "Unset":
			counts[r.Status]++
		default:
			counts["unknown"]++
		}
	}
	expected := exploration.SpanStatusSummary{Count: strconv.Itoa(len(rows)), Error: strconv.Itoa(counts["Error"]), OK: strconv.Itoa(counts["Ok"]), Unset: strconv.Itoa(counts["Unset"]), Unknown: strconv.Itoa(counts["unknown"])}
	return status != nil && *status == expected
}

func etTraceScopes(l etLiveLedger) bool {
	if len(l.Trace.Spans.Rows) != 1 {
		return false
	}
	var row struct {
		ID    string `json:"SpanId"`
		Trace string `json:"TraceId"`
		Name  string `json:"SpanName"`
	}
	if json.Unmarshal(l.Trace.Spans.Rows[0], &row) != nil || row.Name != "ordinary handled-exception" {
		return false
	}
	where := &exploration.Predicate{All: []exploration.Predicate{{Field: &exploration.Field{Kind: "column", Key: "trace_id"}, Op: "eq", Value: row.Trace}}}
	base := exploration.Query{Signal: "spans", Environment: l.Environment, FromMS: l.Result.Question.FromMS, ToMS: l.Result.Question.ToMS, Limit: 100, Where: where, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}, Summary: "span-duration", SpanStatus: true}
	if !nhLiveSameQuery(l.Trace.Spans.Question, base) {
		return false
	}
	logs := exploration.Query{Signal: "logs", Environment: l.Environment, FromMS: base.FromMS, ToMS: base.ToMS, Limit: 100, Where: where}
	if !nhLiveSameQuery(l.Trace.Logs.Question, logs) {
		return false
	}
	logs.Where = comparisonClone(where)
	logs.Where.All = append(logs.Where.All, exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "span_id"}, Op: "eq", Value: row.ID})
	return nhLiveSameQuery(l.Trace.SelectedLogs.Question, logs)
}

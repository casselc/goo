package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

func tsLeaf(kind, key, op, value string) exploration.Predicate {
	return exploration.Predicate{Field: &exploration.Field{Kind: kind, Key: key}, Op: op, Value: value}
}
func tsGeneratedRequest(seed, count, limit int) exploration.TraceSearchRequest {
	metric := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.metric"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
	conditions := []exploration.Predicate{{All: []exploration.Predicate{tsLeaf("column", "service", "eq", "checkout"), tsLeaf("column", "duration_ns", "gte", "50")}}, tsLeaf("column", "status", "eq", "Error"), {Numeric: &metric, Op: "gte", Value: "9007199254740993"}, tsLeaf("resource", "team", "eq", "core")}
	from := int64(1800000000000 + seed*1000)
	return exploration.TraceSearchRequest{Version: exploration.TraceSearchVersion, Environment: "trace-search-property", FromMS: from, ToMS: from + 1000, Conditions: conditions[:count], Limit: limit}
}
func tsGeneratedRecord(request exploration.TraceSearchRequest, id string, ordinal, flags int) tsRecord {
	fields := map[exploration.Field]string{}
	put := func(kind, key, value string) { fields[exploration.Field{Kind: kind, Key: key}] = value }
	put("column", "trace_id", id)
	put("column", "span_id", fmt.Sprintf("%016x", ordinal+1))
	put("column", "parent_span_id", "")
	put("column", "timestamp_ns", strconv.FormatInt(request.FromMS*1000000+int64(ordinal), 10))
	put("column", "duration_ns", "49")
	put("column", "service", "other")
	put("column", "name", "observed work")
	put("column", "kind", "Internal")
	put("column", "status", "Ok")
	put("column", "status_class", "ok")
	put("resource", "team", "other")
	if flags&1 != 0 {
		put("column", "service", "checkout")
		put("column", "duration_ns", "50")
	}
	if flags&2 != 0 {
		put("column", "status", "Error")
		put("column", "status_class", "error")
	}
	if flags&4 != 0 {
		put("attribute", "work.metric", "9007199254740993")
	} else {
		raw := []string{"", "+9007199254740993", "09007199254740993", "9007199254740992", "9223372036854775808"}[ordinal%5]
		put("attribute", "work.metric", raw)
	}
	if flags&8 != 0 {
		put("resource", "team", "core")
	}
	return tsRecord{Environment: request.Environment, Values: fields}
}
func tsGeneratedRecords(request exploration.TraceSearchRequest, flags []int) []tsRecord {
	records := []tsRecord{}
	for i, mask := range flags {
		id := fmt.Sprintf("%032x", i+100)
		for j := 0; j < 4; j++ {
			records = append(records, tsGeneratedRecord(request, id, j, (mask>>(j*4))&15))
		}
	}
	// Guaranteed same-span intersection, duplicate records, separated false join,
	// invalid/all-zero IDs, and half-open/environment controls.
	shared := tsGeneratedRecord(request, "fffffffffffffffffffffffffffffff0", 0, 15)
	records = append(records, shared, shared)
	records = append(records, tsGeneratedRecord(request, "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeee1", 0, 1), tsGeneratedRecord(request, "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeee2", 0, 14))
	records = append(records, tsGeneratedRecord(request, strings.Repeat("0", 32), 0, 15), tsGeneratedRecord(request, "ABCDEFABCDEFABCDEFABCDEFABCDEFAB", 0, 15))
	wrong := tsGeneratedRecord(request, "ddddddddddddddddddddddddddddddd0", 0, 15)
	wrong.Environment += "-other"
	records = append(records, wrong)
	late := tsGeneratedRecord(request, "ddddddddddddddddddddddddddddddd1", 0, 15)
	late.Values[exploration.Field{Kind: "column", Key: "timestamp_ns"}] = strconv.FormatInt(request.ToMS*1000000, 10)
	records = append(records, late)
	return records
}
func tsCheckReference(request exploration.TraceSearchRequest, records []tsRecord) bool {
	want, total := tsOracle(request, records)
	logical := tsLogical(records)
	snapshot := func() string {
		out := make([]tsRecord, len(logical))
		for i, r := range logical {
			out[i] = tsRecord{Environment: r.Environment, Values: r.Fields}
		}
		return tsSnapshot(out)
	}
	before := snapshot()
	requestBefore, _ := json.Marshal(request)
	items, actual, e := exploration.ReferenceTraceSearch(request, logical)
	requestAfter, _ := json.Marshal(request)
	return e == nil && before == snapshot() && string(requestBefore) == string(requestAfter) && actual == total && tsItemsEqual(items, want)
}
func TestHegelTraceSearchIndependentIntersection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		count := hegel.Draw(tc, hegel.Integers(1, 4))
		limit := hegel.Draw(tc, hegel.Integers(1, 5))
		n := hegel.Draw(tc, hegel.Integers(2, 10))
		flags := make([]int, n)
		for i := range flags {
			flags[i] = hegel.Draw(tc, hegel.Integers(0, 65535))
		}
		request := tsGeneratedRequest(seed, count, limit)
		records := tsGeneratedRecords(request, flags)
		requestBefore, _ := json.Marshal(request)
		before := string(requestBefore) + tsSnapshot(records)
		if !tsCheckReference(request, records) {
			tc.Errorf("independent full-population intersection/witness mismatch")
			return
		}
		requestAfter, _ := json.Marshal(request)
		after := string(requestAfter) + tsSnapshot(records)
		if before != after {
			tc.Errorf("reference mutated request/records")
		}
		// Condition permutation changes only witness condition indexes/order.
		reversed := request
		reversed.Conditions = append([]exploration.Predicate{}, request.Conditions...)
		for i, j := 0, len(reversed.Conditions)-1; i < j; i, j = i+1, j-1 {
			reversed.Conditions[i], reversed.Conditions[j] = reversed.Conditions[j], reversed.Conditions[i]
		}
		if !tsCheckReference(reversed, records) {
			tc.Errorf("condition permutation mismatch")
		}
		_, oldTotal := tsOracle(request, records)
		_, reversedTotal := tsOracle(reversed, records)
		if oldTotal != reversedTotal {
			tc.Errorf("condition order changed trace intersection")
		}
		// Additional matching records cannot remove an ID. A newly matching lower ID
		// may move the bounded list prefix, so only the full population is monotone.
		more := append([]tsRecord{}, records...)
		more = append(more, tsGeneratedRecord(request, "00000000000000000000000000000001", 0, 15))
		if !tsCheckReference(request, more) {
			tc.Errorf("monotone added trace mismatch")
		}
		_, newTotal := tsOracle(request, more)
		oldN, _ := strconv.Atoi(oldTotal)
		if newTotal != strconv.Itoa(oldN+1) {
			tc.Errorf("matching record monotonicity lost")
		}
		// Whole input record permutation must preserve deterministic representatives.
		permuted := append([]tsRecord{}, records...)
		for i, j := 0, len(permuted)-1; i < j; i, j = i+1, j-1 {
			permuted[i], permuted[j] = permuted[j], permuted[i]
		}
		if !tsCheckReference(request, permuted) {
			tc.Errorf("record permutation changed witnesses")
		}
	}, options(90)...)
}
func TestTraceSearchSameSpanAndCrossTraceBoundary(t *testing.T) {
	r := tsGeneratedRequest(1, 2, 100)
	a := tsGeneratedRecord(r, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 0, 1)
	b := tsGeneratedRecord(r, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", 0, 2)
	if !tsCheckReference(r, []tsRecord{a, b}) {
		t.Fatal("cross trace reference mismatch")
	}
	items, total := tsOracle(r, []tsRecord{a, b})
	if total != "0" || len(items) != 0 {
		t.Fatal("conditions were joined across different trace IDs")
	}
	both := tsGeneratedRecord(r, "cccccccccccccccccccccccccccccccc", 0, 3)
	if !tsCheckReference(r, []tsRecord{both, both}) {
		t.Fatal("same-span/duplicate reference mismatch")
	}
	items, total = tsOracle(r, []tsRecord{both, both})
	if total != "1" || len(items) != 1 || items[0].Witnesses[0].SpanID != items[0].Witnesses[1].SpanID || items[0].Witnesses[0].MatchingRecords != "2" || items[0].Witnesses[1].MatchingRecords != "2" {
		t.Fatal("same-span matching or duplicate record count lost")
	}
}
func TestHegelTraceSearchNumericGuard(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		scale := hegel.Draw(tc, hegel.Integers(1, 9))
		decimalMode := hegel.Draw(tc, hegel.Integers(0, 1))
		r := tsGeneratedRequest(seed, 1, 100)
		binding := exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "work.metric"}, Type: "int64", Unit: "items", Parse: "canonical-decimal"}
		threshold := "9007199254740993"
		raws := []string{"9007199254740993", "9007199254740992", "-9007199254740993", "9223372036854775807", "9223372036854775808", "+1", "01", "-0", " 1", "1e1", ""}
		if decimalMode == 1 {
			binding.Type = "decimal64"
			binding.Unit = "USD"
			binding.Parse = "canonical-decimal"
			binding.Scale = scale
			threshold = decText("9007199254740993", scale)
			raws = []string{threshold, decText("9007199254740992", scale), decText("-9007199254740993", scale), decText("9223372036854775807", scale), decText("9223372036854775808", scale), "+" + threshold, "0" + threshold, threshold + "0", "-0." + strings.Repeat("0", scale), "1e1", ""}
		}
		records := []tsRecord{}
		for i, raw := range raws {
			record := tsGeneratedRecord(r, fmt.Sprintf("%032x", i+1), 0, 15)
			record.Values[binding.Field] = raw
			records = append(records, record)
		}
		missing := tsGeneratedRecord(r, "ffffffffffffffffffffffffffffffff", 0, 15)
		delete(missing.Values, binding.Field)
		records = append(records, missing)
		for _, op := range []string{"eq", "ne", "gt", "gte", "lt", "lte", "valid", "missing", "empty", "invalid"} {
			value := threshold
			if op == "valid" || op == "missing" || op == "empty" || op == "invalid" {
				value = ""
			}
			r.Conditions = []exploration.Predicate{{Numeric: &binding, Op: op, Value: value}}
			if !tsCheckReference(r, records) {
				tc.Errorf("typed parse guard mismatch type%s scale%d operator%s", binding.Type, scale, op)
			}
		}
	}, options(70)...)
}

func TestHegelTraceSearchContextualClientRequest(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		limit := hegel.Draw(tc, hegel.Integers(1, 100))
		request := tsGeneratedRequest(seed, 3, limit)
		parent := exploration.Query{Signal: "spans", Environment: request.Environment, FromMS: request.FromMS, ToMS: request.ToMS, Limit: 100, Offset: 100, Where: &request.Conditions[0], GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}}
		row := map[string]string{"ServiceName": "checkout", "SpanName": "observed work"}
		seedPredicate := exploration.Predicate{All: []exploration.Predicate{tsLeaf("column", "service", "eq", "checkout"), tsLeaf("column", "name", "eq", "observed work")}}
		input, _ := json.Marshal(map[string]any{"parent": parent, "row": row, "second": request.Conditions[1], "limit": limit})
		cmd := exec.Command("node", "-e", `const fs=require('fs'),m=require('../ui/trace-search'),x=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(x),seed=m.seedFromRow(x.row),first=m.seedCondition(x.parent,seed),request=m.buildRequest(x.parent,[first,x.second],x.limit),saved=structuredClone(request);request.conditions[0].op='mutated';process.stdout.write(JSON.stringify({request:saved,unchanged:before===JSON.stringify(x)}));`)
		cmd.Stdin = bytes.NewReader(input)
		raw, e := cmd.Output()
		if e != nil {
			tc.Errorf("actual contextual helper %v", e)
			return
		}
		var out struct {
			Request   exploration.TraceSearchRequest
			Unchanged bool
		}
		if json.Unmarshal(raw, &out) != nil || !out.Unchanged || out.Request.Validate() != nil || out.Request.Environment != parent.Environment || out.Request.FromMS != parent.FromMS || out.Request.ToMS != parent.ToMS || out.Request.Limit != limit || len(out.Request.Conditions) != 2 {
			tc.Errorf("client request lost typed scope/ownership %s", raw)
			return
		}
		records := tsGeneratedRecords(request, []int{65535, 4369, 8738, 4660})
		for _, record := range records {
			wantFirst := tsPredicate(*parent.Where, record.Values) && tsPredicate(seedPredicate, record.Values)
			if tsPredicate(out.Request.Conditions[0], record.Values) != wantFirst || tsPredicate(out.Request.Conditions[1], record.Values) != tsPredicate(request.Conditions[1], record.Values) {
				tc.Errorf("parent filters leaked into added existential condition")
			}
		}
		if !tsCheckReference(out.Request, records) {
			tc.Errorf("client request independent intersection mismatch")
		}
	}, options(70)...)
}
func TestHegelTraceSearchTypedCompilerBindings(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		value := fmt.Sprintf("tenant-%d'); DROP TABLE telemetry; --", seed)
		r := tsGeneratedRequest(seed, 2, 3)
		r.Conditions[0] = tsLeaf("attribute", "tenant'key", "eq", value)
		r.Conditions[1] = exploration.Predicate{Any: []exploration.Predicate{tsLeaf("resource", "tenant'key", "eq", value+" different"), tsLeaf("attribute", "present", "missing", "")}}
		plan, e := exploration.CompileTraceSearch(r)
		if e != nil {
			tc.Errorf("valid parameterized query %v", e)
			return
		}
		if strings.Contains(plan.SQL, value) || strings.Contains(plan.SQL, "tenant'key") || strings.Contains(plan.SQL, r.Environment) {
			tc.Errorf("user values/keys interpolated into SQL")
		}
		seen := map[string]bool{}
		for _, v := range plan.Parameters {
			seen[v] = true
		}
		for _, v := range []string{r.Environment, strconv.FormatInt(r.FromMS, 10), strconv.FormatInt(r.ToMS, 10), value, value + " different", "tenant'key", "present"} {
			if !seen[v] {
				tc.Errorf("condition scope/value lost typed parameter %q", v)
			}
		}
	}, options(70)...)
}

func TestTraceSearchQualifiedWitnessTie(t *testing.T) {
	r := tsGeneratedRequest(3, 1, 100)
	r.Conditions = []exploration.Predicate{{Any: []exploration.Predicate{tsLeaf("attribute", "same.key", "missing", ""), tsLeaf("resource", "same.key", "eq", "y")}}}
	a := tsGeneratedRecord(r, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 0, 15)
	a.Values[exploration.Field{Kind: "resource", Key: "same.key"}] = "y"
	b := tsGeneratedRecord(r, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 0, 15)
	b.Values[exploration.Field{Kind: "resource", Key: "same.key"}] = "y"
	b.Values[exploration.Field{Kind: "attribute", Key: "same.key"}] = ""
	records := []tsRecord{b, a}
	if !tsCheckReference(r, records) {
		t.Fatal("qualified missing/empty field tie lost")
	}
	items, total := tsOracle(r, records)
	if total != "1" || len(items) != 1 || items[0].Witnesses[0].MatchingRecords != "2" || len(items[0].Witnesses[0].FieldValues) != 2 || items[0].Witnesses[0].FieldValues[0].Field.Kind != "attribute" || items[0].Witnesses[0].FieldValues[0].Value != nil || items[0].Witnesses[0].FieldValues[1].Value == nil || *items[0].Witnesses[0].FieldValues[1].Value != "y" {
		t.Fatal("witness tie order or namespace collapsed")
	}
}

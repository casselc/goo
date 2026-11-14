package verification

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/querycel"
	"hegel.dev/go/hegel"
)

// The actual browser identity is checked once over the whole generated batch,
// rather than approximating its normalization in Go or launching Node per draw.
func TestHegelCELRepresentation436Identity(t *testing.T) {
	type pair struct {
		Before exploration.Query `json:"before"`
		After  exploration.Query `json:"after"`
	}
	var batch []pair
	hegel.Test(t, func(tc *hegel.T) {
		signal := hegel.Draw(tc, hegel.SampledFrom([]string{"logs", "spans"}))
		literal := hegel.Draw(tc, hegel.SampledFrom([]string{"", "quote'\"\\", "雪🦀\nline", "x' OR 1=1 --"}))
		threshold := hegel.Draw(tc, hegel.SampledFrom([]uint64{0, 17, 255, 9007199254740993, ^uint64(0)}))
		textop := hegel.Draw(tc, hegel.SampledFrom([]string{"eq", "ne", "contains"}))
		numop := hegel.Draw(tc, hegel.SampledFrom([]string{"eq", "ne", "gte", "lte"}))
		shape := hegel.Draw(tc, hegel.Integers(0, 3))
		textkey, numkey := "body", "severity_number"
		if signal == "spans" {
			textkey, numkey = "name", "duration_ns"
		}
		text := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: textkey}, Op: textop, Value: literal}
		number := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: numkey}, Op: numop, Value: strconv.FormatUint(threshold, 10)}
		service := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: "api"}
		predicate := text
		switch shape {
		case 1:
			predicate = exploration.Predicate{All: []exploration.Predicate{text, number}}
		case 2:
			predicate = exploration.Predicate{Any: []exploration.Predicate{text, number}}
		case 3:
			predicate = exploration.Predicate{All: []exploration.Predicate{{Any: []exploration.Predicate{text, number}}, service}}
		}
		source, err := querycel.PrintPredicate(signal, &predicate)
		if err != nil {
			tc.Fatal(err)
		}
		program, err := querycel.Compile(signal, source)
		if err != nil {
			tc.Fatal(err)
		}
		lowered := program.Predicate()
		reprinted, err := querycel.PrintPredicate(signal, lowered)
		if err != nil || source != reprinted {
			tc.Fatal("representation is not idempotent", source, reprinted, err)
		}
		q := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 17, Offset: 3, Where: &predicate}
		r := q
		r.Where = lowered
		batch = append(batch, pair{q, r})
		for _, body := range []string{literal, "ordinary", literal + "suffix"} {
			for _, n := range []uint64{0, 17, 255, 9007199254740993, ^uint64(0)} {
				for _, svc := range []string{"api", "other"} {
					textmatch := body == literal
					if textop == "ne" {
						textmatch = body != literal
					}
					if textop == "contains" {
						textmatch = strings.Contains(body, literal)
					}
					nummatch := n == threshold
					switch numop {
					case "ne":
						nummatch = n != threshold
					case "gte":
						nummatch = n >= threshold
					case "lte":
						nummatch = n <= threshold
					}
					want := textmatch
					switch shape {
					case 1:
						want = textmatch && nummatch
					case 2:
						want = textmatch || nummatch
					case 3:
						want = (textmatch || nummatch) && svc == "api"
					}
					values := map[exploration.Field]string{{Kind: "column", Key: textkey}: body, {Kind: "column", Key: numkey}: strconv.FormatUint(n, 10), {Kind: "column", Key: "service"}: svc}
					got, err := program.Eval(values)
					if err != nil || got != want || exploration.Match(lowered, values) != want {
						tc.Fatal("printed membership differs from independent oracle", source, err)
					}
				}
			}
		}
	}, options(120)...)
	if len(batch) == 0 {
		t.Fatal("no generated representations")
	}
	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	script := `import assert from 'node:assert/strict';import fs from 'node:fs';import {identity} from '../ui/question-session.mjs';const rows=JSON.parse(fs.readFileSync(0,'utf8'));for(const {before,after} of rows){assert.equal(identity(before),identity(after));assert.equal(identity(JSON.parse(identity(before))),identity(before));const changed=structuredClone(after);changed.where={field:{kind:'column',key:'service'},op:'eq',value:'dropped-condition'};assert.notEqual(identity(before),identity(changed));if(after.where?.all?.length===2){const reversed=structuredClone(after);reversed.where.all.reverse();assert.notEqual(identity(after),identity(reversed));}}console.log(rows.length+' real QuestionSession identity pairs');`
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(data)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual identity bridge: %v\n%s", err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

func TestCELRepresentation436RefusesPartialConversion(t *testing.T) {
	fixed := exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "eq", Value: "x"}
	for _, p := range []*exploration.Predicate{
		nil,
		{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "edge"},
		{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "exists"},
		{Field: &exploration.Field{Kind: "column", Key: "severity_number"}, Op: "eq", Value: "01"},
		{Numeric: &exploration.NumericBinding{Field: exploration.Field{Kind: "attribute", Key: "count"}, Type: "uint64", Unit: "1", Parse: "canonical-decimal"}, Op: "eq", Value: "1"},
		{All: []exploration.Predicate{fixed}},
		{All: []exploration.Predicate{fixed, fixed, fixed}},
		{All: []exploration.Predicate{fixed, {Field: &exploration.Field{Kind: "resource", Key: "zone"}, Op: "missing"}}},
	} {
		text, err := querycel.PrintPredicate("logs", p)
		var diagnostic *querycel.Diagnostic
		if text != "" || !errors.As(err, &diagnostic) {
			t.Fatalf("partial or untyped conversion: %q %v", text, err)
		}
	}
}

// Real Go receipts exercise the same pure JS owner future editor adapters use.
// Supplied parents contain predicates outside CEL's vocabulary and must survive.
func TestCELRepresentation436AuthoringInterop(t *testing.T) {
	type scenario struct {
		Base         exploration.Query    `json:"base"`
		BaseIdentity string               `json:"baseIdentity"`
		First        *querycel.RecordPlan `json:"first"`
		Second       *querycel.RecordPlan `json:"second"`
	}
	var batch []scenario
	for _, signal := range []string{"logs", "spans"} {
		field := "log.body"
		if signal == "spans" {
			field = "span.name"
		}
		base := exploration.Query{Signal: signal, Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 13, Offset: 4, Where: &exploration.Predicate{Any: []exploration.Predicate{
			{Field: &exploration.Field{Kind: "attribute", Key: "team"}, Op: "eq", Value: "edge"},
			{Field: &exploration.Field{Kind: "resource", Key: "zone"}, Op: "missing"},
		}}}
		firstRequest, err := querycel.NewPlanRequest(base, field+` == "first雪"`, "condition", "mount", "2")
		if err != nil {
			t.Fatal(err)
		}
		first, err := querycel.PlanRecord(firstRequest)
		if err != nil {
			t.Fatal(err)
		}
		secondRequest, err := querycel.NewPlanRequest(base, field+` == "second🦀"`, "condition", "mount", "4")
		if err != nil {
			t.Fatal(err)
		}
		second, err := querycel.PlanRecord(secondRequest)
		if err != nil {
			t.Fatal(err)
		}
		batch = append(batch, scenario{base, first.Binding.BaseIdentity, first, second})
	}
	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	script := `import assert from 'node:assert/strict';import fs from 'node:fs';import {AuthoringSlot} from '../ui/cel-authoring.mjs';import {identity} from '../ui/question-session.mjs';const rows=JSON.parse(fs.readFileSync(0,'utf8'));for(const row of rows){const owner=new AuthoringSlot({base:row.base,baseIdentity:row.baseIdentity,parentRevision:'1',slot:'condition',epoch:'mount'});owner.edit(row.first.expression);const first=owner.beginCheck();assert.deepEqual(first.request.binding,row.first.binding);assert.equal(identity(first.request.base),identity(row.base));assert.equal(owner.accept(first.token,row.first),true);assert.equal(identity(owner.stage(first.token)),identity(row.first.question));owner.edit(row.second.expression);const second=owner.beginCheck();assert.deepEqual(second.request.binding,row.second.binding);assert.equal(owner.accept(first.token,row.first),false);assert.equal(owner.stage(first.token),null);assert.equal(owner.accept(second.token,row.second),true);const staged=owner.stage(second.token);assert.equal(identity(staged),identity(row.second.question));assert.equal(identity({...staged,where:staged.where.all[0]}),identity(row.base));assert.equal(staged.where.all.length,2);const clear=owner.clear();assert.equal(identity(owner.stage(clear)),identity(row.base));owner.dispose();assert.equal(owner.stage(clear),null);}console.log(rows.length+' real Go→AuthoringSlot plans');`
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(data)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("actual authoring bridge: %v\n%s", err, out)
	} else {
		t.Log(strings.TrimSpace(string(out)))
	}
}

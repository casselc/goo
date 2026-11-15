package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

type rawContext449Fixture struct {
	Data               json.RawMessage
	Key                string
	FromMS, ToMS, AtNS int64
	Resource           map[string]string
}
type rawContext449Plan struct {
	Question                                             exploration.Query
	FromNS, ToNS, AtNS, RadiusNS, Scope, Relation, Grain string
	Origin                                               json.RawMessage
}

func rawContext449Data(t *testing.T, family int, q measurements.DiscoveryQuestion, row map[string]any) (json.RawMessage, string) {
	t.Helper()
	wire, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	var value any
	switch family {
	case 0:
		request := gauge438Question()
		request.Question = q
		result, e, _ := gauge438Read(t, metricScopeSchema432(t), []json.RawMessage{wire}, request)
		if e != nil {
			t.Fatal(e)
		}
		value = result
	case 1:
		request := sum441Question(2, false)
		request.Question = q
		result, e, _ := sum441Read(metricScopeSchema432(t)["counter"], []json.RawMessage{wire}, request)
		if e != nil {
			t.Fatal(e)
		}
		value = result
	default:
		request := histogram447Question(7)
		request.Question = q
		result, e, _ := histogram447Read(histogram447Schema(t), []json.RawMessage{wire}, request)
		if e != nil {
			t.Fatal(e)
		}
		value = result
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var bank struct {
		Rows []struct{ OccurrenceKey string }
	}
	if json.Unmarshal(raw, &bank) != nil || len(bank.Rows) != 1 {
		t.Fatal("source did not admit one occurrence")
	}
	return raw, bank.Rows[0].OccurrenceKey
}

func TestHegelRawObservationContext449BoundsAndScope(t *testing.T) {
	fixtures := []rawContext449Fixture{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 100000))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		nsMode := hegel.Draw(tc, hegel.Integers(0, 2))
		instanceMode := hegel.Draw(tc, hegel.Integers(0, 2))
		width := hegel.Draw(tc, hegel.SampledFrom([]int64{1, 2, 60000, 120001, 7200000}))
		position := hegel.Draw(tc, hegel.Integers(0, 2))
		fromMS := int64(1790000000000) + int64(seed)
		if seed%11 == 0 {
			fromMS = 0
		}
		toMS := fromMS + width
		lo, hi := fromMS*1000000, toMS*1000000
		at := []int64{lo, lo + (hi-lo)/2, hi - 1}[position]
		resource := map[string]string{"service.name": "api", "unrelated.resource": "not-a-filter"}
		q := measurements.DiscoveryQuestion{Service: "api", FromMS: fromMS, ToMS: toMS}
		switch mode {
		case 0:
			q.Environment = "prod"
			resource[identity.EnvironmentKey] = "prod"
		case 1:
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "current-only", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
			resource[identity.EnvironmentKey] = "prod"
		case 2:
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
			resource[identity.LegacyEnvironmentKey] = "prod"
		case 3:
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "empty"}}
			resource[identity.EnvironmentKey] = ""
			resource[identity.LegacyEnvironmentKey] = "prod"
		case 4:
			q.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "unreported"}}
		}
		if nsMode > 0 {
			resource[identity.NamespaceKey] = []string{"", "", "workers"}[nsMode]
		}
		if instanceMode > 0 {
			resource["service.instance.id"] = []string{"", "", "worker-" + strconv.Itoa(seed)}[instanceMode]
		}
		for family := range 3 {
			var row map[string]any
			switch family {
			case 0:
				row = gauge438Input("0", uint32(seed%3), 0)
			case 1:
				row = sum441Input(sum441Question(2, false), "0", uint32(seed%3), 0)
			case 2:
				row = histogram447Input(histogram447Question(7), 0, 9007199254740993)
				row["flags"] = uint32(seed % 3)
			}
			row["resource"] = resource
			row["end_ns"] = strconv.FormatInt(at, 10)
			row["attributes"] = map[string]string{"metric.only": "never-copy", "service.instance.id": "not-a-resource-identity"}
			data, key := rawContext449Data(t, family, q, row)
			fixtures = append(fixtures, rawContext449Fixture{Data: data, Key: key, FromMS: fromMS, ToMS: toMS, AtNS: at, Resource: resource})
		}
	}, options(40)...)
	script := `import fs from 'node:fs';import assert from 'node:assert/strict';import {plan} from '../ui/raw-observation-context.mjs';const cases=JSON.parse(fs.readFileSync(0,'utf8'));globalThis.fetch=()=>{throw Error('pure context planning must not acquire telemetry')};const outputs=[];for(const c of cases){const before=JSON.stringify(c.Data);outputs.push(['logs','spans'].map(signal=>plan(c.Data,c.Key,signal)));for(const signal of ['metrics','both','Spans',''])assert.throws(()=>plan(c.Data,c.Key,signal));assert.throws(()=>plan(c.Data,c.Key+'-missing','logs'));const duplicate=structuredClone(c.Data);duplicate.rows.push(structuredClone(duplicate.rows[0]));assert.throws(()=>plan(duplicate,c.Key,'logs'));assert.equal(JSON.stringify(c.Data),before)}process.stdout.write(JSON.stringify(outputs));`
	input, _ := json.Marshal(fixtures)
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("raw point context planner: %v\n%s", err, out)
	}
	var plans [][]rawContext449Plan
	if json.Unmarshal(out, &plans) != nil || len(plans) != len(fixtures) {
		t.Fatalf("incomplete context plans: %s", out)
	}
	for i, c := range fixtures {
		if len(plans[i]) != 2 {
			t.Fatal("missing family context")
		}
		for j, actual := range plans[i] {
			signal := []string{"logs", "spans"}[j]
			from, to := max(c.FromMS*1000000, c.AtNS-60000000000), min(c.ToMS*1000000, c.AtNS+60000000000)
			if actual.FromNS != strconv.FormatInt(from, 10) || actual.ToNS != strconv.FormatInt(to, 10) || actual.AtNS != strconv.FormatInt(c.AtNS, 10) || actual.RadiusNS != "60000000000" || actual.Relation != "raw-observation-context" || actual.Grain != "raw-observation-context" {
				t.Fatal("lost exact context geometry")
			}
			var bank struct {
				Question measurements.DiscoveryQuestion
			}
			if json.Unmarshal(c.Data, &bank) != nil {
				t.Fatal("invalid source fixture")
			}
			leaf := func(kind, key, op, value string) exploration.Predicate {
				return exploration.Predicate{Field: &exploration.Field{Kind: kind, Key: key}, Op: op, Value: value}
			}
			where := []exploration.Predicate{leaf("column", "service", "eq", "api")}
			if value, present := c.Resource[identity.NamespaceKey]; present {
				where = append(where, leaf("resource", identity.NamespaceKey, "eq", value))
			} else {
				where = append(where, leaf("resource", identity.NamespaceKey, "missing", ""))
			}
			scope := "service"
			if value, present := c.Resource["service.instance.id"]; present {
				scope = "reported-instance"
				where = append(where, leaf("resource", "service.instance.id", "eq", value))
			}
			where = append(where, leaf("column", "timestamp_ns", "gte", strconv.FormatInt(from, 10)), leaf("column", "timestamp_ns", "lte", strconv.FormatInt(to-1, 10)))
			expected := exploration.Query{Signal: signal, Environment: bank.Question.Environment, EnvironmentSelector: bank.Question.EnvironmentSelector, FromMS: from / 1000000, ToMS: (to + 999999) / 1000000, Limit: 100, Where: &exploration.Predicate{All: where}}
			if actual.Question.Validate() != nil || !reflect.DeepEqual(actual.Question, expected) || actual.Scope != scope {
				t.Fatalf("scope/context differs\n got %+v\nwant %+v scope %s", actual.Question, expected, scope)
			}
		}
	}
	t.Logf("%d actual Go-admitted banks, %d independently checked nearby questions", len(fixtures), len(fixtures)*2)
}

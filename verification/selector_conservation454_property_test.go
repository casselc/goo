package verification

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestHegelSelectorConservation454(t *testing.T) {
	type entry struct {
		Name   string          `json:"name"`
		Query  json.RawMessage `json:"query"`
		Limit  int             `json:"limit"`
		Offset int             `json:"offset"`
	}
	raw, err := os.ReadFile("../tests/fixtures/selector-conservation454.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank struct {
		Version        string
		Cases, Invalid []entry
	}
	if err = json.Unmarshal(raw, &bank); err != nil || bank.Version != "goo/selector-conservation/v1" || len(bank.Cases) != 9 {
		t.Fatal("invalid casebank", err)
	}
	type fixture struct {
		Query    exploration.Query `json:"query"`
		Limit    int               `json:"limit"`
		Offset   int               `json:"offset"`
		Expected exploration.Query `json:"-"`
	}
	fixtures := []fixture{}
	hegel.Test(t, func(tc *hegel.T) {
		family := hegel.Draw(tc, hegel.Integers(0, 4))
		mode := hegel.Draw(tc, hegel.Integers(0, 4))
		numeric := hegel.Draw(tc, hegel.Integers(0, 1))
		projection := hegel.Draw(tc, hegel.Integers(0, 1))
		start := int64(1700000000000 + hegel.Draw(tc, hegel.Integers(0, 1000000)))
		width := int64(hegel.Draw(tc, hegel.Integers(1, 7200000)))
		limit := hegel.Draw(tc, hegel.Integers(1, 100))
		offset := hegel.Draw(tc, hegel.Integers(0, 1000))
		index := family
		if family == 4 {
			index = 8
		}
		var parent exploration.Query
		if err := json.Unmarshal(bank.Cases[index].Query, &parent); err != nil {
			t.Fatal(err)
		}
		parent.FromMS, parent.ToMS = start, start+width
		parent.BucketMS = 60000
		if family != 4 {
			parent.Environment = "prod"
			parent.EnvironmentSelector = nil
			if mode > 0 {
				parent.Environment = ""
				parent.EnvironmentSelector = &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
				if mode == 1 {
					parent.EnvironmentSelector.Policy = "current-only"
				}
				if mode == 3 {
					parent.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "empty"}
				}
				if mode == 4 {
					parent.EnvironmentSelector.Scope = identity.EnvironmentScope{Mode: "unreported"}
				}
			}
		}
		if numeric == 1 && family != 4 {
			var donor exploration.Query
			if err := json.Unmarshal(bank.Cases[7].Query, &donor); err != nil {
				t.Fatal(err)
			}
			parent.Numeric, parent.NumericHistogram = donor.Numeric, donor.NumericHistogram
			parent.Where.All = append(parent.Where.All, donor.Where.All[len(donor.Where.All)-1])
		}
		if projection == 0 {
			parent.Numeric = nil
			parent.NumericHistogram = nil
			parent.GroupBy = nil
			parent.Summary = ""
			parent.SpanStatus = false
			parent.LogSeverity = false
			parent.BucketMS = 0
			parent.GroupTimeSeries = false
		}
		if err := parent.Validate(); err != nil {
			t.Fatal("generated original refused", err)
		}
		before, _ := json.Marshal(parent)
		out, err := exploration.PopulationQuery(parent, limit, offset)
		if err != nil {
			t.Fatal(err)
		}
		// Independently enumerate population restrictions, not projection-removal steps.
		expected := exploration.Query{Signal: parent.Signal, Environment: parent.Environment, EnvironmentSelector: parent.EnvironmentSelector, FromMS: parent.FromMS, ToMS: parent.ToMS, Where: parent.Where, ReportedEdge: parent.ReportedEdge, LinksTo: parent.LinksTo, Limit: limit, Offset: offset}
		if !reflect.DeepEqual(out, expected) || out.Validate() != nil {
			t.Fatal("population changed")
		}
		fixtures = append(fixtures, fixture{parent, limit, offset, expected})
		if out.Where != nil {
			out.Where.Value = "changed"
			if len(out.Where.All) > 0 {
				out.Where.All[0].Field.Key = "changed"
				last := &out.Where.All[len(out.Where.All)-1]
				if last.Numeric != nil {
					last.Numeric.Unit = "changed"
				}
			}
		}
		if out.LinksTo != nil {
			out.LinksTo.TraceID = "changed"
		}
		if out.ReportedEdge != nil {
			out.ReportedEdge.Parent.Service = "changed"
		}
		if out.EnvironmentSelector != nil {
			out.EnvironmentSelector.Scope.Value = "changed"
		}
		after, _ := json.Marshal(parent)
		if !bytes.Equal(before, after) {
			t.Fatal("Go projection aliases parent")
		}
		bad := parent
		bad.Limit = 0
		if _, err := exploration.PopulationQuery(bad, 100, 0); err == nil {
			t.Fatal("parent page repaired")
		}
		bad = parent
		bad.GroupTimeSeries = true
		bad.GroupBy = nil
		if _, err := exploration.PopulationQuery(bad, 100, 0); err == nil {
			t.Fatal("invalid original projection erased")
		}
	}, options(60)...)
	for _, bad := range bank.Invalid {
		var q exploration.Query
		if err := json.Unmarshal(bad.Query, &q); err == nil {
			if out, err := exploration.PopulationQuery(q, bad.Limit, bad.Offset); err == nil || !reflect.DeepEqual(out, exploration.Query{}) {
				t.Fatalf("invalid wire returned partial/admitted output %s", bad.Name)
			}
		}
	}
	payload, err := json.Marshal(struct {
		Fixtures []fixture `json:"fixtures"`
		Invalid  []entry   `json:"invalid"`
	}{fixtures, bank.Invalid})
	if err != nil {
		t.Fatal(err)
	}
	script := `import fs from 'node:fs';import {populationQuery} from '../ui/record-population.mjs';globalThis.fetch=()=>{throw Error('unexpected source IO')};const data=JSON.parse(fs.readFileSync(0,'utf8')),results=[];for(const f of data.fixtures){const before=JSON.stringify(f.query),out=populationQuery(f.query,{limit:f.limit,offset:f.offset});results.push(structuredClone(out));for(const key of ['numeric','numericHistogram','groupBy','summary','spanStatus','logSeverity','bucketMs','groupTimeSeries'])if(Object.hasOwn(out,key))throw Error('projection retained');if(out.where){out.where.value='changed';if(out.where.all?.length){out.where.all[0].field.key='changed';const last=out.where.all.at(-1);if(last.numeric)last.numeric.unit='changed'}}if(out.linksTo)out.linksTo.traceId='changed';if(out.reportedEdge)out.reportedEdge.parent.service='changed';if(out.environmentSelector)out.environmentSelector.scope.value='changed';if(JSON.stringify(f.query)!==before)throw Error('input aliased');for(const mutate of [q=>q.limit=0,q=>{q.groupTimeSeries=true;delete q.groupBy},q=>q.linksTo=null,q=>q.where=null]){const bad=structuredClone(f.query);mutate(bad);let refused=false;try{populationQuery(bad)}catch{refused=true}if(!refused)throw Error('invalid parent repaired')}}for(const bad of data.invalid){let refused=false;try{populationQuery(bad.query,{limit:bad.limit,offset:bad.offset})}catch{refused=true}if(!refused)throw Error('invalid shared wire admitted '+bad.name)}process.stdout.write(JSON.stringify(results));`
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(payload)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JS projection %v %s", err, out)
	}
	var plans []exploration.Query
	if err := json.Unmarshal(out, &plans); err != nil {
		t.Fatal(err, string(out))
	}
	if len(plans) != len(fixtures) {
		t.Fatal("lost plan")
	}
	for i, q := range plans {
		if !reflect.DeepEqual(q, fixtures[i].Expected) || q.Validate() != nil {
			t.Fatalf("Go/JS selectors differ fixture %d", i)
		}
	}
	t.Logf("%d Hegel casebank variants: exact none/edge/links/both/log selectors, policies, windows/pages, numeric predicate retained, projections stripped, clones and seven wire refusals", len(fixtures))
}

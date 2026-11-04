package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

func TestHegelKeptTraceContinuation327(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := hegel.Draw(tc, hegel.SampledFrom([]string{"legacy", "current", "named", "empty", "unreported"}))
		signal := []string{"logs", "spans"}[seed%2]
		name := fmt.Sprintf("prod-%d'雪", seed%13)
		scope := identity.EnvironmentScope{Mode: mode}
		if mode == "legacy" || mode == "current" || mode == "named" {
			scope.Mode = "named"
			scope.Value = name
		}
		selected, e := identity.EffectiveEnvironment(scope)
		if mode == "legacy" || mode == "current" {
			selected, e = identity.CurrentEnvironment(name)
		}
		if e != nil {
			tc.Fatal(e)
			return
		}
		selector, _ := selected.Selector()
		ns := identity.Namespace{Mode: "missing"}
		if seed%3 != 0 {
			ns.Mode = "value"
			if seed%3 == 2 {
				ns.Value = "team'雪"
			}
		}
		hash := func(v any) string { b, _ := json.Marshal(v); h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
		base := int64(1790000000000) + int64(seed)*1000
		record := func(from int64, index int) json.RawMessage {
			resource := map[string]string{"service.name": "api"}
			switch mode {
			case "legacy", "current":
				resource[identity.EnvironmentKey] = name
			case "named":
				resource[identity.LegacyEnvironmentKey] = name
			case "empty":
				resource[identity.EnvironmentKey] = ""
				resource[identity.LegacyEnvironmentKey] = name
			}
			if ns.Mode == "value" {
				resource[identity.NamespaceKey] = ns.Value
			}
			r := map[string]any{"timestamp_ns": strconv.FormatInt(from*1000000+int64(index)+1, 10), "ServiceName": "api", "ResourceAttributes": resource, "Attributes": map[string]string{"tenant": fmt.Sprint(seed)}, "TraceId": fmt.Sprintf("%032x", seed+index), "SpanId": fmt.Sprintf("%016x", seed+index)}
			if signal == "logs" {
				r["Body"] = "kept log"
			} else {
				r["SpanName"] = "kept span"
			}
			raw, _ := json.Marshal(r)
			return raw
		}
		tailSelector := exploration.TailSelector{Signal: signal, EnvironmentSelector: &selector, Where: &exploration.Predicate{All: []exploration.Predicate{*exploration.NamespacePredicate(&ns), {Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "eq", Value: fmt.Sprint(seed)}}}}
		if signal == "logs" && seed%3 == 0 {
			tailSelector.Signal = ""
		}
		if mode == "legacy" {
			tailSelector.Environment = name
			tailSelector.EnvironmentSelector = nil
		}
		tail := exploration.KeptObservations{Version: exploration.KeptObservationsVersion, ID: strings.Repeat("a", 64), IntentDigest: strings.Repeat("b", 64), Connection: strings.Repeat("c", 64), KeptAt: time.UnixMilli(base + 600000), Selector: tailSelector}
		for i := 0; i < 2; i++ {
			from := base + int64(i)*120000
			r := record(from, i)
			tail.Members = append(tail.Members, exploration.KeptObservation{Key: hash(r) + ":1", FromMS: from, ToMS: from + 60000, ObservedAt: time.UnixMilli(from + 61000), Basis: strings.Repeat("d", 64), Limited: i == 1, Late: i == 1, Record: r})
		}
		tail.Digest = hash(tail)
		if e = tail.Validate(); e != nil {
			tc.Fatal("invalid actual Go tail fixture", e)
			return
		}
		q := exploration.Query{Signal: signal, Environment: tailSelector.Environment, EnvironmentSelector: tailSelector.EnvironmentSelector, FromMS: base + 300000, ToMS: base + 420000, Limit: 100, Offset: seed % 4 * 100, GroupBy: []exploration.Field{{Kind: "attribute", Key: "cohort"}}, Where: tailSelector.Where}
		if signal == "spans" {
			q.Summary = "span-duration"
			q.SpanStatus = true
		}
		page := exploration.KeptRecords{Version: exploration.KeptRecordsVersion, ID: strings.Repeat("e", 64), IntentDigest: strings.Repeat("f", 64), KeptAt: time.UnixMilli(base + 600000), Origin: "result-page", Question: q, ObservedAt: time.UnixMilli(base + 421000).UTC().Format(time.RFC3339Nano), Basis: "original-page", State: "observed", ReturnedRecords: 10, RecordsLimited: true}
		for i := 0; i < 2; i++ {
			r := record(q.FromMS, i)
			page.Members = append(page.Members, exploration.KeptRecord{Index: i + 2, Digest: hash(r), Record: r})
		}
		page.Digest = hash(page)
		if e = page.Validate(); e != nil {
			tc.Fatal("invalid actual Go page fixture", e)
			return
		}
		capture := page
		capture.Origin = "saved-capture"
		capture.CaptureID = strings.Repeat("1", 64)
		capture.CaptureAt = time.UnixMilli(base + 500000).UTC().Format(time.RFC3339Nano)
		capture.Digest = ""
		capture.Digest = hash(capture)
		if e = capture.Validate(); e != nil {
			tc.Fatal("invalid actual Go capture fixture", e)
			return
		}
		input, _ := json.Marshal([]any{tail, page, capture})
		cmd := exec.Command("node", "--input-type=module", "-e", keptTrace327JS)
		cmd.Stdin = bytes.NewReader(input)
		out, e := cmd.CombinedOutput()
		if e != nil {
			tc.Fatal(e, string(out))
			return
		}
		var views []workbench.ExplorationView
		if e = json.Unmarshal(out, &views); e != nil {
			tc.Fatal(e, string(out))
			return
		}
		if len(views) != 6 {
			tc.Fatal("lost origin/member cases")
			return
		}
		for _, v := range views {
			if e = v.Validate(); e != nil {
				tc.Fatal(e)
				return
			}
			if _, e = exploration.Compile(v.Queries[0]); e != nil {
				tc.Fatal(e)
				return
			}
		}
	}, options(30)...)
}

const keptTrace327JS = `
import assert from 'node:assert/strict';import fs from 'node:fs';import * as K from '../ui/kept-saved.mjs';import * as E from '../ui/record-environment.mjs';import * as C from '../ui/view-composition.mjs';import * as S from '../ui/question-session.mjs';
const values=JSON.parse(fs.readFileSync(0,'utf8')),views=[];
for(const value of values){const before=JSON.stringify(value),tail=value.version==='goo/kept-observations/v1',origin=tail?value.selector:value.question;for(let index=0;index<2;index++){const member=value.members[index],from=tail?member.fromMs:origin.fromMs,to=tail?member.toMs:origin.toMs,v=K.traceView(value,index),q=v.queries[0];assert.equal(q.signal,'spans');assert(E.sameQuestionEnvironment(q,origin));assert.deepEqual(E.questionFields(q),E.questionFields(origin));assert.equal(q.fromMs,from);assert.equal(q.toMs,to);assert.equal(q.offset,0);assert.equal(q.limit,100);assert.deepEqual(q.where,{all:[{field:{kind:'column',key:'trace_id'},op:'eq',value:member.record.TraceId}]});assert.equal(q.groupBy,undefined);for(const key of ['summary','spanStatus','numeric','numericHistogram','bucketMs','groupTimeSeries','reportedEdge'])assert(!q[key]);assert(C.recordOnly(v.composition));assert.equal(C.records(v.composition).kind.name,'record-span-waterfall');views.push(v);const other=K.traceView(value,index);q.where.all[0].value='changed';assert.equal(JSON.stringify(value),before);assert.equal(other.queries[0].where.all[0].value,member.record.TraceId);q.where.all[0].value=member.record.TraceId;
for(const trace of [undefined,'0'.repeat(32),'not-a-trace','a'.repeat(31)]){const bad=structuredClone(value);if(trace===undefined)delete bad.members[index].record.TraceId;else bad.members[index].record.TraceId=trace;assert.throws(()=>K.traceView(bad,index));}}
for(const index of [-1,2,0.5,'0',null])assert.throws(()=>K.traceView(value,index));const foreign=structuredClone(value);foreign.members[0].record.ResourceAttributes['deployment.environment.name']='foreign';assert.throws(()=>K.traceView(foreign,0));assert.equal(JSON.stringify(value),before);const edited=structuredClone(value),planned=K.traceView(edited,0),expected=S.identity(planned.queries[0]);edited.members[0].record.TraceId='9'.repeat(32);if(tail)edited.members[0].fromMs--;else edited.question.fromMs--;assert.equal(S.identity(planned.queries[0]),expected);}
process.stdout.write(JSON.stringify(views));
`

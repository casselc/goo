package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

type scopedViewEvent struct {
	Action     string `json:"action"`
	Reads      int    `json:"reads"`
	Generation int    `json:"generation"`
}

// Replay observed source effects, not clicks. The browser supplies the source-read
// counter and harness ordinal of observed successful applied Runs after each
// operation has settled; this is not the opaque dataset scope-generation token.
func TestQuintScopedViewLedger(t *testing.T) {
	path := os.Getenv("GOO_SCOPED_LEDGER")
	if path == "" {
		t.Skip("observed scoped-view browser ledger required")
	}
	raw, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var events []scopedViewEvent
	if e = json.Unmarshal(raw, &events); e != nil {
		t.Fatal(e)
	}
	seen := map[string]bool{}
	for _, event := range events {
		seen[event.Action] = true
	}
	if len(events) < 6 || events[0] != (scopedViewEvent{Action: "run", Reads: 1, Generation: 1}) {
		t.Fatal("missing initial observed run")
	}
	for _, action := range []string{"run", "switch", "select", "stage", "refuse"} {
		if !seen[action] {
			t.Fatalf("missing observed %s", action)
		}
	}
	if events[len(events)-1].Reads < 2 || events[len(events)-1].Generation < 2 {
		t.Fatal("missing explicitly rerun applied dataset")
	}
	spec, e := os.ReadFile("model/scoped_view_261.qnt")
	if e != nil {
		t.Fatal(e)
	}
	dir := t.TempDir()
	if e = os.WriteFile(filepath.Join(dir, "scoped_view_261.qnt"), spec, 0600); e != nil {
		t.Fatal(e)
	}
	binding := `spec: scoped_view_261.qnt
module: scoped_view_261
constants: {LIMIT: '8'}
invariants: [safety]
actions:
  applyRun:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  switch:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  selectRow:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  stage:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
  refuse:
    expect: 'reads == {{.obs.reads}} and generation == {{.obs.generation}}'
`
	bindingPath := filepath.Join(dir, "binding.yaml")
	if e = os.WriteFile(bindingPath, []byte(binding), 0600); e != nil {
		t.Fatal(e)
	}
	check := func(trace []scopedViewEvent) bool {
		t.Helper()
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		for _, event := range trace {
			action := event.Action
			if action == "run" {
				action = "applyRun"
			}
			if action == "select" {
				action = "selectRow"
			}
			rec.Record(context.Background(), action, nil, "obs.reads", event.Reads, "obs.generation", event.Generation)
		}
		checker, e := validate.NewChecker(bindingPath, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(events) {
		t.Fatal("observed scope-view effects rejected")
	}
	// One fault at a time, including simultaneous illicit read/scope changes that
	// preserve the simple reads==generation invariant but violate action ownership.
	for _, action := range []string{"switch", "select", "stage", "refuse"} {
		index := -1
		for i, event := range events {
			if event.Action == action {
				index = i
				break
			}
		}
		for _, fault := range []string{"read", "generation", "both"} {
			changed := append([]scopedViewEvent(nil), events...)
			if fault != "generation" {
				changed[index].Reads++
			}
			if fault != "read" {
				changed[index].Generation++
			}
			if check(changed) {
				t.Fatalf("accepted illicit %s effect on %s", fault, action)
			}
		}
	}
}

// Generate actual public-contract operations rather than validating a second Go
// implementation of the same model. Run is host-supplied data replacement; every
// other operation calls the real JS contracts and must preserve the input basis.
func TestHegelScopedViewContractSequences(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(6, 30))
		actions := make([]string, count)
		names := []string{"switch", "select", "stage", "refuse", "run"}
		for i := range actions {
			actions[i] = names[hegel.Draw(tc, hegel.Integers(0, len(names)-1))]
		}
		wire, e := json.Marshal(actions)
		if e != nil {
			t.Fatal(e)
		}
		script := `const fs=require('node:fs'),assert=require('node:assert/strict'),v=require('../ui/view-model.mjs');
const actions=JSON.parse(fs.readFileSync(0,'utf8'));let generation=1;
const make=()=>({version:'goo.dataset/1',id:'ranked',revision:'r'+generation,scope:{id:'q'+generation,generation,signal:'spans',selector:{}},shape:'ranked',fields:[{key:'category',label:'Service',type:'string'},{key:'count',label:'Count',type:'number'}],rows:[{key:'key-'+generation,values:{category:'api',count:'9007199254740993'}}],quality:{extent:'unknown',basis:'captured-query',value:'recorded',sampling:'unknown',freshness:'fixed'}});
let snapshot=make();for(const action of actions){const before=JSON.stringify(snapshot),previous=snapshot;if(action==='run'){const old=v.selectionIntent(snapshot,snapshot.rows[0].key);generation++;snapshot=make();assert.notEqual(old.revision,snapshot.revision);assert.notEqual(old.scopeId,snapshot.scope.id);assert.notEqual(old.generation,snapshot.scope.generation);assert.throws(()=>v.selectionIntent(snapshot,old.key));continue}if(action==='switch'){for(const kind of v.eligibleKinds(snapshot))assert.equal(v.validateDisplay(snapshot,{kind,category:'category',value:'count'}).ok,true)}else if(action==='refuse'){assert.equal(v.validateDisplay(snapshot,{kind:'bar',category:'category',value:'count',requiresComplete:true}).ok,false)}else{const intent=action==='select'?v.selectionIntent(snapshot,snapshot.rows[0].key):v.scopeIntent(snapshot,'category','api');assert.equal(intent.datasetId,snapshot.id);assert.equal(intent.revision,snapshot.revision);assert.equal(intent.scopeId,snapshot.scope.id);assert.equal(intent.generation,generation)}assert.equal(snapshot,previous);assert.equal(JSON.stringify(snapshot),before)}process.stdout.write('ok');`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(wire)
		out, e := cmd.CombinedOutput()
		if e != nil || string(out) != "ok" {
			t.Fatalf("public scoped contracts: %v: %s", e, out)
		}
	}, options(30)...)
}

// Vary the window, interval width and clipped final bucket against the actual JS
// interval contract. A known aggregate interval is a window, never record refs.
func TestHegelScopedTimeIntervalContracts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input := struct{ From, Count, Bucket, Clip, Missing int }{
			From:   hegel.Draw(tc, hegel.Integers(0, 2000000)),
			Count:  hegel.Draw(tc, hegel.Integers(1, 24)),
			Bucket: hegel.Draw(tc, hegel.SampledFrom([]int{1000, 10000, 60000, 300000})),
		}
		input.Clip = hegel.Draw(tc, hegel.Integers(0, input.Bucket-1))
		input.Missing = hegel.Draw(tc, hegel.Integers(0, input.Count-1))
		wire, e := json.Marshal(input)
		if e != nil {
			t.Fatal(e)
		}
		script := `const fs=require('node:fs'),assert=require('node:assert/strict'),v=require('../ui/view-model.mjs');
const input=JSON.parse(fs.readFileSync(0,'utf8')),from=BigInt(input.From),bucket=BigInt(input.Bucket),to=from+BigInt(input.Count)*bucket-BigInt(input.Clip);
const s={version:'goo.dataset/1',id:'intervals',revision:'r1',scope:{id:'q1',generation:1,signal:'logs',selector:{}},shape:'time-series',fields:[{key:'fromMs',label:'From UTC',type:'timestamp',unit:'ms'},{key:'toMs',label:'To UTC',type:'timestamp',unit:'ms'},{key:'count',label:'Captured records',type:'number',unit:'records',nullable:true}],interval:{fromMs:String(from),toMs:String(to),bucketMs:String(bucket)},rows:[],quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'r1',scopeId:'q1',aspects:['intervals']}};
for(let i=0;i<input.Count;i++){const start=from+BigInt(i)*bucket,end=start+bucket<to?start+bucket:to;s.rows.push({key:'interval-'+i,values:{fromMs:String(start),toMs:String(end),count:i%2?'0':'9007199254740993'},state:i%2?'no_data':'observed'})}
s.origin='source-read';s.scope.selector={signal:'logs',environment:'test',fromMs:Number(from),toMs:Number(to),bucketMs:Number(bucket),limit:100,offset:0};const admitted=require('../ui/snapshots.mjs').createSourceAdapter({id:'hegel-time'}).adoptSnapshot(s,{question:s.scope.selector});assert.equal(v.validateSnapshot(s).ok,true);assert.equal(v.validateDisplay(admitted,{kind:'time-bar',requiresComplete:true}).ok,true);assert.deepEqual(v.eligibleKinds(s),['table','time-bar']);
const original=JSON.stringify(s);for(const row of s.rows){const intent=v.intervalIntent(s,row.key);assert.deepEqual(intent,{type:'scope-window',datasetId:s.id,revision:s.revision,scopeId:s.scope.id,generation:s.scope.generation,key:row.key,fromMs:row.values.fromMs,toMs:row.values.toMs});assert.equal(Object.hasOwn(intent,'members'),false)}assert.equal(JSON.stringify(s),original);
const missing=structuredClone(s),removed=missing.rows.splice(input.Missing,1)[0];assert.equal(v.validateSnapshot(missing).ok,false);missing.quality.extent='partial';delete missing.completeness;assert.equal(v.validateSnapshot(missing).ok,true);assert.throws(()=>v.intervalIntent(missing,removed.key));assert.equal(v.validateDisplay(missing,{kind:'time-bar',requiresComplete:true}).ok,false);
const unknown=structuredClone(s),row=unknown.rows[input.Missing];row.state='unknown';row.values.count=null;assert.equal(v.validateSnapshot(unknown).ok,false);unknown.quality.extent='unknown';delete unknown.completeness;assert.equal(v.validateSnapshot(unknown).ok,true);assert.throws(()=>v.intervalIntent(unknown,row.key));row.values.count='0';assert.equal(v.validateSnapshot(unknown).ok,false);
const stale=structuredClone(s);stale.revision='r2';stale.scope.generation=2;stale.completeness.revision='r2';stale.rows=stale.rows.map(row=>({...row,key:'next-'+row.key}));assert.throws(()=>v.intervalIntent(stale,s.rows[0].key));assert.equal(v.intervalIntent(stale,stale.rows[0].key).generation,2);process.stdout.write('ok');`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(wire)
		out, e := cmd.CombinedOutput()
		if e != nil || string(out) != "ok" {
			t.Fatalf("public time interval contracts: %v: %s", e, out)
		}
	}, options(30)...)
}

// Compute the expected signed decimal mean with Go's exact rational arithmetic,
// then exercise the actual JS canonical contract and projections. Coverage and
// support remain separate from whether any valid numeric values were captured.
func TestHegelScopedRationalMeasureContracts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		valid := hegel.Draw(tc, hegel.Integers(0, 20))
		missing := hegel.Draw(tc, hegel.Integers(0, 5))
		empty := hegel.Draw(tc, hegel.Integers(0, 5))
		invalid := hegel.Draw(tc, hegel.Integers(0, 5))
		sum := hegel.Draw(tc, hegel.Integers(-1000000, 1000000))
		scale := hegel.Draw(tc, hegel.Integers(0, 9))
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)
		var mean, sumExact *string
		if valid > 0 {
			meanValue := new(big.Rat).SetFrac(big.NewInt(int64(sum)), new(big.Int).Mul(big.NewInt(int64(valid)), divisor)).RatString()
			sumValue := new(big.Rat).SetFrac(big.NewInt(int64(sum)), divisor).RatString()
			mean, sumExact = &meanValue, &sumValue
		}
		input := struct {
			Valid, Missing, Empty, Invalid, Total string
			Mean, Sum                             *string
			Unit                                  string
			Scale                                 int
		}{Valid: strconv.Itoa(valid), Missing: strconv.Itoa(missing), Empty: strconv.Itoa(empty), Invalid: strconv.Itoa(invalid), Total: strconv.Itoa(valid + missing + empty + invalid), Mean: mean, Sum: sumExact, Scale: scale, Unit: hegel.Draw(tc, hegel.SampledFrom([]string{"1", "bytes", "{requests}"}))}
		wire, e := json.Marshal(input)
		if e != nil {
			t.Fatal(e)
		}
		script := `const fs=require('node:fs'),assert=require('node:assert/strict'),v=require('../ui/view-model.mjs');
const input=JSON.parse(fs.readFileSync(0,'utf8')),known=input.Valid!=='0',state=known?'derived':input.Total==='0'?'no_data':'no_valid_values',support={total:input.Total,valid:input.Valid,missing:input.Missing,empty:input.Empty,invalid:input.Invalid};
const s={version:'goo.dataset/1',id:'measures',revision:'r1',scope:{id:'q1',generation:1,signal:'spans',selector:{}},shape:'ranked',dimensions:['service'],fields:[{key:'service',label:'Service',type:'string'},{key:'count',label:'Captured count',type:'number',unit:'records'},{key:'numeric.mean',label:'Mean',type:'rational',nullable:true,unit:input.Unit,aggregation:{family:'numeric',kind:'mean',method:'exact-rational'}},{key:'numeric.sum',label:'Sum',type:'rational',nullable:true,unit:input.Unit,aggregation:{family:'numeric',kind:'sum',method:'exact'}},{key:'duration.p95',label:'p95',type:'rational',nullable:true,unit:'ns',aggregation:{family:'duration',kind:'quantile',method:'quantilesExactHigh',quantile:'0.95'}}],rows:[{key:'api',values:{service:'api',count:input.Total,'numeric.mean':input.Mean,'numeric.sum':input.Sum,'duration.p95':input.Total==='0'?null:'9007199254740993'},cells:{'numeric.mean':{state,support},'numeric.sum':{state,support},'duration.p95':{state:input.Total==='0'?'no_data':'derived',support:{total:input.Total,valid:input.Total,missing:'0',empty:'0',invalid:'0'}}}}],quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:'r1',scopeId:'q1',aspects:['groups']}};
s.origin='source-read';s.scope.selector={signal:'spans',environment:'test',fromMs:1000,toMs:2000,limit:100,offset:0,groupBy:[{kind:'column',key:'service'}],summary:'span-duration',numeric:{field:{kind:'attribute',key:'measure'},type:input.Scale?'decimal64':'int64',scale:input.Scale,unit:input.Unit,parse:'canonical-decimal'}};const admitted=require('../ui/snapshots.mjs').createSourceAdapter({id:'hegel-measures'}).adoptSnapshot(s,{question:s.scope.selector});assert.equal(v.validateSnapshot(s).ok,true);const original=JSON.stringify(s),mean=v.measureValue(s,s.rows[0],'numeric.mean'),sum=v.measureValue(s,s.rows[0],'numeric.sum'),p95=v.measureValue(s,s.rows[0],'duration.p95');assert.equal(mean.exact,input.Mean);assert.equal(sum.exact,input.Sum);assert.equal(mean.state,state);assert.deepEqual(mean.support,support);assert.equal(mean.unit,input.Unit);assert.equal(p95.aggregation.method,'quantilesExactHigh');assert.equal(p95.aggregation.quantile,'0.95');assert.equal(p95.state,input.Total==='0'?'no_data':'derived');assert.equal(v.validateDisplay(admitted,{kind:'bar',category:'service',value:'numeric.mean',requiresComplete:true}).ok,true);assert.equal(JSON.stringify(s),original);
if(!known){assert.equal(mean.coordinate,null);assert.equal(sum.coordinate,null);const fakeZero=structuredClone(s);fakeZero.rows[0].values['numeric.sum']='0';assert.equal(v.validateSnapshot(fakeZero).ok,false)}else{assert.equal(typeof mean.coordinate,'number');assert.equal(typeof sum.coordinate,'number')}
const unitChanged=structuredClone(s);unitChanged.fields.find(f=>f.key==='numeric.mean').unit='other-unit';unitChanged.fields.push({key:'unmapped',label:'Unmapped diagnostic',type:'string'});unitChanged.rows[0].values.unmapped='outside mapped measure';assert.equal(v.validateSnapshot(unitChanged).ok,true);const after=v.measureValue(unitChanged,unitChanged.rows[0],'numeric.mean');assert.equal(after.exact,mean.exact);assert.equal(after.coordinate,mean.coordinate);assert.deepEqual(after.support,mean.support);assert.equal(after.unit,'other-unit');assert.throws(()=>v.scopeIntent(s,'numeric.mean',input.Mean));
const changed=structuredClone(s);changed.rows[0].cells['numeric.mean'].support.total=String(BigInt(input.Total)+1n);assert.equal(v.validateSnapshot(changed).ok,false);const badMethod=structuredClone(s);badMethod.fields.find(f=>f.key==='duration.p95').aggregation.method='exact-rational';assert.equal(v.validateSnapshot(badMethod).ok,false);process.stdout.write('ok');`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(wire)
		out, e := cmd.CombinedOutput()
		if e != nil || string(out) != "ok" {
			t.Fatalf("public rational measure contracts: %v: %s", e, out)
		}
	}, options(30)...)
}

// Exact population differences retain two independent bases. This checks the
// browser contract with Go rational oracles, not rounded percentages or ratios.
func TestHegelScopedSummaryComparisonContracts(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		type side struct{ Count, Valid, Missing, Sum, Mean string }
		makeSide := func(allowEmpty bool) side {
			low := 1
			if allowEmpty {
				low = 0
			}
			valid := hegel.Draw(tc, hegel.Integers(low, 15))
			missing := hegel.Draw(tc, hegel.Integers(0, 5))
			sum := hegel.Draw(tc, hegel.Integers(-100000, 100000))
			scale := hegel.Draw(tc, hegel.Integers(0, 6))
			s := side{Count: strconv.Itoa(valid + missing), Valid: strconv.Itoa(valid), Missing: strconv.Itoa(missing), Sum: strconv.Itoa(sum)}
			if valid > 0 {
				denominator := new(big.Int).Mul(big.NewInt(int64(valid)), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil))
				s.Mean = new(big.Rat).SetFrac(big.NewInt(int64(sum)), denominator).RatString()
			}
			return s
		}
		a, b := makeSide(false), makeSide(true)
		countA, _ := new(big.Int).SetString(a.Count, 10)
		countB, _ := new(big.Int).SetString(b.Count, 10)
		deltaCount := new(big.Int).Sub(countB, countA).String()
		deltaMean := ""
		if a.Mean != "" && b.Mean != "" {
			av, _ := new(big.Rat).SetString(a.Mean)
			bv, _ := new(big.Rat).SetString(b.Mean)
			deltaMean = new(big.Rat).Sub(bv, av).RatString()
		}
		input := struct {
			A, B                  side
			DeltaCount, DeltaMean string
		}{a, b, deltaCount, deltaMean}
		wire, e := json.Marshal(input)
		if e != nil {
			t.Fatal(e)
		}
		script := `const fs=require('node:fs'),assert=require('node:assert/strict'),v=require('../ui/view-model.mjs'),{validateComparison}=require('../ui/view-comparison.mjs');const input=JSON.parse(fs.readFileSync(0,'utf8'));
const make=(data,role)=>({version:'goo.dataset/1',id:role,revision:role+'-content',origin:'browser-held',sourceBasis:'caller-claimed-basis',scope:{id:role+'-scope',generation:role+'-content',signal:'spans',selector:{signal:'spans',fromMs:role==='baseline'?1000:2000,toMs:role==='baseline'?2000:3000,numeric:{field:{kind:'attribute',key:'payload.size'},type:'decimal64',scale:2,unit:'bytes',parse:'canonical-decimal'}}},shape:'summary',fields:[{key:'count',label:'Captured records',type:'number',unit:'records'},{key:'numeric.mean',label:'Mean',type:'rational',nullable:true,unit:'bytes',aggregation:{family:'numeric',kind:'mean',method:'exact-rational'}}],rows:[{key:'population',values:{count:data.Count,'numeric.mean':data.Mean||null},cells:{'numeric.mean':{state:data.Mean?'derived':data.Count==='0'?'no_data':'no_valid_values',support:{total:data.Count,valid:data.Valid,missing:data.Missing,empty:'0',invalid:'0'}}}}],quality:{extent:'unknown',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'}});
const c={version:'goo.comparison/1',state:input.B.Count==='0'?'no_data':'observed',baseline:make(input.A,'baseline'),selection:make(input.B,'selection'),changes:{count:{state:'derived',value:input.DeltaCount,unit:'records'},'numeric.mean':input.DeltaMean?{state:'derived',value:input.DeltaMean,unit:'bytes'}:{state:'unavailable',value:null,unit:'bytes',reason:'Selection has no valid numeric observations'}}};
assert.equal(validateComparison(c).ok,true);const original=JSON.stringify(c);assert.equal(c.baseline.origin,'browser-held');assert.equal(c.baseline.quality.extent,'unknown');assert.notEqual(c.baseline.revision,c.selection.revision);assert.notEqual(c.baseline.scope.id,c.selection.scope.id);assert.equal(c.baseline.sourceBasis,c.selection.sourceBasis);assert.equal(v.capabilities(c.baseline).selectRows,false);assert.throws(()=>v.selectionIntent(c.baseline,'population'));assert.equal(JSON.stringify(c),original);
const forged=structuredClone(c);forged.changes.count.value=String(BigInt(input.DeltaCount)+1n);assert.equal(validateComparison(forged).ok,false);const binding=structuredClone(c);binding.selection.scope.selector.numeric.field.key='other.attribute';binding.changes['numeric.mean']={state:'unavailable',value:null,unit:'bytes',reason:'Different numeric binding'};assert.equal(validateComparison(binding).ok,true);binding.changes['numeric.mean']={state:'derived',value:input.DeltaMean||'0',unit:'bytes'};assert.equal(validateComparison(binding).ok,false);
const diagnostic=structuredClone(c);diagnostic.state='diagnostic';assert.equal(validateComparison(diagnostic).ok,false);for(const [key,value]of Object.entries(diagnostic.changes))diagnostic.changes[key]={state:'unavailable',value:null,unit:value.unit,reason:'Source reads disagree'};assert.equal(validateComparison(diagnostic).ok,true);const complete=structuredClone(c.baseline);complete.quality.extent='complete';complete.completeness={basis:'captured-query',revision:complete.revision,scopeId:complete.scope.id,aspects:['summary']};assert.equal(v.validateSnapshot(complete).ok,false);process.stdout.write('ok');`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(wire)
		out, e := cmd.CombinedOutput()
		if e != nil || string(out) != "ok" {
			t.Fatalf("public population comparison contracts: %v: %s", e, out)
		}
	}, options(30)...)
}

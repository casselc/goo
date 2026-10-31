package verification

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

func nativeHistogramTimeFixture(seed, mode int) (*measurements.Result, error) {
	width := int64(4000)
	q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "native-time-contract", Title: "Native time bucket", Kind: "histogram", Environment: "prod", LookbackMinutes: 2, MaxGapMS: 3000, HistogramBucketMS: &width, Grain: "per-stream", Source: measurements.Binding{Services: []string{"worker"}, Scope: "native.time.contract", Version: "1", Metric: "worker.duration", Unit: "s"}}, FromMS: 100000, ToMS: 110000}
	shift := int64(0)
	if seed%2 == 1 {
		shift = 1
		q.FromNS = "100000000001"
		q.ToNS = "110000000001"
		q.ToMS = 110001
	}
	bounds := [][]float64{{.1, .5}, {0, math.SmallestNonzeroFloat64, .1, math.MaxFloat64}, {}, {1}, {.2, .8}}[seed%5]
	counts := make([]uint64, len(bounds)+1)
	for i := range counts {
		counts[i] = uint64((seed + i) % 5)
	}
	if mode == 1 {
		for i := range counts {
			counts[i] = 0
		}
	}
	if mode == 2 {
		for i := range counts {
			counts[i] = 0
		}
		counts[len(counts)-1] = math.MaxUint64 / 2
	}
	rows := []json.RawMessage{}
	cumulative := make([]uint64, len(counts))
	for i := 0; i < 5; i++ {
		values := make([]string, len(counts))
		var total uint64
		for j, n := range counts {
			if i == 2 || i == 3 {
				n = 0
			}
			if mode == 3 {
				cumulative[j] += n
				n = cumulative[j]
			}
			values[j] = fmt.Sprint(n)
			total += n
		}
		start := 100000000000 + int64(i)*2000000000 + shift
		temp := 1
		if mode == 3 {
			start = 98000000000 + shift
			temp = 2
		}
		row, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": "worker", "service.instance.id": "one", "deployment.environment.name": "prod"}, "resource_schema_url": "", "scope": "native.time.contract", "version": "1", "schema_url": "", "name": "worker.duration", "unit": "s", "attributes": map[string]string{}, "start_ns": fmt.Sprint(start), "end_ns": fmt.Sprint(102000000000 + int64(i)*2000000000 + shift), "temporality": temp, "flags": 0, "bounds": bounds, "bucket_counts": values, "count": fmt.Sprint(total)})
		rows = append(rows, row)
	}
	if mode == 3 { // An initial zero cumulative observation establishes the epoch boundary.
		var row map[string]any
		json.Unmarshal(rows[0], &row)
		zeros := make([]string, len(counts))
		for i := range zeros {
			zeros[i] = "0"
		}
		row["start_ns"] = fmt.Sprint(98000000000 + shift)
		row["end_ns"] = fmt.Sprint(100000000000 + shift)
		row["bucket_counts"] = zeros
		row["count"] = "0"
		first, _ := json.Marshal(row)
		rows = append([]json.RawMessage{first}, rows...)
	}
	return measurements.Evaluate(q, rows, time.Unix(200, 123))
}

const nativeHistogramTimeParityScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {buildNativeHistogramTimeView} from '../ui/native-histogram-view.mjs';
import {validateSnapshot,histogramIntent} from '../ui/view-model.mjs';
import {importSnapshot,hasGuarantee} from '../ui/snapshots.mjs';
const {result,expected,index,mutation=-1}=JSON.parse(fs.readFileSync(0,'utf8'));
const value=buildNativeHistogramTimeView(result,0,index);assert.deepEqual(JSON.parse(JSON.stringify(value)),expected);assert(validateSnapshot(value).ok);assert.equal(value.scope.selector.kind,'measurement-time-bucket');assert.equal(value.scope.selector.widthNS,'4000000000');assert.equal(value.scope.selector.bucketIndex,index);assert.equal(value.sourceBasis,result.histogramTime.basis);assert.equal(value.histogram.total,result.histogramTime.series[0].buckets[index].distribution.count);assert.equal(value.rows[0].values.from,'-inf');assert.equal(value.rows.at(-1).values.to,'+inf');assert(!value.histogram.support&&!value.histogram.binding);assert.throws(()=>histogramIntent(value,'bin:0'));assert(!hasGuarantee(importSnapshot(value),'bins'));
if(mutation>=0){const bad=structuredClone(result);switch(mutation){case 0:bad.histogramTime.sourceBasis='f'.repeat(64);break;case 1:bad.histogramTime.widthNS='5000000000';break;case 2:bad.histogramTime.series[0].buckets[index].fromNS=String(BigInt(bad.histogramTime.series[0].buckets[index].fromNS)+1n);break;case 3:bad.histogramTime.series[0].buckets[index].state='unavailable';break;case 4:bad.histogramTime.series[0].buckets[index].reasons=['boundary-crossing'];break;case 5:bad.histogramTime.series[0].buckets[index].distribution.count=String(BigInt(bad.histogramTime.series[0].buckets[index].distribution.count)+1n);break;case 6:bad.histogramTime.series[0].stream.resource['service.name']='foreign';break;case 7:bad.histogramTime.series[0].temporality='other';break;case 8:bad.histogramTime.series[0].buckets[index].toNS=String(BigInt(bad.histogramTime.toNS)+1n);break;case 9:bad.histogramTime.basis='';break;case 10:bad.histogramTime.series[0].buckets[index].reasons=null;break}assert.equal(buildNativeHistogramTimeView(bad,0,index),null,'malformed evaluated bucket refused');}
`

func nativeHistogramTimeCase(t interface{ Fatal(...any) }, seed, mode, index, mutation int) {
	r, e := nativeHistogramTimeFixture(seed, mode)
	if e != nil {
		t.Fatal(e)
		return
	}
	view, e := measurements.HistogramTimeView(r, 0, index)
	if e != nil {
		t.Fatal(e)
		return
	}
	if mutation >= 0 {
		raw, _ := json.Marshal(r)
		var bad measurements.Result
		json.Unmarshal(raw, &bad)
		h := bad.HistogramTime
		b := &h.Series[0].Buckets[index]
		switch mutation {
		case 0:
			h.SourceBasis = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		case 1:
			h.WidthNS = "5000000000"
		case 2:
			var n uint64
			fmt.Sscan(b.FromNS, &n)
			b.FromNS = fmt.Sprint(n + 1)
		case 3:
			b.State = "unavailable"
		case 4:
			b.Reasons = []string{"boundary-crossing"}
		case 5:
			var n uint64
			fmt.Sscan(b.Distribution.Count, &n)
			b.Distribution.Count = fmt.Sprint(n + 1)
		case 6:
			h.Series[0].Stream.Resource["service.name"] = "foreign"
		case 7:
			h.Series[0].Temporality = "other"
		case 8:
			var n uint64
			fmt.Sscan(h.ToNS, &n)
			b.ToNS = fmt.Sprint(n + 1)
		case 9:
			h.Basis = ""
		case 10:
			b.Reasons = nil
		}
		if _, err := measurements.HistogramTimeView(&bad, 0, index); err == nil {
			t.Fatal("Go accepted malformed native time projection", mutation)
			return
		}
	}
	if e = runNativeHistogramJS(map[string]any{"result": r, "expected": view, "index": index, "mutation": mutation}, nativeHistogramTimeParityScript); e != nil {
		t.Fatal(e)
	}
}
func TestHegelCanonicalNativeHistogramTimeParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		nativeHistogramTimeCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), hegel.Draw(tc, hegel.Integers(0, 3)), hegel.Draw(tc, hegel.Integers(0, 2)), -1)
	}, options(70)...)
}
func TestHegelCanonicalNativeHistogramTimeMalformed(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		nativeHistogramTimeCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), 0, hegel.Draw(tc, hegel.Integers(0, 2)), hegel.Draw(tc, hegel.Integers(0, 10)))
	}, options(30)...)
}

func TestCanonicalNativeHistogramTimeEvaluatedRefusals(t *testing.T) {
	for mode := 0; mode < 6; mode++ {
		fixtureMode := 0
		if mode == 5 {
			fixtureMode = 3
		}
		base, err := nativeHistogramTimeFixture(0, fixtureMode)
		if err != nil {
			t.Fatal(err)
		}
		rows := append([]json.RawMessage(nil), base.Rows...)
		q := base.Question
		target := 1
		if mode == 0 {
			target = 0
		}
		if mode == 5 {
			target = 2
		}
		var row map[string]any
		if err = json.Unmarshal(rows[target], &row); err != nil {
			t.Fatal(err)
		}
		switch mode {
		case 0:
			row["flags"] = 1
		case 1:
			width := int64(3000)
			q.Definition.HistogramBucketMS = &width
		case 2:
			row["bounds"] = []float64{.25, .75}
		case 3:
			rows = append(rows[:1], rows[2:]...)
		case 4:
			q.FromMS = 100001
		case 5:
			row["count"] = "0"
			row["bucket_counts"] = []string{"0", "0", "0"}
		}
		if mode != 3 {
			rows[target], err = json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
		}
		r, err := measurements.Evaluate(q, rows, time.Unix(200, 123))
		if err != nil {
			t.Fatal(err)
		}
		if r.HistogramTime.Series[0].Buckets[0].State != "unavailable" {
			t.Fatalf("mode%d did not exercise an unavailable time bucket: %+v", mode, r.HistogramTime.Series[0].Buckets[0])
		}
		if _, err = measurements.HistogramTimeView(r, 0, 0); err == nil {
			t.Fatalf("mode%d projected unavailable bucket", mode)
		}
		script := `import assert from 'node:assert/strict';import fs from 'node:fs';import {buildNativeHistogramTimeView} from '../ui/native-histogram-view.mjs';const {result}=JSON.parse(fs.readFileSync(0,'utf8'));assert.equal(buildNativeHistogramTimeView(result,0,0),null);`
		if err = runNativeHistogramJS(map[string]any{"result": r}, script); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCanonicalNativeHistogramTimeNullReasons(t *testing.T) {
	nativeHistogramTimeCase(t, 1, 0, 0, 10)
}

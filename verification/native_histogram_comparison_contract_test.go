package verification

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

// Independent retained populations are evaluated by the real native kernels.
// Whole windows cover multiple exports; no inspection-row count is a denominator.
func nativeHistogramPairFixture(seed, mode int) (*measurements.Result, *measurements.Result, metrics.WindowComparison, error) {
	var results [2]*measurements.Result
	bounds := [][]float64{{.1, .5}, {0, math.SmallestNonzeroFloat64, .1, math.MaxFloat64}, {}, {1}, {.2, .8}}[seed%5]
	unit := "s"
	if seed%2 == 1 {
		unit = "ms"
	}
	for side := 0; side < 2; side++ {
		exports := 2
		if mode == 3 && side == 1 {
			exports = 3
		}
		from := int64(100000 + side*20000)
		q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "native-pair-contract", Title: "Native histogram comparison", Kind: "histogram", Environment: "prod", LookbackMinutes: 2, MaxGapMS: 3000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"worker"}, Scope: "native.pair.contract", Version: "1", Metric: "worker.duration", Unit: unit}}, FromMS: from, ToMS: from + int64(exports)*2000}
		shift := int64(0)
		if mode == 3 {
			shift = 1
			q.FromNS = fmt.Sprint(from*1000000 + shift)
			q.ToNS = fmt.Sprint(q.ToMS*1000000 + shift)
			q.ToMS++
		}
		counts := make([]uint64, len(bounds)+1)
		for j := range counts {
			counts[j] = uint64((seed+j)%5+1) * uint64(side+1)
		}
		if mode == 2 {
			for j := range counts {
				counts[j] = 0
			}
			counts[len(counts)-1] = math.MaxUint64 / 8 * uint64(side+1)
		}
		rows := []json.RawMessage{}
		cum := make([]uint64, len(counts))
		epoch := from*1000000 + shift - 2000000000
		row := func(start, end int64, values []uint64, temp int) json.RawMessage {
			b := make([]string, len(values))
			var total uint64
			for j, n := range values {
				b[j] = fmt.Sprint(n)
				total += n
			}
			raw, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": "worker", "service.instance.id": "one", "deployment.environment.name": "prod"}, "resource_schema_url": "", "scope": q.Definition.Source.Scope, "version": "1", "schema_url": "", "name": "worker.duration", "unit": unit, "attributes": map[string]string{"queue": "jobs"}, "start_ns": fmt.Sprint(start), "end_ns": fmt.Sprint(end), "temporality": temp, "flags": 0, "bounds": bounds, "bucket_counts": b, "count": fmt.Sprint(total)})
			return raw
		}
		if mode == 1 {
			rows = append(rows, row(epoch, from*1000000+shift, make([]uint64, len(counts)), 2))
		}
		for i := 0; i < exports; i++ {
			start := from*1000000 + shift + int64(i)*2000000000
			end := start + 2000000000
			values := counts
			temp := 1
			if mode == 1 {
				temp = 2
				start = epoch
				for j, n := range counts {
					cum[j] += n
				}
				values = cum
			}
			rows = append(rows, row(start, end, values, temp))
		}
		var err error
		results[side], err = measurements.Evaluate(q, rows, time.Unix(int64(200+side), 123))
		if err != nil {
			return nil, nil, metrics.WindowComparison{}, err
		}
	}
	windows := metrics.CompareHistogramWindows(results[0].Histograms, results[1].Histograms)
	if len(windows) != 1 {
		return nil, nil, metrics.WindowComparison{}, fmt.Errorf("want one evaluated window pair, got %d", len(windows))
	}
	return results[0], results[1], windows[0], nil
}

const nativeHistogramPairParityScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {buildNativeHistogramComparison,buildNativeHistogramComparisonFromEvidence} from '../ui/native-histogram-view.mjs';
import {validateHistogramComparison} from '../ui/view-histogram-comparison.mjs';
import {model} from '../ui/aligned-histogram.mjs';
import {asHistogramComparison,hasGuarantee} from '../ui/snapshots.mjs';
const {baseline,selection,row,expected,evidence}=JSON.parse(fs.readFileSync(0,'utf8'));
const value=buildNativeHistogramComparison(baseline,selection,row);assert.deepEqual(JSON.parse(JSON.stringify(value)),expected);
if(value){assert.deepEqual(buildNativeHistogramComparisonFromEvidence(evidence,evidence.nativeWindows[0]),expected);assert(validateHistogramComparison(value).ok);assert.equal(value.baseline.scope.selector.kind,'measurement-window');assert.equal(value.selection.scope.selector.kind,'measurement-window');assert.equal(value.baseline.scope.selector.fromNS,row.baseline.fromNS);assert.equal(value.selection.scope.selector.toNS,row.selection.toNS);for(const side of ['baseline','selection']){assert.equal(value[side].origin,'browser-held');assert.equal(value[side].quality.extent,'unknown');assert(!value[side].completeness);assert(!value[side].histogram.support);assert.equal(value[side].histogram.total,row[side].distribution.count)}const held=asHistogramComparison(value);assert(!hasGuarantee(held.baseline,'bins'));assert(!hasGuarantee(held.selection,'bins'));const counts=model(held,{scale:'count'}),share=model(held,{scale:'observation-share'});assert(counts&&share);assert.equal(share.ceiling,'100');assert.equal(model(held,{scale:'valid-share'}),null);assert.equal(model(held,{scale:'population-share'}),null);for(let i=0;i<counts.all.length;i++){assert.deepEqual(counts.all[i].counts,[row.baseline.distribution.bucketCounts[i],row.selection.distribution.bucketCounts[i]])}}
`

func nativeHistogramPairCase(t interface{ Fatal(...any) }, seed, mode, mutation int) {
	a, b, w, e := nativeHistogramPairFixture(seed, mode)
	if e != nil {
		t.Fatal(e)
		return
	}
	capture := func(r *measurements.Result, title string) *exploration.SavedCapture {
		return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, ID: r.Basis, Title: title, CapturedAt: r.ReadAt, Question: exploration.CaptureQuestion{Kind: "measurement", Measurement: &r.Question}.Normalized()}, Measurement: r}
	}
	evidence, comparisonErr := exploration.CompareEvidence(capture(a, "Baseline"), capture(b, "Comparison"))
	if comparisonErr != nil {
		t.Fatal(comparisonErr)
		return
	}
	if mutation >= 0 {
		switch mutation {
		case 0:
			w.State = "unavailable"
			w.Reasons = []string{"another-member-withholds-comparison"}
		case 1:
			w.Selection.State = "unavailable"
			w.Selection.Distribution = nil
		case 2:
			w.Selection.Distribution.Unit = "bytes"
		case 3:
			w.Selection.Temporality = "other"
		case 4:
			w.Stream.Resource["service.instance.id"] = "foreign"
		case 5:
			w.Selection.FromNS = fmt.Sprint(b.Question.FromMS*1000000 + 1)
		case 6:
			w.Selection.Distribution.Count = "0"
		case 7:
			w.Baseline.Reasons = []string{"incomplete-window"}
		case 8:
			b.Histograms.Basis = ""
		case 9:
			w.Selection.Distribution.Bounds = []float64{123, 456}
		}
	}
	pair, err := exploration.NativeHistogramComparison(a, b, w)
	if mutation >= 0 {
		if pair != nil || err == nil {
			t.Fatal("Go accepted withheld or malformed native pair", mutation)
			return
		}
	} else if err != nil || pair == nil {
		t.Fatal("valid actual native pair refused", err)
		return
	}
	if e = runNativeHistogramJS(map[string]any{"baseline": a, "selection": b, "row": w, "expected": pair, "evidence": evidence}, nativeHistogramPairParityScript); e != nil {
		t.Fatal(e)
	}
}
func TestHegelCanonicalNativeHistogramPairParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		nativeHistogramPairCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), hegel.Draw(tc, hegel.Integers(0, 3)), -1)
	}, options(70)...)
}
func TestHegelCanonicalNativeHistogramPairRefusals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		nativeHistogramPairCase(tc, hegel.Draw(tc, hegel.Integers(0, 1000)), 0, hegel.Draw(tc, hegel.Integers(0, 9)))
	}, options(30)...)
}

func TestCanonicalNativeHistogramPairWholeComparisonWithheld(t *testing.T) {
	a, b, _, err := nativeHistogramPairFixture(1, 0)
	if err != nil {
		t.Fatal(err)
	}
	duplicate := func(r *measurements.Result, missing bool) (*measurements.Result, error) {
		rows := append([]json.RawMessage(nil), r.Rows...)
		for i, raw := range r.Rows {
			var row map[string]any
			if e := json.Unmarshal(raw, &row); e != nil {
				return nil, e
			}
			row["resource"].(map[string]any)["service.instance.id"] = "two"
			if missing && i == 0 {
				row["flags"] = 1
			}
			copy, _ := json.Marshal(row)
			rows = append(rows, copy)
		}
		at, e := time.Parse(time.RFC3339Nano, r.ReadAt)
		if e != nil {
			return nil, e
		}
		return measurements.Evaluate(r.Question, rows, at)
	}
	a, err = duplicate(a, false)
	if err != nil {
		t.Fatal(err)
	}
	b, err = duplicate(b, true)
	if err != nil {
		t.Fatal(err)
	}
	capture := func(r *measurements.Result) *exploration.SavedCapture {
		return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, ID: r.Basis, Title: "Native pair", CapturedAt: r.ReadAt, Question: exploration.CaptureQuestion{Kind: "measurement", Measurement: &r.Question}.Normalized()}, Measurement: r}
	}
	evidence, err := exploration.CompareEvidence(capture(a), capture(b))
	if err != nil {
		t.Fatal(err)
	}
	if evidence.State != "unavailable" || len(evidence.NativeWindows) != 2 {
		t.Fatal("want withheld whole comparison", evidence.State)
	}
	for _, row := range evidence.NativeWindows {
		if pair, err := exploration.NativeHistogramComparison(a, b, row); err == nil || pair != nil {
			t.Fatal("withheld window produced a native pair")
		}
	}
	script := `import fs from 'node:fs';import assert from 'node:assert/strict';import {buildNativeHistogramComparisonFromEvidence} from '../ui/native-histogram-view.mjs';const {evidence}=JSON.parse(fs.readFileSync(0,'utf8'));for(const row of evidence.nativeWindows)assert.equal(buildNativeHistogramComparisonFromEvidence(evidence,row),null);`
	if err = runNativeHistogramJS(map[string]any{"evidence": evidence}, script); err != nil {
		t.Fatal(err)
	}
}

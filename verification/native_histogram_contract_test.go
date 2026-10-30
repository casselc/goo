package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

// Real evaluator output, including an unrelated unknown interval in the same
// stream. A selected known interval is not a claim of whole-window coverage.
func nativeHistogramFixture(seed, mode int) (*measurements.Result, int, error) {
	unit := "s"
	if seed%2 == 1 {
		unit = "ms"
	}
	q := measurements.Question{Definition: measurements.Definition{APIVersion: measurements.Version, Name: "native-histogram-contract", Title: "Native interval", Kind: "histogram", Environment: "prod", LookbackMinutes: 2, MaxGapMS: 3000, Grain: "per-stream", Source: measurements.Binding{Services: []string{"worker"}, Scope: "native.contract", Version: "1", SchemaURL: "https://opentelemetry.io/schemas/1.30.0", Metric: "worker.duration", Unit: unit}}, FromMS: 100000, ToMS: 104000}
	bounds := [][]float64{{.1, .5}, {0, math.SmallestNonzeroFloat64, .1, math.MaxFloat64}, {}, {1}, {.2, .8}}[seed%5]
	counts := make([]string, len(bounds)+1)
	total := uint64(0)
	for i := range counts {
		n := uint64((seed + i) % 5)
		if mode == 1 {
			n = 0
		}
		counts[i] = strconv.FormatUint(n, 10)
		total += n
	}
	if mode == 2 {
		for i := range counts {
			counts[i] = "0"
		}
		counts[len(counts)-1] = "18446744073709551615"
		total = math.MaxUint64
	}
	row := func(start, end string, c []string, n string, flags int, temp int) json.RawMessage {
		raw, _ := json.Marshal(map[string]any{"resource": map[string]string{"service.name": "worker", "service.instance.id": "one", "deployment.environment.name": "prod"}, "resource_schema_url": "", "scope": "native.contract", "version": "1", "schema_url": q.Definition.Source.SchemaURL, "name": "worker.duration", "unit": unit, "attributes": map[string]string{"queue": "jobs"}, "start_ns": start, "end_ns": end, "temporality": temp, "flags": flags, "bounds": bounds, "bucket_counts": c, "count": n})
		return raw
	}
	zero := make([]string, len(counts))
	for i := range zero {
		zero[i] = "0"
	}
	rows := []json.RawMessage{row("100000000000", "102000000000", counts, strconv.FormatUint(total, 10), 0, 1), row("102000000000", "104000000000", zero, "0", 1, 1)}
	if mode == 3 {
		rows = []json.RawMessage{row("100000000000", "101000000000", zero, "0", 0, 2), row("100000000000", "102000000000", counts, strconv.FormatUint(total, 10), 0, 2), row("100000000000", "104000000000", zero, "0", 1, 2)}
	}
	result, err := measurements.Evaluate(q, rows, time.Unix(200, 123))
	return result, 0, err
}

func runNativeHistogramJS(payload map[string]any, script string) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("native JS parity: %w: %s", err, out)
	}
	return nil
}

const nativeHistogramParityScript = `
import assert from 'node:assert/strict';import fs from 'node:fs';
import {buildNativeHistogramView} from '../ui/native-histogram-view.mjs';
import {validateSnapshot,histogramIntent} from '../ui/view-model.mjs';
import {importSnapshot,hasGuarantee} from '../ui/snapshots.mjs';
const {result,expected,index,edges,mutation=-1}=JSON.parse(fs.readFileSync(0,'utf8'));
const value=buildNativeHistogramView(result,0,index);assert.deepEqual(JSON.parse(JSON.stringify(value)),expected);
assert(validateSnapshot(value).ok);assert.equal(value.histogram.grain,'observations');assert.equal(value.histogram.total,result.histograms.series[0].intervals[index].distribution.count);assert.notEqual(value.histogram.total,undefined);assert.equal(value.rows[0].values.from,'-inf');assert.equal(value.rows.at(-1).values.to,'+inf');assert.deepEqual(value.rows.slice(0,-1).map(r=>r.values.to),edges);assert(!value.histogram.support&&!value.histogram.binding);assert.throws(()=>histogramIntent(value,'bin:0'));
const held=importSnapshot(value);assert(!hasGuarantee(held,'bins'));assert.equal(held.origin,'browser-held');
assert.equal(buildNativeHistogramView(result,0,result.histograms.series[0].intervals.length-1),null);
if(mutation>=0){const bad=structuredClone(expected);switch(mutation){case 0:bad.histogram.closure='left-closed-right-open';break;case 1:bad.histogram.grain='records';break;case 2:bad.histogram.unit='bytes';break;case 3:bad.rows[0].values.from='0';break;case 4:bad.rows.at(-1).values.to='18446744073709551615';break;case 5:bad.histogram.total=String(BigInt(bad.histogram.total)+1n);break;case 6:bad.histogram.support={total:'0',valid:'0',missing:'0',empty:'0',invalid:'0'};break;case 7:bad.scope.selector.stream.resource['service.name']='foreign';break;case 8:bad.scope.selector.fromNS=bad.scope.selector.toNS;break;case 9:bad.rows[0].values.count=0;break}assert(!validateSnapshot(bad).ok,'malformed native contract refused');}
`

func nativeHistogramCase(t interface{ Fatal(...any) }, seed, mode, mutation int) {
	r, index, err := nativeHistogramFixture(seed, mode)
	if err != nil {
		t.Fatal(err)
		return
	}
	view, err := measurements.HistogramView(r, 0, index)
	if err != nil {
		t.Fatal(err)
		return
	}
	edges := []string{}
	for _, bound := range r.Histograms.Series[0].Intervals[index].Distribution.Bounds {
		edges = append(edges, new(big.Rat).SetFloat64(bound).RatString())
	}
	if err = runNativeHistogramJS(map[string]any{"result": r, "expected": view, "index": index, "edges": edges, "mutation": mutation}, nativeHistogramParityScript); err != nil {
		t.Fatal(err)
	}
}

func TestHegelCanonicalNativeHistogramParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		nativeHistogramCase(tc, seed, mode, -1)
	}, options(70)...)
}
func TestHegelCanonicalNativeHistogramMalformed(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		mutation := hegel.Draw(tc, hegel.Integers(0, 9))
		nativeHistogramCase(tc, seed, 0, mutation)
	}, options(30)...)
}

func TestCanonicalNativeHistogramEvaluatedRefusals(t *testing.T) {
	for mode := 0; mode < 6; mode++ {
		base, _, err := nativeHistogramFixture(4, 3)
		if err != nil {
			t.Fatal(err)
		}
		rows := append([]json.RawMessage(nil), base.Rows...)
		q := base.Question
		index := 1
		var row map[string]any
		if err = json.Unmarshal(rows[2], &row); err != nil {
			t.Fatal(err)
		}
		row["flags"] = 0
		switch mode {
		case 0:
			row["flags"] = 1
		case 1:
			q.Definition.MaxGapMS = 500
			index = 0
		case 2: // A lower cumulative count is an evaluator diagnostic, not a zero bin.
		case 3:
			row["bounds"] = []float64{.25, .75}
		case 4:
			row["start_ns"] = "102000000000"
		case 5:
			q.FromMS = 101001
			index = 0
		}
		rows[2], err = json.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		r, err := measurements.Evaluate(q, rows, time.Unix(200, 123))
		if err != nil {
			t.Fatal(err)
		}
		if r.Histograms.Series[0].Intervals[index].State != "unknown" {
			t.Fatalf("mode%d did not exercise evaluator refusal", mode)
		}
		if _, err = measurements.HistogramView(r, 0, index); err == nil {
			t.Fatalf("mode%d projected unknown as observations", mode)
		}
		script := `import assert from 'node:assert/strict';import fs from 'node:fs';import {buildNativeHistogramView} from '../ui/native-histogram-view.mjs';const {result,index}=JSON.parse(fs.readFileSync(0,'utf8'));assert.equal(buildNativeHistogramView(result,0,index),null);`
		if err = runNativeHistogramJS(map[string]any{"result": r, "index": index}, script); err != nil {
			t.Fatal(err)
		}
	}
}

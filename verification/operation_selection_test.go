package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/annotations"
	"hegel.dev/go/hegel"
)

// Exercise the actual browser selection kernel with generated exact-ns evidence.
func TestHegelOperationHeatmapSelection(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		const epoch = uint64(1792209000000000000)
		count := hegel.Draw(tc, hegel.Integers(0, 40))
		edges := []uint64{0, 999999, 1000000, 9999999, 10000000, 49999999, 50000000, 99999999, 100000000, 499999999, 500000000, 999999999, 1000000000, 9999999999, 10000000000, 18446744073709551615}
		spans := []annotations.OperationSpan{}
		expected := []string{}
		for i := 0; i < count; i++ {
			offset := hegel.Draw(tc, hegel.Integers(0, 3))
			duration := edges[hegel.Draw(tc, hegel.Integers(0, len(edges)-1))]
			operation := []string{"checkout.reserve", "checkout.stock"}[hegel.Draw(tc, hegel.Integers(0, 1))]
			outcome := []string{"ok", "error", "panic"}[hegel.Draw(tc, hegel.Integers(0, 2))]
			start := epoch + []uint64{0, 1, 59999999999, 60000000000}[offset]
			span := annotations.OperationSpan{SpanID: fmt.Sprint(i), TraceID: fmt.Sprint(i % 3), Operation: operation, Role: "server", Outcome: outcome, StartNS: strconv.FormatUint(start, 10), DurationNS: strconv.FormatUint(duration, 10)}
			spans = append(spans, span)
			if operation == "checkout.reserve" && outcome == "error" && start < epoch+60000000000 {
				expected = append(expected, span.SpanID)
			}
		}
		input, _ := json.Marshal(spans)
		command := exec.Command("node", "-e", `const fs=require('fs'),o=require('../ui/operations.js');const spans=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(spans);const filter={operation:'checkout.reserve',role:'server',outcome:'error',fromNS:'1792209000000000000',toNS:'1792209060000000000'};const selected=o.select(spans,filter);const cells=o.distribution(selected,1792209000000,1792209060000);const members=cells.flatMap(c=>o.select(selected,c).map(s=>s.spanID));if(cells.some(c=>o.select(selected,c).length!==c.count)||members.length!==selected.length||new Set(members).size!==selected.length)throw Error('cells do not partition selection');o.select(spans,{traceID:'0'});if(JSON.stringify(spans)!==before)throw Error('mutated evidence');const repeated=o.select(spans,filter);console.log(JSON.stringify(repeated.map(s=>s.spanID)));`)
		command.Stdin = bytes.NewReader(input)
		out, err := command.CombinedOutput()
		if err != nil {
			tc.Fatalf("browser kernel: %s %v", out, err)
		}
		var actual []string
		if err = json.Unmarshal(out, &actual); err != nil {
			tc.Fatalf("invalid kernel output: %s", out)
		}
		if !reflect.DeepEqual(actual, expected) {
			tc.Fatalf("selected %v, expected %v", actual, expected)
		}
	}, options(100)...)
}

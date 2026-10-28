package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// tpbPartition measures the covered endpoint cells. Unlike a merging sweep,
// it never extends an active interval or accumulates a merged range.
func tpbPartition(children [][2]uint64) uint64 {
	points := []uint64{}
	for _, c := range children {
		points = append(points, c[0], c[1])
	}
	sort.Slice(points, func(i, j int) bool { return points[i] < points[j] })
	var covered uint64
	for i := 1; i < len(points); i++ {
		left, right := points[i-1], points[i]
		for _, c := range children {
			if c[0] <= left && c[1] >= right {
				covered += right - left
				break
			}
		}
	}
	return covered
}
func tpbRow(id, parent string, start, duration uint64) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"TraceId": "1234567890abcdef1234567890abcdef", "SpanId": id, "ParentSpanId": parent, "timestamp_ns": strconv.FormatUint(start, 10), "duration_ns": strconv.FormatUint(duration, 10), "ServiceName": "checkout", "SpanName": "work"})
	return b
}
func TestHegelTracePageDirectChildIntervals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		width := hegel.Draw(tc, hegel.Integers(1, 80))
		count := hegel.Draw(tc, hegel.Integers(0, 15))
		children := make([][2]uint64, count)
		for i := range children {
			a := hegel.Draw(tc, hegel.Integers(0, width))
			b := hegel.Draw(tc, hegel.Integers(a, width))
			children[i] = [2]uint64{uint64(a), uint64(b)}
		}
		want := tpbPartition(children)
		// A second, deliberately discrete oracle covers every unit time cell.
		grid := uint64(0)
		for x := uint64(0); x < uint64(width); x++ {
			for _, c := range children {
				if c[0] <= x && x < c[1] {
					grid++
					break
				}
			}
		}
		if want != grid {
			tc.Errorf("independent endpoint/grid disagreement")
			return
		}
		// Large epoch and scale retain nanoseconds beyond binary64 exactness.
		for _, transform := range [][2]uint64{{0, 1}, {9007199254740993, 1}, {9007199254740993, 1000000000000}} {
			offset, scale := transform[0], transform[1]
			rows := []json.RawMessage{tpbRow("0000000000000001", "", offset, uint64(width)*scale)}
			for i, c := range children {
				rows = append(rows, tpbRow(fmt.Sprintf("%016x", i+2), "0000000000000001", offset+c[0]*scale, (c[1]-c[0])*scale))
			}
			before, _ := json.Marshal(rows)
			result, e := exploration.AnalyzeTracePage(rows, false)
			after, _ := json.Marshal(rows)
			if e != nil || string(before) != string(after) {
				tc.Errorf("rejected/mutated valid page %v", e)
				return
			}
			r := result.Rows[0]
			if r.State != "available" || r.ChildUnionNS == nil || *r.ChildUnionNS != strconv.FormatUint(want*scale, 10) || r.OutsideChildrenNS == nil || *r.OutsideChildrenNS != strconv.FormatUint((uint64(width)-want)*scale, 10) {
				tc.Errorf("incorrect direct-child union %+v", r)
				return
			}
			// Row permutation changes indexes only; the parent measure is invariant.
			for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
				rows[i], rows[j] = rows[j], rows[i]
			}
			permuted, e := exploration.AnalyzeTracePage(rows, true)
			if e != nil || !permuted.Truncated {
				tc.Errorf("permutation/truncation %v", e)
				return
			}
			p := permuted.Rows[len(rows)-1]
			if p.ChildUnionNS == nil || *p.ChildUnionNS != *r.ChildUnionNS || p.OutsideChildrenNS == nil || *p.OutsideChildrenNS != *r.OutsideChildrenNS {
				tc.Errorf("permutation changed measure")
			}
		}
	}, options(80)...)
}

func TestTracePageBreakdownAuthorityBoundaries(t *testing.T) {
	root := func() json.RawMessage { return tpbRow("0000000000000001", "", 100, 100) }
	linked := tpbRow("0000000000000002", "", 120, 10)
	var linkedObject map[string]any
	json.Unmarshal(linked, &linkedObject)
	linkedObject["linked_traces"] = []string{"1234567890abcdef1234567890abcdef"}
	linkedObject["linked_spans"] = []string{"0000000000000001"}
	linked, _ = json.Marshal(linkedObject)

	cases := []struct {
		name      string
		rows      []json.RawMessage
		index     int
		available bool
		reason    string
		union     string
	}{
		{"zero duration", []json.RawMessage{tpbRow("0000000000000001", "", 100, 0)}, 0, true, "", "0"},
		{"links never become child edges", []json.RawMessage{root(), linked}, 0, true, "", "0"},
		{"missing parent does not hide own children", []json.RawMessage{tpbRow("0000000000000001", "0000000000000099", 100, 100), tpbRow("0000000000000002", "0000000000000001", 120, 10)}, 0, true, "parent-outside-page", "10"},
		{"outside child withholds instead of clipping", []json.RawMessage{root(), tpbRow("0000000000000002", "0000000000000001", 90, 20)}, 0, false, "child-outside-parent", ""},
		{"duplicate child is ambiguous", []json.RawMessage{root(), tpbRow("0000000000000002", "0000000000000001", 120, 10), tpbRow("0000000000000002", "0000000000000001", 140, 10)}, 0, false, "unusable-child-edge", ""},
		{"cycle member unavailable", []json.RawMessage{tpbRow("0000000000000001", "0000000000000002", 100, 100), tpbRow("0000000000000002", "0000000000000001", 100, 100)}, 0, false, "parent-cycle", ""},
		{"unknown child residual does not poison valid ancestor edge", []json.RawMessage{root(), tpbRow("0000000000000002", "0000000000000001", 120, 20), tpbRow("0000000000000003", "0000000000000002", 110, 20)}, 0, true, "", "20"},
		{"overflow interval", []json.RawMessage{tpbRow("0000000000000001", "", 18446744073709551615, 1)}, 0, false, "invalid-interval", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, e := exploration.AnalyzeTracePage(c.rows, false)
			if e != nil {
				t.Fatal(e)
			}
			r := got.Rows[c.index]
			if (r.State == "available") != c.available {
				t.Fatalf("authority state %+v", r)
			}
			if c.available {
				if r.ChildUnionNS == nil || *r.ChildUnionNS != c.union {
					t.Fatalf("union %+v", r)
				}
			} else if r.ChildUnionNS != nil || r.OutsideChildrenNS != nil {
				t.Fatalf("unavailable arithmetic exposed %+v", r)
			}
			if c.reason != "" {
				found := false
				for _, reason := range r.Reasons {
					found = found || reason == c.reason
				}
				if !found {
					t.Fatalf("missing reason %s: %+v", c.reason, r)
				}
			}
		})
	}
	// Group sums are arbitrary-precision inclusive observations, not trace wall time.
	rows := []json.RawMessage{tpbRow("0000000000000001", "", 0, 18446744073709551615), tpbRow("0000000000000002", "", 0, 18446744073709551615)}
	got, e := exploration.AnalyzeTracePage(rows, false)
	if e != nil || len(got.Groups) != 1 || got.Groups[0].InclusiveNS == nil || *got.Groups[0].InclusiveNS != "36893488147419103230" || got.Groups[0].OutsideChildrenNS == nil || *got.Groups[0].OutsideChildrenNS != "36893488147419103230" {
		t.Fatalf("overflowed group aggregation %+v %v", got, e)
	}
}

func TestHegelTracePageJavaScriptParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		mode := hegel.Draw(tc, hegel.Integers(0, 9))
		start := uint64(9007199254740993) + uint64(hegel.Draw(tc, hegel.Integers(0, 1000)))
		rows := []json.RawMessage{tpbRow("0000000000000001", "", start, 100), tpbRow("0000000000000002", "0000000000000001", start+10, 40), tpbRow("0000000000000003", "0000000000000001", start+30, 50)}
		var child map[string]any
		json.Unmarshal(rows[1], &child)
		switch mode {
		case 1:
			child["ParentSpanId"] = "0000000000000099"
		case 2:
			child["timestamp_ns"] = strconv.FormatUint(start-1, 10)
		case 3:
			rows = append(rows, rows[1])
		case 4:
			child["ParentSpanId"] = "0000000000000003"
			var third map[string]any
			json.Unmarshal(rows[2], &third)
			third["ParentSpanId"] = "0000000000000002"
			rows[2], _ = json.Marshal(third)
		case 5:
			child["duration_ns"] = "18446744073709551615"
		case 6:
			child["ParentSpanId"] = nil
		case 7:
			child["ServiceName"] = nil
			child["SpanName"] = ""
		case 8:
			child["timestamp_ns"] = "01"
		case 9:
			child["ServiceName"] = 17
		}
		rows[1], _ = json.Marshal(child)
		want, e := exploration.AnalyzeTracePage(rows, mode%2 == 0)
		if e != nil {
			tc.Errorf("Go rejected page %v", e)
			return
		}
		if !tpbOracleEqual(rows, mode%2 == 0, *want) {
			tc.Errorf("independent graph/interval oracle disagrees mode%d", mode)
			return
		}

		input, _ := json.Marshal(map[string]any{"rows": rows, "truncated": mode%2 == 0})
		cmd := exec.Command("node", "-e", `const fs=require('fs'),m=require('../ui/trace-page-breakdown.mjs'),x=JSON.parse(fs.readFileSync(0,'utf8'));process.stdout.write(JSON.stringify(m.analyze(x.rows,x.truncated)));`)
		cmd.Stdin = bytes.NewReader(input)
		raw, e := cmd.Output()
		if e != nil {
			tc.Errorf("JS failed %v", e)
			return
		}
		expected, _ := json.Marshal(want)
		if !aeJSONEqual(json.RawMessage(raw), json.RawMessage(expected)) {
			tc.Errorf("Go/JS whole DTO mismatch mode%d\nGo %s\nJS %s", mode, expected, raw)
		}
	}, options(70)...)
}

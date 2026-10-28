package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

func tcRow(trace, span, parent, service, operation, kind string, start, duration uint64) json.RawMessage {
	raw, _ := json.Marshal(map[string]any{"TraceId": trace, "SpanId": span, "ParentSpanId": parent, "ServiceName": service, "SpanName": operation, "SpanKind": kind, "timestamp_ns": strconv.FormatUint(start, 10), "duration_ns": strconv.FormatUint(duration, 10)})
	return raw
}
func tcSet(raw json.RawMessage, key string, value any) json.RawMessage {
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	obj[key] = value
	out, _ := json.Marshal(obj)
	return out
}
func tcPages(seed, width, n, mode int) ([]json.RawMessage, []json.RawMessage) {
	makePage := func(current bool) []json.RawMessage {
		trace := fmt.Sprintf("%032x", seed)
		offset := uint64(9007199254740993)
		duration := uint64(width)
		if current {
			trace = fmt.Sprintf("%032x", seed+1)
			offset += 1000000
			duration += 17
		}
		rows := []json.RawMessage{tcRow(trace, "0000000000000001", "", "checkout", "root", "Server", offset, duration)}
		for i := 0; i < n; i++ {
			start := uint64((i * 7) % (width + 1))
			d := uint64((width - int(start)) / 2)
			if current {
				start += 10
				if d > 0 {
					d--
				}
			}
			rows = append(rows, tcRow(trace, fmt.Sprintf("%016x", i+2), "0000000000000001", "worker", fmt.Sprintf("operation-%d", i%3), "Internal", offset+start, d))
		}
		return rows
	}
	a, b := makePage(false), makePage(true)
	switch mode {
	case 1:
		b[0] = tcSet(b[0], "SpanKind", "Client")
	case 2:
		b[0] = tcSet(b[0], "SpanKind", "")
	case 3:
		b = append(b, b[1])
	case 4:
		b[1] = tcSet(b[1], "TraceId", fmt.Sprintf("%032x", seed+2))
	case 5:
		b[0] = tcSet(b[0], "timestamp_ns", "18446744073709551615")
	case 6:
		b[1] = tcSet(b[1], "ParentSpanId", "000000000000ffff")
	case 7:
		b[1] = tcSet(b[1], "timestamp_ns", "9007199254740992")
	case 8:
		b[0] = tcSet(b[0], "ServiceName", nil)
	case 9:
		b[0] = tcSet(b[0], "ServiceName", "")
	case 10:
		b[0] = tcSet(b[0], "ServiceName", 17)
	case 11:
		b[1] = tcSet(b[1], "SpanName", "selection-only")
	case 12:
		b[0] = tcSet(b[0], "timestamp_ns", "not-reported")
	case 13:
		b[1] = tcSet(b[1], "ParentSpanId", nil)
	case 14:
		b[0] = tcSet(b[0], "SpanId", "0000000000000002")
	}
	return a, b
}
func tcGroupComparable(a, b exploration.TracePageComparison) bool {
	// Metamorphisms may change row indexes, absolute timestamps and reason order.
	// They must retain identity unions, complete-known group sums and match states.
	if len(a.Groups) != len(b.Groups) {
		return false
	}
	for i, g := range a.Groups {
		h := b.Groups[i]
		if tcKey(g.Service, g.Operation) != tcKey(h.Service, h.Operation) || g.MatchState != h.MatchState || g.MatchReason != h.MatchReason || !aeJSONEqual(g.InclusiveDifferenceNS, h.InclusiveDifferenceNS) || !aeJSONEqual(g.OutsideDifferenceNS, h.OutsideDifferenceNS) {
			return false
		}
		if (g.Match == nil) != (h.Match == nil) {
			return false
		}
		if g.Match != nil && (!aeJSONEqual(g.Match.InclusiveDifferenceNS, h.Match.InclusiveDifferenceNS) || !aeJSONEqual(g.Match.OutsideDifferenceNS, h.Match.OutsideDifferenceNS) || !aeJSONEqual(g.Match.RelativeStartDifferenceNS, h.Match.RelativeStartDifferenceNS)) {
			return false
		}
	}
	return true
}
func TestHegelTraceComparisonIndependentRawOracle(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 80))
		n := hegel.Draw(tc, hegel.Integers(1, 9))
		mode := hegel.Draw(tc, hegel.Integers(0, 14))
		a, b := tcPages(seed, width, n, mode)
		at, bt := mode%2 == 0, mode%3 == 0
		before, _ := json.Marshal([][]json.RawMessage{a, b})
		got, e := exploration.CompareTracePages(a, b, at, bt)
		after, _ := json.Marshal([][]json.RawMessage{a, b})
		if e != nil || string(before) != string(after) {
			tc.Errorf("pure model rejected/mutated raw inputs %v", e)
			return
		}
		want := tcOracle(a, b, at, bt)
		if !tcOracleEqual(want, *got) {
			tc.Errorf("independent raw oracle mismatch mode%d", mode)
			return
		}
		swapped, e := exploration.CompareTracePages(b, a, bt, at)
		if e != nil || !tcOracleEqual(tcOracle(b, a, bt, at), *swapped) {
			tc.Errorf("side swap mismatch")
			return
		}
		for i, g := range got.Groups {
			z := swapped.Groups[i]
			for j, v := range []*string{g.InclusiveDifferenceNS, g.OutsideDifferenceNS} {
				w := []*string{z.InclusiveDifferenceNS, z.OutsideDifferenceNS}[j]
				if v == nil {
					if w != nil {
						tc.Errorf("unknown delta became known")
					}
				} else if w == nil {
					tc.Errorf("known delta became absent")
				} else {
					x, _ := new(big.Int).SetString(*v, 10)
					y, _ := new(big.Int).SetString(*w, 10)
					if new(big.Int).Add(x, y).Sign() != 0 {
						tc.Errorf("delta sign lost")
					}
				}
			}
		}
		// Simultaneous page permutation cannot justify pairing repeated labels.
		reverse := func(rows []json.RawMessage) []json.RawMessage {
			out := append([]json.RawMessage{}, rows...)
			for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
				out[i], out[j] = out[j], out[i]
			}
			return out
		}
		pa, pb := reverse(a), reverse(b)
		permuted, e := exploration.CompareTracePages(pa, pb, at, bt)
		if e != nil || !tcOracleEqual(tcOracle(pa, pb, at, bt), *permuted) || !tcGroupComparable(*got, *permuted) {
			tc.Errorf("row permutation invented pairing/changed groups")
		}
	}, options(80)...)
}
func TestHegelTraceComparisonJavaScriptParity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 80))
		n := hegel.Draw(tc, hegel.Integers(1, 9))
		mode := hegel.Draw(tc, hegel.Integers(0, 14))
		a, b := tcPages(seed, width, n, mode)
		if mode%2 == 0 {
			a, b = tcAffine(a, 9007199254740993, 1000000000000001), tcAffine(b, 9007199255740993, 1000000000000001)
		}
		at, bt := mode%2 == 0, mode%3 == 0
		got, e := exploration.CompareTracePages(a, b, at, bt)
		if e != nil {
			tc.Errorf("Go projection %v", e)
			return
		}
		if !tcOracleEqual(tcOracle(a, b, at, bt), *got) {
			tc.Errorf("Go vs independent oracle")
			return
		}
		input, _ := json.Marshal(map[string]any{"baseline": a, "selection": b, "baselineTruncated": at, "selectionTruncated": bt})
		cmd := exec.Command("node", "-e", `const fs=require('fs'),m=require('../ui/trace-page-comparison.mjs'),x=JSON.parse(fs.readFileSync(0,'utf8'));process.stdout.write(JSON.stringify(m.compare(x.baseline,x.selection,x.baselineTruncated,x.selectionTruncated)));`)
		cmd.Stdin = bytes.NewReader(input)
		raw, e := cmd.Output()
		if e != nil {
			tc.Errorf("actual JS failed %v", e)
			return
		}
		expected, _ := json.Marshal(got)
		if !aeJSONEqual(json.RawMessage(raw), json.RawMessage(expected)) {
			tc.Errorf("Go/actual Node full DTO parity mode%d\nGo %s\nJS %s", mode, expected, raw)
		}
	}, options(70)...)
}
func TestTraceComparisonKnownnessAndAuthority(t *testing.T) {
	a := []json.RawMessage{tcRow("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0000000000000001", "", "edge", "root", "Server", 0, 0)}
	cases := []struct {
		name                string
		baseline, selection []json.RawMessage
		state, reason       string
		delta               *string
		relative            bool
	}{
		{"zero is known", a, a, "unique-label-match", "", lpcPointer("0"), true},
		{"missing labels are not identity", []json.RawMessage{tcSet(a[0], "ServiceName", nil)}, []json.RawMessage{tcSet(a[0], "ServiceName", nil)}, "unavailable", "unreported-label", lpcPointer("0"), false},
		{"present empty label distinct from missing", []json.RawMessage{tcSet(a[0], "ServiceName", "")}, []json.RawMessage{tcSet(a[0], "ServiceName", "")}, "unique-label-match", "", lpcPointer("0"), true},

		{"absent is not zero", a, []json.RawMessage{}, "baseline-only", "", nil, false},
		{"repeats never zip", append(append([]json.RawMessage{}, a...), tcRow("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "0000000000000002", "", "edge", "root", "Server", 2, 4)), a, "ambiguous", "repeated-label", lpcPointer("-4"), false},
		{"kind mismatch", a, []json.RawMessage{tcSet(a[0], "SpanKind", "Client")}, "unavailable", "span-kind-mismatch-or-unreported", lpcPointer("0"), false},
		{"missing kind", a, []json.RawMessage{tcSet(a[0], "SpanKind", nil)}, "unavailable", "span-kind-mismatch-or-unreported", lpcPointer("0"), false},
		{"timing unknown inclusive known", a, []json.RawMessage{tcSet(a[0], "timestamp_ns", "18446744073709551616")}, "unique-label-match", "", lpcPointer("0"), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, e := exploration.CompareTracePages(c.baseline, c.selection, false, false)
			if e != nil || !tcOracleEqual(tcOracle(c.baseline, c.selection, false, false), *got) {
				t.Fatalf("independent contract %v", e)
			}
			g := got.Groups[0]
			if g.MatchState != c.state || g.MatchReason != c.reason || !aeJSONEqual(g.InclusiveDifferenceNS, c.delta) {
				t.Fatalf("wrong authority/zero %+v", g)
			}
			if c.relative && (g.Match == nil || g.Match.RelativeStartDifferenceNS == nil || *g.Match.RelativeStartDifferenceNS != "0") {
				t.Fatal("known zero relative start lost")
			}
			if !c.relative && g.Match != nil && g.Match.RelativeStartDifferenceNS != nil {
				t.Fatal("relative timing invented")
			}
		})
	}
}

func tcAffine(rows []json.RawMessage, origin, scale uint64) []json.RawMessage {
	var low *big.Int
	for _, raw := range rows {
		a, _ := tcInterval(raw)
		if a != nil && (low == nil || a.Cmp(low) < 0) {
			low = a
		}
	}
	out := make([]json.RawMessage, len(rows))
	for i, raw := range rows {
		a, z := tcInterval(raw)
		if a == nil {
			out[i] = raw
			continue
		}
		duration := new(big.Int).Sub(z, a)
		start := new(big.Int).Sub(a, low)
		start.Mul(start, new(big.Int).SetUint64(scale))
		start.Add(start, new(big.Int).SetUint64(origin))
		duration.Mul(duration, new(big.Int).SetUint64(scale))
		out[i] = tcSet(tcSet(raw, "timestamp_ns", start.String()), "duration_ns", duration.String())
	}
	return out
}
func TestHegelTraceComparisonAffineIntervals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		width := hegel.Draw(tc, hegel.Integers(1, 80))
		n := hegel.Draw(tc, hegel.Integers(1, 9))
		a, b := tcPages(seed, width, n, 0)
		base, e := exploration.CompareTracePages(a, b, false, false)
		if e != nil {
			tc.Errorf("base %v", e)
			return
		}
		for _, scale := range []uint64{1, 1000000000000001} {
			pa, pb := tcAffine(a, 9007199254740993, scale), tcAffine(b, 9007199255740993, scale)
			got, e := exploration.CompareTracePages(pa, pb, false, false)
			if e != nil || !tcOracleEqual(tcOracle(pa, pb, false, false), *got) {
				tc.Errorf("affine independent oracle scale%d %v", scale, e)
				return
			}
			for i, g := range base.Groups {
				z := got.Groups[i]
				if g.MatchState != z.MatchState || g.MatchReason != z.MatchReason {
					tc.Errorf("affine invented authority")
				}
				values := []*string{g.InclusiveDifferenceNS, g.OutsideDifferenceNS}
				scaled := []*string{z.InclusiveDifferenceNS, z.OutsideDifferenceNS}
				if g.Match != nil && z.Match != nil {
					values = append(values, g.Match.RelativeStartDifferenceNS)
					scaled = append(scaled, z.Match.RelativeStartDifferenceNS)
				}
				for j, v := range values {
					w := scaled[j]
					if v == nil {
						if w != nil {
							tc.Errorf("affine filled unknown")
						}
						continue
					}
					if w == nil {
						tc.Errorf("affine dropped known")
						continue
					}
					x, _ := new(big.Int).SetString(*v, 10)
					x.Mul(x, new(big.Int).SetUint64(scale))
					if x.String() != *w {
						tc.Errorf("affine rounded/sign-lost delta")
					}
				}
			}
		}
	}, options(70)...)
}

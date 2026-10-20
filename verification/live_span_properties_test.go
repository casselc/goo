package verification

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

func TestHegelLiveSpanMeasurePopulationAndReversal(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1001, 3000))
		left := hegel.Draw(tc, hegel.Integers(0, 1000000000))
		right := hegel.Draw(tc, hegel.Integers(0, 1000000000))
		aErrors := hegel.Draw(tc, hegel.Integers(0, n))
		bErrors := hegel.Draw(tc, hegel.Integers(0, n))
		makeSide := func(ns, errors int) *exploration.SavedCapture {
			values := make([]uint64, n)
			for i := range values {
				values[i] = uint64(ns)
			}
			summary := exploration.ReferenceDurationSummary(values)
			count := strconv.Itoa(n)
			q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100, Summary: "span-duration", SpanStatus: true}
			return &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, Title: "Span evidence", CapturedAt: "2026-10-20T00:00:00Z", Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Question: q, State: "observed", Count: count, Basis: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Duration: &summary, SpanStatus: &exploration.SpanStatusSummary{Count: count, Error: strconv.Itoa(errors), OK: "0", Unset: "0", Unknown: strconv.Itoa(n - errors)}, Rows: []json.RawMessage{json.RawMessage(`{"duration_ns":"999999999999","StatusCode":"Ok"}`)}}}
		}
		a, b := makeSide(left, aErrors), makeSide(right, bErrors)
		before, _ := json.Marshal([]*exploration.SavedCapture{a, b})
		forward, e := exploration.CompareEvidence(a, b)
		if e != nil {
			tc.Fatal(e)
		}
		reverse, e := exploration.CompareEvidence(b, a)
		if e != nil {
			tc.Fatal(e)
		}
		if forward.Measures.P95DifferenceNS != fmt.Sprint(right-left) || forward.Measures.SelectionDuration.Count != strconv.Itoa(n) || forward.Measures.SelectionStatus.Unknown != strconv.Itoa(n-bErrors) {
			tc.Fatal("bounded rows replaced the source measure population")
		}
		x, _ := new(big.Rat).SetString(forward.Measures.ErrorDifferencePP)
		y, _ := new(big.Rat).SetString(reverse.Measures.ErrorDifferencePP)
		if new(big.Rat).Add(x, y).Sign() != 0 {
			tc.Fatal("swapping sides broke error-share reversal")
		}
		after, _ := json.Marshal([]*exploration.SavedCapture{a, b})
		if string(before) != string(after) {
			tc.Fatal("comparison mutated snapshot")
		}
	}, options(100)...)
}

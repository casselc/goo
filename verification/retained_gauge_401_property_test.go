package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

// Generated carrier admission uses existing Go evaluation/projection and actual
// JavaScript retained owners. It is local semantic exploration, not source data.
func TestHegelRetainedGaugeContext401(t *testing.T) {
	type vector struct {
		Result *measurements.Result                    `json:"result"`
		Page   *exploration.CapturePage                `json:"page"`
		Stream metrics.Stream                          `json:"stream"`
		AtNS   string                                  `json:"atNS"`
		Plans  []*exploration.NativeObservationContext `json:"plans"`
	}
	vectors := []vector{}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		mode := hegel.Draw(tc, hegel.Integers(0, 5))
		namespace := hegel.Draw(tc, hegel.Integers(0, 2))
		width := hegel.Draw(tc, hegel.SampledFrom([]int64{1, 59000000123, 121000000123, 180000000123}))
		lo := int64(1790000000000000000) + int64(seed)
		hi := lo + width
		at := []int64{lo, lo + width/2, hi - 1}[seed%3]
		result, stream := observationResult388(t, seed, mode, namespace, lo, hi, at, seed%2 == 0)
		before, _ := json.Marshal(result)
		question := (exploration.CaptureQuestion{Kind: "measurement", Measurement: &result.Question}).Normalized()
		stored := exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, ID: strings.Repeat("a", 64), Title: "Retained gauge", CapturedAt: time.UnixMilli(result.Question.ToMS + 2000).UTC().Format(time.RFC3339Nano), Question: question}, Measurement: result}
		page, err := stored.Page("")
		if err != nil {
			tc.Fatal(err)
		}
		v := vector{Result: result, Page: page, Stream: stream, AtNS: strconv.FormatInt(at, 10)}
		for _, signal := range []string{"logs", "spans"} {
			p, e := exploration.PlanNativeObservationContext(result, stream, signal, v.AtNS)
			if e != nil {
				tc.Fatal(e)
			}
			v.Plans = append(v.Plans, p)
		}
		after, _ := json.Marshal(result)
		if !bytes.Equal(before, after) {
			tc.Fatal("carrier construction mutated retained result")
		}
		vectors = append(vectors, v)
	}, options(24)...)
	input, _ := json.Marshal(vectors)
	if err := artifact("retained-gauge401-carriers.json", input); err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "node", "retained_gauge_401_fixture.mjs")
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("retained gauge carrier parity: %v\n%s", err, out)
	}
	var got struct{ Cases, Plans, Controls int }
	if json.Unmarshal(out, &got) != nil || got.Cases != len(vectors) || got.Plans != len(vectors)*2 || got.Controls < len(vectors)*15 {
		t.Fatalf("incomplete retained admission %s", out)
	}
	// Distinct clocks stay distinct in every generated full saved carrier.
	for _, v := range vectors {
		if v.Page.Receipt.CapturedAt == v.Result.ReadAt || !reflect.DeepEqual(v.Page.Receipt.Question.Measurement, &v.Result.Question) {
			t.Fatal("custody/read/question identity collapsed")
		}
	}
	t.Logf("%d generated full Saved carriers / %d existing Go+JS observation plans / %d retained-owner refusal controls", got.Cases, got.Plans, got.Controls)
}

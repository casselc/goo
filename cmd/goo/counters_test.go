package main

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/metrics"
	"testing"
)

func TestEvaluateCountersCLI(t *testing.T) {
	var out bytes.Buffer
	if err := run([]string{"evaluate-counters", "-input", "../../examples/metrics/counters.json"}, &out); err != nil {
		t.Fatal(err)
	}
	var result metrics.CounterResult
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Series[0].State != "ok" || result.Series[0].Intervals[1].Delta != "4" {
		t.Fatal("CLI lost counter semantics")
	}
}

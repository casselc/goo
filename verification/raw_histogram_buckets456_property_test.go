package verification

import (
	"bytes"
	"encoding/json"
	"math"
	"os/exec"
	"strconv"
	"testing"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

// Synthetic source-admitted points test display semantics independently of the
// browser journey. Counts stay uint64 strings; geometry has no measurement role.
func TestHegelRawHistogramBuckets456Conservation(t *testing.T) {
	type fixture struct {
		Row       measurements.RawHistogramObservation `json:"row"`
		Counts    []string                             `json:"-"`
		Intervals []string                             `json:"-"`
		State     string                               `json:"-"`
		Eligible  bool                                 `json:"-"`
		Maximum   uint64                               `json:"-"`
	}
	fixtures := []fixture{}
	schema := histogram447Schema(t)
	admit := func(size, seed, mode int) {
		q := histogram447Question([]int32{-7, 0, 1, 2, 7}[seed%5])
		q.Instrument.Unit = "widgets"
		raw := histogram447Input(q, 0, 0)
		counts := make([]string, size)
		bounds := make([]string, size-1)
		var total, maximum uint64
		for i := range counts {
			amount := uint64((i*37 + seed) % 101)
			if mode == 1 {
				amount = 0
			}
			if mode == 2 {
				amount = 0
				if i == 0 {
					amount = math.MaxUint64 - 1
				}
				if i == 1 {
					amount = 1
				}
			}
			counts[i] = strconv.FormatUint(amount, 10)
			total += amount
			maximum = max(maximum, amount)
			if i < len(bounds) {
				bounds[i] = strconv.Itoa(i - size/2)
			}
		}
		raw["count"] = strconv.FormatUint(total, 10)
		raw["bounds"] = bounds
		raw["bucket_counts"] = counts
		f := fixture{Counts: counts, Eligible: true, State: "observed", Maximum: maximum}
		if total == 0 {
			f.State = "no_data"
		}
		for i := range counts {
			interval := "(−∞, +∞)"
			if size > 1 {
				switch i {
				case 0:
					interval = "(−∞, " + bounds[0] + "]"
				case size - 1:
					interval = "(" + bounds[i-1] + ", +∞)"
				default:
					interval = "(" + bounds[i-1] + ", " + bounds[i] + "]"
				}
			}
			f.Intervals = append(f.Intervals, interval)
		}
		switch mode {
		case 3:
			raw["flags"] = uint32(2)
			f.Eligible = false
			f.State = "unknown"
		case 4:
			raw["flags"] = uint32(1)
			f.Eligible = false
			f.State = "unavailable"
			f.Counts = nil
			f.Intervals = nil
			f.Maximum = 0
		case 5:
			raw["count"] = strconv.FormatUint(total+1, 10)
			f.Eligible = false
			f.State = "unavailable"
			f.Counts = nil
			f.Intervals = nil
			f.Maximum = 0
		case 6:
			delete(raw, "flags")
			f.Eligible = false
			f.State = "unknown"
		case 7:
			raw["flags"] = 0.5
			f.Eligible = false
			f.State = "unknown"
		case 8:
			raw["bounds"] = []string{}
			raw["bucket_counts"] = []string{}
			f.Eligible = false
			f.State = "unavailable"
			f.Counts = nil
			f.Intervals = nil
			f.Maximum = 0
		}
		wire, err := json.Marshal(raw)
		if err != nil {
			t.Fatal(err)
		}
		bank, err, _ := histogram447Read(schema, []json.RawMessage{wire}, q)
		if err != nil {
			t.Fatal(err)
		}
		if len(bank.Rows) != 1 {
			t.Fatal("source dropped point")
		}
		f.Row = bank.Rows[0]
		fixtures = append(fixtures, f)
	}
	hegel.Test(t, func(tc *hegel.T) {
		admit(hegel.Draw(tc, hegel.Integers(2, 40)), hegel.Draw(tc, hegel.Integers(0, 10000)), hegel.Draw(tc, hegel.Integers(0, 8)))
	}, options(50)...)
	admit(2, 0, 2)     // UInt64 maximum total, exact maximum-minus-one and one.
	admit(1, 0, 1)     // One all-values bucket is valid even with an empty distribution.
	admit(10001, 1, 0) // Complete vector remains present; presentation pages independently.
	wire, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	script := `import fs from 'node:fs';import {model} from '../ui/raw-histogram-buckets.mjs';globalThis.fetch=()=>{throw Error('bucket display attempted source read')};const out=[];for(const f of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(f.row),m=model(f.row);out.push(structuredClone(m));if(m.rows.length)m.rows[0].count='mutated';if(JSON.stringify(f.row)!==before)throw Error('source aliased');const bad=structuredClone(f.row);bad.bucketCounts=['999999'];if(model(bad).state!=='unavailable')throw Error('forged normalized vector admitted')}process.stdout.write(JSON.stringify(out));`
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(wire)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JS bucket model: %v %s", err, out)
	}
	var results []struct {
		State    string
		Eligible bool
		Maximum  string
		Count    *string
		Rows     []struct {
			Index                        int
			Interval, RawInterval, Count string
			Fraction                     float64
		}
	}
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != len(fixtures) {
		t.Fatal("lost point")
	}
	for i, f := range fixtures {
		m := results[i]
		if m.State != f.State || m.Eligible != f.Eligible || len(m.Rows) != len(f.Counts) || m.Maximum != strconv.FormatUint(f.Maximum, 10) {
			t.Fatalf("fixture%d state/partition: %+v expected %s eligible%v rows%d maximum%d", i, m, f.State, f.Eligible, len(f.Counts), f.Maximum)
		}
		if (m.Count == nil) != (f.Row.Count == nil) || m.Count != nil && *m.Count != *f.Row.Count {
			t.Fatal("exact source count changed")
		}
		for j, bucket := range m.Rows {
			if bucket.Index != j || bucket.Count != f.Counts[j] || bucket.RawInterval != f.Intervals[j] || bucket.Interval != f.Intervals[j] {
				t.Fatalf("fixture%d bucket%d exact count or boundary changed", i, j)
			}
			if math.IsNaN(bucket.Fraction) || math.IsInf(bucket.Fraction, 0) || bucket.Fraction < 0 || bucket.Fraction > 1 {
				t.Fatal("nonfinite/invalid disposable geometry")
			}
			amount, err := strconv.ParseUint(f.Counts[j], 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			want := 0.0
			if f.Maximum != 0 {
				want = float64(amount) / float64(f.Maximum)
			}
			if math.Abs(bucket.Fraction-want) > 1e-15 {
				t.Fatalf("fixture%d bucket%d fraction %g want%g", i, j, bucket.Fraction, want)
			}
		}
	}
	t.Logf("%d synthetic Go-source-admitted points: exact uint64 counts and categorical boundaries, zero/unknown/invalid recording partition, finite geometry, all10001 buckets retained, no IO or aliases", len(fixtures))
}

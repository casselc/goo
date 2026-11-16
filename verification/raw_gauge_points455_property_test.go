package verification

import (
	"bytes"
	"encoding/json"
	"math"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
)

type gauge455Member struct {
	Key, EndNS string
	RawValue   json.RawMessage
}
type gauge455Group struct {
	Members        []gauge455Member
	OccurrenceKeys []string
	RawValues      []json.RawMessage
	Multiplicity   int
	X, Y           float64
}
type gauge455Model struct {
	State    string
	Points   []gauge455Group
	Excluded []struct {
		gauge455Member
		Reason string
	}
	Counts           struct{ Loaded, Eligible, Plotted, Unplotted, NotShownEligible, Undisplayed, UnknownRecording, NoRecordedValue, InvalidValue, RenderedLocations int }
	CoordinateGroups int
	FromNS, ToNS     string
	AtLimit          bool
}

// These are synthetic Go-source-admitted banks, not live ingestion evidence.
// The source adapter owns scalar/flag parsing; this oracle independently partitions
// those admitted facts and computes disposable geometry and member conservation.
func TestHegelRawGaugePoints455Partition(t *testing.T) {
	type fixture struct {
		Data        *measurements.RawGaugeResult `json:"data"`
		Groups      [][]gauge455Member           `json:"-"`
		Coordinates [][2]float64                 `json:"-"`
		Excluded    map[string]string            `json:"-"`
		Eligible    int                          `json:"-"`
	}
	fixtures := []fixture{}
	admit := func(size, seed int, boundary bool) {
		q := gauge438Question()
		q.Question.FromMS = 1700000000000
		q.Question.ToMS = q.Question.FromMS + 1000
		input := []json.RawMessage{}
		for i := 0; i < size; i++ {
			values := []any{"0", "-0", "4.9406564584124654e-324", "-4.9406564584124654e-324", "1.7976931348623157e308", "-1.7976931348623157e308", "12.500", "12.5", "NaN", "+Inf", "bad", nil}
			value := values[(i+seed)%len(values)]
			flags := uint32((i/len(values) + seed) % 4)
			row := gauge438Input(value, flags, 0)
			row["resource"] = map[string]string{"service.name": "api", "deployment.environment.name": "prod"}
			row["attributes"] = map[string]string{}
			delta := int64((i+seed)%4 + 1)
			if boundary {
				value = "1"
				row["value"] = value
				row["flags"] = uint32(0)
				if size == 10000 {
					delta = 1
				} else {
					delta = int64(i + 1)
				}
			} else {
				switch (i + seed) % 9 {
				case 4:
					delete(row, "flags")
				case 5:
					row["flags"] = 0.5
				case 6:
					row["flags"] = uint32(0)
				case 7:
					row["flags"] = uint32(1)
				}
			}
			if !boundary && i < 6 {
				row["value"] = []string{"1.7976931348623157e308", "-1.7976931348623157e308", "4.9406564584124654e-324", "0", "12.500", "12.5"}[i]
				row["flags"] = uint32(0)
				delta = 1
			}
			row["end_ns"] = strconv.FormatInt(q.Question.FromMS*1000000+delta, 10)
			raw, err := json.Marshal(row)
			if err != nil {
				t.Fatal(err)
			}
			input = append(input, raw)
		}
		bank, err, _ := gauge438Read(t, metricScopeSchema432(t), input, q)
		if err != nil {
			t.Fatal(err)
		}
		f := fixture{Data: bank, Excluded: map[string]string{}}
		eligible := []measurements.RawObservation{}
		min, max := 0.0, 0.0
		for _, row := range bank.Rows {
			why := ""
			if row.FlagsState != "valid" || row.Flags == nil {
				why = "recording-unknown"
			} else if *row.Flags&1 != 0 {
				why = "not-recorded"
			} else if *row.Flags != 0 {
				why = "recording-unknown"
			} else if row.ValueState != "finite" || row.Value == nil || math.IsNaN(*row.Value) || math.IsInf(*row.Value, 0) {
				why = "invalid-value"
			}
			if why != "" {
				f.Excluded[row.OccurrenceKey] = why
				continue
			}
			if len(eligible) == 0 || *row.Value < min {
				min = *row.Value
			}
			if len(eligible) == 0 || *row.Value > max {
				max = *row.Value
			}
			eligible = append(eligible, row)
		}
		f.Eligible = len(eligible)
		positions := map[[2]float64]int{}
		for _, row := range eligible {
			ns, err := strconv.ParseInt(row.EndNS, 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			x := float64(ns-q.Question.FromMS*1000000) / float64((q.Question.ToMS-q.Question.FromMS)*1000000)
			y := 0.5
			if min != max {
				span := max - min
				if math.IsInf(span, 0) {
					y = (*row.Value/2 - min/2) / (max/2 - min/2)
				} else {
					y = (*row.Value - min) / span
				}
			}
			if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) || x < 0 || x > 1 || y < 0 || y > 1 {
				t.Fatal("independent coordinate invalid")
			}
			pair := [2]float64{x, y}
			index, exists := positions[pair]
			if !exists {
				index = len(f.Groups)
				positions[pair] = index
				f.Groups = append(f.Groups, nil)
				f.Coordinates = append(f.Coordinates, pair)
			}
			f.Groups[index] = append(f.Groups[index], gauge455Member{row.OccurrenceKey, row.EndNS, row.RawValue})
		}
		fixtures = append(fixtures, f)
	}
	hegel.Test(t, func(tc *hegel.T) {
		admit(hegel.Draw(tc, hegel.Integers(1, 35)), hegel.Draw(tc, hegel.Integers(0, 1000)), false)
	}, options(50)...)
	admit(1001, 0, true)
	admit(10000, 0, true)
	raw, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	script := `import fs from 'node:fs';import {dataset,model} from '../ui/raw-gauge-points.mjs';globalThis.fetch=()=>{throw Error('held point projection attempted IO')};const results=[];for(const f of JSON.parse(fs.readFileSync(0,'utf8'))){const before=JSON.stringify(f.data),d=dataset(f.data),m=model(d);results.push({model:structuredClone(m),quality:structuredClone(d.quality),rows:d.rows.length});if(d.quality.extent==='complete'||d.completeness!==undefined||d.sourceBasis!==undefined)throw Error('invented completeness');if(d.rows.length){d.rows[0].observation.stream.resource['service.name']='mutated';d.rows[0].values.occurrence='mutated'}if(m.points.length){m.points[0].members[0].rawValue='mutated';m.points[0].occurrenceKeys.push('mutated')}if(JSON.stringify(f.data)!==before)throw Error('bank alias');const bad=structuredClone(f.data);if(bad.rows.length){bad.rows[0].endNS='bad';let refused=false;try{dataset(bad)}catch{refused=true}if(!refused)throw Error('malformed bank admitted')}const invalid=structuredClone(d);invalid.quality.extent='complete';if(model(invalid).state!=='unavailable')throw Error('invalid profile plotted')}process.stdout.write(JSON.stringify(results));`
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JS raw points %v %s", err, out)
	}
	var results []struct {
		Model   gauge455Model
		Quality struct{ Extent, Basis, Value, Sampling, Freshness string }
		Rows    int
	}
	if err := json.Unmarshal(out, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != len(fixtures) {
		t.Fatal("lost bank")
	}
	for i, f := range fixtures {
		r, m := results[i], results[i].Model
		plotted := 0
		expectedGroups := min(1000, len(f.Groups))
		if m.State != "observed" || len(m.Points) != expectedGroups || m.CoordinateGroups != len(f.Groups) || r.Rows != len(f.Data.Rows) {
			t.Fatalf("group/bank counts fixture %d", i)
		}
		partition := map[string]string{}
		for j, point := range m.Points {
			if !reflect.DeepEqual(point.Members, f.Groups[j]) || point.Multiplicity != len(f.Groups[j]) || point.X != f.Coordinates[j][0] || point.Y != f.Coordinates[j][1] {
				t.Fatalf("coordinate/group membership fixture %d group%d", i, j)
			}
			for k, member := range point.Members {
				if partition[member.Key] != "" || point.OccurrenceKeys[k] != member.Key || !bytes.Equal(point.RawValues[k], member.RawValue) {
					t.Fatal("plotted duplicate/lost raw value")
				}
				partition[member.Key] = "plotted"
				plotted++
			}
		}
		expectedExcluded := map[string]string{}
		for key, why := range f.Excluded {
			expectedExcluded[key] = why
		}
		for _, group := range f.Groups[expectedGroups:] {
			for _, member := range group {
				expectedExcluded[member.Key] = "visual-budget"
			}
		}
		for _, item := range m.Excluded {
			if partition[item.Key] != "" || expectedExcluded[item.Key] != item.Reason {
				t.Fatal("partition mismatch")
			}
			partition[item.Key] = item.Reason
			for _, row := range f.Data.Rows {
				if row.OccurrenceKey == item.Key && !bytes.Equal(row.RawValue, item.RawValue) {
					t.Fatal("excluded raw value changed")
				}
			}
		}
		c := m.Counts
		reasons := map[string]int{}
		for _, reason := range f.Excluded {
			reasons[reason]++
		}
		if c.UnknownRecording != reasons["recording-unknown"] || c.NoRecordedValue != reasons["not-recorded"] || c.InvalidValue != reasons["invalid-value"] {
			t.Fatal("exclusion reason counts changed")
		}
		if len(partition) != len(f.Data.Rows) || len(m.Excluded) != len(expectedExcluded) || c.Loaded != len(f.Data.Rows) || c.Eligible != f.Eligible || c.Plotted != plotted || c.Unplotted != len(f.Data.Rows)-plotted || c.NotShownEligible != f.Eligible-plotted || c.Undisplayed != f.Eligible-plotted || c.RenderedLocations != expectedGroups {
			t.Fatalf("partition conservation fixture %d", i)
		}
		wantExtent := "unknown"
		if r.Quality.Extent != wantExtent || r.Quality.Value != "unknown" || r.Quality.Sampling != "unknown" || r.Quality.Freshness != "fixed" || r.Quality.Basis != "captured-query" || m.AtLimit != (len(f.Data.Rows) == 10000) || m.FromNS != strconv.FormatInt(f.Data.Question.FromMS*1000000, 10) || m.ToNS != strconv.FormatInt(f.Data.Question.ToMS*1000000, 10) {
			t.Fatal("coverage/time claim changed")
		}
	}
	t.Logf("%d synthetic Go-source banks: finite/flag partition, exact occurrences/raw values, coordinate collision groups, extremes/subnormals,1000-location budget and10000-source limit", len(fixtures))
}

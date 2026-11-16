package verification

import (
	"bytes"
	"encoding/json"
	"math/big"
	"os/exec"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// Generated tracker-admitted records are synthetic observations, not physical
// ingestion evidence. Supplemental malformed public times test the unplaced rail.
func TestHegelHeldChronology458OrderAndConservation(t *testing.T) {
	type expected struct {
		Key, Signal, NS string
		Index           int
		Placed          bool
	}
	type fixture struct {
		Lanes    map[string]any `json:"lanes"`
		Expected []expected     `json:"-"`
	}
	fixtures := []fixture{}
	admit := func(logs, spans, seed int, public bool) {
		f := fixture{Lanes: map[string]any{}}
		for lane, signal := range []string{"logs", "spans"} {
			size := []int{logs, spans}[lane]
			input := []json.RawMessage{}
			for i := 0; i < size; i++ {
				stamps := []string{"0", "0000", "1", "0000000001", "999999999", "1000000000", "1700000000000000000", "1700000000000000001"}
				stamp := stamps[(i+seed)%len(stamps)]
				// Identical record payload in both independent trackers deliberately gives
				// equal source keys across lanes; duplicate records retain ordinal members.
				record := map[string]any{"timestamp_ns": stamp, "Body": "same log", "SpanName": "same span", "ServiceName": "api", "ResourceAttributes": map[string]string{"service.name": "api"}}
				if i > 1 && i%3 == 0 {
					record["timestamp_ns"] = stamps[(i-1+seed)%len(stamps)]
				}
				raw, err := json.Marshal(record)
				if err != nil {
					t.Fatal(err)
				}
				input = append(input, raw)
			}
			var tracker exploration.TailTracker
			rows, err := tracker.Observe(input)
			if err != nil {
				t.Fatal(err)
			}
			items := []any{}
			for i, row := range rows {
				var record map[string]any
				if err = json.Unmarshal(row.Record, &record); err != nil {
					t.Fatal(err)
				}
				if public {
					switch i % 7 {
					case 0:
						delete(record, "timestamp_ns")
					case 1:
						record["timestamp_ns"] = "bad"
					case 2:
						record["timestamp_ns"] = "999999999999999999999999999999999"
					case 3:
						record["timestamp_ns"] = "-1"
					case 4:
						record["timestamp_ns"] = float64(1)
					}
				}
				stamp, _ := record["timestamp_ns"].(string)
				placed := chronology458Renderable(stamp)
				f.Expected = append(f.Expected, expected{Key: signal + ":" + row.Key, Signal: signal, NS: stamp, Index: i, Placed: placed})
				items = append(items, map[string]any{"key": row.Key, "record": record, "readWindow": map[string]any{"fromMs": int64(1700000000000) + int64(i)*1000, "toMs": int64(1700000060000) + int64(i)*1000}, "observedAt": "2026-11-16T12:00:00.123456789Z", "limited": i%2 == 0, "late": row.Late, "membership": map[string]string{"token": "must not reach passive view"}, "connection": strings.Repeat("b", 64), "size": 123})
			}
			selector := map[string]any{"environment": "test"}
			if signal == "spans" {
				selector["signal"] = "spans"
			}
			f.Lanes[signal] = map[string]any{"selector": selector, "phase": "paused", "staged": false, "find": map[string]any{"text": "", "revision": 0}, "readWindow": map[string]any{"fromMs": 1700000000000, "toMs": 1700000060000}, "completedRead": map[string]any{"fromMs": 1700000000000, "toMs": 1700000060000, "rowCount": len(rows), "limited": false, "observedAt": "2026-11-16T12:00:00.123456789Z", "revision": 1}, "connection": strings.Repeat("b", 64), "evicted": 0, "gap": false, "rows": items}
		}
		slices.SortStableFunc(f.Expected, func(a, b expected) int {
			if a.Placed != b.Placed {
				if a.Placed {
					return -1
				}
				return 1
			}
			if a.Placed {
				x, _ := new(big.Int).SetString(a.NS, 10)
				y, _ := new(big.Int).SetString(b.NS, 10)
				if c := x.Cmp(y); c != 0 {
					return -c
				}
			}
			if a.Signal != b.Signal {
				if a.Signal == "logs" {
					return -1
				}
				return 1
			}
			return a.Index - b.Index
		})
		fixtures = append(fixtures, f)
	}
	hegel.Test(t, func(tc *hegel.T) {
		admit(hegel.Draw(tc, hegel.Integers(0, 30)), hegel.Draw(tc, hegel.Integers(0, 30)), hegel.Draw(tc, hegel.Integers(0, 1000)), false)
	}, options(50)...)
	admit(100, 100, 7, false) // Full legal two-lane buffer, not a sampled prefix.
	admit(20, 20, 3, true)    // Public-profile diagnostics, not tracker-admitted timestamps.
	admit(10, 0, 5, false)
	closed := fixtures[len(fixtures)-1].Lanes["spans"].(map[string]any)
	closed["selector"] = nil
	closed["phase"] = "stopped"
	closed["readWindow"] = nil
	closed["completedRead"] = nil
	closed["connection"] = nil

	raw, err := json.Marshal(fixtures)
	if err != nil {
		t.Fatal(err)
	}
	script := `import fs from 'node:fs';import {prepare,model} from '../ui/held-chronology.mjs';globalThis.fetch=()=>{throw Error('chronology attempted IO')};const out=[];for(const f of JSON.parse(fs.readFileSync(0,'utf8'))){const held=prepare(f.lanes),m=model(held);if(!Object.isFrozen(held.lanes.logs.rows))throw Error('mutable held data');if(JSON.stringify(held).includes('must not reach passive view'))throw Error('membership leaked');out.push({rows:m.rows.map(r=>({key:r.key,signal:r.signal,index:r.index,placed:r.placed,timeNS:r.timeNS})),counts:m.counts,unplaced:m.unplaced.length});const before=JSON.stringify(held);for(const signal of ['logs','spans'])if(f.lanes[signal].rows.length)f.lanes[signal].rows[0].record.Body='mutated';if(JSON.stringify(held)!==before)throw Error('caller alias');try{held.lanes.logs.rows.push({});throw Error('mutable rows')}catch(e){if(e.message==='mutable rows')throw e}}process.stdout.write(JSON.stringify(out));`
	cmd := exec.Command("node", "--input-type=module", "-e", script)
	cmd.Stdin = bytes.NewReader(raw)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("JS chronology: %v %s", err, out)
	}
	var results []struct {
		Rows []struct {
			Key, Signal string
			Index       int
			Placed      bool
			TimeNS      *string
		}
		Counts   map[string]int
		Unplaced int
	}
	if err = json.Unmarshal(out, &results); err != nil {
		t.Fatal(err)
	}
	if len(results) != len(fixtures) {
		t.Fatal("lost fixture")
	}
	for i, f := range fixtures {
		got := results[i]
		if len(got.Rows) != len(f.Expected) {
			t.Fatalf("fixture%d lost occurrence", i)
		}
		counts := map[string]int{"logs": 0, "spans": 0}
		unplaced := 0
		keys := map[string]bool{}
		for j, want := range f.Expected {
			row := got.Rows[j]
			if row.Key != want.Key || row.Signal != want.Signal || row.Index != want.Index || row.Placed != want.Placed || row.Placed && (row.TimeNS == nil || *row.TimeNS != want.NS) || !row.Placed && row.TimeNS != nil {
				t.Fatalf("fixture%d row%d got%+v want%+v", i, j, row, want)
			}
			if keys[row.Key] {
				t.Fatal("cross-lane identity collision")
			}
			keys[row.Key] = true
			counts[row.Signal]++
			if !row.Placed {
				unplaced++
			}
		}
		if !reflect.DeepEqual(got.Counts, counts) || got.Unplaced != unplaced {
			t.Fatal("lane/unplaced counts changed")
		}
	}
	t.Logf("%d synthetic banks: 51 Go-tracker-admitted, one public-time diagnostic and one closed-empty-lane bank; exact event ordering, stable ties, duplicate/lane membership, 200-row boundary, unplaced conservation, immutable snapshots and no IO", len(fixtures))
}

var chronology458Stamp = regexp.MustCompile(`^-?\d+$`)

func chronology458Renderable(stamp string) bool {
	if len(stamp) == 0 || len(stamp) > 32 || !chronology458Stamp.MatchString(stamp) {
		return false
	}
	ns, ok := new(big.Int).SetString(stamp, 10)
	if !ok {
		return false
	}
	limit, _ := new(big.Int).SetString("8640000000000000000000", 10)
	return new(big.Int).Abs(ns).Cmp(limit) <= 0
}

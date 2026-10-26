package verification

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"testing"
)

type psMap = map[string]any

func psObj(v any) psMap { m, _ := v.(map[string]any); return m }
func psArr(v any) []any { a, _ := v.([]any); return a }
func psEq(a, b any) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}
func psSelected(f psMap, q psMap, hist bool) []psMap {
	d := psObj(q["definition"])
	s := psObj(psObj(d["source"])["exactStream"])
	res := psObj(s["resource"])
	kind := "points"
	if hist {
		kind = "histograms"
	}
	out := []psMap{}
	for _, v := range psArr(f[kind]) {
		p := psObj(v)
		resource := psMap{"deployment.environment.name": f["environment"], "service.name": f["service"], "service.instance.id": p["instance"], "service.namespace": p["namespace"], "telemetry.sdk.language": "python"}
		for k, v := range psObj(p["resourceExtra"]) {
			resource[k] = v
		}
		if !psEq(resource, res) || p["scope"] != s["scope"] || p["version"] != s["version"] {
			continue
		}

		if p["metric"] != s["name"] || p["unit"] != s["unit"] || p["schemaURL"] != s["schemaURL"] || p["instance"] != res["service.instance.id"] || p["namespace"] != res["service.namespace"] || !psEq(p["attributes"], s["attributes"]) {
			continue
		}
		out = append(out, p)
	}
	return out
}
func psMeasurement(f, r psMap) bool {
	q := psObj(r["question"])
	s := psObj(psObj(psObj(q["definition"])["source"])["exactStream"])
	hist := psObj(q["definition"])["kind"] == "histogram"
	pts := psSelected(f, q, hist)
	rows := psArr(r["rows"])
	if len(pts) == 0 || len(rows) != len(pts) {
		return false
	}
	sort.Slice(pts, func(i, j int) bool { return pts[i]["endNS"].(string) < pts[j]["endNS"].(string) })
	for n, v := range rows {
		row := psObj(v)
		p := pts[n]
		if row["name"] != p["metric"] || row["unit"] != p["unit"] || row["start_ns"] != p["startNS"] || row["end_ns"] != p["endNS"] || row["temporality"] != p["temporality"] || !psEq(row["resource"], s["resource"]) || !psEq(row["attributes"], p["attributes"]) {
			return false
		}
		if hist {
			if !psEq(row["bounds"], p["bounds"]) || !psEq(row["bucket_counts"], p["bucketCounts"]) || row["count"] != p["count"] {
				return false
			}
		} else if row["value"] != p["value"] {
			return false
		}
	}
	if !hist {
		if psObj(q["definition"])["stockDomain"] != "nonnegative-integer" {
			return false
		}
		series := psArr(psObj(r["stocks"])["series"])
		if len(series) != 1 || !psEq(psObj(series[0])["stream"], s) {
			return false
		}
		samples := psArr(psObj(series[0])["samples"])
		if len(samples) != len(pts) {
			return false
		}
		for n, v := range samples {
			a := psObj(v)
			val, e := strconv.ParseFloat(pts[n]["value"].(string), 64)
			if e != nil || a["value"] != val || a["atNS"] != pts[n]["endNS"] || a["startNS"] != pts[n]["startNS"] || a["state"] != "observed" {
				return false
			}
		}
		return true
	}
	series := psArr(psObj(r["histograms"])["series"])
	if len(series) != 1 || !psEq(psObj(series[0])["stream"], s) {
		return false
	}
	intervals := psArr(psObj(series[0])["intervals"])
	if len(intervals) != len(pts) {
		return false
	}
	for n, v := range intervals {
		a := psObj(v)
		d := psObj(a["distribution"])
		p := pts[n]
		if a["fromNS"] != p["startNS"] || a["toNS"] != p["endNS"] || !psEq(d["bounds"], p["bounds"]) || !psEq(d["bucketCounts"], p["bucketCounts"]) || d["count"] != p["count"] {
			return false
		}
		count, _ := strconv.Atoi(p["count"].(string))
		target := (95*count + 99) / 100
		sum := 0
		idx := -1
		for j, c := range psArr(p["bucketCounts"]) {
			k, _ := strconv.Atoi(c.(string))
			sum += k
			if idx < 0 && sum >= target {
				idx = j
			}
		}
		bounds := psArr(p["bounds"])
		if idx <= 0 || idx >= len(bounds) || psObj(d["p95"])["lower"] != bounds[idx-1] || psObj(d["p95"])["upper"] != bounds[idx] {
			return false
		}
	}
	return true
}
func psObserved(d psMap) bool {
	f := psObj(d["facts"])
	pair := psObj(d["paired"])
	req := psObj(pair["request"])
	parent := psObj(req["parent"])
	stream := psObj(req["stream"])
	attrs := psObj(stream["attributes"])
	expected := psMap{}
	for k, v := range attrs {
		if k != "db.client.connection.state" {
			expected[k] = v
		}
	}
	for _, v := range psArr(d["results"]) {
		r := psObj(v)
		actual := psObj(r["result"])
		q := psObj(actual["question"])
		if !psEq(q, psObj(r["option"])["question"]) || !psMeasurement(f, actual) || !psEq(r["saved"], r["reopened"]) {
			return false
		}
		s := psObj(psObj(psObj(q["definition"])["source"])["exactStream"])
		if !psEq(s["attributes"], expected) || !psEq(s["resource"], stream["resource"]) {
			return false
		}
		for _, k := range []string{"fromMs", "toMs", "fromNS", "toNS"} {
			if q[k] != parent[k] {
				return false
			}
		}
	}
	events := psArr(d["events"])
	sources := []float64{9, 9, 9, 10, 11, 11, 11, 11, 12, 13, 13, 13, 13, 13, 13}
	if len(events) != len(sources) {
		return false
	}
	for n, v := range events {
		if psObj(v)["source"] != sources[n] {
			return false
		}
	}
	return true
}
func TestPoolSignalsObserved(t *testing.T) {
	path := os.Getenv("GOO_POOL_SIGNALS_LEDGER")
	if path == "" {
		t.Skip("set GOO_POOL_SIGNALS_LEDGER to the actual journey197 ledger")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	if json.Unmarshal(b, &d) != nil || !psObserved(d) {
		t.Fatal("raw membership, independent values/buckets, exact scope or source discipline")
	}
	for _, v := range psArr(d["results"]) {
		r := psObj(v)
		a := psObj(r["storedBefore"])
		z := psObj(r["storedAfter"])
		before, e := os.ReadFile(filepath.Join(filepath.Dir(path), a["filename"].(string)))
		if e != nil {
			t.Fatal(e)
		}
		after, e := os.ReadFile(filepath.Join(filepath.Dir(path), z["filename"].(string)))
		sum := sha256.Sum256(before)
		var stored exploration.SavedCapture
		if json.Unmarshal(before, &stored) != nil || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) {
			t.Fatal("complete stored content-addressed receipt")
		}
		raw, _ := json.Marshal(stored.Measurement)
		var measured psMap
		_ = json.Unmarshal(raw, &measured)
		if !psMeasurement(psObj(d["facts"]), measured) || !psEq(measured["question"], psObj(r["result"])["question"]) {
			t.Fatal("saved raw source or exact question")
		}

		if e != nil || string(before) != string(after) || hex.EncodeToString(sum[:]) != a["sha256"] || a["sha256"] != z["sha256"] || float64(len(before)) != a["bytes"] {
			t.Fatal("full custody bytes")
		}
	}
	controls := []func(psMap){func(x psMap) { psObj(psObj(psObj(psArr(x["results"])[0])["result"])["question"])["toMs"] = float64(1) }, func(x psMap) { psObj(psArr(psObj(psObj(psArr(x["results"])[0])["result"])["rows"])[0])["value"] = "99" }, func(x psMap) {
		psObj(psObj(psObj(psObj(psArr(x["results"])[0])["result"])["question"])["definition"])["stockDomain"] = ""
	}, func(x psMap) { psObj(psArr(x["events"])[1])["source"] = float64(10) }, func(x psMap) { psObj(psObj(psArr(x["results"])[1])["reopened"])["changed"] = true }}
	for n, change := range controls {
		var x psMap
		_ = json.Unmarshal(b, &x)
		change(x)
		if reflect.DeepEqual(x, d) || psObserved(x) {
			t.Fatalf("changed control %d accepted", n)
		}
	}
}
func TestPoolSignalsOwnershipObserved(t *testing.T) {
	path := os.Getenv("GOO_POOL_SIGNALS_OWNER_LEDGER")
	if path == "" {
		t.Skip("set GOO_POOL_SIGNALS_OWNER_LEDGER")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var d psMap
	_ = json.Unmarshal(b, &d)
	check := func(x psMap) bool {
		a, z := psObj(x["before"]), psObj(x["after"])
		held := psObj(x["held"])
		return a["source"] == float64(9) && psEq(a, z) && len(psArr(psObj(held["response"])["capacity"])) > 0 && len(psArr(psObj(held["response"])["wait"])) > 0
	}
	if !check(d) {
		t.Fatal("actual held positive plan was not ignored after navigation")
	}
	for _, k := range []string{"source", "optionCount", "url"} {
		var x psMap
		_ = json.Unmarshal(b, &x)
		psObj(x["after"])[k] = "changed"
		if check(x) {
			t.Fatal("owner changed control", k)
		}
	}
}

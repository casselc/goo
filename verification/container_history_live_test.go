package verification

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/qtrace"
	"github.com/casselc/observability/quintgo/validate"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/measurements"
)

func chBasis(r *measurements.Result) bool {
	if r == nil {
		return false
	}
	b, e := json.Marshal(struct {
		Q    measurements.Question
		At   string
		Rows []json.RawMessage
	}{r.Question, r.ReadAt, r.Rows})
	if e != nil {
		return false
	}
	hash := sha256.Sum256(b)
	return hex.EncodeToString(hash[:]) == r.Basis
}

// chHistoryOK derives timestamp groups and all point classifications from the
// original physical rows. The raw adapter separately binds these rows to OTLP
// facts; no production composer or gauge sample is used as the expected result.
func chHistoryOK(f psMap, input measurements.ContainerResourceHistoryRequest, got measurements.ContainerResourceHistory) bool {
	if got.Version != measurements.ContainerResourceHistoryVersion || !aeJSONEqual(input.Parent, got.Parent) || len(got.Warnings) == 0 {
		return false
	}
	type role struct {
		family, name, unit string
		result             *measurements.Result
	}
	roles := []role{}
	if input.Memory != nil {
		roles = append(roles, role{"memory", "working-set", "By", input.Memory.WorkingSet}, role{"memory", "request", "By", input.Memory.Request}, role{"memory", "limit", "By", input.Memory.Limit})
	}
	if input.CPU != nil {
		roles = append(roles, role{"cpu", "usage", "{cpu}", input.CPU.Usage}, role{"cpu", "request", "{cpu}", input.CPU.Request}, role{"cpu", "limit", "{cpu}", input.CPU.Limit})
	}
	if len(roles) == 0 || len(got.Series) != len(roles) {
		return false
	}
	total := 0
	for i, role := range roles {
		s := got.Series[i]
		r := role.result
		if s.Key != role.family+"/"+role.name || s.Family != role.family || s.Role != role.name || s.Unit != role.unit || s.Title == "" || s.Points == nil {
			return false
		}
		if r == nil {
			if s.State != "unavailable" || s.Reason != "not-supplied" || len(s.Points) != 0 || s.Question != nil || s.Basis != "" || s.ReadAt != "" {
				return false
			}
			continue
		}
		if !chBasis(r) || !rsRaw(f, r) || s.Question == nil || !aeJSONEqual(*s.Question, r.Question) || s.Basis != r.Basis || s.ReadAt != r.ReadAt {
			return false
		}
		actualRole := cmRole(*r.Question.Definition.Source.ExactStream)
		if role.family == "cpu" {
			actualRole = cpRole(*r.Question.Definition.Source.ExactStream)
		}
		if actualRole != role.name || !cmContext(input.Parent.Resource, r.Question.Definition.Source.ExactStream.Resource) || r.Question.FromMS != input.Parent.FromMS || r.Question.ToMS != input.Parent.ToMS || r.Question.Definition.Environment != input.Parent.Environment {
			return false
		}

		from, to := r.Question.BoundsNS()
		if s.FromNS != from || s.ToNS != to || s.MaxGapNS != strconv.FormatInt(r.Question.Definition.MaxGapMS*1000000, 10) {
			return false
		}
		end, ok := new(big.Int).SetString(to, 10)
		if !ok {
			return false
		}
		gap, ok := new(big.Int).SetString(s.MaxGapNS, 10)
		if !ok {
			return false
		}
		type row struct {
			At    string `json:"end_ns"`
			Value string `json:"value"`
			Flags uint32 `json:"flags"`
		}
		rows := make([]row, len(r.Rows))
		groups := map[string][]int{}
		for j, b := range r.Rows {
			if json.Unmarshal(b, &rows[j]) != nil {
				return false
			}
			groups[rows[j].At] = append(groups[rows[j].At], j)
		}
		times := []string{}
		for at := range groups {
			times = append(times, at)
		}
		sort.Slice(times, func(a, b int) bool {
			x, ok := new(big.Int).SetString(times[a], 10)
			y, ok2 := new(big.Int).SetString(times[b], 10)
			return ok && ok2 && x.Cmp(y) < 0
		})
		total += len(times)
		if total > 10000 || len(s.Points) != len(times) {
			return false
		}
		if len(times) == 0 {
			if s.State != "no_data" || s.Reason != "no-observations" {
				return false
			}
			continue
		}
		if s.State != "observations" || s.Reason != "" {
			return false
		}
		for j, at := range times {
			indices := groups[at]
			p := s.Points[j]
			if p.AtNS != at || !aeJSONEqual(p.RawRowIndices, indices) {
				return false
			}
			first := rows[indices[0]]
			v, e := strconv.ParseFloat(first.Value, 64)
			if e != nil {
				return false
			}
			conflict := false
			for _, idx := range indices[1:] {
				next := rows[idx]
				n, e := strconv.ParseFloat(next.Value, 64)
				if e != nil {
					return false
				}
				conflict = conflict || n != v || (next.Flags&1) != (first.Flags&1)
			}
			quality, reason := "observed", ""
			timestamp, ok := new(big.Int).SetString(at, 10)
			if !ok {
				return false
			}
			switch {
			case conflict:
				quality, reason = "unknown", "conflicting observations at one timestamp"
			case first.Flags&1 != 0:
				quality, reason = "unknown", "no recorded value"
			case new(big.Int).Sub(end, timestamp).Cmp(gap) > 0:
				quality, reason = "stale", "older than declared freshness"
			}
			state, value, domainReason := quality, "", reason
			if quality != "unknown" {
				var exact *big.Rat
				for _, idx := range indices {
					raw := rows[idx].Value
					var n *big.Rat
					canonical := ""
					if role.family == "memory" {
						if integer := cmSafeBytes(raw); integer != nil {
							n = new(big.Rat).SetInt(integer)
							canonical = integer.String()
						}
					} else {
						var valid bool
						n, canonical, valid = cpDecimal(raw)
						if !valid {
							n = nil
						}
					}
					if n == nil || exact != nil && exact.Cmp(n) != 0 {
						state = "invalid"
						value = ""
						break
					}
					exact = n
					value = canonical
				}
				if state == "invalid" {
					domainReason = "invalid-byte-domain"
					if role.family == "cpu" {
						domainReason = "invalid-core-domain"
					}
				}
			}
			if p.State != state || p.Value != value || p.Reason != domainReason || p.ObservationState != quality || p.ObservationReason != reason {
				return false
			}
		}
	}
	return true
}

func chScopeLive(q exploration.Query, resource map[string]string, lo, hi int64, env, signal string) bool {
	if q.Environment != env {
		return false
	}
	q.Environment = "history-property"
	return chScope(q, resource, lo, hi, signal)
}
func chContext(f psMap, resource map[string]string, v psMap) bool {
	var q exploration.Query
	var result exploration.Result
	if !rsDecode(v["question"], &q) || !rsDecode(v["result"], &result) || q.Signal != fmt.Sprint(v["signal"]) || !nhLiveSameQuery(q, result.Question) {
		return false
	}
	r := psObj(v["range"])
	lo, e := strconv.ParseInt(fmt.Sprint(r["fromNS"]), 10, 64)
	if e != nil {
		return false
	}
	hi, e := strconv.ParseInt(fmt.Sprint(r["toNS"]), 10, 64)
	if e != nil || !chScopeLive(q, resource, lo, hi, fmt.Sprint(f["environment"]), q.Signal) {
		return false
	}
	if _, e = exploration.Compile(q); e != nil {
		return false
	}
	expected := map[string]psMap{}
	keys := []string{"container.id", "k8s.pod.uid", "k8s.container.name", "k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name"}
	for _, v := range psArr(f["records"]) {
		p := psObj(v)
		if p["family"] != q.Signal {
			continue
		}
		at, e := strconv.ParseInt(fmt.Sprint(p["timeNS"]), 10, 64)
		if e != nil {
			return false
		}
		match := at >= lo && at < hi
		raw := psObj(p["resource"])
		for _, k := range keys {
			a, present := resource[k]
			b, has := raw[k]
			match = match && present == has && (!present || a == b)
		}
		if match {
			expected[fmt.Sprint(p["id"])] = p
		}
	}
	ids := []string{}
	for id := range expected {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	if len(expected) != 3 || result.Count != strconv.Itoa(len(expected)) || len(result.Rows) != len(expected) || !aeJSONEqual(ids, v["expectedMembers"]) {
		return false
	}
	for _, row := range result.Rows {
		b, _ := json.Marshal(row)
		var p psMap
		if json.Unmarshal(b, &p) != nil {
			return false
		}
		attrs := psObj(p["Attributes"])
		id := fmt.Sprint(attrs["work.id"])
		fact, ok := expected[id]
		if !ok || !aeJSONEqual(p["ResourceAttributes"], fact["resource"]) || !aeJSONEqual(attrs, fact["attributes"]) || p["TraceId"] != fact["traceId"] || p["timestamp_ns"] != fact["timeNS"] {
			return false
		}
		delete(expected, id)
	}
	if _, e = time.Parse(time.RFC3339Nano, result.CapturedAt); e != nil {
		return false
	}
	basis, e := hex.DecodeString(result.Basis)
	return e == nil && len(basis) == 32 && len(expected) == 0
}
func chPair(f psMap, v psMap, family string, origin exploration.PodDiscoveryQuestion, responses []any) bool {
	if !rsPair(f, v, family, origin) || !rsNetwork(responses, "/api/container-"+family+"/view", v["response"]) {
		return false
	}
	for _, key := range []string{"workingSet", "usage", "request", "limit"} {
		if r := psObj(v["request"])[key]; r != nil && !rsNetwork(responses, "/api/measurements", r) {
			return false
		}
	}
	history := psObj(v["history"])
	var req measurements.ContainerResourceHistoryRequest
	var response measurements.ContainerResourceHistory
	if !rsDecode(history["request"], &req) || !rsDecode(history["response"], &response) || !chHistoryOK(f, req, response) || !rsNetwork(responses, "/api/container-resources/history", history["response"]) {
		return false
	}
	if !aeJSONEqual(req.Parent, psObj(v["request"])["parent"]) {
		return false
	}
	// History must project these exact already-read roles; it cannot swap another
	// equally valid read, clock, source variant or family into the evidence.
	if family == "memory" {
		if req.CPU != nil || req.Memory == nil {
			return false
		}
		return aeJSONEqual(req.Memory.WorkingSet, psObj(v["request"])["workingSet"]) && aeJSONEqual(req.Memory.Request, psObj(v["request"])["request"]) && aeJSONEqual(req.Memory.Limit, psObj(v["request"])["limit"])
	}
	if req.Memory != nil || req.CPU == nil {
		return false
	}
	return aeJSONEqual(req.CPU.Usage, psObj(v["request"])["usage"]) && aeJSONEqual(req.CPU.Request, psObj(v["request"])["request"]) && aeJSONEqual(req.CPU.Limit, psObj(v["request"])["limit"])
}

func chSnapshot(history measurements.ContainerResourceHistory, family string, snap psMap, focus string, offset int) bool {
	type point struct {
		p           measurements.ContainerResourceHistoryPoint
		role, key   string
		index, rank int
	}
	all := []point{}
	maximum := new(big.Rat)
	for rank, s := range history.Series {
		if s.Family != family {
			continue
		}
		for index, p := range s.Points {
			all = append(all, point{p, s.Role, s.Key + "@" + p.AtNS, index, rank})
			if p.Value != "" {
				value, ok := new(big.Rat).SetString(p.Value)
				if !ok {
					return false
				}
				if value.Cmp(maximum) > 0 {
					maximum = value
				}
			}
		}
	}
	sort.SliceStable(all, func(i, j int) bool {
		a, _ := new(big.Int).SetString(all[i].p.AtNS, 10)
		b, _ := new(big.Int).SetString(all[j].p.AtNS, 10)
		if c := a.Cmp(b); c != 0 {
			return c < 0
		}
		return all[i].rank < all[j].rank
	})
	filtered := []point{}
	for _, p := range all {
		if focus == "all" || focus == p.role {
			filtered = append(filtered, p)
		}
	}
	plot := psObj(snap["plot"])
	domain := psObj(plot["domain"])
	max, ok := new(big.Rat).SetString(fmt.Sprint(domain["max"]))
	if !ok || max.Cmp(maximum) != 0 || domain["min"] != "0" || plot["fromNS"] != strconv.FormatInt(history.Parent.FromMS*1000000, 10) || plot["toNS"] != strconv.FormatInt(history.Parent.ToMS*1000000, 10) || snap["focus"] != focus || plot["offset"] != float64(offset) || plot["total"] != float64(len(filtered)) || plot["allTotal"] != float64(len(all)) {
		return false
	}
	stop := offset + 256
	if stop > len(filtered) {
		stop = len(filtered)
	}
	if offset > stop {
		return false
	}
	points := psArr(plot["points"])
	marks := psArr(snap["marks"])
	if len(points) != stop-offset || len(marks) != len(points) {
		return false
	}
	for i, v := range points {
		actual := psObj(v)
		want := filtered[offset+i]
		state := want.p.State
		if state == "invalid" {
			state = "unknown"
		}
		if actual["key"] != want.key || actual["fromNS"] != want.p.AtNS || actual["toNS"] != want.p.AtNS || actual["state"] != state || actual["series"] != want.role || actual["role"] != want.role || actual["index"] != float64(want.index) {
			return false
		}
		value, present := actual["value"]
		if present != (want.p.Value != "") || present && value != want.p.Value {
			return false
		}
		mark := psObj(marks[i])
		if mark["key"] != want.key || mark["series"] != want.role {
			return false
		}
	}
	return true
}
func chMainOK(d psMap) bool {
	f := psObj(d["facts"])
	initial := psObj(d["initial"])
	origin, ok := rsOrigin(initial["request"])
	var inventory exploration.PodDiscovery
	if !ok || !rsDecode(initial["response"], &inventory) || !pdOracle(pdObservedFacts(f, origin), &inventory) {
		return false
	}
	responses := psArr(d["responses"])
	for _, key := range []string{"opened", "second"} {
		var read pnRead
		if !rsDecode(d[key], &read) || !ppMetadata(f, read) || !rsNetwork(responses, "/api/pod-metric-discovery", read.Discovery) || !rsNetwork(responses, "/api/pod-dashboard/plan", read.Plan) {
			return false
		}
	}
	for _, family := range []string{"memory", "cpu"} {
		if !chPair(f, psObj(d[family]), family, origin, responses) {
			return false
		}
	}
	var h measurements.ContainerResourceHistory
	if !rsDecode(psObj(psObj(d["cpu"])["history"])["response"], &h) {
		return false
	}
	if !chSnapshot(h, "cpu", psObj(d["initialCPU"]), "all", 0) || !chSnapshot(h, "cpu", psObj(d["usageFirst"]), "usage", 0) || !chSnapshot(h, "cpu", psObj(d["usageLater"]), "usage", 256) || !chSnapshot(h, "cpu", psObj(d["selected"]), "request", 0) {
		return false
	}
	selected := psObj(d["selected"])
	var request measurements.ContainerResourceHistorySeries
	for _, s := range h.Series {
		if s.Role == "request" {
			request = s
		}
	}
	if len(request.Points) == 0 {
		return false
	}
	point := request.Points[len(request.Points)-1]
	if !aeJSONEqual(point, d["selectedPoint"]) || selected["selectedKey"] != request.Key+"@"+point.AtNS {
		return false
	}
	indices := []int{}
	hasBasis, hasClock := false, false
	for _, list := range psArr(selected["statLists"]) {
		for _, v := range psArr(list) {
			p := psObj(v)
			if p["label"] == "Raw row index" {
				index, e := strconv.Atoi(fmt.Sprint(p["value"]))
				if e != nil {
					return false
				}
				indices = append(indices, index)
			}
			hasBasis = hasBasis || p["label"] == "Evidence basis" && p["value"] == request.Basis
			hasClock = hasClock || p["label"] == "Source read" && p["value"] == request.ReadAt
		}
	}
	if !aeJSONEqual(indices, point.RawRowIndices) || !hasBasis || !hasClock {
		return false
	}
	var primary exploration.PodResourceObservation
	if !rsDecode(d["primary"], &primary) {
		return false
	}
	found := false
	for _, r := range inventory.Observations {
		if aeJSONEqual(r, primary) {
			found = true
		}
	}
	if !found {
		return false
	}
	var opened pnRead
	if !rsDecode(d["opened"], &opened) || !aeJSONEqual(opened.DiscoveryRequest.Resource, primary.Resource) || !aeJSONEqual(opened.DiscoveryRequest.ResourceSchemaURL, primary.ResourceSchemaURL) {
		return false
	}
	for _, family := range []string{"memory", "cpu"} {
		var parent measurements.PodMetricDiscoveryQuestion
		if !rsDecode(psObj(psObj(d[family])["request"])["parent"], &parent) || !cmContext(primary.Resource, parent.Resource) {
			return false
		}
	}

	contexts := psArr(d["context"])
	if len(contexts) != 2 {
		return false
	}
	seen := map[string]bool{}
	for _, v := range contexts {
		c := psObj(v)
		signal, ok := c["signal"].(string)
		if !ok || signal != "logs" && signal != "spans" || seen[signal] || !chContext(f, primary.Resource, c) || !rsNetwork(responses, "/api/explore", c["result"]) {
			return false
		}
		seen[signal] = true
	}
	var brush exploration.Query
	if !rsDecode(d["brushQuestion"], &brush) {
		return false
	}
	r := psObj(d["brushRange"])
	lo, e := strconv.ParseInt(fmt.Sprint(r["fromNS"]), 10, 64)
	if e != nil {
		return false
	}
	hi, e := strconv.ParseInt(fmt.Sprint(r["toNS"]), 10, 64)
	if e != nil || !chScopeLive(brush, primary.Resource, lo, hi, origin.Environment, "logs") {
		return false
	}
	var individual measurements.Result
	if !rsDecode(d["individual"], &individual) || !rsRaw(f, &individual) || !aeJSONEqual(individual.Question, request.Question) || !rsNetwork(responses, "/api/measurements", d["individual"]) || !aeJSONEqual(d["saved"], d["reopened"]) || len(psArr(d["blocked"])) != 0 {
		return false
	}
	actions := []string{"shared-metadata", "both-held-histories", "local-focus-page-and-point", "pointer-range-local", "keyboard-form-range-local", "logs-run-retained-back", "spans-run-retained-back", "original-native-stage", "original-native-capture", "frozen-reopen"}
	source := []float64{3, 9, 9, 9, 9, 10, 11, 11, 13, 13}
	plans := []float64{2, 2, 2, 2, 2, 2, 2, 3, 3, 3}
	histories := []float64{0, 2, 2, 2, 2, 2, 2, 2, 2, 2}
	events := psArr(d["events"])
	if len(events) != len(actions) {
		return false
	}
	for i, v := range events {
		p := psObj(v)
		compose := float64(2)
		if i == 0 {
			compose = 0
		}
		if p["action"] != actions[i] || p["source"] != source[i] || p["plans"] != plans[i] || p["histories"] != histories[i] || p["compositions"] != compose {
			return false
		}
	}
	counts := map[string]int{}
	effects := map[string]int{}
	for _, v := range psArr(d["requests"]) {
		p := psObj(v)
		endpoint, _ := p["endpoint"].(string)
		if effect, ok := p["effect"].(string); ok {
			effects[effect]++
		}
		if p["method"] == "POST" {
			counts[endpoint]++
		}
	}
	return effects["source"] == 13 && effects["plan"] == 3 && effects["composition"] == 2 && effects["history"] == 2 && counts["/api/pod-discovery"] == 1 && counts["/api/pod-metric-discovery"] == 2 && counts["/api/measurements"] == 7 && counts["/api/explore"] == 2 && counts["/api/captures"] == 1 && counts["/api/container-resources/history"] == 2
}

func chOwnerOK(d psMap) bool {
	f := psObj(d["facts"])
	cases := psArr(d["cases"])
	reasons := []string{"role", "disclosure", "scope", "close", "sibling-load-survives"}
	if len(cases) != 5 {
		return false
	}
	for i, v := range cases {
		c := psObj(v)
		if c["reason"] != reasons[i] {
			return false
		}
		held := psObj(c["held"])
		var req measurements.ContainerResourceHistoryRequest
		var response measurements.ContainerResourceHistory
		if held["status"] != float64(200) || !rsDecode(held["request"], &req) || !rsDecode(held["body"], &response) || req.Memory == nil || req.CPU != nil || !chHistoryOK(f, req, response) {
			return false
		}
		initial := psObj(c["initial"])
		origin, ok := rsOrigin(initial["request"])
		if !ok || req.Parent.FromMS != origin.FromMS || req.Parent.ToMS != origin.ToMS || req.Parent.Environment != origin.Environment {
			return false
		}
		before, after := psObj(c["before"]), psObj(c["after"])
		heldCounts, afterCounts := psObj(c["countsHeld"]), psObj(c["countsAfter"])
		if heldCounts["source"] != float64((i+1)*3) || heldCounts["plans"] != float64(i+1) || heldCounts["compositions"] != float64(i+1) || heldCounts["histories"] != float64(i+1) {
			return false
		}
		if i < 4 {
			if !aeJSONEqual(before, after) || afterCounts["source"] != heldCounts["source"] || afterCounts["plans"] != heldCounts["plans"] || afterCounts["compositions"] != heldCounts["compositions"] || afterCounts["histories"] != heldCounts["histories"] {
				return false
			}
			if memory := rsFamily(after, "memory"); memory != nil && strings.Contains(fmt.Sprint(memory["html"]), "container-resource-history") {
				return false
			}
		} else {
			if !chPair(f, psObj(c["sibling"]), "cpu", origin, psArr(d["responses"])) || afterCounts["source"] != heldCounts["source"].(float64)+1 || afterCounts["plans"] != heldCounts["plans"] || afterCounts["compositions"] != heldCounts["compositions"].(float64)+1 || afterCounts["histories"] != heldCounts["histories"].(float64)+1 {
				return false
			}
			for _, family := range []string{"memory", "cpu"} {
				s := rsFamily(after, family)
				if s == nil || s["pending"] != false || s["completed"] != true || !strings.Contains(fmt.Sprint(s["html"]), "container-resource-history") {
					return false
				}
			}
		}
	}
	effects := map[string]int{}
	for _, v := range psArr(d["requests"]) {
		p := psObj(v)
		if effect, ok := p["effect"].(string); ok {
			effects[effect]++
		}
	}
	return effects["source"] == 16 && effects["plan"] == 5 && effects["composition"] == 6 && effects["history"] == 6
}
func TestContainerHistoryObserved(t *testing.T) {
	d, raw, path := rsRead(t, "GOO_CONTAINER_HISTORY_LEDGER")
	if !chMainOK(d) {
		t.Fatal("history raw projection/context/clock/effects oracle")
	}
	before, after := psObj(d["storedBefore"]), psObj(d["storedAfter"])
	a, e := os.ReadFile(filepath.Join(filepath.Dir(path), fmt.Sprint(before["file"])))
	if e != nil {
		t.Fatal(e)
	}
	z, e := os.ReadFile(filepath.Join(filepath.Dir(path), fmt.Sprint(after["file"])))
	sum := sha256.Sum256(a)
	var stored exploration.SavedCapture
	var page exploration.CapturePage
	if e != nil || !bytes.Equal(a, z) || len(a) != int(before["bytes"].(float64)) || hex.EncodeToString(sum[:]) != before["sha256"] || before["sha256"] != after["sha256"] || json.Unmarshal(a, &stored) != nil || !rsDecode(d["saved"], &page) || stored.ID != page.Receipt.ID || !reflect.DeepEqual(stored.CaptureReceipt, page.Receipt) || !peLiveCaptureHash(exploration.CapturePage{Receipt: stored.CaptureReceipt, Measurement: stored.Measurement}) || !rsRaw(psObj(d["facts"]), stored.Measurement) || !aeJSONEqual(stored.Measurement.Question, psObj(d["individual"])["question"]) {
		t.Fatal("complete immutable individual capture")
	}
	historyPoint := func(x psMap) psMap {
		return psObj(psArr(psObj(psArr(psObj(psObj(psObj(x["cpu"])["history"])["response"])["series"])[0])["points"])[0])
	}
	changes := []func(psMap){
		func(x psMap) { historyPoint(x)["value"] = "999" },
		func(x psMap) { historyPoint(x)["rawRowIndices"] = []any{float64(999)} },
		func(x psMap) { historyPoint(x)["observationState"] = "unknown" },
		func(x psMap) {
			psObj(psArr(psObj(psObj(psObj(x["cpu"])["history"])["response"])["series"])[0])["basis"] = "wrong"
		},
		func(x psMap) {
			psObj(psArr(psObj(psObj(psObj(x["cpu"])["history"])["response"])["series"])[0])["readAt"] = "wrong"
		},
		func(x psMap) { psObj(psObj(x["usageLater"])["plot"])["offset"] = float64(0) },
		func(x psMap) { psObj(psObj(psObj(x["usageFirst"])["plot"])["domain"])["max"] = "999" },
		func(x psMap) { x["context"] = []any{psArr(x["context"])[0], psArr(x["context"])[0]} },
		func(x psMap) { psObj(psObj(psArr(x["context"])[0])["result"])["count"] = "4" },
		func(x psMap) { psObj(psArr(x["events"])[2])["source"] = float64(10) },
		func(x psMap) { psObj(psArr(x["events"])[2])["histories"] = float64(3) },
		func(x psMap) { psObj(psObj(x["reopened"])["receipt"])["id"] = "wrong" },
		func(x psMap) {
			a, b := psObj(psArr(x["context"])[0]), psObj(psArr(x["context"])[1])
			b["question"] = a["question"]
			b["result"] = a["result"]
			b["expectedMembers"] = a["expectedMembers"]
		},
	}
	for i, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		// Keep the duplicated transport copy coherent for raw-history controls:
		// rejection must come from the independent row/provenance oracle.
		if i < 5 {
			for _, v := range psArr(x["responses"]) {
				r := psObj(v)
				body := psObj(r["body"])
				series := psArr(body["series"])
				if r["endpoint"] == "/api/container-resources/history" && len(series) > 0 && psObj(series[0])["family"] == "cpu" {
					r["body"] = psObj(psObj(x["cpu"])["history"])["response"]
				}
			}
		}
		if aeJSONEqual(x, d) || chMainOK(x) {
			t.Fatalf("history changed main control %d accepted", i)
		}
	}
}
func TestContainerHistoryOwnershipObserved(t *testing.T) {
	d, raw, _ := rsRead(t, "GOO_CONTAINER_HISTORY_OWNER_LEDGER")
	if !chOwnerOK(d) {
		t.Fatal("history held owner withdrawal/sibling survival oracle")
	}
	changes := []func(psMap){func(x psMap) { psObj(psObj(psArr(x["cases"])[0])["countsAfter"])["source"] = float64(99) }, func(x psMap) { psObj(psObj(psArr(x["cases"])[0])["after"])["published"] = true }, func(x psMap) { psObj(psObj(psArr(x["cases"])[4])["countsAfter"])["histories"] = float64(99) }, func(x psMap) { psObj(psObj(psObj(psArr(x["cases"])[0])["held"])["body"])["version"] = "wrong" }}
	for i, change := range changes {
		var x psMap
		_ = json.Unmarshal(raw, &x)
		change(x)
		if aeJSONEqual(x, d) || chOwnerOK(x) {
			t.Fatalf("history changed owner control %d accepted", i)
		}
	}
}

func chTrace(d psMap) *qobs.MemorySink {
	sink := &qobs.MemorySink{}
	actions := []string{"metadata", "both_histories", "local_point", "local_range", "local_range", "context_run_back", "context_run_back", "native_stage", "native_run_capture", "frozen"}
	for i, v := range psArr(d["events"]) {
		p := psObj(v)
		sink.Record(context.Background(), &qtrace.Step{Seq: uint64(i + 1), Recorder: "history-main", Outcome: "ok", Action: actions[i], Obs: map[string]any{"source": p["source"], "plans": p["plans"], "compositions": p["compositions"], "histories": p["histories"]}})
	}
	return sink
}
func chOwnerTrace(c psMap) *qobs.MemorySink {
	sink := &qobs.MemorySink{}
	held, after := psObj(c["countsHeld"]), psObj(c["countsAfter"])
	obs := func(p psMap) map[string]any {
		return map[string]any{"source": p["source"].(float64) - held["source"].(float64) + 3, "plans": p["plans"].(float64) - held["plans"].(float64) + 1, "compositions": p["compositions"].(float64) - held["compositions"].(float64) + 1, "histories": p["histories"].(float64) - held["histories"].(float64) + 1}
	}
	sink.Record(context.Background(), &qtrace.Step{Seq: 1, Recorder: "history-owner", Outcome: "ok", Action: "hold_history", Obs: obs(held)})
	if c["reason"] == "sibling-load-survives" {
		sink.Record(context.Background(), &qtrace.Step{Seq: 2, Recorder: "history-owner", Outcome: "ok", Action: "sibling_load", Obs: obs(after)})
		sink.Record(context.Background(), &qtrace.Step{Seq: 3, Recorder: "history-owner", Outcome: "ok", Action: "held_finishes", Obs: obs(after)})
	} else {
		sink.Record(context.Background(), &qtrace.Step{Seq: 2, Recorder: "history-owner", Outcome: "ok", Action: "withdraw", Obs: obs(after)})
		sink.Record(context.Background(), &qtrace.Step{Seq: 3, Recorder: "history-owner", Outcome: "ok", Action: "late_ignored", Obs: obs(after)})
	}
	return sink
}
func TestQuintContainerHistoryObserved(t *testing.T) {
	main, _, path := rsRead(t, "GOO_CONTAINER_HISTORY_LEDGER")
	owner, _, _ := rsRead(t, "GOO_CONTAINER_HISTORY_OWNER_LEDGER")
	if !chMainOK(main) || !chOwnerOK(owner) {
		t.Fatal("raw semantic precondition")
	}
	checker, e := validate.NewChecker("model/container_history.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	index := 0
	check := func(name string, sink *qobs.MemorySink) bool {
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()
		report, e := checker.Check(ctx, sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		b, _ := json.MarshalIndent(sink.Steps(), "", "  ")
		_ = os.WriteFile(filepath.Join(filepath.Dir(path), fmt.Sprintf("history-protocol-%02d-%s.json", index, name)), b, 0644)
		_ = os.WriteFile(filepath.Join(filepath.Dir(path), fmt.Sprintf("history-protocol-%02d-%s.txt", index, name)), []byte(report.Summary()), 0644)
		index++
		return report.OK()
	}
	if !check("main", chTrace(main)) {
		t.Fatal("actual history main trace")
	}
	for i, v := range psArr(owner["cases"]) {
		if !check(fmt.Sprintf("owner-%d", i), chOwnerTrace(psObj(v))) {
			t.Fatal("actual held history owner", i)
		}
	}
	bad := chTrace(main)
	bad.Steps()[2].Obs["source"] = float64(10)
	if check("bad-local-read", bad) {
		t.Fatal("extra local read accepted")
	}
	bad = chTrace(main)
	steps := bad.Steps()
	steps[0].Action = "context_run_back"
	bad = &qobs.MemorySink{}
	for i := range steps {
		bad.Record(context.Background(), &steps[i])
	}
	if bad.Steps()[0].Action != "context_run_back" {
		t.Fatal("order mutation did not persist")
	}
	if check("bad-context-order", bad) {
		t.Fatal("context before held range accepted")
	}
	bad = chOwnerTrace(psObj(psArr(owner["cases"])[0]))
	bad.Steps()[2].Obs["histories"] = float64(2)
	if check("bad-late-history", bad) {
		t.Fatal("late history effect accepted")
	}
}

package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type entryRaw390 struct {
	ServiceName, ScopeName, ScopeVersion, ScopeSchemaUrl, ResourceSchemaUrl, MetricName, MetricUnit string
	AtNS, StartNS, Value, Count, Flags, Temporality                                                 string
	IsMonotonic                                                                                     bool
	ResourceAttributes, Attributes                                                                  map[string]string
	BucketCounts                                                                                    []string
	ExplicitBounds                                                                                  []float64
}
type entryCounters390 struct{ Calls, Source, Store, SQL, Writes int }
type entryFact390 struct {
	Name          string
	Before, After entryCounters390
}
type entryRead390 struct {
	Name     string
	Request  measurements.Question
	Response json.RawMessage
	Status   int
}
type entryCase390 struct {
	Surface, State, ReadName, CaptureName string
	Width                                 int
	Before, After                         entryCounters390
	Geometry                              entryGeometry390
}
type entryCapture390 struct {
	Name, ID, StoredBytesBefore, StoredBytesAfter string
	Page                                          exploration.CapturePage
}
type entryLedger390 struct {
	Version, Mode string
	Errors        []string
	Fixture       struct {
		Counter, Gauge, Unknown, Empty, Histogram measurements.Question
		Inventory                                 exploration.InventoryQuestion
	}
	Raw struct {
		Sums, Gauges, Histograms []entryRaw390
		Spans, Logs              []json.RawMessage
	}
	Reads    []entryRead390
	Captures []entryCapture390
	Entries  []entryCase390
	Facts    []entryFact390
	Calls    []struct {
		Method, Path string
		Body         json.RawMessage
	}
	ServicePlan  exploration.DashboardPlan
	ServiceLoads []struct {
		Width, FromCall, ToCall int
		ExpectedPaths           []string
		Plan                    exploration.DashboardPlan
	}
	ProxyCalls []struct {
		SQL        string
		Parameters map[string]string
		Status     int
	}
	Actor struct {
		Snapshots           int
		SQLBefore, SQLAfter int
		Complete            map[string]json.RawMessage
	}
	LocalTitles []struct {
		Surface           string
		Width             int
		Initial, Authored string
	}
	LocalEvidence []struct {
		Surface       string
		Width         int
		Before, After json.RawMessage
	}
}

func rawIdentity390(q measurements.Question, r entryRaw390) bool {
	b := q.Definition.Source
	if !slices.Contains(b.Services, r.ServiceName) || b.Scope != r.ScopeName || b.Version != r.ScopeVersion || b.SchemaURL != r.ScopeSchemaUrl || b.Metric != r.MetricName || b.Unit != r.MetricUnit || !continuationNamespace352(b.Namespace, r.ResourceAttributes) {
		return false
	}
	mode, name := 0, q.Definition.Environment
	if s := q.Definition.EnvironmentSelector; s != nil {
		name = s.Scope.Value
		if s.Policy == "effective" {
			mode = 2
		}
		if s.Scope.Mode == "empty" {
			mode = 3
		} else if s.Scope.Mode == "unreported" {
			mode = 4
		}
	}
	if !continuationEnvironment352(mode, name, r.ResourceAttributes) {
		return false
	}
	for k, v := range b.Attributes {
		if actual, ok := r.Attributes[k]; !ok || actual != v {
			return false
		}
	}
	if s := b.ExactStream; s != nil {
		if !reflect.DeepEqual(s.Resource, r.ResourceAttributes) || !reflect.DeepEqual(s.Attributes, r.Attributes) || s.ResourceSchemaURL != r.ResourceSchemaUrl {
			return false
		}
	}
	return true
}
func flags390(v string) (uint64, error) {
	n, e := strconv.ParseUint(v, 10, 32)
	if e != nil || strconv.FormatUint(n, 10) != v {
		return 0, fmt.Errorf("noncanonical source flag")
	}
	return n, nil
}
func admitRead390(read entryRead390, report entryLedger390) (*measurements.Result, error) {
	fail := func(reason string) (*measurements.Result, error) {
		return nil, fmt.Errorf("native390 %s: %s", read.Name, reason)
	}
	if read.Request.Validate() != nil {
		return fail("invalid requested question")
	}
	if read.Status == 502 {
		var body struct {
			Refusal struct{ Kind string }
			Error   string
		}
		if json.Unmarshal(read.Response, &body) != nil || body.Refusal.Kind == "" || body.Error == "" {
			return fail("refusal was not observed")
		}
		return nil, nil
	}
	if read.Status != 200 {
		return fail("unexpected status")
	}
	var result measurements.Result
	if json.Unmarshal(read.Response, &result) != nil || !reflect.DeepEqual(read.Request, result.Question) || measurements.ValidateResult(&result) != nil {
		return fail("measurement replay/question identity changed")
	}
	loText, hiText := read.Request.BoundsNS()
	lo, _ := strconv.ParseInt(loText, 10, 64)
	hi, _ := strconv.ParseInt(hiText, 10, 64)
	switch read.Request.Definition.Kind {
	case "gauge":
		rows := []gaugeRaw386{}
		for _, r := range report.Raw.Gauges {
			flags, e := flags390(r.Flags)
			if e != nil {
				return fail(e.Error())
			}
			rows = append(rows, gaugeRaw386{AtNS: r.AtNS, Value: r.Value, ServiceName: r.ServiceName, MetricName: r.MetricName, MetricUnit: r.MetricUnit, ScopeName: r.ScopeName, ScopeVersion: r.ScopeVersion, ScopeSchemaUrl: r.ScopeSchemaUrl, ResourceSchemaUrl: r.ResourceSchemaUrl, Flags: int(flags), ResourceAttributes: r.ResourceAttributes, Attributes: r.Attributes})
		}
		want := gaugeReference386(read.Request, rows)
		if result.Gauges == nil || result.Count != strconv.Itoa(len(want)) {
			return fail("raw gauge population differs")
		}
		if len(want) == 0 {
			if len(result.Gauges.Series) != 0 || result.State != "no_data" {
				return fail("empty gauge invented data")
			}
			return &result, nil
		}
		if len(result.Gauges.Series) != 1 || len(result.Gauges.Series[0].Samples) != len(want) {
			return fail("gauge sample identity missing")
		}
		for i, w := range want {
			got := result.Gauges.Series[0].Samples[i]
			if got.AtNS != w.AtNS || got.State != w.State || got.Reason != w.Reason {
				return fail("gauge time/freshness state changed")
			}
			if w.Value == nil {
				if got.Value != nil {
					return fail("unknown became a number")
				}
			} else {
				var text string
				json.Unmarshal(*w.Value, &text)
				value, e := strconv.ParseFloat(text, 64)
				if e != nil || got.Value == nil || *got.Value != value {
					return fail("raw gauge value changed")
				}
			}
		}
	case "counter":
		selected := []entryRaw390{}
		for _, r := range report.Raw.Sums {
			if !rawIdentity390(read.Request, r) {
				continue
			}
			a, e := strconv.ParseInt(r.StartNS, 10, 64)
			b, e2 := strconv.ParseInt(r.AtNS, 10, 64)
			if e != nil || e2 != nil || b <= a {
				return fail("invalid raw export bounds")
			}
			if b > lo && a < hi {
				selected = append(selected, r)
			}
		}
		sort.Slice(selected, func(i, j int) bool { return selected[i].AtNS < selected[j].AtNS })
		if result.Counters == nil || len(result.Counters.Series) != 1 || result.Count != strconv.Itoa(len(selected)) || len(result.Counters.Series[0].Intervals) != len(selected) {
			return fail("raw counter population differs")
		}
		for i, r := range selected {
			if r.Temporality != "1" || !r.IsMonotonic {
				return fail("bounded delta counter oracle profile changed")
			}
			a, _ := strconv.ParseInt(r.StartNS, 10, 64)
			b, _ := strconv.ParseInt(r.AtNS, 10, 64)
			flags, e := flags390(r.Flags)
			if e != nil {
				return fail(e.Error())
			}
			got := result.Counters.Series[0].Intervals[i]
			if got.FromNS != r.StartNS || got.ToNS != r.AtNS {
				return fail("raw delta interval membership changed")
			}
			if flags&1 != 0 {
				if got.State != "unknown" || got.Reason != "no-recorded-value" || got.PerSecond != nil || got.Delta != "" {
					return fail("absent export became a number")
				}
				continue
			}
			if a < lo || b > hi || b-a > read.Request.Definition.MaxGapMS*1000000 {
				return fail("fixture unexpectedly needs another counter oracle profile")
			}
			n, ok := new(big.Int).SetString(r.Value, 10)
			if !ok || n.Sign() < 0 {
				return fail("invalid raw integer increment")
			}
			rate := new(big.Rat).Mul(new(big.Rat).SetInt(n), big.NewRat(1000000000, b-a))
			v, _ := rate.Float64()
			if got.State != "ok" || got.Delta != r.Value || got.PerSecond == nil || *got.PerSecond != v {
				return fail("raw increment/rate changed")
			}
		}
	case "histogram":
		selected := []entryRaw390{}
		for _, r := range report.Raw.Histograms {
			if rawIdentity390(read.Request, r) {
				a, e := strconv.ParseInt(r.StartNS, 10, 64)
				b, e2 := strconv.ParseInt(r.AtNS, 10, 64)
				if e != nil || e2 != nil {
					return fail("invalid raw histogram bounds")
				}
				if b > lo && a < hi {
					selected = append(selected, r)
				}
			}
		}
		if len(selected) != 1 || result.Count != "1" || result.Histograms == nil || len(result.Histograms.Series) != 1 || len(result.Histograms.Series[0].Intervals) != 1 {
			return fail("physical histogram export differs")
		}
		r := selected[0]
		got := result.Histograms.Series[0].Intervals[0]
		if r.Temporality != "1" || r.Flags != "0" || got.State != "ok" || got.FromNS != r.StartNS || got.ToNS != r.AtNS || got.Distribution == nil || got.Distribution.Count != r.Count || !reflect.DeepEqual(got.Distribution.Bounds, r.ExplicitBounds) {
			return fail("histogram export value/bounds changed")
		}
		var total uint64
		if len(got.Distribution.BucketCounts) != len(r.BucketCounts) {
			return fail("raw bucket cardinality changed")
		}
		for i, text := range r.BucketCounts {
			n, e := strconv.ParseUint(text, 10, 64)
			if e != nil || strconv.FormatUint(n, 10) != text {
				return fail("raw histogram bucket is not a canonical count")
			}
			total += n
			if got.Distribution.BucketCounts[i] != text {
				return fail("raw bucket population changed")
			}
		}
		if strconv.FormatUint(total, 10) != r.Count {
			return fail("raw histogram count does not reconcile")
		}
	default:
		return fail("unexpected native family")
	}
	return &result, nil
}

func entrySource390(path string) bool {
	return slices.Contains([]string{"/api/measurements", "/api/explore", "/api/explore/page", "/api/discovery", "/api/inventory", "/api/metric-discovery", "/api/record-aggregates"}, path)
}
func serviceCost390(plan exploration.DashboardPlan) int {
	n := 5
	for _, w := range plan.Widgets {
		if w.Query != nil || w.Measurement != nil {
			n++
		}
	}
	return n
}
func expectedFacts390() []string {
	out := []string{}
	for _, width := range []int{1280, 390} {
		s := strconv.Itoa(width)
		for _, kind := range []string{"counter", "gauge", "unknown-only", "empty"} {
			out = append(out, "ordinary-"+kind+"-run-"+s, "ordinary-"+kind+"-local-"+s)
		}
		out = append(out, "ordinary-refused-run-"+s, "ordinary-retry-"+s, "saved-counter-open-"+s, "service-load-"+s, "service-expand-local-"+s)
	}
	return append(out, "complete")
}

func admitEntryLedger390(t *testing.T, r entryLedger390, baseline *entryLedger390) error {
	fail := func(reason string) error { return fmt.Errorf("native390 ledger: %s", reason) }
	if r.Version != "goo.nativeEntry390/1" || !slices.Contains([]string{"baseline", "final"}, r.Mode) || len(r.Errors) != 0 || len(r.Entries) != 14 || len(r.Raw.Sums) != 3 || len(r.Raw.Gauges) != 4 || len(r.Raw.Histograms) != 1 || len(r.Raw.Spans) == 0 || len(r.Raw.Spans) != len(r.Raw.Logs) {
		return fail("incomplete actual report/bank")
	}
	reads := map[string]*measurements.Result{}
	readQuestions := map[string]measurements.Question{}
	for _, read := range r.Reads {
		if _, ok := readQuestions[read.Name]; ok {
			return fail("duplicate read identity")
		}
		result, e := admitRead390(read, r)
		if e != nil {
			return e
		}
		reads[read.Name] = result
		readQuestions[read.Name] = read.Request
	}
	if len(reads) != 14 {
		return fail("required successful/refused/retry/service reads missing")
	}
	for _, read := range r.Reads {
		if strings.HasPrefix(read.Name, "service-histogram-") {
			matched := false
			for _, w := range r.ServicePlan.Widgets {
				if w.Measurement != nil && reflect.DeepEqual(*w.Measurement, read.Request) {
					matched = true
				}
			}
			if !matched {
				return fail("service measurement differs from observed widget plan")
			}
			continue
		}
		want := r.Fixture.Counter
		if strings.HasPrefix(read.Name, "ordinary-gauge-") {
			want = r.Fixture.Gauge
		} else if strings.HasPrefix(read.Name, "ordinary-unknown-only-") {
			want = r.Fixture.Unknown
		} else if strings.HasPrefix(read.Name, "ordinary-empty-") {
			want = r.Fixture.Empty
		}
		if !reflect.DeepEqual(read.Request, want) {
			return fail("ordinary request inherited a different question")
		}
	}
	captures := map[string]entryCapture390{}
	for _, c := range r.Captures {
		if c.StoredBytesBefore == "" || c.StoredBytesBefore != c.StoredBytesAfter || c.Page.Receipt.ID != c.ID {
			return fail("immutable capture bytes or receipt changed")
		}
		var saved exploration.SavedCapture
		if json.Unmarshal([]byte(c.StoredBytesBefore), &saved) != nil || saved.ID != c.ID {
			return fail("stored capture shape changed")
		}
		id := saved.ID
		saved.ID = ""
		canonical, e := json.Marshal(saved)
		if e != nil {
			return e
		}
		hash := sha256.Sum256(canonical)
		if id != hex.EncodeToString(hash[:]) {
			return fail("physical capture hash changed")
		}
		saved.ID = id
		page, e := saved.Page("")
		if e != nil || !reflect.DeepEqual(page, &c.Page) {
			return fail("capture delivery differs from physical evidence")
		}
		if c.Name == "counter" {
			if c.Page.Measurement == nil || !reflect.DeepEqual(c.Page.Measurement.Question, r.Fixture.Counter) {
				return fail("captured measurement question changed")
			}
			b, _ := json.Marshal(c.Page.Measurement)
			if _, e := admitRead390(entryRead390{Name: "held-counter", Request: r.Fixture.Counter, Response: b, Status: 200}, r); e != nil {
				return e
			}
		}
		captures[c.Name] = c
	}
	if len(captures) != 2 || captures["inventory"].Page.Inventory == nil || captures["counter"].Page.Measurement == nil {
		return fail("required captured origins missing")
	}
	if len(r.ServiceLoads) != 2 {
		return fail("service entry acquisitions missing")
	}
	servicePaths := []string{"/api/discovery", "/api/discovery", "/api/explore", "/api/metric-discovery", "/api/inventory", "/api/record-aggregates", "/api/record-aggregates", "/api/measurements"}
	for _, load := range r.ServiceLoads {
		if !slices.Contains([]int{1280, 390}, load.Width) || load.FromCall < 0 || load.ToCall > len(r.Calls) || load.FromCall >= load.ToCall || !slices.Equal(load.ExpectedPaths, servicePaths) {
			return fail("service acquisition range changed")
		}
		paths := []string{}
		nativeQueries := []measurements.Question{}
		for _, call := range r.Calls[load.FromCall:load.ToCall] {
			if call.Method == "POST" && entrySource390(call.Path) {
				paths = append(paths, call.Path)
				if call.Path == "/api/measurements" {
					var q measurements.Question
					if json.Unmarshal(call.Body, &q) != nil {
						return fail("service native request malformed")
					}
					nativeQueries = append(nativeQueries, q)
				}
			}
		}
		if !slices.Equal(paths, servicePaths) || serviceCost390(load.Plan) != len(servicePaths) || len(nativeQueries) != 1 {
			return fail("service source operations differ from explicit observed flow")
		}
		matched := false
		for _, w := range load.Plan.Widgets {
			if w.Measurement != nil && reflect.DeepEqual(*w.Measurement, nativeQueries[0]) {
				matched = true
			}
		}
		if !matched {
			return fail("service actual native request changed from its plan")
		}
	}
	seen := map[string]bool{}
	for _, e := range r.Entries {
		key := e.Surface + "/" + strconv.Itoa(e.Width)
		if seen[key] || !slices.Contains([]string{"ordinary-counter", "ordinary-gauge", "ordinary-unknown-only", "ordinary-empty", "ordinary-refused", "saved-counter", "service-histogram"}, e.Surface) || !slices.Contains([]int{1280, 390}, e.Width) {
			return fail("entry matrix identity differs")
		}
		seen[key] = true
		if err := admitGeometry390(e.Geometry, e.Width, r.Mode == "final"); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		var result *measurements.Result
		if e.Surface == "saved-counter" {
			if e.CaptureName != "counter" || e.After.Source != e.Before.Source || e.After.SQL != e.Before.SQL || e.After.Store-e.Before.Store != 1 {
				return fail("saved entry reread telemetry")
			}
			result = captures["counter"].Page.Measurement
		} else {
			var ok bool
			result, ok = reads[e.ReadName]
			if !ok {
				return fail("entry does not name observed read")
			}
		}
		if e.Surface == "ordinary-refused" {
			if result != nil || e.State != "refused" || e.Geometry.Primary.Tag == "GOO-VALUE-TIMELINE" {
				return fail("refusal invented data")
			}
			continue
		}
		if result == nil {
			return fail("successful entry has no admitted data")
		}
		if r.Mode == "final" {
			s := e.Geometry.Summary
			if !s.Visible || s.Clip == nil || !fullyInside390(s.Rect, *s.Clip) {
				return fail("critical native state/window summary is clipped")
			}
			c, err := paintClip390(e.Width, int(e.Geometry.Viewport.Height), s.ClipAncestors)
			if err != nil || !sameClip390(c, *s.Clip) {
				return fail("summary clipping differs from its ancestors")
			}
			lo, hi := result.Question.BoundsNS()
			if s.FromNS != lo || s.ToNS != hi || !strings.Contains(s.Text, result.State) || !strings.Contains(s.Text, result.Count+" source records") {
				return fail("visible summary changed state/count/fixed evaluation bounds")
			}
		}
		if e.Surface == "ordinary-empty" {
			if result.State != "no_data" || result.Count != "0" || e.Geometry.Timeline != nil {
				return fail("empty state invented a numeric chart")
			}
			continue
		}
		if err := admitEntryTimeline390(e.Geometry, *result, r.Mode == "final"); err != nil {
			return fmt.Errorf("%s: %w", key, err)
		}
		if r.Mode == "final" && baseline != nil {
			for _, b := range baseline.Entries {
				if b.Surface == e.Surface && b.Width == e.Width && e.Geometry.Primary.Rect.AbsoluteDocumentTop >= b.Geometry.Primary.Rect.AbsoluteDocumentTop {
					return fail("primary entry moved only by scrolling rather than layout compaction")
				}
			}
		}
	}
	expected := expectedFacts390()
	if len(r.Facts) != len(expected) {
		return fail("incomplete measured interaction prefix")
	}
	for i, f := range r.Facts {
		if f.Name != expected[i] || f.Before.Source < 0 || f.After.Writes != 0 || f.Before.Writes != 0 {
			return fail("workflow identity/write boundary changed")
		}
		if i > 0 {
			p := r.Facts[i-1].After
			if f.Before.Source != p.Source || f.Before.Store != p.Store || f.Before.SQL != p.SQL {
				return fail("unmeasured read between boundaries")
			}
		}
		wantSource, wantStore := 0, 0
		if strings.Contains(f.Name, "-run-") || strings.HasPrefix(f.Name, "ordinary-retry-") {
			wantSource = 1
		} else if strings.HasPrefix(f.Name, "saved-counter-open-") {
			wantStore = 1
		} else if strings.HasPrefix(f.Name, "service-load-") {
			wantSource = serviceCost390(r.ServicePlan)
			wantStore = 1
		}
		if f.After.Source-f.Before.Source != wantSource || f.After.Store-f.Before.Store != wantStore || wantSource == 0 && f.After.SQL != f.Before.SQL {
			return fail("local/config/explicit-read budget changed")
		}
	}
	if r.Facts[0].Before.Source != 0 || r.Facts[0].Before.Store != 0 {
		return fail("source baseline hidden")
	}
	if r.Mode == "final" {
		if len(r.LocalTitles) != 4 || len(r.LocalEvidence) != 8 {
			return fail("local disclosure/title outcomes missing")
		}
		seenTitles := map[string]bool{}
		for _, item := range r.LocalTitles {
			key := item.Surface + "/" + strconv.Itoa(item.Width)
			if seenTitles[key] || !slices.Contains([]string{"ordinary-counter", "ordinary-gauge"}, item.Surface) || !slices.Contains([]int{1280, 390}, item.Width) || item.Initial == "" || strings.Contains(strings.ToLower(item.Initial), "timeline") || strings.Contains(strings.ToLower(item.Initial), "table") || item.Authored != "Authored context view" {
				return fail("display default title or authored preservation changed")
			}
			seenTitles[key] = true
		}
		seenEvidence := map[string]bool{}
		for _, item := range r.LocalEvidence {
			key := item.Surface + "/" + strconv.Itoa(item.Width)
			if seenEvidence[key] || len(item.Before) == 0 || nativeReuseJSON371(item.Before) != nativeReuseJSON371(item.After) {
				return fail("local disclosure or kind switch changed supplied evidence")
			}
			seenEvidence[key] = true
			read := reads[item.Surface+"-"+strconv.Itoa(item.Width)]
			if read == nil {
				return fail("local disclosure names no actual read")
			}
			var held struct {
				Timeline    *entryTimeline390
				SelectedKey *string
				Provenance  []string
			}
			if json.Unmarshal(item.Before, &held) != nil {
				return fail("local evidence is not an actual snapshot")
			}
			origin := false
			for _, text := range held.Provenance {
				var p struct {
					Question      measurements.Question
					Basis, ReadAt string
				}
				if json.Unmarshal([]byte(text), &p) == nil && p.Basis != "" {
					if p.Basis != read.Basis || p.ReadAt != read.ReadAt || !reflect.DeepEqual(p.Question, read.Question) {
						return fail("held source/custody clocks or question changed")
					}
					origin = true
				}
			}
			if !origin {
				return fail("local evidence has no retained source origin")
			}
		}
		var phase string
		json.Unmarshal(r.Actor.Complete["phase"], &phase)
		if r.Actor.Snapshots <= 0 || phase != "complete" || r.Actor.SQLAfter-r.Actor.SQLBefore != 1 {
			return fail("positive actual actor outcome missing")
		}
		for _, key := range []string{"observed", "current", "paint", "stateVisible", "selected", "actionsPresent", "budgetSafe"} {
			var v bool
			if json.Unmarshal(r.Actor.Complete[key], &v) != nil || !v {
				return fail("actor outcome missing: " + key)
			}
		}
		for key, want := range map[string]int{"source": 1, "store": 0, "writes": 0} {
			var n int
			if json.Unmarshal(r.Actor.Complete[key], &n) != nil || n != want {
				return fail("actor read/publication budget changed")
			}
		}
		var actorRead entryRead390
		actorRead.Name = "actor-unknown-only"
		if json.Unmarshal(r.Actor.Complete["request"], &actorRead.Request) != nil || !reflect.DeepEqual(actorRead.Request, r.Fixture.Unknown) || json.Unmarshal(r.Actor.Complete["status"], &actorRead.Status) != nil {
			return fail("actor's actual source question changed")
		}
		actorRead.Response = r.Actor.Complete["response"]
		if _, err := admitRead390(actorRead, r); err != nil {
			return err
		}
	}
	if err := admitPhysicalNative390(r); err != nil {
		return err
	}
	return nil
}

// The two fixture publications precede the UI counters. Counter custody performs
// one native source read; inventory has a different explicitly observed query.
// Native requests below include every actual ordinary/service/retry read and,
// for the final report, the independent actor's read. Other service record and
// discovery operations are separately bounded by the endpoint ledger.
func admitPhysicalNative390(r entryLedger390) error {
	type expected struct{ Count int }
	expectations := map[string]*expected{}
	knownSQL := map[string]bool{}
	key := func(sql string, parameters map[string]string) string {
		b, _ := json.Marshal(parameters)
		return sql + "\x00" + string(b)
	}
	add := func(q measurements.Question, status int) error {
		p, e := measurements.Compile(q)
		if e != nil {
			return e
		}
		s := status
		if s >= 500 {
			s = 503
		}
		k := key(p.SQL, p.Parameters) + "\x00" + strconv.Itoa(s)
		if expectations[k] == nil {
			expectations[k] = &expected{}
		}
		expectations[k].Count++
		knownSQL[p.SQL] = true
		return nil
	}
	if e := add(r.Fixture.Counter, 200); e != nil {
		return e
	}
	for _, read := range r.Reads {
		if e := add(read.Request, read.Status); e != nil {
			return e
		}
	}
	if r.Mode == "final" {
		if e := add(r.Fixture.Unknown, 200); e != nil {
			return e
		}
	}
	actual := 0
	for _, call := range r.ProxyCalls {
		if !knownSQL[call.SQL] {
			continue
		}
		params := map[string]string{}
		for k, v := range call.Parameters {
			if strings.HasPrefix(k, "param_") {
				params[strings.TrimPrefix(k, "param_")] = v
			}
		}
		k := key(call.SQL, params) + "\x00" + strconv.Itoa(call.Status)
		want := expectations[k]
		if want == nil || want.Count <= 0 || call.Parameters["readonly"] != "1" {
			return fmt.Errorf("native390 physical source SQL/typed parameters/status/readonly changed")
		}
		want.Count--
		actual++
	}
	for _, want := range expectations {
		if want.Count != 0 {
			return fmt.Errorf("native390 native physical source request missing")
		}
	}
	count := len(r.Reads) + 1
	if r.Mode == "final" {
		count++
	}
	if actual != count {
		return fmt.Errorf("native390 unexpected native source request count")
	}
	return nil
}

func admitEntryTimeline390(g entryGeometry390, r measurements.Result, final bool) error {
	fail := func(s string) error { return fmt.Errorf("native390 supplied data: %s", s) }
	if g.Timeline == nil {
		return fail("actual primary timeline missing")
	}
	var want []entryPoint390
	if r.Gauges != nil {
		if len(r.Gauges.Series) != 1 {
			return fail("selected gauge stream differs")
		}
		for i, p := range r.Gauges.Series[0].Samples {
			point := entryPoint390{Key: strconv.Itoa(i), FromNS: p.AtNS, ToNS: p.AtNS, State: p.State}
			if p.Value != nil {
				v := strconv.FormatFloat(*p.Value, 'f', -1, 64)
				point.Value = &v
			} else {
				point.State = "unknown"
			}
			want = append(want, point)
		}
	} else if r.Counters != nil {
		if len(r.Counters.Series) != 1 {
			return fail("selected counter stream differs")
		}
		for i, p := range r.Counters.Series[0].Intervals {
			point := entryPoint390{Key: strconv.Itoa(i), FromNS: p.FromNS, ToNS: p.ToNS, State: "unknown"}
			if p.State == "ok" && p.PerSecond != nil {
				v := strconv.FormatFloat(*p.PerSecond, 'f', -1, 64)
				point.Value = &v
				point.State = "observed"
			}
			want = append(want, point)
		}
	} else if r.Histograms != nil {
		if len(r.Histograms.Series) != 1 {
			return fail("selected histogram stream differs")
		}
		for i, p := range r.Histograms.Series[0].Intervals {
			point := entryPoint390{Key: strconv.Itoa(i), FromNS: p.FromNS, ToNS: p.ToNS, State: "unknown"}
			if p.State == "ok" && p.Distribution != nil {
				v := p.Distribution.Count
				point.Value = &v
				point.State = "observed"
			}
			want = append(want, point)
		}
	} else {
		return fail("unexpected family")
	}
	if !reflect.DeepEqual(g.Timeline.Points, want) || len(g.Marks) != len(want) || g.Timeline.Kind != r.Question.Definition.Kind {
		return fail("display values/state/time differ from admitted source result")
	}
	known := 0
	fromText, toText := r.Question.BoundsNS()
	if g.Timeline.FromNS != fromText || g.Timeline.ToNS != toText || g.Timeline.Total != len(want) || g.Timeline.Omitted != 0 {
		return fail("native fixed window or retained population changed")
	}
	lo, _ := new(big.Rat).SetString(fromText)
	hi, _ := new(big.Rat).SetString(toText)
	// Independent numeric domain includes zero and every retained known value.
	// A button's border-box center represents its supplied ordinate; parent
	// button margins must not shift it away from that data coordinate.
	minimum, maximum := new(big.Rat), new(big.Rat)
	for _, point := range want {
		if point.Value == nil {
			continue
		}
		v, ok := new(big.Rat).SetString(*point.Value)
		if !ok {
			return fail("known value is not a finite rational")
		}
		if v.Cmp(minimum) < 0 {
			minimum.Set(v)
		}
		if v.Cmp(maximum) > 0 {
			maximum.Set(v)
		}
	}
	for i, m := range g.Marks {
		p := want[i]
		if m.Key != p.Key || m.State != p.State || m.FromNS != p.FromNS || m.ToNS != p.ToNS || !reflect.DeepEqual(m.Value, p.Value) {
			return fail("painted marker changed supplied value")
		}
		if p.Value != nil {
			known++
		}
		if !final {
			continue
		}
		a, _ := new(big.Rat).SetString(p.FromNS)
		b, _ := new(big.Rat).SetString(p.ToNS)
		at := new(big.Rat).Quo(new(big.Rat).Add(a, b), big.NewRat(2, 1))
		fraction, _ := new(big.Rat).Quo(new(big.Rat).Sub(at, lo), new(big.Rat).Sub(hi, lo)).Float64()
		if math.Abs(m.Rect.Left+m.Rect.Width/2-(g.NumericPlot.Left+g.NumericPlot.Width*fraction)) > 1 {
			return fail("actual marker time coordinate changed")
		}
		if p.Value != nil {
			fraction := 0.0
			if maximum.Cmp(minimum) != 0 {
				v, _ := new(big.Rat).SetString(*p.Value)
				fraction, _ = new(big.Rat).Quo(new(big.Rat).Sub(v, minimum), new(big.Rat).Sub(maximum, minimum)).Float64()
			}
			// Remove the plot's one-pixel bottom border from its content height.
			// The remaining tolerance covers subpixel browser quantization.
			if math.Abs(m.Rect.Top+m.Rect.Height/2-(g.NumericPlot.Top+(g.NumericPlot.Height-1)*(1-fraction))) > 0.1 {
				return fail("actual marker numeric coordinate changed")
			}
		}
		if p.Value == nil && m.Rect.Top < g.NumericPlot.Bottom-1 {
			return fail("unknown mark occupies numeric domain")
		}
	}
	if final && known == 0 && (g.NumericPlot == nil || g.NumericPlot.Height > 1.1) {
		return fail("unknown-only data retained an empty numeric scale")
	}
	return nil
}

func TestNativeEntryBaseline390(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_ENTRY_390_BASELINE")
	if path == "" {
		t.Skip("actual old-binary baseline required")
	}
	b, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var r entryLedger390
	if e := json.Unmarshal(b, &r); e != nil {
		t.Fatal("invalid baseline report:", e)
	}
	if r.Mode != "baseline" {
		t.Fatal("baseline mode required")
	}
	if e := admitEntryLedger390(t, r, nil); e != nil {
		t.Fatal(e)
	}
	t.Logf("admitted actual baseline: %d entries, %d facts, %d native read outcomes", len(r.Entries), len(r.Facts), len(r.Reads))
}

func TestQuintNativeEntryObserved390(t *testing.T) {
	path := os.Getenv("GOO_NATIVE_ENTRY_390_REPORT")
	if path == "" {
		t.Skip("actual native entry journey required")
	}
	load := func(path string) entryLedger390 {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		var r entryLedger390
		if json.Unmarshal(b, &r) != nil {
			t.Fatal("invalid native entry report")
		}
		return r
	}
	r := load(path)
	baselinePath := os.Getenv("GOO_NATIVE_ENTRY_390_BASELINE")
	if baselinePath == "" {
		t.Fatal("actual old-binary baseline required")
	}
	before := load(baselinePath)
	if before.Mode != "baseline" || r.Mode != "final" {
		t.Fatal("before/final modes required")
	}
	if e := admitEntryLedger390(t, before, nil); e != nil {
		t.Fatal(e)
	}
	if e := admitEntryLedger390(t, r, &before); e != nil {
		t.Fatal(e)
	}
	if !replayEntry390(t, r, false) {
		t.Fatal("actual measured budget violates model")
	}
	raw, _ := json.Marshal(r)
	mutations := []func(*entryLedger390){func(r *entryLedger390) {
		r.Entries[0].Geometry.FirstItem.Rect.Top += 1000
		r.Entries[0].Geometry.FirstItem.Rect.Y += 1000
		r.Entries[0].Geometry.FirstItem.Rect.Bottom += 1000
	}, func(r *entryLedger390) { r.Captures[0].StoredBytesAfter += " " }, func(r *entryLedger390) { r.Entries[1].Geometry.Timeline.Points[1].Value = new("999") }, func(r *entryLedger390) {
		for i := range r.Raw.Gauges {
			if r.Raw.Gauges[i].Flags == "0" {
				r.Raw.Gauges[i].Value = "999"
				return
			}
		}
	}, func(r *entryLedger390) { r.Facts[1].After.Source++ }, func(r *entryLedger390) {
		r.Entries[2].Geometry.NumericPlot.Height = 150
		r.Entries[2].Geometry.NumericPlot.Bottom = r.Entries[2].Geometry.NumericPlot.Top + 150
	}}
	for i, mutate := range mutations {
		var bad entryLedger390
		json.Unmarshal(raw, &bad)
		mutate(&bad)
		if admitEntryLedger390(t, bad, &before) == nil {
			t.Fatalf("semantic mutant %d escaped", i)
		}
	}
	if replayEntry390(t, r, true) {
		t.Fatal("local source-read model control escaped")
	}
}

func replayEntry390(t *testing.T, r entryLedger390, mutate bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, f := range r.Facts {
		action := "local"
		if strings.Contains(f.Name, "-run-") || strings.HasPrefix(f.Name, "ordinary-retry-") {
			action = "read"
		} else if strings.HasPrefix(f.Name, "saved-counter-open-") {
			action = "openSaved"
		} else if strings.HasPrefix(f.Name, "service-load-") {
			action = "serviceLoad"
		} else if f.Name == "complete" {
			action = "finish"
		}
		reads := f.After.Source
		if mutate && f.Name == "ordinary-counter-local-1280" {
			reads++
		}
		rec.Record(context.Background(), action, nil, "obs.reads", reads, "obs.stores", f.After.Store, "obs.writes", f.After.Writes, "obs.finished", f.Name == "complete")
	}
	label := "actual"
	if mutate {
		label = "local-read-mutant"
	}
	wire, _ := json.Marshal(sink.Steps())
	if e := artifact("native-entry390-"+label+"-steps.json", wire); e != nil {
		t.Fatal(e)
	}
	spec, e := filepath.Abs("model/native_entry390.qnt")
	if e != nil {
		t.Fatal(e)
	}
	binding := fmt.Sprintf("spec: %q\nmodule: native_entry390\nconstants: {SERVICE_READS: '%d'}\ninvariants: [safety]\nactions:\n", spec, serviceCost390(r.ServicePlan))
	for _, name := range []string{"local", "read", "openSaved", "serviceLoad", "finish"} {
		binding += "  " + name + ": {expect: 'reads == {{.obs.reads}} and stores == {{.obs.stores}} and writes == {{.obs.writes}} and finished == {{.obs.finished}}'}\n"
	}
	dir := t.TempDir()
	bindingPath := filepath.Join(dir, "binding.yaml")
	if os.WriteFile(bindingPath, []byte(binding), 0600) != nil {
		t.Fatal("cannot write replay binding")
	}
	checker, e := validate.NewChecker(bindingPath, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(t.Context(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	if e := artifact("native-entry390-"+label+"-model.txt", []byte(result.Summary())); e != nil {
		t.Fatal(e)
	}
	return result.OK()
}

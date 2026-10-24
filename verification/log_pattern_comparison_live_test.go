package verification

import (
	"bytes"
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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

// All expected groups come from raw source bodies and separately fixed windows.
// The across-time display is not a new telemetry grain or statistical detector.
func lpcLiveCohort(c exploration.Cohort, environment string, records []lpLiveRecord, allowChanged bool) bool {
	if c.Question.Signal != "logs" || len(c.Question.GroupBy) != 1 || c.Question.GroupBy[0] != (exploration.Field{Kind: "column", Key: "body_pattern_v1"}) || c.Question.Numeric != nil || c.Numeric != nil || c.NumericHistogram != nil || c.Duration != nil || c.SpanStatus != nil || !peLiveHex(c.Basis) {
		return false
	}
	rows, e := lpLiveRows(c.Question, environment, records)
	if e != nil || c.Count != strconv.Itoa(len(rows)) {
		return false
	}
	state := "observed"
	if len(rows) == 0 {
		state = "no_data"
	}
	if c.State != state && !(allowChanged && c.State == "changed_between_reads") {
		return false
	}
	groups := map[string]int{}
	for _, r := range rows {
		groups[lpOracle(r.Body).Identity]++
	}
	if len(groups) != len(c.Groups) {
		return false
	}
	seen := map[string]bool{}
	for _, g := range c.Groups {
		if g.Value == nil || !lpCanonical(*g.Value) || seen[*g.Value] || g.Count != strconv.Itoa(groups[*g.Value]) || g.Numeric != nil || g.Duration != nil || g.SpanStatus != nil {
			return false
		}
		seen[*g.Value] = true
	}
	return true
}
func lpcLiveAPI(request exploration.ComparisonRequest, response exploration.CohortComparison, environment string, records []lpLiveRecord, allowChanged bool) bool {
	if request.Numeric || request.Distribution || !lpcLiveCohort(request.Baseline, environment, records, allowChanged) || !lpcLiveCohort(request.Selection, environment, records, allowChanged) || response.Field != (exploration.Field{Kind: "column", Key: "body_pattern_v1"}) || response.BaselineCount != request.Baseline.Count || response.SelectionCount != request.Selection.Count || response.BaselineBasis != request.Baseline.Basis || response.SelectionBasis != request.Selection.Basis || response.Measures != nil {
		return false
	}
	canonical := request
	canonical.Baseline.Groups = append([]exploration.ValueCount(nil), request.Baseline.Groups...)
	canonical.Selection.Groups = append([]exploration.ValueCount(nil), request.Selection.Groups...)
	sortNumericGroups(canonical.Baseline.Groups)
	sortNumericGroups(canonical.Selection.Groups)
	raw, _ := json.Marshal(canonical)
	hash := sha256.Sum256(raw)
	if response.Basis != hex.EncodeToString(hash[:]) {
		return false
	}
	state := "ok"
	if request.Baseline.State == "changed_between_reads" || request.Selection.State == "changed_between_reads" {
		state = "unknown"
	} else if request.Baseline.Count == "0" || request.Selection.Count == "0" {
		state = "no_data"
	}
	if response.State != state {
		return false
	}
	if state != "ok" {
		return len(response.Rows) == 0
	}
	expected := lpcExpected(request, "share")
	if len(response.Rows) != len(expected.Rows) {
		return false
	}
	for i, row := range response.Rows {
		want := expected.Rows[i]
		if row.Value == nil || *row.Value != want.ID || row.BaselineCount != *want.BaselineCount || row.SelectionCount != *want.CurrentCount || row.Measures != nil {
			return false
		}
		a, _ := new(big.Rat).SetString(*want.BaselineShare)
		z, _ := new(big.Rat).SetString(*want.CurrentShare)
		d, _ := new(big.Rat).SetString(*want.ShareDifference)
		delta := d.FloatString(2)
		if delta == "-0.00" {
			delta = "0.00"
		}
		if row.BaselinePercent != a.FloatString(2) || row.SelectionPercent != z.FloatString(2) || row.DifferencePP != delta {
			return false
		}
	}
	return true
}

type lpcLiveModelRow struct {
	lpcModelRow
	Label, State string
	Classified   bool
}
type lpcLiveModel struct {
	State, Metric, BaselineCount, CurrentCount, BaselineWidthMS, CurrentWidthMS string
	Rows                                                                        []lpcLiveModelRow
}

func lpcLiveModelCheck(actual lpcLiveModel, request exploration.ComparisonRequest, records []lpLiveRecord) bool {
	if actual.Metric != "rate" && actual.Metric != "count" && actual.Metric != "share" {
		return false
	}
	plain := lpcModel{State: actual.State, Metric: actual.Metric, BaselineCount: actual.BaselineCount, CurrentCount: actual.CurrentCount, BaselineWidthMS: actual.BaselineWidthMS, CurrentWidthMS: actual.CurrentWidthMS, Rows: []lpcModelRow{}}
	labels := map[string]exploration.LogPattern{}
	for _, r := range records {
		p := lpOracle(r.Body)
		labels[p.Identity] = p
	}
	for _, row := range actual.Rows {
		p, ok := labels[row.ID]
		if !ok || row.Label != p.Label || row.State != p.State || row.Classified != (p.State == "pattern") {
			return false
		}
		plain.Rows = append(plain.Rows, row.lpcModelRow)
	}
	return reflect.DeepEqual(plain, lpcExpected(request, actual.Metric))
}
func lpcLivePairSources(request exploration.ComparisonRequest, baseline, selection exploration.Result, environment string, records []lpLiveRecord) bool {
	if !lpLiveSource(baseline, environment, records) || !lpLiveSource(selection, environment, records) {
		return false
	}
	a, e := decLiveCohort(baseline)
	if e != nil {
		return false
	}
	z, e := decLiveCohort(selection)
	return e == nil && reflect.DeepEqual(a, request.Baseline) && reflect.DeepEqual(z, request.Selection)
}
func lpcProfileQuestion(parent exploration.Query) exploration.Query {
	q := comparisonClone(parent)
	q.Numeric = nil
	q.NumericHistogram = nil
	q.Summary = ""
	q.SpanStatus = false
	q.GroupTimeSeries = false
	q.GroupBy = []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}
	q.Offset = 0
	q.Limit = 100
	for _, bucket := range []int64{1000, 10000, 60000, 300000} {
		if (q.ToMS-q.FromMS+bucket-1)/bucket <= 240 {
			q.BucketMS = bucket
			break
		}
	}
	return q
}

type lpcLivePair struct {
	Baseline, Selection exploration.Result
	Request             exploration.ComparisonRequest
	Response            exploration.CohortComparison
}
type lpcLiveCell struct {
	Value, Max, Side, Exact string
	FillPercent             *string
	FillWidth               *float64
	TrackWidth              float64
}
type lpcLiveDOMRow struct {
	Pattern, Label                                                                                                                                  string
	BaselineCount, CurrentCount, BaselineRate, CurrentRate, BaselineShare, CurrentShare, CountDifference, RateDifference, ShareDifference, Presence string
	Cells                                                                                                                                           []lpcLiveCell
}
type lpcLiveView struct {
	Metric string
	Model  lpcLiveModel
	Rows   []lpcLiveDOMRow
}
type lpcLiveEvent struct {
	Action                                           string
	SourceOperations, Protocols, Writes, CaptureGets int
}
type lpcLiveRequest struct {
	Effect, Endpoint string
	Payload          json.RawMessage
}
type lpcLiveMain struct {
	Version              int
	Environment          string
	Records, LateRecords []lpLiveRecord
	Selector             struct {
		Signal, Environment string
		Where               *exploration.Predicate
	}
	Windows                                                       struct{ Current, Previous, Saved struct{ FromMS, ToMS int64 } }
	SavedQuestion                                                 exploration.Query
	Current, LateReadyProbe                                       exploration.Result
	Previous, SavedPair, Reused                                   lpcLivePair
	PreviousViews                                                 []lpcLiveView
	SavedViews                                                    map[string]lpcLiveView
	FrozenView                                                    lpcLiveView
	ScopeCapture, LoadedScope, ScopeCaptureAfter, Saved, Reopened exploration.CapturePage
	CaptureSources                                                struct{ Baseline, Selection exploration.CapturePage }
	Custody                                                       decLiveCustody
	SourceStored                                                  exploration.SavedCapture
	SourceBytes, AfterBytes                                       []byte
	Reusable                                                      workbench.ExplorationView
	SideOpen                                                      struct {
		Side, ID         string
		Parent, Question exploration.Query
		Result           exploration.Result
	}
	Geometry map[string]struct{ Baseline, Current lpcLiveCell }
	Mobile   []struct {
		ID    string
		Cells []struct {
			Left, Right, Viewport float64
			Exact                 string
			NoTextClip            bool
		}
	}
	Events          []lpcLiveEvent
	Requests        []lpcLiveRequest
	SourceResponses []struct {
		Request  exploration.Query
		Status   int
		Response exploration.Result
	}
}

func lpcLiveViewCheck(view lpcLiveView, request exploration.ComparisonRequest, records []lpLiveRecord) bool {
	if view.Metric != view.Model.Metric || !lpcLiveModelCheck(view.Model, request, records) || len(view.Rows) != len(view.Model.Rows) {
		return false
	}
	max := new(big.Rat)
	unit := "logs/s"
	if view.Metric == "count" {
		unit = "logs"
	}
	if view.Metric == "share" {
		unit = "%"
		max.SetInt64(100)
	}
	value := func(row lpcModelRow, side int) *string {
		if view.Metric == "count" {
			return []*string{row.BaselineCount, row.CurrentCount}[side]
		}
		if view.Metric == "share" {
			return []*string{row.BaselineShare, row.CurrentShare}[side]
		}
		return []*string{row.BaselineRate, row.CurrentRate}[side]
	}
	for _, row := range view.Model.Rows {
		for side := 0; side < 2; side++ {
			if raw := value(row.lpcModelRow, side); raw != nil {
				v, ok := new(big.Rat).SetString(*raw)
				if !ok {
					return false
				}
				if v.Cmp(max) > 0 {
					max = v
				}
			}
		}
	}
	rendered := func(s *string) string {
		if s == nil {
			return "unavailable"
		}
		return *s
	}
	for i, row := range view.Rows {
		expected := view.Model.Rows[i]
		if row.Pattern != expected.ID || row.Label != expected.Label || row.BaselineCount != rendered(expected.BaselineCount) || row.CurrentCount != rendered(expected.CurrentCount) || row.BaselineRate != rendered(expected.BaselineRate) || row.CurrentRate != rendered(expected.CurrentRate) || row.BaselineShare != rendered(expected.BaselineShare) || row.CurrentShare != rendered(expected.CurrentShare) || row.CountDifference != rendered(expected.CountDifference) || row.RateDifference != rendered(expected.RateDifference) || row.ShareDifference != rendered(expected.ShareDifference) || row.Presence != expected.Presence || len(row.Cells) != 2 {
			return false
		}
		for side, c := range row.Cells {
			raw := value(expected.lpcModelRow, side)
			exact := "Unavailable"
			if raw != nil {
				exact = *raw + " " + unit + " (exact)"
			}
			if c.Side != []string{"baseline", "selection"}[side] || c.Value != rendered(raw) || c.Exact != exact || c.Max != max.RatString() || c.TrackWidth <= 0 {
				return false
			}
			if raw == nil {
				if c.FillPercent != nil || c.FillWidth != nil {
					return false
				}
				continue
			}
			if c.FillPercent == nil || c.FillWidth == nil {
				return false
			}
			actual, e := strconv.ParseFloat(strings.TrimSuffix(*c.FillPercent, "%"), 64)
			if e != nil {
				return false
			}
			ratio := new(big.Rat)
			if max.Sign() > 0 {
				v, _ := new(big.Rat).SetString(*raw)
				ratio.Mul(new(big.Rat).Quo(v, max), big.NewRat(100, 1))
			}
			// Shared visual bars floor to hundredths of a percentage point. This
			// affects CSS geometry only; source/model/title fractions stay exact.
			quantum := new(big.Int).Quo(new(big.Int).Mul(ratio.Num(), big.NewInt(100)), ratio.Denom())
			want, _ := new(big.Rat).SetFrac(quantum, big.NewInt(100)).Float64()
			if math.Abs(actual-want) > 0.0001 || math.Abs(*c.FillWidth-c.TrackWidth*want/100) > 1.1 {
				return false
			}
		}
	}
	return true
}
func lpcLivePairCheck(pair lpcLivePair, environment string, records []lpLiveRecord) bool {
	return lpcLivePairSources(pair.Request, pair.Baseline, pair.Selection, environment, records) && lpcLiveAPI(pair.Request, pair.Response, environment, records, false)
}
func lpcClockAtOrAfter(later, earlier string) bool {
	a, e := time.Parse(time.RFC3339Nano, later)
	if e != nil {
		return false
	}
	b, e := time.Parse(time.RFC3339Nano, earlier)
	return e == nil && !a.Before(b)
}
func lpcParentRetained(child, parent *exploration.Predicate) bool {
	if parent == nil {
		return true
	}
	if reflect.DeepEqual(child, parent) {
		return true
	}
	if child == nil {
		return false
	}
	for i := range child.All {
		if reflect.DeepEqual(&child.All[i], parent) {
			return true
		}
	}
	if len(parent.All) > 0 && len(child.All) >= len(parent.All) {
		return reflect.DeepEqual(child.All[:len(parent.All)], parent.All)
	}
	return false
}
func lpcLiveMainScope(l lpcLiveMain) bool {
	if l.Version != 1 || l.Current.Question.Signal != "logs" || l.Selector.Signal != "logs" || l.Selector.Environment != l.Environment {
		return false
	}
	current := exploration.Query{Signal: "logs", Environment: l.Environment, FromMS: l.Windows.Current.FromMS, ToMS: l.Windows.Current.ToMS, Where: l.Selector.Where, Limit: 100, GroupBy: []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}, BucketMS: 1000}
	if !nhLiveSameQuery(current, l.Current.Question) || !nhLiveSameQuery(current, l.Previous.Request.Selection.Question) {
		return false
	}
	previous := current
	previous.ToMS = current.FromMS
	previous.FromMS -= current.ToMS - current.FromMS
	if !nhLiveSameQuery(previous, l.Previous.Request.Baseline.Question) || previous.FromMS != l.Windows.Previous.FromMS || previous.ToMS != l.Windows.Previous.ToMS {
		return false
	}
	baseline := lpcProfileQuestion(l.SavedQuestion)
	if !nhLiveSameQuery(baseline, l.SavedPair.Request.Baseline.Question) || !nhLiveSameQuery(current, l.SavedPair.Request.Selection.Question) || !nhLiveSameQuery(l.Reused.Request.Baseline.Question, baseline) || !nhLiveSameQuery(l.Reused.Request.Selection.Question, current) {
		return false
	}
	if !nhLiveSameQuery(l.SideOpen.Parent, previous) || l.SideOpen.Side != "baseline" || !lpCanonical(l.SideOpen.ID) || !lpcParentRetained(l.SideOpen.Question.Where, previous.Where) || l.SideOpen.Question.Signal != previous.Signal || l.SideOpen.Question.Environment != previous.Environment || l.SideOpen.Question.FromMS != previous.FromMS || l.SideOpen.Question.ToMS != previous.ToMS || l.SideOpen.Question.BucketMS != previous.BucketMS || len(l.SideOpen.Question.GroupBy) != 0 || !nhLiveSameQuery(l.SideOpen.Question, l.SideOpen.Result.Question) {
		return false
	}
	wantSide := previous
	wantSide.Where = l.SideOpen.Question.Where
	wantSide.GroupBy = nil
	wantSide.GroupTimeSeries = false
	if !nhLiveSameQuery(wantSide, l.SideOpen.Question) {
		return false
	}
	rows, e := lpLiveRows(previous, l.Environment, l.Records)
	if e != nil {
		return false
	}
	wanted := map[string]bool{}
	for _, r := range rows {
		if lpOracle(r.Body).Identity == l.SideOpen.ID {
			wanted[r.ID] = true
		}
	}
	actual, e := lpLiveRows(l.SideOpen.Question, l.Environment, l.Records)
	if e != nil || len(actual) != len(wanted) {
		return false
	}
	for _, r := range actual {
		if !wanted[r.ID] {
			return false
		}
	}
	return true
}
func lpcLiveScopeCustody(l lpcLiveMain) bool {
	before := sha256.Sum256(l.SourceBytes)
	if !bytes.Equal(l.SourceBytes, l.AfterBytes) || !aeJSONEqual(l.ScopeCapture, l.LoadedScope) || !aeJSONEqual(l.ScopeCapture, l.ScopeCaptureAfter) || l.Custody.SHA256 != hex.EncodeToString(before[:]) {
		return false
	}
	if !lpcLiveOriginalStored(l) || !nhLiveSameQuery(l.LateReadyProbe.Question, l.SavedQuestion) || !lpcClockAtOrAfter(l.LateReadyProbe.CapturedAt, l.SourceStored.Explore.CapturedAt) {
		return false
	}
	all := append(append([]lpLiveRecord{}, l.Records...), l.LateRecords...)
	if !lpLiveSource(l.LateReadyProbe, l.Environment, all) || !lpcClockAtOrAfter(l.SavedPair.Baseline.CapturedAt, l.LateReadyProbe.CapturedAt) || !lpcClockAtOrAfter(l.SavedPair.Selection.CapturedAt, l.LateReadyProbe.CapturedAt) {
		return false
	}
	return true
}
func lpcLiveFrozen(l lpcLiveMain) bool {
	if l.Saved.Receipt.Version != 1 || l.Saved.Receipt.Question.Kind != "comparison" || l.Saved.Receipt.Question.Comparison == nil || l.Saved.Receipt.Question.Comparison.Distribution || l.Saved.Comparison == nil || l.Saved.Comparison.Result == nil || !peLiveCaptureHash(l.Saved) || !aeJSONEqual(l.Saved, l.Reopened) {
		return false
	}
	all := append(append([]lpLiveRecord{}, l.Records...), l.LateRecords...)
	p := l.Saved.Comparison
	if !lpcLiveAPI(p.Request, *p.Result, l.Environment, all, false) || !nhLiveSameQuery(p.Request.Baseline.Question, l.SavedPair.Request.Baseline.Question) || !nhLiveSameQuery(p.Request.Selection.Question, l.SavedPair.Request.Selection.Question) || !nhLiveSameQuery(l.Saved.Receipt.Question.Comparison.Baseline, p.Request.Baseline.Question) || !nhLiveSameQuery(l.Saved.Receipt.Question.Comparison.Selection, p.Request.Selection.Question) {
		return false
	}
	for i, page := range []exploration.CapturePage{l.CaptureSources.Baseline, l.CaptureSources.Selection} {
		id := []string{p.BaselineID, p.SelectionID}[i]
		cohort := []exploration.Cohort{p.Request.Baseline, p.Request.Selection}[i]
		prior := []exploration.Result{l.SavedPair.Baseline, l.SavedPair.Selection}[i]
		if page.Receipt.ID != id || page.Explore == nil || page.Receipt.Question.Explore == nil || !nhLiveSameQuery(*page.Receipt.Question.Explore, cohort.Question) || !lpLiveSource(*page.Explore, l.Environment, all) || !lpcClockAtOrAfter(page.Explore.CapturedAt, prior.CapturedAt) || !lpcClockAtOrAfter(page.Receipt.CapturedAt, page.Explore.CapturedAt) || !lpcClockAtOrAfter(l.Saved.Receipt.CapturedAt, page.Receipt.CapturedAt) {
			return false
		}
		actual, e := decLiveCohort(*page.Explore)
		if e != nil || !reflect.DeepEqual(actual, cohort) {
			return false
		}
	}
	return lpcLiveViewCheck(l.FrozenView, p.Request, all)
}
func lpcLiveReusable(l lpcLiveMain) bool {
	v := l.Reusable
	if v.Version != 1 || len(v.Queries) != 1 || !nhLiveSameQuery(v.Queries[0], l.SavedPair.Request.Selection.Question) || v.Breakdowns["logs"].Baseline == nil || !nhLiveSameQuery(*v.Breakdowns["logs"].Baseline, l.SavedPair.Request.Baseline.Question) {
		return false
	}
	id := v.ID
	v.ID = ""
	raw, _ := json.Marshal(v)
	hash := sha256.Sum256(raw)
	return id == hex.EncodeToString(hash[:])
}
func lpcLiveMainRequests(l lpcLiveMain) bool {
	expected := []exploration.Query{l.Current.Question, l.Previous.Baseline.Question, l.Previous.Selection.Question, l.SideOpen.Question, l.SavedPair.Baseline.Question, l.SavedPair.Selection.Question, l.Reused.Baseline.Question, l.Reused.Selection.Question}
	index, captures, protocols, writes, gets := 0, 0, 0, 0, 0
	for _, r := range l.Requests {
		switch r.Effect {
		case "stored-scope":
			gets++
			allowed := false
			for _, id := range []string{l.ScopeCapture.Receipt.ID, l.Saved.Receipt.ID, l.CaptureSources.Baseline.Receipt.ID, l.CaptureSources.Selection.Receipt.ID} {
				if r.Endpoint == "/api/captures/"+id || r.Endpoint == "/api/captures/"+id+"/receipt" {
					allowed = true
				}
			}
			if !allowed {
				return false
			}
		case "source":
			if r.Endpoint == "/api/captures" {
				var in struct {
					Title    string
					Question exploration.CaptureQuestion
				}
				if json.Unmarshal(r.Payload, &in) != nil || in.Title != l.Saved.Receipt.Title || !reflect.DeepEqual(in.Question, l.Saved.Receipt.Question) {
					return false
				}
				captures++
				continue
			}
			var q exploration.Query
			if r.Endpoint != "/api/explore" || index >= len(expected) || json.Unmarshal(r.Payload, &q) != nil {
				return false
			}
			for _, start := range []int{1, 4, 6} {
				if index == start && nhLiveSameQuery(q, expected[start+1]) {
					expected[start], expected[start+1] = expected[start+1], expected[start]
				}
			}
			if !nhLiveSameQuery(q, expected[index]) {
				return false
			}
			index++
		case "comparison":
			var in exploration.ComparisonRequest
			if r.Endpoint != "/api/compare" || json.Unmarshal(r.Payload, &in) != nil || protocols >= 3 || !reflect.DeepEqual(in, []exploration.ComparisonRequest{l.Previous.Request, l.SavedPair.Request, l.Reused.Request}[protocols]) {
				return false
			}
			protocols++
		case "configuration":
			var in workbench.ExplorationView
			expected := l.Reusable
			expected.ID = ""
			if r.Endpoint != "/api/exploration-views" || json.Unmarshal(r.Payload, &in) != nil || !reflect.DeepEqual(in, expected) {
				return false
			}
			writes++
		default:
			return false
		}
	}
	if index != 8 || captures != 1 || protocols != 3 || writes != 1 || len(l.Events) == 0 || l.Events[len(l.Events)-1].SourceOperations != 9 || l.Events[len(l.Events)-1].Protocols != 3 || l.Events[len(l.Events)-1].Writes != 1 || l.Events[len(l.Events)-1].CaptureGets != gets || len(l.SourceResponses) != 8 {
		return false
	}
	for _, s := range l.SourceResponses {
		all := append(append([]lpLiveRecord{}, l.Records...), l.LateRecords...)
		if s.Status != 200 || !nhLiveSameQuery(s.Request, s.Response.Question) || !lpLiveSource(s.Response, l.Environment, all) {
			return false
		}
	}
	return true
}
func lpcLiveMainSemantic(l lpcLiveMain) bool {
	all := append(append([]lpLiveRecord{}, l.Records...), l.LateRecords...)
	if !lpcLiveMainScope(l) || !lpcLiveScopeCustody(l) || !lpcLiveFrozen(l) || !lpcLiveReusable(l) || !lpcLiveMainRequests(l) || !lpLiveSource(l.Current, l.Environment, l.Records) || !lpLiveSource(l.SideOpen.Result, l.Environment, l.Records) || !lpcLivePairCheck(l.Previous, l.Environment, l.Records) || !lpcLivePairCheck(l.SavedPair, l.Environment, all) || !lpcLivePairCheck(l.Reused, l.Environment, all) {
		return false
	}
	if len(l.PreviousViews) != 3 || len(l.SavedViews) != 3 {
		return false
	}
	for i, v := range l.PreviousViews {
		if v.Metric != []string{"rate", "count", "share"}[i] || !lpcLiveViewCheck(v, l.Previous.Request, l.Records) {
			return false
		}
	}
	for key, v := range l.SavedViews {
		if key != v.Metric || !lpcLiveViewCheck(v, l.SavedPair.Request, all) {
			return false
		}
	}
	rate, ok := l.Geometry["rate"]
	if !ok || rate.Baseline.FillWidth == nil || rate.Current.FillWidth == nil || rate.Baseline.Value != "8/3" || rate.Current.Value != "8/3" || math.Abs(*rate.Baseline.FillWidth-*rate.Current.FillWidth) > 1 {
		return false
	}
	count, ok := l.Geometry["count"]
	if !ok || count.Baseline.Value != "320" || count.Current.Value != "160" || count.Baseline.FillPercent == nil || count.Current.FillPercent == nil {
		return false
	}
	a, e := strconv.ParseFloat(strings.TrimSuffix(*count.Baseline.FillPercent, "%"), 64)
	if e != nil {
		return false
	}
	z, e := strconv.ParseFloat(strings.TrimSuffix(*count.Current.FillPercent, "%"), 64)
	if e != nil || math.Abs(a-2*z) > 0.0001 {
		return false
	}
	if len(l.Mobile) != len(l.SavedViews["rate"].Rows) {
		return false
	}
	mobileRows := map[string]lpcLiveModelRow{}
	for _, r := range l.SavedViews["rate"].Model.Rows {
		mobileRows[r.ID] = r
	}
	seenMobile := map[string]bool{}
	for _, row := range l.Mobile {
		expected, ok := mobileRows[row.ID]
		if !ok || seenMobile[row.ID] || len(row.Cells) != 2 {
			return false
		}
		seenMobile[row.ID] = true
		for i, c := range row.Cells {
			value := []*string{expected.BaselineRate, expected.CurrentRate}[i]
			title := "Unavailable"
			if value != nil {
				title = *value + " logs/s (exact)"
			}
			if !c.NoTextClip || c.Left < 0 || c.Right > c.Viewport || c.Right <= c.Left || c.Exact != title {
				return false
			}
		}
	}
	return true
}
func lpcLiveReplay(t *testing.T, events []lpcLiveEvent, model string) bool {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, e := range events {
		rec.Record(context.Background(), strings.ReplaceAll(e.Action, "-", "_"), nil, "obs.coherent", true, "obs.sourceOperations", e.SourceOperations, "obs.protocols", e.Protocols, "obs.writes", e.Writes, "obs.captureGets", e.CaptureGets)
	}
	checker, e := validate.NewChecker(model, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if e != nil {
		t.Fatal(e)
	}
	result, e := checker.Check(context.Background(), sink.Steps())
	if e != nil {
		t.Fatal(e)
	}
	return result.OK()
}
func TestQuintLogPatternTimeLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_LOG_PATTERN_TIME_LEDGER")
	if file == "" {
		t.Skip("requires actual across-time log-pattern source ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base lpcLiveMain
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	base.SourceBytes, e = os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.json"))
	if e != nil {
		t.Fatal(e)
	}
	base.AfterBytes, e = os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.after.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(base.SourceBytes, &base.SourceStored); e != nil {
		t.Fatal(e)
	}
	if !lpcLiveMainSemantic(base) {
		t.Fatalf("independent across-time evidence rejected scope=%v custody=%v frozen=%v config=%v effects=%v previous=%v saved=%v", lpcLiveMainScope(base), lpcLiveScopeCustody(base), lpcLiveFrozen(base), lpcLiveReusable(base), lpcLiveMainRequests(base), lpcLivePairCheck(base.Previous, base.Environment, base.Records), lpcLivePairCheck(base.SavedPair, base.Environment, append(base.Records, base.LateRecords...)))
	}
	if !lpcLiveReplay(t, base.Events, "model/log_pattern_time.yaml") {
		t.Fatal("actual across-time lifecycle rejected")
	}

	controls := []struct {
		name   string
		change func(*lpcLiveMain)
	}{
		{"current-window", func(l *lpcLiveMain) { l.Current.Question.ToMS++ }},
		{"previous-adjacency", func(l *lpcLiveMain) { l.Previous.Request.Baseline.Question.ToMS++ }},
		{"saved-original-selector", func(l *lpcLiveMain) { l.SavedQuestion.Where = nil }},
		{"saved-width", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) { v.Model.BaselineWidthMS = v.Model.CurrentWidthMS })
		}},
		{"wrong-rate-rational", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) { v.Model.Rows[0].BaselineRate = lpcPointer("1234567/19") })
		}},
		{"rate-replaced-by-count", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) { v.Model.Rows[0].BaselineRate = v.Model.Rows[0].BaselineCount })
		}},
		{"share-wrong-denominator", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) { v.Model.Rows[0].BaselineShare = lpcPointer("100") })
		}},
		{"negative-sign-lost", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) {
				for i, r := range v.Model.Rows {
					if r.RateDifference != nil && strings.HasPrefix(*r.RateDifference, "-") {
						v.Model.Rows[i].RateDifference = lpcPointer(strings.TrimPrefix(*r.RateDifference, "-"))
						return
					}
				}
			})
		}},
		{"missing-union-id", func(l *lpcLiveMain) { lpcAlterView(l, func(v *lpcLiveView) { v.Model.Rows = v.Model.Rows[1:] }) }},
		{"duplicate-union-id", func(l *lpcLiveMain) { lpcAlterView(l, func(v *lpcLiveView) { v.Model.Rows[1] = v.Model.Rows[0] }) }},
		{"unclassified-omitted", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) {
				for i, r := range v.Model.Rows {
					if r.ID == "lp1:o" {
						v.Model.Rows = append(v.Model.Rows[:i], v.Model.Rows[i+1:]...)
						return
					}
				}
			})
		}},
		{"false-absence", func(l *lpcLiveMain) { lpcAlterView(l, func(v *lpcLiveView) { v.Model.Rows[0].Presence = "unknown" }) }},
		{"wrong-side", func(l *lpcLiveMain) { l.SideOpen.Side = "selection" }},
		{"side-lost-parent", func(l *lpcLiveMain) { l.SideOpen.Question.Where = nil }},
		{"side-wrong-window", func(l *lpcLiveMain) { l.SideOpen.Question.FromMS++ }},
		{"late-probe-old-count", func(l *lpcLiveMain) { l.LateReadyProbe.Count = "400" }},
		{"frozen-count-used-for-reread", func(l *lpcLiveMain) { l.SavedPair.Request.Baseline.Count = "400" }},
		{"loaded-new-evidence-instead-of-scope", func(l *lpcLiveMain) { l.LoadedScope = l.CaptureSources.Baseline }},
		{"stored-duplicate-row", func(l *lpcLiveMain) { l.SourceStored.Explore.Rows[1] = l.SourceStored.Explore.Rows[0] }},
		{"stored-raw-bytes", func(l *lpcLiveMain) { l.SourceBytes[0] = ' ' }},
		{"stored-after-bytes", func(l *lpcLiveMain) { l.AfterBytes = append(l.AfterBytes, ' ') }},
		{"custody-id", func(l *lpcLiveMain) { l.Custody.ReceiptID = "wrong" }},
		{"custody-sha", func(l *lpcLiveMain) { l.Custody.SHA256 = "wrong" }},
		{"comparison-capture-id", func(l *lpcLiveMain) { l.Saved.Receipt.ID = "wrong" }},
		{"comparison-capture-count", func(l *lpcLiveMain) { l.Saved.Comparison.Result.BaselineCount = "400" }},
		{"frozen-view-rate", func(l *lpcLiveMain) { l.FrozenView.Model.Rows[0].BaselineRate = lpcPointer("0") }},
		{"fresh-clock-is-old-scope-clock", func(l *lpcLiveMain) { l.CaptureSources.Baseline.Explore.CapturedAt = l.SourceStored.Explore.CapturedAt }},
		{"reopened-clock", func(l *lpcLiveMain) { l.Reopened.Receipt.CapturedAt = "wrong" }},
		{"config-baseline-window", func(l *lpcLiveMain) { l.Reusable.Breakdowns["logs"].Baseline.FromMS++ }},
		{"config-current-window", func(l *lpcLiveMain) { l.Reusable.Queries[0].ToMS++ }},
		{"source-body-corruption", func(l *lpcLiveMain) {
			lpLiveMutateRow(&l.SourceResponses[0].Response.Rows[0], "Body", "wrong raw body")
		}},
		{"duplicate-current-row", func(l *lpcLiveMain) { l.Current.Rows[1] = l.Current.Rows[0] }},
		{"separate-side-scale", func(l *lpcLiveMain) { lpcAlterView(l, func(v *lpcLiveView) { v.Rows[0].Cells[1].Max = "1" }) }},
		{"bar-side-metadata", func(l *lpcLiveMain) { lpcAlterView(l, func(v *lpcLiveView) { v.Rows[0].Cells[0].Side = "selection" }) }},
		{"equal-rate-geometry", func(l *lpcLiveMain) { v := l.Geometry["rate"]; v.Current.Value = "0"; l.Geometry["rate"] = v }},
		{"count-geometry-ratio", func(l *lpcLiveMain) {
			v := l.Geometry["count"]
			v.Current.FillPercent = lpcPointer("99%")
			l.Geometry["count"] = v
		}},
		{"mobile-clipped", func(l *lpcLiveMain) { l.Mobile[0].Cells[0].NoTextClip = false }},
		{"mobile-rounded-exact-title", func(l *lpcLiveMain) { l.Mobile[0].Cells[0].Exact = "2.67 logs/s (exact)" }},
		{"source-endpoint", func(l *lpcLiveMain) { l.Requests[0].Endpoint = "/api/inventory" }},
		{"late-raw-shape", func(l *lpcLiveMain) { l.LateRecords[0].Body += " literal-corrupt" }},

		{"new-pattern-baseline-not-zero", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) {
				for i, r := range v.Model.Rows {
					if r.Presence == "not-observed-in-baseline" {
						v.Model.Rows[i].BaselineCount = lpcPointer("1")
						return
					}
				}
			})
		}},
		{"absent-known-count-made-unavailable", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) {
				for i, r := range v.Model.Rows {
					if r.Presence == "not-observed-in-current" {
						v.Model.Rows[i].CurrentCount = nil
						return
					}
				}
			})
		}},
		{"oversize-claimed-as-shape", func(l *lpcLiveMain) {
			lpcAlterView(l, func(v *lpcLiveView) {
				for i, r := range v.Model.Rows {
					if r.ID == "lp1:o" {
						v.Model.Rows[i].Classified = true
						return
					}
				}
			})
		}},
	}
	for _, action := range []string{"previous-stage", "previous-local-metrics", "side-stage", "side-back", "parent-return", "saved-scope-load", "saved-stage", "saved-local-metrics", "view-save", "capture-reopen", "view-reopen", "original-unchanged"} {
		a := action
		controls = append(controls, struct {
			name   string
			change func(*lpcLiveMain)
		}{"hidden-read-" + a, func(l *lpcLiveMain) {
			for i := range l.Events {
				if l.Events[i].Action == a {
					l.Events[i].SourceOperations++
					return
				}
			}
		}})
	}
	for _, control := range controls {
		t.Run("reject-"+control.name, func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l lpcLiveMain
			json.Unmarshal(before, &l)
			control.change(&l)
			after, _ := json.Marshal(l)
			if bytes.Equal(before, after) {
				t.Fatal("negative control did not change serialized evidence")
			}
			if lpcLiveMainSemantic(l) && lpcLiveReplay(t, l.Events, "model/log_pattern_time.yaml") {
				t.Fatal("corrupted across-time evidence accepted")
			}
		})
	}
}

func lpcPointer(value string) *string { return &value }
func lpcAlterView(l *lpcLiveMain, change func(*lpcLiveView)) {
	v := l.SavedViews["rate"]
	change(&v)
	l.SavedViews["rate"] = v
}

type lpcLiveEmpty struct {
	Version         int
	Environment     string
	Records         []lpLiveRecord
	Current         exploration.Result
	Previous        exploration.Query
	ReadyProbes     []exploration.Result
	Pair            lpcLivePair
	Views           map[string]lpcLiveView
	Events          []lpcLiveEvent
	Requests        []lpcLiveRequest
	SourceResponses []struct {
		Request  exploration.Query
		Status   int
		Response exploration.Result
	}
}

func lpcLiveEmptySemantic(l lpcLiveEmpty) bool {
	if l.Version != 1 || !nhLiveSameQuery(l.Current.Question, l.Previous) || !lpLiveSource(l.Current, l.Environment, l.Records) || !lpcLivePairCheck(l.Pair, l.Environment, l.Records) || l.Pair.Request.Baseline.Count != "0" || l.Pair.Request.Baseline.State != "no_data" || !nhLiveSameQuery(l.Current.Question, l.Pair.Request.Selection.Question) || len(l.Views) != 3 {
		return false
	}
	q := l.Current.Question
	width := q.ToMS - q.FromMS
	q.ToMS = q.FromMS
	q.FromMS -= width
	if !nhLiveSameQuery(q, l.Pair.Request.Baseline.Question) {
		return false
	}
	for k, v := range l.Views {
		if k != v.Metric || !lpcLiveViewCheck(v, l.Pair.Request, l.Records) {
			return false
		}
	}
	expected := []exploration.Query{l.Current.Question, l.Pair.Baseline.Question, l.Pair.Selection.Question}
	index, protocols := 0, 0
	for _, r := range l.Requests {
		switch r.Effect {
		case "source":
			var q exploration.Query
			if r.Endpoint != "/api/explore" || index >= 3 || json.Unmarshal(r.Payload, &q) != nil {
				return false
			}
			if index == 1 && nhLiveSameQuery(q, expected[2]) {
				expected[1], expected[2] = expected[2], expected[1]
			}
			if !nhLiveSameQuery(q, expected[index]) {
				return false
			}
			index++
		case "comparison":
			var request exploration.ComparisonRequest
			if r.Endpoint != "/api/compare" || json.Unmarshal(r.Payload, &request) != nil || !reflect.DeepEqual(request, l.Pair.Request) {
				return false
			}
			protocols++
		default:
			return false
		}
	}
	if index != 3 || protocols != 1 || len(l.SourceResponses) != 3 {
		return false
	}
	for _, r := range l.SourceResponses {
		if r.Status != 200 || !nhLiveSameQuery(r.Request, r.Response.Question) || !lpLiveSource(r.Response, l.Environment, l.Records) {
			return false
		}
	}
	return true
}
func TestQuintLogPatternTimeEmptyLedger(t *testing.T) {
	file := os.Getenv("GOO_LOG_PATTERN_TIME_EMPTY_LEDGER")
	if file == "" {
		t.Skip("requires actual known-empty across-time source ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base lpcLiveEmpty
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	if !lpcLiveEmptySemantic(base) || !lpcLiveReplay(t, base.Events, "model/log_pattern_time_empty.yaml") {
		t.Fatal("actual known-empty evidence rejected")
	}
	for i := 0; i < 18; i++ {
		n := i
		t.Run(fmt.Sprintf("reject-empty-%02d", n), func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l lpcLiveEmpty
			json.Unmarshal(before, &l)
			switch n {
			case 0:
				v := l.Views["rate"]
				v.Model.Rows[0].BaselineRate = nil
				l.Views["rate"] = v
			case 1:
				v := l.Views["share"]
				v.Model.Rows[0].BaselineShare = lpcPointer("0")
				l.Views["share"] = v
			case 2:
				v := l.Views["share"]
				v.Model.Rows[0].ShareDifference = lpcPointer("0")
				l.Views["share"] = v
			case 3:
				v := l.Views["rate"]
				v.Model.Rows[0].BaselineCount = nil
				l.Views["rate"] = v
			case 4:
				v := l.Views["rate"]
				v.Model.Rows[0].Presence = "unknown"
				l.Views["rate"] = v
			case 5:
				v := l.Views["rate"]
				v.Model.Rows = v.Model.Rows[1:]
				l.Views["rate"] = v
			case 6:
				l.Pair.Request.Baseline.Count = "1"
			case 7:
				l.Pair.Request.Baseline.State = "changed_between_reads"
			case 8:
				l.Pair.Response.Rows = []exploration.ValueComparison{{Value: lpcPointer("lp1:e"), BaselineCount: "0", SelectionCount: "0"}}
			case 9:
				l.Pair.Baseline.Question.ToMS++
			case 10:
				v := l.Views["share"]
				v.Rows[0].Cells[0].Exact = "0 % (exact)"
				l.Views["share"] = v
			case 11:
				v := l.Views["rate"]
				v.Rows[0].Cells[0].Value = "unavailable"
				l.Views["rate"] = v
			case 12:
				l.Requests[0].Endpoint = "/api/inventory"
			case 13:
				l.Events[1].SourceOperations++
			case 14:
				l.Events[2].SourceOperations++
			case 15:
				l.Events[2].Protocols++
			case 16:
				l.Events[1].CaptureGets++
			case 17:
				l.Pair.Response.Basis = "wrong"
			}
			after, _ := json.Marshal(l)
			if bytes.Equal(before, after) {
				t.Fatal("negative control did not change evidence")
			}
			if lpcLiveEmptySemantic(l) && lpcLiveReplay(t, l.Events, "model/log_pattern_time_empty.yaml") {
				t.Fatal("zero/unavailable conflation accepted")
			}
		})
	}
}

type lpcLiveOwnership struct {
	Version      int
	Environment  string
	Records      []lpLiveRecord
	Current      exploration.Result
	Previous     exploration.Query
	Recovered    lpcLivePair
	HeldResponse struct {
		Request        exploration.Query
		Status         int
		Response       exploration.Result
		TransportFault string
	}
	Late struct {
		Request  exploration.Query
		Response exploration.Result
	}
	Observations []struct {
		Action                                                        string
		ComparisonSections, AbsentLabels, SourceOperations, Protocols int
		Before                                                        struct {
			URL       string
			Stages    int
			Connected bool
		}
		After struct {
			URL    string
			Stages int
		}
	}
	Events          []lpcLiveEvent
	Requests        []lpcLiveRequest
	SourceResponses []struct {
		Request  exploration.Query
		Status   int
		Response exploration.Result
	}
	FailedRequests []struct {
		Endpoint string
		Request  exploration.Query
		Failure  struct{ ErrorText string }
	}
}

func lpcLiveOwnershipSemantic(l lpcLiveOwnership) bool {
	if l.Version != 1 || !lpLiveSource(l.Current, l.Environment, l.Records) || !lpcLivePairCheck(l.Recovered, l.Environment, l.Records) || !nhLiveSameQuery(l.Recovered.Request.Selection.Question, l.Current.Question) || !nhLiveSameQuery(l.Recovered.Request.Baseline.Question, l.Previous) || len(l.Observations) != 3 || len(l.FailedRequests) != 2 || len(l.SourceResponses) != 5 {
		return false
	}
	if l.HeldResponse.Status != 200 || l.HeldResponse.TransportFault == "" || !nhLiveSameQuery(l.HeldResponse.Request, l.Previous) || !nhLiveSameQuery(l.HeldResponse.Response.Question, l.Previous) || !lpLiveSource(l.HeldResponse.Response, l.Environment, l.Records) || !nhLiveSameQuery(l.Late.Request, l.Previous) || !nhLiveSameQuery(l.Late.Response.Question, l.Previous) || !lpLiveSource(l.Late.Response, l.Environment, l.Records) || !lpcClockAtOrAfter(l.Recovered.Baseline.CapturedAt, l.Late.Response.CapturedAt) || !lpcClockAtOrAfter(l.Recovered.Selection.CapturedAt, l.Late.Response.CapturedAt) {
		return false
	}
	for i, o := range l.Observations {
		if o.Action != []string{"failed-baseline", "held-return", "stale-side"}[i] || o.SourceOperations != []int{3, 5, 7}[i] {
			return false
		}
		if i < 2 && (o.ComparisonSections != 0 || o.AbsentLabels != 0 || o.Protocols != 0) {
			return false
		}
		if i == 2 && (!o.Before.Connected || !strings.HasSuffix(o.Before.URL, "#definitions") || o.Before.URL != o.After.URL || o.Before.Stages != o.After.Stages) {
			return false
		}
	}
	failed, canceled := false, false
	for _, r := range l.FailedRequests {
		if r.Endpoint != "/api/explore" || !nhLiveSameQuery(r.Request, l.Previous) {
			return false
		}
		failed = failed || strings.Contains(r.Failure.ErrorText, "ERR_FAILED")
		canceled = canceled || strings.Contains(r.Failure.ErrorText, "ERR_ABORTED")
	}
	if !failed || !canceled {
		return false
	}
	expected := []exploration.Query{l.Current.Question, l.Current.Question, l.Previous, l.Current.Question, l.Previous, l.Current.Question, l.Previous}
	index, protocols := 0, 0
	for _, r := range l.Requests {
		switch r.Effect {
		case "source":
			var q exploration.Query
			if r.Endpoint != "/api/explore" || index >= 7 || json.Unmarshal(r.Payload, &q) != nil {
				return false
			}
			for _, start := range []int{1, 3, 5} {
				if index == start && nhLiveSameQuery(q, expected[start+1]) {
					expected[start], expected[start+1] = expected[start+1], expected[start]
				}
			}
			if !nhLiveSameQuery(q, expected[index]) {
				return false
			}
			index++
		case "comparison":
			var request exploration.ComparisonRequest
			if r.Endpoint != "/api/compare" || json.Unmarshal(r.Payload, &request) != nil || !reflect.DeepEqual(request, l.Recovered.Request) || index != 7 {
				return false
			}
			protocols++
		default:
			return false
		}
	}
	if index != 7 || protocols != 1 {
		return false
	}
	sourceCurrent, sourcePrevious := 0, 0
	clocks := map[string]bool{}
	for _, r := range l.SourceResponses {
		if r.Status != 200 || !nhLiveSameQuery(r.Request, r.Response.Question) || !lpLiveSource(r.Response, l.Environment, l.Records) || clocks[r.Response.CapturedAt] {
			return false
		}
		clocks[r.Response.CapturedAt] = true
		if nhLiveSameQuery(r.Request, l.Current.Question) {
			sourceCurrent++
		} else if nhLiveSameQuery(r.Request, l.Previous) {
			sourcePrevious++
		} else {
			return false
		}
	}
	return sourceCurrent == 4 && sourcePrevious == 1
}
func TestQuintLogPatternTimeOwnershipLedger(t *testing.T) {
	file := os.Getenv("GOO_LOG_PATTERN_TIME_OWNERSHIP_LEDGER")
	if file == "" {
		t.Skip("requires actual failed/held across-time owner ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base lpcLiveOwnership
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	if !lpcLiveOwnershipSemantic(base) || !lpcLiveReplay(t, base.Events, "model/log_pattern_time_ownership.yaml") {
		t.Fatal("actual failed/held ownership rejected")
	}
	for i := 0; i < 20; i++ {
		n := i
		t.Run(fmt.Sprintf("reject-owner-%02d", n), func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l lpcLiveOwnership
			json.Unmarshal(before, &l)
			switch n {
			case 0:
				l.Observations[0].ComparisonSections = 1
			case 1:
				l.Observations[0].AbsentLabels = 1
			case 2:
				l.Observations[0].Protocols = 1
			case 3:
				l.Observations[1].ComparisonSections = 1
			case 4:
				l.Observations[1].Protocols = 1
			case 5:
				l.Observations[2].Before.Connected = false
			case 6:
				l.Observations[2].After.URL += "-stale"
			case 7:
				l.Observations[2].After.Stages++
			case 8:
				l.Observations[2].SourceOperations++
			case 9:
				l.HeldResponse.Status = 500
			case 10:
				l.HeldResponse.Response.Count = "0"
			case 11:
				l.Late.Request.FromMS++
			case 12:
				l.FailedRequests[0].Failure.ErrorText = "net::OK"
			case 13:
				l.Events[2].Protocols++
			case 14:
				l.Events[4].SourceOperations++
			case 15:
				l.Events[6].SourceOperations++
			case 16:
				l.Requests[0].Endpoint = "/api/inventory"
			case 17:
				l.Recovered.Request.Baseline.Question.ToMS++
			case 18:
				l.Recovered.Baseline.CapturedAt = l.HeldResponse.Response.CapturedAt
			case 19:
				l.SourceResponses[1] = l.SourceResponses[0]
			}
			after, _ := json.Marshal(l)
			if bytes.Equal(before, after) {
				t.Fatal("negative control did not change evidence")
			}
			if lpcLiveOwnershipSemantic(l) && lpcLiveReplay(t, l.Events, "model/log_pattern_time_ownership.yaml") {
				t.Fatal("failed read supplied zero/absence or stale owner republished evidence")
			}
		})
	}
}

func lpcLiveOriginalStored(l lpcLiveMain) bool {
	s, p, c := l.SourceStored, l.ScopeCapture, l.Custody
	hash := sha256.Sum256(l.SourceBytes)
	if c.File != "scope-capture.json" || c.Bytes != len(l.SourceBytes) || c.Bytes <= 0 || c.SHA256 != hex.EncodeToString(hash[:]) || c.ReceiptID != s.ID || c.Source != "read-only actual query-container capture store" || s.Explore == nil || s.Question.Explore == nil || p.Explore == nil || s.Question.Kind != "explore" || !reflect.DeepEqual(s.CaptureReceipt, p.Receipt) || !nhLiveSameQuery(*s.Question.Explore, l.SavedQuestion) || !peLiveCaptureHash(exploration.CapturePage{Receipt: s.CaptureReceipt, Explore: s.Explore}) || !lpcClockAtOrAfter(c.CopiedAt, s.CapturedAt) || !lpcClockAtOrAfter(c.CopiedAt, s.Explore.CapturedAt) {
		return false
	}
	rows, e := lpLiveRows(l.SavedQuestion, l.Environment, l.Records)
	if e != nil || len(s.Explore.Rows) != len(rows) || s.Explore.HasMore || p.RetainedRecords != len(rows) || p.RecordsLimited {
		return false
	}
	n := len(rows)
	if n > 100 {
		n = 100
	}
	if !aeJSONEqual(s.Explore.Rows[:n], p.Explore.Rows) || p.Explore.HasMore != (len(rows) > n) {
		return false
	}
	full := *s.Explore
	full.Question.Limit = len(rows)
	return lpLiveSource(full, l.Environment, l.Records) && lpLiveSource(*p.Explore, l.Environment, l.Records)
}

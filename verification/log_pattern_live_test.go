package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"math/big"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type lpLiveRecord struct {
	Environment, Signal, ID, TraceID, Service, Namespace, Body, Severity string
	TimeMS                                                               int64 `json:"timeMs"`
	Attributes, Resource                                                 map[string]string
}

func lpLiveField(f exploration.Field, r lpLiveRecord) (*string, error) {
	if f.Kind == "attribute" {
		v, ok := r.Attributes[f.Key]
		if !ok {
			return nil, nil
		}
		return &v, nil
	}
	if f.Kind == "resource" {
		v, ok := r.Resource[f.Key]
		if f.Key == "service.name" {
			v, ok = r.Service, true
		}
		if f.Key == "deployment.environment.name" {
			v, ok = r.Environment, true
		}
		if !ok {
			return nil, nil
		}
		return &v, nil
	}
	var v string
	switch f.Key {
	case "service":
		v = r.Service
	case "trace_id":
		v = r.TraceID
	case "span_id":
		v = r.ID
	case "body":
		v = r.Body
	case "severity":
		v = r.Severity
	case "body_pattern_v1":
		v = lpOracle(r.Body).Identity
	default:
		return nil, fmt.Errorf("unsupported fixture field %+v", f)
	}
	return &v, nil
}
func lpLiveMatch(p *exploration.Predicate, r lpLiveRecord) (bool, error) {
	if p == nil {
		return true, nil
	}
	if p.Numeric != nil {
		return false, fmt.Errorf("unexpected numeric pattern predicate")
	}
	if len(p.All) > 0 {
		for i := range p.All {
			yes, e := lpLiveMatch(&p.All[i], r)
			if e != nil || !yes {
				return false, e
			}
		}
		return true, nil
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			yes, e := lpLiveMatch(&p.Any[i], r)
			if e != nil {
				return false, e
			}
			if yes {
				return true, nil
			}
		}
		return false, nil
	}
	if p.Field == nil {
		return false, fmt.Errorf("missing fixture field")
	}
	v, e := lpLiveField(*p.Field, r)
	if e != nil {
		return false, e
	}
	switch p.Op {
	case "exists":
		return v != nil, nil
	case "missing":
		return v == nil, nil
	case "eq":
		return v != nil && *v == p.Value, nil
	case "ne":
		return v != nil && *v != p.Value, nil
	}
	return false, fmt.Errorf("unsupported fixture operator %s", p.Op)
}
func lpLiveRows(q exploration.Query, environment string, records []lpLiveRecord) ([]lpLiveRecord, error) {
	if q.Signal != "logs" || q.Environment != environment || q.ToMS <= q.FromMS {
		return nil, fmt.Errorf("invalid fixed source scope")
	}
	rows := []lpLiveRecord{}
	for _, r := range records {
		if r.Signal != "logs" || r.Environment != environment || r.TimeMS < q.FromMS || r.TimeMS >= q.ToMS {
			continue
		}
		yes, e := lpLiveMatch(q.Where, r)
		if e != nil {
			return nil, e
		}
		if yes {
			rows = append(rows, r)
		}
	}
	return rows, nil
}
func lpLiveSource(result exploration.Result, environment string, records []lpLiveRecord) bool {
	q := result.Question
	rows, e := lpLiveRows(q, environment, records)
	if e != nil || result.Count != strconv.Itoa(len(rows)) || !peLiveClock(result.CapturedAt) || !peLiveHex(result.Basis) || q.Numeric != nil || result.Numeric != nil {
		return false
	}
	state := "observed"
	if len(rows) == 0 {
		state = "no_data"
	}
	if result.State != state {
		return false
	}
	n := len(rows) - q.Offset
	if n < 0 {
		n = 0
	}
	if n > q.Limit {
		n = q.Limit
	}
	if len(result.Rows) != n || result.HasMore != (len(rows) > q.Offset+n) {
		return false
	}
	byID := map[string]lpLiveRecord{}
	for _, r := range rows {
		if _, exists := byID[r.ID]; exists {
			return false
		}
		byID[r.ID] = r
	}
	seen := map[string]bool{}
	for _, raw := range result.Rows {
		var row struct {
			Time       string `json:"timestamp_ns"`
			ID         string `json:"SpanId"`
			TraceID    string `json:"TraceId"`
			Service    string `json:"ServiceName"`
			Severity   string `json:"SeverityText"`
			Body       string
			Attributes map[string]string
			Resource   map[string]string `json:"ResourceAttributes"`
			Pattern    *string           `json:"body_pattern_v1"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return false
		}
		fact, ok := byID[row.ID]
		if !ok || seen[row.ID] || row.Time != strconv.FormatInt(fact.TimeMS*1000000, 10) || row.TraceID != fact.TraceID || row.Service != fact.Service || row.Severity != fact.Severity || row.Body != fact.Body || !reflect.DeepEqual(row.Attributes, fact.Attributes) {
			return false
		}
		seen[row.ID] = true
		expected := map[string]string{"service.name": fact.Service, "deployment.environment.name": fact.Environment}
		for k, v := range fact.Resource {
			expected[k] = v
		}
		if !reflect.DeepEqual(row.Resource, expected) {
			return false
		}
		if row.Pattern != nil && *row.Pattern != lpOracle(fact.Body).Identity {
			return false
		}
	}
	groups := map[string]int{}
	if len(q.GroupBy) > 0 {
		for _, r := range rows {
			values := []*string{}
			for _, f := range q.GroupBy {
				v, e := lpLiveField(f, r)
				if e != nil {
					return false
				}
				values = append(values, v)
			}
			key, _ := json.Marshal(values)
			groups[string(key)]++
		}
	}
	if len(groups) != len(result.Groups) {
		return false
	}
	seenGroups := map[string]bool{}
	for _, raw := range result.Groups {
		var group map[string]json.RawMessage
		if json.Unmarshal(raw, &group) != nil {
			return false
		}
		var count string
		if json.Unmarshal(group["count"], &count) != nil {
			return false
		}
		values := []*string{}
		for i := range q.GroupBy {
			var v *string
			if json.Unmarshal(group[fmt.Sprintf("g%d", i)], &v) != nil {
				return false
			}
			values = append(values, v)
		}
		key, _ := json.Marshal(values)
		expected, ok := groups[string(key)]
		if !ok || seenGroups[string(key)] || count != strconv.Itoa(expected) {
			return false
		}
		seenGroups[string(key)] = true
	}
	if q.BucketMS > 0 {
		expected := (q.ToMS - q.FromMS + q.BucketMS - 1) / q.BucketMS
		if int64(len(result.TimeSeries)) != expected {
			return false
		}
		for i, p := range result.TimeSeries {
			from := q.FromMS + int64(i)*q.BucketMS
			to := from + q.BucketMS
			if to > q.ToMS {
				to = q.ToMS
			}
			count := 0
			for _, r := range rows {
				if r.TimeMS >= from && r.TimeMS < to {
					count++
				}
			}
			state := "observed"
			if count == 0 {
				state = "no_data"
			}
			if p.FromMS != from || p.ToMS != to || p.Count != strconv.Itoa(count) || p.State != state {
				return false
			}
		}
	} else if len(result.TimeSeries) != 0 {
		return false
	}
	if !q.GroupTimeSeries && len(result.GroupedTimeSeries) != 0 {
		return false
	}
	return result.Plan.Count != "" && !strings.Contains(strings.ToUpper(result.Plan.Count), " LIMIT ") && result.Plan.Parameters["environment"] == environment && result.Plan.Parameters["from"] == strconv.FormatInt(q.FromMS, 10) && result.Plan.Parameters["to"] == strconv.FormatInt(q.ToMS, 10)
}

type lpLiveDisplay struct{ ID, State, Label, Count, Share string }
type lpLiveLedger struct {
	Version     int
	Environment string
	Records     []lpLiveRecord
	Selector    struct {
		Signal, Environment string
		Where               *exploration.Predicate
	}
	Applied struct {
		Selector struct {
			Signal, Environment string
			Where               *exploration.Predicate
		}
		ReadWindow struct{ FromMS, ToMS int64 }
	}
	Original, SelectedStage               exploration.Query
	Result, Selected, ParentAgain, Reused exploration.Result
	FrequentID                            string
	InitialRows, RareRows                 []lpLiveDisplay
	RawDetail                             json.RawMessage
	Comparison                            struct {
		Request  exploration.ComparisonRequest
		Response exploration.CohortComparison
	}
	Saved, Reopened exploration.CapturePage
	Stored          exploration.SavedCapture
	Custody         decLiveCustody
	StoredBytes     int
	StoredSHA       string
	YAML            string
	Draft           exploration.PackagePlan
	Reusable        exploration.SavedPackage
	Events          []cpLiveEvent
	Requests        []struct {
		Effect, Endpoint string
		Payload          json.RawMessage
	}
}

func lpLiveDisplayCheck(l lpLiveLedger, actual []lpLiveDisplay, least bool) bool {
	rows, e := lpLiveRows(l.Original, l.Environment, l.Records)
	if e != nil {
		return false
	}
	groups := map[string]int{}
	labels := map[string]exploration.LogPattern{}
	for _, r := range rows {
		p := lpOracle(r.Body)
		groups[p.Identity]++
		labels[p.Identity] = p
	}
	keys := []string{}
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, z := keys[i], keys[j]
		if groups[a] == groups[z] {
			return a < z
		}
		if least {
			return groups[a] < groups[z]
		}
		return groups[a] > groups[z]
	})
	if len(actual) != len(keys) {
		return false
	}
	for i, row := range actual {
		key := keys[i]
		p := labels[key]
		state := p.State
		if row.ID != key || row.State != state || row.Label != p.Label || row.Count != strconv.Itoa(groups[key]) || row.Share != row.Count+"/"+strconv.Itoa(len(rows))+" of all matching records (exact)" {
			return false
		}
	}
	return true
}
func lpLivePartition(l lpLiveLedger) bool {
	parent, e := lpLiveRows(l.Original, l.Environment, l.Records)
	if e != nil || !lpCanonical(l.FrequentID) || !strings.HasPrefix(l.FrequentID, "lp1:p:") {
		return false
	}
	selected, e := lpLiveRows(l.SelectedStage, l.Environment, l.Records)
	if e != nil {
		return false
	}
	want := map[string]bool{}
	for _, r := range parent {
		if lpOracle(r.Body).Identity == l.FrequentID {
			want[r.ID] = true
		}
	}
	if len(selected) != len(want) || len(selected) <= 100 {
		return false
	}
	for _, r := range selected {
		if !want[r.ID] {
			return false
		}
	}
	if !nhLiveSameQuery(l.Selected.Question, l.SelectedStage) || len(l.SelectedStage.GroupBy) != 0 || l.SelectedStage.FromMS != l.Original.FromMS || l.SelectedStage.ToMS != l.Original.ToMS || l.SelectedStage.Environment != l.Environment || l.SelectedStage.Signal != "logs" || l.SelectedStage.BucketMS != l.Original.BucketMS {
		return false
	}
	for i, c := range []exploration.Cohort{l.Comparison.Request.Baseline, l.Comparison.Request.Selection} {
		rows, e := lpLiveRows(c.Question, l.Environment, l.Records)
		if e != nil || c.Count != strconv.Itoa(len(rows)) || c.State != "observed" || !peLiveHex(c.Basis) || c.Numeric != nil || len(c.Question.GroupBy) != 1 || c.Question.GroupBy[0] != (exploration.Field{Kind: "column", Key: "service"}) || c.Question.FromMS != l.Original.FromMS || c.Question.ToMS != l.Original.ToMS {
			return false
		}
		expected := map[string]bool{}
		for _, r := range parent {
			if want[r.ID] == (i == 1) {
				expected[r.ID] = true
			}
		}
		if len(rows) != len(expected) {
			return false
		}
		groups := map[string]int{}
		for _, r := range rows {
			if !expected[r.ID] {
				return false
			}
			groups[r.Service]++
		}
		if len(c.Groups) != len(groups) {
			return false
		}
		seen := map[string]bool{}
		for _, g := range c.Groups {
			if g.Value == nil || seen[*g.Value] || g.Count != strconv.Itoa(groups[*g.Value]) || g.Numeric != nil {
				return false
			}
			seen[*g.Value] = true
		}
	}
	req, res := l.Comparison.Request, l.Comparison.Response
	if req.Numeric || req.Distribution || res.State != "ok" || res.BaselineCount != req.Baseline.Count || res.SelectionCount != req.Selection.Count || res.BaselineBasis != req.Baseline.Basis || res.SelectionBasis != req.Selection.Basis || res.Field != (exploration.Field{Kind: "column", Key: "service"}) {
		return false
	}
	canonical := req
	canonical.Baseline.Groups = append([]exploration.ValueCount{}, req.Baseline.Groups...)
	canonical.Selection.Groups = append([]exploration.ValueCount{}, req.Selection.Groups...)
	sortNumericGroups(canonical.Baseline.Groups)
	sortNumericGroups(canonical.Selection.Groups)
	raw, _ := json.Marshal(canonical)
	hash := sha256.Sum256(raw)
	if res.Basis != hex.EncodeToString(hash[:]) {
		return false
	}
	ag, zg := map[string]string{}, map[string]string{}
	keys := map[string]bool{}
	for _, g := range req.Baseline.Groups {
		k := explanationValueKey(g.Value)
		ag[k] = g.Count
		keys[k] = true
	}
	for _, g := range req.Selection.Groups {
		k := explanationValueKey(g.Value)
		zg[k] = g.Count
		keys[k] = true
	}
	if len(res.Rows) != len(keys) || len(keys) < 2 {
		return false
	}
	seen := map[string]bool{}
	previous := big.NewRat(1000, 1)
	previousKey := ""
	different := false
	for _, row := range res.Rows {
		k := explanationValueKey(row.Value)
		a, z := ag[k], zg[k]
		if a == "" {
			a = "0"
		}
		if z == "" {
			z = "0"
		}
		ap := new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(a), ncLiveInt(req.Baseline.Count)), big.NewRat(100, 1))
		zp := new(big.Rat).Mul(new(big.Rat).SetFrac(ncLiveInt(z), ncLiveInt(req.Selection.Count)), big.NewRat(100, 1))
		delta := new(big.Rat).Sub(zp, ap)
		display := delta.FloatString(2)
		if display == "-0.00" {
			display = "0.00"
		}
		score := new(big.Rat).Abs(delta)
		if score.Sign() > 0 {
			different = true
		}
		if !keys[k] || seen[k] || row.BaselineCount != a || row.SelectionCount != z || row.BaselinePercent != ap.FloatString(2) || row.SelectionPercent != zp.FloatString(2) || row.DifferencePP != display || previous.Cmp(score) < 0 || previous.Cmp(score) == 0 && previousKey > k {
			return false
		}
		seen[k] = true
		previous = score
		previousKey = k
	}
	return different
}
func lpLiveStored(l lpLiveLedger) bool {
	s, p := l.Stored, l.Saved
	if s.Explore == nil || s.Question.Explore == nil || p.Explore == nil || !reflect.DeepEqual(s.CaptureReceipt, p.Receipt) || !nhLiveSameQuery(*s.Question.Explore, l.Original) || !peLiveCaptureHash(exploration.CapturePage{Receipt: s.CaptureReceipt, Explore: s.Explore}) || !aeJSONEqual(p, l.Reopened) {
		return false
	}
	rows, e := lpLiveRows(l.Original, l.Environment, l.Records)
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
	return lpLiveSource(full, l.Environment, l.Records) && lpLiveSource(*p.Explore, l.Environment, l.Records) && decLiveCustodyCheck(decLiveLedger{Stored: l.Stored, Saved: l.Saved, Custody: l.Custody, StoredBytes: l.StoredBytes, StoredSHA: l.StoredSHA})
}
func lpLivePackage(l lpLiveLedger) bool {
	p := l.Reusable.Definition
	if p.Source.Signal != "logs" || p.Source.Environment != l.Environment || p.Source.Adapter != "clickhouse.otel/v1" || p.Numeric != nil || p.NumericHistogram != nil || p.Measure != "captured-record-count" || p.LookbackMinutes != 3 || !reflect.DeepEqual(p.Where, l.Original.Where) || !reflect.DeepEqual(p.GroupBy, l.Original.GroupBy) || p.BucketMS != l.Original.BucketMS || !reflect.DeepEqual(p, l.Draft.Definition) || l.Draft.Revision != l.Reusable.Revision || l.Draft.YAML != l.YAML {
		return false
	}
	raw, _ := json.Marshal(p)
	hash := sha256.Sum256(raw)
	if l.Reusable.Revision != hex.EncodeToString(hash[:]) {
		return false
	}
	q := l.Original
	q.FromMS = l.Reused.Question.FromMS
	q.ToMS = l.Reused.Question.ToMS
	return nhLiveSameQuery(q, l.Reused.Question) && q.ToMS-q.FromMS == 180000 && q.ToMS >= l.Original.ToMS && lpLiveSource(l.Reused, l.Environment, l.Records)
}
func lpLiveRequests(l lpLiveLedger) bool {
	expected := []exploration.Query{l.Result.Question, l.Selected.Question, l.Comparison.Request.Baseline.Question, l.Comparison.Request.Selection.Question, l.ParentAgain.Question, l.Reused.Question}
	index, captures, streams, protocols, writes := 0, 0, 0, 0, 0
	for _, r := range l.Requests {
		switch r.Effect {
		case "stream":
			var p struct {
				Environment, Signal string
				Where               *exploration.Predicate
			}
			if r.Endpoint != "/api/tail" || json.Unmarshal(r.Payload, &p) != nil || p.Environment != l.Environment || p.Signal != "" && p.Signal != "logs" || !reflect.DeepEqual(p.Where, l.Selector.Where) {
				return false
			}
			streams++
		case "source":
			if r.Endpoint == "/api/captures" {
				var p struct {
					Title    string
					Question exploration.CaptureQuestion
				}
				if json.Unmarshal(r.Payload, &p) != nil || p.Question.Explore == nil || p.Title != l.Saved.Receipt.Title || !nhLiveSameQuery(*p.Question.Explore, l.Original) || captures != 0 {
					return false
				}
				captures++
				continue
			}
			var q exploration.Query
			if r.Endpoint != "/api/explore" || index >= len(expected) || json.Unmarshal(r.Payload, &q) != nil {
				return false
			}
			if index == 2 && nhLiveSameQuery(q, expected[3]) {
				expected[2], expected[3] = expected[3], expected[2]
			}
			if !nhLiveSameQuery(q, expected[index]) {
				return false
			}
			index++
		case "comparison":
			var q exploration.ComparisonRequest
			if r.Endpoint != "/api/compare" || json.Unmarshal(r.Payload, &q) != nil || !reflect.DeepEqual(q, l.Comparison.Request) {
				return false
			}
			protocols++
		case "publication":
			var p struct{ YAML string }
			if r.Endpoint != "/api/telemetry-packages" || json.Unmarshal(r.Payload, &p) != nil || p.YAML != l.YAML {
				return false
			}
			writes++
		default:
			return false
		}
	}
	return index == 6 && captures == 1 && streams == 1 && protocols == 1 && writes == 1
}
func lpLiveSemantic(l lpLiveLedger) bool {
	if l.Version != 1 || l.Applied.Selector.Signal != "logs" || l.Applied.Selector.Environment != l.Environment || !reflect.DeepEqual(l.Applied.Selector.Where, l.Selector.Where) {
		return false
	}
	original := exploration.Query{Signal: "logs", Environment: l.Environment, Where: l.Selector.Where, FromMS: l.Applied.ReadWindow.FromMS, ToMS: l.Applied.ReadWindow.ToMS, GroupBy: []exploration.Field{{Kind: "column", Key: "body_pattern_v1"}}, BucketMS: 1000, Limit: 100}
	if !nhLiveSameQuery(original, l.Original) || !nhLiveSameQuery(l.Result.Question, l.Original) || !nhLiveSameQuery(l.ParentAgain.Question, l.Original) || !lpLiveSource(l.Result, l.Environment, l.Records) || !lpLiveSource(l.ParentAgain, l.Environment, l.Records) || !lpLiveSource(l.Selected, l.Environment, l.Records) || !lpLiveDisplayCheck(l, l.InitialRows, false) || !lpLiveDisplayCheck(l, l.RareRows, true) || !lpLivePartition(l) || !lpLiveStored(l) || !lpLivePackage(l) || !lpLiveRequests(l) {
		return false
	}
	found := false
	for _, r := range l.Result.Rows {
		if aeJSONEqual(r, l.RawDetail) {
			found = true
		}
	}
	return found
}
func lpLiveReplay(t *testing.T, l lpLiveLedger) bool {
	return lpLiveReplayModel(t, l.Events, "model/log_patterns.yaml")
}
func lpLiveReplayModel(t *testing.T, events []cpLiveEvent, model string) bool {
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, e := range events {
		rec.Record(context.Background(), strings.ReplaceAll(e.Action, "-", "_"), nil, "obs.coherent", true, "obs.reads", e.Reads, "obs.plans", e.Plans, "obs.protocols", e.Protocols, "obs.writes", e.Writes, "obs.streams", e.Streams)
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
func TestQuintLogPatternLiveLedger(t *testing.T) {
	file := os.Getenv("GOO_LOG_PATTERN_LEDGER")
	if file == "" {
		t.Skip("requires actual log-pattern journey source ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base lpLiveLedger
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	stored, e := os.ReadFile(filepath.Join(filepath.Dir(file), "scope-capture.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(stored, &base.Stored); e != nil {
		t.Fatal(e)
	}
	base.StoredBytes = len(stored)
	hash := sha256.Sum256(stored)
	base.StoredSHA = hex.EncodeToString(hash[:])
	custody, e := os.ReadFile(filepath.Join(filepath.Dir(file), "capture-custody.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(custody, &base.Custody); e != nil {
		t.Fatal(e)
	}
	if !lpLiveSemantic(base) {
		t.Fatalf("independent log-pattern evidence rejected source=%v/%v/%v scope/display=%v/%v partition=%v stored=%v package=%v requests=%v", lpLiveSource(base.Result, base.Environment, base.Records), lpLiveSource(base.Selected, base.Environment, base.Records), lpLiveSource(base.ParentAgain, base.Environment, base.Records), lpLiveDisplayCheck(base, base.InitialRows, false), lpLiveDisplayCheck(base, base.RareRows, true), lpLivePartition(base), lpLiveStored(base), lpLivePackage(base), lpLiveRequests(base))
	}
	if !lpLiveReplay(t, base) {
		t.Fatal("actual log-pattern lifecycle rejected")
	}

	controls := []struct {
		name   string
		change func(*lpLiveLedger)
	}{
		{"window", func(l *lpLiveLedger) { l.Original.ToMS++ }},
		{"original-predicate", func(l *lpLiveLedger) { l.Original.Where = nil }},
		{"pattern-field", func(l *lpLiveLedger) { l.Original.GroupBy[0].Key = "body" }},
		{"grouped-trend-cap", func(l *lpLiveLedger) { l.Original.GroupTimeSeries = true }},
		{"full-count", func(l *lpLiveLedger) { l.Result.Count += "0" }},
		{"duplicate-row", func(l *lpLiveLedger) { l.Result.Rows[1] = l.Result.Rows[0] }},
		{"raw-body", func(l *lpLiveLedger) { lpLiveMutateRow(&l.Result.Rows[0], "Body", "corrupted raw body") }},
		{"raw-time", func(l *lpLiveLedger) { lpLiveMutateRow(&l.Result.Rows[0], "timestamp_ns", "1") }},
		{"raw-trace", func(l *lpLiveLedger) { lpLiveMutateRow(&l.Result.Rows[0], "TraceId", "wrong") }},
		{"derived-row", func(l *lpLiveLedger) { lpLiveMutateRow(&l.Result.Rows[0], "body_pattern_v1", "lp2:corrupt") }},
		{"duplicate-group", func(l *lpLiveLedger) { l.Result.Groups[1] = l.Result.Groups[0] }},
		{"group-count", func(l *lpLiveLedger) { lpLiveMutateRow(&l.Result.Groups[0], "count", "999999") }},
		{"group-identity", func(l *lpLiveLedger) { lpLiveMutateRow(&l.Result.Groups[0], "g0", "lp2:e") }},
		{"time-count", func(l *lpLiveLedger) { l.Result.TimeSeries[0].Count += "9" }},
		{"time-boundary", func(l *lpLiveLedger) { l.Result.TimeSeries[0].ToMS++ }},
		{"display-label", func(l *lpLiveLedger) { l.InitialRows[0].Label += "corrupt" }},
		{"display-share", func(l *lpLiveLedger) { l.InitialRows[0].Share = "100% of displayed examples" }},
		{"rare-order", func(l *lpLiveLedger) {
			l.RareRows[0], l.RareRows[len(l.RareRows)-1] = l.RareRows[len(l.RareRows)-1], l.RareRows[0]
		}},
		{"selected-equals-whole-parent", func(l *lpLiveLedger) { l.SelectedStage.Where = l.Original.Where }},
		{"selected-window", func(l *lpLiveLedger) { l.SelectedStage.FromMS++ }},
		{"rest-missing-unclassified", func(l *lpLiveLedger) {
			l.Comparison.Request.Baseline.Question.Where = l.Comparison.Request.Selection.Question.Where
		}},
		{"comparison-count", func(l *lpLiveLedger) { l.Comparison.Request.Baseline.Count += "0" }},
		{"comparison-sign", func(l *lpLiveLedger) { l.Comparison.Response.Rows[0].DifferencePP = "999.99" }},
		{"comparison-basis", func(l *lpLiveLedger) { l.Comparison.Response.Basis = "wrong" }},
		{"raw-inspector", func(l *lpLiveLedger) { lpLiveMutateRow(&l.RawDetail, "Body", "rendered template instead of raw body") }},
		{"stored-id", func(l *lpLiveLedger) { l.Stored.ID = "wrong" }},
		{"stored-body", func(l *lpLiveLedger) { lpLiveMutateRow(&l.Stored.Explore.Rows[0], "Body", "changed") }},
		{"reopened-clock", func(l *lpLiveLedger) { l.Reopened.Receipt.CapturedAt = "wrong" }},
		{"custody-hash", func(l *lpLiveLedger) { l.Custody.SHA256 = "wrong" }},
		{"package-field", func(l *lpLiveLedger) { l.Reusable.Definition.GroupBy[0].Key = "service" }},
		{"reuse-scope", func(l *lpLiveLedger) { l.Reused.Question.Where = nil }},
		{"source-endpoint", func(l *lpLiveLedger) { l.Requests[1].Endpoint = "/api/inventory" }},
	}
	for _, action := range []string{"patterns-stage", "sort-local", "raw-inspection", "pattern-trend-stage", "pattern-back", "comparison-stage", "reopen", "reuse-stage", "return-paused"} {
		a := action
		controls = append(controls, struct {
			name   string
			change func(*lpLiveLedger)
		}{"hidden-read-" + a, func(l *lpLiveLedger) {
			for i := range l.Events {
				if l.Events[i].Action == a {
					l.Events[i].Reads++
					return
				}
			}
		}})
	}
	for _, control := range controls {
		t.Run("reject-"+control.name, func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l lpLiveLedger
			json.Unmarshal(before, &l)
			control.change(&l)
			after, _ := json.Marshal(l)
			if string(before) == string(after) {
				t.Fatal("negative control did not change ledger")
			}
			if lpLiveSemantic(l) && lpLiveReplay(t, l) {
				t.Fatal("corrupted log pattern accepted")
			}
		})
	}
}

func lpLiveMutateRow(raw *json.RawMessage, key, value string) {
	var row map[string]json.RawMessage
	json.Unmarshal(*raw, &row)
	row[key], _ = json.Marshal(value)
	*raw, _ = json.Marshal(row)
}

type lpLiveOwnership struct {
	Version      int
	Environment  string
	Records      []lpLiveRecord
	Question     exploration.Query
	Result       exploration.Result
	FrequentID   string
	Observations []struct {
		Action string
		Before struct {
			URL       string
			Stages    int
			Connected bool
		}
		After struct {
			URL    string
			Stages int
		}
		Reads int
	}
	Events   []cpLiveEvent
	Requests []struct {
		Effect, Endpoint string
		Payload          json.RawMessage
	}
}

func lpLiveOwnershipSemantic(l lpLiveOwnership) bool {
	if l.Version != 1 || len(l.Observations) != 2 || len(l.Requests) != 1 || !lpLiveSource(l.Result, l.Environment, l.Records) || !nhLiveSameQuery(l.Question, l.Result.Question) || !lpCanonical(l.FrequentID) {
		return false
	}
	for i, o := range l.Observations {
		if o.Action != []string{"hidden-owner", "dirty-owner"}[i] || o.Before.URL == "" || o.Before.URL != o.After.URL || o.Before.Stages != o.After.Stages || o.Reads != 1 || i == 0 && !o.Before.Connected {
			return false
		}
	}
	var q exploration.Query
	r := l.Requests[0]
	return r.Effect == "source" && r.Endpoint == "/api/explore" && json.Unmarshal(r.Payload, &q) == nil && nhLiveSameQuery(q, l.Question)
}
func TestQuintLogPatternOwnershipLedger(t *testing.T) {
	file := os.Getenv("GOO_LOG_PATTERN_OWNERSHIP_LEDGER")
	if file == "" {
		t.Skip("requires actual log-pattern stale-owner ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base lpLiveOwnership
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	if !lpLiveOwnershipSemantic(base) || !lpLiveReplayModel(t, base.Events, "model/log_patterns_ownership.yaml") {
		t.Fatal("actual log-pattern ownership rejected")
	}
	for i := 0; i < 12; i++ {
		n := i
		t.Run(fmt.Sprintf("reject-owner-%02d", n), func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l lpLiveOwnership
			json.Unmarshal(before, &l)
			switch n {
			case 0:
				l.Observations[0].Before.Connected = false
			case 1:
				l.Observations[0].After.URL += "-stale"
			case 2:
				l.Observations[1].After.URL += "-stale"
			case 3:
				l.Observations[0].After.Stages++
			case 4:
				l.Observations[1].After.Stages++
			case 5:
				l.Observations[0].Reads++
			case 6:
				l.Observations[1].Reads++
			case 7:
				l.Requests[0].Endpoint = "/api/inventory"
			case 8:
				l.Events[1].Reads++
			case 9:
				l.Events[2].Reads++
			case 10:
				l.Events[2].Streams++
			case 11:
				l.Result.Rows[1] = l.Result.Rows[0]
			}
			after, _ := json.Marshal(l)
			if string(before) == string(after) {
				t.Fatal("control unchanged")
			}
			if lpLiveOwnershipSemantic(l) && lpLiveReplayModel(t, l.Events, "model/log_patterns_ownership.yaml") {
				t.Fatal("obsolete owner accepted")
			}
		})
	}
}

type lpLiveCardinality struct {
	Version          int
	Environment      string
	Records          []lpLiveRecord
	ReadyProbe       exploration.Result
	Question         exploration.Query
	DistinctPatterns int
	Refusal          struct {
		Status   int
		Request  exploration.Query
		Response map[string]json.RawMessage
	}
	Events   []cpLiveEvent
	Requests []struct {
		Effect, Endpoint string
		Payload          json.RawMessage
	}
}

func lpLiveCardinalitySemantic(l lpLiveCardinality) bool {
	if l.Version != 1 || !lpLiveSource(l.ReadyProbe, l.Environment, l.Records) || !nhLiveSameQuery(l.Refusal.Request, l.Question) || l.Question.GroupTimeSeries || len(l.Question.GroupBy) != 1 || l.Question.GroupBy[0] != (exploration.Field{Kind: "column", Key: "body_pattern_v1"}) || l.Refusal.Status != 502 || len(l.Requests) != 1 {
		return false
	}
	probeQuestion := l.Question
	probeQuestion.GroupBy = nil
	probeQuestion.BucketMS = 0
	if !nhLiveSameQuery(probeQuestion, l.ReadyProbe.Question) {
		return false
	}
	rows, e := lpLiveRows(l.Question, l.Environment, l.Records)
	if e != nil {
		return false
	}
	groups := map[string]bool{}
	for _, r := range rows {
		groups[lpOracle(r.Body).Identity] = true
	}
	if len(groups) <= 1000 || l.DistinctPatterns != len(groups) || l.ReadyProbe.Count != strconv.Itoa(len(rows)) {
		return false
	}
	var message string
	if json.Unmarshal(l.Refusal.Response["error"], &message) != nil || !strings.Contains(message, "more than 1000 groups; narrow query; no partial grouping totals") {
		return false
	}
	for _, key := range []string{"rows", "groups", "count", "timeSeries"} {
		if _, ok := l.Refusal.Response[key]; ok {
			return false
		}
	}
	var q exploration.Query
	r := l.Requests[0]
	return r.Effect == "source" && r.Endpoint == "/api/explore" && json.Unmarshal(r.Payload, &q) == nil && nhLiveSameQuery(q, l.Question)
}
func TestQuintLogPatternCardinalityLedger(t *testing.T) {
	file := os.Getenv("GOO_LOG_PATTERN_CARDINALITY_LEDGER")
	if file == "" {
		t.Skip("requires actual high-cardinality log-pattern refusal ledger")
	}
	raw, e := os.ReadFile(file)
	if e != nil {
		t.Fatal(e)
	}
	var base lpLiveCardinality
	if e = json.Unmarshal(raw, &base); e != nil {
		t.Fatal(e)
	}
	if !lpLiveCardinalitySemantic(base) || !lpLiveReplayModel(t, base.Events, "model/log_patterns_cardinality.yaml") {
		t.Fatal("actual high-cardinality refusal rejected")
	}
	for i := 0; i < 12; i++ {
		n := i
		t.Run(fmt.Sprintf("reject-cardinality-%02d", n), func(t *testing.T) {
			before, _ := json.Marshal(base)
			var l lpLiveCardinality
			json.Unmarshal(before, &l)
			switch n {
			case 0:
				l.DistinctPatterns = 1000
			case 1:
				l.Refusal.Status = 200
			case 2:
				l.Refusal.Response["groups"] = json.RawMessage(`[]`)
			case 3:
				l.Refusal.Response["count"] = json.RawMessage(`"1000"`)
			case 4:
				l.Refusal.Response["rows"] = json.RawMessage(`[]`)
			case 5:
				l.Refusal.Response["error"] = json.RawMessage(`"sampled top patterns"`)
			case 6:
				l.Refusal.Request.Where = nil
			case 7:
				l.Requests[0].Endpoint = "/api/inventory"
			case 8:
				l.ReadyProbe.Count = "1000"
			case 9:
				l.Events[0].Reads++
			case 10:
				l.Events[1].Reads++
			case 11:
				l.Events[1].Writes++
			}
			after, _ := json.Marshal(l)
			if string(before) == string(after) {
				t.Fatal("control unchanged")
			}
			if lpLiveCardinalitySemantic(l) && lpLiveReplayModel(t, l.Events, "model/log_patterns_cardinality.yaml") {
				t.Fatal("partial top list accepted")
			}
		})
	}
}

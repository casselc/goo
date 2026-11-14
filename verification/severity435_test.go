package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
)

var severityBands435 = []string{"trace", "debug", "info", "warn", "error", "fatal", "unspecified", "invalid"}

func severityPredicate435(band string) exploration.Predicate {
	f := &exploration.Field{Kind: "column", Key: "severity_number"}
	if band == "unspecified" {
		return exploration.Predicate{Field: f, Op: "eq", Value: "0"}
	}
	if band == "invalid" {
		return exploration.Predicate{Field: f, Op: "gte", Value: "25"}
	}
	for i, b := range severityBands435[:6] {
		if b == band {
			return exploration.Predicate{All: []exploration.Predicate{{Field: f, Op: "gte", Value: fmt.Sprint(1 + 4*i)}, {Field: f, Op: "lte", Value: fmt.Sprint(4 + 4*i)}}}
		}
	}
	panic("unknown band")
}

func severityMap435(s exploration.LogSeveritySummary) map[string]string {
	return map[string]string{"count": s.Count, "trace": s.Trace, "debug": s.Debug, "info": s.Info, "warn": s.Warn, "error": s.Error, "fatal": s.Fatal, "unspecified": s.Unspecified, "invalid": s.Invalid}
}

// Compare the production classifier and exact numeric predicate evaluator to an
// independent band partition, including omitted/unspecified and invalid values.
func TestHegelNumericSeverity435(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := uint64(hegel.Draw(tc, hegel.Integers(0, 1000000)))
		if hegel.Draw(tc, hegel.Booleans()) {
			n = uint64(hegel.Draw(tc, hegel.Integers(0, 255)))
		}
		want := "invalid"
		if n == 0 {
			want = "unspecified"
		} else if n <= 24 {
			want = severityBands435[(n-1)/4]
		}
		if got := exploration.LogSeverityClass(n); got != want {
			tc.Fatal("severity partition", n, got, want)
		}
		matches := 0
		for _, band := range severityBands435 {
			p := severityPredicate435(band)
			values := map[exploration.Field]string{{Kind: "column", Key: "severity_number"}: strconv.FormatUint(n, 10), {Kind: "column", Key: "severity"}: "FATAL"}
			got := exploration.Match(&p, values)
			if got {
				matches++
			}
			if got != (band == want) {
				tc.Fatal("exact numeric filter differs from band", n, band, want)
			}
			delete(values, exploration.Field{Kind: "column", Key: "severity_number"})
			if exploration.Match(&p, values) {
				tc.Fatal("missing numeric severity inferred from text")
			}
		}
		if matches != 1 {
			tc.Fatal("severity partition is not disjoint/exhaustive", n, matches)
		}
		value := uint64(hegel.Draw(tc, hegel.Integers(0, 1000000)))
		s := exploration.LogSeveritySummary{Count: fmt.Sprint(value), Trace: "0", Debug: "0", Info: "0", Warn: "0", Error: fmt.Sprint(value), Fatal: "0", Unspecified: "0", Invalid: "0"}
		if err := s.Validate(); err != nil {
			tc.Fatal(err)
		}
		bad := s
		bad.Info = "1"
		if bad.Validate() == nil {
			tc.Fatal("nonconserved severity admitted")
		}
		bad = s
		bad.Error = "0" + s.Error
		if bad.Validate() == nil {
			tc.Fatal("noncanonical severity admitted")
		}
		bad = s
		bad.Count = "18446744073709551616"
		bad.Error = bad.Count
		if bad.Validate() == nil {
			tc.Fatal("overflow severity admitted")
		}
	}, options(180)...)
}

// Equal grand totals cannot make a per-band drift actionable. Exercise the real
// aggregate source with generated counts, drift placement, and text labels.
func TestHegelSeverityConservation435(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 1000000))
		drift := hegel.Draw(tc, hegel.SampledFrom([]string{"none", "groups", "time"}))
		q := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 1, GroupBy: []exploration.Field{{Kind: "column", Key: "severity"}}, BucketMS: 10000, LogSeverity: true}
		plan, err := exploration.Compile(q)
		if err != nil {
			tc.Fatal(err)
		}
		reads := 0
		source := exploration.Source{Endpoint: "http://fixture.invalid", Client: &http.Client{Transport: cvTransport(func(r *http.Request) (*http.Response, error) {
			b, _ := io.ReadAll(r.Body)
			sql := string(b)
			reads++
			row := map[string]string{"count": fmt.Sprint(n), "severity_trace": "0", "severity_debug": "0", "severity_info": "0", "severity_warn": "0", "severity_error": fmt.Sprint(n), "severity_fatal": "0", "severity_unspecified": "0", "severity_invalid": "0"}
			if sql == plan.Groups {
				row["g0"] = "not-an-ERROR-label"
			}
			if sql == plan.TimeSeries {
				row["from_ms"] = "1000"
			}
			if drift == "groups" && sql == plan.Groups || drift == "time" && sql == plan.TimeSeries {
				row["severity_error"] = fmt.Sprint(n - 1)
				row["severity_fatal"] = "1"
			}
			if sql != plan.Count && sql != plan.Groups && sql != plan.TimeSeries {
				return nil, fmt.Errorf("unexpected inspection read")
			}
			body, _ := json.Marshal(row)
			return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(string(body)))}, nil
		})}}
		result, err := source.FetchRecordAggregates(context.Background(), q)
		if err != nil {
			tc.Fatal(err)
		}
		want := "observed"
		if drift != "none" {
			want = "changed_between_reads"
		}
		if result.State != want || reads != 3 || result.LogSeverity.Error != fmt.Sprint(n) {
			tc.Fatal("aggregate conservation/read boundary", result, reads, want)
		}
		for _, point := range result.TimeSeries[1:] {
			if point.LogSeverity == nil || point.LogSeverity.Validate() != nil || point.LogSeverity.Count != "0" {
				tc.Fatal("zero bucket lost full severity partition", point)
			}
		}
	}, options(80)...)
}

type severityReport435 struct {
	Version   string          `json:"version"`
	Status    string          `json:"status"`
	Errors    []string        `json:"errors"`
	Selector  json.RawMessage `json:"selector"`
	Aggregate struct {
		Question exploration.Query                  `json:"question"`
		Result   exploration.RecordAggregatesResult `json:"result"`
	} `json:"aggregate"`
	SelectedBand struct {
		Band    string   `json:"band"`
		Points  []string `json:"points"`
		Ceiling string   `json:"ceiling"`
	} `json:"selectedBand"`
	FilterWhere     *exploration.Predicate `json:"filterWhere"`
	AppliedSelector json.RawMessage        `json:"appliedSelector"`
	Steps           []struct {
		Name        string `json:"name"`
		SourceReads *int   `json:"sourceReads"`
		Band        string `json:"band"`
	} `json:"steps"`
	Raw []struct {
		Timestamp string `json:"timestamp_ns"`
		Number    uint64 `json:"SeverityNumber"`
	} `json:"raw"`
	HeldRecords []struct {
		Timestamp string  `json:"timestamp_ns"`
		Number    *uint64 `json:"SeverityNumber"`
	} `json:"heldRecords"`
	Explanation struct {
		FromMS     int64    `json:"fromMs"`
		ToMS       int64    `json:"toMs"`
		HeldPoints []string `json:"heldPoints"`
	} `json:"explanation"`
	Calls []struct {
		Path   string          `json:"path"`
		Method string          `json:"method"`
		Body   json.RawMessage `json:"body"`
	} `json:"calls"`
}
type severityStep435 struct {
	Action                   string
	Reads                    int
	ScopeOK, BandsOK, ViewOK bool
}

var severityActions435 = []string{"loaded", "volume", "band", "stale_refused", "explain", "filter", "apply", "retire"}

func severityConserved435(r exploration.RecordAggregatesResult) bool {
	if r.LogSeverity == nil || r.LogSeverity.Validate() != nil || r.LogSeverity.Count != r.Count || r.State != "observed" && r.State != "no_data" {
		return false
	}
	total := severityMap435(*r.LogSeverity)
	for _, projection := range []string{"groups", "time"} {
		sums := map[string]*big.Int{}
		for key := range total {
			sums[key] = new(big.Int)
		}
		add := func(summary *exploration.LogSeveritySummary, count string) bool {
			if summary == nil || summary.Validate() != nil || summary.Count != count {
				return false
			}
			for key, value := range severityMap435(*summary) {
				n, _ := new(big.Int).SetString(value, 10)
				sums[key].Add(sums[key], n)
			}
			return true
		}
		if projection == "groups" {
			for _, raw := range r.Groups {
				var g struct {
					Count   string                          `json:"count"`
					Summary *exploration.LogSeveritySummary `json:"logSeverity"`
				}
				if json.Unmarshal(raw, &g) != nil || !add(g.Summary, g.Count) {
					return false
				}
			}
		} else {
			if exploration.ValidateTimeSeries(r.Question, r.TimeSeries) != nil {
				return false
			}
			for _, point := range r.TimeSeries {
				if !add(point.LogSeverity, point.Count) {
					return false
				}
			}
		}
		for key, value := range total {
			if sums[key].String() != value {
				return false
			}
		}
	}
	return true
}

func severityObserved435(r severityReport435) ([]severityStep435, error) {
	q := r.Aggregate.Question
	if q.Validate() != nil || !q.LogSeverity || q.Signal != "logs" || q.ToMS-q.FromMS != 60000 || q.BucketMS != 10000 || len(r.Steps) != 15 || !sparseSame432(q, r.Aggregate.Result.Question) {
		return nil, fmt.Errorf("missing valid numeric severity observations")
	}
	selector, err := sparseSelector432(r.Selector, q.ToMS)
	if err != nil {
		return nil, err
	}
	applied, err := sparseSelector432(r.AppliedSelector, q.ToMS)
	if err != nil {
		return nil, err
	}
	band := r.SelectedBand.Band
	found := false
	for _, b := range severityBands435 {
		found = found || b == band
	}
	if !found {
		return nil, fmt.Errorf("missing chosen severity band")
	}
	predicate := severityPredicate435(band)
	wantWhere := &predicate
	if selector.Where != nil {
		wantWhere = &exploration.Predicate{All: []exploration.Predicate{*selector.Where, predicate}}
	}
	wantApplied := selector
	wantApplied.Where = wantWhere
	filterOK := sparseSame432(wantWhere, r.FilterWhere)
	applyOK := tailSameScope433(applied, wantApplied)
	qScope := q.Signal == "logs" && q.SameEnvironment(selector.Query(q.ToMS)) && sparseSame432(q.Where, selector.Where) && len(q.GroupBy) == 1 && q.GroupBy[0] == (exploration.Field{Kind: "column", Key: "severity"})
	bandsOK := severityConserved435(r.Aggregate.Result)
	rawCounts := map[string]uint64{}
	for _, band := range severityBands435 {
		rawCounts[band] = 0
	}
	for _, row := range r.Raw {
		band := "invalid"
		if row.Number == 0 {
			band = "unspecified"
		} else if row.Number <= 24 {
			band = severityBands435[(row.Number-1)/4]
		}
		rawCounts[band]++
		n, ok := new(big.Int).SetString(row.Timestamp, 10)
		if !ok {
			return nil, fmt.Errorf("invalid independent event time")
		}
		from := new(big.Int).Mul(big.NewInt(q.FromMS), big.NewInt(1000000))
		to := new(big.Int).Mul(big.NewInt(q.ToMS), big.NewInt(1000000))
		bandsOK = bandsOK && n.Cmp(from) >= 0 && n.Cmp(to) < 0
	}
	if r.Aggregate.Result.LogSeverity == nil {
		return nil, fmt.Errorf("missing summary")
	}
	observed := severityMap435(*r.Aggregate.Result.LogSeverity)
	bandsOK = bandsOK && fmt.Sprint(len(r.Raw)) == r.Aggregate.Result.Count
	for band, count := range rawCounts {
		bandsOK = bandsOK && observed[band] == fmt.Sprint(count)
	}
	points := []string{}
	ceiling := big.NewInt(1)
	for _, point := range r.Aggregate.Result.TimeSeries {
		if point.LogSeverity == nil {
			return nil, fmt.Errorf("missing bucket severity")
		}
		points = append(points, severityMap435(*point.LogSeverity)[band])
		n, ok := new(big.Int).SetString(point.Count, 10)
		if !ok {
			return nil, fmt.Errorf("invalid count")
		}
		if n.Cmp(ceiling) > 0 {
			ceiling = n
		}
	}
	viewOK := reflect.DeepEqual(points, r.SelectedBand.Points) && ceiling.String() == r.SelectedBand.Ceiling
	heldCounts := make([]uint64, 6)
	guideOK := len(r.HeldRecords) == 100 && r.Explanation.FromMS >= q.FromMS && r.Explanation.ToMS <= q.ToMS && r.Explanation.ToMS-r.Explanation.FromMS == 10000 && (r.Explanation.FromMS-q.FromMS)%10000 == 0
	if guideOK {
		index := (r.Explanation.FromMS - q.FromMS) / 10000
		guideOK = points[index] != "0"
	}
	for _, row := range r.HeldRecords {
		if row.Number == nil {
			guideOK = false
			continue
		}
		// The explanation guide describes retained matching observations only;
		// it must never substitute the full-population aggregate as held rows.
		if exploration.LogSeverityClass(*row.Number) != band {
			continue
		}
		n, ok := new(big.Int).SetString(row.Timestamp, 10)
		if !ok {
			return nil, fmt.Errorf("invalid held event time")
		}
		ms := new(big.Int).Quo(n, big.NewInt(1000000))
		if !ms.IsInt64() {
			guideOK = false
			continue
		}
		at := ms.Int64()
		if at >= q.FromMS && at < q.ToMS {
			heldCounts[(at-q.FromMS)/10000]++
		}
	}
	heldPoints := []string{}
	for _, n := range heldCounts {
		heldPoints = append(heldPoints, fmt.Sprint(n))
	}
	guideOK = guideOK && reflect.DeepEqual(heldPoints, r.Explanation.HeldPoints)
	reads := 0
	steps := []severityStep435{}
	seenBands := map[string]bool{}
	wantActions := []string{"loaded", "volume", "band", "band", "band", "band", "band", "band", "band", "band", "stale_refused", "explain", "filter", "apply", "retire"}
	for i, fact := range r.Steps {
		if fact.Name != wantActions[i] || fact.SourceReads == nil || *fact.SourceReads < 0 {
			return nil, fmt.Errorf("missing action/read fact %d", i)
		}
		reads += *fact.SourceReads
		action := fact.Name
		if action == "filter" {
			action = "stage_filter"
		}
		s := severityStep435{Action: action, Reads: reads, ScopeOK: true, BandsOK: true, ViewOK: true}
		switch fact.Name {
		case "volume":
			s.ScopeOK = qScope
			s.BandsOK = bandsOK
		case "band":
			validBand := false
			for _, band := range severityBands435 {
				validBand = validBand || band == fact.Band
			}
			s.ViewOK = viewOK && validBand && !seenBands[fact.Band]
			seenBands[fact.Band] = true
		case "filter":
			s.ScopeOK = filterOK
		case "explain":
			s.ViewOK = guideOK
		case "apply":
			s.ScopeOK = applyOK
		}
		steps = append(steps, s)
	}
	return steps, nil
}

func severityCheck435(t *testing.T, steps []severityStep435) (bool, error) {
	t.Helper()
	quint, err := exec.LookPath("quint")
	if err != nil {
		return false, err
	}
	spec, err := filepath.Abs("model/severity435.qnt")
	if err != nil {
		return false, err
	}
	manifest := "spec: " + spec + "\nmodule: severity435\nconstants: {INITIAL: '0'}\ninvariants: [safety]\nactions:\n"
	for _, name := range severityActions435 {
		if name == "filter" {
			name = "stage_filter"
		}
		manifest += "  " + name + ":\n    expect: 'reads == {{.obs.reads}} and scopeOK == {{.obs.scopeOK}} and bandsOK == {{.obs.bandsOK}} and viewOK == {{.obs.viewOK}}'\n"
	}
	file := filepath.Join(t.TempDir(), "binding.yaml")
	if err = os.WriteFile(file, []byte(manifest), 0600); err != nil {
		return false, err
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink)
	for _, s := range steps {
		rec.Record(context.Background(), s.Action, nil, "obs.reads", s.Reads, "obs.scopeOK", s.ScopeOK, "obs.bandsOK", s.BandsOK, "obs.viewOK", s.ViewOK)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		return false, err
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		return false, err
	}
	return result.OK(), nil
}

func TestQuintObservedNumericSeverity435(t *testing.T) {
	file := os.Getenv("GOO_SEVERITY435_REPORT")
	if file == "" {
		t.Skip("set GOO_SEVERITY435_REPORT to the real completed severity journey report for required acceptance")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var report severityReport435
	if json.Unmarshal(raw, &report) != nil || report.Version != "goo.activitySeverity435/1" || report.Status != "passed" || len(report.Errors) > 0 {
		t.Fatal("incomplete real severity report")
	}
	// Link the scope facts to actual outgoing browser requests, not a draft-only
	// projection. The independent actor uses its own unique environment.
	q := report.Aggregate.Question
	origin, err := sparseSelector432(report.Selector, q.ToMS)
	if err != nil {
		t.Fatal(err)
	}
	applied, err := sparseSelector432(report.AppliedSelector, q.ToMS)
	if err != nil {
		t.Fatal(err)
	}
	tailBefore, tailAfter, aggregate := 0, 0, 0
	for _, call := range report.Calls {
		if call.Method != "POST" {
			continue
		}
		if call.Path == "/api/tail" {
			selector, e := sparseSelector432(call.Body, q.ToMS)
			if e != nil {
				continue
			}
			if tailSameScope433(selector, origin) {
				tailBefore++
			}
			if tailSameScope433(selector, applied) {
				tailAfter++
			}
		}
		if call.Path == "/api/record-aggregates" {
			var actual exploration.Query
			if json.Unmarshal(call.Body, &actual) == nil && sparseSame432(actual, q) {
				aggregate++
			}
		}
	}
	if tailBefore != 1 || tailAfter != 1 || aggregate != 1 {
		t.Fatal("missing/repeated actual source requests", tailBefore, tailAfter, aggregate)
	}
	steps, err := severityObserved435(report)
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := severityCheck435(t, steps); err != nil || !ok {
		t.Fatal("actual severity trace refused", ok, err)
	}
	for _, control := range []string{"group-band-drift", "time-band-drift", "narrow-scale", "wrong-band-points", "held-guide-drift", "explanation-window", "parent-dropped", "wrong-numeric-filter", "implicit-band-read", "implicit-filter-read", "retired-owner-read"} {
		t.Run(control, func(t *testing.T) {
			var bad severityReport435
			encoded, _ := json.Marshal(report)
			if json.Unmarshal(encoded, &bad) != nil {
				t.Fatal("copy report")
			}
			switch control {
			case "group-band-drift":
				var group map[string]json.RawMessage
				json.Unmarshal(bad.Aggregate.Result.Groups[0], &group)
				var s exploration.LogSeveritySummary
				json.Unmarshal(group["logSeverity"], &s)
				s = severityDrift435(s)
				group["logSeverity"], _ = json.Marshal(s)
				bad.Aggregate.Result.Groups[0], _ = json.Marshal(group)
			case "time-band-drift":
				for i, point := range bad.Aggregate.Result.TimeSeries {
					if point.Count != "0" {
						*bad.Aggregate.Result.TimeSeries[i].LogSeverity = severityDrift435(*point.LogSeverity)
						break
					}
				}
			case "narrow-scale":
				ceiling := big.NewInt(1)
				for _, value := range bad.SelectedBand.Points {
					n, _ := new(big.Int).SetString(value, 10)
					if n.Cmp(ceiling) > 0 {
						ceiling = n
					}
				}
				if ceiling.String() == bad.SelectedBand.Ceiling {
					ceiling.Add(ceiling, big.NewInt(1))
				}
				bad.SelectedBand.Ceiling = ceiling.String()
			case "wrong-band-points":
				n, _ := new(big.Int).SetString(bad.SelectedBand.Points[0], 10)
				bad.SelectedBand.Points[0] = n.Add(n, big.NewInt(1)).String()
			case "held-guide-drift":
				n, _ := new(big.Int).SetString(bad.Explanation.HeldPoints[0], 10)
				bad.Explanation.HeldPoints[0] = n.Add(n, big.NewInt(1)).String()
			case "explanation-window":
				bad.Explanation.FromMS--
			case "parent-dropped":
				bad.FilterWhere = &bad.FilterWhere.All[1]
			case "wrong-numeric-filter":
				bad.FilterWhere = &exploration.Predicate{All: []exploration.Predicate{bad.FilterWhere.All[0], severityPredicate435("unspecified")}}
			case "implicit-band-read":
				*bad.Steps[2].SourceReads++
			case "implicit-filter-read":
				*bad.Steps[12].SourceReads++
			case "retired-owner-read":
				*bad.Steps[14].SourceReads++
			}
			trace, err := severityObserved435(bad)
			if err != nil {
				t.Fatal("control malformed instead of semantically rejected", err)
			}
			ok, err := severityCheck435(t, trace)
			if err != nil {
				t.Fatal(err)
			}
			if ok {
				t.Fatal("unsafe altered report admitted", control)
			}
		})
	}
}

func severityDrift435(s exploration.LogSeveritySummary) exploration.LogSeveritySummary {
	pointers := []*string{&s.Trace, &s.Debug, &s.Info, &s.Warn, &s.Error, &s.Fatal, &s.Unspecified, &s.Invalid}
	for i, p := range pointers {
		n, _ := new(big.Int).SetString(*p, 10)
		if n.Sign() > 0 {
			n.Sub(n, big.NewInt(1))
			*p = n.String()
			target := pointers[(i+1)%len(pointers)]
			m, _ := new(big.Int).SetString(*target, 10)
			*target = m.Add(m, big.NewInt(1)).String()
			return s
		}
	}
	panic("drift requires nonempty summary")
}

package exploration

import (
	"encoding/json"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func heldViewCohort(t *testing.T, b *NumericBinding, values []*string, durations []uint64) Cohort {
	t.Helper()
	r := measureGroupFixture(t, b, values, durations)
	c := Cohort{Question: detached(r.Question), Basis: "caller-claimed-basis", Count: r.Count, State: "observed", Numeric: detached(r.Numeric), Duration: detached(r.Duration), Groups: []ValueCount{}}
	if c.Count == "0" {
		c.State = "no_data"
	} else {
		var g ValueCount
		if json.Unmarshal(r.Groups[0], &g) != nil {
			t.Fatal("fixture")
		}
		c.Groups = append(c.Groups, g)
	}
	return c
}
func TestSummaryDatasetSourceAspectAndEmptySupport(t *testing.T) {
	b := decimalBinding(2)
	r := measureGroupFixture(t, &b, numericValues("0.01", "0.02"), []uint64{1, 2})
	d := summaryViewDataset(r, true)
	if d == nil || d.Shape != "summary" || d.Origin != "source-read" || len(d.Rows) != 1 || d.Rows[0].Key != "population" || len(d.Dimensions) != 0 || d.Interval != nil || d.Completeness == nil || !reflect.DeepEqual(d.Completeness.Aspects, []string{"summary"}) || *d.Rows[0].Values["numeric.mean"] != "3/200" {
		t.Fatal(d)
	}
	if summaryViewDataset(r, false).Completeness != nil {
		t.Fatal("source receipt invented from count agreement")
	}
	r.State = "changed_between_reads"
	if summaryViewDataset(r, true).Completeness != nil {
		t.Fatal("diagnostic summary attested")
	}
	empty := measureGroupFixture(t, &b, nil, []uint64{})
	empty.State = "no_data"
	d = summaryViewDataset(empty, true)
	if d == nil || d.Completeness == nil || *d.Rows[0].Values["count"] != "0" || d.Rows[0].Values["numeric.sum"] != nil || d.Rows[0].Cells["duration.mean"].State != "no_data" {
		t.Fatal(d)
	}
}
func TestViewComparisonGeneratedScaledRationalDifferences(t *testing.T) {
	for scale := 1; scale <= 9; scale++ {
		for seed := 1; seed <= 12; seed++ {
			b := decimalBinding(scale)
			aValues := numericValues(formatFixedCoefficient(big.NewInt(int64(seed)), scale), formatFixedCoefficient(big.NewInt(int64(seed+1)), scale), "bad")
			zValues := numericValues(formatFixedCoefficient(big.NewInt(int64(-seed)), scale), formatFixedCoefficient(big.NewInt(int64(seed+2)), scale), formatFixedCoefficient(big.NewInt(9007199254740993), scale))
			a, z := heldViewCohort(t, &b, aValues, nil), heldViewCohort(t, &b, zValues, nil)
			a.CapturedAt = "2026-10-30T00:00:00Z"
			z.CapturedAt = "2026-10-30T00:00:01Z"
			z.Question.FromMS += 1000
			z.Question.ToMS += 1000
			result, err := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: z})
			if err != nil || result.ViewComparison == nil {
				t.Fatal(scale, seed, result, err)
			}
			v := result.ViewComparison
			reference := func(values []*string) *big.Rat {
				sum := new(big.Rat)
				n := int64(0)
				for _, value := range values {
					if value == nil {
						continue
					}
					parsed, ok := new(big.Rat).SetString(*value)
					if ok {
						sum.Add(sum, parsed)
						n++
					}
				}
				return sum.Quo(sum, big.NewRat(n, 1))
			}
			want := new(big.Rat).Sub(reference(zValues), reference(aValues)).RatString()
			if v.State != "observed" || v.Changes["numeric.mean"].State != "derived" || *v.Changes["numeric.mean"].Value != want {
				t.Fatal(scale, seed, v.Changes, want)
			}
			for _, side := range []*ViewDataset{v.Baseline, v.Selection} {
				if side.Origin != "browser-held" || side.Quality.Extent != "unknown" || side.Completeness != nil || side.SourceBasis != "caller-claimed-basis" || len(side.Revision) != 64 || side.Revision == a.Basis {
					t.Fatal(side)
				}
			}
			if v.Baseline.Scope.ID == v.Selection.Scope.ID || v.Baseline.CapturedAt == v.Selection.CapturedAt {
				t.Fatal("independent context flattened")
			}
		}
	}
}
func TestViewComparisonEmptyCountDeltaAndUnavailableMeasures(t *testing.T) {
	b := numericBinding()
	a := heldViewCohort(t, &b, nil, []uint64{})
	z := heldViewCohort(t, &b, numericValues("2", "3"), []uint64{5, 6})
	result, err := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: z})
	if err != nil || result.ViewComparison == nil {
		t.Fatal(result, err)
	}
	v := result.ViewComparison
	if v.State != "no_data" || *v.Changes["count"].Value != "2" || v.Changes["count"].State != "derived" {
		t.Fatal(v)
	}
	for _, key := range []string{"numeric.mean", "numeric.sum", "duration.p95"} {
		if v.Changes[key].Value != nil || v.Changes[key].State != "unavailable" || v.Changes[key].Reason == "" {
			t.Fatal(key, v.Changes[key])
		}
	}
	invalid := heldViewCohort(t, &b, []*string{nil, viewText("bad")}, []uint64{1, 2})
	result, err = CompareCohorts(ComparisonRequest{Numeric: true, Baseline: invalid, Selection: z})
	if err != nil || result.ViewComparison.Changes["numeric.sum"].Value != nil || result.ViewComparison.Changes["count"].Value == nil {
		t.Fatal(result, err)
	}
}
func TestViewComparisonSignedDurationChangesAreNotPooledQuantiles(t *testing.T) {
	a := heldViewCohort(t, nil, nil, []uint64{4, 4, 4})
	z := heldViewCohort(t, nil, nil, []uint64{0, 0, 1})
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: z})
	if err != nil || result.ViewComparison == nil {
		t.Fatal(result, err)
	}
	v := result.ViewComparison
	if *v.Changes["duration.mean"].Value != "-11/3" || *v.Changes["duration.p95"].Value != "-3" || v.Changes["duration.p95"].Aggregation.Method != "quantilesExactHigh" {
		t.Fatal(v)
	}
}
func TestViewComparisonDiagnosticIncompatibleAndBadBindings(t *testing.T) {
	a := heldViewCohort(t, nil, nil, []uint64{1, 2})
	z := heldViewCohort(t, nil, nil, []uint64{3, 4})
	a.State = "changed_between_reads"
	result, err := CompareCohorts(ComparisonRequest{Baseline: a, Selection: z})
	if err != nil || result.ViewComparison.State != "diagnostic" {
		t.Fatal(result, err)
	}
	for _, change := range result.ViewComparison.Changes {
		if change.Value != nil || change.State != "unavailable" || change.Reason == "" {
			t.Fatal(change)
		}
	}
	a = cohort("2", ValueCount{Value: str("svc"), Count: "2"})
	a.Question.GroupBy = z.Question.GroupBy
	result, err = CompareCohorts(ComparisonRequest{Baseline: a, Selection: z})
	if err != nil || result.ViewComparison == nil || result.ViewComparison.State != "incompatible" {
		t.Fatal(result, err)
	}
	for _, change := range result.ViewComparison.Changes {
		if change.Value != nil {
			t.Fatal("incompatible value coerced")
		}
	}
	b := numericBinding()
	x := heldViewCohort(t, &b, numericValues("1"), nil)
	y := detached(x)
	y.Question.Numeric.Unit = "other"
	if _, err = CompareCohorts(ComparisonRequest{Numeric: true, Baseline: x, Selection: y}); err == nil {
		t.Fatal("mismatched binding coerced")
	}
}
func TestViewComparisonHeldClaimsClocksAndImmutableIdentity(t *testing.T) {
	b := numericBinding()
	a := heldViewCohort(t, &b, numericValues("1", "2"), nil)
	z := heldViewCohort(t, &b, numericValues("3", "4"), nil)
	a.Basis = strings.Repeat("a", 64)
	z.Basis = a.Basis
	result, err := CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: z})
	if err != nil || result.ViewComparison == nil {
		t.Fatal(result, err)
	}
	v := result.ViewComparison
	if v.Baseline.CapturedAt != "" || v.Selection.CapturedAt != "" || v.Baseline.Revision == v.Selection.Revision || v.Baseline.Completeness != nil || v.Baseline.Origin != "browser-held" {
		t.Fatal(v)
	}
	raw, _ := json.Marshal(v.Baseline)
	var wire map[string]any
	json.Unmarshal(raw, &wire)
	if _, present := wire["capturedAt"]; present {
		t.Fatal("missing clock invented")
	}
	before, _ := json.Marshal(v)
	a.Numeric.Sum = "0"
	a.Question.Numeric.Unit = "mutated"
	z.Groups[0].Numeric.Valid = "0"
	after, _ := json.Marshal(v)
	if string(before) != string(after) {
		t.Fatal("caller mutation changed comparison")
	}
	z.CapturedAt = "made-up-clock"
	if _, err = CompareCohorts(ComparisonRequest{Numeric: true, Baseline: a, Selection: z}); err == nil {
		t.Fatal("bad claimed clock accepted")
	}
}

func TestCohortConstructorsRetainReadClockWithoutCustodySubstitution(t *testing.T) {
	r := measureGroupFixture(t, nil, nil, []uint64{1, 2})
	r.CapturedAt = "2026-10-30T02:03:04.123456789Z"
	c, err := CohortFromResult(r)
	if err != nil || c.CapturedAt != r.CapturedAt {
		t.Fatal(c, err)
	}
	r.CapturedAt = ""
	c, err = CohortFromResult(r)
	if err != nil || c.CapturedAt != "" {
		t.Fatal("unknown clock replaced", c, err)
	}
	// The explicit aggregate-only distribution constructor is a separate path.
	dc := distributionCohort("int64", map[string][]*string{"/a": numericValues("1", "2")})
	dc.Question.GroupBy = nil
	dc.Groups = nil
	dr := &Result{Question: dc.Question, Count: dc.Count, State: dc.State, Basis: dc.Basis, Numeric: dc.Numeric, NumericHistogram: dc.NumericHistogram, CapturedAt: "2026-10-30T03:04:05Z"}
	d, err := CohortFromResultDistribution(dr)
	if err != nil || d.CapturedAt != dr.CapturedAt {
		t.Fatal(d, err)
	}
}

package activitylogs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type step struct {
	Name                  string
	Reads, SQL, Adoptions int
}

func modelCheck(t *testing.T, steps []step) bool {
	t.Helper()
	quint := os.Getenv("QUINT_BIN")
	if quint == "" {
		quint = "../../node_modules/.bin/quint"
	}
	quint, err := filepath.Abs(quint)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = os.Stat(quint); err != nil {
		t.Fatalf("Quint required, not skipped: %v", err)
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(t.Name()))
	for _, s := range steps {
		rec.Record(context.Background(), s.Name, nil, "obs.reads", s.Reads, "obs.sql", s.SQL, "obs.adoptions", s.Adoptions)
	}
	checker, err := validate.NewChecker("binding.yaml", validate.Options{Quint: quint, Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	report, err := checker.Check(t.Context(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return report.OK()
}
func TestQuintOptionalLogReadFreezeAndRetirement(t *testing.T) {
	good := []step{{"show", 1, 3, 0}, {"complete", 1, 3, 1}, {"refresh", 2, 6, 1}, {"complete", 2, 6, 2}, {"freeze", 2, 6, 2}, {"choose_interval", 2, 6, 2}, {"explain", 2, 6, 2}, {"return_parent", 2, 6, 2}, {"retire", 2, 6, 2}, {"late_result", 2, 6, 2}}
	if !modelCheck(t, good) {
		t.Fatal("valid optional aggregate protocol refused")
	}
	for _, i := range []int{5, 6, 7} {
		bad := append([]step(nil), good...)
		bad[i].Reads++
		bad[i].SQL += 3
		bad[i].Adoptions++
		if modelCheck(t, bad) {
			t.Fatal("implicit local read accepted")
		}
	}
	// A new read may be retired before completion. Its late response must not adopt.
	pending := []step{{"show", 1, 3, 0}, {"retire", 1, 3, 0}, {"late_result", 1, 3, 0}}
	if !modelCheck(t, pending) {
		t.Fatal("retired pending read refused")
	}
	pending[2].Adoptions = 1
	if modelCheck(t, pending) {
		t.Fatal("retired response adopted")
	}
	wrong := append([]step(nil), good...)
	wrong[0].SQL = 5
	if modelCheck(t, wrong) {
		t.Fatal("inspection/field read budget accepted")
	}
	if modelCheck(t, []step{{"show", 1, 3, 0}, {"complete", 1, 3, 1}, {"choose_interval", 1, 3, 1}}) {
		t.Fatal("selection before freezing accepted")
	}

}

type counters struct{ Posts, SQL int }
type fact struct {
	Name          string
	Before, After counters
}
type observedRead struct {
	Request  exploration.Query
	Response json.RawMessage
	Status   int
}
type observedReport struct {
	Version  string
	Reads    []observedRead
	Facts    []fact
	Errors   []string
	Selected struct{ Parent, Selection exploration.Query }
}

func observedSteps(r observedReport) ([]step, error) {
	if r.Version != "goo.activityLogs/1" || len(r.Reads) == 0 || len(r.Errors) != 0 {
		return nil, fmt.Errorf("actual complete log report required")
	}
	seen := map[string]bool{}
	steps := []step{}
	reads, sql := 0, 0
	frozen := false
	for _, f := range r.Facts {
		name := f.Name
		if name == "read" {
			name = "show"
		}
		if name == "return" {
			name = "return_parent"
		}
		if name == "select" {
			name = "choose_interval"
		}
		if name != "show" && name != "refresh" && name != "choose_interval" && name != "explain" && name != "return_parent" {
			continue
		}
		posts, statements := f.After.Posts-f.Before.Posts, f.After.SQL-f.Before.SQL
		if name == "show" || name == "refresh" {
			if posts != 1 || statements != 3 {
				return nil, fmt.Errorf("%s requires one request/three SQL", name)
			}
			reads++
			sql += 3
			frozen = false
		} else {
			if posts != 0 || statements != 0 {
				return nil, fmt.Errorf("%s performed an implicit read", name)
			}
			if !frozen {
				steps = append(steps, step{"freeze", reads, sql, reads})
				frozen = true
			}
		}
		seen[name] = true
		if name == "show" || name == "refresh" {
			steps = append(steps, step{name, reads, sql, reads - 1}, step{"complete", reads, sql, reads})
		} else {
			steps = append(steps, step{name, reads, sql, reads})
		}
	}
	if !seen["show"] || !seen["choose_interval"] || !seen["return_parent"] {
		return nil, fmt.Errorf("show/select/return evidence required")
	}
	return steps, nil
}
func checkObservedProjection(t *testing.T, r observedReport) error {
	t.Helper()
	inputs := []any{}
	matched := false
	for _, read := range r.Reads {
		if read.Status != 200 {
			return fmt.Errorf("selected aggregate read not successful")
		}
		q := read.Request
		if err := exploration.ValidateRecordAggregatesQuery(q); err != nil {
			return err
		}
		if q.Signal != "logs" || q.ToMS-q.FromMS != 60000 || q.BucketMS != 10000 || q.Limit != 1 || q.Offset != 0 || !reflect.DeepEqual(q.GroupBy, []exploration.Field{{Kind: "column", Key: "severity"}}) {
			return fmt.Errorf("actual log projection changed")
		}
		input := map[string]any{"request": q, "dto": read.Response}
		if q.FromMS == r.Selected.Parent.FromMS && q.ToMS == r.Selected.Parent.ToMS {
			parent := q
			parent.Limit = 100
			parent.GroupBy = nil
			parent.BucketMS = 0
			if !reflect.DeepEqual(parent, r.Selected.Parent) {
				return fmt.Errorf("selected parent selector changed")
			}
			sel := r.Selected.Selection
			a, b := sel.FromMS-q.FromMS, sel.ToMS-q.FromMS
			if a < 0 || b > 60000 || a >= b || a%10000 != 0 || b%10000 != 0 {
				return fmt.Errorf("selected range escaped frozen buckets")
			}
			input["a"] = a / 10000
			input["b"] = b/10000 - 1
			matched = true
		}
		inputs = append(inputs, input)
	}
	if !matched {
		return fmt.Errorf("selection does not cite any actual aggregate window")
	}
	out := project(t, inputs)
	for i, p := range out {
		if !reflect.DeepEqual(p.Query, r.Reads[i].Request) {
			return fmt.Errorf("aggregate question changed")
		}
		if p.Query.FromMS == r.Selected.Parent.FromMS && p.Query.ToMS == r.Selected.Parent.ToMS && !reflect.DeepEqual(p.Selection, r.Selected.Selection) {
			return fmt.Errorf("selected UTC range/selector changed")
		}
	}
	return nil
}
func TestQuintObservedLogActivityLedger(t *testing.T) {
	path := os.Getenv("GOO_ACTIVITY_LOGS_REPORT")
	if path == "" {
		t.Skip("explicit actual browser report required; strict protocol test remains required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var report observedReport
	if err = json.Unmarshal(raw, &report); err != nil {
		t.Fatal(err)
	}
	if err = checkObservedProjection(t, report); err != nil {
		t.Fatal(err)
	}
	steps, err := observedSteps(report)
	if err != nil || !modelCheck(t, steps) {
		t.Fatalf("actual log read ledger refused: %v", err)
	}
	for _, change := range []func(*observedReport){func(r *observedReport) { r.Selected.Selection.ToMS++ }, func(r *observedReport) { r.Selected.Parent.FromMS++ }, func(r *observedReport) {
		for i := range r.Facts {
			if r.Facts[i].Name == "select" {
				r.Facts[i].After.SQL++
				return
			}
		}
	}} {
		var bad observedReport
		_ = json.Unmarshal(raw, &bad)
		change(&bad)
		if _, e := observedSteps(bad); e == nil {
			if e = checkObservedProjection(t, bad); e == nil {
				t.Fatal("altered actual scope/read ledger accepted")
			}
		}
	}
}

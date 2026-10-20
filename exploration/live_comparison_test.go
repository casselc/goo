package exploration

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestLiveComparisonPreservesBaselineAndFreezesDisplayedPopulation(t *testing.T) {
	source, total, closeSource := captureFixture(t)
	defer closeSource()
	store := CaptureStore{Dir: t.TempDir()}
	baseline, err := store.Create(context.Background(), source, "Baseline", savedQuestion(), "")
	if err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(baseline)
	total.Store(1400)
	now := time.UnixMilli(1800000000000)
	snapshot, pair, err := source.LiveRecordComparison(context.Background(), baseline, now)
	if err != nil {
		t.Fatal(err)
	}
	q := snapshot.Question.Explore
	bq := baseline.Question.Explore
	if pair.State != "comparable" || pair.CountDifference != "200" || snapshot.Explore.Count != "1400" || len(snapshot.Explore.Rows) != 1000 || q.ToMS != now.UnixMilli() || q.ToMS-q.FromMS != bq.ToMS-bq.FromMS {
		t.Fatal("population/window", pair, q)
	}
	after, _ := json.Marshal(baseline)
	if string(before) != string(after) {
		t.Fatal("baseline mutated")
	}
	total.Store(2000)
	closeSource()
	frozen, err := store.FreezeSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	reopened, err := store.Load(frozen.ID)
	if err != nil || reopened.Explore.Count != "1400" || reopened.Explore.CapturedAt != snapshot.Explore.CapturedAt {
		t.Fatal("freeze reread or changed snapshot", err)
	}
}

package measurements

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestPureEvaluationFixedTimeAndStorageBoundary(t *testing.T) {
	gauge := definition(t, "queue-lag")
	lifecycle := definition(t, "reservation-lifecycle")
	q := Question{Definition: gauge, FromMS: 100000, ToMS: 220000}
	rows := []json.RawMessage{raw(gaugeRow(gauge, "worker", 210000, "7"))}
	at := time.UnixMilli(230000)
	// A clean working directory makes accidental relative file publication visible.
	dir := t.TempDir()
	t.Chdir(dir)
	first, err := Evaluate(q, rows, at)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Evaluate(q, rows, at)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatal("same supplied input/time produced different results")
	}
	if err := ValidateResult(first); err != nil {
		t.Fatal(err)
	}
	if _, err := Evaluate(Question{Definition: lifecycle, FromMS: q.FromMS, ToMS: q.ToMS}, nil, at); err == nil {
		t.Fatal("pure evaluator admitted lifecycle custody")
	}
	if _, err := ObserveLifecycle(q, rows, at, "custody", nil); err == nil {
		t.Fatal("storage operation admitted a numeric definition")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("evaluation boundary published files")
	}
}

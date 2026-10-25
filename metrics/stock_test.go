package metrics

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

func stockPoint(at, start string, value float64) StockPoint {
	return StockPoint{Stream: Stream{Resource: map[string]string{"service.name": "api", "service.instance.id": "one"}, Scope: "otel", Name: "db.client.connection.count", Unit: "{connection}", Attributes: map[string]string{"db.client.connection.pool.name": "main"}}, Temporality: "cumulative", StartNS: start, EndNS: at, Value: value}
}
func TestStockObservationsPreserveDecreasesEpochsAndMissing(t *testing.T) {
	points := []StockPoint{stockPoint("10", "0", 8), stockPoint("70", "0", 3.5), stockPoint("80", "75", -2), stockPoint("90", "75", 0)}
	points[3].NoRecordedValue = true
	got, err := EvaluateStocks("0", "100", "30", points)
	if err != nil {
		t.Fatal(err)
	}
	s := got.Series[0]
	if s.State != "unknown" || s.Samples[0].State != "stale" || *s.Samples[1].Value != 3.5 || *s.Samples[2].Value != -2 || s.Samples[2].StartNS != "75" || s.Samples[3].Value != nil {
		t.Fatalf("lost stock semantics: %+v", s)
	}
	if points[0].EndNS != "10" {
		t.Fatal("input mutated")
	}
	got.Series[0].Stream.Resource["service.name"] = "changed"
	if points[0].Stream.Resource["service.name"] != "api" {
		t.Fatal("input aliased")
	}
	empty, err := EvaluateStocks("0", "100", "30", nil)
	if err != nil || len(empty.Series) != 0 {
		t.Fatal("absence became zero")
	}
}
func TestStockOrderingReplayConflictAndIsolation(t *testing.T) {
	points := []StockPoint{stockPoint("80", "0", 5), stockPoint("80", "0", 5), stockPoint("90", "0", 4), stockPoint("90", "85", 4), stockPoint("90", "0", 0)}
	points[4].Stream.Resource["service.instance.id"] = "two"
	want, err := EvaluateStocks("0", "100", "30", points)
	if err != nil {
		t.Fatal(err)
	}
	if len(want.Series) != 2 || len(want.Series[0].Samples) != 2 || want.Series[0].Samples[1].State != "unknown" || want.Series[0].Samples[1].Value != nil || want.Series[0].Samples[1].StartNS != "" || *want.Series[1].Samples[0].Value != 0 {
		t.Fatalf("bad isolation/conflict: %+v", want)
	}
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}
	got, err := EvaluateStocks("0", "100", "30", points)
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatal("order changed result", err)
	}
}
func TestStockRejectsUnsupportedInputs(t *testing.T) {
	cases := map[string]func(*StockPoint){
		"delta":        func(p *StockPoint) { p.Temporality = "delta" },
		"monotonic":    func(p *StockPoint) { p.Monotonic = true },
		"epoch":        func(p *StockPoint) { p.StartNS = "91" },
		"noncanonical": func(p *StockPoint) { p.EndNS = "090" },
		"right-bound":  func(p *StockPoint) { p.EndNS = "100" },
		"left-bound":   func(p *StockPoint) { p.EndNS = "9"; p.StartNS = "0" },
		"nan":          func(p *StockPoint) { p.Value = math.NaN() },
		"infinite":     func(p *StockPoint) { p.Value = math.Inf(1) },
		"identity":     func(p *StockPoint) { delete(p.Stream.Resource, "service.instance.id") },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			p := stockPoint("90", "0", 2)
			change(&p)
			if _, err := EvaluateStocks("10", "100", "30", []StockPoint{p}); err == nil {
				t.Fatal("accepted invalid input")
			}
		})
	}
	for _, w := range [][3]string{{"0", "0", "30"}, {"0", "100", "0"}, {"-1", "100", "30"}} {
		if _, err := EvaluateStocks(w[0], w[1], w[2], nil); err == nil {
			t.Fatal("accepted invalid window")
		}
	}
	if _, err := EvaluateStocks("0", "100", "30", make([]StockPoint, 10001)); err == nil {
		t.Fatal("accepted excessive points")
	}
}

func TestStockCanonicalReplayBytes(t *testing.T) {
	a := stockPoint("90", "0", 0)
	a.Stream.Attributes = nil
	b := stockPoint("90", "0", math.Copysign(0, -1))
	b.Stream.Attributes = map[string]string{}
	first, err := EvaluateStocks("0", "100", "30", []StockPoint{a, b})
	if err != nil {
		t.Fatal(err)
	}
	second, err := EvaluateStocks("0", "100", "30", []StockPoint{b, a})
	if err != nil {
		t.Fatal(err)
	}
	x, _ := json.Marshal(first)
	y, _ := json.Marshal(second)
	if string(x) != string(y) || len(first.Series) != 1 || len(first.Series[0].Samples) != 1 || math.Signbit(*first.Series[0].Samples[0].Value) {
		t.Fatalf("noncanonical replay: %s vs %s", x, y)
	}
}

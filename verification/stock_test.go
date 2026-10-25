package verification

import (
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
	"reflect"
	"strconv"
	"testing"
)

func TestHegelStockObservations(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(2, 40))
		shift := hegel.Draw(tc, hegel.Integers(-1000, 1000))
		points := []metrics.StockPoint{}
		for i := 0; i < n; i++ {
			start := "0"
			if i >= n/2 {
				start = strconv.Itoa(n / 2)
			}
			points = append(points, metrics.StockPoint{Stream: metrics.Stream{Resource: map[string]string{"service.name": "ordinary", "service.instance.id": "one"}, Scope: "sdk", Name: "stock", Unit: "1"}, Temporality: "cumulative", StartNS: start, EndNS: strconv.Itoa(i + 1), Value: float64(shift-i) / 4})
		}
		result, err := metrics.EvaluateStocks("0", "100", "100", points)
		if err != nil {
			tc.Fatal(err)
		}
		if len(result.Series) != 1 || len(result.Series[0].Samples) != n {
			tc.Fatal("invented or dropped samples")
		}
		for i, s := range result.Series[0].Samples {
			if s.State != "observed" || s.Value == nil || *s.Value != float64(shift-i)/4 || s.StartNS != points[i].StartNS {
				tc.Fatal("stock became delta or lost epoch")
			}
		}
		for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
			points[i], points[j] = points[j], points[i]
		}
		points = append(points, points...)
		replay, err := metrics.EvaluateStocks("0", "100", "100", points)
		if err != nil || !reflect.DeepEqual(result, replay) {
			tc.Fatal("arrival order or identical replay changed observation")
		}
	})
}

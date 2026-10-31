package exploration

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

// NativeHistogramComparison projects one existing comparable native-window row.
// It performs no aggregation or source reads. Callers retain the surrounding
// comparison's global withholding policy; a withheld row cannot mint a pair.
func NativeHistogramComparison(a, b *measurements.Result, w metrics.WindowComparison) (*ViewHistogramComparison, error) {
	refuse := func(reason string) (*ViewHistogramComparison, error) {
		return nil, fmt.Errorf("native histogram comparison: %s", reason)
	}
	if a == nil || b == nil || w.State != "comparable" || w.Reasons == nil || len(w.Reasons) != 0 {
		return refuse("require a comparable whole-window row")
	}
	ad, _ := json.Marshal(a.Question.Definition)
	bd, _ := json.Marshal(b.Question.Definition)
	if !bytes.Equal(ad, bd) {
		return refuse("measurement definitions differ")
	}
	x, err := measurements.HistogramWindowView(a, w.Stream, w.Baseline)
	if err != nil {
		return nil, err
	}
	y, err := measurements.HistogramWindowView(b, w.Stream, w.Selection)
	if err != nil {
		return nil, err
	}
	if w.Baseline.State != "observed" || w.Selection.State != "observed" || x.Histogram.Total == "0" || y.Histogram.Total == "0" || x.Histogram.Unit != y.Histogram.Unit || x.Histogram.Temporality != y.Histogram.Temporality || len(x.Rows) != len(y.Rows) {
		return refuse("require positive compatible window populations")
	}
	for i := range x.Rows {
		if *x.Rows[i].Values["from"] != *y.Rows[i].Values["from"] || *x.Rows[i].Values["to"] != *y.Rows[i].Values["to"] {
			return refuse("bucket layouts differ")
		}
	}
	return &ViewHistogramComparison{Version: "goo.histogram-comparison/1", Baseline: x, Selection: y, State: "observed"}, nil
}

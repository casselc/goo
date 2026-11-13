package exploration

import (
	"encoding/json"
	"errors"
	"math/big"
	"reflect"
	"strconv"

	"github.com/casselc/goo/scopedview"
)

const RecordRateKey = "record.rate"

// RecordRateQuery admits only the initial ordinary, ungrouped record profile.
// Count exposure is the declared query window, regardless of narrower predicates.
func RecordRateQuery(q Query) bool {
	return (q.Signal == "logs" || q.Signal == "spans") && len(q.GroupBy) == 0 && !q.GroupTimeSeries && q.NumericHistogram == nil && q.ReportedEdge == nil && q.Validate() == nil
}

// RecordRateSummaryQuery also admits the existing ordinary categorical count
// profile. A population summary uses the whole count and query-window exposure;
// its grouping does not turn the rate into a per-category or interval measure.
// Timeline admission remains the narrower RecordRateQuery contract.
func RecordRateSummaryQuery(q Query) bool {
	return RecordRateQuery(q) || ValidateCapturedCountQuery(q) == nil
}

func recordRateField(q Query) ViewField {
	label := "Captured log records/s"
	if q.Signal == "spans" {
		label = "Captured span records/s"
	}
	return ViewField{Key: RecordRateKey, Label: label, Type: "rational", Unit: "records/s", Nullable: true, Aggregation: &ViewAggregation{Family: "rate", Kind: "records-per-second", Method: "exact-rational"}}
}

// RecordRateCell constructs exact count/exposure support. Zero remains derived
// zero; an unknown interval has no invented count or exposure facts.
func RecordRateCell(count string, elapsedMS int64, unknown bool) (*string, ViewCell, error) {
	if elapsedMS <= 0 {
		return nil, ViewCell{}, errors.New("record rate requires positive elapsed milliseconds")
	}
	if unknown {
		if count != "" {
			return nil, ViewCell{}, errors.New("unknown record rate cannot supply count")
		}
		return nil, ViewCell{State: "unknown"}, nil
	}
	n, err := strconv.ParseUint(count, 10, 64)
	if err != nil || strconv.FormatUint(n, 10) != count {
		return nil, ViewCell{}, errors.New("record rate requires canonical captured count")
	}
	numerator := new(big.Int).Mul(new(big.Int).SetUint64(n), big.NewInt(1000))
	value := new(big.Rat).SetFrac(numerator, big.NewInt(elapsedMS)).RatString()
	return &value, ViewCell{State: "derived", Rate: &scopedview.RateFacts{Count: count, ElapsedMS: strconv.FormatInt(elapsedMS, 10)}}, nil
}
func addRecordRate(row *ViewRow, elapsedMS int64, unknown bool) bool {
	count := ""
	if !unknown {
		if row.Values["count"] == nil {
			return false
		}
		count = *row.Values["count"]
	}
	value, cell, err := RecordRateCell(count, elapsedMS, unknown)
	if err != nil {
		return false
	}
	row.Values[RecordRateKey] = value
	if row.Cells == nil {
		row.Cells = map[string]ViewCell{}
	}
	row.Cells[RecordRateKey] = cell
	return true
}

// ValidateRecordRateDataset checks the finite rate profile and exact same-row
// proof. It does not attest source coverage, sampling or completeness.
func ValidateRecordRateDataset(d *ViewDataset) error {
	if d == nil {
		return errors.New("record rate dataset required")
	}
	var rate *ViewField
	for i := range d.Fields {
		f := &d.Fields[i]
		if f.Key == RecordRateKey || f.Aggregation != nil && f.Aggregation.Family == "rate" {
			if rate != nil {
				return errors.New("one captured record rate field required")
			}
			rate = f
		}
	}
	if rate == nil {
		for _, row := range d.Rows {
			for _, cell := range row.Cells {
				if cell.Rate != nil {
					return errors.New("rate facts require declared rate")
				}
			}
		}
		return nil
	}
	var q Query
	if json.Unmarshal(d.Scope.Selector, &q) != nil || d.Scope.Signal != q.Signal || d.Quality.Freshness != "fixed" || (d.Shape != "summary" && d.Shape != "time-series") || !reflect.DeepEqual(*rate, recordRateField(q)) {
		return errors.New("rate requires its ordinary fixed record profile")
	}
	if d.Shape == "summary" && !RecordRateSummaryQuery(q) || d.Shape == "time-series" && !RecordRateQuery(q) {
		return errors.New("rate requires its supported summary or interval profile")
	}
	for _, row := range d.Rows {
		elapsed := q.ToMS - q.FromMS
		if d.Shape == "time-series" {
			from, to := row.Values["fromMs"], row.Values["toMs"]
			if from == nil || to == nil {
				return errors.New("rate interval bounds required")
			}
			a, e := strconv.ParseInt(*from, 10, 64)
			b, f := strconv.ParseInt(*to, 10, 64)
			if e != nil || f != nil || strconv.FormatInt(a, 10) != *from || strconv.FormatInt(b, 10) != *to || a < q.FromMS || b > q.ToMS {
				return errors.New("rate interval differs from query")
			}
			elapsed = b - a
		}
		count := ""
		if row.Values["count"] != nil {
			count = *row.Values["count"]
		}
		unknown := d.Shape == "time-series" && row.State == "unknown"
		value, cell, err := RecordRateCell(count, elapsed, unknown)
		actual, ok := row.Cells[RecordRateKey]
		_, hasValue := row.Values[RecordRateKey]
		if err != nil || !ok || !hasValue || !reflect.DeepEqual(value, row.Values[RecordRateKey]) || !reflect.DeepEqual(cell, actual) {
			return errors.New("rate differs from exact captured count and elapsed time")
		}
		for key, c := range row.Cells {
			if key != RecordRateKey && c.Rate != nil {
				return errors.New("rate facts belong only to record rate")
			}
		}
	}
	return nil
}

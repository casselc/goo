package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
)

// SpanStatusClass binds the ClickHouse OTel exporter's strings. Unset is not OK;
// empty/unrecognized values remain unknown rather than silently becoming success.
func SpanStatusClass(status string) string {
	switch status {
	case "Error":
		return "error"
	case "Ok":
		return "ok"
	case "Unset":
		return "unset"
	default:
		return "unknown"
	}
}
func statusClassSQL() string {
	return "multiIf(StatusCode = 'Error', 'error', StatusCode = 'Ok', 'ok', StatusCode = 'Unset', 'unset', 'unknown')"
}
func statusAggregateSQL() string {
	return ", toString(countIf(StatusCode = 'Error')) AS status_error, toString(countIf(StatusCode = 'Ok')) AS status_ok, toString(countIf(StatusCode = 'Unset')) AS status_unset, toString(countIf(StatusCode NOT IN ('Error','Ok','Unset'))) AS status_unknown"
}

type SpanStatusSummary struct {
	Count   string `json:"count"`
	Error   string `json:"error"`
	OK      string `json:"ok"`
	Unset   string `json:"unset"`
	Unknown string `json:"unknown"`
}

func emptySpanStatus() SpanStatusSummary { return SpanStatusSummary{"0", "0", "0", "0", "0"} }
func (s SpanStatusSummary) Validate() error {
	total, e := decimal(s.Count)
	if e != nil {
		return e
	}
	sum := new(big.Int)
	for _, v := range []string{s.Error, s.OK, s.Unset, s.Unknown} {
		n, e := decimal(v)
		if e != nil {
			return e
		}
		sum.Add(sum, n)
	}
	if sum.Cmp(total) != 0 {
		return fmt.Errorf("span status counts disagree with population")
	}
	return nil
}
func (s *SpanStatusSummary) add(v SpanStatusSummary) {
	for _, pair := range [][2]*string{{&s.Count, &v.Count}, {&s.Error, &v.Error}, {&s.OK, &v.OK}, {&s.Unset, &v.Unset}, {&s.Unknown, &v.Unknown}} {
		a, _ := new(big.Int).SetString(*pair[0], 10)
		b, _ := new(big.Int).SetString(*pair[1], 10)
		*pair[0] = a.Add(a, b).String()
	}
}
func DecodeSpanStatus(raw json.RawMessage) (*SpanStatusSummary, error) {
	var r struct {
		Count   string `json:"count"`
		Error   string `json:"status_error"`
		OK      string `json:"status_ok"`
		Unset   string `json:"status_unset"`
		Unknown string `json:"status_unknown"`
	}
	if json.Unmarshal(raw, &r) != nil {
		return nil, fmt.Errorf("invalid span status row")
	}
	s := &SpanStatusSummary{r.Count, r.Error, r.OK, r.Unset, r.Unknown}
	return s, s.Validate()
}

package measurements

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// LatestCounterEndpoint inspects validated adapter observations, including bad
// intervals. It never falls back to an older healthy interval across a reset.
func LatestCounterEndpoint(r *Result) (int64, error) {
	if r == nil || r.Question.Definition.Kind != "counter" {
		return 0, fmt.Errorf("counter search required")
	}
	var latest int64
	for _, raw := range r.Rows {
		var row metricRow
		if e := json.Unmarshal(raw, &row); e != nil {
			return 0, e
		}
		n, e := strconv.ParseInt(row.EndNS, 10, 64)
		if e != nil || n < 0 {
			return 0, fmt.Errorf("invalid counter endpoint")
		}
		if n > latest {
			latest = n
		}
	}
	return latest, nil
}

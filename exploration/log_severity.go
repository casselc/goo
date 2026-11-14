package exploration

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
)

// LogSeverityClass follows OTel SeverityNumber, independently of SeverityText.
// The ClickHouse exporter retains an unsigned physical number: 0 is unspecified;
// values above the OTel range remain invalid rather than becoming errors/fatal.
func LogSeverityClass(number uint64) string {
	switch {
	case number == 0:
		return "unspecified"
	case number <= 4:
		return "trace"
	case number <= 8:
		return "debug"
	case number <= 12:
		return "info"
	case number <= 16:
		return "warn"
	case number <= 20:
		return "error"
	case number <= 24:
		return "fatal"
	default:
		return "invalid"
	}
}

type LogSeveritySummary struct {
	Count       string `json:"count"`
	Trace       string `json:"trace"`
	Debug       string `json:"debug"`
	Info        string `json:"info"`
	Warn        string `json:"warn"`
	Error       string `json:"error"`
	Fatal       string `json:"fatal"`
	Unspecified string `json:"unspecified"`
	Invalid     string `json:"invalid"`
}

func emptyLogSeverity() LogSeveritySummary {
	return LogSeveritySummary{"0", "0", "0", "0", "0", "0", "0", "0", "0"}
}

func (s LogSeveritySummary) Validate() error {
	total, err := decimal(s.Count)
	if err != nil {
		return err
	}
	sum := new(big.Int)
	for _, value := range []string{s.Trace, s.Debug, s.Info, s.Warn, s.Error, s.Fatal, s.Unspecified, s.Invalid} {
		n, err := decimal(value)
		if err != nil {
			return err
		}
		sum.Add(sum, n)
	}
	if sum.Cmp(total) != 0 {
		return fmt.Errorf("log severity bands disagree with population")
	}
	return nil
}

func (s *LogSeveritySummary) add(value LogSeveritySummary) {
	for _, pair := range [][2]*string{{&s.Count, &value.Count}, {&s.Trace, &value.Trace}, {&s.Debug, &value.Debug}, {&s.Info, &value.Info}, {&s.Warn, &value.Warn}, {&s.Error, &value.Error}, {&s.Fatal, &value.Fatal}, {&s.Unspecified, &value.Unspecified}, {&s.Invalid, &value.Invalid}} {
		a, _ := new(big.Int).SetString(*pair[0], 10)
		b, _ := new(big.Int).SetString(*pair[1], 10)
		*pair[0] = a.Add(a, b).String()
	}
}

func logSeverityAggregateSQL() string {
	parts := []string{}
	for _, band := range []struct{ name, predicate string }{
		{"trace", "SeverityNumber BETWEEN 1 AND 4"}, {"debug", "SeverityNumber BETWEEN 5 AND 8"},
		{"info", "SeverityNumber BETWEEN 9 AND 12"}, {"warn", "SeverityNumber BETWEEN 13 AND 16"},
		{"error", "SeverityNumber BETWEEN 17 AND 20"}, {"fatal", "SeverityNumber BETWEEN 21 AND 24"},
		{"unspecified", "SeverityNumber = 0"}, {"invalid", "SeverityNumber > 24"},
	} {
		parts = append(parts, "toString(countIf("+band.predicate+")) AS severity_"+band.name)
	}
	return ", " + strings.Join(parts, ", ")
}

func DecodeLogSeverity(raw json.RawMessage) (*LogSeveritySummary, error) {
	var row struct {
		Count       string `json:"count"`
		Trace       string `json:"severity_trace"`
		Debug       string `json:"severity_debug"`
		Info        string `json:"severity_info"`
		Warn        string `json:"severity_warn"`
		Error       string `json:"severity_error"`
		Fatal       string `json:"severity_fatal"`
		Unspecified string `json:"severity_unspecified"`
		Invalid     string `json:"severity_invalid"`
	}
	if json.Unmarshal(raw, &row) != nil {
		return nil, fmt.Errorf("invalid log severity row")
	}
	summary := &LogSeveritySummary{row.Count, row.Trace, row.Debug, row.Info, row.Warn, row.Error, row.Fatal, row.Unspecified, row.Invalid}
	return summary, summary.Validate()
}

// Each independently read projection must conserve every band, not just count.
func decodeLogSeverityGroups(groups []json.RawMessage, total *LogSeveritySummary) (bool, error) {
	sum := emptyLogSeverity()
	for i, raw := range groups {
		value, err := DecodeLogSeverity(raw)
		if err != nil || value.Count == "0" {
			return false, fmt.Errorf("invalid grouped log severity")
		}
		sum.add(*value)
		var group map[string]json.RawMessage
		if json.Unmarshal(raw, &group) != nil {
			return false, fmt.Errorf("invalid log severity group")
		}
		for _, key := range []string{"trace", "debug", "info", "warn", "error", "fatal", "unspecified", "invalid"} {
			delete(group, "severity_"+key)
		}
		group["logSeverity"], _ = json.Marshal(value)
		groups[i], _ = json.Marshal(group)
	}
	return sum != *total, nil
}

func logSeveritySeriesAgree(total *LogSeveritySummary, points []TimePoint) bool {
	sum := emptyLogSeverity()
	for _, point := range points {
		if point.LogSeverity == nil || point.LogSeverity.Validate() != nil || point.LogSeverity.Count != point.Count {
			return false
		}
		sum.add(*point.LogSeverity)
	}
	return sum == *total
}

// Saved result admission must not accept unrequested, missing or inconsistent
// severity. A changed-between-reads result retains diagnostics, not a census.
func validateLogSeverityResult(result *Result) error {
	if !result.Question.LogSeverity {
		if result.LogSeverity != nil {
			return fmt.Errorf("unexpected log severity summary")
		}
		for _, raw := range result.Groups {
			var group map[string]json.RawMessage
			if json.Unmarshal(raw, &group) != nil || group["logSeverity"] != nil {
				return fmt.Errorf("unexpected grouped log severity")
			}
		}
		return nil
	}
	if result.LogSeverity == nil || result.LogSeverity.Count != result.Count || result.LogSeverity.Validate() != nil {
		return fmt.Errorf("invalid log severity summary")
	}
	if len(result.Question.GroupBy) > 0 {
		sum := emptyLogSeverity()
		for _, raw := range result.Groups {
			var group struct {
				Count string              `json:"count"`
				Value *LogSeveritySummary `json:"logSeverity"`
			}
			if json.Unmarshal(raw, &group) != nil || group.Value == nil || group.Value.Validate() != nil || group.Value.Count != group.Count {
				return fmt.Errorf("invalid saved grouped log severity")
			}
			sum.add(*group.Value)
		}
		if result.State != "changed_between_reads" && sum != *result.LogSeverity {
			return fmt.Errorf("saved log severity groups disagree")
		}
	}
	if result.Question.BucketMS != 0 && result.State != "changed_between_reads" && !logSeveritySeriesAgree(result.LogSeverity, result.TimeSeries) {
		return fmt.Errorf("saved log severity intervals disagree")
	}
	return nil
}

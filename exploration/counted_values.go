package exploration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"
	"unicode/utf8"
)

const CountedValuesVersion = "goo/counted-values/v1"

type CountedValue struct {
	Value *string `json:"g0"`
	Count string  `json:"count"`
}
type CountedValuesRead struct {
	ReadAt string `json:"readAt"`
	Basis  string `json:"basis"`
}
type CountedValuesReads struct {
	Count  CountedValuesRead `json:"count"`
	Groups CountedValuesRead `json:"groups"`
}

// Groups in changed_between_reads are diagnostics only: they must not be used
// as an actionable population ranking or conserved snapshot.
type CountedValuesResult struct {
	Version    string             `json:"version"`
	Question   Query              `json:"question"`
	CapturedAt string             `json:"capturedAt"`
	Basis      string             `json:"basis"`
	Count      string             `json:"count"`
	State      string             `json:"state"`
	Groups     []CountedValue     `json:"groups"`
	Reads      CountedValuesReads `json:"reads"`
}

func ValidateCountedValuesQuery(q Query) error {
	if err := q.Validate(); err != nil {
		return err
	}
	if len(q.GroupBy) != 1 || q.Numeric != nil || q.NumericHistogram != nil || q.Summary != "" || q.BucketMS != 0 || q.GroupTimeSeries || q.SpanStatus || q.Offset != 0 {
		return fmt.Errorf("counted values requires one string group, offset zero and no measure or series projections")
	}
	typ, e := q.GroupBy[0].Type(q.Signal)
	if e != nil {
		return e
	}
	if typ != "string" {
		return fmt.Errorf("counted values requires a string group")
	}
	return nil
}
func countedValuesHash(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
func countedValuesUint(text string, positive bool) (uint64, error) {
	n, e := strconv.ParseUint(text, 10, 64)
	if e != nil || strconv.FormatUint(n, 10) != text || positive && n == 0 {
		return 0, fmt.Errorf("invalid canonical count")
	}
	return n, nil
}

// FetchCountedValues makes exactly two bounded aggregate reads. Clocks identify
// separate observations, not a transactional population snapshot.
func (s Source) FetchCountedValues(ctx context.Context, q Query) (*CountedValuesResult, error) {
	if e := ValidateCountedValuesQuery(q); e != nil {
		return nil, e
	}
	// Detach caller-owned predicates and reported-edge scope before any I/O.
	encoded, e := json.Marshal(q)
	if e != nil {
		return nil, e
	}
	var question Query
	if e = json.Unmarshal(encoded, &question); e != nil {
		return nil, e
	}
	plan, e := Compile(question)
	if e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	rawCount, e := s.read(ctx, plan.Count, plan.Parameters)
	if e != nil {
		return nil, e
	}
	countRead := CountedValuesRead{ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Basis: countedValuesHash(rawCount)}
	var total struct {
		Count *string `json:"count"`
	}
	if len(rawCount) != 1 || json.Unmarshal(rawCount[0], &total) != nil || total.Count == nil {
		return nil, fmt.Errorf("invalid counted-values count result")
	}
	count, e := countedValuesUint(*total.Count, false)
	if e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	rawGroups, e := s.read(ctx, plan.Groups, plan.Parameters)
	if e != nil {
		return nil, e
	}
	groupRead := CountedValuesRead{ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Basis: countedValuesHash(rawGroups)}
	if len(rawGroups) > 1000 {
		return nil, fmt.Errorf("more than 1000 groups; narrow query; no partial grouping totals")
	}
	groups := make([]CountedValue, 0, len(rawGroups))
	seen := map[string]bool{}
	sum := new(big.Int)
	field := question.GroupBy[0]
	for _, raw := range rawGroups {
		var obj map[string]json.RawMessage
		if json.Unmarshal(raw, &obj) != nil {
			return nil, fmt.Errorf("invalid counted-values group")
		}
		valueRaw, ok := obj["g0"]
		if !ok {
			return nil, fmt.Errorf("counted-values group missing g0")
		}
		var value *string
		if !bytes.Equal(bytes.TrimSpace(valueRaw), []byte("null")) {
			var text string
			if json.Unmarshal(valueRaw, &text) != nil || !utf8.Valid(valueRaw) {
				return nil, fmt.Errorf("invalid counted-values group value")
			}
			value = &text
		}
		if value == nil && !field.CanBeMissing() {
			return nil, fmt.Errorf("nonnullable counted-values column is missing")
		}
		if value != nil && field.Kind == "column" && field.Key == "body_pattern_v1" {
			if _, e := DecodeLogPattern(*value); e != nil {
				return nil, e
			}
		}
		var groupCount string
		if b, ok := obj["count"]; !ok || json.Unmarshal(b, &groupCount) != nil {
			return nil, fmt.Errorf("invalid counted-values group count")
		}
		n, e := countedValuesUint(groupCount, true)
		if e != nil {
			return nil, e
		}
		key := string(valueRaw)
		if value == nil {
			key = "null"
		} else {
			b, _ := json.Marshal(*value)
			key = string(b)
		}
		if seen[key] {
			return nil, fmt.Errorf("duplicate counted-values group")
		}
		seen[key] = true
		sum.Add(sum, new(big.Int).SetUint64(n))
		groups = append(groups, CountedValue{Value: value, Count: groupCount})
	}
	state := "observed"
	if count == 0 {
		state = "no_data"
	}
	if sum.Cmp(new(big.Int).SetUint64(count)) != 0 {
		state = "changed_between_reads"
	}
	r := &CountedValuesResult{Version: CountedValuesVersion, Question: question, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Count: *total.Count, State: state, Groups: groups, Reads: CountedValuesReads{Count: countRead, Groups: groupRead}}
	basis := *r
	basis.CapturedAt = ""
	basis.Reads.Count.ReadAt = ""
	basis.Reads.Groups.ReadAt = ""
	r.Basis = countedValuesHash(basis)
	return r, nil
}

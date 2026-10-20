package exploration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const TailWindowMS int64 = 60000
const TailReadLimit = 1000

// TailSelector deliberately has no fixed window or aggregate settings.
// Every poll reads a rolling event-time window; this is not an ingestion cursor.
type TailSelector struct {
	Environment string     `json:"environment"`
	Where       *Predicate `json:"where,omitempty"`
}

func (s TailSelector) Query(now int64) Query {
	return Query{Signal: "logs", Environment: s.Environment, Where: s.Where, FromMS: now - TailWindowMS, ToMS: now, Limit: 100}
}
func (s Source) PollTail(ctx context.Context, selector TailSelector, now int64) ([]json.RawMessage, bool, error) {
	q := selector.Query(now)
	p, e := compile(q, TailReadLimit)
	if e != nil {
		return nil, false, e
	}
	// Bias bounded inspection toward recent events. Equal timestamps use the
	// compiler's remaining deterministic ordering; no physical identity is claimed.
	sql := strings.Replace(p.Rows, " ORDER BY Timestamp,", " ORDER BY Timestamp DESC,", 1)
	rows, e := s.read(ctx, sql, p.Parameters)
	if e != nil {
		return nil, false, e
	}
	limited := len(rows) > TailReadLimit
	if limited {
		rows = rows[:TailReadLimit]
	}
	size := 0
	for i, row := range rows {
		size += len(row)
		if size > 512<<10 {
			rows = rows[:i]
			limited = true
			break
		}
	}
	// Present each bounded poll in event-time order so retaining the tail of a
	// batch keeps its newest observations. Later polls may still append late data.
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows, limited, nil
}

type TailRow struct {
	Key    string          `json:"key"`
	Record json.RawMessage `json:"record"`
	Late   bool            `json:"late"`
}

// TailTracker tracks maximum observed multiplicity per canonical record within
// one bounded connection. IDs are observation handles, not physical event IDs.
type TailTracker struct {
	seen    map[string]int
	maximum int64
}

func (t *TailTracker) Observe(rows []json.RawMessage) ([]TailRow, error) {
	if t.seen == nil {
		t.seen = map[string]int{}
	}
	counts := map[string]int{}
	out := []TailRow{}
	maximum := t.maximum
	for _, row := range rows {
		var value map[string]any
		d := json.NewDecoder(bytes.NewReader(row))
		d.UseNumber()
		if e := d.Decode(&value); e != nil {
			return nil, e
		}
		stamp, ok := value["timestamp_ns"].(string)
		if !ok {
			return nil, fmt.Errorf("tail record has no event time")
		}
		ns, e := strconv.ParseInt(stamp, 10, 64)
		if e != nil || ns < 0 {
			return nil, fmt.Errorf("invalid tail event time")
		}
		canonical, e := json.Marshal(value)
		if e != nil {
			return nil, e
		}
		hash := sha256.Sum256(canonical)
		key := hex.EncodeToString(hash[:])
		counts[key]++
		if counts[key] > t.seen[key] {
			out = append(out, TailRow{Key: fmt.Sprintf("%s:%d", key, counts[key]), Record: append(json.RawMessage(nil), row...), Late: ns < t.maximum})
		}
		if ns > maximum {
			maximum = ns
		}
	}
	if len(t.seen)+len(counts) > 20000 {
		return nil, fmt.Errorf("tail identity budget reached; reconnect with a continuity gap")
	}
	for key, n := range counts {
		if n > t.seen[key] {
			t.seen[key] = n
		}
	}
	t.maximum = maximum
	return out, nil
}

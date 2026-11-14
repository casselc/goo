package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/casselc/goo/internal/sourcecap"
	"github.com/casselc/goo/internal/sourcefailure"
)

// SourceCapabilitiesQuestion has no implicit environment, service or family.
// It performs bounded source discovery, never an all-environments Query.
type SourceCapabilitiesQuestion struct {
	FromMS int64 `json:"fromMs"`
	ToMS   int64 `json:"toMs"`
}

func (q SourceCapabilitiesQuestion) Validate() error {
	if q.FromMS < 0 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 120*60000 || q.ToMS > 4102444800000 {
		return fmt.Errorf("choose a half-open window of at most 120 minutes before 2100")
	}
	return nil
}

type SourceObservation struct {
	State string `json:"state"`
	Unit  string `json:"unit"`
}
type SourceCapability struct {
	sourcecap.Family
	Observation SourceObservation `json:"observation"`
}
type SourceCapabilitiesResult struct {
	Version    string                     `json:"version"`
	Question   SourceCapabilitiesQuestion `json:"question"`
	CapturedAt string                     `json:"capturedAt"`
	Families   []SourceCapability         `json:"families"`
}

// DiscoverSourceCapabilities reports readiness separately from a LIMIT 1 presence
// witness. An empty witness is not a population total or an ingestion verdict.
func (s Source) DiscoverSourceCapabilities(ctx context.Context, q SourceCapabilitiesQuestion) (*SourceCapabilitiesResult, error) {
	if err := q.Validate(); err != nil {
		return nil, err
	}
	families, err := sourcecap.Inspect(ctx, s.read)
	if err != nil {
		return nil, err
	}
	result := &SourceCapabilitiesResult{Version: "goo/source-capabilities/v1", Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Families: make([]SourceCapability, 0, len(families))}
	params := map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10)}
	for _, family := range families {
		unit, clock := "records", "Timestamp"
		if family.Kind != "logs" && family.Kind != "spans" {
			unit = "exported-points"
			clock = "TimeUnix"
		}
		item := SourceCapability{Family: family, Observation: SourceObservation{State: "unknown", Unit: unit}}
		if family.Compatible() {
			sql := "SELECT 1 AS observed FROM " + family.Table + " WHERE " + clock + ">=fromUnixTimestamp64Milli({from:Int64}) AND " + clock + "<fromUnixTimestamp64Milli({to:Int64}) LIMIT 1 FORMAT JSONEachRow"
			rows, readErr := s.read(ctx, sql, params)
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if readErr != nil {
				item.Observation.State = "unreadable"
				item.Detail = sourcefailure.Describe(readErr).Error
			} else if len(rows) == 0 {
				item.Observation.State = "empty"
			} else {
				var witness struct {
					Observed int `json:"observed"`
				}
				if len(rows) != 1 || json.Unmarshal(rows[0], &witness) != nil || witness.Observed != 1 {
					item.Observation.State = "unreadable"
					item.Detail = "Source presence response rejected."
				} else {
					item.Observation.State = "observed"
				}
			}
		}
		result.Families = append(result.Families, item)
	}
	return result, nil
}

package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/casselc/goo/identity"
)

const EnvironmentScopeLimit = 64

// AdoptionWindow is deliberately not Query: no implicit named scope, arbitrary
// predicates, snapshot bank, publication or unsupported continuation is offered.
type AdoptionWindow struct {
	Signal string `json:"signal"`
	FromMS int64  `json:"fromMs"`
	ToMS   int64  `json:"toMs"`
}

func (q AdoptionWindow) Validate() error {
	if q.Signal != "logs" && q.Signal != "spans" || q.FromMS < 0 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 120*60000 || q.ToMS > 4102444800000 {
		return fmt.Errorf("choose logs or spans and a half-open window of at most 120 minutes before 2100")
	}
	return nil
}
func (q AdoptionWindow) params() map[string]string {
	return map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10)}
}
func (q AdoptionWindow) source() string {
	if q.Signal == "logs" {
		return "otel_logs"
	}
	return "otel_traces"
}
func (q AdoptionWindow) windowSQL() string {
	return "Timestamp>=fromUnixTimestamp64Milli({from:Int64}) AND Timestamp<fromUnixTimestamp64Milli({to:Int64})"
}

// EnvironmentPredicate lowers only a checked logical selector. Named selection
// is presence-aware current-first fallback, unlike legacy Query's current-only rule.
func EnvironmentPredicate(scope identity.EnvironmentScope) (string, map[string]string, error) {
	selection, err := identity.EffectiveEnvironment(scope)
	if err != nil {
		return "", nil, err
	}
	return environmentSelectionPredicate(selection)
}

type EnvironmentScopeGroup struct {
	Observation identity.EnvironmentObservation `json:"observation"`
	Count       string                          `json:"count"`
	Selectable  bool                            `json:"selectable"`
}
type EnvironmentScopesResult struct {
	Version  string                  `json:"version"`
	Question AdoptionWindow          `json:"question"`
	ReadAt   string                  `json:"readAt"`
	Groups   []EnvironmentScopeGroup `json:"groups"`
	HasMore  bool                    `json:"hasMore"`
}

func (s Source) DiscoverEnvironmentScopes(ctx context.Context, q AdoptionWindow) (*EnvironmentScopesResult, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	sql := "SELECT mapContains(ResourceAttributes,'" + identity.EnvironmentKey + "') AS current_present, mapContains(ResourceAttributes,'" + identity.LegacyEnvironmentKey + "') AS legacy_present, ResourceAttributes['" + identity.EnvironmentKey + "'] AS current_value, ResourceAttributes['" + identity.LegacyEnvironmentKey + "'] AS legacy_value, toString(count()) AS count FROM " + q.source() + " WHERE " + q.windowSQL() + " GROUP BY current_present,legacy_present,current_value,legacy_value ORDER BY current_present,legacy_present,current_value,legacy_value LIMIT 65 FORMAT JSONEachRow"
	rows, e := s.read(ctx, sql, q.params())
	if e != nil {
		return nil, e
	}
	out := &EnvironmentScopesResult{Version: "goo/environment-scopes/v1", Question: q, ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Groups: []EnvironmentScopeGroup{}, HasMore: len(rows) > EnvironmentScopeLimit}
	if len(rows) > 65 {
		return nil, fmt.Errorf("environment scope response exceeds group bound")
	}
	retained := 0
	for _, wire := range rows[:min(len(rows), EnvironmentScopeLimit)] {
		retained += len(wire)
		if retained > 1<<20 {
			return nil, fmt.Errorf("environment scope diagnostics exceed 1 MiB; narrow the read window")
		}
		var row struct {
			CurrentPresent *uint8  `json:"current_present"`
			LegacyPresent  *uint8  `json:"legacy_present"`
			CurrentValue   *string `json:"current_value"`
			LegacyValue    *string `json:"legacy_value"`
			Count          string  `json:"count"`
		}
		if e = json.Unmarshal(wire, &row); e != nil || row.CurrentPresent == nil || row.LegacyPresent == nil || *row.CurrentPresent > 1 || *row.LegacyPresent > 1 || row.CurrentValue == nil || row.LegacyValue == nil {
			return nil, fmt.Errorf("invalid environment scope row")
		}
		n, e := strconv.ParseUint(row.Count, 10, 64)
		if e != nil || n == 0 || strconv.FormatUint(n, 10) != row.Count {
			return nil, fmt.Errorf("invalid environment scope count")
		}
		if (*row.CurrentPresent == 0 && *row.CurrentValue != "") || (*row.LegacyPresent == 0 && *row.LegacyValue != "") {
			return nil, fmt.Errorf("absent environment key has a reported value")
		}
		resource := map[string]string{}
		if *row.CurrentPresent == 1 {
			resource[identity.EnvironmentKey] = *row.CurrentValue
		}
		if *row.LegacyPresent == 1 {
			resource[identity.LegacyEnvironmentKey] = *row.LegacyValue
		}
		observation := identity.ObserveEnvironment(resource)
		out.Groups = append(out.Groups, EnvironmentScopeGroup{observation, row.Count, observation.Scope.Validate() == nil})
	}
	return out, nil
}

type AdoptionRecordsQuestion struct {
	AdoptionWindow
	Scope identity.EnvironmentScope `json:"scope"`
	Limit int                       `json:"limit"`
}

func (q AdoptionRecordsQuestion) Validate() error {
	if e := q.AdoptionWindow.Validate(); e != nil {
		return e
	}
	if e := q.Scope.Validate(); e != nil {
		return e
	}
	if q.Limit < 1 || q.Limit > 100 {
		return fmt.Errorf("adoption preview limit must be 1..100")
	}
	return nil
}

type AdoptionRecordsResult struct {
	Selector identity.EnvironmentSelector `json:"selector"`
	Version  string                       `json:"version"`
	Question AdoptionRecordsQuestion      `json:"question"`
	ReadAt   string                       `json:"readAt"`
	Rows     []json.RawMessage            `json:"rows"`
	Limited  bool                         `json:"limited"`
}

func (s Source) PreviewAdoptionRecords(ctx context.Context, q AdoptionRecordsQuestion) (*AdoptionRecordsResult, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	selection, e := identity.EffectiveEnvironment(q.Scope)
	if e != nil {
		return nil, e
	}
	selector, e := selection.Selector()
	if e != nil {
		return nil, e
	}
	predicate, params, e := environmentSelectionPredicate(selection)
	if e != nil {
		return nil, e
	}
	for k, v := range q.params() {
		params[k] = v
	}
	projection, order := recordProjection(q.Signal)
	rows, e := s.read(ctx, "SELECT "+projection+" FROM "+q.source()+" WHERE "+q.windowSQL()+" AND "+predicate+" ORDER BY "+order+fmt.Sprintf(" LIMIT %d FORMAT JSONEachRow", q.Limit+1), params)
	if e != nil {
		return nil, e
	}
	if len(rows) > q.Limit+1 {
		return nil, fmt.Errorf("adoption preview exceeded row bound")
	}
	out := &AdoptionRecordsResult{Selector: selector, Version: "goo/adoption-records/v1", Question: q, ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Rows: []json.RawMessage{}, Limited: len(rows) > q.Limit}
	retained := 0
	budgetLimited := false
	for i, wire := range rows {
		if e = validateSelectedRecord(selection, q.Signal, q.FromMS, q.ToMS, wire); e != nil {
			return nil, e
		}
		if i >= q.Limit || budgetLimited {
			continue
		}
		retained += len(wire)
		if retained > 1<<20 {
			out.Limited = true
			budgetLimited = true
			continue
		}
		out.Rows = append(out.Rows, wire)
	}
	return out, nil
}

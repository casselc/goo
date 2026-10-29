package lifecycle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/casselc/goo/subjectfacts"
)

type scopeKey struct{ Environment, Tenant, Service string }
type subjectKey struct {
	Scope scopeKey
	ID    string
}
type eventKey struct {
	Scope scopeKey
	ID    string
}
type normalized struct {
	Event
	time int64
}

func scopeOf(environment, tenant, service string) scopeKey {
	return scopeKey{environment, tenant, service}
}

func encoded(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func orderedKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return encoded(keys[i]) < encoded(keys[j]) })
	return keys
}

func (p Policy) Prepare(events []Event) (*Prepared, error) {
	raw := make([]json.RawMessage, len(events))
	for i, e := range events {
		raw[i], _ = json.Marshal(e)
	}
	return p.PrepareRaw(raw)
}

// PrepareRaw retains per-record validation failures as evidence impairment.
// Invalid JSON syntax for the complete input is an ingestion error instead.
func (p Policy) PrepareRaw(raw []json.RawMessage) (*Prepared, error) {
	b, err := p.validate()
	if err != nil {
		return nil, err
	}
	// Sort copies: callers' policy/event inputs are never mutated.
	p.Scopes = append([]Scope(nil), p.Scopes...)
	p.TerminalOutcomes = append([]string(nil), p.TerminalOutcomes...)
	sort.Slice(p.Scopes, func(i, j int) bool { return encoded(p.Scopes[i]) < encoded(p.Scopes[j]) })
	sort.Strings(p.TerminalOutcomes)
	out := &Prepared{Subjects: []Subject{}, Facts: []subjectfacts.Fact{}}
	r := &out.Receipt
	r.Policy = p
	r.CohortStartNS, r.CohortEndNS = strconv.FormatInt(b.start, 10), strconv.FormatInt(b.end, 10)
	r.Counts = map[string]int{"eligible": 0, "good": 0, "bad": 0, "unknown": 0, "pending": 0}
	receipts := map[scopeKey]*ScopeReceipt{}
	reasons := map[scopeKey]map[string]bool{}
	for _, s := range p.Scopes {
		k := scopeOf(s.Environment, s.TenantID, s.ServiceName)
		receipts[k] = &ScopeReceipt{Scope: s, Reasons: []string{}}
		reasons[k] = map[string]bool{}
		if s.Coverage != "complete" {
			reasons[k]["coverage_"+s.Coverage] = true
		}
	}
	poisoned := map[subjectKey]bool{}
	invalid := func(e Event, reason string) {
		r.InvalidRecords++
		k := scopeOf(e.Environment, e.TenantID, e.ServiceName)
		if receipt := receipts[k]; receipt != nil {
			receipt.InvalidRecords++
			reasons[k][reason] = true
			if e.SubjectID != "" {
				poisoned[subjectKey{k, e.SubjectID}] = true
			}
		} else {
			// Missing scope fields cannot establish that a malformed record is
			// outside the request. Conservatively impair all declared scopes.
			for k, receipt := range receipts {
				receipt.InvalidRecords++
				reasons[k]["unscoped_invalid_record"] = true
			}
		}
	}
	variants := map[eventKey]map[string]normalized{}
	canonicalInput := make([]string, 0, len(raw))
	visibleInput := []string{}
	for _, record := range raw {
		// UseNumber avoids float conversion while digesting malformed numeric
		// timestamps. All event times still require strings in the event API.
		var value any
		decoder := json.NewDecoder(bytes.NewReader(record))
		decoder.UseNumber()
		if err := decoder.Decode(&value); err != nil {
			return nil, fmt.Errorf("invalid event JSON: %w", err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("each event must be a single JSON value")
		}
		canonicalInput = append(canonicalInput, encoded(value))
		var e Event
		dec := json.NewDecoder(bytes.NewReader(record))
		dec.DisallowUnknownFields()
		decodeErr := dec.Decode(&e)
		k := scopeOf(e.Environment, e.TenantID, e.ServiceName)
		if e.Environment != "" && e.TenantID != "" && e.ServiceName != "" && receipts[k] == nil {
			r.OutsideScope++
			continue
		}
		received, receivedErr := parseNS(e.ReceivedTimeNS)
		if receivedErr == nil && received >= b.cutoff {
			r.AfterCutoff++
			continue
		}
		r.VisibleDeliveries++
		visibleInput = append(visibleInput, encoded(value))
		time, timeErr := parseNS(e.EventTimeNS)
		if decodeErr != nil || receivedErr != nil || timeErr != nil || receipts[k] == nil || e.EventID == "" || e.SubjectID == "" {
			invalid(e, "invalid_event")
			continue
		}
		if p.KnowledgeAfterEvaluation && time > b.evaluation {
			invalid(e, "beyond_evaluation")
			continue
		}
		if time > received {
			invalid(e, "clock_violation")
			continue
		}
		if e.Kind != "accepted" && e.Kind != "terminal" && e.Kind != "retry" {
			invalid(e, "invalid_event_kind")
			continue
		}
		if e.Kind == "accepted" && (e.Release == "" || e.Outcome != "") ||
			e.Kind == "terminal" && !has(p.TerminalOutcomes, e.Outcome) ||
			e.Kind == "retry" && e.Outcome != "" {
			invalid(e, "invalid_event_payload")
			continue
		}
		e.EventTimeNS = strconv.FormatInt(time, 10)
		e.ReceivedTimeNS = "" // Delivery time is not immutable business payload.
		ek := eventKey{k, e.EventID}
		if variants[ek] == nil {
			variants[ek] = map[string]normalized{}
		}
		payload := encoded(e)
		if _, exists := variants[ek][payload]; exists {
			r.DuplicateDeliveries++
		} else {
			variants[ek][payload] = normalized{e, time}
		}
	}
	bySubject := map[subjectKey][]normalized{}
	for _, k := range orderedKeys(variants) {
		values := variants[k]
		if len(values) > 1 {
			r.ConflictingEventIDs++
			reasons[k.Scope]["conflicting_event_id"] = true
		}
		for _, payload := range orderedKeys(values) {
			e := values[payload]
			sk := subjectKey{scopeOf(k.Scope.Environment, "a", k.Scope.Service), e.SubjectID}
			bySubject[sk] = append(bySubject[sk], e)
			if len(values) > 1 {
				poisoned[sk] = true
			}
		}
	}
	for _, key := range orderedKeys(bySubject) {
		events := bySubject[key]
		accepts := map[string]normalized{}
		terminals := map[string]normalized{}
		ids := map[string]bool{}
		for _, e := range events {
			ids[e.EventID] = true
			if e.Kind == "accepted" {
				accepts[encoded([]string{e.EventTimeNS, e.Release})] = e
			} else if e.Kind == "terminal" {
				terminals[encoded([]string{e.EventTimeNS, e.Outcome})] = e
			}
		}
		if len(accepts) == 0 {
			receipts[key.Scope].OrphanSubjects++
			reason := "orphan_retry"
			if len(terminals) > 0 {
				reason = "orphan_terminal"
			}
			reasons[key.Scope][reason] = true
			continue
		}
		selected, inCohort, mature := false, false, false
		for _, a := range accepts {
			selected = selected || a.time >= b.start && a.time <= b.evaluation
			inCohort = inCohort || a.time >= b.start && a.time < b.end
			mature = mature || a.time <= b.end
		}
		if !selected {
			continue
		}
		s := Subject{Environment: key.Scope.Environment, TenantID: key.Scope.Tenant,
			ServiceName: key.Scope.Service, SubjectID: key.ID,
			InCohort: inCohort, IsMature: mature, EventIDs: orderedKeys(ids)}
		var a normalized
		if len(accepts) == 1 {
			for _, e := range accepts {
				a = e
			}
			s.AcceptedTimeNS, s.Release = a.EventTimeNS, a.Release
		}
		s.Classification, s.Reason = "unknown", "conflicting_acceptance"
		switch {
		case poisoned[key]:
			s.Reason = "invalid_or_conflicting_event"
		case len(accepts) != 1:
			// No arbitrary minimum timestamp or release becomes authoritative.
		case len(terminals) > 1:
			s.Reason = "conflicting_terminal"
		default:
			var terminal normalized
			for _, e := range terminals {
				terminal = e
			}
			if len(terminals) == 1 {
				s.TerminalTimeNS = terminal.EventTimeNS
			}
			switch {
			case len(terminals) == 1 && terminal.time < a.time:
				s.Reason = "negative_duration"
			case !mature:
				s.Classification, s.Reason = "pending", "not_mature"
			case len(terminals) == 0:
				// Mature implies a.time + deadline + allowance <= T, so
				// these additions cannot overflow after bound validation.
				if receipts[key.Scope].Scope.Coverage == "complete" && b.cutoff > a.time+b.deadline+b.allowance {
					s.Classification, s.Reason = "bad", "no_completion"
				} else {
					s.Reason = "absence_not_established"
				}
			case terminal.Outcome != p.SuccessOutcome:
				s.Classification, s.Reason = "bad", "terminal_failure"
			case terminal.time <= a.time+b.deadline:
				s.Classification, s.Reason = "good", "completed_on_time"
			default:
				s.Classification, s.Reason = "bad", "completed_late"
			}
		}
		if s.Classification == "unknown" && inCohort {
			reasons[key.Scope][s.Reason] = true
		}
		if inCohort {
			r.Counts["eligible"]++
			r.Counts[s.Classification]++
		} else if s.Classification == "pending" {
			r.Counts["pending"]++
		}
		out.Subjects = append(out.Subjects, s)
	}
	r.State = "ok"
	for _, k := range orderedKeys(receipts) {
		sr := receipts[k]
		sr.Reasons = orderedKeys(reasons[k])
		sr.Quality = "complete"
		if len(sr.Reasons) > 0 {
			sr.Quality = "impaired"
			if len(sr.Reasons) == 1 && sr.Reasons[0] == "coverage_unverified" {
				sr.Quality = "unverified"
			}
			r.State = "unknown"
		}
		r.Scopes = append(r.Scopes, *sr)
	}
	if r.State == "ok" && r.Counts["eligible"] == 0 {
		r.State = "no_data"
	}
	for _, s := range out.Subjects {
		quality := receipts[scopeOf(s.Environment, s.TenantID, s.ServiceName)].Quality
		out.Facts = append(out.Facts, subjectfacts.Fact{
			"evaluation_id": p.EvaluationID, "environment": s.Environment, "tenant_id": s.TenantID,
			"service_name": s.ServiceName, p.SubjectField: s.SubjectID, "release": s.Release,
			"classification": s.Classification, "quality_impaired": quality != "complete",
			"is_mature": s.IsMature, "in_cohort": s.InCohort,
		})
	}
	sort.Strings(canonicalInput)
	sort.Strings(visibleInput)
	r.InputManifestDigest = digest(canonicalInput)
	r.EvidenceDigest = digest(visibleInput)
	r.SnapshotDigest = digest(struct {
		Policy   Policy
		Evidence string
	}{p, r.EvidenceDigest})
	return out, nil
}

func digest(value any) string {
	h := sha256.Sum256([]byte(encoded(value)))
	return hex.EncodeToString(h[:])
}

func has(values []string, s string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}

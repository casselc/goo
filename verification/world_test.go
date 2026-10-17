package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/casselc/goo/lifecycle"
	"github.com/casselc/observability/quintgo/qobs"
)

type projection struct {
	eligible, good, bad int
	unknown             bool
}
type archive struct {
	value   *lifecycle.Prepared
	encoded string
}
type world struct {
	accepted, terminal bool
	basis              int
	events             []lifecycle.Event
	snapshots          []archive
	recorder           *qobs.Recorder
	ctx                context.Context
}

func newWorld() *world { return &world{basis: 80, ctx: qobs.Thread(context.Background())} }

func policy(basis int) lifecycle.Policy {
	return lifecycle.Policy{
		APIVersion: "goo/lifecycle/v1alpha1", EvaluationID: "property-basis-" + strconv.Itoa(basis),
		DefinitionRevision: "test/v1", SourceRevision: "normalized/v1", EvidenceBasis: "ledger-" + strconv.Itoa(basis),
		CustodyClock: "fixture-durable-receipt", EvaluationTimeNS: "100", EvidenceCutoffNS: strconv.Itoa(basis),
		DeadlineNS: "10", AllowanceNS: "3", WindowNS: "30", SubjectField: "reservation_id",
		SuccessOutcome: "confirmed", TerminalOutcomes: []string{"confirmed", "rejected"},
		Scopes: []lifecycle.Scope{{Environment: "test", TenantID: "a", ServiceName: "reservations", Coverage: "complete", EvidenceRef: "synthetic-ledger"}},
	}
}

func event(kind string) lifecycle.Event {
	e := lifecycle.Event{EventID: kind, Kind: kind, Environment: "test", TenantID: "a", ServiceName: "reservations", SubjectID: "one", Release: "v1", EventTimeNS: "60", ReceivedTimeNS: "60"}
	if kind == "terminal" {
		e.EventTimeNS = "70"
		e.ReceivedTimeNS = "95"
		e.Outcome = "confirmed"
	}
	return e
}

func (w *world) prepare() (*lifecycle.Prepared, error) { return policy(w.basis).Prepare(w.events) }

func (w *world) actual() (projection, error) {
	p, err := w.prepare()
	if err != nil {
		return projection{}, err
	}
	return projection{p.Receipt.Counts["eligible"], p.Receipt.Counts["good"], p.Receipt.Counts["bad"], p.Receipt.State == "unknown"}, nil
}

func (w *world) expected() projection {
	p := projection{}
	if w.accepted {
		p.eligible = 1
		p.bad = 1
	}
	if w.accepted && w.terminal && w.basis > 95 {
		p.good = 1
		p.bad = 0
	}
	p.unknown = !w.accepted && w.terminal && w.basis > 95
	return p
}

func (w *world) apply(action string) (bool, error) {
	switch action {
	case "accept":
		if w.accepted {
			return false, nil
		}
		w.events = append(w.events, event("accepted"))
		w.accepted = true
	case "terminal":
		if w.terminal {
			return false, nil
		}
		w.events = append(w.events, event("terminal"))
		w.terminal = true
	case "duplicate":
		if len(w.events) == 0 {
			return false, nil
		}
		w.events = append(w.events, w.events[len(w.events)-1])
	case "advance95", "advance96", "advance100":
		n, _ := strconv.Atoi(action[len("advance"):])
		if n <= w.basis {
			return false, nil
		}
		w.basis = n
	case "snapshot":
		if len(w.snapshots) >= 3 {
			return false, nil
		}
		p, err := w.prepare()
		if err != nil {
			return false, err
		}
		b, _ := json.Marshal(p)
		w.snapshots = append(w.snapshots, archive{p, string(b)})
	default:
		return false, fmt.Errorf("unknown action %s", action)
	}
	if w.recorder != nil {
		p, err := w.actual()
		if err != nil {
			return false, err
		}
		w.recorder.Record(w.ctx, action, nil, "obs.eligible", p.eligible, "obs.good", p.good, "obs.bad", p.bad, "obs.unknown", p.unknown, "obs.basis", w.basis, "obs.snapshots", len(w.snapshots))
	}
	return true, nil
}

func (w *world) check() error {
	actual, err := w.actual()
	if err != nil {
		return err
	}
	if actual != w.expected() {
		return fmt.Errorf("projection %+v, want %+v", actual, w.expected())
	}
	for _, a := range w.snapshots {
		b, _ := json.Marshal(a.value)
		if string(b) != a.encoded {
			return fmt.Errorf("archived evaluation was rewritten")
		}
	}
	return nil
}

func artifact(name string, data []byte) error {
	dir := os.Getenv("GOO_VERIFICATION_OUT")
	if dir == "" {
		dir = "../artifacts/verification"
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), data, 0644)
}

package verification

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"strconv"
	"strings"
	"testing"
	"time"
)

type tailMembershipMachine struct {
	now        time.Time
	store      *exploration.TailMembershipSnapshots
	connection string
	selector   exploration.TailSelector
	caps       []exploration.TailMembershipCapability
	keys       map[string][]string
}

func (m *tailMembershipMachine) RuleIssue(tc hegel.TestCase) {
	n := hegel.Draw(tc, hegel.Integers(1, 4))
	rows := make([]json.RawMessage, n)
	stamp := m.now.UnixMilli() - 1000
	for i := range rows {
		raw, _ := json.Marshal(map[string]any{"timestamp_ns": strconv.FormatInt(stamp*1000000, 10), "Body": "same", "exact": json.Number("18446744073709551615")})
		rows[i] = raw
	}
	tracker := exploration.TailTracker{}
	observed, e := tracker.Observe(rows)
	if e != nil {
		panic(e)
	}
	cap, e := m.store.Freeze(exploration.TailMembershipBatch{Connection: m.connection, Selector: m.selector, FromMS: m.now.UnixMilli() - 60000, ToMS: m.now.UnixMilli(), ObservedAt: m.now, Rows: observed, Limited: n%2 == 0})
	if e != nil {
		if !errors.Is(e, exploration.ErrTailMembershipCapacity) {
			panic(e)
		}
		return
	}
	m.caps = append(m.caps, cap)
	for _, row := range observed {
		m.keys[cap.Token] = append(m.keys[cap.Token], row.Key)
	}
}
func (m *tailMembershipMachine) RuleAdvance(tc hegel.TestCase) {
	m.now = m.now.Add(time.Duration(hegel.Draw(tc, hegel.Integers(0, 180))) * time.Second)
}
func (m *tailMembershipMachine) RuleResolve(tc hegel.TestCase) {
	if len(m.caps) == 0 {
		return
	}
	cap := m.caps[hegel.Draw(tc, hegel.Integers(0, len(m.caps)-1))]
	keys := m.keys[cap.Token]
	key := keys[hegel.Draw(tc, hegel.Integers(0, len(keys)-1))]
	refs := []exploration.TailMembershipRef{{Token: cap.Token, Key: key}}
	draft, e := m.store.Resolve(m.connection, m.selector, refs)
	if !m.now.Before(cap.ExpiresAt) {
		if draft != nil || !errors.Is(e, exploration.ErrTailMembershipUnavailable) {
			panic("expired capability returned evidence")
		}
		return
	}
	if e != nil || len(draft.Members) != 1 || draft.Members[0].Ref.Key != key || draft.Members[0].Basis != cap.Basis || !bytes.Contains(draft.Members[0].Record, []byte("18446744073709551615")) {
		panic("membership or exact payload lost")
	}
	again, e := m.store.Resolve(m.connection, m.selector, refs)
	if e != nil || again.Digest != draft.Digest {
		panic("unchanged resolution changed intent")
	}
	wrong := m.selector
	wrong.Environment = "other"
	if d, e := m.store.Resolve(m.connection, wrong, refs); d != nil || e == nil {
		panic("changed selector accepted")
	}
	if d, e := m.store.Resolve(strings.Repeat("b", 64), m.selector, refs); d != nil || e == nil {
		panic("different connection accepted")
	}
	if d, e := m.store.Resolve(m.connection, m.selector, append(refs, refs[0])); d != nil || e == nil {
		panic("duplicate occurrence accepted")
	}
}
func (m *tailMembershipMachine) InvariantMembershipTTL(tc hegel.TestCase) {
	for _, cap := range m.caps {
		draft, err := m.store.Resolve(m.connection, m.selector, []exploration.TailMembershipRef{{Token: cap.Token, Key: m.keys[cap.Token][0]}})
		if m.now.Before(cap.ExpiresAt) {
			if err != nil || draft == nil {
				panic("live membership disappeared")
			}
		} else if draft != nil || !errors.Is(err, exploration.ErrTailMembershipUnavailable) {
			panic("expired membership returned evidence")
		}
	}
}
func TestHegelTailMembershipStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &tailMembershipMachine{now: time.UnixMilli(200000), connection: strings.Repeat("a", 64), selector: exploration.TailSelector{Environment: "test"}, keys: map[string][]string{}}
		m.store = exploration.NewTailMembershipSnapshots(func() time.Time { return m.now })
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantMembershipTTL"))
	}, options(70)...)
}
func TestQuintObservedTailMembership(t *testing.T) {
	check := func(bad int) bool {
		now := time.UnixMilli(200000)
		s := exploration.NewTailMembershipSnapshots(func() time.Time { return now })
		selector := exploration.TailSelector{Environment: "test"}
		connection := strings.Repeat("a", 64)
		tracker := exploration.TailTracker{}
		row := json.RawMessage(`{"timestamp_ns":"150000000000","Body":"same","exact":18446744073709551615}`)
		rows, e := tracker.Observe([]json.RawMessage{row, row})
		if e != nil {
			t.Fatal(e)
		}
		cap, e := s.Freeze(exploration.TailMembershipBatch{Connection: connection, Selector: selector, FromMS: 100000, ToMS: 160000, ObservedAt: now, Rows: rows})
		if e != nil {
			t.Fatal(e)
		}
		refs := []exploration.TailMembershipRef{{Token: cap.Token, Key: rows[0].Key}, {Token: cap.Token, Key: rows[1].Key}}
		draft, e := s.Resolve(connection, selector, refs)
		if e != nil {
			t.Fatal(e)
		}
		now = cap.ExpiresAt
		expired, expiredErr := s.Resolve(connection, selector, refs)
		count, available, denied := len(draft.Members), now.Before(cap.ExpiresAt), expired == nil && errors.Is(expiredErr, exploration.ErrTailMembershipUnavailable)
		switch bad {
		case 0:
			count++
		case 1:
			available = true
		case 2:
			draft.Members[0].FromMS++
		case 3:
			denied = false
		}
		coherent := draft.Members[0].FromMS == 100000 && draft.Members[1].FromMS == 100000 && draft.Members[0].ToMS == 160000 && draft.Members[0].Basis == cap.Basis && draft.Members[0].Ref.Key != draft.Members[1].Ref.Key && bytes.Contains(draft.Members[0].Record, []byte("18446744073709551615"))
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(action string, n int, live, rejected bool) {
			rec.Record(context.Background(), action, nil, "obs.coherent", coherent, "obs.count", n, "obs.available", live, "obs.denied", rejected)
		}
		record("issue", 0, true, false)
		record("resolve_members", count, true, false)
		record("expire", 0, available, false)
		record("refuse", 0, false, denied)
		c, e := validate.NewChecker("model/tail_membership.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := c.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("valid membership trace rejected")
	}
	for bad := 0; bad < 4; bad++ {
		if check(bad) {
			t.Fatal("forged membership trace accepted", bad)
		}
	}
}

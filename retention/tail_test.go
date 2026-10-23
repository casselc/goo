package retention

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
)

func tailFixture(t *testing.T) (*exploration.TailMembershipSnapshots, *time.Time, TailIntent) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Millisecond)
	bank := exploration.NewTailMembershipSnapshots(func() time.Time { return now })
	raw := json.RawMessage(`{"timestamp_ns":"` + strconv.FormatInt(now.Add(-time.Second).UnixNano(), 10) + `","n":18446744073709551615,"Body":"<check> café 🤝"}`)
	tracker := exploration.TailTracker{}
	rows, e := tracker.Observe([]json.RawMessage{raw, raw})
	if e != nil {
		t.Fatal(e)
	}
	in := TailIntent{Version: TailEvidenceVersion, RequestID: hash([]byte(t.Name() + now.String())), Connection: strings.Repeat("c", 64), Selector: exploration.TailSelector{Environment: "prod"}, Tags: []string{"review:tail"}}
	cap, e := bank.Freeze(exploration.TailMembershipBatch{Connection: in.Connection, Selector: in.Selector, FromMS: now.UnixMilli() - 60000, ToMS: now.UnixMilli(), ObservedAt: now, Rows: rows, Limited: true})
	if e != nil {
		t.Fatal(e)
	}
	for _, row := range rows {
		in.Refs = append(in.Refs, exploration.TailMembershipRef{Token: cap.Token, Key: row.Key})
	}
	return bank, &now, in
}
func TestTailManifestAndIntent(t *testing.T) {
	bank, _, in := tailFixture(t)
	normalized, digest, e := normalizeTailIntent(in)
	if e != nil {
		t.Fatal(e)
	}
	in.Refs[0], in.Refs[1] = in.Refs[1], in.Refs[0]
	in.Selector.Signal = "logs"
	_, other, e := normalizeTailIntent(in)
	if e != nil || digest != other {
		t.Fatal("order/default changed intent", e)
	}
	if normalized.Refs[0] == in.Refs[0] {
		t.Fatal("normalization aliased references")
	}
	d, e := bank.Resolve(in.Connection, in.Selector, in.Refs)
	if e != nil {
		t.Fatal(e)
	}
	m := TailManifest{Version: TailEvidenceVersion, Selection: *d}
	raw, _ := json.Marshal(m)
	m.ID = hash(raw)
	if e = m.Validate(); e != nil {
		t.Fatal(e)
	}
	// Rehashing an invalid payload cannot turn it into admissible evidence.
	m.Selection.Members[0].Record = json.RawMessage(`{"timestamp_ns":"0"}`)
	m.Selection.Digest = ""
	raw, _ = json.Marshal(m.Selection)
	m.Selection.Digest = hash(raw)
	m.ID = ""
	raw, _ = json.Marshal(m)
	m.ID = hash(raw)
	if m.Validate() == nil {
		t.Fatal("accepted fabricated member under a valid outer digest")
	}
}
func TestTailStoreOutagePrecedesMembership(t *testing.T) {
	_, _, in := tailFixture(t)
	_, e := (Store{URL: "postgres://127.0.0.1:1/xtdb?connect_timeout=1"}).PublishTail(context.Background(), nil, in)
	if e == nil || errors.Is(e, exploration.ErrTailMembershipUnavailable) {
		t.Fatal("outage was reported as absent proof", e)
	}
}
func TestXTDBTailPublication(t *testing.T) {
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		t.Skip("optional XTDB instance required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	bank, now, in := tailFixture(t)
	s := Store{URL: url}
	interrupted := s
	interrupted.beforeCommit = func() error { return errors.New("interrupted") }
	if _, e := interrupted.PublishTail(ctx, bank, in); e == nil {
		t.Fatal("interruption succeeded")
	}
	draft, e := bank.Resolve(in.Connection, in.Selector, in.Refs)
	if e != nil {
		t.Fatal(e)
	}
	expected := TailManifest{Version: TailEvidenceVersion, Selection: *draft}
	raw, _ := json.Marshal(expected)
	expected.ID = hash(raw)
	if _, e = s.LoadTail(ctx, expected.ID); e == nil {
		t.Fatal("interruption left partial evidence")
	}
	lost := s
	lost.afterTailCommit = func() error { return errors.New("acknowledgement lost") }
	if _, e = lost.PublishTail(ctx, bank, in); e == nil {
		t.Fatal("expected lost acknowledgement")
	}
	r, e := s.PublishTail(ctx, nil, in)
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Manifest.Selection.Members) != 2 || !strings.Contains(string(r.Manifest.Selection.Members[0].Record), "18446744073709551615") || !r.Manifest.Selection.Members[0].Limited || r.RetainedAt.IsZero() {
		t.Fatal("lost provenance")
	}
	// Discard the acknowledgement and all process-local proof, then retry.
	*now = now.Add(3 * time.Minute)
	recovered, e := s.PublishTail(ctx, nil, in)
	if e != nil {
		t.Fatal(e)
	}
	r.AwaitToken = ""
	recovered.AwaitToken = ""
	if !reflect.DeepEqual(r, recovered) {
		t.Fatal("retry changed durable receipt")
	}
	loaded, e := s.LoadTail(ctx, r.Manifest.ID)
	if e != nil || !reflect.DeepEqual(*loaded, r.Manifest) {
		t.Fatal("origin resolution", e)
	}
	if _, e = s.Load(ctx, r.Manifest.ID); e == nil {
		t.Fatal("tail entered v1 reader")
	}
	for _, mutate := range []func(*TailIntent){
		func(v *TailIntent) { v.Note = "changed" },
		func(v *TailIntent) { v.Tags = []string{"review:other"} },
		func(v *TailIntent) { v.Connection = strings.Repeat("d", 64) },
		func(v *TailIntent) { v.Selector.Environment = "staging" },
		func(v *TailIntent) { v.Refs = v.Refs[:1] },
	} {
		changed := in
		mutate(&changed)
		if _, e = s.PublishTail(ctx, nil, changed); !errors.Is(e, ErrTailIntentConflict) {
			t.Fatal("changed intent", e)
		}
	}
	absent := in
	absent.RequestID = hash([]byte(in.RequestID + "absent"))
	if _, e = s.PublishTail(ctx, bank, absent); !errors.Is(e, exploration.ErrTailMembershipUnavailable) {
		t.Fatal("expired proof", e)
	}
	// A second fresh intent races before either response has been received.
	fresh, _, intent := tailFixture(t)
	intent.RequestID = hash([]byte(in.RequestID + "race"))
	var wg sync.WaitGroup
	results := make(chan *TailReceipt, 2)
	errs := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); r, e := s.PublishTail(ctx, fresh, intent); results <- r; errs <- e }()
	}
	wg.Wait()
	close(results)
	close(errs)
	for e := range errs {
		if e != nil {
			t.Fatal("concurrent publication", e)
		}
	}
	var id string
	for r := range results {
		if id != "" && id != r.Annotation.ID {
			t.Fatal("race created distinct receipts")
		}
		id = r.Annotation.ID
	}
}

func TestTailManifestMixedPollProvenance(t *testing.T) {
	bank, now, in := tailFixture(t)
	first, e := bank.Resolve(in.Connection, in.Selector, in.Refs[:1])
	if e != nil {
		t.Fatal(e)
	}
	row := exploration.TailRow{Key: in.Refs[1].Key, Record: first.Members[0].Record, Late: true}
	cap, e := bank.Freeze(exploration.TailMembershipBatch{Connection: in.Connection, Selector: in.Selector, FromMS: now.UnixMilli() - 60500, ToMS: now.UnixMilli() - 500, ObservedAt: now.Add(-100 * time.Millisecond), Rows: []exploration.TailRow{row}})
	if e != nil {
		t.Fatal(e)
	}
	refs := []exploration.TailMembershipRef{in.Refs[0], {Token: cap.Token, Key: row.Key}}
	draft, e := bank.Resolve(in.Connection, in.Selector, refs)
	if e != nil {
		t.Fatal(e)
	}
	m := TailManifest{Version: TailEvidenceVersion, Selection: *draft}
	raw, _ := json.Marshal(m)
	m.ID = hash(raw)
	if e = m.Validate(); e != nil {
		t.Fatal(e)
	}
	var late, limited bool
	for _, v := range m.Selection.Members {
		if v.Ref.Token == cap.Token {
			late = v.Late && !v.Limited && v.ToMS == now.UnixMilli()-500 && v.ObservedAt.Equal(now.Add(-100*time.Millisecond)) && v.Basis == cap.Basis
		} else {
			limited = v.Limited && !v.Late && v.ToMS == now.UnixMilli() && v.ObservedAt.Equal(*now)
		}
	}
	if !late || !limited {
		t.Fatal("mixed-poll provenance collapsed")
	}
}

func TestTailManifestWhitespaceTransport(t *testing.T) {
	bank, _, in := tailFixture(t)
	draft, e := bank.Resolve(in.Connection, in.Selector, in.Refs)
	if e != nil {
		t.Fatal(e)
	}
	m := TailManifest{Version: TailEvidenceVersion, Selection: *draft}
	raw, _ := json.Marshal(m)
	m.ID = hash(raw)
	pretty, _ := json.MarshalIndent(m, "", "  ")
	var proxy map[string]any
	decoder := json.NewDecoder(bytes.NewReader(pretty))
	decoder.UseNumber()
	if e = decoder.Decode(&proxy); e != nil {
		t.Fatal(e)
	}
	var encoded bytes.Buffer
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if e = encoder.Encode(proxy); e != nil {
		t.Fatal(e)
	}
	pretty = bytes.ReplaceAll(encoded.Bytes(), []byte("café 🤝"), []byte(`caf\u00e9 \ud83e\udd1d`))

	var transported TailManifest
	if json.Unmarshal(pretty, &transported) != nil {
		t.Fatal("invalid fixture transport")
	}
	var fields map[string]json.RawMessage
	if e = json.Unmarshal(transported.Selection.Members[0].Record, &fields); e != nil {
		t.Fatal(e)
	}
	transported.Selection.Members[0].Record = json.RawMessage(`{"timestamp_ns":` + string(fields["timestamp_ns"]) + `,"n":` + string(fields["n"]) + `,"Body":` + string(fields["Body"]) + `}`)
	before := append([]byte(nil), transported.Selection.Members[0].Record...)
	if e = transported.Validate(); e != nil {
		t.Fatal("JSON transport formatting invalidated canonical proof", e)
	}
	if !bytes.Equal(before, transported.Selection.Members[0].Record) {
		t.Fatal("validator rewrote caller-owned transport bytes")
	}

	if transported.ID != m.ID || !strings.Contains(string(transported.Selection.Members[0].Record), "18446744073709551615") {
		t.Fatal("transport changed identity or numeric tokens")
	}
}

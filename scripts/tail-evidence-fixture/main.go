// Development journey fixture: poll actual OTLP/ClickStack data through the
// trusted Go source adapter and publish two exact observed occurrences. This is
// not a public endpoint and never accepts client row payloads.
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
)

func value(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func randomID() string {
	var v [32]byte
	if _, e := rand.Read(v[:]); e != nil {
		panic(e)
	}
	return hex.EncodeToString(v[:])
}
func run() error {
	if len(os.Args) != 3 {
		return errors.New("usage: tail-evidence-fixture ENVIRONMENT OUTPUT_JSON")
	}
	url := os.Getenv("GOO_XTDB_TEST_URL")
	if url == "" {
		return errors.New("GOO_XTDB_TEST_URL required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	source := exploration.Source{Endpoint: value("GOO_TEST_CLICKHOUSE_URL", "http://127.0.0.1:8123"), User: value("GOO_TEST_CLICKHOUSE_USER", "lab"), Password: value("GOO_TEST_CLICKHOUSE_PASSWORD", "lab-local")}
	selector := exploration.TailSelector{Signal: "spans", Environment: os.Args[1]}
	now := time.Now().UTC()
	to := now.UnixMilli()
	records, limited, e := source.PollTail(ctx, selector, to)
	if e != nil {
		return e
	}
	if len(records) != 2 {
		return fmt.Errorf("fixture requires two actual source occurrences, got %d", len(records))
	}
	tracker := exploration.TailTracker{}
	rows, e := tracker.Observe(records)
	if e != nil {
		return e
	}
	bank := exploration.NewTailMembershipSnapshots(nil)
	connection := randomID()
	// Freeze different original poll windows/read clocks for the two members. The
	// second window is itself read through the source adapter before membership.
	refs := []exploration.TailMembershipRef{}
	for i, row := range rows {
		windowTo := to - int64(i)
		polled, bounded, e := source.PollTail(ctx, selector, windowTo)
		if e != nil {
			return e
		}
		candidate, e := (&exploration.TailTracker{}).Observe(polled)
		if e != nil {
			return e
		}
		found := false
		for _, v := range candidate {
			if v.Key == row.Key {
				row = v
				found = true
				break
			}
		}
		if !found {
			return errors.New("selected occurrence absent from its actual poll")
		}
		observed := time.Now().UTC()
		cap, e := bank.Freeze(exploration.TailMembershipBatch{Connection: connection, Selector: selector, FromMS: windowTo - exploration.TailWindowMS, ToMS: windowTo, ObservedAt: observed, Limited: limited || bounded, Rows: []exploration.TailRow{row}})
		if e != nil {
			return e
		}
		refs = append(refs, exploration.TailMembershipRef{Token: cap.Token, Key: row.Key})
	}
	in := retention.TailIntent{Version: retention.TailEvidenceVersion, RequestID: randomID(), Connection: connection, Selector: selector, Refs: refs, Tags: []string{"review:tail-ui"}, Note: "Observed retries during a service rollout"}
	receipt, e := (retention.Store{URL: url}).PublishTail(ctx, bank, in)
	if e != nil {
		return e
	}
	// Output is a fixture receipt, never credentials. Keep exact JSON numeric tokens.
	raw, e := json.MarshalIndent(receipt, "", "  ")
	if e != nil {
		return e
	}
	if len(raw) > retention.MaxBytes+16384 {
		return errors.New("fixture receipt exceeds bound")
	}
	return os.WriteFile(os.Args[2], raw, 0600)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

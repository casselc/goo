package verification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

// This is readiness for immutable occurrence references, not a review/task
// producer conformance claim. Expected membership is simply rawRows[index].
func TestHegelCapturedRecord365(t *testing.T) {
	fixture, err := os.ReadFile("../testdata/storage/exploration-capture.json")
	if err != nil {
		t.Fatal(err)
	}
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(101, 240))
		index := hegel.Draw(tc, hegel.Integers(100, n-1))
		payload := hegel.Draw(tc, hegel.Integers(0, 1000000))
		for _, signal := range []string{"logs", "spans"} {
			var c exploration.SavedCapture
			if err := json.Unmarshal(fixture, &c); err != nil {
				tc.Fatal(err)
			}
			base := bytes.Clone(c.Explore.Rows[0])
			rows := make([]json.RawMessage, n)
			for i := range rows {
				var fields map[string]json.RawMessage
				if err := json.Unmarshal(base, &fields); err != nil {
					tc.Fatal(err)
				}
				fields["occurrenceProbe"] = json.RawMessage(strconv.Itoa(i))
				fields["largeIntegerProbe"] = json.RawMessage(fmt.Sprintf("900719925475%04d", payload%10000))
				raw, err := json.Marshal(fields)
				if err != nil {
					tc.Fatal(err)
				}
				rows[i] = raw
			}
			rows[100] = bytes.Clone(rows[0]) // Equal payloads are different occurrences.
			c.Explore.Rows = rows
			c.Explore.Count = strconv.Itoa(n + 17)
			c.Explore.HasMore = true
			c.Explore.TraceExamples = nil
			c.Explore.Groups = nil
			c.Explore.Question.Signal = signal
			c.Question.Explore.Signal = signal
			c.ID = ""
			digestWire, err := json.Marshal(c)
			if err != nil {
				tc.Fatal(err)
			}
			hash := sha256.Sum256(digestWire)
			c.ID = hex.EncodeToString(hash[:])
			wire, err := json.Marshal(c)
			if err != nil {
				tc.Fatal(err)
			}
			dir := t.TempDir()
			store := exploration.CaptureStore{Dir: dir}
			file := filepath.Join(dir, c.ID+".json")
			if err = os.WriteFile(file, wire, 0600); err != nil {
				tc.Fatal(err)
			}
			loaded, err := store.Load(c.ID)
			if err != nil {
				tc.Fatal("fixture must pass public immutable load", err)
			}
			first, err := store.ResolveRecord(exploration.CaptureRecordRef{CaptureID: c.ID, Index: 0})
			if err != nil {
				tc.Fatal(err)
			}
			duplicate, err := store.ResolveRecord(exploration.CaptureRecordRef{CaptureID: c.ID, Index: 100})
			if err != nil {
				tc.Fatal(err)
			}
			if !bytes.Equal(first.Record, duplicate.Record) || first.Reference == duplicate.Reference || first.RecordDigest != duplicate.RecordDigest {
				tc.Fatal("equal payloads collapsed occurrence identity")
			}
			for _, at := range []int{0, 100, index, n - 1} {
				ref := exploration.CaptureRecordRef{CaptureID: c.ID, Index: at}
				got, err := store.ResolveRecord(ref)
				if err != nil {
					tc.Fatal(err)
				}
				expected := rows[at]
				sum := sha256.Sum256(expected)
				if got.Reference != ref || !bytes.Equal(got.Record, expected) || got.RecordDigest != hex.EncodeToString(sum[:]) || !reflect.DeepEqual(got.Receipt, loaded.CaptureReceipt) || got.SourceReadAt != c.Explore.CapturedAt || got.Basis != c.Explore.Basis || got.ObservationState != c.Explore.State || got.MatchingRecords != c.Explore.Count || got.RetainedRecords != n || !got.RecordsLimited {
					tc.Fatal("absolute raw occurrence or frozen evidence metadata changed")
				}
				got.Record[0] = '['
				got.Receipt.Question.Explore.Environment = "caller mutation"
				loaded.Explore.Rows[at][0] = '['
				again, err := store.ResolveRecord(ref)
				if err != nil || !bytes.Equal(again.Record, expected) || again.Receipt.Question.Explore.Environment != c.Question.Explore.Environment {
					tc.Fatal("caller-owned result mutated stored evidence", err)
				}
			}
			for _, at := range []int{-1, n, exploration.CaptureRowLimit, exploration.CaptureRowLimit + 1} {
				if _, err := store.ResolveRecord(exploration.CaptureRecordRef{CaptureID: c.ID, Index: at}); err == nil {
					tc.Fatal("out of retained bounds admitted", at)
				}
			}
			for _, id := range []string{"../escape", "", string(bytes.Repeat([]byte("f"), 64))} {
				if _, err := store.ResolveRecord(exploration.CaptureRecordRef{CaptureID: id, Index: 0}); err == nil {
					tc.Fatal("invalid/missing capture admitted")
				}
			}
			before, err := os.ReadFile(file)
			if err != nil || !bytes.Equal(before, wire) {
				tc.Fatal("read/mutation changed immutable bytes", err)
			}
			corrupt := bytes.Replace(before, []byte("18446744073709551615"), []byte("18446744073709551614"), 1)
			if bytes.Equal(corrupt, before) {
				tc.Fatal("fixture lacks exact uint64 token")
			}
			if err = os.WriteFile(file, corrupt, 0600); err != nil {
				tc.Fatal(err)
			}
			if _, err := store.ResolveRecord(exploration.CaptureRecordRef{CaptureID: c.ID, Index: index}); err == nil {
				tc.Fatal("corrupt content accepted despite capture digest")
			}
		}
	}, options(40)...)
}

package exploration

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestResolveCapturedRecordAbsoluteOccurrence365(t *testing.T) {
	wire, err := os.ReadFile("../testdata/storage/exploration-capture.json")
	if err != nil {
		t.Fatal(err)
	}
	var capture SavedCapture
	if err = json.Unmarshal(wire, &capture); err != nil {
		t.Fatal(err)
	}
	duplicate := bytes.Clone(capture.Explore.Rows[0])
	capture.Explore.Rows = make([]json.RawMessage, 150)
	for i := range capture.Explore.Rows {
		capture.Explore.Rows[i] = bytes.Clone(duplicate)
	}
	capture.Explore.Count = "900"
	capture.Explore.HasMore = true
	capture.Explore.TraceExamples = nil
	capture.Explore.Groups = nil
	store := CaptureStore{Dir: t.TempDir()}
	for _, signal := range []string{"logs", "spans"} {
		capture.Explore.Question.Signal = signal
		capture.Question.Explore.Signal = signal
		saved, err := store.publish(capture)
		if err != nil {
			t.Fatal(err)
		}
		first, err := store.ResolveRecord(CaptureRecordRef{CaptureID: saved.ID, Index: 0})
		if err != nil {
			t.Fatal(err)
		}
		later, err := store.ResolveRecord(CaptureRecordRef{CaptureID: saved.ID, Index: 100})
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(first.Record, later.Record) || first.Reference == later.Reference || first.RecordDigest != later.RecordDigest {
			t.Fatal("duplicate values became the same occurrence")
		}
		page0, err := saved.Page("")
		if err != nil {
			t.Fatal(err)
		}
		page1, err := saved.Page(page0.NextCursor)
		if err != nil || page1.Explore.Question.Offset != 100 || !bytes.Equal(page1.Explore.Rows[0], later.Record) {
			t.Fatal("page-local zero did not map to absolute 100", err)
		}
		sum := sha256.Sum256(later.Record)
		if later.RecordDigest != hex.EncodeToString(sum[:]) || !strings.Contains(string(later.Record), "18446744073709551615") || !strings.Contains(string(later.Record), "1700000000000000123") {
			t.Fatal("stored number tokens changed")
		}
		if later.MatchingRecords != "900" || later.RetainedRecords != 150 || !later.RecordsLimited || later.SourceReadAt != saved.Explore.CapturedAt || !reflect.DeepEqual(later.Receipt, saved.CaptureReceipt) {
			t.Fatal("custody or extent lost")
		}
		later.Record[0] = '['
		later.Receipt.Question.Explore.Environment = "changed"
		again, err := store.ResolveRecord(CaptureRecordRef{CaptureID: saved.ID, Index: 100})
		if err != nil || !bytes.Equal(again.Record, duplicate) || again.Receipt.Question.Explore.Environment != "local" {
			t.Fatal("caller mutation changed stored evidence", err)
		}
		for _, index := range []int{-1, 150, CaptureRowLimit} {
			if _, err := store.ResolveRecord(CaptureRecordRef{CaptureID: saved.ID, Index: index}); err == nil {
				t.Fatal("invalid index admitted", index)
			}
		}
		file := filepath.Join(store.Dir, saved.ID+".json")
		original, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		corrupt := bytes.Replace(original, []byte("18446744073709551615"), []byte("18446744073709551614"), 1)
		if err := os.WriteFile(file, corrupt, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := store.ResolveRecord(CaptureRecordRef{CaptureID: saved.ID, Index: 100}); err == nil {
			t.Fatal("corrupt capture admitted")
		}
	}
	q := InventoryQuestion{Environment: "local", FromMS: 0, ToMS: 60000}
	inventory, err := store.publish(SavedCapture{CaptureReceipt: CaptureReceipt{Version: 1, Title: "Inventory is not record evidence", CapturedAt: "2026-11-07T00:00:00Z", Question: CaptureQuestion{Kind: "inventory", Inventory: &q}}, Inventory: &InventoryResult{Question: q, Basis: strings.Repeat("a", 64), Count: "0"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.ResolveRecord(CaptureRecordRef{CaptureID: inventory.ID}); err == nil || !strings.Contains(err.Error(), "ordinary logs or spans") {
		t.Fatal("unsupported capture family not refused", err)
	}
	for _, id := range []string{"../escape", strings.Repeat("a", 64)} {
		if _, err := store.ResolveRecord(CaptureRecordRef{CaptureID: id}); err == nil {
			t.Fatal("invalid or missing capture admitted")
		}
	}
}

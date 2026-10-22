package measurements

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/casselc/goo/lifecycle"
)

func custodyStorageFixture(t *testing.T) (lifecycle.Event, []byte, string) {
	t.Helper()
	wire, e := os.ReadFile("../testdata/storage/lifecycle-custody.json")
	if e != nil {
		t.Fatal(e)
	}
	var event lifecycle.Event
	if e = json.Unmarshal(wire, &event); e != nil {
		t.Fatal(e)
	}
	var ids map[string]string
	b, e := os.ReadFile("../testdata/storage/identities.json")
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &ids); e != nil {
		t.Fatal(e)
	}
	copy := event
	copy.ReceivedTimeNS = ""
	id := Hash(copy)
	if id != ids["lifecycleCustodyID"] {
		t.Fatal("custody identity changed", id)
	}
	return event, wire, id
}
func TestCustodyStorageContract(t *testing.T) {
	event, wire, id := custodyStorageFixture(t)
	dir := t.TempDir()
	ns, e := strconv.ParseInt(event.ReceivedTimeNS, 10, 64)
	if e != nil {
		t.Fatal(e)
	}
	input := event
	input.ReceivedTimeNS = "caller-spoof"
	got, e := observe(dir, input, time.Unix(0, ns))
	if e != nil || got != event {
		t.Fatal("first observation", got, e)
	}
	file := filepath.Join(dir, id+".json")
	stored, e := os.ReadFile(file)
	if e != nil || !bytes.Equal(stored, wire) {
		t.Fatal("custody bytes changed", e)
	}
	// A later call carrying an earlier observation time must keep the first winner.
	got, e = observe(dir, input, time.Unix(0, ns-1000))
	if e != nil || got != event {
		t.Fatal("custody clock rewritten", got, e)
	}
	input.EventTimeNS = "1700000000123456790"
	other, e := observe(dir, input, time.Unix(0, ns+1000))
	if e != nil || other.EventTimeNS == event.EventTimeNS || other.ReceivedTimeNS == event.ReceivedTimeNS {
		t.Fatal("business payload identity collapsed", e)
	}
	for name, bad := range map[string][]byte{"truncated": wire[:len(wire)/2], "changed event": bytes.Replace(wire, []byte("job-1"), []byte("job-2"), 1), "invalid custody clock": bytes.Replace(wire, []byte(event.ReceivedTimeNS), []byte("not-an-integer"), 1)} {
		t.Run(name, func(t *testing.T) {
			if e := os.WriteFile(file, bad, 0600); e != nil {
				t.Fatal(e)
			}
			if _, e := observe(dir, event, time.Unix(0, ns)); e == nil {
				t.Fatal("bad custody accepted")
			}
			after, e := os.ReadFile(file)
			if e != nil || !bytes.Equal(after, bad) {
				t.Fatal("corrupt winner repaired implicitly", e)
			}
		})
	}
	// Received time is excluded from identity: a valid integer edit is not detected.
	changedClock := bytes.Replace(wire, []byte(event.ReceivedTimeNS), []byte("1792700000000000999"), 1)
	if e = os.WriteFile(file, changedClock, 0600); e != nil {
		t.Fatal(e)
	}
	got, e = observe(dir, event, time.Unix(0, ns))
	if e != nil || got.ReceivedTimeNS != "1792700000000000999" {
		t.Fatal("existing clock-validation policy changed", e)
	}
	if _, e = observe("", event, time.Unix(0, ns)); e == nil {
		t.Fatal("ephemeral custody accepted")
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 2 {
		t.Fatal("scratch leaked", e)
	}
}
func TestCustodyDistinctObservationClocksHaveOneWinner(t *testing.T) {
	event, _, _ := custodyStorageFixture(t)
	dir := t.TempDir()
	start := make(chan struct{})
	results := make(chan lifecycle.Event, 16)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			got, e := observe(dir, event, time.Unix(0, 1792700000000000000+int64(i)))
			if e != nil {
				t.Error(e)
				return
			}
			results <- got
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	winner := ""
	count := 0
	for got := range results {
		count++
		if winner == "" {
			winner = got.ReceivedTimeNS
		}
		n, e := strconv.ParseInt(got.ReceivedTimeNS, 10, 64)
		if e != nil || n < 1792700000000000000 || n > 1792700000000000015 || got.ReceivedTimeNS != winner || got.EventTimeNS != event.EventTimeNS {
			t.Fatal("mixed custody winner", got, e)
		}
	}
	if count != 16 {
		t.Fatal("missing observations", count)
	}
	entries, e := os.ReadDir(dir)
	if e != nil || len(entries) != 1 {
		t.Fatal("scratch leaked", e)
	}
}

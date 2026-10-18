package measurements

import (
	"encoding/json"
	"github.com/casselc/goo/lifecycle"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func definition(t *testing.T, name string) Definition {
	t.Helper()
	b, e := os.ReadFile("../config/measurements/" + name + ".yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := Read(string(b))
	if e != nil {
		t.Fatal(e)
	}
	return d
}
func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func gaugeRow(d Definition, instance string, at int64, value string) metricRow {
	return metricRow{Resource: map[string]string{"service.name": d.Source.Services[0], "service.instance.id": instance, "deployment.environment.name": d.Environment}, Scope: d.Source.Scope, Version: d.Source.Version, SchemaURL: d.Source.SchemaURL, Name: d.Source.Metric, Unit: d.Source.Unit, Attributes: d.Source.Attributes, EndNS: ns(at), Value: value}
}
func TestBuiltinsAndParameterizedPlans(t *testing.T) {
	files, _ := filepath.Glob("../config/measurements/*.yaml")
	for _, f := range files {
		b, _ := os.ReadFile(f)
		d, e := Read(string(b))
		if e != nil {
			t.Fatal(f, e)
		}
		q := Question{d, 100000, 220000}
		p, e := Compile(q)
		if e != nil {
			t.Fatal(e)
		}
		if strings.Contains(p.SQL, d.Environment) || !strings.Contains(p.SQL, "LIMIT 10001") {
			t.Fatal("unbounded or literal plan")
		}
		d.Environment = "' OR 1=1"
		q.Definition = d
		p, e = Compile(q)
		if e != nil || strings.Contains(p.SQL, d.Environment) {
			t.Fatal("SQL parameter escaped binding")
		}
	}
	b, _ := os.ReadFile("../config/measurements/queue-lag.yaml")
	for _, suffix := range []string{"\nunknown: true", "\n---\na: b", "\ntitle: duplicate"} {
		if _, e := Read(string(b) + suffix); e == nil {
			t.Fatal("invalid definition accepted")
		}
	}
	f, e := os.Open("../config/workflows.json")
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if _, e = ReadWorkflows(f); e != nil {
		t.Fatal(e)
	}
}
func TestGaugeReplayAndIsolation(t *testing.T) {
	d := definition(t, "queue-lag")
	q := Question{d, 100000, 220000}
	row := gaugeRow(d, "worker-a", 210000, "7")
	r, e := Evaluate(q, []json.RawMessage{raw(row)}, time.UnixMilli(230000), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	if e = ValidateResult(r); e != nil {
		t.Fatal(e)
	}
	d.Source.Attributes["injected"] = "changed"
	if r.Question.Definition.Source.Attributes["injected"] != "" {
		t.Fatal("question mutated")
	}
	r.Gauges.Series[0].Samples[0].Value = new(float64)
	if ValidateResult(r) == nil {
		t.Fatal("tampered result accepted")
	}
	row.Unit = "wrong"
	if _, e = Evaluate(q, []json.RawMessage{raw(row)}, time.UnixMilli(230000), "", nil); e == nil {
		t.Fatal("unit mismatch accepted")
	}
}
func TestLifecycleCustodyCorrectionAndReplay(t *testing.T) {
	d := definition(t, "reservation-lifecycle")
	q := Question{d, 100000, 220000}
	dir := t.TempDir()
	event := func(id, body, at, outcome string) json.RawMessage {
		return raw(map[string]any{"body": body, "service": "lab-grpc", "resource": map[string]string{"deployment.environment.name": "local"}, "attributes": map[string]string{"goo.event.id": id, "goo.event.time_ns": at, "goo.subject.id": "s1", "goo.subject.tenant": "demo", "app.release": "r1", "app.outcome": outcome}})
	}
	rows := []json.RawMessage{event("e1", d.Lifecycle.Accepted, ns(150000), "")}
	r, e := Evaluate(q, rows, time.UnixMilli(230000), dir, nil)
	if e != nil {
		t.Fatal(e)
	}
	if e = ValidateResult(r); e != nil {
		t.Fatal(e)
	}
	if r.Lifecycle.Receipt.Counts["unknown"] != 1 || r.Lifecycle.Receipt.Counts["bad"] != 0 {
		t.Fatal("missing delivery classified bad", r.Lifecycle.Receipt.Counts)
	}
	rows = append(rows, event("e2", d.Lifecycle.Terminal, ns(155000), "confirmed"))
	next, e := Evaluate(q, rows, time.UnixMilli(240000), dir, r)
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(next.Question, r.Question) || next.Events[0].ReceivedTimeNS != r.Events[0].ReceivedTimeNS || next.Lifecycle.Receipt.Counts["good"] != 1 {
		t.Fatal("correction changed question or custody")
	}
	if e = ValidateResult(next); e != nil {
		t.Fatal(e)
	}
	forgedPolicy := next.Lifecycle.Receipt.Policy
	forgedPolicy.Scopes[0].Coverage = "complete"
	forged, e := forgedPolicy.Prepare(next.Events)
	if e != nil {
		t.Fatal(e)
	}
	next.Lifecycle = forged
	next.State = forged.Receipt.State
	if ValidateResult(next) == nil {
		t.Fatal("complete coverage forgery accepted")
	}
}
func TestCounterPrecisionAndBoundedInputs(t *testing.T) {
	d := definition(t, "checkout-calls")
	q := Question{d, 100000, 220000}
	row := gaugeRow(d, "a", 210000, "9007199254740992")
	row.StartNS = ns(90000)
	row.Temporality = 2
	row.Monotonic = true
	if _, e := Evaluate(q, []json.RawMessage{raw(row)}, time.Now(), "", nil); e == nil {
		t.Fatal("unsafe Float64 integer accepted")
	}
	if _, e := Evaluate(q, make([]json.RawMessage, RowLimit+1), time.Now(), "", nil); e == nil {
		t.Fatal("partial result accepted")
	}
}

func TestEvidenceSurvivesJSONTransportSpellings(t *testing.T) {
	d := definition(t, "worker-active")
	q := Question{d, 100000, 220000}
	row := gaugeRow(d, "worker", 210000, "1")
	data, _ := json.Marshal(row)
	// ClickHouse commonly escapes slashes; JSON transports are free to unescape.
	escaped := strings.ReplaceAll(string(data), "/", `\/`)
	r, e := Evaluate(q, []json.RawMessage{json.RawMessage(escaped)}, time.UnixMilli(230000), "", nil)
	if e != nil {
		t.Fatal(e)
	}
	wire, _ := json.Marshal(r)
	var transport any
	json.Unmarshal(wire, &transport)
	wire, _ = json.Marshal(transport)
	var reopened Result
	json.Unmarshal(wire, &reopened)
	if e = ValidateResult(&reopened); e != nil {
		t.Fatal("transport changed evidence identity", e)
	}
}

func TestCustodyConcurrentPublication(t *testing.T) {
	dir := t.TempDir()
	event := lifecycle.Event{EventID: "e1", Kind: "accepted", Environment: "local", TenantID: "demo", ServiceName: "jobs", SubjectID: "s1", Release: "r1", EventTimeNS: "100"}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			got, e := observe(dir, event, time.Unix(0, 200))
			if e != nil || got.ReceivedTimeNS != "200" {
				t.Errorf("atomic custody publication: %v %+v", e, got)
			}
		}()
	}
	close(start)
	wg.Wait()
	files, e := os.ReadDir(dir)
	if e != nil || len(files) != 1 {
		t.Fatal("partial custody files survived", e, len(files))
	}
}

func TestLifecycleRejectsUnusedScopeSelectors(t *testing.T) {
	d := definition(t, "reservation-lifecycle")
	for _, field := range []string{"scope", "version", "schema"} {
		copy := d
		switch field {
		case "scope":
			copy.Source.Scope = "ignored-filter"
		case "version":
			copy.Source.Version = "ignored-filter"
		case "schema":
			copy.Source.SchemaURL = "ignored-filter"
		}
		if copy.Validate() == nil {
			t.Fatal("ignored lifecycle selector accepted", field)
		}
	}
}

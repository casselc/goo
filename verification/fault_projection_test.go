package verification

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"testing"

	"hegel.dev/go/hegel"
)

type projectionOracle struct {
	mu  sync.Mutex
	in  io.Writer
	out *bufio.Scanner
}
type projectionSpan struct{ SpanName, StatusCode string }
type projectionLog struct {
	Body       string
	Attributes map[string]string
}

func (o *projectionOracle) read(spans []projectionSpan, logs []projectionLog) faultFacts {
	o.mu.Lock()
	defer o.mu.Unlock()
	data, err := json.Marshal(map[string]any{"spans": spans, "logs": logs})
	if err != nil {
		panic(err)
	}
	if _, err = fmt.Fprintf(o.in, "%s\n", data); err != nil {
		panic(err)
	}
	if !o.out.Scan() {
		panic("fault projection process did not return an observation")
	}
	var facts faultFacts
	if err = json.Unmarshal(o.out.Bytes(), &facts); err != nil {
		panic(err)
	}
	return facts
}

type faultProjectionMachine struct {
	oracle *projectionOracle
	spans  []projectionSpan
	logs   []projectionLog
}

func (m *faultProjectionMachine) RuleAttempt(tc hegel.TestCase) {
	status := hegel.Draw(tc, hegel.SampledFrom([]string{"Unset", "Error"}))
	m.spans = append(m.spans, projectionSpan{"inventory process", status})
}
func (m *faultProjectionMachine) RuleRetry(tc hegel.TestCase) {
	m.logs = append(m.logs, projectionLog{Body: "reservation.retry"})
}
func (m *faultProjectionMachine) RuleStall(tc hegel.TestCase) {
	m.logs = append(m.logs, projectionLog{Body: "reservation.stalled"})
}
func (m *faultProjectionMachine) RuleTerminal(tc hegel.TestCase) {
	outcome := hegel.Draw(tc, hegel.SampledFrom([]string{"confirmed", "rejected", ""}))
	m.logs = append(m.logs, projectionLog{Body: "reservation.terminal", Attributes: map[string]string{"app.outcome": outcome}})
}
func (m *faultProjectionMachine) RuleReset(tc hegel.TestCase) { m.spans = nil; m.logs = nil }
func (m *faultProjectionMachine) RuleDuplicateLog(tc hegel.TestCase) {
	if len(m.logs) == 0 {
		return
	}
	before := m.oracle.read(m.spans, m.logs)
	m.logs = append(m.logs, m.logs[hegel.Draw(tc, hegel.Integers(0, len(m.logs)-1))])
	if after := m.oracle.read(m.spans, m.logs); after != before {
		panic("duplicate log changed observed meaning")
	}
}
func (m *faultProjectionMachine) InvariantHonestyAndOrder(tc hegel.TestCase) {
	facts := m.oracle.read(m.spans, m.logs)
	terminals := map[string]bool{}
	for _, log := range m.logs {
		if log.Body == "reservation.terminal" {
			terminals[log.Attributes["app.outcome"]] = true
		}
	}
	if len(terminals) == 0 && (facts.Meaning == "rejected" || facts.Meaning == "confirmed-after-retry") {
		panic("invented terminal outcome")
	}
	if facts.Meaning == "healthy" {
		panic("capture absence became health")
	}
	if facts.TerminalSeen && facts.Meaning == "unknown-no-terminal" {
		panic("present but malformed terminal was treated as absent")
	}
	if len(terminals) > 1 && facts.Meaning != "unknown-conflicting-terminals" {
		panic("conflicting outcomes were resolved without evidence")
	}
	spans := append([]projectionSpan(nil), m.spans...)
	logs := append([]projectionLog(nil), m.logs...)
	for i, j := 0, len(spans)-1; i < j; i, j = i+1, j-1 {
		spans[i], spans[j] = spans[j], spans[i]
	}
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}
	if reordered := m.oracle.read(spans, logs); reordered != facts {
		panic("delivery order changed captured interpretation")
	}
}
func TestHegelFaultEvidenceProjectionStateSwarm(t *testing.T) {
	// Exercise the actual browser-test projection through one persistent Node
	// process. These are protocol fixtures, not simulated integration evidence.
	script := `const r=require('node:readline').createInterface({input:process.stdin});const f=require('../journeys/support/fault-evidence');r.on('line',line=>{const d=JSON.parse(line);console.log(JSON.stringify(f.summarizeCaptured(d.spans||[],d.logs||[])))})`
	cmd := exec.Command("node", "-e", script)
	in, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		in.Close()
		if err := cmd.Wait(); err != nil {
			t.Errorf("projection process: %v", err)
		}
	}()
	oracle := &projectionOracle{in: in, out: bufio.NewScanner(out)}
	hegel.Test(t, func(tc *hegel.T) {
		m := &faultProjectionMachine{oracle: oracle}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(20), hegel.WithAlwaysCheckInvariants("InvariantHonestyAndOrder"))
	}, options(50)...)
}

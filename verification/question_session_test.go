package verification

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"os/exec"
	"reflect"
	"testing"
)

type sessionCommand struct {
	Action   string             `json:"action"`
	Reset    bool               `json:"reset"`
	Token    int                `json:"token"`
	Question *exploration.Query `json:"question,omitempty"`
}
type sessionOracle struct {
	Generation int
	Phase      string
	Applied    *questionApplied
	Requested  *exploration.Query
	Bounds     *questionBounds
}
type sessionMachine struct {
	commands  []sessionCommand
	expected  []sessionOracle
	state     sessionOracle
	lastToken int
	questions map[int]*exploration.Query
}

func (m *sessionMachine) save(c sessionCommand) {
	m.commands = append(m.commands, c)
	b, _ := json.Marshal(m.state)
	var copy sessionOracle
	json.Unmarshal(b, &copy)
	m.expected = append(m.expected, copy)
}
func (m *sessionMachine) RuleStart(tc hegel.TestCase) {
	q := exploration.Query{Signal: []string{"spans", "logs"}[hegel.Draw(tc, hegel.Integers(0, 1))], Environment: "ordinary", FromMS: 1000, ToMS: 61000, Limit: 100, Offset: 0}
	m.state.Generation++
	m.lastToken = m.state.Generation
	m.state.Phase = "running"
	m.state.Requested = &q
	m.state.Bounds = &questionBounds{1000, 61000}
	m.questions[m.lastToken] = &q
	m.save(sessionCommand{Action: "start", Question: &q})
}
func (m *sessionMachine) RuleEdit(tc hegel.TestCase) {
	reset := hegel.Draw(tc, hegel.Integers(0, 1)) == 1
	m.state.Generation++
	m.state.Phase = "draft"
	m.state.Requested = nil
	if reset {
		m.state.Bounds = nil
	}
	m.save(sessionCommand{Action: "edit", Reset: reset})
}
func (m *sessionMachine) RuleCancel(tc hegel.TestCase) {
	m.state.Generation++
	m.state.Phase = "cancelled"
	m.state.Requested = nil
	m.save(sessionCommand{Action: "cancel"})
}
func (m *sessionMachine) RuleCompletion(tc hegel.TestCase) {
	token := m.lastToken
	if hegel.Draw(tc, hegel.Integers(0, 1)) == 0 {
		token = 0
	}
	action := []string{"publish", "fail"}[hegel.Draw(tc, hegel.Integers(0, 1))]
	if token == m.state.Generation && m.state.Phase == "running" {
		if action == "publish" {
			m.state.Phase = "applied"
			m.state.Applied = &questionApplied{Question: *m.questions[token], Basis: "basis", Count: "1", State: "observed"}
		} else {
			m.state.Phase = "failed"
		}
		m.state.Requested = nil
	}
	m.save(sessionCommand{Action: action, Token: token})
}
func TestHegelQuestionSessionStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &sessionMachine{state: sessionOracle{Phase: "initial"}, questions: map[int]*exploration.Query{}}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(25))
		data, _ := json.Marshal(m.commands)
		script := `const fs=require('node:fs'),{Session}=require('../ui/question-session.mjs');const s=new Session(),questions={},states=[];for(const c of JSON.parse(fs.readFileSync(0,'utf8'))){if(c.action==='start'){const token=s.start(c.question);questions[token]=c.question}else if(c.action==='edit')s.edit(c.reset);else if(c.action==='cancel')s.cancel();else if(c.action==='fail')s.fail(c.token);else s.publish(c.token,{question:questions[c.token]||{},basis:'basis',count:'1',state:'observed'});states.push(s.snapshot())}process.stdout.write(JSON.stringify(states))`
		cmd := exec.Command("node", "-e", script)
		cmd.Stdin = bytes.NewReader(data)
		out, e := cmd.CombinedOutput()
		if e != nil {
			t.Fatalf("JS lifecycle: %s: %v", out, e)
		}
		var actual []sessionOracle
		if json.Unmarshal(out, &actual) != nil || !reflect.DeepEqual(actual, m.expected) {
			t.Fatalf("state oracle mismatch: %s", out)
		}
	}, options(100)...)
}

package verification

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

func TestQuintObservedOperationNavigation(t *testing.T) {
	// Execute the actual browser kernel. Retain IDs, rather than validating only counts.
	script := `const o=require('../ui/operations.mjs');const make=(id,t,d,trace,outcome)=>({spanID:id,traceID:trace,operation:'reserve',role:'server',outcome,startNS:String(1792209000000000000n+BigInt(t)),durationNS:String(d)});const snapshot=[make('1',0,500000,'a','error'),make('2',1,25000000,'b','error'),make('3',2,25000000,'b','ok')];const subject={environment:'local',service:'checkout',namespace:'reservation',tenant:'demo',id:'same'};snapshot.forEach((s,i)=>s.events=[{timeNS:s.startNS,name:'accepted',subject:i===2?{...subject,tenant:'other'}:subject}]);const before=JSON.stringify(snapshot);let frames=[];const emit=(action,rows)=>frames.push({action,selected:rows.reduce((m,s)=>m|2**(Number(s.spanID)-1),0),basis:JSON.stringify(snapshot)===before?1:2});const errors=o.select(snapshot,{outcome:'error'});emit('select_errors',errors);const cell=o.distribution(errors,1792209000000,1792209060000).find(c=>c.band===2);const members=o.select(errors,cell);emit('select_cell',members);emit('open_trace',o.select(snapshot,{traceID:members[0].traceID}));emit('follow_subject',o.subjectSpans(snapshot,subject));emit('reset',o.select(snapshot,{}));console.log(JSON.stringify(frames));`
	out, err := exec.Command("node", "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v", out, err)
	}
	var frames []struct {
		Action   string `json:"action"`
		Selected int    `json:"selected"`
		Basis    int    `json:"basis"`
	}
	if err = json.Unmarshal(out, &frames); err != nil {
		t.Fatal(err)
	}
	if len(frames) != 5 {
		t.Fatal("no navigation trace")
	}
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor("navigation"), qobs.WithProcess("fixture"))
	for _, frame := range frames {
		rec.Record(context.Background(), frame.Action, nil, "obs.selected", frame.Selected, "obs.basis", frame.Basis)
	}
	checker, err := validate.NewChecker("model/operation_navigation.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	steps := sink.Steps()
	report, err := checker.Check(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK() {
		t.Fatal(report.Summary())
	}
	if err := artifact("operation-navigation.txt", []byte(report.Summary())); err != nil {
		t.Fatal(err)
	}
	for _, actor := range report.Actors {
		if actor.ITF != "" {
			data, err := os.ReadFile(actor.ITF)
			if err != nil {
				t.Fatal(err)
			}
			if err = artifact("subject-navigation.itf.json", data); err != nil {
				t.Fatal(err)
			}
		}
	}
	var lines []byte
	for i := range steps {
		line, err := steps[i].MarshalLine()
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, line...)
		lines = append(lines, '\n')
	}
	if err := artifact("subject-navigation-steps.jsonl", lines); err != nil {
		t.Fatal(err)
	}
	// A selected span outside the cell must be rejected, even if basis is retained.
	steps[1].Obs["selected"] = int64(3)
	bad, err := checker.Check(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if bad.OK() {
		t.Fatal("accepted wrong cell membership")
	}
	if err := artifact("operation-navigation-negative.txt", []byte(bad.Summary())); err != nil {
		t.Fatal(err)
	}
	steps[1].Obs["selected"] = int64(2)
	steps[3].Obs["selected"] = int64(7)
	badSubject, err := checker.Check(context.Background(), steps)
	if err != nil {
		t.Fatal(err)
	}
	if badSubject.OK() {
		t.Fatal("accepted cross-tenant subject membership")
	}
	if err := artifact("subject-navigation-negative.txt", []byte(badSubject.Summary())); err != nil {
		t.Fatal(err)
	}

}

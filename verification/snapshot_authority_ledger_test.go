package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type snapshotAuthorityStep struct {
	Action   string `json:"action"`
	Reads    int    `json:"reads"`
	Rows     int    `json:"rows"`
	Complete bool   `json:"complete"`
	Entry    string `json:"entry"`
}

const snapshotAuthorityTraceScript = `
import fs from 'node:fs';import {createSourceAdapter,importSnapshot,getSnapshotInfo,hasGuarantee,selectRows} from '../ui/snapshots.mjs';
const {snapshot:raw,question}=JSON.parse(fs.readFileSync(0,'utf8'));let reads=1;globalThis.fetch=()=>{reads++;throw Error('Unexpected constructor source IO')};const adapter=createSourceAdapter({id:'model-source'});const original=adapter.adoptSnapshot(raw,{question});
const ledger=[];const record=(action,snapshot)=>ledger.push({action,reads,rows:snapshot.rows.length,complete:hasGuarantee(snapshot,'groups'),entry:getSnapshotInfo(snapshot).entry});record('source',original);
const sliced=structuredClone(original);sliced.rows=sliced.rows.slice(0,1);const copy=importSnapshot(sliced);record('copy',copy);
const subset=selectRows(original,[original.rows[0].key]);record('subset',subset);record('restore',original);process.stdout.write(JSON.stringify(ledger));
`

// This deliberately small projected model checks ownership/effect observations,
// not authentication, source truth, completeness of collected telemetry or every
// possible projection. Reads count exploration requests, not internal SQL
// statements. Broader constructor schedules are exercised with Hegel.
func newSnapshotAuthorityChecker(t *testing.T) *validate.Checker {
	t.Helper()
	dir := t.TempDir()
	spec, err := filepath.Abs("model/snapshot_authority_265.qnt")
	if err != nil {
		t.Fatal(err)
	}
	binding := fmt.Sprintf(`spec: %q
module: snapshot_authority_265
constants: {ORIGINAL_ROWS: '2'}
invariants: [safety]
actions:
  source:
    expect: 'sourceReads == {{.obs.reads}} and rows == {{.obs.rows}} and complete == {{.obs.complete}} and entry == {{.obs.entry}}'
  copy:
    expect: 'sourceReads == {{.obs.reads}} and rows == {{.obs.rows}} and complete == {{.obs.complete}} and entry == {{.obs.entry}}'
  subset:
    expect: 'sourceReads == {{.obs.reads}} and rows == {{.obs.rows}} and complete == {{.obs.complete}} and entry == {{.obs.entry}}'
  restore:
    expect: 'sourceReads == {{.obs.reads}} and rows == {{.obs.rows}} and complete == {{.obs.complete}} and entry == {{.obs.entry}}'
`, spec)
	file := filepath.Join(dir, "binding.yaml")
	if err = os.WriteFile(file, []byte(binding), 0600); err != nil {
		t.Fatal(err)
	}
	checker, err := validate.NewChecker(file, validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	return checker
}

func checkSnapshotAuthorityLedger(t *testing.T, checker *validate.Checker, steps []snapshotAuthorityStep) bool {
	t.Helper()
	if len(steps) != 4 {
		return false
	}
	for i, name := range []string{"source", "copy", "subset", "restore"} {
		if steps[i].Action != name {
			return false
		}
	}
	sink := &qobs.MemorySink{}
	recorder := qobs.NewRecorder(sink)
	for _, step := range steps {
		if step.Entry != "source" && step.Entry != "held" && step.Entry != "projection" {
			return false
		}
		recorder.Record(context.Background(), step.Action, nil, "obs.reads", step.Reads, "obs.rows", step.Rows, "obs.complete", step.Complete, "obs.entry", step.Entry)
	}
	result, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	return result.OK()
}

func validateSnapshotAuthorityMutants(t *testing.T, steps []snapshotAuthorityStep) {
	t.Helper()
	checker := newSnapshotAuthorityChecker(t)
	if !checkSnapshotAuthorityLedger(t, checker, steps) {
		t.Fatal("actual constructor/projection ownership trace rejected")
	}
	for mutant := 0; mutant < 7; mutant++ {
		altered := append([]snapshotAuthorityStep(nil), steps...)
		switch mutant {
		case 0:
			altered[1].Complete = true
		case 1:
			altered[2].Complete = true
		case 2:
			altered[1].Reads++
		case 3:
			altered[2].Rows++
		case 4:
			altered[1].Entry = "source"
		case 5:
			altered[3].Rows--
		case 6:
			altered[3].Complete = false
		}
		if checkSnapshotAuthorityLedger(t, checker, altered) {
			t.Fatal("incorrect ownership/effect trace accepted", mutant)
		}
	}
	if checkSnapshotAuthorityLedger(t, checker, steps[:3]) {
		t.Fatal("conforming prefix mistaken for completed journey")
	}
}

func TestQuintSnapshotConstructionLedger(t *testing.T) {
	raw, q, err := snapshotSourceFixture(17)
	if err != nil {
		t.Fatal(err)
	}
	output, err := runSnapshotJS(raw, q, nil, snapshotAuthorityTraceScript)
	if err != nil {
		t.Fatal(err)
	}
	var steps []snapshotAuthorityStep
	if err = json.Unmarshal(output, &steps); err != nil {
		t.Fatal(err)
	}
	validateSnapshotAuthorityMutants(t, steps)
}

func TestQuintSnapshotAuthorityObservedBrowserLedger(t *testing.T) {
	file := os.Getenv("GOO_SNAPSHOT_AUTHORITY_LEDGER")
	if file == "" {
		t.Skip("actual browser ownership ledger required; set GOO_SNAPSHOT_AUTHORITY_LEDGER")
	}
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var steps []snapshotAuthorityStep
	if err = json.Unmarshal(raw, &steps); err != nil {
		t.Fatal(err)
	}
	validateSnapshotAuthorityMutants(t, steps)
}

package verification

import (
	"bytes"
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"math/big"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"testing"
)

func TestHegelFreshScopeContract(t *testing.T) {
	raw, e := os.ReadFile("../config/measurements/checkout-calls.yaml")
	if e != nil {
		t.Fatal(e)
	}
	d, e := measurements.Read(string(raw))
	if e != nil {
		t.Fatal(e)
	}
	hegel.Test(t, func(tc *hegel.T) {
		offset := hegel.Draw(tc, hegel.Integers(0, 999999))
		width := hegel.Draw(tc, hegel.Integers(1, 1000000000))
		from := int64(1790000000000000000) + int64(offset)
		to := from + int64(width)
		step := workbench.Step{Version: 1, Kind: "measurement", Scope: workbench.StepScope{Environment: d.Environment, FromMS: from / 1000000, ToMS: (to + 999999) / 1000000, FromNS: strconv.FormatInt(from, 10), ToNS: strconv.FormatInt(to, 10)}, View: "Overview", Measurement: &d}
		input, _ := json.Marshal(map[string]any{"step": step, "now": int64(1800000000000)})
		cmd := exec.Command("node", "-e", `const p=JSON.parse(require('fs').readFileSync(0,'utf8')),before=JSON.stringify(p.step),fresh=require('../ui/fresh-comparison').propose(p.step,p.now);process.stdout.write(JSON.stringify({fresh,unchanged:before===JSON.stringify(p.step)}))`)
		cmd.Stdin = bytes.NewReader(input)
		out, e := cmd.CombinedOutput()
		if e != nil {
			tc.Fatalf("node projection: %v %s", e, out)
		}
		var result struct {
			Fresh     workbench.Step
			Unchanged bool
		}
		if e = json.Unmarshal(out, &result); e != nil {
			tc.Fatal(e)
		}
		if e = result.Fresh.Validate(); e != nil {
			tc.Fatal(e)
		}
		a, _ := new(big.Int).SetString(result.Fresh.Scope.FromNS, 10)
		b, _ := new(big.Int).SetString(result.Fresh.Scope.ToNS, 10)
		if !result.Unchanged || new(big.Int).Sub(b, a).Cmp(big.NewInt(int64(width))) != 0 || !reflect.DeepEqual(result.Fresh.Measurement, step.Measurement) || result.Fresh.Scope.ToNS != "1800000000000000000" {
			tc.Fatal("fresh projection changed intent, duration or original")
		}
	}, options(100)...)
}

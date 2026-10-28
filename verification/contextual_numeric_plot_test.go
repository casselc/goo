package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// Invoke the actual client transformation. The oracle checks custody and Go
// compilation independently; it does not implement a second plotQuestion.
func TestHegelContextualNumericPlotScope(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 1000000))
		typ := hegel.Draw(tc, hegel.SampledFrom([]string{"int64", "uint64", "decimal64"}))
		kind := hegel.Draw(tc, hegel.SampledFrom([]string{"attribute", "resource"}))
		signal := hegel.Draw(tc, hegel.SampledFrom([]string{"spans", "logs"}))
		b := exploration.NumericBinding{Field: exploration.Field{Kind: kind, Key: fmt.Sprintf("amount.雪-%d", seed)}, Type: typ, Unit: "USD", Parse: "canonical-decimal"}
		if typ == "decimal64" {
			b.Scale = hegel.Draw(tc, hegel.Integers(1, 9))
		}
		exact := predicateBinding("uint64", "attribute")
		parent := exploration.Query{Signal: signal, Environment: fmt.Sprintf("scope-%d", seed), FromMS: 1800000000000 + int64(seed)*1000, Limit: 1 + seed%999, Offset: seed % 90, GroupBy: []exploration.Field{{Kind: "column", Key: "service"}}, BucketMS: 1000, GroupTimeSeries: true, Where: &exploration.Predicate{All: []exploration.Predicate{{Numeric: &exact, Op: "gte", Value: "9007199254740993"}, {Any: []exploration.Predicate{{Field: &exploration.Field{Kind: "attribute", Key: "tenant"}, Op: "missing"}, {Field: &exploration.Field{Kind: "resource", Key: "tenant"}, Op: "eq", Value: ""}}}}}}
		if signal == "spans" && seed%2 == 0 {
			ns := &identity.Namespace{Mode: "missing"}
			if seed%4 == 0 {
				ns = &identity.Namespace{Mode: "value", Value: fmt.Sprintf("team-%d", seed)}
			}
			parent.ReportedEdge = &exploration.ReportedEdge{Parent: exploration.EdgeService{Service: "checkout", Namespace: ns}, Child: exploration.EdgeService{Service: "worker", Namespace: ns}, DiscoveryNamespace: ns}
		}
		parent.ToMS = parent.FromMS + int64(1+seed%600)*1000
		if signal == "spans" {
			parent.Summary = "span-duration"
			parent.SpanStatus = true
		}
		input := struct {
			Parent  exploration.Query          `json:"parent"`
			Binding exploration.NumericBinding `json:"binding"`
		}{parent, b}
		raw, err := json.Marshal(input)
		if err != nil {
			tc.Fatal(err)
		}
		const script = `const fs=require('fs'),p=require('./ui/numeric-plot.mjs');const x=JSON.parse(fs.readFileSync(0,'utf8')),before=JSON.stringify(x),q=p.plotQuestion(x.parent,x.binding),snapshot=structuredClone(q);q.where.all[0].value='0';q.numeric.field.key='mutated';console.log(JSON.stringify({question:snapshot,unchanged:JSON.stringify(x)===before}));`
		cmd := exec.Command("node", "-e", script)
		cmd.Dir = ".."
		cmd.Stdin = bytes.NewReader(raw)
		output, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("actual plot transform: %v %s", err, output)
		}
		var got struct {
			Question  exploration.Query `json:"question"`
			Unchanged bool              `json:"unchanged"`
		}
		if err = json.Unmarshal(output, &got); err != nil {
			tc.Fatal(err)
		}
		q := got.Question
		if !got.Unchanged || !reflect.DeepEqual(q.Where, parent.Where) || q.Signal != parent.Signal || q.Environment != parent.Environment || q.FromMS != parent.FromMS || q.ToMS != parent.ToMS || !reflect.DeepEqual(q.ReportedEdge, parent.ReportedEdge) {
			tc.Fatal("original scope/custody changed")
		}
		if q.Offset != 0 || q.Limit != 100 || len(q.GroupBy) != 0 || q.GroupTimeSeries || q.BucketMS != 0 || q.Summary != "" || q.SpanStatus || !reflect.DeepEqual(q.Numeric, &b) || q.NumericHistogram == nil || q.NumericHistogram.Binning != "powers-of-two" {
			tc.Fatal("distribution carries stale presentation or wrong binding")
		}
		compiled, err := exploration.Compile(q)
		if err != nil {
			tc.Fatal(err)
		}
		if compiled.Parameters["from"] != strconv.FormatInt(parent.FromMS, 10) || compiled.Parameters["to"] != strconv.FormatInt(parent.ToMS, 10) || compiled.Parameters["environment"] != parent.Environment {
			tc.Fatal("Go compiler changed scoped query")
		}
		// Retained >2^53 predicates remain strings across Node authoring and Go wire.
		if q.Where.All[0].Value != "9007199254740993" {
			tc.Fatal("typed predicate rounded")
		}
	}, options(70)...)
}

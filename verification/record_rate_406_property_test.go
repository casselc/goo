package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
)

type rateTransport406 func(*http.Request) (*http.Response, error)

func (f rateTransport406) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type rateVector406 struct {
	Data        *exploration.Result             `json:"data"`
	Destination exploration.Query               `json:"destination"`
	Summary     workbench.ExplorationView       `json:"summary"`
	Timeline    workbench.ExplorationView       `json:"timeline"`
	Expansion   workbench.ViewTemplateExpansion `json:"expansion"`
}

// This synthetic transport exercises actual public source projection and carrier
// constructors. Its arbitrary UInt64 counts are arithmetic/property inputs, not
// additional observations; the separate live ledger supplies raw-source proof.
func rateSource406(t testing.TB, q exploration.Query, count uint64) *exploration.Result {
	t.Helper()
	var reads atomic.Int32
	source := exploration.Source{Endpoint: "http://rate-property.invalid", Client: &http.Client{Transport: rateTransport406(func(r *http.Request) (*http.Response, error) {
		reads.Add(1)
		wire, e := io.ReadAll(r.Body)
		if e != nil {
			return nil, e
		}
		sql := string(wire)
		body := ""
		switch {
		case strings.Contains(sql, "GROUP BY from_ms"):
			for from := q.FromMS; from < q.ToMS; from += q.BucketMS {
				n := uint64(0)
				if from == q.FromMS && count > 0 {
					n = count - 1
				}
				if from+q.BucketMS >= q.ToMS && count > 0 {
					n = 1
				}
				if n > 0 {
					body += fmt.Sprintf("{\"from_ms\":\"%d\",\"count\":\"%d\"}\n", from, n)
				}
			}
		case strings.Contains(sql, "SELECT DISTINCT"):
		case strings.Contains(sql, "count()"):
			body = fmt.Sprintf("{\"count\":\"%d\"}", count)
		default:
			for i := uint64(0); i < min(count, 2); i++ {
				body += fmt.Sprintf("{\"timestamp_ns\":\"%d\",\"duration_ns\":\"0\",\"SpanId\":\"%016x\"}\n", q.FromMS*1000000+int64(i), i+1)
			}
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
	})}}
	data, err := source.Fetch(t.Context(), q)
	if err != nil {
		t.Fatal(err)
	}
	if reads.Load() != 4 || data.SummaryDataset == nil || data.TimeDataset == nil {
		t.Fatalf("four existing source statements must supply both datasets: %d", reads.Load())
	}
	return data
}
func rateExpected406(count string, elapsed int64) string {
	n, _ := new(big.Int).SetString(count, 10)
	n.Mul(n, big.NewInt(1000))
	return new(big.Rat).SetFrac(n, big.NewInt(elapsed)).RatString()
}
func rateView406(t testing.TB, q exploration.Query, timeline bool) workbench.ExplorationView {
	t.Helper()
	records := map[string]any{"records": map[string]any{"version": "goo.view/1", "scope": map[string]string{"kind": "record-query", "signal": q.Signal}, "kind": map[string]any{"name": "record-table", "version": 1}, "display": map[string]any{"columns": []string{"utc", "summary"}}}}
	role := "summary"
	kind := "measure-summary"
	display := map[string]any{"fields": []string{"record.rate", "count"}, "title": "Captured activity"}
	if timeline {
		role = "timeline"
		kind = "time-bar"
		display = map[string]any{"value": "record.rate", "title": "Captured activity"}
	}
	leaf := map[string]any{role: map[string]any{"version": "goo.view/1", "scope": map[string]string{"kind": "record-query", "signal": q.Signal}, "kind": map[string]any{"name": kind, "version": 1}, "display": display}}
	children := []any{leaf, records}
	if timeline {
		summary := map[string]any{"summary": map[string]any{"version": "goo.view/1", "scope": map[string]string{"kind": "record-query", "signal": q.Signal}, "kind": map[string]any{"name": "measure-summary", "version": 1}, "display": map[string]any{"fields": []string{"count"}}}}
		children = []any{leaf, summary, records}
	}
	wire := map[string]any{"version": 1, "title": "Rate facts", "queries": []exploration.Query{q}, "composition": map[string]any{"version": "goo.composition/1", "kind": map[string]any{"name": "stack", "version": 1}, "children": children}}
	if !timeline {
		wire["template"] = map[string]string{"version": "goo.template/1", "group": "none", "profile": "ordinary-summary"}
	}
	raw, _ := json.Marshal(wire)
	var view workbench.ExplorationView
	if err := json.Unmarshal(raw, &view); err != nil {
		t.Fatal(err)
	}
	if err := view.Validate(); err != nil {
		t.Fatal(err)
	}
	return view
}
func TestHegelRecordRate406(t *testing.T) {
	vectors := []rateVector406{}
	store := workbench.Store{Dir: t.TempDir()}
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		for _, signal := range []string{"logs", "spans"} {
			for _, count := range []uint64{0, 1, 7, 9007199254740993, 18446744073709551615} {
				q := exploration.Query{Signal: signal, Environment: "rate406", FromMS: int64(1000 + seed), ToMS: int64(3501 + seed), BucketMS: 1000, Limit: 1, Offset: 0}
				data := rateSource406(t, q, count)
				for _, d := range []*exploration.ViewDataset{data.SummaryDataset, data.TimeDataset} {
					for _, row := range d.Rows {
						elapsed := q.ToMS - q.FromMS
						if d.Shape == "time-series" {
							from, _ := strconv.ParseInt(*row.Values["fromMs"], 10, 64)
							to, _ := strconv.ParseInt(*row.Values["toMs"], 10, 64)
							elapsed = to - from
						}
						if row.Values["record.rate"] == nil || *row.Values["record.rate"] != rateExpected406(*row.Values["count"], elapsed) {
							tc.Fatal("rate does not equal exact independent count/exposure fraction")
						}
					}
				}
				summary, timeline := rateView406(t, q, false), rateView406(t, q, true)
				for _, view := range []workbench.ExplorationView{summary, timeline} {
					saved, e := store.CreateExplorationView(view)
					if e != nil {
						tc.Fatal(e)
					}
					loaded, e := store.LoadExplorationView(saved.ID)
					if e != nil || !reflect.DeepEqual(saved, loaded) {
						tc.Fatal("whole rate view store round trip", e)
					}
				}
				destination := q
				destination.Environment = "destination406"
				destination.FromMS += 100000
				destination.ToMS += 100137
				destination.Limit = 7
				before, _ := json.Marshal(summary)
				expanded, e := workbench.ExpandViewTemplate(workbench.ViewTemplateExpansionRequest{Source: &summary, Query: &destination})
				if e != nil {
					tc.Fatal(e)
				}
				after, _ := json.Marshal(summary)
				if !bytes.Equal(before, after) || !reflect.DeepEqual(expanded.View.Queries, []exploration.Query{destination}) {
					tc.Fatal("rate template replaced destination query or donor")
				}
				vectors = append(vectors, rateVector406{data, destination, summary, timeline, expanded})
			}
		}
	}, options(8)...)
	raw, _ := json.Marshal(vectors)
	cmd := exec.CommandContext(t.Context(), "node", "verification/record_rate_406_fixture.mjs")
	cmd.Dir = ".."
	cmd.Stdin = bytes.NewReader(raw)
	out, e := cmd.CombinedOutput()
	if e != nil {
		t.Fatalf("rate Go/JS bridge: %v %s", e, out)
	}
	t.Log(strings.TrimSpace(string(out)))
	if e = os.MkdirAll("../artifacts/verification", 0755); e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile("../artifacts/verification/record-rate406-carriers.json", raw, 0644); e != nil {
		t.Fatal(e)
	}
	t.Logf("%d complete summary/timeline/template source carriers including exact large counts and clipped buckets", len(vectors))
}

func TestRecordRate406SupportProof(t *testing.T) {
	q := exploration.Query{Signal: "logs", Environment: "rate406", FromMS: 1000, ToMS: 3501, BucketMS: 1000, Limit: 1}
	data := rateSource406(t, q, 7)
	for _, original := range []*exploration.ViewDataset{data.SummaryDataset, data.TimeDataset} {
		if err := exploration.ValidateRecordRateDataset(original); err != nil {
			t.Fatal(err)
		}
		for _, mutate := range []func(*exploration.ViewDataset){
			func(d *exploration.ViewDataset) { wrong := "99"; d.Rows[0].Values["record.rate"] = &wrong },
			func(d *exploration.ViewDataset) { d.Rows[0].Cells["record.rate"].Rate.Count = "99" },
			func(d *exploration.ViewDataset) { d.Rows[0].Cells["record.rate"].Rate.ElapsedMS = "1" },
			func(d *exploration.ViewDataset) {
				cell := d.Rows[0].Cells["record.rate"]
				cell.Rate = nil
				d.Rows[0].Cells["record.rate"] = cell
			},
			func(d *exploration.ViewDataset) {
				var selector exploration.Query
				_ = json.Unmarshal(d.Scope.Selector, &selector)
				selector.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}}
				d.Scope.Selector, _ = json.Marshal(selector)
			},
		} {
			raw, _ := json.Marshal(original)
			var bad exploration.ViewDataset
			_ = json.Unmarshal(raw, &bad)
			mutate(&bad)
			if exploration.ValidateRecordRateDataset(&bad) == nil {
				t.Fatal("forged exact rate support admitted")
			}
		}
	}

	rawUnknown, _ := json.Marshal(data.TimeDataset)
	var unknown exploration.ViewDataset
	_ = json.Unmarshal(rawUnknown, &unknown)
	unknown.Rows[0].State = "unknown"
	unknown.Rows[0].Values["count"] = nil
	unknown.Rows[0].Values["record.rate"] = nil
	unknown.Rows[0].Cells["record.rate"] = exploration.ViewCell{State: "unknown"}
	if err := exploration.ValidateRecordRateDataset(&unknown); err != nil {
		t.Fatal("unreported interval must remain expressible", err)
	}
	zero := "0"
	unknown.Rows[0].Values["record.rate"] = &zero
	if exploration.ValidateRecordRateDataset(&unknown) == nil {
		t.Fatal("unknown interval accepted fabricated zero")
	}
	for _, count := range []string{"0", "1", "18446744073709551615"} {
		value, cell, err := exploration.RecordRateCell(count, 501, false)
		if err != nil || value == nil || *value != rateExpected406(count, 501) || cell.State != "derived" || cell.Rate.Count != count || cell.Rate.ElapsedMS != "501" {
			t.Fatal("exact clipped support", count, err)
		}
	}
	value, cell, err := exploration.RecordRateCell("", 501, true)
	if err != nil || value != nil || cell.State != "unknown" || cell.Rate != nil {
		t.Fatal("unknown must have neither invented zero nor exposure facts")
	}
	for _, count := range []string{"01", "-1", "18446744073709551616", "1.0"} {
		if _, _, err := exploration.RecordRateCell(count, 1000, false); err == nil {
			t.Fatal("invalid UInt64", count)
		}
	}
	for _, elapsed := range []int64{0, -1} {
		if _, _, err := exploration.RecordRateCell("1", elapsed, false); err == nil {
			t.Fatal("nonpositive exposure")
		}
	}
}

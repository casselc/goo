package exploration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestViewPartitionUsesExactGroupReadIncludingMissingAndEmpty(t *testing.T) {
	r := viewFixture()
	// Neither an independently read total nor a bounded record page defines this
	// partition. Source.Fetch separately checks cross-read coherence in production.
	r.Count = "7"
	r.Rows = []json.RawMessage{json.RawMessage(`{"timestamp_ns":"1"}`)}
	d := groupViewDataset(r, true)
	want := &ViewPartition{Kind: "disjoint-record-groups", Measure: "count", Total: "18446744073709551615"}
	if d == nil || !reflect.DeepEqual(d.Partition, want) {
		t.Fatalf("group partition: %#v", d)
	}
	if d.Rows[0].Values["category"] != nil || *d.Rows[1].Values["category"] != "" || len(d.Rows) != 2 {
		t.Fatal("missing and empty groups were not separate partition members", d.Rows)
	}
	wire, err := json.Marshal(d)
	if err != nil || !strings.Contains(string(wire), `"total":"18446744073709551615"`) {
		t.Fatal("partition total lost exact decimal encoding", string(wire), err)
	}
}

func TestViewPartitionRequiresExhaustiveCoherentGroups(t *testing.T) {
	for _, tc := range []struct {
		name       string
		exhaustive bool
		state      string
	}{
		{"unbounded claim cannot establish partition", false, "observed"},
		{"cross-read disagreement", true, "changed_between_reads"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := viewFixture()
			r.State = tc.state
			d := groupViewDataset(r, tc.exhaustive)
			if d == nil || d.Partition != nil || d.Completeness != nil {
				t.Fatal("unknown groups gained an exhaustive partition", d)
			}
		})
	}
	r := viewFixture()
	r.State, r.Count, r.Groups = "no_data", "0", nil
	d := groupViewDataset(r, true)
	if d == nil || d.Partition == nil || d.Partition.Total != "0" || len(d.Rows) != 0 || d.Completeness == nil {
		t.Fatal("empty captured partition must retain exact zero, not invent a percentage", d)
	}
}

func TestViewPartitionOverflowOmitsDenominatorWithoutWrappingRows(t *testing.T) {
	r := viewFixture()
	r.Groups = []json.RawMessage{
		json.RawMessage(`{"g0":"a","count":"18446744073709551615"}`),
		json.RawMessage(`{"g0":"b","count":"1"}`),
		json.RawMessage(`{"g0":"c","count":"18446744073709551615"}`),
	}
	d := groupViewDataset(r, true)
	if d == nil || d.Partition != nil || d.Completeness == nil || len(d.Rows) != 3 || *d.Rows[0].Values["count"] != "18446744073709551615" {
		t.Fatal("overflow must make shares unavailable without replacing exact supplied values", d)
	}
}

func TestViewPartitionPreservesMeasureSupportAndExactMethods(t *testing.T) {
	b := numericBinding()
	r := measureGroupFixture(t, &b, numericValues("1", "2", "3"), []uint64{0, 0, 1})
	before, err := json.Marshal(r.Groups)
	if err != nil {
		t.Fatal(err)
	}
	d := groupViewDataset(r, true)
	if d == nil || d.Partition == nil || d.Partition.Total != "3" || *d.Rows[0].Values["numeric.mean"] != "2" || *d.Rows[0].Values["duration.mean"] != "1/3" || *d.Rows[0].Values["duration.p95"] != "1" {
		t.Fatal("partition changed exact measure semantics", d)
	}
	if got := d.Rows[0].Cells["numeric.mean"].Support; got == nil || got.Total != "3" || got.Valid != "3" {
		t.Fatal("numeric support changed", got)
	}
	for _, field := range d.Fields {
		if field.Key == "duration.p95" && field.Aggregation.Method != "quantilesExactHigh" {
			t.Fatal("quantile method changed", field)
		}
	}
	after, _ := json.Marshal(r.Groups)
	if string(after) != string(before) {
		t.Fatal("partition construction mutated its source groups")
	}
}

func TestSourcePartitionCoherenceWithoutExtraReads(t *testing.T) {
	for _, total := range []string{"3", "5"} {
		t.Run("separate count "+total, func(t *testing.T) {
			reads := 0
			source := Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: countedValuesTransport(func(req *http.Request) (*http.Response, error) {
				reads++
				body, _ := io.ReadAll(req.Body)
				sql, text := string(body), ""
				switch {
				case strings.Contains(sql, "GROUP BY"):
					text = "{\"g0\":null,\"count\":\"2\"}\n{\"g0\":\"\",\"count\":\"1\"}"
				case strings.Contains(sql, "SELECT DISTINCT"):
				case strings.Contains(sql, "count()"):
					text = `{"count":"` + total + `"}`
				default:
					text = `{"timestamp_ns":"1","duration_ns":"0"}`
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(text))}, nil
			})}}
			q := basic()
			q.GroupBy = []Field{{Kind: "attribute", Key: "team"}}
			r, err := source.Fetch(context.Background(), q)
			if err != nil || reads != 4 || r == nil || r.ViewDataset == nil {
				t.Fatal("source read count changed", r, err, reads)
			}
			if total == "3" {
				if r.State != "observed" || r.ViewDataset.Partition == nil || r.ViewDataset.Partition.Total != "3" || len(r.Rows) != 1 {
					t.Fatal("partition was not emitted from its exhaustive groups independently of the bounded row page", r)
				}
			} else if r.State != "changed_between_reads" || r.ViewDataset.Partition != nil || r.ViewDataset.Completeness != nil {
				t.Fatal("incoherent source gained a full-share denominator", r)
			}
		})
	}
}

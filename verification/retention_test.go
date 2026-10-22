package verification

import (
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/retention"
	"hegel.dev/go/hegel"
	"reflect"
	"strconv"
	"testing"
)

func TestHegelRetainedSelectionIdentity(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 20))
		value := hegel.Draw(tc, hegel.Integers(0, 1000000))
		raw := json.RawMessage(`{ "timestamp_ns":"1792629000000000001", "empty":"", "large":9007199254740993, "value":` + strconv.Itoa(value) + `}`)
		q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}
		c := &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Version: 1, ID: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", CapturedAt: "2026-10-22T00:00:00Z", Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Count: "100", Basis: "fixed", Rows: []json.RawMessage{}}}
		indexes := []int{}
		reverse := []int{}
		for i := 0; i < n; i++ {
			c.Explore.Rows = append(c.Explore.Rows, append(json.RawMessage{}, raw...))
			indexes = append(indexes, i)
			reverse = append([]int{i}, reverse...)
		}
		m, e := retention.Select(c, indexes)
		if e != nil {
			tc.Fatal(e)
		}
		other, e := retention.Select(c, reverse)
		if e != nil || !reflect.DeepEqual(m, other) || m.Validate() != nil {
			tc.Fatal("selection order changed identity", e)
		}
		if len(m.Members) != n {
			tc.Fatal("identical occurrences collapsed")
		}
		for _, member := range m.Members {
			var decoded map[string]json.RawMessage
			if json.Unmarshal(member.Row, &decoded) != nil || string(decoded["large"]) != "9007199254740993" || string(decoded["timestamp_ns"]) != `"1792629000000000001"` {
				tc.Fatal("numeric/time token changed")
			}
		}
		m.Members[0].Row = json.RawMessage(`{}`)
		if m.Validate() == nil {
			tc.Fatal("mutated record accepted")
		}
		if _, e = retention.Select(c, append(indexes, indexes[0])); e == nil {
			tc.Fatal("duplicate selection accepted")
		}
	}, options(100)...)
}

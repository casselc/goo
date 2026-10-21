package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
	"reflect"
	"strconv"
	"testing"
)

func TestHegelTracePageExamples(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(1, 30))
		rows := []json.RawMessage{}
		for i := 0; i < n; i++ {
			d := uint64(hegel.Draw(tc, hegel.Integers(0, 1000000)))
			if i%3 == 0 {
				d += 9007199254740992
			}
			if i%7 == 0 {
				d = 18446744073709551615
			}
			status := "Ok"
			if i%4 == 0 {
				status = "Error"
			}
			b, _ := json.Marshal(map[string]string{"TraceId": fmt.Sprintf("%032x", i/2+1), "SpanId": fmt.Sprintf("%016x", i+1), "duration_ns": strconv.FormatUint(d, 10), "timestamp_ns": "1500000000", "StatusCode": status})
			rows = append(rows, b)
		}
		r := exploration.Result{Question: exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 2000, Limit: 100}, State: "observed", Rows: rows}
		a := exploration.SelectTraceExamples(r)
		for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
			rows[i], rows[j] = rows[j], rows[i]
		}
		b := exploration.SelectTraceExamples(r)
		if !reflect.DeepEqual(a, b) || len(a.Items) > 3 {
			tc.Fatal("order changed examples or exceeded bound")
		}
		seen := map[string]bool{}
		for _, v := range a.Items {
			if seen[v.TraceID] {
				tc.Fatal("duplicate trace")
			}
			seen[v.TraceID] = true
			found := false
			for _, raw := range rows {
				var row map[string]string
				json.Unmarshal(raw, &row)
				if row["TraceId"] == v.TraceID && row["SpanId"] == v.SpanID && row["duration_ns"] == v.DurationNS {
					found = true
				}
			}
			if !found {
				tc.Fatal("invented page member")
			}
		}
		r.State = "changed_between_reads"
		if len(exploration.SelectTraceExamples(r).Items) != 0 {
			tc.Fatal("incoherent examples exposed")
		}
	}, options(100)...)
}

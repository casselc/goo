package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// Generated receipt membership is an independent oracle, not a replay of a
// production matching helper. Other-signal and unadmitted fields are decoys.
func TestKeptContentSearchProperties330(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		count := hegel.Draw(tc, hegel.Integers(21, 75))
		text := hegel.Draw(tc, hegel.SampledFrom([]string{"Needle", ".*", "界🙂", "é", " e "}))
		stride := hegel.Draw(tc, hegel.Integers(1, 9))
		filter := hegel.Draw(tc, hegel.SampledFrom([]string{"", "logs", "spans"}))
		store := exploration.KeptObservationStore{Dir: t.TempDir()}
		env, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
		dto, _ := env.Selector()
		type expectation struct {
			records, index int
			field          string
		}
		oracle := map[string]expectation{}
		eligible := 0
		for i := 0; i < count; i++ {
			signal, primary, other := "logs", "Body", "SpanName"
			if i%2 == 1 {
				signal, primary, other = "spans", "SpanName", "Body"
			}
			rows := []json.RawMessage{}
			want := expectation{index: -1}
			for j := 0; j < 1+i%3; j++ {
				record := map[string]any{"timestamp_ns": "150000000000", "ResourceAttributes": map[string]string{}, primary: "ordinary", other: text, "Hidden": text, "ServiceName": strings.ToLower(text)}
				if (i+j)%stride == 0 {
					record[[]string{primary, "ServiceName", "TraceId", "SpanId"}[(i+j)%4]] = strings.Repeat("界", 400) + text + strings.Repeat("🙂", 400)
				}
				for _, field := range []string{primary, "ServiceName", "TraceId", "SpanId"} {
					if value, ok := record[field].(string); ok && strings.Contains(value, text) {
						want.records++
						if want.index < 0 {
							want.index = j
							want.field = field
						}
						break
					}
				}
				raw, e := json.Marshal(record)
				if e != nil {
					tc.Fatal(e)
				}
				rows = append(rows, raw)
			}
			q := exploration.Query{Signal: signal, EnvironmentSelector: &dto, FromMS: 100000, ToMS: 160000, Limit: 100}
			bank := new(exploration.ResultSnapshots)
			token, e := bank.Freeze(&exploration.Result{Question: q, Rows: rows, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Basis: strings.Repeat("a", 64), State: "observed"})
			if e != nil {
				tc.Fatal(e)
			}
			indexes := []int{}
			for j := range rows {
				indexes = append(indexes, j)
			}
			receipt, e := store.KeepRecords(context.Background(), bank, exploration.CaptureStore{}, exploration.KeepRecordsIntent{Version: exploration.KeepRecordsVersion, RequestID: fmt.Sprintf("%064x", i+1), ResultToken: token, Indexes: indexes})
			if e != nil {
				tc.Fatal(e)
			}
			if filter == "" || filter == signal {
				eligible++
				if want.records > 0 {
					oracle[receipt.ID] = want
				}
			}
		}
		seen := map[string]bool{}
		cursor := ""
		checked, records := 0, 0
		for n := 0; ; n++ {
			if n > count {
				tc.Fatal("cursor failed to progress")
			}
			page, e := store.Search(context.Background(), exploration.KeptSearchQuery{Text: text, Signal: filter}, cursor)
			if e != nil {
				tc.Fatal(e)
			}
			if page.Total != count || page.Eligible != eligible || page.Checked < 1 || page.Checked > 64 || page.Bytes < 1 || page.Bytes > 32<<20 || len(page.Items) > 20 || page.Matched != len(page.Items) {
				tc.Fatal("page accounting", page)
			}
			checked += page.Checked
			pageRecords := 0
			for _, item := range page.Items {
				want, ok := oracle[item.ID]
				if !ok || seen[item.ID] || item.Match == nil {
					tc.Fatal("false positive or duplicate", item.ID)
				}
				seen[item.ID] = true
				m := item.Match
				if m.Records != want.records || m.Witness.MemberIndex != want.index || m.Witness.Field != want.field || !strings.Contains(m.Witness.Text, text) || !utf8.ValidString(m.Witness.Text) || utf8.RuneCountInString(m.Witness.Text) > 320 {
					tc.Fatal("witness/occurrence mismatch", m, want)
				}
				pageRecords += m.Records
			}
			if page.MatchingRecords != pageRecords {
				tc.Fatal("matching record accounting")
			}
			records += pageRecords
			if !page.HasMore {
				if page.NextCursor != "" {
					tc.Fatal("terminal cursor")
				}
				break
			}
			if page.NextCursor == "" || page.NextCursor == cursor {
				tc.Fatal("nonprogress cursor")
			}
			if _, e = store.Search(context.Background(), exploration.KeptSearchQuery{Text: text + "x", Signal: filter}, page.NextCursor); e == nil {
				tc.Fatal("cursor accepted changed query")
			}
			cursor = page.NextCursor
		}
		expectedRecords := 0
		for _, want := range oracle {
			expectedRecords += want.records
		}
		if len(seen) != len(oracle) || checked != eligible || records != expectedRecords {
			tc.Fatal("incomplete continuation", len(seen), len(oracle), checked, eligible)
		}
	}, options(20)...)
}

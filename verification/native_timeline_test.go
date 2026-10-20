package verification

import (
	"context"
	"encoding/json"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestQuintNativeTimelineLedger(t *testing.T) {
	name := os.Getenv("GOO_NATIVE_TIMELINE_LEDGER")
	if name == "" {
		t.Skip("real-source opt-in")
	}
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	check := func(mutation int) bool {
		type point struct {
			Key, FromNS, ToNS, State, Label string
			Value                           *string
		}
		var l struct {
			Entries []struct {
				Result measurements.Result
				Data   struct {
					FromNS, ToNS, Kind, Unit string
					Total, Omitted           int
					Points                   []point
				}
				Before, After int
			}
			Capture, Reopened exploration.CapturePage
			RecipientReads    int
		}
		if e := json.Unmarshal(raw, &l); e != nil {
			t.Fatal(e)
		}
		switch mutation {
		case 0:
			l.Entries[0].Data.Points[0].Value = stringPtr("999")
		case 1:
			l.Entries[0].Data.Points[3].Value = stringPtr("0")
		case 2:
			l.Entries[1].Data.Points[0].Value = stringPtr("10")
		case 3:
			l.Entries[1].Data.Points[1].State = "observed"
		case 4:
			l.Entries[2].Data.Points[2].FromNS = l.Entries[2].Data.Points[1].ToNS
		case 5:
			l.Entries[2].Data.Points[0].Label = "P95 0.5"
		case 6:
			l.Entries[0].After++
		case 7:
			l.Capture.Receipt.Question.Measurement.Definition.Source.Unit = "wrong"
		case 8:
			l.Reopened.Receipt.ID = "wrong"
		}
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink)
		record := func(a string, b bool) { rec.Record(context.Background(), a, nil, "obs.coherent", b) }
		g, c, h := l.Entries[0], l.Entries[1], l.Entries[2]
		gp, cp, hp := g.Data.Points, c.Data.Points, h.Data.Points
		record("gauge", len(gp) == 4 && gp[0].Value != nil && *gp[0].Value == "-2" && gp[0].State == "stale" && gp[1].Value != nil && *gp[1].Value == "0" && gp[3].State == "unknown" && gp[3].Value == nil)
		record("counter", len(cp) == 4 && cp[0].Value != nil && *cp[0].Value == "5" && cp[1].State == "unknown" && cp[1].Value == nil && strings.Contains(cp[1].Label, "counter-decrease") && cp[3].State == "unknown")
		hs := h.Result.Histograms.Series[0]
		bounds := len(hp) == 4 && hp[2].FromNS == hs.Intervals[2].FromNS && hp[2].FromNS != hp[1].ToNS && hp[1].Value == nil && hp[2].Value != nil && *hp[2].Value == "0" && strings.Contains(hp[2].Label, "P95: no observations") && strings.Contains(hp[0].Label, "(0.5, +∞)") && strings.Contains(hp[3].Label, "(0.8, +∞)")
		record("histogram", bounds)
		readonly := true
		for _, e := range l.Entries {
			readonly = readonly && e.Before == e.After
			d := e.Result.Question.Definition
			readonly = readonly && d.Source.Namespace != nil && d.Source.Namespace.Value == "payments" && len(d.Source.Services) == 1 && d.Source.Services[0] == "api"
		}
		record("inspect", readonly)
		record("capture", reflect.DeepEqual(l.Capture.Receipt.Question.Measurement, &h.Result.Question) && reflect.DeepEqual(l.Capture.Measurement.Histograms, h.Result.Histograms))
		record("reopen", l.RecipientReads == 0 && reflect.DeepEqual(l.Capture, l.Reopened))
		checker, e := validate.NewChecker("model/native_timeline.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		result, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return result.OK()
	}
	if !check(-1) {
		t.Fatal("actual native timeline ledger rejected")
	}
	for i := 0; i < 9; i++ {
		if check(i) {
			t.Fatal("corrupt native timeline trace accepted", i)
		}
	}
}
func stringPtr(s string) *string { return &s }

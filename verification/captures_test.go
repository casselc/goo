package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

type captureMachine struct {
	store           exploration.CaptureStore
	source          exploration.Source
	count, revision int
	captures        []*exploration.SavedCapture
	first           map[string]string
}

func (m *captureMachine) RuleArrive(tc hegel.TestCase) {
	m.count += hegel.Draw(tc, hegel.Integers(1, 60))
	m.revision++
}
func (m *captureMachine) RuleResetSource(tc hegel.TestCase) {
	m.count = hegel.Draw(tc, hegel.Integers(0, 1200))
	m.revision++
}
func (m *captureMachine) save(parent string) {
	q := exploration.Query{Signal: "spans", Environment: "prod", FromMS: 1000, ToMS: 61000, Limit: 100}
	c, e := m.store.Create(context.Background(), m.source, "Investigation", exploration.CaptureQuestion{Kind: "explore", Explore: &q}, parent)
	if e != nil {
		panic(e)
	}
	page, _ := c.Page("")
	b, _ := json.Marshal(page)
	m.first[c.ID] = string(b)
	m.captures = append(m.captures, c)
	if len(m.captures) > 4 {
		m.captures = m.captures[1:]
	}
	if c.Explore.Count != strconv.Itoa(m.count) || len(c.Explore.Rows) != min(m.count, 1000) || c.Explore.HasMore != (m.count > 1000) {
		panic("capture changed count or retention semantics")
	}
}
func (m *captureMachine) RuleCapture(tc hegel.TestCase) { m.save("") }
func (m *captureMachine) RuleRefresh(tc hegel.TestCase) {
	if len(m.captures) == 0 {
		return
	}
	p := m.captures[hegel.Draw(tc, hegel.Integers(0, len(m.captures)-1))]
	m.save(p.ID)
}
func (m *captureMachine) InvariantStableEvidence(tc hegel.TestCase) {
	for _, old := range m.captures {
		c, e := m.store.Load(old.ID)
		if e != nil {
			panic(e)
		}
		first, e := c.Page("")
		if e != nil {
			panic(e)
		}
		b, _ := json.Marshal(first)
		if string(b) != m.first[c.ID] {
			panic("source arrival or refresh mutated predecessor")
		}
		rows := []json.RawMessage{}
		cur := ""
		for {
			p, e := c.Page(cur)
			if e != nil {
				panic(e)
			}
			if p.Explore.Count != old.Explore.Count || p.Explore.Basis != old.Explore.Basis || len(p.Explore.Rows) > 100 {
				panic("page changed basis or population")
			}
			rows = append(rows, p.Explore.Rows...)
			if p.NextCursor == "" {
				break
			}
			cur = p.NextCursor
		}
		if !reflect.DeepEqual(rows, old.Explore.Rows) {
			panic("paging skipped, duplicated or changed records")
		}
	}
}

// Search interleaves with source arrivals, capture/refresh and immutable reopen.
func (m *captureMachine) RuleSearch(tc hegel.TestCase) {
	titles := []string{"", "investigation", "[.*", "not present"}
	q := exploration.CaptureSearchQuery{Title: titles[hegel.Draw(tc, hegel.Integers(0, len(titles)-1))]}
	if hegel.Draw(tc, hegel.Booleans()) {
		q.Order = "oldest"
	} else {
		q.Order = "newest"
	}
	env := "other"
	switch hegel.Draw(tc, hegel.Integers(0, 2)) {
	case 1:
		env = "prod"
		q.Environment = &env
	case 2:
		q.Environment = &env
	}
	expected := []*exploration.SavedCapture{}
	seen := map[string]bool{}
	for _, wire := range m.first {
		var page exploration.CapturePage
		if err := json.Unmarshal([]byte(wire), &page); err != nil {
			panic(err)
		}
		c := &exploration.SavedCapture{CaptureReceipt: page.Receipt}
		if !seen[c.ID] && strings.Contains(strings.ToLower(c.Title), q.Title) && (q.Environment == nil || *q.Environment == "prod") {
			expected = append(expected, c)
			seen[c.ID] = true
		}
	}
	slices.SortFunc(expected, func(a, b *exploration.SavedCapture) int {
		x, _ := time.Parse(time.RFC3339Nano, a.CapturedAt)
		y, _ := time.Parse(time.RFC3339Nano, b.CapturedAt)
		n := x.Compare(y)
		if q.Order == "newest" {
			n = -n
		}
		if n == 0 {
			return strings.Compare(a.ID, b.ID)
		}
		return n
	})
	actual := []string{}
	cursor := ""
	for {
		p, e := m.store.Search(context.Background(), q, cursor)
		if e != nil {
			panic(e)
		}
		for _, r := range p.Items {
			actual = append(actual, r.ID)
		}
		if !p.HasMore {
			break
		}
		cursor = p.NextCursor
	}
	if len(actual) != len(expected) {
		panic("search changed catalog membership")
	}
	for i, c := range expected {
		if actual[i] != c.ID {
			panic("search order or pagination diverged")
		}
	}
}
func TestHegelCaptureStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		m := &captureMachine{store: exploration.CaptureStore{Dir: t.TempDir()}, count: 1005, first: map[string]string{}}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			b, _ := io.ReadAll(r.Body)
			switch {
			case strings.Contains(string(b), "SELECT DISTINCT"):
			case strings.Contains(string(b), "toString(count()) AS count"):
				fmt.Fprintf(w, "{\"count\":\"%d\"}\n", m.count)
			default:
				for i := 0; i < min(m.count, 1001); i++ {
					fmt.Fprintf(w, "{\"timestamp_ns\":\"%d\",\"duration_ns\":\"0\",\"source_revision\":%d}\n", i+1, m.revision)
				}
			}
		}))
		defer server.Close()
		m.source = exploration.Source{Endpoint: server.URL}
		hegel.RunStateful(tc, m, hegel.WithStatefulStepCount(15), hegel.WithAlwaysCheckInvariants("InvariantStableEvidence"))
	}, options(60)...)
}

type captureLedger struct {
	Original            exploration.CapturePage `json:"original"`
	Second              exploration.CapturePage `json:"second"`
	Refreshed           exploration.CapturePage `json:"refreshed"`
	Reopened            exploration.CapturePage `json:"reopened"`
	CrossRevisionStatus int                     `json:"crossRevisionStatus"`
}

func TestQuintCaptureLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_CAPTURES_LEDGER")
	if name == "" {
		t.Skip("live ledger opt-in")
	}
	b, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l captureLedger
	if e = json.Unmarshal(b, &l); e != nil {
		t.Fatal(e)
	}
	check := func(l captureLedger) (bool, string) {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("captures"))
		a := l.Original
		s := l.Second
		n := l.Refreshed
		valid := a.Explore != nil && a.Receipt.Question.Kind == "explore" && a.RetainedRecords == 1000 && a.RecordsLimited && len(a.Explore.Rows) == 100 && a.Explore.Count == "1205" && a.Explore.Question.Offset == 0 && a.Receipt.Question.Explore.Offset == 0
		rec.Record(context.Background(), "save", nil, "obs.coherent", valid)
		expected := *a.Receipt.Question.Explore
		expected.Offset = 100
		paged := s.Receipt.ID == a.Receipt.ID && reflect.DeepEqual(s.Explore.Question, expected) && s.Explore.Basis == a.Explore.Basis && s.Explore.Count == a.Explore.Count && s.RetainedRecords == a.RetainedRecords && s.RecordsLimited == a.RecordsLimited && len(s.Explore.Rows) == 100 && s.PreviousCursor != "" && s.NextCursor != ""
		rec.Record(context.Background(), "page", nil, "obs.coherent", paged)
		refreshed := n.Receipt.ID != a.Receipt.ID && n.Receipt.Parent == a.Receipt.ID && reflect.DeepEqual(n.Receipt.Question, a.Receipt.Question) && n.Explore.Count == "1206" && n.Explore.Basis != a.Explore.Basis && l.CrossRevisionStatus == 400
		rec.Record(context.Background(), "refresh", nil, "obs.coherent", refreshed)
		rec.Record(context.Background(), "reopen", nil, "obs.coherent", reflect.DeepEqual(a, l.Reopened))
		checker, e := validate.NewChecker("model/captures.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
		if e != nil {
			t.Fatal(e)
		}
		report, e := checker.Check(context.Background(), sink.Steps())
		if e != nil {
			t.Fatal(e)
		}
		return report.OK(), report.Summary()
	}
	ok, summary := check(l)
	artifact("captures-live.txt", []byte(summary))
	if !ok {
		t.Fatal(summary)
	}
	for i := 0; i < 4; i++ {
		data, _ := json.Marshal(l)
		var bad captureLedger
		json.Unmarshal(data, &bad)
		switch i {
		case 0:
			bad.Reopened.Explore.Count = "1206"
		case 1:
			bad.Second.Explore.Basis = "shifted"
		case 2:
			bad.Refreshed.Receipt.Question.Explore.Environment = "other"
		case 3:
			bad.CrossRevisionStatus = 200
		}
		ok, summary := check(bad)
		artifact(fmt.Sprintf("captures-negative-%d.txt", i), []byte(summary))
		if ok {
			t.Fatal("negative control accepted", i)
		}
	}
}

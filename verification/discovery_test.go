package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
	"hegel.dev/go/hegel"
	"os"
	"reflect"
	"strconv"
	"testing"
)

type contextMachine struct {
	center, radius int
	offset         int
}

func (m *contextMachine) RuleAnchor(tc hegel.TestCase) {
	m.center = hegel.Draw(tc, hegel.Integers(1000, 99999))
}
func (m *contextMachine) RuleRadius(tc hegel.TestCase) {
	m.radius = hegel.Draw(tc, hegel.Integers(1000, 300000))
}
func (m *contextMachine) RulePage(tc hegel.TestCase) {
	m.offset = hegel.Draw(tc, hegel.Integers(0, 10)) * 100
}
func (m *contextMachine) InvariantBoundedContext(tc hegel.TestCase) {
	origin := exploration.Query{Signal: "logs", Environment: "prod", FromMS: 1000, ToMS: 100000, Limit: 1, Where: &exploration.Predicate{Field: &exploration.Field{Kind: "column", Key: "body"}, Op: "contains", Value: "original"}}
	anchor := strconv.FormatInt(int64(m.center)*1_000_000+123, 10)
	q, e := exploration.SurroundingQuestion(origin, anchor, "worker", int64(m.radius), m.offset)
	if e != nil {
		panic(e)
	}
	if q.FromMS < origin.FromMS || q.ToMS > origin.ToMS || q.FromMS > int64(m.center) || q.ToMS <= int64(m.center) || q.Environment != origin.Environment || q.Where.Field.Key != "service" || q.Where.Value != "worker" || q.Where.Op != "eq" || q.Offset != m.offset {
		panic("context leaked scope or search predicates")
	}
	p, e := exploration.Compile(q)
	if e != nil {
		panic(e)
	}
	q.Offset = 0
	first, _ := exploration.Compile(q)
	if p.Count != first.Count || !reflect.DeepEqual(p.Parameters, first.Parameters) {
		panic("page changed context count")
	}
}
func TestHegelDiscoveryContextStateSwarm(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		hegel.RunStateful(tc, &contextMachine{center: 1500, radius: 30000}, hegel.WithStatefulStepCount(30), hegel.WithAlwaysCheckInvariants("InvariantBoundedContext"))
	}, options(100)...)
}
func TestHegelDurationBucketPartition(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		n := hegel.Draw(tc, hegel.Integers(0, 2000000000))
		bounds := []uint64{0, 1000000, 5000000, 10000000, 50000000, 100000000, 500000000, 1000000000}
		labels := []string{"[0,1) ms", "[1,5) ms", "[5,10) ms", "[10,50) ms", "[50,100) ms", "[100,500) ms", "[500,1000) ms", "[1000,+inf) ms"}
		members := 0
		for i, lo := range bounds {
			if uint64(n) >= lo && (i+1 == len(bounds) || uint64(n) < bounds[i+1]) {
				members++
				if exploration.DurationBucket(uint64(n)) != labels[i] {
					panic("wrong duration bucket")
				}
			}
		}
		if members != 1 {
			panic("duration missing or counted twice")
		}
	}, options(200)...)
}

type discoveryLedger struct {
	Origin exploration.Query `json:"origin"`
	Anchor struct {
		Timestamp string `json:"timestamp_ns"`
		Service   string `json:"ServiceName"`
	} `json:"anchor"`
	Context exploration.Query `json:"context"`
	Trace   exploration.Query `json:"trace"`
	TraceID string            `json:"traceID"`
	Buckets exploration.Query `json:"buckets"`
	Package exploration.Query `json:"package"`
}

func TestQuintDiscoveryLiveLedger(t *testing.T) {
	name := os.Getenv("GOO_DISCOVERY_LEDGER")
	if name == "" {
		t.Skip("live browser ledger is opt-in")
	}
	data, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	var l discoveryLedger
	if e = json.Unmarshal(data, &l); e != nil {
		t.Fatal(e)
	}
	check := func(l discoveryLedger) (bool, string) {
		sink := &qobs.MemorySink{}
		rec := qobs.NewRecorder(sink, qobs.WithActor("discovery"), qobs.WithProcess("live-browser"))
		expected, e := exploration.SurroundingQuestion(l.Origin, l.Anchor.Timestamp, l.Anchor.Service, 30000, 0)
		bounded := e == nil && reflect.DeepEqual(expected, l.Context)
		rec.Record(context.Background(), "surround", nil, "obs.bounded", bounded, "obs.reusable", true)
		trace, e := exploration.TraceQuestion(l.Buckets.Environment, l.Buckets.FromMS, l.Buckets.ToMS, l.TraceID, "", "spans", 0)
		trace.GroupBy = []exploration.Field{{Kind: "column", Key: "service"}, {Kind: "column", Key: "name"}}
		trace.Summary = "span-duration"
		trace.SpanStatus = true
		rec.Record(context.Background(), "trace", nil, "obs.bounded", e == nil && reflect.DeepEqual(trace, l.Trace), "obs.reusable", true)
		reusable := l.Package.Signal == l.Buckets.Signal && l.Package.Environment == l.Buckets.Environment && reflect.DeepEqual(l.Package.Where, l.Buckets.Where) && reflect.DeepEqual(l.Package.GroupBy, l.Buckets.GroupBy)
		rec.Record(context.Background(), "save", nil, "obs.bounded", true, "obs.reusable", reusable)
		checker, e := validate.NewChecker("model/discovery.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
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
	t.Log(summary)
	artifact("discovery-live.txt", []byte(summary))
	if !ok {
		t.Fatal(summary)
	}
	for i := 0; i < 3; i++ {
		bad := l
		switch i {
		case 0:
			bad.Context.FromMS--
		case 1:
			bad.Trace.Environment = "other"
		case 2:
			bad.Package.GroupBy = nil
		}
		ok, summary := check(bad)
		t.Logf("negative control %d: %s", i, summary)
		artifact(fmt.Sprintf("discovery-negative-%d.txt", i), []byte(summary))
		if ok {
			t.Fatal("forged context or package accepted", i)
		}
	}
}

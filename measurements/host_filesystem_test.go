package measurements

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func filesystemQuestion(metric string) Question {
	q := cpuTimeQuestion()
	q.Definition.Name = "filesystem"
	q.Definition.Kind = "stock"
	q.Definition.CounterDomain = ""
	q.Definition.StockDomain = "nonnegative-integer"
	s := *q.Definition.Source.ExactStream
	s.Name = metric
	s.Unit = "By"
	s.Attributes = map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data"}
	if metric == "system.filesystem.usage" {
		s.Attributes["system.filesystem.state"] = "used"
	}
	q.Definition.Source.Metric = s.Name
	q.Definition.Source.Unit = s.Unit
	q.Definition.Source.ExactStream = &s
	return q
}
func filesystemRow(q Question, at, value string) json.RawMessage {
	return json.RawMessage(strings.Replace(string(cpuTimeRow(q, at, value)), `"monotonic":true`, `"monotonic":false`, 1))
}
func TestFilesystemStocksRetainDecreasesAndRequireOriginalAdmission(t *testing.T) {
	for _, metric := range []string{"system.filesystem.usage", "system.filesystem.limit"} {
		q := filesystemQuestion(metric)
		rows := []json.RawMessage{filesystemRow(q, "1000000000", "4096"), filesystemRow(q, "2000000000", "2048")}
		r, e := Evaluate(q, rows, time.Unix(3, 0), "", nil)
		if e != nil || len(r.Stocks.Series) != 1 || *r.Stocks.Series[0].Samples[1].Value != 2048 {
			t.Fatal(r, e)
		}
		for _, bad := range []string{"-1", "1.00000000000000000001", "9007199254740992"} {
			if _, e := Evaluate(q, []json.RawMessage{filesystemRow(q, "1000000000", bad)}, time.Unix(3, 0), "", nil); e == nil {
				t.Fatal("accepted invalid bytes", bad)
			}
		}
		for _, change := range [][2]string{{`"flags":0`, `"flags":null`}, {`"flags":0`, `"flags":0.000000000000001`}, {`"temporality":2`, `"temporality":1`}, {`"temporality":2`, `"temporality":2.0000000000000001`}, {`"monotonic":false`, `"monotonic":null`}, {`"monotonic":false`, `"monotonic":true`}, {`"attributes":{`, `"attributes":null,"ignored":{`}} {
			raw := strings.Replace(string(rows[0]), change[0], change[1], 1)
			if _, e := Evaluate(q, []json.RawMessage{json.RawMessage(raw)}, time.Unix(3, 0), "", nil); e == nil {
				t.Fatal("accepted malformed stock metadata", change)
			}
		}
		missing := strings.Replace(string(rows[1]), `"flags":0`, `"flags":1`, 1)
		missing = strings.Replace(missing, `"value":"2048"`, `"value":"-1"`, 1)
		r, e = Evaluate(q, []json.RawMessage{rows[0], json.RawMessage(missing)}, time.Unix(3, 0), "", nil)
		if e != nil || r.Stocks.Series[0].Samples[1].Value != nil || r.Stocks.Series[0].Samples[1].State != "unknown" {
			t.Fatal("missing value interpreted", r, e)
		}
	}
}
func TestFilesystemPlanKeepsDefaultsAndReportedRolesExplicit(t *testing.T) {
	rows := hostTestRows()[:5]
	attrs := map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data", "system.filesystem.state": "used"}
	for _, metric := range []string{"system.filesystem.usage", "system.filesystem.limit"} {
		a := map[string]string{}
		for k, v := range attrs {
			a[k] = v
		}
		if metric == "system.filesystem.limit" {
			delete(a, "system.filesystem.state")
		}
		rows = append(rows, hostTestRow(metric, "counter", "By", a, nil))
	}
	inv, e := discoverHostRows(hostTestQuestion(), rows)
	if e != nil {
		t.Fatal(e)
	}
	plan, e := PlanHostDashboard(HostDashboardRequest{Question: inv.Question, Host: inv.Hosts[0]})
	if e != nil || len(plan.Widgets) != 2 {
		t.Fatal(plan, e)
	}
	found := 0
	for _, w := range plan.Widgets {
		if HostFilesystemDefinition(w.Question.Definition) {
			t.Fatal("filesystem became default")
		}
	}
	for _, w := range plan.Candidates {
		if HostFilesystemDefinition(w.Question.Definition) {
			found++
		}
	}
	if found != 2 {
		t.Fatal("filesystem pair missing", found)
	}
	for _, bad := range []struct {
		metric string
		attrs  map[string]string
	}{{"system.filesystem.usage", map[string]string{"system.device": "/dev/sda", "system.filesystem.mountpoint": "/data", "system.filesystem.state": "free"}}, {"system.filesystem.limit", attrs}, {"system.filesystem.usage", map[string]string{"system.filesystem.state": "used"}}} {
		inv, e := discoverHostRows(hostTestQuestion(), []json.RawMessage{hostTestRow(bad.metric, "counter", "By", bad.attrs, nil)})
		if e != nil {
			t.Fatal(e)
		}
		p, e := PlanHostDashboard(HostDashboardRequest{Question: inv.Question, Host: inv.Hosts[0]})
		if e != nil || len(p.Candidates) != 0 || len(p.Widgets) != 0 {
			t.Fatal("guessed filesystem role", p, e)
		}
	}
}

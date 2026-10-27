package measurements

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
)

func containerHistoryFixture(t *testing.T) ContainerResourceHistoryRequest {
	memory := containerMemoryFixture(t)
	cpu := containerCPUFixture(t)
	return ContainerResourceHistoryRequest{Parent: memory.Parent, Memory: &ContainerMemoryHistoryResults{memory.WorkingSet, memory.Request, memory.Limit}, CPU: &ContainerCPUHistoryResults{cpu.Usage, cpu.Request, cpu.Limit}}
}
func TestContainerResourceHistoryEveryTimestampAndQuality(t *testing.T) {
	in := containerHistoryFixture(t)
	in.Memory.WorkingSet.Question.Definition.MaxGapMS = 20000
	replacePoolRows(t, in.Memory.WorkingSet, func(rows []metricRow) []metricRow {
		first, duplicate, latest, missing := rows[0], rows[0], rows[0], rows[0]
		first.EndNS = ns(110000)
		first.Value = "10"
		duplicate.Value = "50.00000000000000001"
		latest.EndNS = ns(150000)
		latest.Value = "100"
		missing.EndNS = ns(140000)
		missing.Flags = 1
		return []metricRow{latest, rows[0], first, duplicate, missing}
	})
	before, _ := json.Marshal(in)
	history, e := ComposeContainerResourceHistory(in)
	if e != nil {
		t.Fatal(e)
	}
	if len(history.Series) != 6 || history.Series[0].Key != "memory/working-set" || history.Series[3].Key != "cpu/usage" {
		t.Fatal(history)
	}
	s := history.Series[0]
	if s.MaxGapNS != "20000000000" || s.FromNS != ns(in.Parent.FromMS) || s.ToNS != ns(in.Parent.ToMS) || len(s.Points) != 4 {
		t.Fatal(s)
	}
	if s.Points[0].AtNS != ns(110000) || s.Points[0].State != "stale" || s.Points[0].Value != "10" || s.Points[0].RawRowIndices[0] != 2 {
		t.Fatal(s.Points[0])
	}
	invalid := s.Points[1]
	if invalid.State != "invalid" || invalid.Value != "" || invalid.ObservationState != "stale" || invalid.ObservationReason != "older than declared freshness" || invalid.Reason != "invalid-byte-domain" || len(invalid.RawRowIndices) != 2 || invalid.RawRowIndices[0] != 1 || invalid.RawRowIndices[1] != 3 {
		t.Fatal(invalid)
	}
	if s.Points[2].State != "unknown" || s.Points[2].ObservationReason != "no recorded value" || s.Points[2].Value != "" || s.Points[3].Value != "100" {
		t.Fatal(s.Points)
	}
	history.Parent.Resource["container.id"] = "mutated"
	history.Series[0].Question.Definition.Source.ExactStream.Resource["k8s.pod.uid"] = "mutated"
	after, _ := json.Marshal(in)
	if string(before) != string(after) {
		t.Fatal("held evidence mutated")
	}
	if file := os.Getenv("GOO_CONTAINER_HISTORY_FIXTURE"); file != "" {
		view, e := ComposeContainerResourceHistory(in)
		if e != nil {
			t.Fatal(e)
		}
		b, _ := json.MarshalIndent(struct {
			Request  ContainerResourceHistoryRequest `json:"request"`
			Response *ContainerResourceHistory       `json:"response"`
		}{in, view}, "", "  ")
		if e = os.WriteFile(file, append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
}
func TestContainerResourceHistoryRawCPUAndMissingFamilies(t *testing.T) {
	in := containerHistoryFixture(t)
	in.Memory = nil
	replacePoolRows(t, in.CPU.Usage, func(rows []metricRow) []metricRow {
		first, duplicate, invalid, latest := rows[0], rows[0], rows[0], rows[0]
		first.EndNS = ns(110000)
		first.Value = "1"
		duplicate.EndNS = first.EndNS
		duplicate.Value = "1.00000000000000001"
		invalid.EndNS = ns(130000)
		invalid.Value = "-0.25"
		latest.EndNS = ns(150000)
		latest.Value = "0.000000000000000001"
		return []metricRow{latest, first, duplicate, invalid}
	})
	in.CPU.Request = nil
	history, e := ComposeContainerResourceHistory(in)
	if e != nil || len(history.Series) != 3 || history.Series[0].Points[0].State != "invalid" || history.Series[0].Points[1].Reason != "invalid-core-domain" || history.Series[0].Points[2].Value != "0.000000000000000001" || history.Series[1].State != "unavailable" {
		t.Fatal(history, e)
	}
	if _, e = ComposeContainerResourceHistory(ContainerResourceHistoryRequest{Parent: in.Parent}); e == nil {
		t.Fatal("no family admitted")
	}
	empty, e := ComposeContainerResourceHistory(ContainerResourceHistoryRequest{Parent: in.Parent, CPU: &ContainerCPUHistoryResults{}})
	if e != nil || len(empty.Series) != 3 || empty.Series[0].Points == nil {
		t.Fatal(empty, e)
	}
	wrong := containerHistoryFixture(t)
	wrong.CPU.Usage.Gauges.Series[0].Samples[0].Value = nil
	if _, e = ComposeContainerResourceHistory(wrong); e == nil {
		t.Fatal("forged held evidence admitted")
	}
	wrong = containerHistoryFixture(t)
	wrong.CPU.Usage = wrong.Memory.WorkingSet
	if _, e = ComposeContainerResourceHistory(wrong); e == nil {
		t.Fatal("memory signal admitted as CPU")
	}
}
func TestContainerResourceHistoryPointAndInputLimits(t *testing.T) {
	in := containerHistoryFixture(t)
	in.Memory.Request = nil
	in.Memory.Limit = nil
	in.CPU.Request = nil
	in.CPU.Limit = nil
	replacePoolRows(t, in.Memory.WorkingSet, func(rows []metricRow) []metricRow {
		out := make([]metricRow, ContainerResourceHistoryPointLimit)
		for i := range out {
			out[i] = rows[0]
			out[i].EndNS = strconv.FormatInt(120000000000+int64(i), 10)
		}
		return out
	})
	if _, e := ComposeContainerResourceHistory(in); e == nil || !strings.Contains(e.Error(), "10000 total points") {
		t.Fatal("point budget", e)
	}
	in.CPU = nil
	history, e := ComposeContainerResourceHistory(in)
	if e != nil || len(history.Series[0].Points) != ContainerResourceHistoryPointLimit {
		t.Fatal("exact point boundary", e)
	}
	in = containerHistoryFixture(t)
	in.Memory.WorkingSet.Rows[0] = json.RawMessage(strings.Repeat("x", (8<<20)+1))
	if _, e = ComposeContainerResourceHistory(in); e == nil || !strings.Contains(e.Error(), "8 MiB") {
		t.Fatal("raw budget before replay", e)
	}
}

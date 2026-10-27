package measurements

import (
	"encoding/json"
	"fmt"
	"math/big"
)

const ContainerResourceHistoryVersion = "goo/container-resources-history/v1"
const ContainerResourceHistoryPointLimit = 10000

type ContainerMemoryHistoryResults struct {
	WorkingSet *Result `json:"workingSet,omitempty"`
	Request    *Result `json:"request,omitempty"`
	Limit      *Result `json:"limit,omitempty"`
}
type ContainerCPUHistoryResults struct {
	Usage   *Result `json:"usage,omitempty"`
	Request *Result `json:"request,omitempty"`
	Limit   *Result `json:"limit,omitempty"`
}
type ContainerResourceHistoryRequest struct {
	Parent PodMetricDiscoveryQuestion     `json:"parent"`
	Memory *ContainerMemoryHistoryResults `json:"memory,omitempty"`
	CPU    *ContainerCPUHistoryResults    `json:"cpu,omitempty"`
}
type ContainerResourceHistoryPoint struct {
	AtNS              string `json:"atNS"`
	Value             string `json:"value,omitempty"`
	State             string `json:"state"`
	ObservationState  string `json:"observationState"`
	Reason            string `json:"reason,omitempty"`
	ObservationReason string `json:"observationReason,omitempty"`
	RawRowIndices     []int  `json:"rawRowIndices"`
}
type ContainerResourceHistorySeries struct {
	Key      string                          `json:"key"`
	Family   string                          `json:"family"`
	Role     string                          `json:"role"`
	Title    string                          `json:"title"`
	Unit     string                          `json:"unit"`
	State    string                          `json:"state"`
	Reason   string                          `json:"reason,omitempty"`
	Question *Question                       `json:"question,omitempty"`
	ReadAt   string                          `json:"readAt,omitempty"`
	Basis    string                          `json:"basis,omitempty"`
	FromNS   string                          `json:"fromNS,omitempty"`
	ToNS     string                          `json:"toNS,omitempty"`
	MaxGapNS string                          `json:"maxGapNS,omitempty"`
	Points   []ContainerResourceHistoryPoint `json:"points"`
}
type ContainerResourceHistory struct {
	Version  string                           `json:"version"`
	Parent   PodMetricDiscoveryQuestion       `json:"parent"`
	Series   []ContainerResourceHistorySeries `json:"series"`
	Warnings []string                         `json:"warnings"`
}

// ComposeContainerResourceHistory projects only held replay-validated evidence.
// Quality is evaluated at the original fixed window end; observations do not
// establish continuity, a configuration-change timestamp or an atomic snapshot.
func ComposeContainerResourceHistory(input ContainerResourceHistoryRequest) (*ContainerResourceHistory, error) {
	if input.Memory == nil && input.CPU == nil {
		return nil, fmt.Errorf("container resource history requires a supplied family")
	}
	results := []*Result{}
	if input.Memory != nil {
		results = append(results, input.Memory.WorkingSet, input.Memory.Request, input.Memory.Limit)
	}
	if input.CPU != nil {
		results = append(results, input.CPU.Usage, input.CPU.Request, input.CPU.Limit)
	}
	in, e := cloneContainerResourceEvidence(input, input.Parent, results, "container resource history")
	if e != nil {
		return nil, e
	}
	out := &ContainerResourceHistory{Version: ContainerResourceHistoryVersion, Parent: in.Parent, Series: []ContainerResourceHistorySeries{}, Warnings: []string{
		"Every point is a reported observation certified from its original raw rows; absence is not zero. No interpolation or held-forward request or limit is supplied.",
		"Observation quality and declared freshness are evaluated at the original fixed question end, not at the instant the observation was recorded. Sparse observations do not establish continuity or an outage.",
		"Different reported values establish only different observations, not a configuration-change event or its time. No percentage history, atomic snapshot, health or causal relationship is inferred.",
		"Raw row indices refer to the corresponding supplied result in its original order. Each series retains independent source read time and evidence basis; no new custody or source read is created.",
	}}
	type family struct {
		key, unit string
		entries   []containerResourceEntry
		policy    containerResourcePolicy
	}
	families := []family{}
	if in.Memory != nil {
		families = append(families, family{"memory", "By", []containerResourceEntry{{"working-set", in.Memory.WorkingSet}, {"request", in.Memory.Request}, {"limit", in.Memory.Limit}}, containerMemoryPolicy()})
	}
	if in.CPU != nil {
		families = append(families, family{"cpu", "{cpu}", []containerResourceEntry{{"usage", in.CPU.Usage}, {"request", in.CPU.Request}, {"limit", in.CPU.Limit}}, containerCPUPolicy()})
	}
	total := 0
	for _, f := range families {
		// Admission shares the existing exact role, window, context and stream rules.
		// This latest-only pass is linear; the timestamp index below also visits each
		// raw row once, instead of rescanning the complete result for every sample.
		signals, _, e := composeContainerResourceSignals(in.Parent, f.entries, f.policy)
		if e != nil {
			return nil, e
		}
		for index, entry := range f.entries {
			signal := signals[index]
			series := ContainerResourceHistorySeries{Key: f.key + "/" + entry.role, Family: f.key, Role: entry.role, Title: signal.Title, Unit: f.unit, State: "unavailable", Reason: "not-supplied", Points: []ContainerResourceHistoryPoint{}}
			result := entry.result
			if result == nil {
				out.Series = append(out.Series, series)
				continue
			}
			series.Question = &result.Question
			series.ReadAt = result.ReadAt
			series.Basis = result.Basis
			series.FromNS, series.ToNS = result.Question.BoundsNS()
			series.MaxGapNS = ns(result.Question.Definition.MaxGapMS)
			series.State = "no_data"
			series.Reason = "no-observations"
			if len(result.Gauges.Series) == 0 {
				out.Series = append(out.Series, series)
				continue
			}
			samples := result.Gauges.Series[0].Samples
			if len(samples) > ContainerResourceHistoryPointLimit-total {
				return nil, fmt.Errorf("container resource history exceeds 10000 total points; narrow the held window")
			}
			total += len(samples)
			rows := make([]metricRow, len(result.Rows))
			byTime := map[string][]int{}
			for i, raw := range result.Rows {
				if e = json.Unmarshal(raw, &rows[i]); e != nil {
					return nil, fmt.Errorf("invalid container resource history raw row")
				}
				byTime[rows[i].EndNS] = append(byTime[rows[i].EndNS], i)
			}
			for _, sample := range samples {
				indices := byTime[sample.AtNS]
				if len(indices) == 0 {
					return nil, fmt.Errorf("container resource history observation has no original raw row")
				}
				point := ContainerResourceHistoryPoint{AtNS: sample.AtNS, State: sample.State, ObservationState: sample.State, Reason: sample.Reason, ObservationReason: sample.Reason, RawRowIndices: indices}
				if sample.Value != nil {
					var exact *big.Rat
					valid := true
					for _, i := range indices {
						row := rows[i]
						value, ok := f.policy.parse(row.Value)
						if row.Flags&1 != 0 || !ok || exact != nil && exact.Cmp(value) != 0 {
							valid = false
							break
						}
						exact = value
					}
					if valid && exact != nil {
						point.Value = f.policy.format(exact)
					} else {
						point.State = "invalid"
						point.Reason = f.policy.invalidReason
					}
				}
				series.Points = append(series.Points, point)
			}
			if len(series.Points) > 0 {
				series.State = "observations"
				series.Reason = ""
			}
			out.Series = append(out.Series, series)
		}
	}
	encoded, e := json.Marshal(out)
	if e != nil {
		return nil, e
	}
	if len(encoded) > 8<<20 {
		return nil, fmt.Errorf("container resource history output exceeds 8 MiB; narrow the held window")
	}
	return out, nil
}

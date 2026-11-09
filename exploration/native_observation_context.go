package exploration

import (
	"fmt"
	"strconv"

	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
)

// NativeObservationRadiusNS is a fixed lookaround, not gauge coverage. The upper
// endpoint is exclusive; both endpoints are clipped to the held native question.
const NativeObservationRadiusNS uint64 = 60_000_000_000

// NativeObservationContext associates ordinary records with one reported gauge
// timestamp. It does not infer causality, an exemplar or an observation interval.
type NativeObservationContext struct {
	Question Query  `json:"question"`
	FromNS   string `json:"fromNS"`
	ToNS     string `json:"toNS"`
	Grain    string `json:"grain"`
	Relation string `json:"relation"`
	AtNS     string `json:"atNS"`
	RadiusNS string `json:"radiusNS"`
}

// PlanNativeObservationContext checks retained gauge evidence without source or
// custody I/O. Stale and unknown observations retain their reported timestamps;
// neither a numeric magnitude nor the current clock supplies missing identity.
func PlanNativeObservationContext(result *measurements.Result, stream metrics.Stream, signal, atNS string) (*NativeObservationContext, error) {
	if result == nil {
		return nil, fmt.Errorf("observation context requires retained gauge evidence")
	}
	parent := result.Question
	d := parent.Definition
	if d.Kind != "gauge" || d.Grain != "per-stream" || d.Source.Resource != nil {
		return nil, fmt.Errorf("observation context requires a service-bound per-stream gauge")
	}
	if signal != "logs" && signal != "spans" {
		return nil, fmt.Errorf("observation context requires logs or spans")
	}
	if old := d.Source.ExactStream; old != nil && !nativeContextSameStream(*old, stream) {
		return nil, fmt.Errorf("observation stream differs from pinned native stream")
	}
	checked := parent
	checked.Definition.Source.ExactStream = &stream
	if err := checked.Validate(); err != nil {
		return nil, err
	}
	if err := measurements.ValidateResult(result); err != nil {
		return nil, err
	}
	if result.Gauges == nil {
		return nil, fmt.Errorf("observation context requires retained gauge samples")
	}
	at, err := strconv.ParseUint(atNS, 10, 64)
	if err != nil || strconv.FormatUint(at, 10) != atNS {
		return nil, fmt.Errorf("observation context requires a canonical reported timestamp")
	}
	seriesCount, sampleCount := 0, 0
	for _, series := range result.Gauges.Series {
		if nativeContextSameStream(series.Stream, stream) {
			seriesCount++
			for _, sample := range series.Samples {
				if sample.AtNS == atNS {
					sampleCount++
				}
			}
		}
	}
	if seriesCount != 1 || sampleCount != 1 {
		return nil, fmt.Errorf("observation must identify one retained stream sample")
	}
	fromText, toText := parent.BoundsNS()
	lo, _ := strconv.ParseUint(fromText, 10, 64)
	hi, _ := strconv.ParseUint(toText, 10, 64)
	if at < lo || at >= hi {
		return nil, fmt.Errorf("observation timestamp is outside the native question")
	}
	// Compare distances before subtracting or adding: canonical uint64 bounds can
	// be close to zero, and no lookaround may escape the original fixed window.
	from, to := lo, hi
	if at-lo > NativeObservationRadiusNS {
		from = at - NativeObservationRadiusNS
	}
	if hi-at > NativeObservationRadiusNS {
		to = at + NativeObservationRadiusNS
	}
	context, err := nativeRecordContextQuery(parent, stream, signal, from, to, "observation-context")
	if err != nil {
		return nil, err
	}
	return &NativeObservationContext{Question: context.Question, FromNS: context.FromNS, ToNS: context.ToNS, Grain: context.Grain, Relation: "observation-context", AtNS: atNS, RadiusNS: strconv.FormatUint(NativeObservationRadiusNS, 10)}, nil
}

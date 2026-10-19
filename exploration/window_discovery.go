package exploration

import (
	"fmt"
	"github.com/casselc/goo/metrics"
)

type CapturedWindowDiscovery struct {
	Receipt CaptureReceipt          `json:"receipt"`
	ReadAt  string                  `json:"readAt"`
	Support metrics.WindowDiscovery `json:"support"`
}

// DiscoverCapturedWindows is source-free and does not alter the original capture.
func DiscoverCapturedWindows(c *SavedCapture) (*CapturedWindowDiscovery, error) {
	if c == nil {
		return nil, fmt.Errorf("stored native capture required")
	}
	if e := validateCapture(c); e != nil {
		return nil, e
	}
	if c.Measurement == nil {
		return nil, fmt.Errorf("window discovery requires a counter or explicit histogram capture")
	}
	r := c.Measurement
	out := &CapturedWindowDiscovery{Receipt: c.CaptureReceipt, ReadAt: r.ReadAt}
	switch r.Question.Definition.Kind {
	case "counter":
		out.Support = metrics.DiscoverCounterWindows(r.Counters)
	case "histogram":
		out.Support = metrics.DiscoverHistogramWindows(r.Histograms)
	default:
		return nil, fmt.Errorf("window discovery requires a counter or explicit histogram capture")
	}
	return out, nil
}

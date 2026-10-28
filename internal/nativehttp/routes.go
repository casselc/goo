package nativehttp

import (
	"context"
	"net/http"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/measurements"
)

// RegisterMeasurements binds one configured source, including its CustodyDir.
// Only Run may record lifecycle custody; discovery and supplied-data plans do not.
// Definition catalogs and package publication are composed separately by cmd/goo.
func RegisterMeasurements(mux *http.ServeMux, source measurements.Source) {
	registerRead(mux, "/metric-discovery", measurements.DiscoveryQuestion.Validate, source.Discover)
	registerRead(mux, "/measurements", measurements.Question.Validate, func(ctx context.Context, q measurements.Question) (*measurements.Result, error) {
		return source.Run(ctx, q, nil)
	})
	registerRead(mux, "/host-discovery", measurements.HostDiscoveryQuestion.Validate, source.DiscoverHosts)
	registerRead(mux, "/pod-metric-discovery", measurements.PodMetricDiscoveryQuestion.Validate, source.DiscoverPodMetrics)
	registerPlan(mux, "/host-dashboard/plan", measurements.PlanHostDashboard)
	registerPlan(mux, "/pod-dashboard/plan", measurements.PlanPodDashboard)
}

// RegisterPodDiscovery uses the record adapter's cross-source resource inventory,
// not the native metrics source. Its admission and limits remain in this family.
func RegisterPodDiscovery(mux *http.ServeMux, source exploration.Source) {
	registerRead(mux, "/pod-discovery", exploration.PodDiscoveryQuestion.Validate, source.DiscoverPods)
}

func registerRead[Q, R any](mux *http.ServeMux, path string, validate func(Q) error, read func(context.Context, Q) (R, error)) {
	policy, ok := PolicyFor("POST", path)
	if !ok {
		panic("nativehttp: missing source-read policy for " + path)
	}
	mux.HandleFunc("POST "+path, queryhttp.ReadHandler(policy, validate, read))
}
func registerPlan[Q, R any](mux *http.ServeMux, path string, plan func(Q) (R, error)) {
	policy, ok := PolicyFor("POST", path)
	if !ok {
		panic("nativehttp: missing planner policy for " + path)
	}
	mux.HandleFunc("POST "+path, queryhttp.PlanHandler(policy, plan))
}

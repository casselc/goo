package main

import (
	"net/http"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/recordhttp"
)

var resultSnapshots = &exploration.ResultSnapshots{}

func registerExploration(mux *http.ServeMux, source exploration.Source, banks ...*exploration.TailMembershipSnapshots) {
	recordhttp.Register(mux, source, resultSnapshots)
	registerPodDiscovery(mux, source)
	registerRecipes(mux)
	registerPoolSignals(mux)
	registerPoolView(mux)
	registerPopulationExplanations(mux, source)
	registerTail(mux, source, banks...)
}

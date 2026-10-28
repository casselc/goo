package main

import (
	"net/http"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/internal/recordhttp"
)

func registerExploration(mux *http.ServeMux, source exploration.Source, resultSnapshots *exploration.ResultSnapshots, tailMembership *exploration.TailMembershipSnapshots) {
	recordhttp.Register(mux, source, resultSnapshots)
	nativehttp.RegisterPodDiscovery(mux, source)
	registerRecipes(mux)
	registerPoolSignals(mux)
	registerPoolView(mux)
	registerPopulationExplanations(mux, source, resultSnapshots)
	registerTail(mux, source, tailMembership)
}

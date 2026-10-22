package main

import (
	"net/http"
	"os"

	"github.com/casselc/goo/exploration"
)

// explorationSourceFromEnvironment resolves record-source configuration at startup.
// Registration receives a source value; query and SSE owners keep their lifetimes.
func explorationSourceFromEnvironment(client *http.Client) exploration.Source {
	endpoint := os.Getenv("CLICKHOUSE_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8123"
	}
	return exploration.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
}

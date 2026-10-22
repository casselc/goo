package main

import (
	"net/http"
	"os"

	"github.com/casselc/goo/measurements"
)

// nativeMeasurementSourceFromEnvironment is startup wiring, not a request hook.
// Registration receives a source value; no native handler rereads process config.
func nativeMeasurementSourceFromEnvironment(client *http.Client) measurements.Source {
	endpoint := os.Getenv("CLICKHOUSE_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8123"
	}
	return measurements.Source{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
}

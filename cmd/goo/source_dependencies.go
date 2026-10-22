package main

import (
	"net/http"
	"os"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/measurements"
)

// clickHouseConnection is startup configuration, not a query or transport owner.
// Typed source adapters retain compilation, limits, semantics and custody policy.
type clickHouseConnection struct {
	Endpoint, User, Password string
	Client                   *http.Client
}

func clickHouseConnectionFromEnvironment(client *http.Client) clickHouseConnection {
	endpoint := os.Getenv("CLICKHOUSE_URL")
	if endpoint == "" {
		endpoint = "http://localhost:8123"
	}
	return clickHouseConnection{Endpoint: endpoint, User: os.Getenv("CLICKHOUSE_USER"), Password: os.Getenv("CLICKHOUSE_PASSWORD"), Client: client}
}
func (c clickHouseConnection) measurementSource() measurements.Source {
	return measurements.Source{Endpoint: c.Endpoint, User: c.User, Password: c.Password, Client: c.Client}
}
func (c clickHouseConnection) explorationSource() exploration.Source {
	return exploration.Source{Endpoint: c.Endpoint, User: c.User, Password: c.Password, Client: c.Client}
}

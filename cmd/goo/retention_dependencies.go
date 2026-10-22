package main

import (
	"os"

	"github.com/casselc/goo/retention"
)

// retentionConnection is optional startup configuration. Read basis is supplied
// per operation; no caller's await token becomes mutable server-wide state.
type retentionConnection struct{ URL string }

func retentionConnectionFromEnvironment() retentionConnection {
	return retentionConnection{URL: os.Getenv("GOO_XTDB_URL")}
}
func (c retentionConnection) store(readBasis string) retention.Store {
	return retention.Store{URL: c.URL, AwaitToken: readBasis}
}

package measurements

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/casselc/goo/identity"
)

func TestAllEnvironments437CannotBecomeNativeScope(t *testing.T) {
	wire, _ := identity.AllEnvironments().Selector()
	if _, err := metricEnvironmentPredicate(identity.AllEnvironments(), func(value, typ string) string { t.Fatal("All reached native parameter binding"); return "" }); err == nil {
		t.Fatal("All reached native lowering")
	}
	var reads atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reads.Add(1) }))
	defer server.Close()
	q := DiscoveryQuestion{EnvironmentSelector: &wire, Service: "worker", FromMS: 1000, ToMS: 2000}
	if _, err := (Source{Endpoint: server.URL, Client: server.Client()}).Discover(t.Context(), q); err == nil || !strings.Contains(err.Error(), "all-environments") {
		t.Fatal("native discovery did not explicitly refuse All", err)
	}
	if reads.Load() != 0 {
		t.Fatal("unsupported native scope performed metadata/source work")
	}
}

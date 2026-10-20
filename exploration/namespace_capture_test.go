package exploration

import (
	"context"
	"github.com/casselc/goo/identity"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestInventoryNamespaceCaptureRoundTripAndRefresh(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	source := Source{Endpoint: server.URL}
	for _, ns := range []*identity.Namespace{{Mode: "missing"}, {Mode: "value"}, {Mode: "value", Value: "rollout-canary"}} {
		t.Run(ns.Mode+ns.Value, func(t *testing.T) {
			store := CaptureStore{Dir: t.TempDir()}
			q := CaptureQuestion{Kind: "inventory", Inventory: &InventoryQuestion{Namespace: ns.Clone(), Environment: "prod", FromMS: 1000, ToMS: 61000}}
			saved, e := store.Create(context.Background(), source, "Scoped inventory", q, "")
			if e != nil {
				t.Fatal(e)
			}
			loaded, e := store.Load(saved.ID)
			if e != nil || !reflect.DeepEqual(loaded.Question, q) || !reflect.DeepEqual(loaded.Inventory.Question, *q.Inventory) {
				t.Fatal(loaded, e)
			}
			next, e := store.Create(context.Background(), source, "Scoped inventory", loaded.Question, loaded.ID)
			if e != nil || next.Parent != loaded.ID || !reflect.DeepEqual(next.Question, loaded.Question) {
				t.Fatal(next, e)
			}
			if _, e = store.Load(next.ID); e != nil {
				t.Fatal(e)
			}
			// Independent wire values must compare equal; a semantic scope change must fail.
			loaded.Question.Inventory.Namespace = &identity.Namespace{Mode: "value", Value: "foreign"}
			if e = validateCapture(loaded); e == nil {
				t.Fatal("mismatched inventory namespace accepted")
			}
		})
	}
}

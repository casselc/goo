package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInventorySingleReadAndExactAccounting(t *testing.T) {
	calls := 0
	row := InventoryRow{Service: "api'\\\n", Name: "<work>", Count: "9007199254740993", Errors: "3", Buckets: []string{"9007199254740990", "0", "0", "0", "0", "0", "0", "3"}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("readonly") != "1" {
			t.Error("not read-only")
		}
		json.NewEncoder(w).Encode(row)
	}))
	defer server.Close()
	q := InventoryQuestion{Environment: "prod'\\\n", FromMS: 1000, ToMS: 100000}
	p, e := CompileInventory(q)
	if e != nil {
		t.Fatal(e)
	}
	if strings.Contains(p.Groups, q.Environment) || strings.Contains(p.Groups, "OFFSET") || !strings.Contains(p.Groups, "LIMIT 1001") || p.Parameters["environment"] != q.Environment {
		t.Fatal("unsafe or page-defined inventory", p)
	}
	s := Source{Endpoint: server.URL}
	a, e := s.Inventory(context.Background(), q)
	if e != nil {
		t.Fatal(e)
	}
	b, e := s.Inventory(context.Background(), q)
	if e != nil || a.Basis != b.Basis || a.Count != row.Count || calls != 2 || a.State != "observed" {
		t.Fatal(a, b, e, calls)
	}
}
func TestInventoryRejectsPartialOrContradictoryEvidence(t *testing.T) {
	for _, mode := range []string{"overflow", "partition", "errors", "duplicate", "leading-zero", "missing-dimension", "null-dimension"} {
		t.Run(mode, func(t *testing.T) {
			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				row := InventoryRow{Service: "api", Name: "work", Count: "2", Errors: "1", Buckets: []string{"2", "0", "0", "0", "0", "0", "0", "0"}}
				n := 1
				switch mode {
				case "overflow":
					n = 1001
				case "duplicate":
					n = 2
				case "partition":
					row.Buckets[0] = "1"
				case "errors":
					row.Errors = "3"
				case "leading-zero":
					row.Count = "02"
				case "null-dimension":
					fmt.Fprintln(w, `{"service": null ,"name":"work","count":"2","errors":"1","buckets":["2","0","0","0","0","0","0","0"]}`)
					return
				case "missing-dimension":
					fmt.Fprintln(w, `{"count":"2","errors":"1","buckets":["2","0","0","0","0","0","0","0"]}`)
					return
				}
				for range n {
					json.NewEncoder(w).Encode(row)
				}
			}))
			defer s.Close()
			if _, e := (Source{Endpoint: s.URL}).Inventory(context.Background(), InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 100000}); e == nil {
				t.Fatal("accepted incomplete or contradictory inventory")
			}
		})
	}
}
func TestInventoryNoDataAndValidation(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer s.Close()
	r, e := (Source{Endpoint: s.URL}).Inventory(context.Background(), InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 100000})
	if e != nil || r.State != "no_data" || r.Count != "0" || len(r.Rows) != 0 {
		t.Fatal(r, e)
	}
	if _, e := CompileInventory(InventoryQuestion{Environment: "", FromMS: 1000, ToMS: 100000}); e == nil {
		t.Fatal("missing environment")
	}
}

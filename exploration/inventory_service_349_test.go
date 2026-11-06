package exploration

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/casselc/goo/identity"
)

func inventoryServiceRow349(service string) InventoryRow {
	return InventoryRow{Service: service, Name: "operation-'雪", Count: "9007199254740993", Errors: "1", Buckets: []string{"9007199254740993", "0", "0", "0", "0", "0", "0", "0"}}
}
func inventorySource349(body func(*http.Request) string) Source {
	return Source{Endpoint: "http://source.invalid", Client: &http.Client{Transport: recordWireTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body(r)))}, nil
	})}}
}
func TestInventoryExactServiceCompileAndWire349(t *testing.T) {
	for _, name := range []string{"", "api-'\\雪\n", strings.Repeat("x", 128)} {
		for _, namespace := range []*identity.Namespace{nil, {Mode: "missing"}, {Mode: "value"}, {Mode: "value", Value: "team-'雪"}} {
			q := InventoryQuestion{EnvironmentSelector: &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod-'雪"}}, Namespace: namespace, Service: new(name), FromMS: 1000, ToMS: 61000}
			plan, err := CompileInventory(q)
			if err != nil {
				t.Fatal(err)
			}
			query := q.Query()
			if query.Validate() != nil {
				t.Fatal("invalid exact-service record question")
			}
			predicate, err := Compile(query)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(plan.Groups, "LIMIT 1001") || !strings.Contains(plan.Groups, "ServiceName = {") || strings.Index(plan.Groups, "ServiceName = {") > strings.Index(plan.Groups, " GROUP BY ") {
				t.Fatal("service filter not before aggregate cap", plan.Groups)
			}
			if name != "" && strings.Contains(plan.Groups, name) {
				t.Fatal("literal service interpolated")
			}
			found := false
			for key, value := range plan.Parameters {
				if value == name && strings.Contains(plan.Groups, "{"+key+":String}") {
					found = true
				}
			}
			if !found || predicate.Parameters["from"] != "1000" || predicate.Parameters["to"] != "61000" {
				t.Fatal("missing typed parameters", plan.Parameters)
			}
			raw, _ := json.Marshal(q)
			var wire InventoryQuestion
			if err = json.Unmarshal(raw, &wire); err != nil || !reflect.DeepEqual(q, wire) {
				t.Fatal("service presence or typed scope lost", err)
			}
			clone := q.clone()
			*q.Service = "caller mutation"
			if *clone.Service != name {
				t.Fatal("service selector aliases caller")
			}
		}
	}
	base := `{"environment":"prod","fromMs":1000,"toMs":61000`
	for _, tail := range []string{`,"service":null}`, `,"service":1}`, `,"service":{}}`, `,"service":"\u0000"}`, `,"service":"\ud800"}`} {
		var q InventoryQuestion
		err := json.Unmarshal([]byte(base+tail), &q)
		if err == nil && q.Validate() == nil {
			t.Fatal("invalid service admitted", tail)
		}
	}
	for _, name := range []string{strings.Repeat("x", 129), "bad\x00", string([]byte{0xff})} {
		q := InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000, Service: new(name)}
		if q.Validate() == nil {
			t.Fatal("invalid service name admitted")
		}
	}
	malformed := append([]byte(base+`,"service":"`), 0xff)
	malformed = append(malformed, []byte(`"}`)...)
	var q InventoryQuestion
	if json.Unmarshal(malformed, &q) == nil {
		t.Fatal("raw malformed UTF-8 repaired into service")
	}
}
func TestInventoryExactServiceSourceAndScopeIsolation349(t *testing.T) {
	for _, name := range []string{"", "api-'雪"} {
		q := InventoryQuestion{Environment: "prod", Service: new(name), FromMS: 1000, ToMS: 61000}
		row := inventoryServiceRow349(name)
		raw, _ := json.Marshal(row)
		attempts := 0
		source := inventorySource349(func(r *http.Request) string {
			attempts++
			*q.Service = "caller changed after admission"
			return string(raw)
		})
		result, err := source.Inventory(context.Background(), q)
		if err != nil || attempts != 1 || result.Count != row.Count || result.Question.Service == nil || *result.Question.Service != name {
			t.Fatal("whole selected scope not frozen", result, err)
		}
		q.Service = new(name)
		source = inventorySource349(func(*http.Request) string {
			foreign := inventoryServiceRow349("foreign")
			data, _ := json.Marshal(foreign)
			return string(data)
		})
		if _, err = source.Inventory(context.Background(), q); err == nil {
			t.Fatal("foreign service source row admitted", name)
		}
		source = inventorySource349(func(*http.Request) string { return "" })
		result, err = source.Inventory(context.Background(), q)
		if err != nil || result.State != "no_data" || result.Count != "0" || result.Question.Service == nil || *result.Question.Service != name {
			t.Fatal("empty read widened selected service", result, err)
		}
		if ValidateServiceDiscoveryQuestion(q) == nil {
			t.Fatal("service leaked into multi-signal discovery")
		}
		if _, err = CompileServiceDiscovery(q); err == nil {
			t.Fatal("compiler ignored or partially applied service")
		}
		if _, err = PlanServiceDashboard(DashboardRequest{Scope: q, Service: name}, nil); err == nil {
			t.Fatal("duplicate dashboard service scope admitted")
		}
	}
}
func TestInventoryServiceCapAndCapture349(t *testing.T) {
	name := "api-'雪"
	q := InventoryQuestion{Environment: "prod", Service: new(name), Namespace: &identity.Namespace{Mode: "missing"}, FromMS: 1000, ToMS: 61000}
	source := inventorySource349(func(r *http.Request) string {
		for key, values := range r.URL.Query() {
			if strings.HasPrefix(key, "param_") && len(values) == 1 && values[0] == name {
				raw, _ := json.Marshal(inventoryServiceRow349(name))
				return string(raw)
			}
		}
		var out strings.Builder
		for range 1001 {
			raw, _ := json.Marshal(inventoryServiceRow349("unrelated"))
			out.Write(raw)
			out.WriteByte('\n')
		}
		return out.String()
	})
	selected, err := source.Inventory(context.Background(), q)
	if err != nil || len(selected.Rows) != 1 {
		t.Fatal("selected service blocked by unrelated cardinality", selected, err)
	}
	global := q.clone()
	global.Service = nil
	if _, err = source.Inventory(context.Background(), global); err == nil {
		t.Fatal("global cardinality sentinel weakened")
	}
	for _, service := range []string{name, ""} {
		q.Service = new(service)
		row := inventoryServiceRow349(service)
		raw, _ := json.Marshal(row)
		exact := inventorySource349(func(*http.Request) string { return string(raw) })
		store := CaptureStore{Dir: t.TempDir()}
		question := CaptureQuestion{Kind: "inventory", Inventory: &q}
		saved, err := store.Create(context.Background(), exact, "Selected inventory", question, "")
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := store.Load(saved.ID)
		if err != nil || !reflect.DeepEqual(loaded.Question, question) || loaded.Inventory.Question.Service == nil || *loaded.Inventory.Question.Service != service {
			t.Fatal("capture lost exact service presence", loaded, err)
		}
		loaded.Inventory.Rows[0].Service = "foreign"
		if validateCapture(loaded) == nil {
			t.Fatal("stored foreign service row admitted")
		}
	}
}

func TestInventoryServiceUnicodeIdentity349(t *testing.T) {
	for _, valid := range []string{`"\ufffd"`, `"\ud83d\ude42"`, `"�"`} {
		raw := []byte(`{"environment":"prod","fromMs":1000,"toMs":61000,"service":` + valid + `}`)
		var q InventoryQuestion
		if err := json.Unmarshal(raw, &q); err != nil || q.Validate() != nil {
			t.Fatal("valid Unicode identity refused", valid, err)
		}
		body := `{"service":` + valid + `,"name":` + valid + `,"count":"1","errors":"0","buckets":["1","0","0","0","0","0","0","0"]}`
		source := inventorySource349(func(*http.Request) string { return body })
		if _, err := source.Inventory(context.Background(), q); err != nil {
			t.Fatal("valid Unicode row refused", valid, err)
		}
	}
	for _, dimension := range []string{"service", "name"} {
		for _, invalid := range []string{`"\ud800"`, `"\udfff"`, string([]byte{'"', 0xff, '"'})} {
			body := `{"service":"�","name":"operation","count":"1","errors":"0","buckets":["1","0","0","0","0","0","0","0"]}`
			if dimension == "service" {
				body = strings.Replace(body, `"service":"�"`, `"service":`+invalid, 1)
			} else {
				body = strings.Replace(body, `"name":"operation"`, `"name":`+invalid, 1)
			}
			source := inventorySource349(func(*http.Request) string { return body })
			q := InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000, Service: new("�")}
			if _, err := source.Inventory(context.Background(), q); err == nil {
				t.Fatal("malformed raw dimension repaired into exact identity", dimension, invalid)
			}
		}
	}
}

func TestInventoryStoredRawServiceIdentity349(t *testing.T) {
	for _, service := range []string{"", "�"} {
		q := InventoryQuestion{Environment: "prod", FromMS: 1000, ToMS: 61000, Service: new(service)}
		row := inventoryServiceRow349(service)
		encoded, _ := json.Marshal(row)
		source := inventorySource349(func(*http.Request) string { return string(encoded) })
		store := CaptureStore{Dir: t.TempDir()}
		saved, err := store.Create(context.Background(), source, "Exact service", CaptureQuestion{Kind: "inventory", Inventory: &q}, "")
		if err != nil {
			t.Fatal(err)
		}
		filename := filepath.Join(store.Dir, saved.ID+".json")
		original, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		needle := `"rows":[{"service":"` + service + `",`
		for _, replacement := range []string{`"rows":[{`, `"rows":[{"service":null,`, `"rows":[{"service":"\ud800",`} {
			if !strings.Contains(string(original), needle) {
				t.Fatal("capture row location not found")
			}
			altered := strings.Replace(string(original), needle, replacement, 1)
			if err = os.WriteFile(filename, []byte(altered), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err = store.Load(saved.ID); err == nil {
				t.Fatal("stored raw service absence or lossy Unicode admitted", service, replacement)
			}
		}
	}
}

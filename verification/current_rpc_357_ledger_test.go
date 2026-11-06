package verification

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"slices"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/observability/quintgo/qobs"
	"github.com/casselc/observability/quintgo/validate"
)

type rpcCounter357 struct{ Calls, Source, Plans int }
type rpcFact357 struct {
	Name          string
	Before, After rpcCounter357
}
type rpcObservedCase357 struct {
	Key                     string
	Scope                   exploration.InventoryQuestion
	Automatic, Legacy, Plan exploration.DashboardPlan
	Widget                  exploration.DashboardWidget
	Initial                 struct {
		Request  exploration.Query
		Response struct{ Refusal struct{ Kind string } }
		Status   int
	}
	Retry struct {
		Request                     exploration.Query
		Response                    exploration.RecordAggregatesResult
		Status                      int
		HealthyBefore, HealthyAfter []json.RawMessage
	}
	RecordsRun struct {
		Request  exploration.Query
		Response exploration.Result
	}
	Oracle                    []struct{ ID string }
	ParentBefore, ParentAfter []json.RawMessage
	Retired                   struct{ Before, After json.RawMessage }
	Facts                     []rpcFact357
}
type rpcObservedLedger357 struct {
	Version string
	Fixture struct {
		Environment, Service, Namespace string
		FromMS, ToMS                    int64
		Rows                            []struct {
			ID, Service     string
			Namespace       *string
			Env, Attributes map[string]string
			At              int64
			Kind            int
			DurationNS      string
		}
	}
	Cases  []rpcObservedCase357
	Calls  []serviceContinuationCall352
	Errors []string
}

var rpcFactNames357 = []string{"profile-apply", "retry", "records-stage", "records-run", "return"}

func rpcObservedAdmission357(l rpcObservedLedger357, c rpcObservedCase357) error {
	q := c.Widget.Query
	if q == nil || q.Validate() != nil || c.Widget.State != "supported" || c.Widget.Family != "rpc-server-current" || c.Automatic.Profile != "rpc-server-current" || c.Legacy.Profile != "rpc-server" || c.Plan.Profile != "rpc-server-current" || c.Plan.Service != l.Fixture.Service || continuationJSON352(c.Plan.Scope) != continuationJSON352(c.Scope) {
		return fmt.Errorf("distinct explicit current/legacy plans")
	}
	if q.Environment != c.Scope.Environment || continuationJSON352(q.EnvironmentSelector) != continuationJSON352(c.Scope.EnvironmentSelector) || q.FromMS != c.Scope.FromMS || q.ToMS != c.Scope.ToMS || q.Summary != "span-duration" || !q.SpanStatus || q.Offset != 0 {
		return fmt.Errorf("complete fixed current profile scope")
	}
	if c.Initial.Status != 502 || c.Initial.Response.Refusal.Kind != "unavailable" || c.Retry.Status != 200 || continuationJSON352(c.Initial.Request) != continuationJSON352(*q) || continuationJSON352(c.Retry.Request) != continuationJSON352(*q) || continuationJSON352(c.Retry.Response.Question) != continuationJSON352(*q) {
		return fmt.Errorf("actual source fault/retry descriptor")
	}
	mode := 1
	name := ""
	if c.Scope.EnvironmentSelector == nil {
		return fmt.Errorf("native typed policy proof missing")
	}
	name = c.Scope.EnvironmentSelector.Scope.Value
	if c.Scope.EnvironmentSelector.Policy == "effective" {
		mode = 2
	}
	f := rpcFixture357{Mode: mode, Scope: c.Scope, Service: l.Fixture.Service}
	expectedIDs := []string{}
	sum := new(big.Int)
	for _, raw := range l.Fixture.Rows {
		resource := map[string]string{}
		for k, v := range raw.Env {
			resource[k] = v
		}
		if raw.Namespace != nil {
			resource["service.namespace"] = *raw.Namespace
		}
		kind := "Other"
		if raw.Kind == 2 {
			kind = "Server"
		}
		if raw.Kind == 3 {
			kind = "Client"
		}
		r := rpcRaw357{ID: raw.ID, Service: raw.Service, Kind: kind, At: raw.At, Resource: resource, Attributes: raw.Attributes}
		want := r.Service == f.Service && r.Kind == "Server" && r.At >= q.FromMS && r.At < q.ToMS && continuationEnvironment352(mode, name, r.Resource) && continuationNamespace352(c.Scope.Namespace, r.Resource) && r.Attributes["rpc.system.name"] != ""
		values := map[exploration.Field]string{{Kind: "column", Key: "service"}: r.Service, {Kind: "column", Key: "kind"}: r.Kind}
		for k, v := range resource {
			values[exploration.Field{Kind: "resource", Key: k}] = v
		}
		for k, v := range r.Attributes {
			values[exploration.Field{Kind: "attribute", Key: k}] = v
		}
		got := hdPredicate(q.Where, values) && r.At >= q.FromMS && r.At < q.ToMS && continuationEnvironment352(mode, name, resource)
		if got != want {
			return fmt.Errorf("literal current required system only predicate")
		}
		if want {
			expectedIDs = append(expectedIDs, raw.ID)
			d, ok := new(big.Int).SetString(raw.DurationNS, 10)
			if !ok || d.Sign() < 0 {
				return fmt.Errorf("raw fixture duration")
			}
			sum.Add(sum, d)
		}
	}
	slices.Sort(expectedIDs)
	oracle := []string{}
	for _, r := range c.Oracle {
		oracle = append(oracle, r.ID)
	}
	slices.Sort(oracle)
	gotIDs, err := reportedResultIDs353(c.RecordsRun.Response)
	if err != nil {
		return err
	}
	if len(expectedIDs) == 0 || !slices.Equal(expectedIDs, oracle) || !slices.Equal(expectedIDs, gotIDs) || c.Retry.Response.Count != fmt.Sprint(len(expectedIDs)) || c.RecordsRun.Response.Count != fmt.Sprint(len(expectedIDs)) {
		return fmt.Errorf("independent current raw IDs/count")
	}
	duration := c.Retry.Response.Duration
	if duration == nil || duration.Validate() != nil || duration.Count != fmt.Sprint(len(expectedIDs)) || duration.SumNS != sum.String() {
		return fmt.Errorf("independent raw duration sum")
	}
	if continuationJSON352(c.RecordsRun.Request) != continuationJSON352(*q) || continuationJSON352(c.RecordsRun.Response.Question) != continuationJSON352(*q) {
		return fmt.Errorf("manual records fixed question")
	}
	if len(c.Retry.HealthyBefore) == 0 || continuationJSON352(c.Retry.HealthyBefore) != continuationJSON352(c.Retry.HealthyAfter) || len(c.ParentBefore) < 2 || continuationJSON352(c.ParentBefore) != continuationJSON352(c.ParentAfter) || len(c.Retired.Before) == 0 || continuationJSON352(c.Retired.Before) != continuationJSON352(c.Retired.After) {
		return fmt.Errorf("healthy retry/back receipts or stale callback authority")
	}
	if len(c.Facts) != len(rpcFactNames357) {
		return fmt.Errorf("missing actual lifecycle fact")
	}
	for i, fact := range c.Facts {
		if fact.Name != rpcFactNames357[i] || fact.Before.Calls < 0 || fact.After.Calls < fact.Before.Calls || fact.After.Calls > len(l.Calls) {
			return fmt.Errorf("fact order/boundaries")
		}
		source, plans := 0, 0
		for _, call := range l.Calls[fact.Before.Calls:fact.After.Calls] {
			if call.Path == "/api/dashboards/plan" {
				plans++
			}
			if !serviceTemplateSource356(call) {
				continue
			}
			source++
			if i == 1 || i == 3 {
				path := "/api/record-aggregates"
				if i == 3 {
					path = "/api/explore"
				}
				var actual exploration.Query
				if call.Path != path || json.Unmarshal(call.Body, &actual) != nil || continuationJSON352(actual) != continuationJSON352(*q) {
					return fmt.Errorf("actual fixed retry/records source request")
				}
			}
		}
		want := 0
		if i == 1 || i == 3 {
			want = 1
		}
		if i == 0 {
			if plans != 1 || source < 2 {
				return fmt.Errorf("explicit profile Apply source work")
			}
		} else if source != want || plans != 0 {
			return fmt.Errorf("passive or single-widget source budget")
		}
		if fact.After.Source-fact.Before.Source != source || fact.After.Plans-fact.Before.Plans != plans {
			return fmt.Errorf("raw call/counter disagreement")
		}
	}
	return nil
}
func rpcModel357(t *testing.T, label string, actions []string, coherent bool) bool {
	t.Helper()
	sink := &qobs.MemorySink{}
	rec := qobs.NewRecorder(sink, qobs.WithActor(label))
	for i, a := range actions {
		rec.Record(context.Background(), a, nil, "obs.coherent", coherent || i != len(actions)-1)
	}
	checker, err := validate.NewChecker("model/widget_recovery_354.yaml", validate.Options{Quint: quintPath(t), Backend: "typescript", Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	r, err := checker.Check(context.Background(), sink.Steps())
	if err != nil {
		t.Fatal(err)
	}
	artifact("current-rpc357-"+label+".txt", []byte(r.Summary()))
	return r.OK()
}

// Protocol unit smoke invokes the actual reused binding wrapper, without
// claiming synthetic steps as browser observations.
func TestQuintCurrentRPCProtocol357(t *testing.T) {
	good := []string{"start", "source_failed", "retry", "succeed", "refused", "finish"}
	if !rpcModel357(t, "protocol-complete", good, true) {
		t.Fatal("complete recovery protocol refused")
	}
	if rpcModel357(t, "protocol-busy", []string{"start", "source_failed", "retry", "retry", "succeed", "finish"}, true) {
		t.Fatal("busy duplicate retry accepted")
	}
}
func TestQuintCurrentRPCObserved357(t *testing.T) {
	path := os.Getenv("GOO_CURRENT_RPC_357_REPORT")
	if path == "" {
		t.Skip("explicit actual native report required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var l rpcObservedLedger357
	if json.Unmarshal(raw, &l) != nil || l.Version != "goo.currentRpc357/1" || len(l.Cases) != 2 || len(l.Errors) != 0 {
		t.Fatal("incomplete actual native RPC campaign")
	}
	actions := []string{"start", "source_failed", "retry", "succeed", "refused", "finish"}
	check := func(label string, c rpcObservedCase357, want bool) {
		t.Helper()
		err := rpcObservedAdmission357(l, c)
		if (err == nil) != want {
			t.Fatalf("%s independent admission: %v", label, err)
		}
		if rpcModel357(t, label, actions, err == nil) != want {
			t.Fatalf("%s model verdict", label)
		}
	}
	for _, c := range l.Cases {
		check("actual-"+c.Key, c, true)
	}
	poisons := []func(*rpcObservedCase357){func(c *rpcObservedCase357) { c.Retry.Request.FromMS++ }, func(c *rpcObservedCase357) {
		c.Widget.Query.Where = &exploration.Predicate{Field: &exploration.Field{Kind: "attribute", Key: "rpc.method"}, Op: "ne"}
	}, func(c *rpcObservedCase357) { c.Retry.Response.Duration.SumNS = "0" }, func(c *rpcObservedCase357) { c.Retry.HealthyAfter = nil }, func(c *rpcObservedCase357) { c.ParentAfter = nil }, func(c *rpcObservedCase357) { c.Facts[2].After.Source++ }}
	for i, poison := range poisons {
		var fresh rpcObservedLedger357
		json.Unmarshal(raw, &fresh)
		c := fresh.Cases[0]
		poison(&c)
		check(fmt.Sprintf("poison-%d", i), c, false)
	}
}

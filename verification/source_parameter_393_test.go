package verification

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/measurements"
	"github.com/casselc/goo/metrics"
	"hegel.dev/go/hegel"
)

func literalQuestions393(value string) (exploration.Query, measurements.Question) {
	field := exploration.Field{Kind: "attribute", Key: "literal"}
	record := exploration.Query{Signal: "logs", Environment: "audit", FromMS: 1000, ToMS: 2000, Limit: 5, Where: &exploration.Predicate{Field: &field, Op: "eq", Value: value}}
	instrument := hdInstrument("literal-audit", "audit.literal", "gauge", "1", exploration.InventoryQuestion{Environment: "audit", FromMS: 1000, ToMS: 2000}, "audit-worker")
	instrument.Definition.Source.Attributes = map[string]string{"literal": value}
	return record, measurements.Question{Definition: *instrument.Definition, FromMS: 1000, ToMS: 2000}
}
func hasLiteral393(parameters map[string]string, value string) bool {
	for _, v := range parameters {
		if v == value {
			return true
		}
	}
	return false
}

// This property proves only public typed validation/compilation and immutable
// literal intent. The separate actual test below supplies ClickHouse proof.
func TestHegelSourceParameterIntent393(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(1, 100000))
		control := hegel.Draw(tc, hegel.Integers(1, 31))
		value := fmt.Sprintf("case-%c-%d", 'a'+seed%17, seed) + string(rune(control)) + []string{"\\x01", "\\n", "雪", "e\u0301", "\U0001f680", "\x7f"}[seed%6]
		if !utf8.ValidString(value) || strings.ContainsRune(value, 0) || len(value) > 512 {
			tc.Fatal("generator crossed common legal boundary")
		}
		q, n := literalQuestions393(value)
		before, _ := json.Marshal([]any{q, n})
		rp, e := exploration.Compile(q)
		if e != nil {
			tc.Fatal(e)
		}
		np, e := measurements.Compile(n)
		if e != nil {
			tc.Fatal(e)
		}
		if !hasLiteral393(rp.Parameters, value) || !hasLiteral393(np.Parameters, value) {
			tc.Fatal("compiler changed literal bytes")
		}
		alternate := value + "\\"
		q2, n2 := literalQuestions393(alternate)
		r2, e := exploration.Compile(q2)
		if e != nil {
			tc.Fatal(e)
		}
		m2, e := measurements.Compile(n2)
		if e != nil {
			tc.Fatal(e)
		}
		if rp.Rows != r2.Rows || np.SQL != m2.SQL || reflect.DeepEqual(rp.Parameters, r2.Parameters) || reflect.DeepEqual(np.Parameters, m2.Parameters) || !hasLiteral393(r2.Parameters, alternate) || !hasLiteral393(m2.Parameters, alternate) {
			tc.Fatal("literal mutation changed SQL structure or lost distinct bytes")
		}
		after, _ := json.Marshal([]any{q, n})
		if string(before) != string(after) {
			tc.Fatal("compile mutated caller questions")
		}
	}, options(32)...)
}

func TestTypedSourceParameterRefusals393(t *testing.T) {
	for _, c := range []struct {
		name, value string
		recordOK    bool
	}{
		{"NUL", "a\x00b", false}, {"invalid-UTF8", string([]byte{'a', 0xff, 'b'}), false}, {"native-overbound", strings.Repeat("x", 513), true}, {"record-overbound", strings.Repeat("x", 1025), false},
	} {
		t.Run(c.name, func(t *testing.T) {
			q, n := literalQuestions393(c.value)
			_, recordError := exploration.Compile(q)
			_, nativeError := measurements.Compile(n)
			if (recordError == nil) != c.recordOK || nativeError == nil {
				t.Fatal("typed compiler refusal boundary changed")
			}
			attempts := 0
			client := &http.Client{Transport: snapshotTransport(func(*http.Request) (*http.Response, error) {
				attempts++
				return nil, fmt.Errorf("refused input reached HTTP")
			})}
			if !c.recordOK {
				_, e := (exploration.Source{Endpoint: "http://source.invalid", Client: client}).Fetch(t.Context(), q)
				if e == nil || attempts != 0 {
					t.Fatal("typed record refusal performed source IO")
				}
			}
			_, e := (measurements.Source{Endpoint: "http://source.invalid", Client: client}).Run(t.Context(), n, nil)
			if e == nil || attempts != 0 {
				t.Fatalf("typed native refusal performed source IO: attempts=%d error=%v", attempts, e)
			}
		})
	}
}

type literalVector393 struct{ Name, Value, ID, UTF8Hex, SpanID, TimeNS, GaugeValue string }
type literalRawLog393 struct {
	ServiceName, SpanID, TraceID, Body, SeverityText, AtNS, Literal, LiteralHex, IDHex string
	LiteralPresent                                                                     int
	ResourceAttributes, LogAttributes                                                  map[string]string
}
type literalRawGauge393 struct {
	ServiceName, MetricName, MetricUnit, ScopeName, ScopeVersion, ScopeSchemaUrl, ResourceSchemaUrl, AtNS, Value, Literal, LiteralHex, IDHex string
	LiteralPresent                                                                                                                           int
	ResourceAttributes, Attributes                                                                                                           map[string]string
}
type literalBank393 struct {
	Version, Environment, Service, Namespace, Scope, SchemaURL, Metric, Unit, AttributeKey string
	FromMS, ToMS                                                                           int64
	Vectors                                                                                []literalVector393
	RawLogs                                                                                []literalRawLog393
	RawGauges                                                                              []literalRawGauge393
}
type literalWire393 struct {
	SQL        string
	Parameters map[string]string
	Status     int
}
type literalTransport393 struct {
	Base  http.RoundTripper
	mu    sync.Mutex
	Calls []literalWire393
}

func (t *literalTransport393) RoundTrip(req *http.Request) (*http.Response, error) {
	sql, e := io.ReadAll(req.Body)
	if e != nil {
		return nil, e
	}
	req.Body = io.NopCloser(strings.NewReader(string(sql)))
	parameters := map[string]string{}
	for key, values := range req.URL.Query() {
		if len(values) != 1 {
			return nil, fmt.Errorf("duplicate source parameter")
		}
		parameters[key] = values[0]
	}
	response, e := t.Base.RoundTrip(req)
	status := 0
	if response != nil {
		status = response.StatusCode
	}
	t.mu.Lock()
	t.Calls = append(t.Calls, literalWire393{string(sql), parameters, status})
	t.mu.Unlock()
	return response, e
}

type literalOutcome393 struct {
	Name, Arm, LiteralHex  string
	ExpectedIDs, ActualIDs []string
	Question               json.RawMessage
	Calls                  []literalWire393
	Error                  string `json:"error,omitempty"`
}

func admitLiteralBank393(bank literalBank393) error {
	if bank.Version != "goo.literalTransport393Bank/1" || len(bank.Vectors) != 50 || len(bank.RawLogs) != 101 || len(bank.RawGauges) != 101 || bank.AttributeKey != "literal" || bank.Unit != "1" || bank.ToMS <= bank.FromMS {
		return fmt.Errorf("finite settled bank missing")
	}
	values, ids := map[string]bool{}, map[string]bool{}
	for _, v := range bank.Vectors {
		if !utf8.ValidString(v.Value) || strings.ContainsRune(v.Value, 0) || len(v.Value) > 512 || values[v.Value] || ids[v.ID] || strings.ToUpper(hex.EncodeToString([]byte(v.Value))) != v.UTF8Hex {
			return fmt.Errorf("invalid/duplicate original UTF8 bank vector")
		}
		values[v.Value] = true
		ids[v.ID] = true
		logFound, gaugeFound := false, false
		for _, row := range bank.RawLogs {
			if row.LogAttributes["vector.id"] != v.ID {
				continue
			}
			if logFound || row.SpanID != v.SpanID || row.AtNS != v.TimeNS || row.LiteralPresent != 1 || row.Literal != v.Value || row.LiteralHex != v.UTF8Hex || row.ServiceName != bank.Service || row.ResourceAttributes["deployment.environment.name"] != bank.Environment {
				return fmt.Errorf("physical log bytes/membership differ from original OTLP")
			}
			logFound = true
		}
		for _, row := range bank.RawGauges {
			if row.Attributes["vector.id"] != v.ID {
				continue
			}
			if gaugeFound || row.AtNS != v.TimeNS || row.LiteralPresent != 1 || row.Literal != v.Value || row.LiteralHex != v.UTF8Hex || row.Value != v.GaugeValue || row.ServiceName != bank.Service || row.MetricName != bank.Metric || row.MetricUnit != bank.Unit || row.ScopeName != bank.Scope || row.ScopeSchemaUrl != bank.SchemaURL || row.ResourceAttributes["deployment.environment.name"] != bank.Environment {
				return fmt.Errorf("physical metric bytes/identity differ from original OTLP")
			}
			gaugeFound = true
		}
		if !logFound || !gaugeFound {
			return fmt.Errorf("unsettled source bank")
		}
	}
	return nil
}

func recordQuestion393(bank literalBank393, value string) exploration.Query {
	return exploration.Query{Signal: "logs", Environment: bank.Environment, FromMS: bank.FromMS, ToMS: bank.ToMS, Limit: 5, Where: &exploration.Predicate{All: []exploration.Predicate{
		{Field: &exploration.Field{Kind: "column", Key: "service"}, Op: "eq", Value: bank.Service},
		{Field: &exploration.Field{Kind: "resource", Key: "service.namespace"}, Op: "eq", Value: bank.Namespace},
		{Field: &exploration.Field{Kind: "attribute", Key: bank.AttributeKey}, Op: "eq", Value: value},
	}}}
}
func nativeQuestion393(bank literalBank393, value string, exact *literalRawGauge393) measurements.Question {
	_, q := literalQuestions393(value)
	d := &q.Definition
	d.Environment = bank.Environment
	d.MaxGapMS = 600000
	d.Source.Services = []string{bank.Service}
	d.Source.Namespace = &identity.Namespace{Mode: "value", Value: bank.Namespace}
	d.Source.Scope = bank.Scope
	d.Source.Version = "1"
	d.Source.SchemaURL = bank.SchemaURL
	d.Source.Metric = bank.Metric
	d.Source.Unit = bank.Unit
	d.Source.Attributes = map[string]string{bank.AttributeKey: value}
	q.FromMS = bank.FromMS
	q.ToMS = bank.ToMS
	if exact != nil {
		d.Source.Attributes = nil
		d.Source.ExactStream = &metrics.Stream{Resource: exact.ResourceAttributes, Attributes: exact.Attributes, Scope: exact.ScopeName, Version: exact.ScopeVersion, SchemaURL: exact.ScopeSchemaUrl, ResourceSchemaURL: exact.ResourceSchemaUrl, Name: exact.MetricName, Unit: exact.MetricUnit}
	}
	return q
}
func expectedLiteral393(bank literalBank393, value string) []string {
	ids := []string{}
	for _, row := range bank.RawLogs {
		if row.LiteralPresent == 1 && row.Literal == value {
			ids = append(ids, row.LogAttributes["vector.id"])
		}
	}
	slices.Sort(ids)
	return ids
}
func recordIDs393(result *exploration.Result, literal string) ([]string, error) {
	ids := []string{}
	for _, wire := range result.Rows {
		var row struct{ Attributes map[string]string }
		if e := json.Unmarshal(wire, &row); e != nil {
			return nil, e
		}
		if value, ok := row.Attributes["literal"]; !ok || value != literal {
			return nil, fmt.Errorf("returned record literal bytes changed")
		}
		id, ok := row.Attributes["vector.id"]
		if !ok {
			return nil, fmt.Errorf("record vector identity absent")
		}
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids, nil
}
func gaugeIDs393(result *measurements.Result, literal string) ([]string, error) {
	ids := []string{}
	if result.Gauges == nil {
		return nil, fmt.Errorf("native source supplied no gauges")
	}
	for _, series := range result.Gauges.Series {
		if len(series.Samples) != 1 {
			return nil, fmt.Errorf("native stream changed sample population")
		}
		if value, ok := series.Stream.Attributes["literal"]; !ok || value != literal {
			return nil, fmt.Errorf("returned native literal bytes changed")
		}
		id, ok := series.Stream.Attributes["vector.id"]
		if !ok {
			return nil, fmt.Errorf("native vector identity absent")
		}
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids, nil
}

type echoLiteral393 struct{ Encoded, Original string }

func TestActualSourceParameterRoundTrips393(t *testing.T) {
	path := os.Getenv("GOO_SOURCE_PARAMETER_393_BANK")
	if path == "" {
		t.Skip("settled ordinary OTLP bank required")
	}
	wire, e := os.ReadFile(path)
	if e != nil {
		t.Fatal(e)
	}
	var bank literalBank393
	if e = json.Unmarshal(wire, &bank); e != nil {
		t.Fatal(e)
	}
	if e = admitLiteralBank393(bank); e != nil {
		t.Fatal(e)
	}
	endpoint := os.Getenv("GOO_SOURCE_PARAMETER_393_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("GOO_TEST_CLICKHOUSE_URL")
	}
	if endpoint == "" {
		endpoint = "http://127.0.0.1:8123"
	}
	user, password := os.Getenv("GOO_TEST_CLICKHOUSE_USER"), os.Getenv("GOO_TEST_CLICKHOUSE_PASSWORD")
	if user == "" {
		user = "lab"
	}
	if password == "" {
		password = "lab-local"
	}
	outcomes := []literalOutcome393{}
	echoes := map[echoLiteral393]bool{}
	negatives := []literalOutcome393{}
	defer func() {
		report := map[string]any{"version": "goo.sourceParameter393/1", "bank": path, "vectors": len(bank.Vectors), "outcomes": outcomes, "negativeControls": negatives, "typedStringEchoes": len(echoes)}
		data, _ := json.MarshalIndent(report, "", "  ")
		if e := artifact("source-parameter393-audit.json", data); e != nil {
			t.Error(e)
		}
	}()
	checkCalls := func(calls []literalWire393, parameters map[string]string, metric bool) {
		if len(calls) != map[bool]int{false: 3, true: 1}[metric] {
			t.Errorf("unexpected public source read count: metric=%v calls=%d", metric, len(calls))
		}
		for _, call := range calls {
			if call.Status != 200 || call.Parameters["readonly"] != "1" || call.Parameters["max_execution_time"] != "10" || call.Parameters["max_bytes_to_read"] != "268435456" || call.Parameters["result_overflow_mode"] != "throw" || call.Parameters["read_overflow_mode"] != "throw" {
				t.Error("source status/readonly/budgets changed")
			}
			expectedLimit := "1002"
			if metric {
				expectedLimit = "10001"
			}
			if call.Parameters["max_result_rows"] != expectedLimit {
				t.Error("public source row budget widened")
			}
			if !metric && (call.Parameters["max_memory_usage"] != "268435456" || call.Parameters["cancel_http_readonly_queries_on_client_close"] != "1") {
				t.Error("record source memory/cancellation settings changed")
			}
			for key, original := range parameters {
				if strings.Contains(call.SQL, "{"+key+":String}") {
					encoded, ok := call.Parameters["param_"+key]
					if !ok {
						t.Error("typed literal parameter absent")
					}
					echoes[echoLiteral393{encoded, original}] = true
				}
			}
		}
	}
	run := func(name, arm, value string, exact *literalRawGauge393) literalOutcome393 {
		transport := &literalTransport393{Base: http.DefaultTransport}
		client := &http.Client{Timeout: 15 * time.Second, Transport: transport}
		wanted := expectedLiteral393(bank, value)
		outcome := literalOutcome393{Name: name, Arm: arm, LiteralHex: strings.ToUpper(hex.EncodeToString([]byte(value))), ExpectedIDs: wanted}
		if arm == "record" {
			q := recordQuestion393(bank, value)
			before, _ := json.Marshal(q)
			outcome.Question = before
			plan, err := exploration.Compile(q)
			if err != nil {
				t.Fatal(err)
			}
			result, err := (exploration.Source{Endpoint: endpoint, User: user, Password: password, Client: client}).Fetch(t.Context(), q)
			if err != nil {
				outcome.Error = err.Error()
				t.Errorf("%s/%s actual source: %v", name, arm, err)
			} else {
				outcome.ActualIDs, err = recordIDs393(result, value)
				if err != nil {
					t.Error(err)
				}
				if result.Count != strconv.Itoa(len(wanted)) || result.HasMore || !reflect.DeepEqual(result.Question, q) {
					t.Error("actual record count/question changed")
				}
			}
			after, _ := json.Marshal(q)
			if string(before) != string(after) {
				t.Error("record source mutated original question")
			}
			checkCalls(transport.Calls, plan.Parameters, false)
		} else {
			if exact != nil {
				wanted = []string{exact.Attributes["vector.id"]}
				outcome.ExpectedIDs = wanted
			}
			q := nativeQuestion393(bank, value, exact)
			before, _ := json.Marshal(q)
			outcome.Question = before
			plan, err := measurements.Compile(q)
			if err != nil {
				t.Fatal(err)
			}
			result, err := (measurements.Source{Endpoint: endpoint, User: user, Password: password, Client: client}).Run(t.Context(), q, nil)
			if err != nil {
				outcome.Error = err.Error()
				t.Errorf("%s/%s actual source: %v", name, arm, err)
			} else {
				outcome.ActualIDs, err = gaugeIDs393(result, value)
				if err != nil {
					t.Error(err)
				}
				for _, series := range result.Gauges.Series {
					id := series.Stream.Attributes["vector.id"]
					matched := false
					for _, raw := range bank.RawGauges {
						if raw.Attributes["vector.id"] != id {
							continue
						}
						matched = true
						expected, e := strconv.ParseFloat(raw.Value, 64)
						sample := series.Samples[0]
						if e != nil || sample.Value == nil || *sample.Value != expected || sample.AtNS != raw.AtNS || !reflect.DeepEqual(series.Stream.Resource, raw.ResourceAttributes) || !reflect.DeepEqual(series.Stream.Attributes, raw.Attributes) {
							t.Error("returned native source bytes/value/identity differ from settled raw row")
						}
					}
					if !matched {
						t.Error("returned native row absent from raw bank")
					}
				}

				if result.Count != strconv.Itoa(len(wanted)) || !reflect.DeepEqual(result.Question, q) {
					t.Error("actual native count/question changed")
				}
			}
			after, _ := json.Marshal(q)
			if string(before) != string(after) {
				t.Error("native source mutated original question")
			}
			checkCalls(transport.Calls, plan.Parameters, true)
		}
		outcome.Calls = transport.Calls
		if !slices.Equal(outcome.ActualIDs, wanted) {
			t.Errorf("%s/%s literal membership mismatch: want=%v got=%v hex=%s", name, arm, wanted, outcome.ActualIDs, outcome.LiteralHex)
		}
		return outcome
	}
	for _, v := range bank.Vectors {
		var exact *literalRawGauge393
		for i := range bank.RawGauges {
			if bank.RawGauges[i].Attributes["vector.id"] == v.ID {
				exact = &bank.RawGauges[i]
				break
			}
		}
		if exact == nil {
			t.Fatal("settled exact gauge absent")
		}
		for _, arm := range []string{"record", "native-attributes", "native-exact-json"} {
			var selected *literalRawGauge393
			if arm == "native-exact-json" {
				selected = exact
			}
			outcomes = append(outcomes, run(v.Name, arm, v.Value, selected))
		}
	}
	// Actual near-miss and decode-looking controls select different known rows.
	// Their successful source reads must fail the original target-ID oracle.
	for _, name := range []string{"ascii-01", "escapedLF", "doubleSlashLF", "escapedHexA", "nfc", "decomposed"} {
		for _, v := range bank.Vectors {
			if v.Name != name {
				continue
			}
			changed := v.Value + "!"
			if name == "escapedLF" {
				changed = "\n"
			}
			if name == "doubleSlashLF" {
				changed = "\\n"
			}
			if name == "escapedHexA" {
				changed = "A"
			}
			for _, arm := range []string{"record", "native-attributes"} {
				outcome := run("negative-"+name, arm, changed, nil)
				if slices.Equal(outcome.ActualIDs, []string{v.ID}) || len(outcome.ActualIDs) != 1 {
					t.Error("literal/decode mutation escaped original membership oracle")
				}
				negatives = append(negatives, outcome)
			}
		}
	}
	if !t.Failed() {
		if e := echoSourceLiterals393(t.Context(), endpoint, user, password, echoes); e != nil {
			t.Error(e)
		}
	}
	t.Logf("%d real public-source arms, %d altered/decode controls, %d distinct typed String parameter echoes", len(outcomes), len(negatives), len(echoes))
}

// Send observed parameter text unchanged through a separately bounded server
// String decoder probe. The server returns hex bytes, not an encoder's spelling.
func echoSourceLiterals393(ctx context.Context, endpoint, user, password string, values map[echoLiteral393]bool) error {
	pending := make([]echoLiteral393, 0, len(values))
	for value := range values {
		pending = append(pending, value)
	}
	slices.SortFunc(pending, func(a, b echoLiteral393) int {
		return strings.Compare(a.Original+"\x00"+a.Encoded, b.Original+"\x00"+b.Encoded)
	})
	for start := 0; start < len(pending); start += 24 {
		group := pending[start:min(start+24, len(pending))]
		target, e := url.Parse(endpoint)
		if e != nil {
			return e
		}
		params := target.Query()
		for key, value := range map[string]string{"readonly": "1", "max_execution_time": "10", "max_result_rows": "1", "result_overflow_mode": "throw", "max_bytes_to_read": "268435456", "read_overflow_mode": "throw"} {
			params.Set(key, value)
		}
		columns := []string{}
		for index, value := range group {
			key := "v" + strconv.Itoa(index)
			params.Set("param_"+key, value.Encoded)
			columns = append(columns, "hex({"+key+":String}) AS "+key)
		}
		target.RawQuery = params.Encode()
		req, e := http.NewRequestWithContext(ctx, "POST", target.String(), strings.NewReader("SELECT "+strings.Join(columns, ",")+" FORMAT JSONEachRow"))
		if e != nil {
			return e
		}
		req.SetBasicAuth(user, password)
		response, e := (&http.Client{Timeout: 15 * time.Second}).Do(req)
		if e != nil {
			return e
		}
		body, e := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		response.Body.Close()
		if e != nil {
			return e
		}
		if response.StatusCode != 200 {
			return fmt.Errorf("typed literal echo HTTP%d", response.StatusCode)
		}
		var actual map[string]string
		if e = json.Unmarshal(body, &actual); e != nil {
			return e
		}
		for index, value := range group {
			want := strings.ToUpper(hex.EncodeToString([]byte(value.Original)))
			if actual["v"+strconv.Itoa(index)] != want {
				return fmt.Errorf("typed parameter server bytes differ: want=%s got=%s", want, actual["v"+strconv.Itoa(index)])
			}
		}
	}
	return nil
}

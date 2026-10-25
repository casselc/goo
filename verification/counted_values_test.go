package verification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/casselc/goo/exploration"
	"hegel.dev/go/hegel"
)

type cvFact struct{ Number, Value *string }

func cvReference(facts []cvFact, threshold int64) (string, []exploration.CountedValue) {
	counts := map[string]int{}
	values := map[string]*string{}
	total := 0
	for _, f := range facts {
		if f.Number == nil {
			continue
		}
		n, e := strconv.ParseInt(*f.Number, 10, 64)
		if e != nil || strconv.FormatInt(n, 10) != *f.Number || n < threshold {
			continue
		}
		key := "missing"
		if f.Value != nil {
			key = "present:" + *f.Value
		}
		counts[key]++
		values[key] = f.Value
		total++
	}
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	groups := []exploration.CountedValue{}
	for _, k := range keys {
		groups = append(groups, exploration.CountedValue{Value: values[k], Count: strconv.Itoa(counts[k])})
	}
	return strconv.Itoa(total), groups
}
func cvQuestion(kind string, threshold int64) exploration.Query {
	field := exploration.Field{Kind: "attribute", Key: "work.items"}
	binding := exploration.NumericBinding{Field: field, Type: "int64", Unit: "{item}", Parse: "canonical-decimal"}
	return exploration.Query{Signal: "spans", Environment: "ordinary", FromMS: 1000, ToMS: 61000, Limit: 17, GroupBy: []exploration.Field{{Kind: kind, Key: "work.category"}}, Where: &exploration.Predicate{Numeric: &binding, Op: "gte", Value: strconv.FormatInt(threshold, 10)}}
}
func cvRows(groups []exploration.CountedValue) []string {
	out := []string{}
	for _, g := range groups {
		b, _ := json.Marshal(g)
		out = append(out, string(b))
	}
	return out
}

type cvTransport func(*http.Request) (*http.Response, error)

func (f cvTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type cvFakeServer struct {
	URL    string
	client *http.Client
}

func (s *cvFakeServer) Client() *http.Client { return s.client }
func (s *cvFakeServer) Close()               {}
func cvFake(count string, groups []string, mutate func(int)) (*cvFakeServer, *atomic.Int32) {
	calls := &atomic.Int32{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		body, _ := io.ReadAll(r.Body)
		if r.Method != "POST" || r.URL.Query().Get("readonly") != "1" || r.URL.Query().Get("param_environment") != "ordinary" || !strings.Contains(string(body), "otel_traces") {
			http.Error(w, "wrong typed source request", 500)
			return
		}
		if mutate != nil {
			mutate(n)
		}
		w.Header().Set("Content-Type", "application/json")
		switch n {
		case 1:
			fmt.Fprintln(w, count)
		case 2:
			for _, g := range groups {
				fmt.Fprintln(w, g)
			}
		default:
			http.Error(w, "unexpected third source read", 500)
		}
	})
	server := &cvFakeServer{URL: "http://counted.invalid"}
	server.client = &http.Client{Transport: cvTransport(func(r *http.Request) (*http.Response, error) {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		return w.Result(), nil
	})}
	return server, calls
}
func cvRawBasis(rows []string) string {
	raw := []json.RawMessage{}
	for _, r := range rows {
		raw = append(raw, json.RawMessage(r))
	}
	b, _ := json.Marshal(raw)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func TestHegelCountedValuesReferenceAndSourceBoundary(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(-100, 100))
		threshold := int64(hegel.Draw(tc, hegel.Integers(-8, 8)))
		facts := []cvFact{}
		for i := 0; i < 140; i++ {
			raw := strconv.Itoa((i+seed)%23 - 11)
			var number *string = &raw
			switch i % 13 {
			case 0:
				number = nil
			case 1:
				number = lpcPointer("")
			case 2:
				number = lpcPointer("+3")
			case 3:
				number = lpcPointer("03")
			case 4:
				number = lpcPointer("9223372036854775808")
			}
			var value *string
			switch i % 4 {
			case 0:
				value = nil
			case 1:
				value = lpcPointer("")
			case 2:
				value = lpcPointer("checkout'\\\nπ")
			case 3:
				value = lpcPointer("queue")
			}
			facts = append(facts, cvFact{number, value})
		}
		expected, groups := cvReference(facts, threshold)
		for _, kind := range []string{"attribute", "resource"} {
			q := cvQuestion(kind, threshold)
			before, _ := json.Marshal(q)
			rawCount := fmt.Sprintf(`{"count":%q}`, expected)
			groupRows := cvRows(groups)
			server, calls := cvFake(rawCount, groupRows, nil)
			result, e := (exploration.Source{Endpoint: server.URL, Client: server.Client()}).FetchCountedValues(context.Background(), q)
			server.Close()
			after, _ := json.Marshal(q)
			if e != nil || calls.Load() != 2 || result.Version != "goo/counted-values/v1" || result.Count != expected || !reflect.DeepEqual(result.Groups, groups) || !nhLiveSameQuery(result.Question, q) || string(before) != string(after) {
				tc.Errorf("counted source/reference boundary %v", e)
				return
			}
			state := "observed"
			if expected == "0" {
				state = "no_data"
			}
			if result.State != state || result.Reads.Count.Basis != cvRawBasis([]string{rawCount}) || result.Reads.Groups.Basis != cvRawBasis(groupRows) {
				tc.Errorf("counts/state/read receipt")
				return
			}
			a, e1 := time.Parse(time.RFC3339Nano, result.Reads.Count.ReadAt)
			b, e2 := time.Parse(time.RFC3339Nano, result.Reads.Groups.ReadAt)
			c, e3 := time.Parse(time.RFC3339Nano, result.CapturedAt)
			if e1 != nil || e2 != nil || e3 != nil || b.Before(a) || c.Before(b) {
				tc.Errorf("read clocks")
				return
			}
			// Group observations may race independently of the count. Preserve them
			// diagnostically, but withhold an actionable coherent population.
			wrong, _ := strconv.Atoi(expected)
			changedServer, _ := cvFake(fmt.Sprintf(`{"count":%q}`, strconv.Itoa(wrong+1)), groupRows, nil)
			changed, e := (exploration.Source{Endpoint: changedServer.URL, Client: changedServer.Client()}).FetchCountedValues(context.Background(), q)
			changedServer.Close()
			if e != nil || changed.State != "changed_between_reads" || !reflect.DeepEqual(changed.Groups, groups) {
				tc.Errorf("changed accounting hidden %v", e)
				return
			}
		}
	}, options(24)...)
}
func TestCountedValuesMalformedGroupsAndDetachedCaller(t *testing.T) {
	q := cvQuestion("attribute", 0)
	for _, c := range []struct {
		name, count string
		groups      []string
	}{
		{"noncanonical-total", `{"count":"01"}`, nil},
		{"negative-total", `{"count":"-1"}`, nil},
		{"zero-group", `{"count":"0"}`, []string{`{"g0":"x","count":"0"}`}},
		{"duplicate-empty", `{"count":"2"}`, []string{`{"g0":"","count":"1"}`, `{"g0":"","count":"1"}`}},
		{"duplicate-missing", `{"count":"2"}`, []string{`{"g0":null,"count":"1"}`, `{"g0":null,"count":"1"}`}},
		{"missing-g0", `{"count":"1"}`, []string{`{"count":"1"}`}},
		{"numeric-value", `{"count":"1"}`, []string{`{"g0":7,"count":"1"}`}},
		{"overflow-group", `{"count":"1"}`, []string{`{"g0":"x","count":"18446744073709551616"}`}},
	} {
		t.Run(c.name, func(t *testing.T) {
			s, _ := cvFake(c.count, c.groups, nil)
			defer s.Close()
			if _, e := (exploration.Source{Endpoint: s.URL, Client: s.Client()}).FetchCountedValues(context.Background(), q); e == nil {
				t.Fatal("malformed counted source accepted")
			}
		})
	}
	detached := cvQuestion("attribute", 0)
	original, _ := json.Marshal(detached)
	s, _ := cvFake(`{"count":"1"}`, []string{`{"g0":null,"count":"1"}`}, func(n int) {
		if n == 1 {
			detached.Where.Value = "999"
			detached.GroupBy[0].Key = "caller-mutated"
		}
	})
	defer s.Close()
	result, e := (exploration.Source{Endpoint: s.URL, Client: s.Client()}).FetchCountedValues(context.Background(), detached)
	if e != nil {
		t.Fatal(e)
	}
	got, _ := json.Marshal(result.Question)
	if string(got) != string(original) {
		t.Fatal("caller mutation changed detached question")
	}
}

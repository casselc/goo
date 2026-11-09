//go:build ignore

// A journey-only HTTP workload. Every exported span surrounds a real loopback
// HTTP request. Partial instrumentation cases are deliberate data-quality controls,
// not semantic-convention conformance examples. Run from examples/report-app so
// the already pinned SDK/exporter module supplies dependencies; no product deps.
package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const schemaURL = "https://opentelemetry.io/schemas/1.45.0"

type scenario struct {
	Label, Method, Path, Address, Scope                                    string
	MissingMethod, EmptyMethod, LegacyMethod, MissingAddress, EmptyAddress bool
	DelayBeforeClient                                                      time.Duration
}
type observedCall struct {
	Label            string   `json:"label"`
	Method           string   `json:"method"`
	Address          string   `json:"address"`
	Port             int      `json:"port"`
	Status           int      `json:"status"`
	Environment      string   `json:"environment"`
	Namespace        string   `json:"namespace"`
	TraceID          string   `json:"traceId"`
	RootSpanID       string   `json:"rootSpanId"`
	ClientSpanID     string   `json:"clientSpanId"`
	DependencySpanID string   `json:"dependencySpanId"`
	PartialFields    []string `json:"partialFields"`
}
type output struct {
	Version            string         `json:"version"`
	Token              string         `json:"token"`
	Service            string         `json:"service"`
	DependencyService  string         `json:"dependencyService"`
	Namespace          string         `json:"namespace"`
	OtherNamespace     string         `json:"otherNamespace"`
	Environment        string         `json:"environment"`
	ForeignEnvironment string         `json:"foreignEnvironment"`
	FromMS             int64          `json:"fromMs"`
	ToMS               int64          `json:"toMs"`
	Calls              []observedCall `json:"calls"`
	EmptyRootSpanID    string         `json:"emptyRootSpanId"`
	Convention         string         `json:"convention"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	token := rand.Text()[:10]
	f := output{Version: "goo.httpClient385Workload/1", Token: token, Service: "http-outgoing385-" + token, DependencyService: "http-peer385-" + token, Namespace: "http385", OtherNamespace: "other-http385", Environment: "http385-" + token, ForeignEnvironment: "foreign-http385-" + token, FromMS: time.Now().Add(-time.Second).UnixMilli(), Convention: "OpenTelemetry semantic conventions v1.45.0; manual real HTTP instrumentation with explicitly labeled partial controls"}
	endpoint := os.Getenv("GOO_HTTP_CLIENT_385_OTLP")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:4318/v1/traces"
	}
	providers := map[string]*sdktrace.TracerProvider{}
	create := func(key, service, namespace, environment string) error {
		exp, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpointURL(endpoint))
		if err != nil {
			return err
		}
		attrs := []attribute.KeyValue{attribute.String("service.name", service), attribute.String("service.namespace", namespace), attribute.String("deployment.environment.name", environment), attribute.String("service.instance.id", token+"-"+key)}
		if key == "foreign" {
			attrs = append(attrs, attribute.String("deployment.environment", f.Environment))
		}
		providers[key] = sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp), sdktrace.WithResource(resource.NewWithAttributes(schemaURL, attrs...)))
		return nil
	}
	for _, entry := range []struct{ key, service, namespace, environment string }{{"caller", f.Service, f.Namespace, f.Environment}, {"other", f.Service, f.OtherNamespace, f.Environment}, {"foreign", f.Service, f.Namespace, f.ForeignEnvironment}, {"peer", f.DependencyService, "peer-http385", f.Environment}} {
		if err := create(entry.key, entry.service, entry.namespace, entry.environment); err != nil {
			return err
		}
	}
	defer func() {
		for _, p := range providers {
			shutdown, c := context.WithTimeout(context.Background(), 5*time.Second)
			_ = p.Shutdown(shutdown)
			c()
		}
	}()
	propagator := propagation.TraceContext{}
	methodAttrs := func(method string) []attribute.KeyValue {
		switch method {
		case "GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE":
			return []attribute.KeyValue{attribute.String("http.request.method", method)}
		default:
			return []attribute.KeyValue{attribute.String("http.request.method", "_OTHER"), attribute.String("http.request.method_original", method)}
		}
	}
	setStatus := func(span trace.Span, status int, client bool) {
		if status >= 500 || client && status >= 400 {
			span.SetStatus(codes.Error, "")
			span.SetAttributes(attribute.String("error.type", strconv.Itoa(status)))
		}
	}
	peer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parent := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		attrs := methodAttrs(r.Method)
		host, port, _ := net.SplitHostPort(r.Host)
		p, _ := strconv.Atoi(port)
		attrs = append(attrs, attribute.String("server.address", host), attribute.Int("server.port", p), attribute.String("url.scheme", "http"), attribute.String("url.path", r.URL.Path), attribute.String("network.protocol.version", "1.1"), attribute.String("http.route", r.URL.Path))
		_, span := providers["peer"].Tracer("goo.journey.http-client385.server", trace.WithSchemaURL(schemaURL)).Start(parent, r.Method+" "+r.URL.Path, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attrs...))
		defer span.End()
		status := 200
		switch r.URL.Path {
		case "/server-error":
			status = 500
		case "/not-found":
			status = 404
		case "/slow":
			time.Sleep(10 * time.Millisecond)
		}
		span.SetAttributes(attribute.Int("http.response.status_code", status))
		setStatus(span, status, false)
		w.Header().Set("X-Goo385-Trace", span.SpanContext().TraceID().String())
		w.Header().Set("X-Goo385-Span", span.SpanContext().SpanID().String())
		w.WriteHeader(status)
		_, _ = fmt.Fprintf(w, "actual HTTP %d\n", status)
	}))
	defer peer.Close()
	peerAddress := peer.Listener.Addr().String()
	_, portText, _ := net.SplitHostPort(peerAddress)
	peerPort, _ := strconv.Atoi(portText)
	transport := &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
		_, port, err := net.SplitHostPort(address)
		if err != nil || port != portText {
			return nil, fmt.Errorf("unexpected workload destination %q", address)
		}
		return (&net.Dialer{}).DialContext(ctx, network, peerAddress)
	}}
	defer transport.CloseIdleConnections()
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	var mu sync.Mutex
	var current scenario
	caller := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		c := current
		mu.Unlock()
		key := c.Scope
		if key == "" {
			key = "caller"
		}
		ns, env := f.Namespace, f.Environment
		if key == "other" {
			ns = f.OtherNamespace
		}
		if key == "foreign" {
			env = f.ForeignEnvironment
		}
		serverAttrs := methodAttrs(r.Method)
		callerHost, callerPortText, _ := net.SplitHostPort(r.Host)
		callerPort, _ := strconv.Atoi(callerPortText)
		serverAttrs = append(serverAttrs, attribute.String("server.address", callerHost), attribute.Int("server.port", callerPort), attribute.String("url.scheme", "http"), attribute.String("url.path", "/call"), attribute.String("network.protocol.version", "1.1"), attribute.String("http.route", "/call"), attribute.String("test.http385.case", c.Label))
		serverCtx, serverSpan := providers[key].Tracer("goo.journey.http-client385.caller", trace.WithSchemaURL(schemaURL)).Start(r.Context(), r.Method+" /call", trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(serverAttrs...))
		defer serverSpan.End()
		if c.DelayBeforeClient > 0 {
			time.Sleep(c.DelayBeforeClient)
		}
		attrs := []attribute.KeyValue{attribute.String("url.full", "http://"+net.JoinHostPort(c.Address, portText)+c.Path), attribute.String("url.scheme", "http"), attribute.String("network.protocol.version", "1.1"), attribute.String("test.http385.case", c.Label)}
		partial := []string{}
		switch {
		case c.MissingMethod:
			partial = append(partial, "http.request.method omitted")
		case c.EmptyMethod:
			attrs = append(attrs, attribute.String("http.request.method", ""))
			partial = append(partial, "http.request.method empty")
		case c.LegacyMethod:
			attrs = append(attrs, attribute.String("http.method", c.Method))
			partial = append(partial, "older method alias only")
		default:
			attrs = append(attrs, methodAttrs(c.Method)...)
		}
		if c.MissingAddress {
			partial = append(partial, "server.address omitted")
		} else if c.EmptyAddress {
			attrs = append(attrs, attribute.String("server.address", ""))
			partial = append(partial, "server.address empty")
		} else {
			attrs = append(attrs, attribute.String("server.address", c.Address))
		}
		attrs = append(attrs, attribute.Int("server.port", peerPort))
		if len(partial) > 0 {
			attrs = append(attrs, attribute.Bool("test.http385.partial_instrumentation", true))
		}
		name := c.Method
		if c.Method == "PURGE" {
			name = "HTTP"
		}
		clientCtx, clientSpan := providers[key].Tracer("goo.journey.http-client385.client", trace.WithSchemaURL(schemaURL)).Start(serverCtx, name, trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attrs...))
		req, err := http.NewRequestWithContext(clientCtx, c.Method, "http://"+net.JoinHostPort(c.Address, portText)+c.Path, nil)
		if err != nil {
			clientSpan.End()
			http.Error(w, err.Error(), 500)
			return
		}
		propagator.Inject(clientCtx, propagation.HeaderCarrier(req.Header))
		response, err := client.Do(req)
		if err != nil {
			clientSpan.RecordError(err)
			clientSpan.SetStatus(codes.Error, "")
			clientSpan.End()
			http.Error(w, err.Error(), 500)
			return
		}
		_, copyErr := io.Copy(io.Discard, response.Body)
		_ = response.Body.Close()
		status := response.StatusCode
		clientSpan.SetAttributes(attribute.Int("http.response.status_code", status))
		setStatus(clientSpan, status, true)
		if copyErr != nil {
			clientSpan.RecordError(copyErr)
			clientSpan.SetStatus(codes.Error, "")
		}
		clientSpan.End()
		serverSpan.SetAttributes(attribute.Int("http.response.status_code", status))
		setStatus(serverSpan, status, false)
		fact := observedCall{Label: c.Label, Method: c.Method, Address: c.Address, Port: peerPort, Status: status, Environment: env, Namespace: ns, TraceID: clientSpan.SpanContext().TraceID().String(), RootSpanID: serverSpan.SpanContext().SpanID().String(), ClientSpanID: clientSpan.SpanContext().SpanID().String(), DependencySpanID: response.Header.Get("X-Goo385-Span"), PartialFields: partial}
		if response.Header.Get("X-Goo385-Trace") != fact.TraceID {
			http.Error(w, "trace propagation mismatch", 500)
			return
		}
		mu.Lock()
		f.Calls = append(f.Calls, fact)
		if c.Label == "empty-window-anchor" {
			f.EmptyRootSpanID = fact.RootSpanID
		}
		mu.Unlock()
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, "observed")
	}))
	defer caller.Close()
	driver := &http.Client{Transport: &http.Transport{Proxy: nil}, Timeout: 10 * time.Second}
	runCase := func(c scenario) error {
		mu.Lock()
		current = c
		mu.Unlock()
		req, err := http.NewRequestWithContext(ctx, c.Method, caller.URL+"/call", nil)
		if err != nil {
			return err
		}
		res, err := driver.Do(req)
		if err != nil {
			return err
		}
		_, _ = io.Copy(io.Discard, res.Body)
		_ = res.Body.Close()
		want := 200
		if c.Path == "/server-error" {
			want = 500
		} else if c.Path == "/not-found" {
			want = 404
		}
		if res.StatusCode != want {
			return fmt.Errorf("actual request %s returned %d, want %d", c.Label, res.StatusCode, want)
		}
		mu.Lock()
		observed := len(f.Calls) > 0 && f.Calls[len(f.Calls)-1].Label == c.Label
		mu.Unlock()
		if !observed {
			return fmt.Errorf("actual request %s has no independently correlated Client/Server fact", c.Label)
		}
		return nil
	}
	cases := []scenario{}
	add := func(n int, label, method, path, address string, opts scenario) {
		for i := range n {
			c := opts
			c.Label = fmt.Sprintf("%s-%d", label, i)
			c.Method = method
			c.Path = path
			c.Address = address
			cases = append(cases, c)
		}
	}
	add(70, "ok-inventory", "GET", "/ok", "inventory.http385.test", scenario{})
	add(30, "ok-billing", "GET", "/ok", "billing.http385.test", scenario{})
	add(5, "server-error", "POST", "/server-error", "billing.http385.test", scenario{})
	add(4, "not-found", "GET", "/not-found", "inventory.http385.test", scenario{})
	add(2, "slow", "GET", "/slow", "inventory.http385.test", scenario{})
	add(2, "missing-address", "GET", "/ok", "inventory.http385.test", scenario{MissingAddress: true})
	add(2, "empty-address", "GET", "/ok", "inventory.http385.test", scenario{EmptyAddress: true})
	add(1, "other-method", "PURGE", "/ok", "inventory.http385.test", scenario{})
	add(2, "missing-method", "GET", "/ok", "inventory.http385.test", scenario{MissingMethod: true})
	add(2, "empty-method", "GET", "/ok", "inventory.http385.test", scenario{EmptyMethod: true})
	add(1, "older-method", "GET", "/ok", "inventory.http385.test", scenario{LegacyMethod: true})
	add(2, "other-namespace", "GET", "/ok", "inventory.http385.test", scenario{Scope: "other"})
	add(2, "foreign-environment", "GET", "/ok", "inventory.http385.test", scenario{Scope: "foreign"})
	for _, c := range cases {
		if err := runCase(c); err != nil {
			return err
		}
	}
	time.Sleep(5 * time.Millisecond)
	if err := runCase(scenario{Label: "empty-window-anchor", Method: "GET", Path: "/ok", Address: "inventory.http385.test", DelayBeforeClient: 25 * time.Millisecond}); err != nil {
		return err
	}
	for _, provider := range providers {
		if err := provider.ForceFlush(ctx); err != nil {
			return err
		}
	}
	f.ToMS = time.Now().Add(time.Second).UnixMilli()
	return json.NewEncoder(os.Stdout).Encode(f)
}

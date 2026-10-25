// goo-read-audit measures the existing record executor without changing its reads.
// Input is a fixed exploration.Query on stdin; output contains no record bodies,
// SQL text, parameters, endpoint, or credentials. Use only a trusted local source.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/casselc/goo/exploration"
)

type read struct {
	Kind          string            `json:"kind"`
	SQLBytes      int               `json:"sqlBytes"`
	SQLSHA256     string            `json:"sqlSHA256"`
	ResponseBytes int64             `json:"responseBytes"`
	Status        int               `json:"status"`
	ElapsedMS     float64           `json:"elapsedMs"`
	Summary       map[string]string `json:"clickhouseSummary,omitempty"`
}

func sourceSummary(raw string) map[string]string {
	var supplied map[string]json.RawMessage
	if json.Unmarshal([]byte(raw), &supplied) != nil {
		return nil
	}
	out := map[string]string{}
	for _, key := range []string{"read_rows", "read_bytes", "written_rows", "written_bytes", "total_rows_to_read", "result_rows", "result_bytes", "elapsed_ns"} {
		var value string
		if json.Unmarshal(supplied[key], &value) != nil {
			continue
		}
		n, err := strconv.ParseUint(value, 10, 64)
		if err == nil && strconv.FormatUint(n, 10) == value {
			out[key] = value
		}
	}
	return out
}

type audit struct {
	base  http.RoundTripper
	kinds map[string]string
	reads []*read
}
type measuredBody struct {
	io.ReadCloser
	item  *read
	start time.Time
}

func (b *measuredBody) Read(p []byte) (int, error) {
	n, e := b.ReadCloser.Read(p)
	b.item.ResponseBytes += int64(n)
	return n, e
}
func (b *measuredBody) Close() error {
	e := b.ReadCloser.Close()
	b.item.ElapsedMS = float64(time.Since(b.start).Microseconds()) / 1000
	return e
}
func (a *audit) RoundTrip(req *http.Request) (*http.Response, error) {
	b, e := io.ReadAll(req.Body)
	if e != nil {
		return nil, e
	}
	req.Body.Close()
	req.Body = io.NopCloser(strings.NewReader(string(b)))
	h := sha256.Sum256(b)
	item := &read{Kind: a.kinds[string(b)], SQLBytes: len(b), SQLSHA256: hex.EncodeToString(h[:])}
	a.reads = append(a.reads, item)
	start := time.Now()
	response, e := a.base.RoundTrip(req)
	if e != nil {
		item.ElapsedMS = float64(time.Since(start).Microseconds()) / 1000
		return nil, e
	}
	item.Status = response.StatusCode
	item.Summary = sourceSummary(response.Header.Get("X-ClickHouse-Summary"))
	response.Body = &measuredBody{response.Body, item, start}
	return response, nil
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
func run() error {
	var q exploration.Query
	input, err := io.ReadAll(io.LimitReader(os.Stdin, (1<<20)+1))
	if err != nil || len(input) > 1<<20 {
		return fmt.Errorf("input exceeds budget or cannot be read")
	}
	d := json.NewDecoder(strings.NewReader(string(input)))
	d.DisallowUnknownFields()
	if e := d.Decode(&q); e != nil {
		return fmt.Errorf("invalid input query: %w", e)
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return fmt.Errorf("require exactly one input query")
	}
	p, e := exploration.Compile(q)
	if e != nil {
		return e
	}
	endpoint := os.Getenv("GOO_AUDIT_ENDPOINT")
	if endpoint == "" {
		return fmt.Errorf("GOO_AUDIT_ENDPOINT is required")
	}
	a := &audit{base: http.DefaultTransport, kinds: map[string]string{p.Count: "count", p.Rows: "rows", p.Groups: "groups", p.Fields: "fields", p.TimeSeries: "timeSeries", p.GroupedTimeSeries: "groupedTimeSeries"}, reads: []*read{}}
	s := exploration.Source{Endpoint: endpoint, User: os.Getenv("GOO_AUDIT_USER"), Password: os.Getenv("GOO_AUDIT_PASSWORD"), Client: &http.Client{Transport: a, Timeout: 15 * time.Second}}
	start := time.Now()
	result, err := s.Fetch(context.Background(), q)
	out := map[string]any{"version": "goo/read-audit/v1", "reads": a.reads, "elapsedMs": float64(time.Since(start).Microseconds()) / 1000, "success": err == nil}
	if result != nil {
		b, _ := json.Marshal(result)
		out["resultBytes"] = len(b)
		out["count"] = result.Count
		out["groups"] = len(result.Groups)
		out["state"] = result.State
	}
	// Source errors can contain URL credentials; keep output deliberately generic.
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if e := encoder.Encode(out); e != nil {
		return e
	}
	if err != nil {
		return fmt.Errorf("source execution failed; see status and completed reads in audit")
	}
	return nil
}

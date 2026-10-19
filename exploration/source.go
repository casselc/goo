package exploration

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Result struct {
	Numeric           *NumericSummary    `json:"numeric,omitempty"`
	GroupedTimeSeries []TimeSeriesGroup  `json:"groupedTimeSeries,omitempty"`
	SpanStatus        *SpanStatusSummary `json:"spanStatus,omitempty"`
	TimeSeries        []TimePoint        `json:"timeSeries,omitempty"`
	Duration          *DurationSummary   `json:"duration,omitempty"`
	Question          Query              `json:"question"`
	CapturedAt        string             `json:"capturedAt"`
	Basis             string             `json:"basis"`
	Count             string             `json:"count"`
	State             string             `json:"state"`
	Rows              []json.RawMessage  `json:"rows"`
	HasMore           bool               `json:"hasMore"`
	Groups            []json.RawMessage  `json:"groups"`
	Fields            []Field            `json:"fields"`
	FieldsLimited     bool               `json:"fieldsLimited"`
	Plan              Plan               `json:"plan"`
	Warnings          []string           `json:"warnings"`
}
type Source struct {
	Endpoint, User, Password string
	Client                   *http.Client
}

// ClickHouse HTTP query parameters use the Escaped text format after URL
// decoding. URL encoding alone loses backslashes and rejects literal tabs.
func encodeParameter(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r == '\\' {
			b.WriteString(`\\`)
		} else if r < 32 || r == 127 {
			fmt.Fprintf(&b, `\x%02x`, r)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func (s Source) read(ctx context.Context, sql string, params map[string]string) ([]json.RawMessage, error) {
	endpoint, e := url.Parse(s.Endpoint)
	if e != nil {
		return nil, e
	}
	v := endpoint.Query()
	v.Set("readonly", "1")
	v.Set("cancel_http_readonly_queries_on_client_close", "1")
	v.Set("max_execution_time", "10")
	v.Set("max_result_rows", "1002")
	v.Set("result_overflow_mode", "throw")
	v.Set("max_memory_usage", "268435456")
	v.Set("max_bytes_to_read", "268435456")
	v.Set("read_overflow_mode", "throw")
	for k, p := range params {
		v.Set("param_"+k, encodeParameter(p))
	}
	endpoint.RawQuery = v.Encode()
	req, e := http.NewRequestWithContext(ctx, "POST", endpoint.String(), strings.NewReader(sql))
	if e != nil {
		return nil, e
	}
	req.SetBasicAuth(s.User, s.Password)
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	resp, e := client.Do(req)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("exploration source HTTP %d", resp.StatusCode)
	}
	b, e := io.ReadAll(io.LimitReader(resp.Body, (8<<20)+1))
	if e != nil {
		return nil, e
	}
	if len(b) > 8<<20 {
		return nil, fmt.Errorf("source exceeds 8 MiB")
	}
	rows := []json.RawMessage{}
	scan := bufio.NewScanner(bytes.NewReader(b))
	scan.Buffer(make([]byte, 4096), 8<<20)
	for scan.Scan() {
		line := bytes.TrimSpace(scan.Bytes())
		if len(line) == 0 {
			continue
		}
		var object map[string]json.RawMessage
		if json.Unmarshal(line, &object) != nil || object == nil {
			return nil, fmt.Errorf("invalid source row")
		}
		rows = append(rows, append(json.RawMessage(nil), line...))
		if len(rows) > 1001 {
			return nil, fmt.Errorf("source row budget exceeded")
		}
	}
	return rows, scan.Err()
}
func (s Source) Fetch(ctx context.Context, q Query) (*Result, error) { return s.fetch(ctx, q, q.Limit) }

// captureRows reads one bounded row bank; its counts/groups retain full scope.
func (s Source) captureRows(ctx context.Context, q Query) (*Result, error) {
	return s.fetch(ctx, q, CaptureRowLimit)
}

func (s Source) fetch(ctx context.Context, q Query, rowLimit int) (*Result, error) {
	plan, e := compile(q, rowLimit)
	if e != nil {
		return nil, e
	}
	r := &Result{Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), Rows: []json.RawMessage{}, Groups: []json.RawMessage{}, Fields: []Field{}, Plan: plan, Warnings: []string{"Captured records only; sampling, ingestion completeness and clock accuracy are unverified.", "Rows, counts, groups and field discovery are separate source reads, not a transactional snapshot. Paging can change as telemetry arrives; counts never derive from the page.", "Attributes are exporter strings; missing keys and present empty strings remain distinct. Counts are records, not unique invocations or business outcomes."}}
	count, e := s.read(ctx, plan.Count, plan.Parameters)
	if e != nil {
		return nil, e
	}
	var total struct {
		Count string `json:"count"`
	}
	if len(count) != 1 || json.Unmarshal(count[0], &total) != nil {
		return nil, fmt.Errorf("invalid count result")
	}
	if _, e = strconv.ParseUint(total.Count, 10, 64); e != nil {
		return nil, fmt.Errorf("invalid count")
	}
	r.Count = total.Count
	if q.Numeric != nil {
		r.Numeric, e = decodeNumeric(count[0], *q.Numeric, r.Count)
		if e != nil {
			return nil, e
		}
	}
	if q.Summary == "span-duration" {
		r.Duration, e = DecodeDurationSummary(count[0])
		if e != nil {
			return nil, e
		}
	}
	if q.SpanStatus {
		r.SpanStatus, e = DecodeSpanStatus(count[0])
		if e != nil {
			return nil, e
		}
	}
	r.State = "observed"
	if r.Count == "0" {
		r.State = "no_data"
	}
	rows, e := s.read(ctx, plan.Rows, plan.Parameters)
	if e != nil {
		return nil, e
	}
	if len(rows) > rowLimit+1 {
		return nil, fmt.Errorf("source exceeded page budget")
	}
	r.HasMore = len(rows) > rowLimit
	if r.HasMore {
		rows = rows[:rowLimit]
	}
	r.Rows = rows
	totalValue, _ := new(big.Int).SetString(r.Count, 10)
	observedMinimum := new(big.Int).SetInt64(int64(q.Offset + len(rows)))
	if len(rows) > 0 && totalValue.Cmp(observedMinimum) < 0 || r.HasMore && totalValue.Cmp(new(big.Int).SetInt64(int64(q.Offset+rowLimit))) <= 0 {
		r.State = "changed_between_reads"
	}
	for _, raw := range rows {
		var row struct {
			Time     string `json:"timestamp_ns"`
			Duration string `json:"duration_ns"`
		}
		if json.Unmarshal(raw, &row) != nil {
			return nil, fmt.Errorf("invalid row shape")
		}
		if _, e := strconv.ParseUint(row.Time, 10, 64); e != nil {
			return nil, fmt.Errorf("invalid row timestamp")
		}
		if q.Signal == "spans" {
			if _, e := strconv.ParseUint(row.Duration, 10, 64); e != nil {
				return nil, fmt.Errorf("invalid span duration")
			}
		}
	}

	if plan.Groups != "" {
		r.Groups, e = s.read(ctx, plan.Groups, plan.Parameters)
		if e != nil {
			return nil, e
		}
		if len(r.Groups) > 1000 {
			return nil, fmt.Errorf("more than 1000 groups; narrow query; no partial grouping totals")
		}
	}
	if plan.Groups != "" {
		sum := new(big.Int)
		durationSum := new(big.Int)
		var durationMin, durationMax *big.Int
		statusTotals := emptySpanStatus()
		seenGroups := map[string]bool{}
		for groupIndex, raw := range r.Groups {
			var group map[string]json.RawMessage
			if json.Unmarshal(raw, &group) != nil {
				return nil, fmt.Errorf("invalid group")
			}
			var count string
			if json.Unmarshal(group["count"], &count) != nil {
				return nil, fmt.Errorf("invalid group count")
			}
			n, ok := new(big.Int).SetString(count, 10)
			if !ok || n.Sign() < 0 || n.BitLen() > 64 {
				return nil, fmt.Errorf("invalid group count")
			}
			sum.Add(sum, n)
			if q.Summary == "span-duration" {
				d, err := DecodeDurationSummary(raw)
				if err != nil || d.State != "observed" {
					return nil, fmt.Errorf("invalid grouped duration summary")
				}
				for _, key := range []string{"sum_ns", "min_ns", "max_ns", "quantiles_ns"} {
					delete(group, key)
				}
				group["duration"], _ = json.Marshal(d)
				v, _ := durationDecimal(d.SumNS, 128)
				durationSum.Add(durationSum, v)
				lo, _ := durationDecimal(d.MinNS, 64)
				hi, _ := durationDecimal(d.MaxNS, 64)
				if durationMin == nil || lo.Cmp(durationMin) < 0 {
					durationMin = lo
				}
				if durationMax == nil || hi.Cmp(durationMax) > 0 {
					durationMax = hi
				}
			}
			if q.SpanStatus {
				status, err := DecodeSpanStatus(raw)
				if err != nil || status.Count == "0" {
					return nil, fmt.Errorf("invalid grouped span status")
				}
				statusTotals.add(*status)
				for _, k := range []string{"status_error", "status_ok", "status_unset", "status_unknown"} {
					delete(group, k)
				}
				group["spanStatus"], _ = json.Marshal(status)
			}
			dimensions := []*string{}
			for i := range q.GroupBy {
				key := fmt.Sprintf("g%d", i)
				v, exists := group[key]
				if !exists {
					return nil, fmt.Errorf("missing group dimension")
				}
				var value *string
				if json.Unmarshal(v, &value) != nil {
					return nil, fmt.Errorf("invalid group dimension")
				}
				dimensions = append(dimensions, value)
			}
			key, _ := json.Marshal(dimensions)
			if seenGroups[string(key)] {
				return nil, fmt.Errorf("duplicate group dimensions")
			}
			seenGroups[string(key)] = true
			if q.Numeric != nil {
				summary, err := decodeNumeric(raw, *q.Numeric, count)
				if err != nil {
					return nil, err
				}
				group["numeric"], _ = json.Marshal(summary)
			}
			if q.Summary == "span-duration" || q.SpanStatus || q.Numeric != nil {
				r.Groups[groupIndex], _ = json.Marshal(group)
			}
		}
		if r.SpanStatus != nil && *r.SpanStatus != statusTotals {
			r.State = "changed_between_reads"
		}
		if r.Duration != nil {
			totalSum, _ := durationDecimal(r.Duration.SumNS, 128)
			if durationSum.Cmp(totalSum) != 0 {
				r.State = "changed_between_reads"
			}
			if durationMin != nil && r.Duration.State == "observed" && (durationMin.String() != r.Duration.MinNS || durationMax.String() != r.Duration.MaxNS) {
				r.State = "changed_between_reads"
			}
		}
		if sum.Cmp(totalValue) != 0 {
			r.State = "changed_between_reads"
		}
	}
	fields, e := s.read(ctx, plan.Fields, plan.Parameters)
	if e != nil {
		return nil, e
	}
	if len(fields) > 129 {
		return nil, fmt.Errorf("source exceeded discovery budget")
	}
	if r.Count == "0" && len(fields) > 0 {
		r.State = "changed_between_reads"
	}
	r.FieldsLimited = len(fields) > 128
	if r.FieldsLimited {
		fields = fields[:128]
	}
	names := []string{}
	for k := range columns {
		if _, e := (Field{"column", k}).Type(q.Signal); e == nil {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	for _, k := range names {
		r.Fields = append(r.Fields, Field{"column", k})
	}
	for _, raw := range fields {
		var v struct {
			Field string `json:"field"`
		}
		if json.Unmarshal(raw, &v) != nil {
			return nil, fmt.Errorf("invalid field result")
		}
		parts := strings.SplitN(v.Field, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid discovered field")
		}
		f := Field{parts[0], parts[1]}
		if _, e = f.Type(q.Signal); e != nil {
			r.FieldsLimited = true
			continue
		}
		r.Fields = append(r.Fields, f)
	}
	if q.BucketMS != 0 {
		series, err := s.read(ctx, plan.TimeSeries, plan.Parameters)
		if err != nil {
			return nil, err
		}
		r.TimeSeries, err = DecodeTimeSeries(q, series)
		if err != nil {
			return nil, err
		}
		sum := new(big.Int)
		statusTotal := emptySpanStatus()
		for _, point := range r.TimeSeries {
			if point.SpanStatus != nil {
				statusTotal.add(*point.SpanStatus)
			}
			n, _ := new(big.Int).SetString(point.Count, 10)
			sum.Add(sum, n)
		}
		if r.SpanStatus != nil && *r.SpanStatus != statusTotal {
			r.State = "changed_between_reads"
		}
		if sum.Cmp(totalValue) != 0 {
			r.State = "changed_between_reads"
		}
	}

	if q.GroupTimeSeries {
		raw, err := s.read(ctx, plan.GroupedTimeSeries, plan.Parameters)
		if err != nil {
			return nil, err
		}
		r.GroupedTimeSeries, err = DecodeGroupedTimeSeries(q, raw)
		if err != nil {
			return nil, err
		}
		if !groupedSeriesAgree(r) {
			r.State = "changed_between_reads"
		}
	}
	if q.Numeric != nil {
		if err := ValidateNumericResult(r); err != nil {
			return nil, err
		}
		if !numericReadsAgree(r) {
			r.State = "changed_between_reads"
		}
	}
	if r.State == "changed_between_reads" {
		r.Warnings = append(r.Warnings, "Source reads disagree; do not treat count, groups and page as one population snapshot. Rerun to capture new evidence.")
	}
	// Timestamp is a receipt, not part of the evidence basis. Hash captured query,
	// plan and results with exact JSON numbers retained by RawMessage.
	basis := *r
	basis.CapturedAt = ""
	b, e := json.Marshal(basis)
	if e != nil {
		return nil, e
	}
	h := sha256.Sum256(b)
	r.Basis = hex.EncodeToString(h[:])
	return r, nil
}

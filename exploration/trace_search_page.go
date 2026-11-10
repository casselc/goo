package exploration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/casselc/goo/identity"
)

const TraceSearchPositionVersion = "goo/trace-search-position/v1"
const TraceSearchPageRequestMaxBytes = TraceSearchRequestMaxBytes + 1024

// TraceSearchPosition is execution position, not authored search intent or a
// permission to execute. Its hash binds the complete normalized request wire.
type TraceSearchPosition struct {
	Version     string `json:"version"`
	RequestHash string `json:"requestHash"`
	LastTraceID string `json:"lastTraceId"`
}
type TraceSearchPageRequest struct {
	Request  TraceSearchRequest  `json:"request"`
	Position TraceSearchPosition `json:"position"`
}

// TraceSearchRequestHash hashes sorted-key JSON after the request's existing
// omitempty normalization. HTML is literal UTF-8; U+2028/U+2029 retain Go's JSON
// escapes. This is a public identity check, not a signed token or source lease.
func TraceSearchRequestHash(r TraceSearchRequest) (string, error) {
	if err := r.Validate(); err != nil {
		return "", err
	}
	raw, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err = decoder.Decode(&value); err != nil {
		return "", err
	}
	var canonical bytes.Buffer
	encoder := json.NewEncoder(&canonical)
	encoder.SetEscapeHTML(false)
	if err = encoder.Encode(value); err != nil {
		return "", err
	}
	sum := sha256.Sum256(bytes.TrimSuffix(canonical.Bytes(), []byte{'\n'}))
	return hex.EncodeToString(sum[:]), nil
}
func ValidateTraceSearchPosition(r TraceSearchRequest, p TraceSearchPosition) error {
	hash, err := TraceSearchRequestHash(r)
	if err != nil {
		return err
	}
	if p.Version != TraceSearchPositionVersion || !regexpRevision.MatchString(p.RequestHash) || p.RequestHash != hash || !exampleID(p.LastTraceID, 32) {
		return errors.New("trace search position differs from its search")
	}
	return nil
}
func NextTraceSearchPosition(r TraceSearchRequest, last string) (*TraceSearchPosition, error) {
	hash, err := TraceSearchRequestHash(r)
	if err != nil {
		return nil, err
	}
	p := &TraceSearchPosition{Version: TraceSearchPositionVersion, RequestHash: hash, LastTraceID: last}
	if err = ValidateTraceSearchPosition(r, *p); err != nil {
		return nil, err
	}
	return p, nil
}
func (r TraceSearchPageRequest) Validate() error {
	raw, err := json.Marshal(r)
	if err != nil || len(raw) > TraceSearchPageRequestMaxBytes {
		return errors.New("trace search page exceeds request budget")
	}
	return ValidateTraceSearchPosition(r.Request, r.Position)
}
func (r *TraceSearchPageRequest) UnmarshalJSON(data []byte) error {
	*r = TraceSearchPageRequest{}
	value, err := strictTraceSearchObject(data, reflect.TypeFor[TraceSearchPageRequest](), []string{"request", "position"})
	if err != nil {
		return err
	}
	type wire TraceSearchPageRequest
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var out wire
	if err = json.Unmarshal(raw, &out); err != nil {
		return err
	}
	*r = TraceSearchPageRequest(out)
	return nil
}
func (p *TraceSearchPosition) UnmarshalJSON(data []byte) error {
	*p = TraceSearchPosition{}
	value, err := strictTraceSearchObject(data, reflect.TypeFor[TraceSearchPosition](), []string{"version", "requestHash", "lastTraceId"})
	if err != nil {
		return err
	}
	type wire TraceSearchPosition
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	var out wire
	if err = json.Unmarshal(raw, &out); err != nil {
		return err
	}
	*p = TraceSearchPosition(out)
	return nil
}
func strictTraceSearchObject(data []byte, t reflect.Type, required []string) (map[string]any, error) {
	if !identity.WellFormedJSON(data) {
		return nil, errors.New("trace search position contains malformed Unicode")
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if decoder.Decode(new(any)) != io.EOF || !strictTraceSearchValue(value, t, 0) {
		return nil, errors.New("trace search page requires known non-null fields")
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("trace search page requires object")
	}
	for _, key := range required {
		if _, ok = object[key]; !ok {
			return nil, fmt.Errorf("trace search page requires %s", key)
		}
	}
	return object, nil
}

// CompileTraceSearchPage retains total-before-position and remaining counts even
// when an intervening source change leaves the page empty. The bounded page is
// left-joined to one metadata row; matched CTE expansion can scan the source more
// than once. One SQL execution does not imply transactionality or a stable walk.
func CompileTraceSearchPage(r TraceSearchPageRequest) (TraceSearchPlan, error) {
	if err := r.Validate(); err != nil {
		return TraceSearchPlan{}, err
	}
	aggregate, err := compileTraceSearchAggregate(r.Request)
	if err != nil {
		return aggregate, err
	}
	aggregate.Parameters["trace_after"] = r.Position.LastTraceID
	selectColumns := []string{"totals.matched_traces AS matched_traces", "totals.remaining_traces AS remaining_traces", "page.trace_id AS trace_id", "page.page_present AS page_present"}
	for i := range r.Request.Conditions {
		n := strconv.Itoa(i)
		selectColumns = append(selectColumns, "page.c"+n+" AS c"+n, "page.w"+n+" AS w"+n)
	}
	aggregate.SQL = "WITH matched AS (" + aggregate.SQL + "), totals AS (SELECT toString(count()) AS matched_traces, toString(countIf(trace_id > {trace_after:String})) AS remaining_traces FROM matched), page AS (SELECT *, toUInt8(1) AS page_present FROM matched WHERE trace_id > {trace_after:String} ORDER BY trace_id LIMIT " + strconv.Itoa(r.Request.Limit+1) + ") SELECT " + strings.Join(selectColumns, ", ") + " FROM totals LEFT JOIN page ON 1 ORDER BY page.trace_id SETTINGS join_use_nulls = 0 FORMAT JSONEachRow"
	if len(aggregate.SQL) > 256<<10 {
		return TraceSearchPlan{}, errors.New("trace search page SQL exceeds budget")
	}
	return aggregate, nil
}
func (s Source) SearchTracePage(ctx context.Context, r TraceSearchPageRequest) (*TraceSearchResult, error) {
	if err := r.Validate(); err != nil {
		return nil, err
	}
	return s.searchTraces(ctx, r.Request, &r.Position)
}
func decodeTraceSearchPageMetadata(rows []json.RawMessage, conditions int) ([]json.RawMessage, string, string, error) {
	if len(rows) == 0 {
		return nil, "", "", errors.New("trace search page metadata missing")
	}
	var total, remaining string
	empty := false
	for i, raw := range rows {
		var object map[string]json.RawMessage
		var n, left string
		if json.Unmarshal(raw, &object) != nil || json.Unmarshal(object["matched_traces"], &n) != nil || json.Unmarshal(object["remaining_traces"], &left) != nil {
			return nil, "", "", errors.New("invalid trace search page counts")
		}
		full, err := decimal(n)
		rest, leftErr := decimal(left)
		if err != nil || leftErr != nil || rest.Cmp(full) > 0 || i > 0 && (n != total || left != remaining) {
			return nil, "", "", errors.New("inconsistent trace search page counts")
		}
		total, remaining = n, left
		var present uint8
		if json.Unmarshal(object["page_present"], &present) != nil || present > 1 {
			return nil, "", "", errors.New("trace search page marker missing")
		}
		if present == 0 {
			if len(rows) != 1 || rest.Sign() != 0 || !bytes.Equal(object["trace_id"], []byte(`""`)) {
				return nil, "", "", errors.New("invalid empty trace page")
			}
			for j := range conditions {
				if !bytes.Equal(object["c"+strconv.Itoa(j)], []byte(`""`)) {
					return nil, "", "", errors.New("empty trace page contains count")
				}
				var tuple []json.RawMessage
				if json.Unmarshal(object["w"+strconv.Itoa(j)], &tuple) != nil || len(tuple) != 8 {
					return nil, "", "", errors.New("empty trace page contains witness")
				}
				for _, v := range tuple[:7] {
					var value string
					if !traceSearchString(v, &value) || value != "" {
						return nil, "", "", errors.New("empty trace page contains witness value")
					}
				}
				var fields []json.RawMessage
				if json.Unmarshal(tuple[7], &fields) != nil || fields == nil || len(fields) != 0 {
					return nil, "", "", errors.New("empty trace page contains field values")
				}
			}
			empty = true
		}
	}
	if empty {
		return []json.RawMessage{}, total, remaining, nil
	}
	return rows, total, remaining, nil
}

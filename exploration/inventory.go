package exploration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/internal/sourcefailure"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// InventoryQuestion discovers observed span names; it does not declare fleet
// membership, request identity, or application-owned operation contracts.
type InventoryQuestion struct {
	// Omitted means all service names; a present empty string is an exact name.
	Service             *string                       `json:"service,omitempty"`
	Namespace           *identity.Namespace           `json:"namespace,omitempty"`
	EnvironmentSelector *identity.EnvironmentSelector `json:"environmentSelector,omitempty"`
	Environment         string                        `json:"environment,omitempty"`
	FromMS              int64                         `json:"fromMs"`
	ToMS                int64                         `json:"toMs"`
}

// UnmarshalJSON keeps the checked selector union closed. Explicit null selectors
// or names must not disappear into zero values and acquire a different scope.
func (q *InventoryQuestion) UnmarshalJSON(data []byte) error {
	type wire InventoryQuestion
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for key := range fields {
		switch key {
		case "service", "namespace", "environment", "environmentSelector", "fromMs", "toMs":
		default:
			return fmt.Errorf("json: unknown field %q", key)
		}
	}
	for _, key := range []string{"service", "environment", "environmentSelector"} {
		if raw, ok := fields[key]; ok && bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("%s must be omitted or non-null", key)
		}
	}
	if raw, ok := fields["service"]; ok && !identity.WellFormedJSON(raw) {
		return fmt.Errorf("invalid service name Unicode")
	}
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*q = InventoryQuestion(value)
	return nil
}
func (q InventoryQuestion) clone() InventoryQuestion {
	if q.Service != nil {
		q.Service = new(*q.Service)
	}
	q.Namespace = q.Namespace.Clone()
	q.EnvironmentSelector = copyEnvironmentSelector(q.EnvironmentSelector)
	return q
}
func (q InventoryQuestion) EnvironmentSelection() (identity.EnvironmentSelection, error) {
	return q.Query().EnvironmentSelection()
}

// CurrentEnvironmentName is a proof of representability by existing native
// definitions, never a fallback for effective selectors of the same name.
func (q InventoryQuestion) CurrentEnvironmentName() (string, bool) {
	s, e := q.EnvironmentSelection()
	if e != nil || !s.CurrentOnly() || s.Scope().Mode != "named" {
		return "", false
	}
	return s.Scope().Value, true
}

func (q InventoryQuestion) Validate() error {
	if q.Service != nil && (!utf8.ValidString(*q.Service) || len(*q.Service) > 128 || strings.ContainsRune(*q.Service, 0)) {
		return fmt.Errorf("invalid service")
	}
	if e := q.Namespace.Validate(); e != nil {
		return e
	}
	return q.Query().Validate()
}
func NamespacePredicate(n *identity.Namespace) *Predicate {
	if n == nil {
		return nil
	}
	p := &Predicate{Field: &Field{Kind: "resource", Key: identity.NamespaceKey}, Op: "eq", Value: n.Value}
	if n.Mode == "missing" {
		p.Op = "missing"
	}
	return p
}
func (q InventoryQuestion) Query() Query {
	v := Query{Signal: "spans", Environment: q.Environment, EnvironmentSelector: copyEnvironmentSelector(q.EnvironmentSelector), FromMS: q.FromMS, ToMS: q.ToMS, Limit: 100}
	if q.Namespace != nil {
		v.Where = NamespacePredicate(q.Namespace)
	}
	if q.Service != nil {
		service := Predicate{Field: &Field{Kind: "column", Key: "service"}, Op: "eq", Value: *q.Service}
		if v.Where == nil {
			v.Where = &service
		} else {
			v.Where = &Predicate{All: []Predicate{*v.Where, service}}
		}
	}
	return v
}

// A selected inventory must never admit another service, including when the
// selected name is empty. Used for fresh source rows and stored-capture admission.
func validateInventoryServiceRows(q InventoryQuestion, rows []InventoryRow) error {
	if q.Service != nil {
		for _, row := range rows {
			if row.Service != *q.Service {
				return fmt.Errorf("inventory row service differs from selected service")
			}
		}
	}
	return nil
}

type InventoryRow struct {
	Service string   `json:"service"`
	Name    string   `json:"name"`
	Count   string   `json:"count"`
	Errors  string   `json:"errors"`
	Buckets []string `json:"buckets"`
}

// UnmarshalJSON preserves reported identity in both fresh source rows and stored
// captures. Missing/null dimensions and lossy Unicode cannot become an empty or
// replacement-character service name before the exact-service proof runs.
func (row *InventoryRow) UnmarshalJSON(data []byte) error {
	if !identity.WellFormedJSON(data) {
		return fmt.Errorf("invalid inventory dimension Unicode")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for _, key := range []string{"service", "name"} {
		var value *string
		if json.Unmarshal(fields[key], &value) != nil || value == nil {
			return fmt.Errorf("missing inventory dimension")
		}
	}
	type wire InventoryRow
	return json.Unmarshal(data, (*wire)(row))
}

type InventoryResult struct {
	Question     InventoryQuestion `json:"question"`
	CapturedAt   string            `json:"capturedAt"`
	Basis        string            `json:"basis"`
	State        string            `json:"state"`
	Count        string            `json:"count"`
	Rows         []InventoryRow    `json:"rows"`
	BucketLabels []string          `json:"bucketLabels"`
	Plan         Plan              `json:"plan"`
	Warnings     []string          `json:"warnings"`
}

func CompileInventory(q InventoryQuestion) (Plan, error) {
	if e := q.Validate(); e != nil {
		return Plan{}, e
	}
	selection, err := q.EnvironmentSelection()
	if err != nil {
		return Plan{}, err
	}
	c := compiler{signal: "spans", params: map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10)}}
	base, _, sourceErr := c.source(q.Query(), selection)
	if sourceErr != nil {
		return Plan{}, sourceErr
	}
	// One aggregate read, disjoint duration buckets and a fail-closed group budget.
	sql := "SELECT ServiceName AS service, SpanName AS name, toString(count()) AS count, toString(countIf(StatusCode = 'Error')) AS errors, ["
	for i, label := range durationLabels {
		if i > 0 {
			sql += ", "
		}
		sql += "toString(countIf(" + durationBucketSQL() + " = '" + label + "'))"
	}
	sql += "] AS buckets" + base + " GROUP BY ServiceName, SpanName ORDER BY ServiceName, SpanName LIMIT 1001 FORMAT JSONEachRow"
	return Plan{Groups: sql, Parameters: c.params}, nil
}
func inventoryNumber(s string) (*big.Int, error) {
	n, e := strconv.ParseUint(s, 10, 64)
	if e != nil || strconv.FormatUint(n, 10) != s {
		return nil, fmt.Errorf("invalid inventory count")
	}
	return new(big.Int).SetUint64(n), nil
}
func (s Source) Inventory(ctx context.Context, q InventoryQuestion) (*InventoryResult, error) {
	q = q.clone()
	plan, e := CompileInventory(q)
	if e != nil {
		return nil, e
	}
	raw, e := s.read(ctx, plan.Groups, plan.Parameters)
	if e != nil {
		return nil, e
	}
	if len(raw) > 1000 {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("more than 1000 service/span-name pairs; narrow the window; no partial inventory"))
	}
	r := &InventoryResult{Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), State: "no_data", Rows: []InventoryRow{}, BucketLabels: append([]string(nil), durationLabels...), Plan: plan, Warnings: []string{"Observed captured spans only; sampling, ingestion completeness and fleet membership are unknown.", "Span names are not declared business operations. Error counts use OTel Error span status, not inferred HTTP status or business failures.", "Duration buckets describe individual spans, not end-to-end requests, root traces or native metric histograms. Drilldowns are separate live reads of the fixed window."}}
	total := new(big.Int)
	seen := map[[2]string]bool{}
	for _, v := range raw {
		var row InventoryRow
		if e = json.Unmarshal(v, &row); e != nil {
			return nil, fmt.Errorf("invalid inventory row")
		}
		if e := validateInventoryServiceRows(q, []InventoryRow{row}); e != nil {
			return nil, e
		}
		count, e := inventoryNumber(row.Count)
		if e != nil || count.Sign() == 0 {
			return nil, fmt.Errorf("invalid inventory population")
		}
		errors, e := inventoryNumber(row.Errors)
		if e != nil || errors.Cmp(count) > 0 {
			return nil, fmt.Errorf("invalid inventory errors")
		}
		if len(row.Buckets) != len(durationLabels) {
			return nil, fmt.Errorf("invalid duration partition")
		}
		sum := new(big.Int)
		for _, v := range row.Buckets {
			n, e := inventoryNumber(v)
			if e != nil {
				return nil, e
			}
			sum.Add(sum, n)
		}
		if sum.Cmp(count) != 0 {
			return nil, fmt.Errorf("duration partition disagrees with population")
		}
		key := [2]string{row.Service, row.Name}
		if seen[key] {
			return nil, fmt.Errorf("duplicate inventory pair")
		}
		seen[key] = true
		total.Add(total, count)
		r.Rows = append(r.Rows, row)
	}
	if total.BitLen() > 64 {
		return nil, fmt.Errorf("inventory population overflow")
	}
	r.Count = total.String()
	if total.Sign() > 0 {
		r.State = "observed"
	}
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

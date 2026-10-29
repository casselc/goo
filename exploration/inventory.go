package exploration

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"math/big"
	"strconv"
	"time"
)

// InventoryQuestion discovers observed span names; it does not declare fleet
// membership, request identity, or application-owned operation contracts.
type InventoryQuestion struct {
	Namespace   *identity.Namespace `json:"namespace,omitempty"`
	Environment string              `json:"environment"`
	FromMS      int64               `json:"fromMs"`
	ToMS        int64               `json:"toMs"`
}

func (q InventoryQuestion) Validate() error {
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
	v := Query{Signal: "spans", Environment: q.Environment, FromMS: q.FromMS, ToMS: q.ToMS, Limit: 100}
	if q.Namespace != nil {
		v.Where = NamespacePredicate(q.Namespace)
	}
	return v
}

type InventoryRow struct {
	Service string   `json:"service"`
	Name    string   `json:"name"`
	Count   string   `json:"count"`
	Errors  string   `json:"errors"`
	Buckets []string `json:"buckets"`
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
	selection, err := identity.CurrentEnvironment(q.Environment)
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
	plan, e := CompileInventory(q)
	if e != nil {
		return nil, e
	}
	raw, e := s.read(ctx, plan.Groups, plan.Parameters)
	if e != nil {
		return nil, e
	}
	if len(raw) > 1000 {
		return nil, fmt.Errorf("more than 1000 service/span-name pairs; narrow the window; no partial inventory")
	}
	r := &InventoryResult{Question: q, CapturedAt: time.Now().UTC().Format(time.RFC3339Nano), State: "no_data", Rows: []InventoryRow{}, BucketLabels: append([]string(nil), durationLabels...), Plan: plan, Warnings: []string{"Observed captured spans only; sampling, ingestion completeness and fleet membership are unknown.", "Span names are not declared business operations. Error counts use OTel Error span status, not inferred HTTP status or business failures.", "Duration buckets describe individual spans, not end-to-end requests, root traces or native metric histograms. Drilldowns are separate live reads of the fixed window."}}
	total := new(big.Int)
	seen := map[[2]string]bool{}
	for _, v := range raw {
		var dimensions map[string]json.RawMessage
		if json.Unmarshal(v, &dimensions) != nil {
			return nil, fmt.Errorf("invalid inventory dimensions")
		}
		for _, key := range []string{"service", "name"} {
			var value *string
			if json.Unmarshal(dimensions[key], &value) != nil || value == nil {
				return nil, fmt.Errorf("missing inventory dimension")
			}
		}
		var row InventoryRow
		if e = json.Unmarshal(v, &row); e != nil {
			return nil, fmt.Errorf("invalid inventory row")
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

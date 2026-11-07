package verification

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"reflect"
	"strconv"
	"testing"

	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"hegel.dev/go/hegel"
)

// Exercise actual JS plans, then admit them through Go and compare membership
// against raw facts rather than invoking the builders as their own oracle.
func TestHegelTypedErrorTriage370(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		pick := func() *string {
			switch hegel.Draw(tc, hegel.Integers(0, 2)) {
			case 0:
				return nil
			case 1:
				value := ""
				return &value
			default:
				value := "literal-Ω雪-" + strconv.Itoa(hegel.Draw(tc, hegel.Integers(0, 1000)))
				return &value
			}
		}
		service, typ, namespace := pick(), pick(), pick()
		mode := hegel.Draw(tc, hegel.Integers(0, 3))
		selector := &identity.EnvironmentSelector{Version: identity.EnvironmentSelectorVersion, Policy: "effective", Scope: identity.EnvironmentScope{Mode: "named", Value: "prod"}}
		if mode == 0 {
			selector.Policy = "current-only"
		} else if mode == 2 {
			selector.Scope = identity.EnvironmentScope{Mode: "empty"}
		} else if mode == 3 {
			selector.Scope = identity.EnvironmentScope{Mode: "unreported"}
		}
		signal := []string{"logs", "spans"}[hegel.Draw(tc, hegel.Integers(0, 1))]
		leaf := func(kind, key, op, value string) exploration.Predicate {
			return exploration.Predicate{Field: &exploration.Field{Kind: kind, Key: key}, Op: op, Value: value}
		}
		ns := leaf("resource", "service.namespace", "missing", "")
		if namespace != nil {
			ns = leaf("resource", "service.namespace", "eq", *namespace)
		}
		q := exploration.Query{Signal: signal, EnvironmentSelector: selector, FromMS: 1000, ToMS: 61000, Limit: 17, Offset: 9, Where: &exploration.Predicate{All: []exploration.Predicate{ns, {Any: []exploration.Predicate{leaf("attribute", "tenant", "eq", "inside"), leaf("attribute", "tenant", "eq", "inside-2")}}}}}
		encoded, err := json.Marshal(map[string]any{"parent": q, "group": map[string]any{"g0": service, "g1": typ}})
		if err != nil {
			tc.Fatal(err)
		}
		cmd := exec.CommandContext(t.Context(), "node", "typed_error_triage_370_fixture.mjs")
		cmd.Stdin = bytes.NewReader(encoded)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatalf("actual typed triage fixture: %v\n%s", err, out)
		}
		var actual struct {
			Triage, Inspect, Compare exploration.Query
			ParentUnchanged          bool
		}
		if err = json.Unmarshal(out, &actual); err != nil || !actual.ParentUnchanged {
			tc.Fatalf("unfinished or mutating fixture: %v\n%s", err, out)
		}
		for _, child := range []exploration.Query{actual.Triage, actual.Inspect, actual.Compare} {
			if err = child.Validate(); err != nil {
				tc.Fatal(err)
			}
			if !reflect.DeepEqual(child.EnvironmentSelector, selector) || child.Signal != q.Signal || child.FromMS != q.FromMS || child.ToMS != q.ToMS || child.Offset != 0 || child.Limit != 100 || child.Where == nil || len(child.Where.All) < 2 || !reflect.DeepEqual(child.Where.All[0], *q.Where) {
				tc.Fatal("whole typed parent authority changed")
			}
		}
		for _, query := range []exploration.Query{actual.Inspect, actual.Compare} {
			if !reflect.DeepEqual(query.Where.All[:2], actual.Triage.Where.All) || !reflect.DeepEqual(query.GroupBy, actual.Triage.GroupBy[1:]) || query.BucketMS != actual.Triage.BucketMS {
				tc.Fatal("triage evidence or grouping changed")
			}
		}
		text := func(value string) *string { return &value }
		choices := []*string{nil, text(""), text("foreign"), service, typ, namespace}
		for _, rowService := range choices {
			for _, rowType := range choices {
				for _, rowNamespace := range choices {
					for _, allowedTenant := range []bool{false, true} {
						for _, reported := range []bool{false, true} {
							values := map[exploration.Field]*string{{Kind: "column", Key: "service"}: rowService, {Kind: "resource", Key: "service.namespace"}: rowNamespace, {Kind: "attribute", Key: "tenant"}: text("outside")}
							if allowedTenant {
								values[exploration.Field{Kind: "attribute", Key: "tenant"}] = text("inside")
							}
							if signal == "spans" {
								values[exploration.Field{Kind: "column", Key: "status_class"}] = text("ok")
								if reported {
									values[exploration.Field{Kind: "column", Key: "status_class"}] = text("error")
								}
								values[exploration.Field{Kind: "column", Key: "exception_type"}] = rowType
							} else {
								values[exploration.Field{Kind: "column", Key: "severity_number"}] = text("9")
								values[exploration.Field{Kind: "column", Key: "severity"}] = text("INFO")
								if reported {
									values[exploration.Field{Kind: "column", Key: "severity_number"}] = text("17")
								}
								values[exploration.Field{Kind: "attribute", Key: "exception.type"}] = rowType
							}
							base := reflect.DeepEqual(rowNamespace, namespace) && allowedTenant && reported
							if triagePredicate370(actual.Triage.Where, values) != base || triagePredicate370(actual.Compare.Where, values) != (base && reflect.DeepEqual(rowService, service)) || triagePredicate370(actual.Inspect.Where, values) != (base && reflect.DeepEqual(rowService, service) && reflect.DeepEqual(rowType, typ)) {
								tc.Fatal("independent typed triage membership disagrees")
							}
						}
					}
				}
			}
		}
	}, options(40)...)
}

func triagePredicate370(p *exploration.Predicate, values map[exploration.Field]*string) bool {
	if p == nil {
		return true
	}
	if len(p.All) > 0 {
		for i := range p.All {
			if !triagePredicate370(&p.All[i], values) {
				return false
			}
		}
		return true
	}
	if len(p.Any) > 0 {
		for i := range p.Any {
			if triagePredicate370(&p.Any[i], values) {
				return true
			}
		}
		return false
	}
	v := values[*p.Field]
	if p.Op == "missing" {
		return v == nil
	}
	if v == nil {
		return false
	}
	switch p.Op {
	case "eq":
		return *v == p.Value
	case "ne":
		return *v != p.Value
	case "gte", "lte":
		x, e := strconv.Atoi(*v)
		y, f := strconv.Atoi(p.Value)
		return e == nil && f == nil && (p.Op == "gte" && x >= y || p.Op == "lte" && x <= y)
	default:
		return false
	}
}

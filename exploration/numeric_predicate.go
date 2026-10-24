package exploration

import "fmt"

func validateNumericPredicate(p *Predicate, signal string) error {
	if p.Numeric == nil {
		return fmt.Errorf("numeric predicate requires binding")
	}
	if e := p.Numeric.Validate(signal); e != nil {
		return e
	}
	switch p.Op {
	case "valid", "missing", "empty", "invalid":
		if p.Value != "" {
			return fmt.Errorf("numeric state predicate has no threshold")
		}
	case "eq", "ne", "gt", "gte", "lt", "lte":
		if _, ok := parseNumericCoefficient(p.Value, *p.Numeric); !ok {
			return fmt.Errorf("numeric threshold requires canonical in-range %s decimal", p.Numeric.Type)
		}
	default:
		return fmt.Errorf("unsupported numeric predicate operator")
	}
	return nil
}

// matchNumericPredicate never treats missing, empty or invalid strings as zero.
// Inequality, like every comparison, applies only to valid parsed values.
func matchNumericPredicate(p *Predicate, values map[Field]string) bool {
	if validateNumericPredicate(p, "spans") != nil {
		return false
	}
	raw, present := values[p.Numeric.Field]
	value, valid := parseNumericCoefficient(raw, *p.Numeric)
	switch p.Op {
	case "missing":
		return !present
	case "empty":
		return present && raw == ""
	case "invalid":
		return present && raw != "" && !valid
	case "valid":
		return present && valid
	}
	if !present || !valid {
		return false
	}
	threshold, _ := parseNumericCoefficient(p.Value, *p.Numeric)
	order := value.Cmp(threshold)
	switch p.Op {
	case "eq":
		return order == 0
	case "ne":
		return order != 0
	case "gt":
		return order > 0
	case "gte":
		return order >= 0
	case "lt":
		return order < 0
	case "lte":
		return order <= 0
	}
	return false
}

func (c *compiler) numericPredicate(p *Predicate) string {
	raw, present, value := c.numericValueSQL(*p.Numeric)
	switch p.Op {
	case "missing":
		return "NOT (" + present + ")"
	case "empty":
		return "(" + present + " AND " + raw + " = '')"
	case "invalid":
		return "(" + present + " AND " + raw + " != '' AND isNull(" + value + "))"
	case "valid":
		return "isNotNull(" + value + ")"
	}
	typ := "Int64"
	if p.Numeric.Type == "uint64" {
		typ = "UInt64"
	}
	threshold := p.Value
	if p.Numeric.Type == "decimal64" {
		coefficient, _ := parseNumericCoefficient(p.Value, *p.Numeric)
		threshold = coefficient.String()
	}
	parameter := c.param(threshold, typ)
	op := map[string]string{"eq": "=", "ne": "!=", "gt": ">", "gte": ">=", "lt": "<", "lte": "<="}[p.Op]
	return "ifNull((" + value + " " + op + " " + parameter + "), 0)"
}

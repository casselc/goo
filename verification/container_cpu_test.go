package verification

import (
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Parse coefficient/exponent and compare exact IEEE rounding boundaries. This
// does not reuse the composer, its decimal formatter or ParseFloat eligibility.
func cpDecimal(raw string) (*big.Rat, string, bool) {
	if len(raw) > 128 || !cmDecimal.MatchString(raw) {
		return nil, "", false
	}
	mantissa := raw
	exponent := 0
	if i := strings.IndexAny(raw, "eE"); i >= 0 {
		var e error
		exponent, e = strconv.Atoi(raw[i+1:])
		if e != nil || exponent < -400 || exponent > 400 {
			return nil, "", false
		}
		mantissa = raw[:i]
	}
	negative := strings.HasPrefix(mantissa, "-")
	mantissa = strings.TrimPrefix(strings.TrimPrefix(mantissa, "+"), "-")
	parts := strings.Split(mantissa, ".")
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	digits := strings.TrimLeft(parts[0]+fraction, "0")
	if digits == "" {
		return new(big.Rat), "0", true
	}
	if negative {
		return nil, "", false
	}
	coefficient, ok := new(big.Int).SetString(digits, 10)
	if !ok {
		return nil, "", false
	}
	scale := len(fraction) - exponent
	value := new(big.Rat).SetInt(coefficient)
	normalized := ""
	if scale <= 0 {
		normalized = digits + strings.Repeat("0", -scale)
		value.Mul(value, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(-scale)), nil)))
	} else {
		value.Quo(value, new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(scale)), nil)))
		if scale >= len(digits) {
			normalized = "0." + strings.Repeat("0", scale-len(digits)) + digits
		} else {
			normalized = digits[:len(digits)-scale] + "." + digits[len(digits)-scale:]
		}
		normalized = strings.TrimRight(strings.TrimRight(normalized, "0"), ".")
	}
	if len(normalized) > 512 {
		return nil, "", false
	}
	halfMin := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 1075))
	overflow := new(big.Int).Lsh(new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 54), big.NewInt(1)), 970)
	if value.Cmp(halfMin) <= 0 || value.Cmp(new(big.Rat).SetInt(overflow)) >= 0 {
		return nil, "", false
	}
	return value, normalized, true
}

var cpScientific = regexp.MustCompile(`^~[0-9](\.[0-9]{1,3})?e\+[0-9]{2,4}%$`)

func cpTextOK(value *big.Rat, text string) bool {
	if value.Sign() > 0 && value.Cmp(big.NewRat(1, 10)) < 0 {
		return text == "<0.1%"
	}
	fixed := cmPercentText(value)
	if len(fixed) <= 24 {
		return text == fixed
	}
	if !cpScientific.MatchString(text) {
		return false
	}
	reported, ok := new(big.Rat).SetString(strings.TrimSuffix(strings.TrimPrefix(text, "~"), "%"))
	if !ok {
		return false
	}
	error := new(big.Rat).Abs(new(big.Rat).Sub(reported, value))
	error.Mul(error, big.NewRat(2000, 1))
	return error.Cmp(value) <= 0
}
func cpQuestion(role string, resource map[string]string) measurements.Question {
	mapped := role
	if role == "usage" {
		mapped = "working-set"
	}
	q := cmQuestion(mapped, resource)
	d := &q.Definition
	d.Name = "container-cpu-" + role
	d.Source.Unit = "{cpu}"
	if role == "usage" {
		d.Source.Metric = "container.cpu.usage"
	} else {
		d.Source.Metric = "k8s.container.cpu_" + role
	}
	d.Source.ExactStream.Name = d.Source.Metric
	d.Source.ExactStream.Unit = "{cpu}"
	return q
}
func cpResult(q measurements.Question, value string, at int64, flags int, conflict, empty bool) (*measurements.Result, error) {
	row := hrRow(q, value)
	row["end_ns"] = strconv.FormatInt(at, 10)
	row["flags"] = flags
	row["temporality"] = 0
	row["monotonic"] = false
	b, _ := json.Marshal(row)
	rows := []json.RawMessage{b}
	if conflict {
		row["value"] = "3.75"
		b, _ = json.Marshal(row)
		rows = append(rows, b)
	}
	if empty {
		rows = nil
	}
	clock := int64(200)
	if strings.HasSuffix(q.Definition.Source.Metric, "_request") {
		clock = 201
	} else if strings.HasSuffix(q.Definition.Source.Metric, "_limit") {
		clock = 202
	}
	return measurements.Evaluate(q, rows, time.Unix(clock, 0), "", nil)
}
func TestHegelContainerCPUExportedDecimals(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		a := hegel.Draw(tc, hegel.Integers(1, 2000))
		b := hegel.Draw(tc, hegel.Integers(1, 2000))
		qualifier := hegel.Draw(tc, hegel.Integers(0, 2))
		usageText := strconv.Itoa(a) + "e-3"
		denomText := strconv.Itoa(b) + "e-3"
		resource := map[string]string{"deployment.environment.name": "prod", "container.id": "literal-container", "k8s.pod.uid": "literal-pod", "k8s.container.name": "worker"}
		for _, k := range []string{"k8s.cluster.uid", "k8s.cluster.name", "k8s.namespace.name"} {
			if qualifier == 1 {
				resource[k] = ""
			} else if qualifier == 2 {
				resource[k] = "reported"
			}
		}
		parent := pnQuestion("literal-pod")
		parent.Resource = resource
		schema := "reported-schema"
		parent.ResourceSchemaURL = &schema
		modes := []string{"aligned", "zero-usage", "zero-denominator", "absent-denominator", "negative", "nonshortest", "scientific", "tiny", "underflow", "large", "beyond-byte-integer", "stale", "missing", "conflict", "no-data", "different-time", "different-resource", "different-schema", "different-attrs", "different-freshness", "different-id", "qualifier-presence", "cpu-time"}
		for _, mode := range modes {
			n, d := usageText, denomText
			nAt, dAt := int64(159000000000), int64(159000000000)
			flags := 0
			conflict, empty, refuse := false, false, false
			nq, dq := cpQuestion("usage", resource), cpQuestion("request", resource)
			switch mode {
			case "zero-usage":
				n = "-0.000"
			case "zero-denominator":
				d = "0"
			case "negative":
				n = "-0.125"
			case "nonshortest":
				n = "0.10000000000000000001"
			case "scientific":
				n = "1.2500e+1"
			case "tiny":
				n = "4e-324"
			case "underflow":
				n = "1e-400"
			case "large":
				n = "1e300"
			case "beyond-byte-integer":
				n = "9007199254740992.5"
			case "stale":
				nAt = 150000000000
			case "missing":
				flags = 1
			case "conflict":
				conflict = true
			case "no-data":
				empty = true
			case "different-time":
				dAt++
			case "different-resource":
				dq.Definition.Source.ExactStream.Resource["collector.variant"] = "other"
			case "different-schema":
				dq.Definition.Source.ExactStream.ResourceSchemaURL = "other"
			case "different-attrs":
				dq.Definition.Source.ExactStream.Attributes["dimension"] = "other"
			case "different-freshness":
				dq.Definition.MaxGapMS++
			case "different-id":
				dq.Definition.Source.Resource.ID = "replacement"
				dq.Definition.Source.ExactStream.Resource["container.id"] = "replacement"
				refuse = true
			case "qualifier-presence":
				if _, ok := resource["k8s.cluster.uid"]; ok {
					delete(dq.Definition.Source.ExactStream.Resource, "k8s.cluster.uid")
				} else {
					dq.Definition.Source.ExactStream.Resource["k8s.cluster.uid"] = ""
				}
				refuse = true
			case "cpu-time":
				nq.Definition.Source.Metric = "container.cpu.time"
				nq.Definition.Source.Unit = "s"
				nq.Definition.Source.ExactStream.Name = "container.cpu.time"
				nq.Definition.Source.ExactStream.Unit = "s"
				refuse = true
			}
			usage, e := cpResult(nq, n, nAt, flags, conflict, empty)
			if e != nil {
				tc.Errorf("raw usage %s: %v", mode, e)
				return
			}
			request, e := cpResult(dq, d, dAt, 0, false, false)
			if e != nil {
				tc.Errorf("raw request %s: %v", mode, e)
				return
			}
			input := measurements.ContainerCPUViewRequest{Parent: parent, Usage: usage, Request: request}
			if mode == "absent-denominator" {
				input.Request = nil
			}
			before, _ := json.Marshal(input)
			view, e := measurements.ComposeContainerCPUView(input)
			after, _ := json.Marshal(input)
			if string(before) != string(after) {
				tc.Errorf("mutated raw input %s", mode)
				return
			}
			if refuse {
				if e == nil {
					tc.Errorf("authority/profile accepted %s", mode)
				}
				continue
			}
			if e != nil {
				tc.Errorf("compose %s: %v", mode, e)
				return
			}
			if len(view.Signals) != 3 || view.Signals[0].Role != "usage" || view.Signals[1].Role != "request" || view.Signals[2].Role != "limit" {
				tc.Errorf("role set %s", mode)
				return
			}
			nr, normalized, valid := cpDecimal(n)
			dr, _, denomValid := cpDecimal(d)
			known := valid && denomValid && dr.Sign() > 0
			switch mode {
			case "absent-denominator", "stale", "missing", "conflict", "no-data", "different-time", "different-resource", "different-schema", "different-attrs", "different-freshness":
				known = false
			}
			if mode != "missing" && mode != "conflict" && mode != "no-data" {
				if valid {
					if view.Signals[0].Value != normalized {
						tc.Errorf("reported decimal changed %s got%s want%s", mode, view.Signals[0].Value, normalized)
						return
					}
				} else if view.Signals[0].State != "invalid" || view.Signals[0].Value != "" || view.Signals[0].ObservationState != "observed" {
					tc.Errorf("invalid decimal became cores %s", mode)
					return
				}
			}
			p := view.Percentages.Request
			if known {
				expected := new(big.Rat).Mul(new(big.Rat).Quo(nr, dr), big.NewRat(100, 1))
				if len(expected.RatString()) > 1024 || len(cmPercentText(expected)) > 1024 {
					if p.State != "unavailable" || p.Reason != "bounded-percentage-domain" {
						tc.Errorf("output budget")
						return
					}
				} else if p.State != "observed" || p.Value != expected.RatString() || !cpTextOK(expected, p.Text) {
					tc.Errorf("exact reported percentage %s %+v", mode, p)
					return
				}
			} else if p.State != "unavailable" || p.Value != "" || p.Text != "" {
				tc.Errorf("uncertified percentage %s %+v", mode, p)
				return
			}
			if view.Signals[0].ReadAt != usage.ReadAt || view.Signals[0].Basis != usage.Basis || view.Percentages.Limit.State != "unavailable" {
				tc.Errorf("independent read/absent limit %s", mode)
				return
			}
			if input.Request != nil && (view.Signals[1].ReadAt != request.ReadAt || view.Signals[1].Basis != request.Basis) {
				tc.Errorf("request provenance %s", mode)
				return
			}
			if mode == "aligned" {
				limit, err := cpResult(cpQuestion("limit", resource), d, dAt, 0, false, false)
				if err != nil {
					tc.Errorf("limit setup: %v", err)
					return
				}
				trio := input
				trio.Limit = limit
				full, err := measurements.ComposeContainerCPUView(trio)
				if err != nil || full.Percentages.Limit.Value != p.Value || full.Signals[2].ReadAt != limit.ReadAt || full.Signals[2].Basis != limit.Basis {
					tc.Errorf("separate limit role %v", err)
					return
				}
			}
		}
	}, options(8)...)
}

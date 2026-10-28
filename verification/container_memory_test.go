package verification

import (
	"encoding/json"
	"github.com/casselc/goo/measurements"
	"hegel.dev/go/hegel"
	"math/big"
	"strconv"
	"testing"
	"time"
)

func cmQuestion(role string, resource map[string]string) measurements.Question {
	q := hrQuestion("unused", "gauge", "missing")
	d := &q.Definition
	d.Name = "container-memory-" + role
	d.MaxGapMS = 5000
	d.Source.Resource = &measurements.ResourceBinding{Kind: "kubernetes-container", ID: resource["container.id"]}
	d.Source.Services = []string{}
	d.Source.Namespace = nil
	d.Source.SchemaURL = ""
	d.Source.Unit = "By"
	d.Source.Version = "0.137.0"
	if role == "working-set" {
		d.Source.Metric = "container.memory.working_set"
		d.Source.Scope = measurements.KubeletStatsScope
	} else {
		d.Source.Metric = "k8s.container.memory_" + role
		d.Source.Scope = "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/k8sclusterreceiver"
	}
	s := d.Source.ExactStream
	s.Resource = map[string]string{}
	for k, v := range resource {
		s.Resource[k] = v
	}
	s.ResourceSchemaURL = "reported-schema"
	s.Name = d.Source.Metric
	s.Scope = d.Source.Scope
	s.Version = d.Source.Version
	s.SchemaURL = ""
	s.Unit = "By"
	s.Attributes = map[string]string{}
	return q
}
func cmResult(q measurements.Question, value string, at int64, flags int, conflict bool, empty bool) (*measurements.Result, error) {
	row := hrRow(q, value)
	row["end_ns"] = strconv.FormatInt(at, 10)
	row["flags"] = flags
	row["temporality"] = 0
	row["monotonic"] = false
	b, _ := json.Marshal(row)
	rows := []json.RawMessage{b}
	if conflict {
		row["value"] = "1234567"
		z, _ := json.Marshal(row)
		rows = append(rows, z)
	}
	if empty {
		rows = nil
	}
	clock := int64(200)
	if q.Definition.Source.Metric == "k8s.container.memory_request" {
		clock = 201
	} else if q.Definition.Source.Metric == "k8s.container.memory_limit" {
		clock = 202
	}
	return measurements.Evaluate(q, rows, time.Unix(clock, 0))
}
func TestHegelContainerMemoryIndependentAlignment(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		numerator := hegel.Draw(tc, hegel.Integers(0, 10000))
		denominator := hegel.Draw(tc, hegel.Integers(1, 10000))
		qualifier := hegel.Draw(tc, hegel.Integers(0, 2))
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
		modes := []string{"aligned", "zero-numerator", "zero-denominator", "absent-denominator", "no-data", "stale", "missing", "conflict", "different-time", "different-resource", "different-schema", "different-attrs", "different-freshness", "fraction", "stale-fraction", "rounded-fraction", "negative", "unsafe", "different-id", "different-pod", "different-name", "qualifier-presence"}
		for _, mode := range modes {
			n, d := strconv.Itoa(numerator), strconv.Itoa(denominator)
			nAt, dAt := int64(159000000000), int64(159000000000)
			flags := 0
			conflict, empty := false, false
			nq, dq := cmQuestion("working-set", resource), cmQuestion("request", resource)
			ratioKnown := mode == "aligned" || mode == "zero-numerator"
			refuse := false
			switch mode {
			case "zero-numerator":
				n = "0"
			case "zero-denominator":
				d = "0"
			case "no-data":
				empty = true
			case "stale":
				dAt = 150000000000
			case "missing":
				flags = 1
			case "conflict":
				conflict = true
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
			case "fraction":
				n = "1.5"
			case "stale-fraction":
				n = "1.5"
				nAt = 150000000000
			case "rounded-fraction":
				n = "1.0000000000000000000000000001"
			case "negative":
				n = "-1"
			case "unsafe":
				n = "9007199254740992"
			case "different-id":
				dq.Definition.Source.Resource.ID = "replacement"
				dq.Definition.Source.ExactStream.Resource["container.id"] = "replacement"
				refuse = true
			case "different-pod":
				dq.Definition.Source.ExactStream.Resource["k8s.pod.uid"] = "other"
				refuse = true
			case "different-name":
				dq.Definition.Source.ExactStream.Resource["k8s.container.name"] = "other"
				refuse = true
			case "qualifier-presence":
				if _, ok := resource["k8s.cluster.uid"]; ok {
					delete(dq.Definition.Source.ExactStream.Resource, "k8s.cluster.uid")
				} else {
					dq.Definition.Source.ExactStream.Resource["k8s.cluster.uid"] = ""
				}
				refuse = true
			}
			used, e := cmResult(nq, n, nAt, 0, false, false)
			if e != nil {
				tc.Errorf("numerator %s: %v", mode, e)
				return
			}
			denom, e := cmResult(dq, d, dAt, flags, conflict, empty)
			if e != nil {
				tc.Errorf("denominator %s: %v", mode, e)
				return
			}
			input := measurements.ContainerMemoryViewRequest{Parent: parent, WorkingSet: used, Request: denom}
			if mode == "absent-denominator" {
				input.Request = nil
			}
			before, _ := json.Marshal(input)
			view, e := measurements.ComposeContainerMemoryView(input)
			after, _ := json.Marshal(input)
			if string(before) != string(after) {
				tc.Errorf("input mutation %s", mode)
				return
			}
			if refuse {
				if e == nil {
					tc.Errorf("context mismatch accepted %s", mode)
				}
				continue
			}
			if e != nil {
				tc.Errorf("compose %s: %v", mode, e)
				return
			}
			if len(view.Signals) != 3 || view.Signals[0].Role != "working-set" || view.Signals[1].Role != "request" || view.Signals[2].Role != "limit" {
				tc.Errorf("fixed signal set %s", mode)
				return
			}
			switch mode {
			case "fraction", "stale-fraction", "rounded-fraction", "negative", "unsafe":
				if view.Signals[0].State != "invalid" || view.Signals[0].Value != "" {
					tc.Errorf("raw invalid bytes exposed %s", mode)
					return
				}
				expectedQuality := "observed"
				if mode == "stale-fraction" {
					expectedQuality = "stale"
				}
				if view.Signals[0].ObservationState != expectedQuality {
					tc.Errorf("domain erased observation quality %s", mode)
					return
				}
			case "zero-numerator":
				if view.Signals[0].Value != "0" {
					tc.Errorf("reported zero lost")
					return
				}
			case "stale":
				if view.Signals[1].State != "stale" || view.Signals[1].Value != d {
					tc.Errorf("stale evidence erased")
					return
				}
			case "missing", "conflict":
				if view.Signals[1].State != "unknown" || view.Signals[1].Value != "" {
					tc.Errorf("missing/conflict fabricated value %s", mode)
					return
				}
			}
			p := view.Percentages.Request
			if ratioKnown {
				nn, _ := new(big.Int).SetString(n, 10)
				dd, _ := new(big.Int).SetString(d, 10)
				expected := new(big.Rat).SetFrac(new(big.Int).Mul(nn, big.NewInt(100)), dd)
				if p.State != "observed" || p.Value != expected.RatString() {
					tc.Errorf("independent ratio %s: %+v", mode, p)
					return
				}
			} else if p.State != "unavailable" || p.Value != "" || p.Text != "" {
				tc.Errorf("unavailable fabricated percentage %s: %+v", mode, p)
				return
			}
			if mode == "aligned" {
				limit, err := cmResult(cmQuestion("limit", resource), d, dAt, 0, false, false)
				if err != nil {
					tc.Errorf("limit setup: %v", err)
					return
				}
				trio := input
				trio.Limit = limit
				full, err := measurements.ComposeContainerMemoryView(trio)
				if err != nil || full.Percentages.Limit.State != "observed" || full.Percentages.Limit.Value != p.Value || full.Signals[2].Value != d || full.Signals[2].ReadAt != limit.ReadAt || full.Signals[2].Basis != limit.Basis {
					tc.Errorf("independent limit signal/ratio: %v", err)
					return
				}
			}
			if view.Percentages.Limit.State != "unavailable" || view.Percentages.Limit.Value != "" || view.Signals[0].ReadAt != used.ReadAt || view.Signals[0].Basis != used.Basis {
				tc.Errorf("absent limit/provenance %s", mode)
				return
			}
			if input.Request != nil && (view.Signals[1].ReadAt != denom.ReadAt || view.Signals[1].Basis != denom.Basis) {
				tc.Errorf("denominator provenance %s", mode)
				return
			}
		}
	}, options(8)...)
}

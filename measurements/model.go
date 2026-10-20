// Package measurements binds typed measurement kernels to bounded OTel sources.
package measurements

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/identity"
	"github.com/casselc/goo/lifecycle"
	"github.com/casselc/goo/metrics"
	"gopkg.in/yaml.v3"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const Version = "goo/measurement-package/v1alpha1"
const RowLimit = 10000

var stable = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_.-]{0,127}$`)

type Binding struct {
	Namespace  *identity.Namespace `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Services   []string            `json:"services" yaml:"services"`
	Scope      string              `json:"scope" yaml:"scope"`
	Version    string              `json:"version" yaml:"version"`
	SchemaURL  string              `json:"schemaURL" yaml:"schemaURL"`
	Metric     string              `json:"metric,omitempty" yaml:"metric,omitempty"`
	Unit       string              `json:"unit,omitempty" yaml:"unit,omitempty"`
	Attributes map[string]string   `json:"attributes,omitempty" yaml:"attributes,omitempty"`
}
type LifecycleBinding struct {
	LogicalService string   `json:"logicalService" yaml:"logicalService"`
	Tenants        []string `json:"tenants" yaml:"tenants"`
	Accepted       string   `json:"accepted" yaml:"accepted"`
	Terminal       string   `json:"terminal" yaml:"terminal"`
	EventID        string   `json:"eventID" yaml:"eventID"`
	EventTime      string   `json:"eventTime" yaml:"eventTime"`
	Subject        string   `json:"subject" yaml:"subject"`
	Tenant         string   `json:"tenant" yaml:"tenant"`
	Release        string   `json:"release" yaml:"release"`
	Outcome        string   `json:"outcome" yaml:"outcome"`
	Success        string   `json:"success" yaml:"success"`
	Outcomes       []string `json:"outcomes" yaml:"outcomes"`
	DeadlineMS     int64    `json:"deadlineMs" yaml:"deadlineMs"`
	AllowanceMS    int64    `json:"allowanceMs" yaml:"allowanceMs"`
}
type Definition struct {
	APIVersion      string            `json:"apiVersion" yaml:"apiVersion"`
	Name            string            `json:"name" yaml:"name"`
	Title           string            `json:"title" yaml:"title"`
	Kind            string            `json:"kind" yaml:"kind"`
	Environment     string            `json:"environment" yaml:"environment"`
	LookbackMinutes int               `json:"lookbackMinutes" yaml:"lookbackMinutes"`
	MaxGapMS        int64             `json:"maxGapMs" yaml:"maxGapMs"`
	Grain           string            `json:"grain" yaml:"grain"`
	Source          Binding           `json:"source" yaml:"source"`
	Lifecycle       *LifecycleBinding `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
}
type Question struct {
	Definition Definition `json:"definition"`
	FromMS     int64      `json:"fromMs"`
	ToMS       int64      `json:"toMs"`
	FromNS     string     `json:"fromNS,omitempty"`
	ToNS       string     `json:"toNS,omitempty"`
}
type Plan struct {
	SQL        string            `json:"sql"`
	Parameters map[string]string `json:"parameters"`
	Revision   string            `json:"revision"`
}
type Result struct {
	Events     []lifecycle.Event        `json:"events,omitempty"`
	Question   Question                 `json:"question"`
	Plan       Plan                     `json:"plan"`
	ReadAt     string                   `json:"readAt"`
	Basis      string                   `json:"basis"`
	Count      string                   `json:"count"`
	State      string                   `json:"state"`
	Warnings   []string                 `json:"warnings"`
	Rows       []json.RawMessage        `json:"rows"`
	Gauges     *metrics.GaugeResult     `json:"gauges,omitempty"`
	Counters   *metrics.CounterResult   `json:"counters,omitempty"`
	Histograms *metrics.HistogramResult `json:"histograms,omitempty"`
	Lifecycle  *lifecycle.Prepared      `json:"lifecycle,omitempty"`
}

func Hash(v any) string {
	b, _ := json.Marshal(v)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func text(s string, max int) bool {
	return len(s) <= max && utf8.ValidString(s) && !strings.ContainsRune(s, 0)
}
func (d Definition) Validate() error {
	if err := d.Source.Namespace.Validate(); err != nil {
		return err
	}
	if d.APIVersion != Version || !stable.MatchString(d.Name) || d.Title == "" || !text(d.Title, 160) || d.Environment == "" || !text(d.Environment, 128) || d.LookbackMinutes < 1 || d.LookbackMinutes > 120 || d.MaxGapMS < 1 || d.MaxGapMS > 600000 {
		return fmt.Errorf("require versioned named package, explicit environment, 1..120 minute lookback and 1..600000 ms gap/freshness")
	}
	if d.Kind != "gauge" && d.Kind != "counter" && d.Kind != "histogram" && d.Kind != "lifecycle" {
		return fmt.Errorf("unsupported measurement kind")
	}
	if d.Grain != "per-stream" && d.Grain != "consumer-group" && d.Grain != "subject" {
		return fmt.Errorf("declare measurement grain")
	}
	if len(d.Source.Services) < 1 || len(d.Source.Services) > 8 || !text(d.Source.Scope, 256) || !text(d.Source.Version, 128) || !text(d.Source.SchemaURL, 1024) || len(d.Source.Attributes) > 16 {
		return fmt.Errorf("invalid source binding")
	}
	seen := map[string]bool{}
	for _, s := range d.Source.Services {
		if s == "" || !text(s, 128) || seen[s] {
			return fmt.Errorf("require unique bounded source services")
		}
		seen[s] = true
	}
	for k, v := range d.Source.Attributes {
		if k == "" || !text(k, 128) || !text(v, 512) {
			return fmt.Errorf("invalid attribute selector")
		}
	}
	if d.Kind != "lifecycle" {
		if d.Lifecycle != nil || d.Source.Scope == "" || d.Source.Metric == "" || !text(d.Source.Metric, 256) || d.Source.Unit == "" || !text(d.Source.Unit, 64) || d.Grain == "subject" {
			return fmt.Errorf("require typed metric/unit binding")
		}
		return nil
	}
	l := d.Lifecycle
	if d.Source.Scope != "" || d.Source.Version != "" || d.Source.SchemaURL != "" {
		return fmt.Errorf("lifecycle source scope/version/schema selectors are not supported; bind services, environment and named events")
	}
	if l == nil || d.Grain != "subject" || len(d.Source.Attributes) > 0 || d.Source.Metric != "" || d.Source.Unit != "" || l.DeadlineMS < 1 || l.DeadlineMS > 3600000 || l.AllowanceMS < 0 || l.AllowanceMS > 600000 || len(l.Tenants) < 1 || len(l.Tenants) > 16 || len(l.Outcomes) < 1 || len(l.Outcomes) > 16 {
		return fmt.Errorf("invalid lifecycle binding")
	}
	for _, s := range []string{l.LogicalService, l.Accepted, l.Terminal, l.EventID, l.EventTime, l.Subject, l.Tenant, l.Release, l.Outcome, l.Success} {
		if s == "" || !text(s, 128) {
			return fmt.Errorf("require explicit lifecycle field/event/outcome bindings")
		}
	}
	if l.Accepted == l.Terminal {
		return fmt.Errorf("distinct lifecycle event names required")
	}
	found := false
	for _, list := range [][]string{l.Tenants, l.Outcomes} {
		seen = map[string]bool{}
		for _, s := range list {
			if s == "" || !text(s, 128) || seen[s] {
				return fmt.Errorf("unique nonempty lifecycle scopes/outcomes required")
			}
			seen[s] = true
		}
	}
	for _, s := range l.Outcomes {
		found = found || s == l.Success
	}
	if !found {
		return fmt.Errorf("success must be a terminal outcome")
	}
	return nil
}

// BoundsNS returns exact native bounds; legacy questions retain millisecond semantics.
func (q Question) BoundsNS() (string, string) {
	if q.FromNS != "" {
		return q.FromNS, q.ToNS
	}
	return ns(q.FromMS), ns(q.ToMS)
}

func (q Question) Validate() error {
	if e := q.Definition.Validate(); e != nil {
		return e
	}
	if q.FromMS < 0 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 7200000 || q.ToMS >= 4102444800000 {
		return fmt.Errorf("require fixed half-open window <=120 minutes before 2100")
	}
	if q.FromNS != "" || q.ToNS != "" {
		from, e1 := strconv.ParseInt(q.FromNS, 10, 64)
		to, e2 := strconv.ParseInt(q.ToNS, 10, 64)
		if e1 != nil || e2 != nil || strconv.FormatInt(from, 10) != q.FromNS || strconv.FormatInt(to, 10) != q.ToNS || from < 0 || to <= from || to-from > 7200000000000 || to >= 4102444800000000000 || q.Definition.Kind == "lifecycle" || q.FromMS != from/1000000 || q.ToMS != (to+999999)/1000000 {
			return fmt.Errorf("require canonical paired native nanosecond bounds and matching enclosing millisecond bounds")
		}
	}
	if l := q.Definition.Lifecycle; l != nil && q.FromMS < l.DeadlineMS+l.AllowanceMS {
		return fmt.Errorf("lifecycle cohort predates epoch")
	}
	return nil
}
func Read(text string) (Definition, error) {
	var d Definition
	if len(text) > 32768 {
		return d, fmt.Errorf("package exceeds 32 KiB")
	}
	var n yaml.Node
	if e := yaml.Unmarshal([]byte(text), &n); e != nil {
		return d, e
	}
	var walk func(*yaml.Node, int) error
	walk = func(n *yaml.Node, depth int) error {
		if depth > 16 || n.Kind == yaml.AliasNode || n.Tag == "!!merge" {
			return fmt.Errorf("YAML aliases/merges/depth unsupported")
		}
		for _, c := range n.Content {
			if e := walk(c, depth+1); e != nil {
				return e
			}
		}
		return nil
	}
	if e := walk(&n, 0); e != nil {
		return d, e
	}
	dec := yaml.NewDecoder(bytes.NewBufferString(text))
	dec.KnownFields(true)
	if e := dec.Decode(&d); e != nil {
		return d, e
	}
	if dec.Decode(new(any)) != io.EOF {
		return d, fmt.Errorf("one YAML document required")
	}
	return d, d.Validate()
}

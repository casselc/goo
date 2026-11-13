package exploration

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/casselc/goo/internal/sourcefailure"
)

const ExceptionGroupsVersion = "goo/exception-groups/v1"
const ExceptionGroupsMaxGroups = 1000
const ExceptionGroupsMaxAffected = 32
const ExceptionGroupsMaxResponseBytes = 8 << 20

type ExceptionAffectedIdentity struct {
	Service   string  `json:"service"`
	Namespace *string `json:"namespace"`
	Version   *string `json:"version"`
}
type ExceptionGroup struct {
	Fingerprint     string                      `json:"fingerprint"`
	State           string                      `json:"state"`
	Reason          string                      `json:"reason"`
	Count           string                      `json:"count"`
	FirstNS         string                      `json:"firstNS"`
	LastNS          string                      `json:"lastNS"`
	Inputs          *ExceptionInputs            `json:"inputs"`
	Affected        []ExceptionAffectedIdentity `json:"affected"`
	AffectedCount   string                      `json:"affectedCount"`
	AffectedLimited bool                        `json:"affectedLimited"`
}
type ExceptionGroupsReads struct {
	Count  CountedValuesRead `json:"count"`
	Groups CountedValuesRead `json:"groups"`
}
type ExceptionGroupsResult struct {
	Version    string               `json:"version"`
	Question   Query                `json:"question"`
	CapturedAt string               `json:"capturedAt"`
	Basis      string               `json:"basis"`
	Count      string               `json:"count"`
	State      string               `json:"state"`
	Groups     []ExceptionGroup     `json:"groups"`
	Warnings   []string             `json:"warnings"`
	Reads      ExceptionGroupsReads `json:"reads"`
}
type ExceptionGroupsPlan struct {
	Count      string            `json:"count"`
	Groups     string            `json:"groups"`
	Parameters map[string]string `json:"parameters"`
}

func ValidateExceptionGroupsQuery(q Query) error {
	if e := q.Validate(); e != nil {
		return e
	}
	if len(q.GroupBy) != 0 || q.Offset != 0 || q.Summary != "" || q.SpanStatus || q.BucketMS != 0 || q.Numeric != nil || q.NumericHistogram != nil || q.GroupTimeSeries || q.ReportedEdge != nil {
		return fmt.Errorf("exception groups require an ordinary page-zero log/span scope without display projections")
	}
	return nil
}
func CompileExceptionGroups(q Query) (ExceptionGroupsPlan, error) {
	if e := ValidateExceptionGroupsQuery(q); e != nil {
		return ExceptionGroupsPlan{}, e
	}
	selection, e := q.EnvironmentSelection()
	if e != nil {
		return ExceptionGroupsPlan{}, e
	}
	c := compiler{signal: q.Signal, params: map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10)}}
	base, _, e := c.source(q, selection)
	if e != nil {
		return ExceptionGroupsPlan{}, e
	}
	nullable := func(key string) string {
		return "if(mapContains(ResourceAttributes, '" + key + "'), CAST(ResourceAttributes['" + key + "'] AS Nullable(String)), NULL)"
	}
	projected := "SELECT " + exceptionProjectionSQL(q.Signal) + " AS exception_tuple, ServiceName AS service, " + nullable("service.namespace") + " AS namespace, " + nullable("service.version") + " AS version, toUnixTimestamp64Nano(Timestamp) AS at_ns" + base
	fp := exceptionFingerprintFromProjectionSQL("exception_tuple")
	records := "SELECT " + fp + " AS fingerprint, service, namespace, version, at_ns, if(exception_tuple.1 = '', exception_tuple.2, '') AS input_type, if(exception_tuple.1 = '', exception_tuple.3, '') AS input_stack FROM (" + projected + ")"
	// Group by the full affected identity first: groupArraySorted then produces a
	// deterministic bounded prefix of distinct identities, never sampled records.
	identities := "SELECT fingerprint, service, namespace, version, count() AS n, min(at_ns) AS first_ns, max(at_ns) AS last_ns, min(input_type) AS type_min, max(input_type) AS type_max, min(input_stack) AS stack_min, max(input_stack) AS stack_max FROM (" + records + ") GROUP BY fingerprint, service, namespace, version"
	groups := "SELECT fingerprint, toString(sum(n)) AS count, toString(min(first_ns)) AS first_ns, toString(max(last_ns)) AS last_ns, min(type_min) AS type_min, max(type_max) AS type_max, min(stack_min) AS stack_min, max(stack_max) AS stack_max, groupArraySorted(33)(tuple(service, toUInt8(isNotNull(namespace)), ifNull(namespace, ''), toUInt8(isNotNull(version)), ifNull(version, ''))) AS affected, toString(count()) AS affected_count FROM (" + identities + ") GROUP BY fingerprint ORDER BY toUInt64(count) DESC, fingerprint ASC LIMIT 1001 FORMAT JSONEachRow"
	return ExceptionGroupsPlan{Count: "SELECT toString(count()) AS count" + base + " FORMAT JSONEachRow", Groups: groups, Parameters: c.params}, nil
}

type exceptionGroupWire struct {
	Fingerprint   string              `json:"fingerprint"`
	Count         string              `json:"count"`
	FirstNS       string              `json:"first_ns"`
	LastNS        string              `json:"last_ns"`
	TypeMin       string              `json:"type_min"`
	TypeMax       string              `json:"type_max"`
	StackMin      string              `json:"stack_min"`
	StackMax      string              `json:"stack_max"`
	Affected      [][]json.RawMessage `json:"affected"`
	AffectedCount string              `json:"affected_count"`
}

func decodeExceptionGroup(q Query, raw json.RawMessage) (ExceptionGroup, error) {
	fail := func() (ExceptionGroup, error) {
		return ExceptionGroup{}, fmt.Errorf("invalid exception group source projection")
	}
	if !utf8.Valid(raw) {
		return fail()
	}
	var keys map[string]json.RawMessage
	if json.Unmarshal(raw, &keys) != nil || len(keys) != 10 {
		return fail()
	}
	for _, key := range []string{"fingerprint", "count", "first_ns", "last_ns", "type_min", "type_max", "stack_min", "stack_max", "affected", "affected_count"} {
		if _, ok := keys[key]; !ok {
			return fail()
		}
	}
	var w exceptionGroupWire
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if e := decoder.Decode(&w); e != nil {
		return fail()
	}
	if e := ValidateExceptionFingerprint(w.Fingerprint); e != nil {
		return fail()
	}
	if q.Signal == "logs" && (w.Fingerprint == "ef1:u:none" || w.Fingerprint == "ef1:u:multiple") {
		return fail()
	}
	count, e := countedValuesUint(w.Count, true)
	if e != nil {
		return fail()
	}
	first, e := strconv.ParseInt(w.FirstNS, 10, 64)
	if e != nil || strconv.FormatInt(first, 10) != w.FirstNS {
		return fail()
	}
	last, e := strconv.ParseInt(w.LastNS, 10, 64)
	if e != nil || strconv.FormatInt(last, 10) != w.LastNS || first > last || first < q.FromMS*1_000_000 || last >= q.ToMS*1_000_000 {
		return fail()
	}
	affectedCount, e := countedValuesUint(w.AffectedCount, true)
	if e != nil || affectedCount > count || len(w.Affected) != int(min(affectedCount, uint64(ExceptionGroupsMaxAffected+1))) {
		return fail()
	}
	group := ExceptionGroup{Fingerprint: w.Fingerprint, State: "unclassified", Reason: strings.TrimPrefix(w.Fingerprint, "ef1:u:"), Count: w.Count, FirstNS: w.FirstNS, LastNS: w.LastNS, Affected: []ExceptionAffectedIdentity{}, AffectedCount: w.AffectedCount, AffectedLimited: affectedCount > ExceptionGroupsMaxAffected}
	if strings.HasPrefix(w.Fingerprint, "ef1:c:") {
		if w.TypeMin != w.TypeMax || w.StackMin != w.StackMax {
			return fail()
		}
		expected := ReferenceLogExceptionFingerprint(map[string]string{"exception.type": w.TypeMin, "exception.stacktrace": w.StackMin})
		if expected.State != "classified" || expected.Fingerprint != w.Fingerprint {
			return fail()
		}
		group.State = "classified"
		group.Reason = ""
		group.Inputs = expected.Inputs
	} else if w.TypeMin != "" || w.TypeMax != "" || w.StackMin != "" || w.StackMax != "" {
		return fail()
	}
	seen := map[string]bool{}
	var previous *ExceptionAffectedIdentity
	for i, values := range w.Affected {
		if len(values) != 5 {
			return fail()
		}
		var service, namespace, version string
		var hasNamespace, hasVersion uint8
		if json.Unmarshal(values[0], &service) != nil || json.Unmarshal(values[1], &hasNamespace) != nil || json.Unmarshal(values[2], &namespace) != nil || json.Unmarshal(values[3], &hasVersion) != nil || json.Unmarshal(values[4], &version) != nil || hasNamespace > 1 || hasVersion > 1 || hasNamespace == 0 && namespace != "" || hasVersion == 0 && version != "" {
			return fail()
		}
		for _, v := range []string{service, namespace, version} {
			if !utf8.ValidString(v) || len(v) > 1024 {
				return fail()
			}
		}
		identity := ExceptionAffectedIdentity{Service: service}
		if hasNamespace == 1 {
			identity.Namespace = &namespace
		}
		if hasVersion == 1 {
			identity.Version = &version
		}
		encoded, _ := json.Marshal(identity)
		key := string(encoded)
		if seen[key] || previous != nil && compareExceptionAffected(*previous, identity) >= 0 {
			return fail()
		}
		seen[key] = true
		copy := identity
		previous = &copy
		if i < ExceptionGroupsMaxAffected {
			group.Affected = append(group.Affected, identity)
		}
	}
	return group, nil
}
func compareExceptionAffected(a, b ExceptionAffectedIdentity) int {
	if c := cmp.Compare(a.Service, b.Service); c != 0 {
		return c
	}
	for _, p := range [][2]*string{{a.Namespace, b.Namespace}, {a.Version, b.Version}} {
		if p[0] == nil && p[1] != nil {
			return -1
		}
		if p[0] != nil && p[1] == nil {
			return 1
		}
		if p[0] != nil {
			if c := cmp.Compare(*p[0], *p[1]); c != 0 {
				return c
			}
		}
	}
	return 0
}
func (s Source) FetchExceptionGroups(ctx context.Context, q Query) (result *ExceptionGroupsResult, resultErr error) {
	// Pin one caller-independent question for execution and the returned receipt.
	encoded, e := json.Marshal(q)
	if e != nil {
		return nil, e
	}
	var question Query
	if e = json.Unmarshal(encoded, &question); e != nil {
		return nil, e
	}
	plan, e := CompileExceptionGroups(question)
	if e != nil {
		return nil, e
	}
	defer func() { resultErr = sourcefailure.Protocol(resultErr) }()
	read := func(sql string) ([]json.RawMessage, CountedValuesRead, error) {
		if e := ctx.Err(); e != nil {
			return nil, CountedValuesRead{}, e
		}
		raw, e := s.read(ctx, sql, plan.Parameters)
		if e != nil {
			return nil, CountedValuesRead{}, e
		}
		return raw, CountedValuesRead{ReadAt: time.Now().UTC().Format(time.RFC3339Nano), Basis: countedValuesHash(raw)}, nil
	}
	raw, countRead, e := read(plan.Count)
	if e != nil {
		return nil, e
	}
	var count struct {
		Count string `json:"count"`
	}
	var countKeys map[string]json.RawMessage
	if len(raw) != 1 || json.Unmarshal(raw[0], &countKeys) != nil || len(countKeys) != 1 || json.Unmarshal(raw[0], &count) != nil {
		return nil, fmt.Errorf("invalid exception count source projection")
	}
	if _, e = countedValuesUint(count.Count, false); e != nil {
		return nil, e
	}
	groups, groupRead, e := read(plan.Groups)
	if e != nil {
		return nil, e
	}
	if len(groups) > ExceptionGroupsMaxGroups {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("more than 1000 exception groups; narrow scope"))
	}
	r := &ExceptionGroupsResult{Version: ExceptionGroupsVersion, Question: question, Count: count.Count, State: "observed", Groups: []ExceptionGroup{}, Warnings: []string{"Matching captured records, not exception-event counts or complete upstream traffic.", "First/last are matching record timestamps within the searched window; independent reads do not form a transactional snapshot."}, Reads: ExceptionGroupsReads{Count: countRead, Groups: groupRead}}
	total := new(big.Int)
	seen := map[string]bool{}
	for _, raw := range groups {
		g, e := decodeExceptionGroup(question, raw)
		if e != nil {
			return nil, e
		}
		if seen[g.Fingerprint] {
			return nil, fmt.Errorf("duplicate exception group identity")
		}
		seen[g.Fingerprint] = true
		n, _ := new(big.Int).SetString(g.Count, 10)
		total.Add(total, n)
		r.Groups = append(r.Groups, g)
	}
	if count.Count == "0" {
		r.State = "no_data"
	}
	if total.String() != count.Count {
		r.State = "changed_between_reads"
		r.Warnings = append(r.Warnings, "Independent count/group reads disagree; groups are diagnostic and cannot authorize drilldown.")
	}
	r.CapturedAt = time.Now().UTC().Format(time.RFC3339Nano)
	basis := *r
	basis.CapturedAt = ""
	basis.Reads.Count.ReadAt = ""
	basis.Reads.Groups.ReadAt = ""
	r.Basis = countedValuesHash(basis)
	response, e := json.Marshal(r)
	if e != nil {
		return nil, e
	}
	if len(response) > ExceptionGroupsMaxResponseBytes {
		return nil, sourcefailure.Wrap(sourcefailure.Budget, fmt.Errorf("exception group response exceeds 8 MiB; narrow scope"))
	}
	return r, nil
}

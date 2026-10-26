package exploration

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

const PodDiscoveryVersion = "goo/pod-discovery/v1"
const PodObservationLimit = 100
const PodSourceGroupLimit = 1000
const PodIdentityByteLimit = 16 << 10
const PodDiscoveryByteLimit = 8 << 20

type PodDiscoveryQuestion struct {
	Environment string `json:"environment"`
	FromMS      int64  `json:"fromMs"`
	ToMS        int64  `json:"toMs"`
}
type PodFamilyObservation struct {
	Family  string `json:"family"`
	Count   string `json:"count"`
	FirstNS string `json:"firstNS"`
	LastNS  string `json:"lastNS"`
}
type PodResourceObservation struct {
	Key               string                 `json:"key"`
	Resource          map[string]string      `json:"resource"`
	ResourceSchemaURL *string                `json:"resourceSchemaURL"`
	PodUIDState       string                 `json:"podUIDState"`
	ClusterUIDState   string                 `json:"clusterUIDState"`
	Actionable        bool                   `json:"actionable"`
	Families          []PodFamilyObservation `json:"families"`
}
type PodDiscovery struct {
	Version             string                   `json:"version"`
	Question            PodDiscoveryQuestion     `json:"question"`
	Observations        []PodResourceObservation `json:"observations"`
	SourceGroups        int                      `json:"sourceGroups"`
	ObservationsOmitted int                      `json:"observationsOmitted"`
	ReadAt              string                   `json:"readAt"`
	Basis               string                   `json:"basis"`
	Warning             string                   `json:"warning"`
}

func (q PodDiscoveryQuestion) Validate() error {
	if q.Environment == "" || len(q.Environment) > 128 || !utf8.ValidString(q.Environment) || strings.ContainsRune(q.Environment, 0) || q.FromMS < 0 || q.ToMS <= q.FromMS || q.ToMS-q.FromMS > 7200000 || q.ToMS >= 4102444800000 {
		return fmt.Errorf("pod discovery requires explicit environment and fixed half-open window <=120 minutes before 2100")
	}
	return nil
}
func podDiscoverySQL() string {
	parts := []string{}
	for _, family := range []string{"logs", "spans", "gauge", "sum", "histogram"} {
		table, clock, schema := "otel_metrics_"+family, "TimeUnix", "toNullable(toString(ResourceSchemaUrl))"
		if family == "logs" {
			table = "otel_logs"
			clock = "Timestamp"
		}
		if family == "spans" {
			table = "otel_traces"
			clock = "Timestamp"
			schema = "CAST(NULL AS Nullable(String))"
		}
		parts = append(parts, "SELECT '"+family+"' AS family, arraySort(arrayZip(mapKeys(ResourceAttributes),mapValues(ResourceAttributes))) AS labels, "+schema+" AS resource_schema_url, toString(count()) AS count, toString(min(toUnixTimestamp64Nano("+clock+"))) AS first_ns, toString(max(toUnixTimestamp64Nano("+clock+"))) AS last_ns FROM "+table+" WHERE "+clock+">=fromUnixTimestamp64Milli({from:Int64}) AND "+clock+"<fromUnixTimestamp64Milli({to:Int64}) AND ResourceAttributes['deployment.environment.name']={environment:String} AND (mapContains(ResourceAttributes,'k8s.pod.uid') OR mapContains(ResourceAttributes,'k8s.pod.name')) GROUP BY labels, resource_schema_url")
	}
	return "SELECT family, mapFromArrays(arrayMap(x -> x.1,labels),arrayMap(x -> x.2,labels)) AS resource, resource_schema_url, count, first_ns, last_ns FROM (" + strings.Join(parts, " UNION ALL ") + ") ORDER BY family, labels, resource_schema_url LIMIT 1001 FORMAT JSONEachRow"
}
func (s Source) DiscoverPods(ctx context.Context, q PodDiscoveryQuestion) (*PodDiscovery, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if e := ctx.Err(); e != nil {
		return nil, e
	}
	rows, e := s.read(ctx, podDiscoverySQL(), map[string]string{"from": strconv.FormatInt(q.FromMS, 10), "to": strconv.FormatInt(q.ToMS, 10), "environment": q.Environment})
	if e != nil {
		return nil, e
	}
	if e = ctx.Err(); e != nil {
		return nil, e
	}
	return decodePodDiscovery(q, rows, time.Now().UTC())
}
func podLabelState(resource map[string]string, key string) string {
	v, ok := resource[key]
	if !ok {
		return "missing"
	}
	if v == "" {
		return "empty"
	}
	return "present"
}
func decodePodDiscovery(q PodDiscoveryQuestion, rows []json.RawMessage, at time.Time) (*PodDiscovery, error) {
	if e := q.Validate(); e != nil {
		return nil, e
	}
	if len(rows) > PodSourceGroupLimit {
		return nil, fmt.Errorf("more than 1000 pod source-family groups; narrow the window; no partial inventory")
	}
	result := &PodDiscovery{Version: PodDiscoveryVersion, Question: q, Observations: []PodResourceObservation{}, SourceGroups: len(rows), ReadAt: at.Format(time.RFC3339Nano), Warning: "Reported pod-bearing resource variants, not physical pod inventory or Kubernetes API state. Family counts are captured records or metric data points, not messages, requests or healthy entities. Cluster names, namespace, workload and node labels are reported context; missing cluster UID does not establish absence. Matching resource record questions are contextual, not exact map/schema joins. No ownership, health, capacity or native metric authority is inferred."}
	observations := map[string]*PodResourceObservation{}
	seen := map[string]bool{}
	for _, raw := range rows {
		if !utf8.Valid(raw) {
			return nil, fmt.Errorf("invalid pod discovery UTF-8")
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(raw, &fields) != nil {
			return nil, fmt.Errorf("invalid pod discovery row")
		}
		for _, key := range []string{"family", "resource", "resource_schema_url", "count", "first_ns", "last_ns"} {
			if _, ok := fields[key]; !ok {
				return nil, fmt.Errorf("pod discovery row missing %s", key)
			}
		}
		var row struct {
			Family   string            `json:"family"`
			Resource map[string]string `json:"resource"`
			Schema   *string           `json:"resource_schema_url"`
			Count    string            `json:"count"`
			First    string            `json:"first_ns"`
			Last     string            `json:"last_ns"`
		}
		if json.Unmarshal(raw, &row) != nil || row.Resource == nil || len(row.Resource) > 64 {
			return nil, fmt.Errorf("invalid pod resource shape")
		}
		var labels map[string]json.RawMessage
		if json.Unmarshal(fields["resource"], &labels) != nil {
			return nil, fmt.Errorf("invalid pod resource labels")
		}
		for _, encoded := range labels {
			if len(encoded) == 0 || encoded[0] != '"' {
				return nil, fmt.Errorf("pod resource labels must be reported strings")
			}
		}

		if row.Family != "logs" && row.Family != "spans" && row.Family != "gauge" && row.Family != "sum" && row.Family != "histogram" {
			return nil, fmt.Errorf("invalid pod source family")
		}
		if row.Family == "spans" && row.Schema != nil || row.Family != "spans" && row.Schema == nil {
			return nil, fmt.Errorf("invalid pod resource schema availability")
		}
		for key, value := range row.Resource {
			if key == "" || len(key) > 256 || len(value) > 2048 || strings.ContainsRune(key, 0) || strings.ContainsRune(value, 0) {
				return nil, fmt.Errorf("pod resource labels exceed bounds")
			}
		}
		if row.Resource["deployment.environment.name"] != q.Environment {
			return nil, fmt.Errorf("pod discovery environment mismatch")
		}
		if _, uid := row.Resource["k8s.pod.uid"]; !uid {
			if _, name := row.Resource["k8s.pod.name"]; !name {
				return nil, fmt.Errorf("resource has no reported pod label")
			}
		}
		if row.Schema != nil && (len(*row.Schema) > 1024 || strings.ContainsRune(*row.Schema, 0)) {
			return nil, fmt.Errorf("invalid pod resource schema")
		}
		if _, e := countedValuesUint(row.Count, true); e != nil {
			return nil, e
		}
		first, e := countedValuesUint(row.First, false)
		if e != nil {
			return nil, e
		}
		last, e := countedValuesUint(row.Last, false)
		if e != nil {
			return nil, e
		}
		if first < uint64(q.FromMS)*1000000 || last >= uint64(q.ToMS)*1000000 || first > last {
			return nil, fmt.Errorf("pod source observation clock outside requested window")
		}
		identity := struct {
			Resource map[string]string
			Schema   *string
		}{row.Resource, row.Schema}
		encoded, _ := json.Marshal(identity)
		if len(encoded) > PodIdentityByteLimit {
			return nil, fmt.Errorf("pod resource identity exceeds 16 KiB")
		}
		key := countedValuesHash(identity)
		familyKey := key + ":" + row.Family
		if seen[familyKey] {
			return nil, fmt.Errorf("duplicate pod source-family group")
		}
		seen[familyKey] = true
		if observations[key] == nil {
			observations[key] = &PodResourceObservation{Key: key, Resource: row.Resource, ResourceSchemaURL: row.Schema, PodUIDState: podLabelState(row.Resource, "k8s.pod.uid"), ClusterUIDState: podLabelState(row.Resource, "k8s.cluster.uid"), Actionable: row.Resource["k8s.pod.uid"] != "", Families: []PodFamilyObservation{}}
		}
		observations[key].Families = append(observations[key].Families, PodFamilyObservation{Family: row.Family, Count: row.Count, FirstNS: row.First, LastNS: row.Last})
	}
	keys := []string{}
	for key := range observations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	allObservations := make([]PodResourceObservation, 0, len(keys))
	for index, key := range keys {
		observation := observations[key]
		sort.Slice(observation.Families, func(a, b int) bool { return observation.Families[a].Family < observation.Families[b].Family })
		allObservations = append(allObservations, *observation)
		if index >= PodObservationLimit {
			result.ObservationsOmitted++
			continue
		}
		result.Observations = append(result.Observations, *observation)
	}
	basis := *result
	basis.ReadAt = ""
	result.Basis = countedValuesHash(struct {
		Projection PodDiscovery
		Observed   []PodResourceObservation
	}{basis, allObservations})
	encoded, e := json.Marshal(result)
	if e != nil {
		return nil, e
	}
	if len(encoded) > PodDiscoveryByteLimit {
		return nil, fmt.Errorf("pod discovery reply exceeds 8 MiB")
	}
	return result, nil
}

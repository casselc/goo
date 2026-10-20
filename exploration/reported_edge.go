package exploration

import (
	"fmt"
	"github.com/casselc/goo/identity"
	"strings"
	"unicode/utf8"
)

// ReportedEdge selects child records whose reported parent has one service owner
// in the same environment/window used by discovery. It is not a call-count grain.
type EdgeService struct {
	Service   string              `json:"service" yaml:"service"`
	Namespace *identity.Namespace `json:"namespace" yaml:"namespace"`
}
type ReportedEdge struct {
	Parent             EdgeService         `json:"parent" yaml:"parent"`
	Child              EdgeService         `json:"child" yaml:"child"`
	DiscoveryNamespace *identity.Namespace `json:"discoveryNamespace,omitempty" yaml:"discoveryNamespace,omitempty"`
}

func (e *ReportedEdge) Validate() error {
	if e == nil {
		return nil
	}
	for _, s := range []EdgeService{e.Parent, e.Child} {
		if !utf8.ValidString(s.Service) || len(s.Service) > 128 || strings.ContainsRune(s.Service, 0) || s.Namespace == nil {
			return fmt.Errorf("reported edge requires bounded service identities and explicit namespace presence/value")
		}
		if err := s.Namespace.Validate(); err != nil {
			return err
		}
	}
	if err := e.DiscoveryNamespace.Validate(); err != nil {
		return err
	}
	if e.DiscoveryNamespace != nil && (!e.DiscoveryNamespace.Equal(e.Parent.Namespace) || !e.DiscoveryNamespace.Equal(e.Child.Namespace)) {
		return fmt.Errorf("edge identities differ from discovery namespace")
	}
	if e.Parent.Service == e.Child.Service && e.Parent.Namespace.Equal(e.Child.Namespace) {
		return fmt.Errorf("reported edge requires distinct service identities")
	}
	return nil
}

func (c *compiler) edgeNamespace(n *identity.Namespace) string {
	present := "mapContains(ResourceAttributes,'service.namespace')"
	if n.Mode == "missing" {
		return "NOT " + present
	}
	return present + " AND ResourceAttributes['service.namespace']=" + c.param(n.Value, "String")
}
func (c *compiler) reportedEdge(q Query) string {
	e := q.ReportedEdge
	parentScope := "Timestamp >= fromUnixTimestamp64Milli({from:Int64}) AND Timestamp < fromUnixTimestamp64Milli({to:Int64}) AND mapContains(ResourceAttributes,'deployment.environment.name') AND ResourceAttributes['deployment.environment.name']={environment:String}"
	if e.DiscoveryNamespace != nil {
		parentScope += " AND " + c.edgeNamespace(e.DiscoveryNamespace)
	}
	present := "1"
	if e.Parent.Namespace.Mode == "missing" {
		present = "0"
	}
	owner := "tuple(" + c.param(present, "UInt8") + "," + c.param(e.Parent.Namespace.Value, "String") + "," + c.param(e.Parent.Service, "String") + ")"
	parents := "SELECT TraceId, SpanId FROM otel_traces WHERE " + parentScope + " GROUP BY TraceId, SpanId HAVING groupUniqArray(tuple(mapContains(ResourceAttributes,'service.namespace'),ResourceAttributes['service.namespace'],ServiceName))=[" + owner + "]"
	return "(ServiceName=" + c.param(e.Child.Service, "String") + " AND " + c.edgeNamespace(e.Child.Namespace) + " AND match(TraceId,'^[a-f0-9]{32}$') AND TraceId!=repeat('0',32) AND match(ParentSpanId,'^[a-f0-9]{16}$') AND ParentSpanId!=repeat('0',16) AND (TraceId,ParentSpanId) IN (" + parents + "))"
}

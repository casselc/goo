package verification

import (
	"encoding/json"
	"fmt"
	"github.com/casselc/goo/annotations"
	"hegel.dev/go/hegel"
	"strings"
	"testing"
)

func TestHegelOperationPackagePlanning(t *testing.T) {
	base, err := annotations.LoadOperationPackage("../config/operations/checkout.yaml")
	if err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(base.Definition)
	hegel.Test(t, func(tc *hegel.T) {
		var p annotations.OperationPackage
		if err := json.Unmarshal(encoded, &p); err != nil {
			tc.Fatalf("%v", err)
		}
		suffix := hegel.Draw(tc, hegel.Integers(0, 100000))
		p.Resource.Service = fmt.Sprintf("service-%d'quoted", suffix)
		if hegel.Draw(tc, hegel.Integers(0, 1)) == 1 {
			p.Operations[0], p.Operations[2] = p.Operations[2], p.Operations[0]
		}
		before, _ := json.Marshal(p)
		a, err := p.Bind(base.Manifest)
		if err != nil {
			tc.Fatalf("%v", err)
		}
		after, _ := json.Marshal(p)
		if string(before) != string(after) {
			tc.Fatalf("planning mutated package")
		}
		b, err := p.Bind(base.Manifest)
		if err != nil || a.Revision != b.Revision {
			tc.Fatalf("unstable plan")
		}
		query, params, err := a.Manifest.OperationQuery(a.Source, 0, 1)
		if err != nil || params["service"] != p.Resource.Service || strings.Contains(query, p.Resource.Service) {
			tc.Fatalf("resource escaped typed binding")
		}
		p.Views = []string{"operation-overview", "trace-waterfall", "subject-events"}
		changed, err := p.Bind(base.Manifest)
		if err != nil || changed.Revision == a.Revision {
			tc.Fatalf("view policy absent from revision")
		}
	}, options(150)...)
}

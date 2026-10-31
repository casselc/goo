package queryhttp_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/internal/capturehttp"
	"github.com/casselc/goo/internal/explanationhttp"
	"github.com/casselc/goo/internal/nativehttp"
	"github.com/casselc/goo/internal/operationhttp"
	"github.com/casselc/goo/internal/queryhttp"
	"github.com/casselc/goo/internal/recordhttp"
)

func TestFamilyPoliciesMatchPublishedAPIBoundaries(t *testing.T) {
	data, err := os.ReadFile("../../api/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]struct {
			Post struct {
				Limit      int64 `json:"x-go-body-limit-bytes"`
				Parameters []struct{ Name string }
			}
		}
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	families := map[string]struct {
		lookup func(string, string) (queryhttp.Policy, bool)
		count  int
	}{
		"record":      {recordhttp.PolicyFor, 10},
		"native":      {nativehttp.PolicyFor, 7},
		"explanation": {explanationhttp.PolicyFor, 3},
		"capture":     {capturehttp.PolicyFor, 2},
	}
	for name, family := range families {
		t.Run(name, func(t *testing.T) {
			covered := 0
			for path, operation := range spec.Paths {
				policy, ok := family.lookup("POST", path)
				if !ok {
					continue
				}
				covered++
				if operation.Post.Limit != policy.BodyLimit {
					t.Errorf("%s body bound differs from API", path)
				}
				controlled := false
				for _, parameter := range operation.Post.Parameters {
					controlled = controlled || parameter.Name == "X-Goo-Query-ID"
				}
				if controlled != (policy.Effect != queryhttp.Pure) {
					t.Errorf("%s capability differs from effect policy", path)
				}
			}
			if covered != family.count {
				t.Fatalf("expected %d adopted family routes, covered %d", family.count, covered)
			}
		})
	}
}

func TestOperationGETPoliciesMatchPublishedAdmission(t *testing.T) {
	data, err := os.ReadFile("../../api/openapi.json")
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Paths map[string]map[string]struct{ Parameters []struct{ Name string } }
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		t.Fatal(err)
	}
	count := 0
	for path, methods := range spec.Paths {
		policy, ok := operationhttp.PolicyFor("GET", path)
		if !ok {
			continue
		}
		count++
		controlled := false
		for _, p := range methods["get"].Parameters {
			controlled = controlled || p.Name == "X-Goo-Query-ID"
		}
		if controlled != (policy.Effect != queryhttp.Pure) {
			t.Errorf("%s admission differs", path)
		}
		head, ok := operationhttp.PolicyFor("HEAD", path)
		if !ok || head != policy {
			t.Errorf("%s implicit HEAD differs", path)
		}
	}
	if count != 5 {
		t.Fatalf("expected five GET declarations, got %d", count)
	}
}

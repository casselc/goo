package verification

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/casselc/goo/workbench"
)

// A finite shared wire contract, not an independent universal JSON validator.
func TestNativeHistogramWireCases373(t *testing.T) {
	data, err := os.ReadFile("../testdata/protocol/native-histogram-view-questions-373.json")
	if err != nil {
		t.Fatal(err)
	}
	var bank struct {
		Version string
		Cases   []struct {
			Name         string
			Question     json.RawMessage
			RuntimeValid bool
		}
	}
	if err = json.Unmarshal(data, &bank); err != nil {
		t.Fatal(err)
	}
	if bank.Version != "goo.native-histogram-question-cases/1" || len(bank.Cases) < 50 {
		t.Fatal("finite native boundary bank missing")
	}
	for _, item := range bank.Cases {
		t.Run(item.Name, func(t *testing.T) {
			raw := append([]byte(`{"id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","question":`), item.Question...)
			raw = append(raw, '}')
			var request workbench.NativeViewReuseRequest
			err := json.Unmarshal(raw, &request)
			if err == nil {
				err = (workbench.NativeViewInput{Version: 1, Title: "Wire case", Question: request.Question, Declaration: workbench.NativeViewDeclaration{Version: "goo.view/1", Scope: workbench.NativeViewScope{Kind: "measurement-question", Signal: "metrics"}, Selection: workbench.NativeViewSelection{Kind: "export-interval"}, Kind: workbench.ViewKind{Name: "histogram", Version: 1}, Display: workbench.NativeViewDisplay{Category: "bucket", Value: "observation-count", RequiresComplete: true}}}).Validate()
			}
			if (err == nil) != item.RuntimeValid {
				t.Fatalf("admission=%v want%v: %v", err == nil, item.RuntimeValid, err)
			}
		})
	}
}

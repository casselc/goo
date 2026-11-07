package verification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"hegel.dev/go/hegel"
)

// Mounted production elements own publication, retirement and local Find. This
// fixture's minimal DOM cannot establish browser painting or native focus; the
// live journey checks those separately. No transport or filtering oracle is copied.
func TestHegelAtomicFacetPublication346(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		input := struct {
			Rounds int `json:"rounds"`
			Seed   int `json:"seed"`
		}{hegel.Draw(tc, hegel.Integers(2, 6)), hegel.Draw(tc, hegel.Integers(0, 1000000))}
		raw, err := json.Marshal(input)
		if err != nil {
			tc.Fatal(err)
			return
		}
		cmd := exec.CommandContext(t.Context(), "node", "--no-warnings", "fixtures/atomic-facet-346.mjs")
		cmd.Stdin = bytes.NewReader(raw)
		out, err := cmd.CombinedOutput()
		if err != nil {
			tc.Fatal(err, string(out))
			return
		}
		type values struct {
			Items []struct {
				Value *string
				Count string
			}
		}
		type model struct {
			Values   values
			Disabled bool
		}
		type row struct {
			Key    string
			Hidden bool
			Count  string
		}
		type snapshot struct {
			Data             json.RawMessage
			Find             *string
			Focused, Details bool
			Rows             []row
		}
		type event struct {
			Type   string
			Detail struct {
				Value *string
				Mode  string
			}
		}
		var result struct {
			History []struct {
				Turn                           int
				Same, Connected, Allowed       bool
				Expected, Adopted              json.RawMessage
				Before, After                  snapshot
				CurrentEvents                  []event
				RetiredEvents, DetachedRenders int
				Rendered                       []struct {
					Owner   int
					Adopted json.RawMessage
				}
			}
			Invalid struct {
				Data        any
				Interaction any
				State       string
				NewEvents   int
			}
		}
		if err = json.Unmarshal(out, &result); err != nil {
			tc.Fatal(err, string(out))
			return
		}
		if len(result.History) != input.Rounds || result.Invalid.Data != nil || result.Invalid.Interaction != nil || result.Invalid.State != "unavailable" || result.Invalid.NewEvents != 0 {
			tc.Fatal("unfinished schedule or stale invalid publication", result.Invalid)
			return
		}
		equal := func(a, b json.RawMessage) bool {
			var x, y any
			return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
		}
		// Model the accepted local Find independently from the DOM input value.
		// The fixture assigns "terminal" even for an inactive owner: its raw
		// input changes, but the refused callback must not adopt that query.
		adoptedFind := ""
		for turn, step := range result.History {
			if step.Turn != turn || step.Same != (turn > 0 && (input.Seed+turn)%2 == 0) || step.Connected != ((input.Seed+turn)%3 != 0) || step.Allowed != ((input.Seed+turn)%4 != 0) {
				tc.Fatal("schedule not executed", turn)
				return
			}
			if step.RetiredEvents != 0 || !equal(step.Expected, step.Adopted) || !equal(step.Expected, step.After.Data) {
				tc.Fatal("caller/getter mutation or retired callback changed authority", turn)
				return
			}
			// Inspect the model and interaction observed by every actual render, rather
			// than treating a render count as proof that no mixed pair was published.
			if len(step.Rendered) != 1 {
				tc.Fatal("publication missing or duplicated", turn, len(step.Rendered))
				return
			}
			for _, render := range step.Rendered {
				if render.Owner != turn || !equal(render.Adopted, step.Expected) {
					tc.Fatal("mixed model and owner published", turn, render.Owner)
					return
				}
			}
			if !step.Connected && step.DetachedRenders != 0 {
				tc.Fatal("detached publication rendered", turn)
				return
			}
			var expected model
			if err = json.Unmarshal(step.Expected, &expected); err != nil {
				tc.Fatal(err)
				return
			}
			wantEvents := 0
			if !expected.Disabled && step.Allowed {
				wantEvents = 3
			}
			if len(step.CurrentEvents) != wantEvents {
				tc.Fatal("current owner authority disagrees", turn, len(step.CurrentEvents), wantEvents)
				return
			}
			if wantEvents != 0 {
				action := step.CurrentEvents[0]
				literal := fmt.Sprintf("literal Ω🙂 %d terminal-%d", input.Seed, turn)
				if action.Type != "goo:facet-stage" || action.Detail.Value == nil || *action.Detail.Value != literal || action.Detail.Mode != "include" || step.CurrentEvents[1].Type != "goo:facet-load" || step.CurrentEvents[2].Type != "goo:facet-clear" {
					tc.Fatal("current callbacks emitted wrong literal or intent", turn)
					return
				}
			}
			if !step.Same {
				adoptedFind = ""
			}
			wantFind := adoptedFind
			if step.After.Find == nil || *step.After.Find != wantFind || step.After.Details != step.Same || turn > 0 && !step.After.Focused {
				tc.Fatal("same-read Find/disclosure or mounted focus changed", turn, step.After)
				return
			}
			// Derive visible membership and exact counts from the complete supplied
			// partition, not from a rendered slice or the component's matcher.
			wantRows := []row{}
			for _, item := range expected.Values.Items {
				label, key := "Missing key", "0"
				if item.Value != nil {
					key = "1" + *item.Value
					if *item.Value == "" {
						label = "Present empty string"
					} else {
						encoded, err := json.Marshal(*item.Value)
						if err != nil {
							tc.Fatal(err)
							return
						}
						label = string(encoded)
					}
				}
				if strings.Contains(strings.ToLower(label), strings.ToLower(wantFind)) {
					wantRows = append(wantRows, row{Key: key, Count: item.Count})
				}
			}
			if !reflect.DeepEqual(step.After.Rows, wantRows) {
				tc.Fatal("mounted Find projection changed literal membership or exact counts", turn, step.After.Rows, wantRows)
				return
			}
			if step.Allowed {
				adoptedFind = "terminal"
			}
		}
	}, options(40)...)
}

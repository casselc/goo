package verification

import (
	"encoding/json"
	"github.com/casselc/goo/workbench"
	"hegel.dev/go/hegel"
	"strings"
	"testing"
)

const savedViewFindingParityJS = `import assert from 'node:assert/strict';import fs from 'node:fs';import {validateDraft} from '../ui/investigation-draft.mjs';const {card,accepted}=JSON.parse(fs.readFileSync(0,'utf8'));const draft={parent:null,document:{title:'Investigation',question:'',finding:'',openQuestions:'',cards:[card]},composer:{role:'finding',text:'',editing:null,captureID:'',caption:'',destination:'',comparisonID:'',comparisonFinding:''},openedQuestion:''};let ok=true;try{validateDraft(draft)}catch{ok=false}assert.equal(ok,accepted);`

func TestHegelSavedViewFindingValidation(t *testing.T) {
	hegel.Test(t, func(tc *hegel.T) {
		seed := hegel.Draw(tc, hegel.Integers(0, 1000))
		variant := hegel.Draw(tc, hegel.Integers(0, 5))
		card := workbench.Card{Kind: "view", Reference: strings.Repeat("a", 64), Caption: "Scoped finding", Text: strings.Repeat("é", seed%100) + "\nCaptured count, not a rate."}
		switch variant {
		case 1:
			card.Reference = "not-a-view"
		case 2:
			card.Evidence = []string{strings.Repeat("b", 64)}
		case 3:
			card.Role = "finding"
		case 4:
			card.Text = strings.Repeat("x", 4001)
		case 5:
			card.Review = strings.Repeat("c", 64)
		}
		accepted := card.Validate() == nil
		if accepted != (variant == 0) {
			tc.Fatal("unexpected Go card validation", variant)
		}
		before, _ := json.Marshal(card)
		if e := runNativeHistogramJS(map[string]any{"card": card, "accepted": accepted}, savedViewFindingParityJS); e != nil {
			tc.Fatal(e)
		}
		after, _ := json.Marshal(card)
		if string(before) != string(after) {
			tc.Fatal("validation mutated card")
		}
	}, options(100)...)
}

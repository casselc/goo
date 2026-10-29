package retention

import (
	"github.com/casselc/goo/exploration"
	"github.com/casselc/goo/identity"
	"strings"
	"testing"
)

func TestSelectorRetentionRefused250(t *testing.T) {
	s, _ := identity.EffectiveEnvironment(identity.EnvironmentScope{Mode: "unreported"})
	wire, _ := s.Selector()
	q := exploration.Query{Signal: "logs", EnvironmentSelector: &wire, FromMS: 100000, ToMS: 160000, Limit: 100}
	c := &exploration.SavedCapture{CaptureReceipt: exploration.CaptureReceipt{Question: exploration.CaptureQuestion{Kind: "explore", Explore: &q}}, Explore: &exploration.Result{Question: q}}
	if _, err := Select(c, []int{0}); err == nil || !strings.Contains(err.Error(), "environmentSelector") {
		t.Fatal("retention accepted selector", err)
	}
	m := Manifest{Capture: c.CaptureReceipt}
	if err := m.Validate(); err == nil || !strings.Contains(err.Error(), "environmentSelector") {
		t.Fatal("manifest accepted selector", err)
	}
	in := TailIntent{Selector: exploration.TailSelector{EnvironmentSelector: &wire}}
	if err := in.Validate(); err == nil || !strings.Contains(err.Error(), "environmentSelector") {
		t.Fatal("tail retention accepted selector", err)
	}
}

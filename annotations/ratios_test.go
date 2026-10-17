package annotations

import "testing"

func TestRatioPackagePolicy(t *testing.T) {
	p, e := LoadOperationPackage("../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Definition.Ratios) != 1 {
		t.Fatal("missing ratio")
	}
	for _, mode := range []string{"reference", "subset", "duplicate", "partial"} {
		t.Run(mode, func(t *testing.T) {
			p, e := LoadOperationPackage("../config/operations/checkout.yaml")
			if e != nil {
				t.Fatal(e)
			}
			d := p.Definition
			switch mode {
			case "reference":
				d.Ratios[0].TotalRate = "absent"
			case "subset":
				d.Ratios[0].ErrorOutcomes = []string{"other"}
			case "duplicate":
				d.Ratios = append(d.Ratios, d.Ratios[0])
			case "partial":
				d.Ratios[0].TotalRate = "errors"
			}
			if _, e := d.Bind(p.Manifest); e == nil {
				t.Fatal("accepted invalid ratio policy")
			}
		})
	}
}

func TestRatioCounterSourceGuard(t *testing.T) {
	p, e := LoadOperationPackage("../config/operations/checkout.yaml")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = p.EvaluateRateRows(p.Definition.Rates[0], 100, 200, make([]CounterRow, 3001)); e == nil {
		t.Fatal("source cap did not reject excess rows")
	}
}

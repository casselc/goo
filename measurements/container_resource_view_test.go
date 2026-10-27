package measurements

import "testing"

// Equal Float64 replay values do not authorize a percentage when the original
// reported decimals disagree. Both domains retain the same raw-evidence rule.
func TestContainerResourceRawDuplicateCertification(t *testing.T) {
	cpu := containerCPUFixture(t)
	cpu.Usage = containerCPUTestResult(t, cpu.Parent, "usage", "1")
	replacePoolRows(t, cpu.Usage, func(rows []metricRow) []metricRow {
		duplicate := rows[0]
		duplicate.Value = "1.00000000000000001"
		return append(rows, duplicate)
	})
	cv, e := ComposeContainerCPUView(cpu)
	if e != nil || cv.Signals[0].State != "invalid" || cv.Signals[0].ObservationState != "observed" || cv.Percentages.Request.Reason != "invalid-core-domain" {
		t.Fatal(cv, e)
	}
	memory := containerMemoryFixture(t)
	memory.WorkingSet = containerMemoryTestResult(t, memory.Parent, "working-set", "1")
	replacePoolRows(t, memory.WorkingSet, func(rows []metricRow) []metricRow {
		duplicate := rows[0]
		duplicate.Value = "1.00000000000000001"
		return append(rows, duplicate)
	})
	mv, e := ComposeContainerMemoryView(memory)
	if e != nil || mv.Signals[0].State != "invalid" || mv.Signals[0].ObservationState != "observed" || mv.Percentages.Request.Reason != "invalid-byte-domain" {
		t.Fatal(mv, e)
	}
	// Identical decimal meaning with a different spelling remains valid; the
	// canonical display and arithmetic still belong to the respective domain.
	replacePoolRows(t, cpu.Usage, func(rows []metricRow) []metricRow { rows[1].Value = "1.000"; return rows })
	cv, e = ComposeContainerCPUView(cpu)
	if e != nil || cv.Signals[0].Value != "1" || cv.Percentages.Request.Value != "400" {
		t.Fatal(cv, e)
	}
	replacePoolRows(t, memory.WorkingSet, func(rows []metricRow) []metricRow { rows[1].Value = "1.000"; return rows })
	mv, e = ComposeContainerMemoryView(memory)
	if e != nil || mv.Signals[0].Value != "1" || mv.Percentages.Request.Value != "1" {
		t.Fatal(mv, e)
	}
}

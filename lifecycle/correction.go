package lifecycle

import "fmt"

// ValidateCorrection checks the generation link, not external coverage truth or
// storage publication. The pure Prepare kernel can also evaluate independent
// snapshots; a caller claiming supersession must check this contract.
func (p Policy) ValidateCorrection(previous Receipt) error {
	next, err := p.validate()
	if err != nil {
		return err
	}
	old, err := previous.Policy.validate()
	if err != nil {
		return fmt.Errorf("invalid prior receipt: %w", err)
	}
	prior := previous.Policy
	if p.SupersedesEvaluationID != prior.EvaluationID || p.EvaluationID == prior.EvaluationID {
		return fmt.Errorf("correction must name a distinct prior evaluation ID")
	}
	if next.evaluation != old.evaluation || next.deadline != old.deadline || next.allowance != old.allowance || next.start != old.start || next.end != old.end ||
		p.DefinitionRevision != prior.DefinitionRevision || p.SourceRevision != prior.SourceRevision || p.CustodyClock != prior.CustodyClock || p.SubjectField != prior.SubjectField || p.SuccessOutcome != prior.SuccessOutcome {
		return fmt.Errorf("correction must preserve the original question, revisions and custody clock")
	}
	if next.cutoff <= old.cutoff || p.EvidenceBasis == prior.EvidenceBasis {
		return fmt.Errorf("correction must advance custody cutoff and name a new evidence basis")
	}
	if len(p.TerminalOutcomes) != len(prior.TerminalOutcomes) || len(p.Scopes) != len(prior.Scopes) {
		return fmt.Errorf("correction changed outcomes or scopes")
	}
	for _, v := range p.TerminalOutcomes {
		if !has(prior.TerminalOutcomes, v) {
			return fmt.Errorf("correction changed terminal outcomes")
		}
	}
	scopes := map[scopeKey]bool{}
	for _, s := range prior.Scopes {
		scopes[scopeOf(s.Environment, s.TenantID, s.ServiceName)] = true
	}
	for _, s := range p.Scopes {
		if !scopes[scopeOf(s.Environment, s.TenantID, s.ServiceName)] {
			return fmt.Errorf("correction changed subject scope")
		}
	}
	return nil
}

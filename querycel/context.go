package querycel

import "github.com/casselc/goo/exploration"

// ContextRequest supplies a complete immutable base for source-free authoring.
type ContextRequest struct {
	Base exploration.Query `json:"base"`
}

// RecordContext binds browser authoring to the server's validated content identity.
// Static target support is not source availability, execution authority or a lease.
type RecordContext struct {
	BaseIdentity string `json:"baseIdentity"`
	Operation    string `json:"operation"`
	Library      string `json:"library"`
	Compiler     string `json:"compiler"`
	Target       string `json:"target"`
	Profile      string `json:"profile"`
}

// ContextFor validates supplied ordinary meaning and static target admission.
// It computes identity in Go; browsers need not reproduce Go JSON encoding.
// No expression, clock, source probe, store or retained session is manufactured.
func ContextFor(base exploration.Query) (*RecordContext, error) {
	id, err := BaseIdentity(base)
	if err != nil {
		return nil, planningError("invalid_base", PhaseBinding, "base", err)
	}
	if _, err = exploration.Compile(base); err != nil {
		return nil, planningError("unsupported_target", PhaseTarget, "base", err)
	}
	return &RecordContext{id, RecordOperation, LibraryVersion, CompilerVersion, RecordTarget, RecordProfile}, nil
}

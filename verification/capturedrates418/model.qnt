// Counters describe browser source-API acquisitions and pure comparisons,
// not SQL queries. Stored metadata GETs do not grant source authority.
module capturedrates {
 const INITIAL: int
 var phase: int
 var source: int
 var compares: int
 action init = all {phase'=INITIAL, source'=0, compares'=0}
 action stage_view = all {phase==0, phase'=1, source'=source, compares'=compares}
 action read_population = all {phase==1, phase'=3, source'=source+1, compares'=compares+1}
 action local_display = all {or {phase==2,phase==3}, phase'=phase, source'=source, compares'=compares}
 action choose_baseline = all {phase==3, phase'=2, source'=source, compares'=compares}
 action prepare_baseline = all {phase==2, phase'=3, source'=source, compares'=compares+1}
 action diagnostic_compare = all {phase==3, phase'=phase, source'=source, compares'=compares+1}
 action stage_operand = all {phase==3, phase'=1, source'=source, compares'=compares}
 action read_operand = all {phase==1, phase'=2, source'=source+1, compares'=compares}
 val safety = and {phase>=0,phase<=3,source>=0,compares>=0}
}

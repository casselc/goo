// Configuration admission has no source authority. Only an explicit Run may
// acquire a result; a refused handoff leaves both phase and held evidence intact.
module recordintent {
 const INITIAL: int
 var phase: int
 var reads: int
 var held: bool
 action init = all {phase'=INITIAL, reads'=0, held'=false}
 action local_intent = all {phase==0, phase'=phase, reads'=reads, held'=held}
 action stage_intent = all {or {phase==0,phase==2}, phase'=1, reads'=reads, held'=false}
 action run_intent = all {phase==1, phase'=2, reads'=reads+1, held'=true}
 action refuse_intent = all {phase==2, phase'=phase, reads'=reads, held'=held}
 val safety = and {phase>=0, phase<=2, reads>=0, held==(phase==2)}
}

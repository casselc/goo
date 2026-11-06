// Configuration gestures never acquire telemetry. Only the explicit population
// Run below grants a read; return witnesses completed use_template of that fixed view.
module service_templates_356 {
 const INITIAL: int
 var phase: int
 var reads: int
 var coherent: bool
 action init = all {phase'=INITIAL,reads'=0,coherent'=true}
 action service_context = all {phase==0,phase'=1,reads'=reads,coherent'=true}
 action use_template = all {phase==1,phase'=2,reads'=reads,coherent'=true}
 action inspect_template = all {phase==2,phase'=3,reads'=reads,coherent'=true}
 action stage_template = all {phase==3,phase'=4,reads'=reads,coherent'=true}
 action run_population = all {phase==4,phase'=5,reads'=reads+1,coherent'=true}
 action return_service = all {phase==5,reads==1,phase'=6,reads'=reads,coherent'=true}
 val safety = and {phase>=0,phase<=6,reads>=0,reads<=1,coherent}
}

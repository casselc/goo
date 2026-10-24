module distribution_lifecycle {
 const INITIAL: int
 var phase: int
 var reads: int
 var protocols: int
 var planners: int
 var coherent: bool
 action init = all {phase'=INITIAL,reads'=0,protocols'=0,planners'=0,coherent'=true}
 action stage = all {phase==0,phase'=1,reads'=reads,protocols'=protocols,planners'=planners,coherent'=true}
 action apply = all {phase==1,phase'=2,reads'=reads+2,protocols'=protocols+1,planners'=planners,coherent'=true}
 action brush = all {phase==2,phase'=3,reads'=reads,protocols'=protocols,planners'=planners,coherent'=true}
 action held = all {phase==3,phase'=4,reads'=reads,protocols'=protocols,planners'=planners+1,coherent'=true}
 action return = all {phase==4,phase'=5,reads'=reads,protocols'=protocols,planners'=planners,coherent'=true}
 action fresh_brush = all {phase==5,phase'=6,reads'=reads,protocols'=protocols,planners'=planners,coherent'=true}
 action fresh_stage = all {phase==6,phase'=7,reads'=reads,protocols'=protocols,planners'=planners+1,coherent'=true}
 action fresh_run = all {phase==7,phase'=8,reads'=reads+1,protocols'=protocols,planners'=planners,coherent'=true}
 val safety = coherent and reads>=0 and protocols>=0 and planners>=0
}

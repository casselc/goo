module tail_reviews {
 const INITIAL: int
 var phase: int
 var heads: int
 var active: int
 var visible: bool
 var immutable: bool
 action init = all {phase'=INITIAL,heads'=0,active'=0,visible'=false,immutable'=true}
 action publish = all {phase==0,phase'=1,heads'=1,active'=1,visible'=true,immutable'=true}
 action correct = all {phase==1,phase'=2,heads'=1,active'=1,visible'=true,immutable'=true}
 action withdraw = all {phase==2,phase'=3,heads'=1,active'=0,visible'=false,immutable'=true}
 action retry_original = all {phase==3,phase'=4,heads'=1,active'=0,visible'=false,immutable'=true}
 val safety = immutable and active<=heads
}

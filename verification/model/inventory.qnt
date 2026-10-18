module inventory {
 const INITIAL: int
 var phase: int
 var coherent: bool
 action init = all { phase'=INITIAL, coherent'=true }
 action capture = all { phase==0, phase'=1, coherent'=true }
 action search = all { phase==1, phase'=2, coherent'=true }
 action open = all { phase==2, phase'=3, coherent'=true }
 action package = all { phase==3, phase'=4, coherent'=true }
 val safety = coherent
}

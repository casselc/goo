module capture_catalog {
 const INITIAL: int
 var saved: bool
 var reads: int
 var coherent: bool
 action init = all {saved'=false,reads'=INITIAL,coherent'=true}
 action save = all {not(saved),saved'=true,reads'=2,coherent'=true}
 action browse = all {saved,saved'=saved,reads'=reads,coherent'=true}
 val safety = coherent and (not(saved) or reads == 2)
}

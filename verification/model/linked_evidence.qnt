module linked_evidence {
 const INITIAL: int
 var reads: int
 var linked: int
 var coherent: bool
 action init = all {reads'=INITIAL,linked'=0,coherent'=true}
 action save = all {reads'=reads,linked'=linked,coherent'=true}
 action capture = all {reads'=reads+1,linked'=linked+1,coherent'=true}
 action inspectRead = all {reads'=reads+1,linked'=linked,coherent'=true}
 action failPublish = all {reads'=reads+1,linked'=linked,coherent'=true}
 action recover = all {reads'=reads,linked'=linked+1,coherent'=true}
 action reopen = all {reads'=reads,linked'=linked,coherent'=true}
 val safety = coherent and linked<=reads and linked>=0
}

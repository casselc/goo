module saved_comparison {
 const INITIAL: int
 var reads: int
 var revisions: int
 var coherent: bool
 action init = all {reads'=INITIAL,revisions'=0,coherent'=true}
 action publish = all {reads'=reads,revisions'=revisions+1,coherent'=true}
 action reopen = all {reads'=reads,revisions'=revisions,coherent'=true}
 action failPublish = all {reads'=reads,revisions'=revisions,coherent'=true}
 action failStage = all {reads'=reads,revisions'=revisions,coherent'=true}
 val safety = coherent and reads==INITIAL and revisions>=0
}

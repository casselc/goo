module reusable_questions {
 const INITIAL: int
 var reads: int
 var revisions: int
 var coherent: bool
 action init = all {reads'=INITIAL,revisions'=1,coherent'=true}
 action stage = all {reads'=reads,revisions'=revisions,coherent'=true}
 action publish_failed = all {reads'=reads,revisions'=revisions,coherent'=true}
 action save = all {reads'=reads,revisions'=revisions+1,coherent'=true}
 action open = all {reads'=reads,revisions'=revisions,coherent'=true}
 action capture = all {reads'=reads+1,revisions'=revisions+1,coherent'=true}
 action fresh = all {reads'=reads,revisions'=revisions,coherent'=true}
 action execute = all {reads'=reads+1,revisions'=revisions,coherent'=true}
 val safety = coherent
}

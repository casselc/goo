module annotations {
 const INITIAL: int
 var reads: int
 var published: int
 var coherent: bool
 action init = all {reads'=INITIAL,published'=0,coherent'=true}
 action inspect = all {reads'=reads,published'=published,coherent'=true}
 action publish = all {reads'=reads,published'=published+1,coherent'=true}
 val safety = reads==0 and published>=0 and coherent
}

module kept_discovery {
 const INITIAL: int
 var sourceReads: int
 var payloadWrites: int
 var discoveryReads: int
 var exactReads: int
 var indexBatches: int
 var coherent: bool
 action init = all { sourceReads'=INITIAL, payloadWrites'=0, discoveryReads'=0, exactReads'=0, indexBatches'=0, coherent'=true }
 action discover = all {sourceReads'=sourceReads, payloadWrites'=payloadWrites, discoveryReads'=discoveryReads+1, exactReads'=exactReads, indexBatches'=indexBatches, coherent'=true}
 action open = all {sourceReads'=sourceReads, payloadWrites'=payloadWrites, discoveryReads'=discoveryReads, exactReads'=exactReads+1, indexBatches'=indexBatches, coherent'=true}
 action repair = all {sourceReads'=sourceReads, payloadWrites'=payloadWrites, discoveryReads'=discoveryReads, exactReads'=exactReads, indexBatches'=indexBatches+1, coherent'=true}
 val safety = sourceReads==0 and payloadWrites==0 and discoveryReads>=0 and exactReads>=0 and indexBatches>=0 and coherent
}

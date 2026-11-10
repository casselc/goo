module trace_continuation403 {
 const READS:List[int]
 const STORES:List[int]
 const CAPTURES:List[int]
 const DOCUMENTS:List[int]
 var phase:int
 var reads:int
 var stores:int
 var captures:int
 var documents:int
 action init=all{phase'=0,reads'=0,stores'=0,captures'=0,documents'=0}
 action step=all{phase<24,phase'=phase+1,reads'=reads+READS.nth(phase),stores'=stores+STORES.nth(phase),captures'=captures+CAPTURES.nth(phase),documents'=documents+DOCUMENTS.nth(phase)}
 val safety=phase>=0 and phase<=24 and reads>=0 and stores>=0 and captures>=0 and captures<=1 and documents>=0 and documents<=1
}

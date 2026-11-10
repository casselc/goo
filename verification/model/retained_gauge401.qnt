module retained_gauge401 {
 const READS:List[int]
 const STORES:List[int]
 var phase:int
 var reads:int
 var stores:int
 var writes:int
 action init=all{phase'=0,reads'=0,stores'=0,writes'=0}
 action step=all{phase<39,phase'=phase+1,reads'=reads+READS.nth(phase),stores'=stores+STORES.nth(phase),writes'=writes}
 val safety=phase>=0 and phase<=39 and reads>=0 and stores>=0 and stores<=6 and writes==0
}

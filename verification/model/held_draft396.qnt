module held_draft396 {
 const READS:List[int]
 const STORES:List[int]
 const WRITES:List[int]
 var phase:int
 var reads:int
 var stores:int
 var writes:int
 var finished:bool
 action init=all{phase'=0,reads'=0,stores'=0,writes'=0,finished'=false}
 action step=all{not(finished),phase<19,phase'=phase+1,reads'=reads+READS.nth(phase),stores'=stores+STORES.nth(phase),writes'=writes+WRITES.nth(phase),finished'=false}
 action finish=all{not(finished),phase==19,phase'=20,reads'=reads,stores'=stores,writes'=writes,finished'=true}
 val safety=reads>=0 and reads<=4 and stores>=0 and stores<=1 and writes<=3
}

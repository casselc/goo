module http_client_native392 {
 const READS:List[int]
 const STORES:List[int]
 const CONFIGS:List[int]
 const CAPTURES:List[int]
 var phase:int
 var reads:int
 var stores:int
 var configs:int
 var captures:int
 var finished:bool
 action init=all{phase'=0,reads'=0,stores'=0,configs'=0,captures'=0,finished'=false}
 action step=all{not(finished),phase<14,phase'=phase+1,reads'=reads+READS.nth(phase),stores'=stores+STORES.nth(phase),configs'=configs+CONFIGS.nth(phase),captures'=captures+CAPTURES.nth(phase),finished'=false}
 action finish=all{not(finished),phase==14,phase'=15,reads'=reads,stores'=stores,configs'=configs,captures'=captures,finished'=true}
 val safety=reads>=0 and stores>=0 and configs<=1 and captures<=1
}

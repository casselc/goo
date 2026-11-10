module grouped_mean398 {
 const READS:List[int]
 const STORES:List[int]
 const CONFIGS:List[int]
 const STEPS:int
 const READ_LIMIT:int
 const STORE_LIMIT:int
 const CONFIG_LIMIT:int
 var phase:int
 var reads:int
 var stores:int
 var configs:int
 var finished:bool
 action init=all{phase'=0,reads'=0,stores'=0,configs'=0,finished'=false}
 action step=all{not(finished),phase<STEPS,phase'=phase+1,reads'=reads+READS.nth(phase),stores'=stores+STORES.nth(phase),configs'=configs+CONFIGS.nth(phase),finished'=false}
 action finish=all{not(finished),phase==STEPS,phase'=phase+1,reads'=reads,stores'=stores,configs'=configs,finished'=true}
 val safety=phase>=0 and reads>=0 and reads<=READ_LIMIT and stores>=0 and stores<=STORE_LIMIT and configs>=0 and configs<=CONFIG_LIMIT
}

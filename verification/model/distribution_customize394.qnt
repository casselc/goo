module distribution_customize394 {
 const READS:List[int]
 const STORES:List[int]
 const CONFIGS:List[int]
 var phase:int
 var reads:int
 var stores:int
 var configs:int
 var finished:bool
 action init=all{phase'=0,reads'=0,stores'=0,configs'=0,finished'=false}
 action step=all{not(finished),phase<34,phase'=phase+1,reads'=reads+READS.nth(phase),stores'=stores+STORES.nth(phase),configs'=configs+CONFIGS.nth(phase),finished'=false}
 action finish=all{not(finished),phase==34,phase'=35,reads'=reads,stores'=stores,configs'=configs,finished'=true}
 val safety=reads>=0 and reads<=14 and stores>=0 and stores<=4 and configs<=2
}

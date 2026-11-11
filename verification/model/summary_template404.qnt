module summary_template404 {
 const READS:List[int]
 const STORES:List[int]
 const CONFIGS:List[int]
 const EXPANDS:List[int]
 var phase:int
 var reads:int
 var stores:int
 var configs:int
 var expands:int
 action init=all{phase'=0,reads'=0,stores'=0,configs'=0,expands'=0}
 action step=all{phase<33,phase'=phase+1,reads'=reads+READS.nth(phase),stores'=stores+STORES.nth(phase),configs'=configs+CONFIGS.nth(phase),expands'=expands+EXPANDS.nth(phase)}
 val safety=phase>=0 and phase<=33 and reads>=0 and reads<=10 and stores>=0 and configs>=0 and configs<=4 and expands>=0 and expands<=3
}

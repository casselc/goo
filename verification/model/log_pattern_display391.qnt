module log_pattern_display391 {
 const INSPECT_STORES:int
 var phase:int
 var reads:int
 var stores:int
 var configs:int
 var captures:int
 var finished:bool
 action init=all{phase'=0,reads'=0,stores'=0,configs'=0,captures'=0,finished'=false}
 action local=all{not(finished),phase'=phase+1,reads'=reads,stores'=stores,configs'=configs,captures'=captures,finished'=false}
 action read=all{not(finished),phase'=phase+1,reads'=reads+1,stores'=stores,configs'=configs,captures'=captures,finished'=false}
 action readPair=all{not(finished),phase'=phase+1,reads'=reads+2,stores'=stores,configs'=configs,captures'=captures,finished'=false}
 action save=all{not(finished),phase'=phase+1,reads'=reads,stores'=stores,configs'=configs+1,captures'=captures,finished'=false}
 action inspect=all{not(finished),phase'=phase+1,reads'=reads,stores'=stores+INSPECT_STORES,configs'=configs,captures'=captures,finished'=false}
 action finish=all{not(finished),phase==11,phase'=12,reads'=reads,stores'=stores,configs'=configs,captures'=captures,finished'=true}
 val safety=reads>=0 and reads<=4 and configs<=1 and captures==0 and stores<=INSPECT_STORES
}

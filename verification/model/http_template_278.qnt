// Template expansion stages configuration; only explicit Run acquires records.
module http_template_278 {
 const BASE_READS:int
 const MAX_READS:int
 var phase:int
 var reads:int
 var staged:bool
 action init=all{phase'=0,reads'=0,staged'=false}
 action service=all{phase==0,phase'=1,reads'=BASE_READS,staged'=false}
 action customize=all{phase==1,phase'=2,reads'=reads,staged'=true}
 action read=all{phase==2,phase'=3,reads'=reads+1,staged'=true}
 action edit=all{phase==3,phase'=4,reads'=reads,staged'=true}
 action save=all{phase==4,phase'=5,reads'=reads,staged'=true}
 action reopen=all{phase==5,phase'=6,reads'=reads,staged'=true}
 action rerun=all{phase==6,phase'=7,reads'=reads+1,staged'=true}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=7
}

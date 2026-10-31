// Public saved-display and source-read boundary; private source witnesses stay runtime-only.
module view_authoring_275 {
 const MAX_READS:int
 var phase:int
 var reads:int
 action init=all{phase'=0,reads'=0}
 action staged=all{phase==0,phase'=1,reads'=reads}
 action read=all{phase==1,phase'=2,reads'=reads+1}
 action edit=all{phase==2,phase'=3,reads'=reads}
 action invalid=all{phase==3,phase'=4,reads'=reads}
 action apply=all{phase==4,phase'=5,reads'=reads}
 action save=all{phase==5,phase'=6,reads'=reads}
 action reopen=all{phase==6,phase'=7,reads'=reads}
 action rerun=all{phase==7,phase'=8,reads'=reads+1}
 val safety=reads>=0 and reads<=MAX_READS and phase>=0 and phase<=8
}

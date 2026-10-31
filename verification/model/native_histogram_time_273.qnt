// Public selected bucket facts/effects, not private source authenticity.
module native_histogram_time_273 {
 const MAX_READS:int
 var reads:int
 var phase:int
 action init=all{reads'=0,phase'=0}
 action source=all{phase==0,reads'=1,phase'=1}
 action table=all{phase==1,reads'=reads,phase'=2}
 action zero=all{phase==2,reads'=reads,phase'=3}
 action unknown=all{phase==3,reads'=reads,phase'=4}
 action final=all{phase==4,reads'=reads,phase'=5}
 action saved=all{phase==5,reads'=reads,phase'=6}
 action complete=all{phase==6,reads'=reads,phase'=7}
 val safety=all{reads>=0,reads<=MAX_READS,phase>=0,phase<=7}
}

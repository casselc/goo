module http_method_facet_311 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var expands:int
 var saves:int
 var lookups:int
 var encodes:int
 action init=all{phase'=0,reads'=0,pages'=0,expands'=0,saves'=0,lookups'=0,encodes'=0}
 action owner=all{phase==0,phase'=1,reads'=0,pages'=0,expands'=0,saves'=0,lookups'=0,encodes'=0}
 action customize=all{phase==1,phase'=2,reads'=0,pages'=0,expands'=1,saves'=0,lookups'=0,encodes'=0}
 action runOriginal=all{phase==2,phase'=3,reads'=1,pages'=0,expands'=1,saves'=0,lookups'=0,encodes'=0}
 action valuesOriginal=all{phase==3,phase'=4,reads'=1,pages'=0,expands'=1,saves'=0,lookups'=1,encodes'=0}
 action presentation=all{phase==4,phase'=5,reads'=1,pages'=0,expands'=1,saves'=0,lookups'=1,encodes'=0}
 action saveOriginal=all{phase==5,phase'=6,reads'=1,pages'=0,expands'=1,saves'=1,lookups'=1,encodes'=0}
 action selectMethods=all{phase==6,phase'=7,reads'=1,pages'=0,expands'=1,saves'=1,lookups'=1,encodes'=0}
 action stageMethods=all{phase==7,phase'=8,reads'=1,pages'=0,expands'=1,saves'=1,lookups'=1,encodes'=0}
 action runSelected=all{phase==8,phase'=9,reads'=2,pages'=0,expands'=1,saves'=1,lookups'=1,encodes'=0}
 action valuesSelected=all{phase==9,phase'=10,reads'=2,pages'=0,expands'=1,saves'=1,lookups'=2,encodes'=0}
 action saveSelected=all{phase==10,phase'=11,reads'=2,pages'=0,expands'=1,saves'=2,lookups'=2,encodes'=0}
 action encodeArtifact=all{phase==11,phase'=12,reads'=2,pages'=0,expands'=1,saves'=2,lookups'=2,encodes'=1}
 action reopen=all{phase==12,phase'=13,reads'=2,pages'=0,expands'=1,saves'=2,lookups'=2,encodes'=1}
 action rerun=all{phase==13,phase'=14,reads'=3,pages'=0,expands'=1,saves'=2,lookups'=2,encodes'=1}
 action destination=all{phase==14,phase'=15,reads'=6,pages'=0,expands'=1,saves'=2,lookups'=2,encodes'=1}
 action applyTemplate=all{phase==15,phase'=16,reads'=6,pages'=0,expands'=2,saves'=2,lookups'=2,encodes'=1}
 action valuesDestination=all{phase==16,phase'=17,reads'=6,pages'=0,expands'=2,saves'=2,lookups'=3,encodes'=1}
 action saveDestination=all{phase==17,phase'=18,reads'=6,pages'=0,expands'=2,saves'=3,lookups'=3,encodes'=1}
 val safety=reads>=0 and reads<=MAX_READS and pages==0 and expands>=0 and expands<=2 and saves>=0 and saves<=3 and lookups>=0 and lookups<=3 and encodes>=0 and encodes<=1 and phase>=0 and phase<=18
}

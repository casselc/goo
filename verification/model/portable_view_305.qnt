module portable_view_305 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var encodes:int
 var decodes:int
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,encodes'=0,decodes'=0}
 action raw=all{phase==0,phase'=1,reads'=3,pages'=0,saves'=0,encodes'=0,decodes'=0}
 action rawRead=all{phase==1,phase'=2,reads'=3,pages'=1,saves'=0,encodes'=0,decodes'=0}
 action add=all{phase==2,phase'=3,reads'=3,pages'=1,saves'=0,encodes'=0,decodes'=0}
 action fullRead=all{phase==3,phase'=4,reads'=4,pages'=1,saves'=0,encodes'=0,decodes'=0}
 action saveSource=all{phase==4,phase'=5,reads'=4,pages'=1,saves'=1,encodes'=0,decodes'=0}
 action encodeArtifact=all{phase==5,phase'=6,reads'=4,pages'=1,saves'=1,encodes'=1,decodes'=0}
 action isolatedStore=all{phase==6,phase'=7,reads'=4,pages'=1,saves'=1,encodes'=1,decodes'=0}
 action preview=all{phase==7,phase'=8,reads'=4,pages'=1,saves'=1,encodes'=1,decodes'=1}
 action saveArtifact=all{phase==8,phase'=9,reads'=4,pages'=1,saves'=2,encodes'=1,decodes'=1}
 action open=all{phase==9,phase'=10,reads'=4,pages'=1,saves'=2,encodes'=1,decodes'=1}
 action runImported=all{phase==10,phase'=11,reads'=5,pages'=1,saves'=2,encodes'=1,decodes'=1}
 action destinationRead=all{phase==11,phase'=12,reads'=9,pages'=1,saves'=2,encodes'=1,decodes'=1}
 action apply=all{phase==12,phase'=13,reads'=9,pages'=1,saves'=2,encodes'=1,decodes'=1}
 action saveDestination=all{phase==13,phase'=14,reads'=9,pages'=1,saves'=3,encodes'=1,decodes'=1}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=1 and saves>=0 and saves<=3 and encodes>=0 and encodes<=1 and decodes>=0 and decodes<=1 and phase>=0 and phase<=14
}

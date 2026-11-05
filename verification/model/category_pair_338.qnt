// Observable endpoint effects; categorical authority is checked by generated tests.
module category_pair_338 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var lookups:int
 var compares:int
 var saves:int
 var encodes:int
 action init=all{phase'=0,reads'=0,lookups'=0,compares'=0,saves'=0,encodes'=0}
 action at(p:int,r:int,l:int,c:int,s:int,e:int):bool=all{phase==p,phase'=p+1,reads'=r,lookups'=l,compares'=c,saves'=s,encodes'=e}
 action owner=at(0,0,0,0,0,0)
 action runParent=at(1,2,0,1,0,0)
 action valuesParent=at(2,2,2,1,0,0)
 action stageBaseline=at(3,2,2,1,0,0)
 action applyBaseline=at(4,2,2,1,0,0)
 action runBaseline=at(5,4,2,2,0,0)
 action backBaseline=at(6,4,2,2,0,0)
 action stageExclude=at(7,4,2,2,0,0)
 action applyExclude=at(8,4,2,2,0,0)
 action runExcluded=at(9,6,2,3,0,0)
 action valuesExcluded=at(10,6,4,3,0,0)
 action stageInclude=at(11,6,4,3,0,0)
 action applyInclude=at(12,6,4,3,0,0)
 action runIncluded=at(13,8,4,4,0,0)
 action openGroup=at(14,8,4,4,0,0)
 action readGroup=at(15,9,4,4,0,0)
 action backGroup=at(16,9,4,4,0,0)
 action save=at(17,9,4,4,1,0)
 action encodeArtifact=at(18,9,4,4,1,1)
 action reopen=at(19,9,4,4,1,1)
 action runReopened=at(20,11,4,5,1,1)
 val safety=phase>=0 and phase<=21 and reads>=0 and reads<=MAX_READS and lookups>=0 and lookups<=4 and compares>=0 and compares<=5 and saves>=0 and saves<=1 and encodes>=0 and encodes<=1
}

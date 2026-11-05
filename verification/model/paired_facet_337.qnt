// Observed effects of the 337 native journey; AST authority belongs to Go/JS checks.
module paired_facet_337 {
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
 action stageExclude=at(3,2,2,1,0,0)
 action applyExclude=at(4,2,2,1,0,0)
 action runExcluded=at(5,4,2,2,0,0)
 action valuesExcluded=at(6,4,4,2,0,0)
 action stageInclude=at(7,4,4,2,0,0)
 action applyInclude=at(8,4,4,2,0,0)
 action runIncluded=at(9,6,4,3,0,0)
 action openGroup=at(10,6,4,3,0,0)
 action readGroup=at(11,7,4,3,0,0)
 action backGroup=at(12,7,4,3,0,0)
 action save=at(13,7,4,3,1,0)
 action encodeArtifact=at(14,7,4,3,1,1)
 action reopen=at(15,7,4,3,1,1)
 action runReopened=at(16,9,4,4,1,1)
 val safety=phase>=0 and phase<=17 and reads>=0 and reads<=MAX_READS and lookups>=0 and lookups<=4 and compares>=0 and compares<=4 and saves>=0 and saves<=1 and encodes>=0 and encodes<=1
}

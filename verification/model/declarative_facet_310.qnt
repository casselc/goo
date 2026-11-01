module declarative_facet_310 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var compares:int
 var saves:int
 var lookups:int
 var encodes:int
 var decodes:int
 action init=all{phase'=0,reads'=0,pages'=0,compares'=0,saves'=0,lookups'=0,encodes'=0,decodes'=0}
 action raw=all{phase==0,phase'=1,reads'=3,pages'=1,compares'=0,saves'=0,lookups'=0,encodes'=0,decodes'=0}
 action invalidFacet=all{phase==1,phase'=2,reads'=3,pages'=1,compares'=0,saves'=0,lookups'=0,encodes'=0,decodes'=0}
 action configureFacet=all{phase==2,phase'=3,reads'=3,pages'=1,compares'=0,saves'=0,lookups'=0,encodes'=0,decodes'=0}
 action save=all{phase==3,phase'=4,reads'=3,pages'=1,compares'=0,saves'=1,lookups'=0,encodes'=0,decodes'=0}
 action encodeArtifact=all{phase==4,phase'=5,reads'=3,pages'=1,compares'=0,saves'=1,lookups'=0,encodes'=1,decodes'=0}
 action previewArtifact=all{phase==5,phase'=6,reads'=3,pages'=1,compares'=0,saves'=1,lookups'=0,encodes'=1,decodes'=1}
 action importDefinition=all{phase==6,phase'=7,reads'=3,pages'=1,compares'=0,saves'=2,lookups'=0,encodes'=1,decodes'=1}
 action openImported=all{phase==7,phase'=8,reads'=3,pages'=1,compares'=0,saves'=2,lookups'=0,encodes'=1,decodes'=1}
 action runImported=all{phase==8,phase'=9,reads'=3,pages'=2,compares'=0,saves'=2,lookups'=0,encodes'=1,decodes'=1}
 action valuesParent=all{phase==9,phase'=10,reads'=3,pages'=2,compares'=0,saves'=2,lookups'=1,encodes'=1,decodes'=1}
 action externalOrigin=all{phase==10,phase'=11,reads'=3,pages'=2,compares'=0,saves'=2,lookups'=1,encodes'=1,decodes'=1}
 action clearOwned=all{phase==11,phase'=12,reads'=3,pages'=2,compares'=0,saves'=2,lookups'=1,encodes'=1,decodes'=1}
 action missingEmpty=all{phase==12,phase'=13,reads'=3,pages'=2,compares'=0,saves'=2,lookups'=1,encodes'=1,decodes'=1}
 action stageSelected=all{phase==13,phase'=14,reads'=3,pages'=2,compares'=0,saves'=2,lookups'=1,encodes'=1,decodes'=1}
 action runSelected=all{phase==14,phase'=15,reads'=3,pages'=3,compares'=0,saves'=2,lookups'=1,encodes'=1,decodes'=1}
 action stale=all{phase==15,phase'=16,reads'=3,pages'=3,compares'=0,saves'=2,lookups'=1,encodes'=1,decodes'=1}
 action valuesSelected=all{phase==16,phase'=17,reads'=3,pages'=3,compares'=0,saves'=2,lookups'=2,encodes'=1,decodes'=1}
 action destination=all{phase==17,phase'=18,reads'=6,pages'=4,compares'=0,saves'=2,lookups'=2,encodes'=1,decodes'=1}
 action applyDisplay=all{phase==18,phase'=19,reads'=6,pages'=4,compares'=0,saves'=2,lookups'=2,encodes'=1,decodes'=1}
 action valuesDestination=all{phase==19,phase'=20,reads'=6,pages'=4,compares'=0,saves'=2,lookups'=3,encodes'=1,decodes'=1}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=4 and compares==0 and saves>=0 and saves<=2 and lookups>=0 and lookups<=3 and encodes>=0 and encodes<=1 and decodes>=0 and decodes<=1 and phase>=0 and phase<=20
}

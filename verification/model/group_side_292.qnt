module group_side_292 {
 const MAX_FULL:int
 var phase:int
 var full:int
 var pages:int
 var baselines:int
 var compares:int
 var child:int
 var staged:int
 var held:int
 action init=all{phase'=0,full'=0,pages'=0,baselines'=0,compares'=0,child'=0,staged'=0,held'=0}
 action open=all{phase==0,phase'=1,full'=0,pages'=0,baselines'=0,compares'=0,child'=0,staged'=1,held'=0}
 action readPopulation=all{phase==1,phase'=2,full'=1,pages'=0,baselines'=1,compares'=1,child'=0,staged'=0,held'=6}
 action pageParent=all{phase==2,phase'=3,full'=1,pages'=1,baselines'=1,compares'=1,child'=0,staged'=0,held'=6}
 action stageBaseline=all{phase==3,phase'=4,full'=1,pages'=1,baselines'=1,compares'=1,child'=1,staged'=1,held'=6}
 action readBaseline=all{phase==4,phase'=5,full'=2,pages'=1,baselines'=1,compares'=1,child'=1,staged'=0,held'=6}
 action returnBaseline=all{phase==5,phase'=6,full'=2,pages'=1,baselines'=1,compares'=1,child'=0,staged'=0,held'=6}
 action rejectOld=all{phase==6,phase'=7,full'=2,pages'=1,baselines'=1,compares'=1,child'=0,staged'=0,held'=6}
 action stageCurrent=all{phase==7,phase'=8,full'=2,pages'=1,baselines'=1,compares'=1,child'=1,staged'=1,held'=6}
 action readCurrent=all{phase==8,phase'=9,full'=3,pages'=1,baselines'=1,compares'=1,child'=1,staged'=0,held'=6}
 action returnCurrent=all{phase==9,phase'=10,full'=3,pages'=1,baselines'=1,compares'=1,child'=0,staged'=0,held'=6}
 action stageMissing=all{phase==10,phase'=11,full'=3,pages'=1,baselines'=1,compares'=1,child'=1,staged'=1,held'=6}
 action readMissing=all{phase==11,phase'=12,full'=4,pages'=1,baselines'=1,compares'=1,child'=1,staged'=0,held'=6}
 action returnMissing=all{phase==12,phase'=13,full'=4,pages'=1,baselines'=1,compares'=1,child'=0,staged'=0,held'=6}
 action stageEmpty=all{phase==13,phase'=14,full'=4,pages'=1,baselines'=1,compares'=1,child'=1,staged'=1,held'=6}
 action readEmpty=all{phase==14,phase'=15,full'=5,pages'=1,baselines'=1,compares'=1,child'=1,staged'=0,held'=6}
 action returnEmpty=all{phase==15,phase'=16,full'=5,pages'=1,baselines'=1,compares'=1,child'=0,staged'=0,held'=6}
 action retire=all{phase==16,phase'=17,full'=5,pages'=1,baselines'=1,compares'=1,child'=0,staged'=1,held'=0}
 val safety=phase>=0 and phase<=17 and full>=0 and full<=MAX_FULL and pages>=0 and pages<=1 and baselines>=0 and baselines<=1 and compares>=0 and compares<=1 and child>=0 and child<=1 and staged>=0 and staged<=1 and held>=0 and held<=6
}

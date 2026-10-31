module authored_comparison_290 {
 const MAX_FULL:int
 var phase:int
 var full:int
 var pages:int
 var baselines:int
 var compares:int
 var saves:int
 var held:int
 var fresh:int
 var pair:int
 action init=all{phase'=0,full'=0,pages'=0,baselines'=0,compares'=0,saves'=0,held'=0,fresh'=0,pair'=0}
 action open=all{phase==0,phase'=1,full'=0,pages'=0,baselines'=0,compares'=0,saves'=0,held'=0,fresh'=0,pair'=0}
 action readPopulation=all{phase==1,phase'=2,full'=1,pages'=0,baselines'=0,compares'=0,saves'=0,held'=3,fresh'=3,pair'=0}
 action prepareBaseline=all{phase==2,phase'=3,full'=1,pages'=0,baselines'=1,compares'=1,saves'=0,held'=3,fresh'=3,pair'=1}
 action arrival=all{phase==3,phase'=4,full'=1,pages'=0,baselines'=1,compares'=1,saves'=0,held'=3,fresh'=3,pair'=1}
 action pageRecords=all{phase==4,phase'=5,full'=1,pages'=1,baselines'=1,compares'=1,saves'=0,held'=3,fresh'=4,pair'=1}
 action saveDefinition=all{phase==5,phase'=6,full'=1,pages'=1,baselines'=1,compares'=1,saves'=1,held'=3,fresh'=4,pair'=1}
 action reopenDefinition=all{phase==6,phase'=7,full'=1,pages'=1,baselines'=1,compares'=1,saves'=1,held'=0,fresh'=0,pair'=0}
 action readAgain=all{phase==7,phase'=8,full'=2,pages'=1,baselines'=2,compares'=2,saves'=1,held'=4,fresh'=4,pair'=1}
 action prepareZero=all{phase==8,phase'=9,full'=2,pages'=1,baselines'=3,compares'=3,saves'=1,held'=4,fresh'=4,pair'=1}
 action refuseIncompatible=all{phase==9,phase'=10,full'=2,pages'=1,baselines'=4,compares'=3,saves'=1,held'=4,fresh'=4,pair'=0}
 action refuseMissing=all{phase==10,phase'=11,full'=2,pages'=1,baselines'=5,compares'=3,saves'=1,held'=4,fresh'=4,pair'=0}
 action retirePreparation=all{phase==11,phase'=12,full'=2,pages'=1,baselines'=6,compares'=3,saves'=1,held'=0,fresh'=0,pair'=0}
 val safety=phase>=0 and phase<=12 and full>=0 and full<=MAX_FULL and pages>=0 and pages<=1 and baselines>=0 and baselines<=6 and compares>=0 and compares<=3 and saves>=0 and saves<=1 and held>=0 and held<=4 and fresh>=0 and fresh<=4 and pair>=0 and pair<=1
}

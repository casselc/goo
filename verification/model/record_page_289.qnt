module record_page_289 {
 const MAX_READS:int
 var phase:int
 var full:int
 var pages:int
 var keeps:int
 var captures:int
 var held:int
 var fresh:int
 action init=all{phase'=0,full'=0,pages'=0,keeps'=0,captures'=0,held'=0,fresh'=0}
 action open=all{phase==0,phase'=1,full'=0,pages'=0,keeps'=0,captures'=0,held'=0,fresh'=0}
 action readPopulation=all{phase==1,phase'=2,full'=1,pages'=0,keeps'=0,captures'=0,held'=3,fresh'=3}
 action arrival=all{phase==2,phase'=3,full'=1,pages'=0,keeps'=0,captures'=0,held'=3,fresh'=3}
 action page=all{phase==3,phase'=4,full'=1,pages'=1,keeps'=0,captures'=0,held'=3,fresh'=4}
 action keep=all{phase==4,phase'=5,full'=1,pages'=1,keeps'=1,captures'=0,held'=3,fresh'=4}
 action rejectPromotion=all{phase==5,phase'=6,full'=1,pages'=1,keeps'=1,captures'=0,held'=3,fresh'=4}
 action retire=all{phase==6,phase'=7,full'=1,pages'=2,keeps'=1,captures'=0,held'=0,fresh'=0}
 action ordinary=all{phase==7,phase'=8,full'=3,pages'=2,keeps'=1,captures'=0,held'=0,fresh'=4}
 action capture=all{phase==8,phase'=9,full'=3,pages'=2,keeps'=1,captures'=1,held'=0,fresh'=4}
 val safety=phase>=0 and phase<=9 and full>=0 and full<=MAX_READS and pages>=0 and pages<=2 and keeps>=0 and keeps<=1 and captures>=0 and captures<=1 and held>=0 and held<=3 and fresh>=0 and fresh<=4
}

module supplied_waterfall_301 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var expands:int
 var keeps:int
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,expands'=0,keeps'=0}
 action source=all{phase==0,phase'=1,reads'=1,pages'=0,saves'=0,expands'=1,keeps'=0}
 action configure=all{phase==1,phase'=2,reads'=1,pages'=0,saves'=0,expands'=1,keeps'=0}
 action read=all{phase==2,phase'=3,reads'=2,pages'=0,saves'=0,expands'=1,keeps'=0}
 action waterfall=all{phase==3,phase'=4,reads'=2,pages'=0,saves'=0,expands'=1,keeps'=0}
 action keep=all{phase==4,phase'=5,reads'=2,pages'=0,saves'=0,expands'=1,keeps'=1}
 action inspect=all{phase==5,phase'=6,reads'=2,pages'=0,saves'=0,expands'=1,keeps'=1}
 action trace=all{phase==6,phase'=7,reads'=4,pages'=0,saves'=0,expands'=1,keeps'=1}
 action saveTemplate=all{phase==7,phase'=8,reads'=4,pages'=0,saves'=1,expands'=1,keeps'=1}
 action target=all{phase==8,phase'=9,reads'=5,pages'=0,saves'=1,expands'=2,keeps'=1}
 action readTarget=all{phase==9,phase'=10,reads'=6,pages'=0,saves'=1,expands'=2,keeps'=1}
 action apply=all{phase==10,phase'=11,reads'=6,pages'=0,saves'=1,expands'=3,keeps'=1}
 action saveTarget=all{phase==11,phase'=12,reads'=6,pages'=0,saves'=2,expands'=3,keeps'=1}
 action reopen=all{phase==12,phase'=13,reads'=6,pages'=0,saves'=2,expands'=3,keeps'=1}
 action rerun=all{phase==13,phase'=14,reads'=7,pages'=0,saves'=2,expands'=3,keeps'=1}
 action page=all{phase==14,phase'=15,reads'=8,pages'=1,saves'=2,expands'=3,keeps'=1}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=1 and saves>=0 and saves<=2 and expands>=0 and expands<=3 and keeps>=0 and keeps<=1 and phase>=0 and phase<=15
}

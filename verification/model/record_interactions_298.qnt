module record_interactions_298 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var keeps:int
 var saves:int
 var kind:str
 var staged:bool
 var selectionCount:int
 action init=all{phase'=0,reads'=0,pages'=0,keeps'=0,saves'=0,kind'="record-table",staged'=false,selectionCount'=0}
 action service=all{phase==0,phase'=1,reads'=1,pages'=0,keeps'=0,saves'=0,kind'="record-table",staged'=false,selectionCount'=0}
 action customize=all{phase==1,phase'=2,reads'=1,pages'=0,keeps'=0,saves'=0,kind'="record-table",staged'=true,selectionCount'=0}
 action read=all{phase==2,phase'=3,reads'=2,pages'=0,keeps'=0,saves'=0,kind'="record-table",staged'=false,selectionCount'=0}
 action list=all{phase==3,phase'=4,reads'=2,pages'=0,keeps'=0,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action inspect=all{phase==4,phase'=5,reads'=2,pages'=0,keeps'=0,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action trace=all{phase==5,phase'=6,reads'=4,pages'=0,keeps'=0,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action selected=all{phase==6,phase'=7,reads'=4,pages'=0,keeps'=0,saves'=0,kind'="record-list",staged'=false,selectionCount'=2}
 action keep=all{phase==7,phase'=8,reads'=4,pages'=0,keeps'=1,saves'=0,kind'="record-list",staged'=false,selectionCount'=2}
 action facet=all{phase==8,phase'=9,reads'=4,pages'=0,keeps'=1,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action back=all{phase==9,phase'=10,reads'=4,pages'=0,keeps'=1,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action retired=all{phase==10,phase'=11,reads'=4,pages'=0,keeps'=1,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action page=all{phase==11,phase'=12,reads'=4,pages'=1,keeps'=1,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action pageKeep=all{phase==12,phase'=13,reads'=4,pages'=1,keeps'=2,saves'=0,kind'="record-list",staged'=false,selectionCount'=1}
 action table=all{phase==13,phase'=14,reads'=4,pages'=1,keeps'=2,saves'=0,kind'="record-table",staged'=false,selectionCount'=0}
 action listAgain=all{phase==14,phase'=15,reads'=4,pages'=1,keeps'=2,saves'=0,kind'="record-list",staged'=false,selectionCount'=0}
 action save=all{phase==15,phase'=16,reads'=4,pages'=1,keeps'=2,saves'=1,kind'="record-list",staged'=false,selectionCount'=0}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=1 and keeps>=0 and keeps<=2 and saves>=0 and saves<=1 and selectionCount>=0 and selectionCount<=20 and phase>=0 and phase<=16
}

module record_list_297 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var kind:str
 var staged:bool
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,kind'="record-table",staged'=false}
 action service=all{phase==0,phase'=1,reads'=1,pages'=0,saves'=0,kind'="record-table",staged'=false}
 action customize=all{phase==1,phase'=2,reads'=reads,pages'=pages,saves'=saves,kind'="record-table",staged'=true}
 action readSource=all{phase==2,phase'=3,reads'=reads+1,pages'=pages,saves'=saves,kind'="record-table",staged'=false}
 action selectionRefusal=all{phase==3,phase'=4,reads'=reads,pages'=pages,saves'=saves,kind'="record-table",staged'=false}
 action list=all{phase==4,phase'=5,reads'=reads,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 action details=all{phase==5,phase'=6,reads'=reads,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 action saveTemplate=all{phase==6,phase'=7,reads'=reads,pages'=pages,saves'=saves+1,kind'="record-list",staged'=false}
 action target=all{phase==7,phase'=8,reads'=reads+1,pages'=pages,saves'=saves,kind'="record-table",staged'=true}
 action groupTarget=all{phase==8,phase'=9,reads'=reads,pages'=pages,saves'=saves,kind'="record-table",staged'=true}
 action readTarget=all{phase==9,phase'=10,reads'=reads+1,pages'=pages,saves'=saves,kind'="record-table",staged'=false}
 action preview=all{phase==10,phase'=11,reads'=reads,pages'=pages,saves'=saves,kind'="record-table",staged'=false}
 action applyList=all{phase==11,phase'=12,reads'=reads,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 action readApplied=all{phase==12,phase'=13,reads'=reads+1,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 action arrival=all{phase==13,phase'=14,reads'=reads,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 action page=all{phase==14,phase'=15,reads'=reads,pages'=pages+1,saves'=saves,kind'="record-list",staged'=false}
 action retiredControl=all{phase==15,phase'=16,reads'=reads,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 action tableLocal=all{phase==16,phase'=17,reads'=reads,pages'=pages,saves'=saves,kind'="record-table",staged'=false}
 action listLocal=all{phase==17,phase'=18,reads'=reads,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 action saveDestination=all{phase==18,phase'=19,reads'=reads,pages'=pages,saves'=saves+1,kind'="record-list",staged'=false}
 action reopen=all{phase==19,phase'=20,reads'=reads,pages'=pages,saves'=saves,kind'="record-list",staged'=true}
 action rerun=all{phase==20,phase'=21,reads'=reads+1,pages'=pages,saves'=saves,kind'="record-list",staged'=false}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=1 and saves>=0 and saves<=2 and phase>=0 and phase<=21
}

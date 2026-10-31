module preserve_attachment_293 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var writes:int
 var cards:int
 action init=all{phase'=0,reads'=0,writes'=0,cards'=0}
 action preserve=all{phase==0,phase'=1,reads'=1,writes'=0,cards'=0}
 action closedRefusal=all{phase==1,phase'=2,reads'=reads,writes'=writes,cards'=0}
 action hiddenParentRefusal=all{phase==2,phase'=3,reads'=reads,writes'=writes,cards'=0}
 action newSaveRefusal=all{phase==3,phase'=4,reads'=reads,writes'=writes,cards'=0}
 action freshSaveAttached=all{phase==4,phase'=5,reads'=reads,writes'=writes+1,cards'=1}
 action navigationRefusal=all{phase==5,phase'=6,reads'=reads,writes'=writes,cards'=0}
 action pendingRefusal=all{phase==6,phase'=7,reads'=reads,writes'=writes,cards'=0}
 action attach=all{phase==7,phase'=8,reads'=reads,writes'=writes,cards'=1}
 action saveDocument=all{phase==8,phase'=9,reads'=reads,writes'=writes+1,cards'=1}
 val safety=reads>=0 and reads<=MAX_READS and writes>=0 and writes<=2 and cards>=0 and cards<=1
}

module preservation_288 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var captures:int
 var staged:bool
 var questions:int
 var views:int
 var evidence:int
 action init=all{phase'=0,reads'=0,captures'=0,staged'=true,questions'=0,views'=0,evidence'=0}
 action openQuestion=all{phase==0,phase'=1,reads'=0,captures'=0,staged'=true,questions'=0,views'=0,evidence'=0}
 action shareQuestion=all{phase==1,phase'=2,reads'=0,captures'=0,staged'=true,questions'=0,views'=0,evidence'=0}
 action pinQuestion=all{phase==2,phase'=3,reads'=0,captures'=0,staged'=true,questions'=1,views'=0,evidence'=0}
 action saveQuestion=all{phase==3,phase'=4,reads'=0,captures'=0,staged'=true,questions'=1,views'=0,evidence'=0}
 action queryQuestion=all{phase==4,phase'=5,reads'=1,captures'=0,staged'=false,questions'=1,views'=0,evidence'=0}
 action openCapture=all{phase==5,phase'=6,reads'=1,captures'=0,staged'=false,questions'=1,views'=0,evidence'=0}
 action captureEvidence=all{phase==6,phase'=7,reads'=1,captures'=1,staged'=false,questions'=1,views'=0,evidence'=0}
 action reopenEvidence=all{phase==7,phase'=8,reads'=1,captures'=1,staged'=false,questions'=1,views'=0,evidence'=0}
 action pinEvidence=all{phase==8,phase'=9,reads'=1,captures'=1,staged'=false,questions'=1,views'=0,evidence'=1}
 action multiView=all{phase==9,phase'=10,reads'=1,captures'=1,staged'=true,questions'=1,views'=0,evidence'=1}
 action chooseSignal=all{phase==10,phase'=11,reads'=1,captures'=1,staged'=true,questions'=1,views'=0,evidence'=1}
 action singleView=all{phase==11,phase'=12,reads'=1,captures'=1,staged'=true,questions'=1,views'=0,evidence'=1}
 action saveView=all{phase==12,phase'=13,reads'=1,captures'=1,staged'=true,questions'=1,views'=0,evidence'=1}
 action pinView=all{phase==13,phase'=14,reads'=1,captures'=1,staged'=true,questions'=1,views'=1,evidence'=1}
 action saveFinding=all{phase==14,phase'=15,reads'=1,captures'=1,staged'=true,questions'=1,views'=1,evidence'=1}
 val safety=reads>=0 and reads<=MAX_READS and captures>=0 and captures<=1 and phase>=0 and phase<=15 and questions>=0 and questions<=1 and views>=0 and views<=1 and evidence>=0 and evidence<=captures
}

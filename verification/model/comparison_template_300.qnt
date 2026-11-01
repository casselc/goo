// Comparison reuse never imports a specimen baseline; preparation is explicit.
module comparison_template_300 {
 const MAX_READS:int
 var phase:int
 var reads:int
 var pages:int
 var saves:int
 var captures:int
 var compares:int
 var expands:int
 var captureReads:int
 action init=all{phase'=0,reads'=0,pages'=0,saves'=0,captures'=0,compares'=0,expands'=0,captureReads'=0}
 action source=all{phase==0,phase'=1,reads'=reads+1,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands+1,captureReads'=captureReads}
 action readSource=all{phase==1,phase'=2,reads'=reads+1,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action captureSource=all{phase==2,phase'=3,reads'=reads,pages'=pages,saves'=saves,captures'=captures+1,compares'=compares,expands'=expands,captureReads'=captureReads}
 action prepareSource=all{phase==3,phase'=4,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares+1,expands'=expands,captureReads'=captureReads+1}
 action saveTemplate=all{phase==4,phase'=5,reads'=reads,pages'=pages,saves'=saves+1,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action target=all{phase==5,phase'=6,reads'=reads+1,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands+1,captureReads'=captureReads}
 action readTarget=all{phase==6,phase'=7,reads'=reads+1,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action captureTarget=all{phase==7,phase'=8,reads'=reads,pages'=pages,saves'=saves,captures'=captures+1,compares'=compares,expands'=expands,captureReads'=captureReads}
 action refreshTarget=all{phase==8,phase'=9,reads'=reads+1,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action missing=all{phase==9,phase'=10,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action previewWrong=all{phase==10,phase'=11,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands+1,captureReads'=captureReads}
 action applyWrong=all{phase==11,phase'=12,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action rejectWrong=all{phase==12,phase'=13,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads+1}
 action retireOld=all{phase==13,phase'=14,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands+1,captureReads'=captureReads}
 action previewTarget=all{phase==14,phase'=15,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands+1,captureReads'=captureReads}
 action applyTarget=all{phase==15,phase'=16,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action prepareTarget=all{phase==16,phase'=17,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares+1,expands'=expands,captureReads'=captureReads+1}
 action page=all{phase==17,phase'=18,reads'=reads,pages'=pages+1,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action saveTarget=all{phase==18,phase'=19,reads'=reads,pages'=pages,saves'=saves+1,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action reopen=all{phase==19,phase'=20,reads'=reads,pages'=pages,saves'=saves,captures'=captures,compares'=compares,expands'=expands,captureReads'=captureReads}
 action rerun=all{phase==20,phase'=21,reads'=reads+1,pages'=pages,saves'=saves,captures'=captures,compares'=compares+1,expands'=expands,captureReads'=captureReads+1}
 val safety=reads>=0 and reads<=MAX_READS and pages>=0 and pages<=1 and saves>=0 and saves<=2 and captures>=0 and captures<=2 and compares>=0 and compares<=3 and expands>=0 and expands<=5 and captureReads>=0 and captureReads<=4 and phase>=0 and phase<=21
}

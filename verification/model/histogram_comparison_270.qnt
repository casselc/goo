// Observed independent operands and local/read effects, not private capability proof.
module histogram_comparison_270 {
 const CAPTURED_TOTAL: int
 var reads:int
 var plans:int
 var scale:str
 var baselineShare:str
 var selectionShare:str
 var started:bool
 action init=all{reads'=0,plans'=0,scale'="",baselineShare'="",selectionShare'="",started'=false}
 action source=all{not(started),reads'=2,plans'=0,scale'="count",baselineShare'="25",selectionShare'="50",started'=true}
 action population=all{started,reads'=reads,plans'=plans,scale'="population-share",baselineShare'="25",selectionShare'="50",started'=started}
 action valid=all{started,reads'=reads,plans'=plans,scale'="valid-share",baselineShare'="50",selectionShare'="200/3",started'=started}
 action table=all{started,reads'=reads,plans'=plans,scale'=scale,baselineShare'=baselineShare,selectionShare'=selectionShare,started'=started}
 action brush=table
 action open=all{started,reads'=reads+1,plans'=plans+1,scale'=scale,baselineShare'=baselineShare,selectionShare'=selectionShare,started'=started}
 val safety=all{CAPTURED_TOTAL==4,reads>=0,reads<=3,plans>=0,plans<=1}
}

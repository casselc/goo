// Stored-page presentation never acquires telemetry. Capture-store GETs are
// separately counted metadata reads; bounded pages never become whole traces.
module saved_waterfall_360 {
 const INITIAL:int
 var phase:int
 var stores:int
 var source:int
 var coherent:bool
 action init=all{phase'=INITIAL,stores'=0,source'=0,coherent'=true}
 action reopen=all{phase==0,phase'=1,stores'=1,source'=source,coherent'=true}
 action waterfall=all{phase==1,phase'=2,stores'=stores,source'=source,coherent'=true}
 action columns=all{phase==2,phase'=3,stores'=stores,source'=source,coherent'=true}
 action inspect_occurrence=all{phase==3,phase'=4,stores'=stores,source'=source,coherent'=true}
 action close_inspector=all{phase==4,phase'=5,stores'=stores,source'=source,coherent'=true}
 action table=all{phase==5,phase'=6,stores'=stores,source'=source,coherent'=true}
 action page_next=all{phase==6,phase'=7,stores'=stores+1,source'=source,coherent'=true}
 action page_previous=all{phase==7,phase'=8,stores'=stores+1,source'=source,coherent'=true}
 action return_inspect=all{phase==8,phase'=9,stores'=stores,source'=source,coherent'=true}
 val safety=and{phase>=0,phase<=9,stores>=0,stores<=3,source==0,coherent}
}

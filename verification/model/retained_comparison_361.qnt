// Actual retained-page visit protocol. Store GETs are distinct from telemetry.
module retained_comparison_361 {
 const INITIAL:int
 var phase:int
 var stores:int
 var source:int
 var paired:bool
 var coherent:bool
 action init=all{phase'=INITIAL,stores'=0,source'=0,paired'=false,coherent'=true}
 action open_reference=all{phase==0,phase'=1,stores'=stores+2,source'=source,paired'=false,coherent'=true}
 action pin_reference=all{phase==1,phase'=2,stores'=stores+0,source'=source,paired'=false,coherent'=true}
 action open_selection=all{phase==2,phase'=3,stores'=stores+1,source'=source,paired'=false,coherent'=true}
 action compare_pages=all{phase==3,phase'=4,stores'=stores+0,source'=source,paired'=true,coherent'=true}
 action open_baseline_side=all{phase==4,phase'=5,stores'=stores+1,source'=source,paired'=true,coherent'=true}
 action inspect_return_baseline=all{phase==5,phase'=6,stores'=stores+0,source'=source,paired'=true,coherent'=true}
 action open_selection_side=all{phase==6,phase'=7,stores'=stores+1,source'=source,paired'=true,coherent'=true}
 action inspect_return_selection=all{phase==7,phase'=8,stores'=stores+0,source'=source,paired'=true,coherent'=true}
 action candidate_page=all{phase==8,phase'=9,stores'=stores+1,source'=source,paired'=true,coherent'=true}
 action replace_baseline=all{phase==9,phase'=10,stores'=stores+0,source'=source,paired'=false,coherent'=true}
 action clear_reference=all{phase==10,phase'=11,stores'=stores+0,source'=source,paired'=false,coherent'=true}
 action open_reference_aba=all{phase==11,phase'=12,stores'=stores+2,source'=source,paired'=false,coherent'=true}
 action pin_reference_aba=all{phase==12,phase'=13,stores'=stores+0,source'=source,paired'=false,coherent'=true}
 action open_selection_aba=all{phase==13,phase'=14,stores'=stores+1,source'=source,paired'=false,coherent'=true}
 action compare_aba=all{phase==14,phase'=15,stores'=stores+0,source'=source,paired'=true,coherent'=true}
 action old_action_aba=all{phase==15,phase'=16,stores'=stores+0,source'=source,paired'=true,coherent'=true}
 action open_baseline_aba=all{phase==16,phase'=17,stores'=stores+1,source'=source,paired'=true,coherent'=true}
 action inspect_return_aba=all{phase==17,phase'=18,stores'=stores+0,source'=source,paired'=true,coherent'=true}
 val safety=and{phase>=0,phase<=18,stores>=0,stores<=10,source==0,coherent}
}

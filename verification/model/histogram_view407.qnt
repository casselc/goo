module histogram_view407 {
 const READS:List[int]
 const PLANS:List[int]
 const LENGTH:int
 var phase:int
 var reads:int
 var plans:int
 action init=all{phase'=0,reads'=0,plans'=0}
 action step=all{phase<LENGTH,phase'=phase+1,reads'=reads+READS.nth(phase),plans'=plans+PLANS.nth(phase)}
 val safety=phase>=0 and phase<=LENGTH and reads>=0 and plans>=0
}

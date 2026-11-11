module record_rate406 {
 const READS:List[int]
 const SQL:List[int]
 var phase:int
 var reads:int
 var sql:int
 action init=all{phase'=0,reads'=0,sql'=0}
 action step=all{phase<24,phase'=phase+1,reads'=reads+READS.nth(phase),sql'=sql+SQL.nth(phase)}
 val safety=phase>=0 and phase<=24 and reads>=0 and reads<=8 and sql>=0 and sql<=28
}

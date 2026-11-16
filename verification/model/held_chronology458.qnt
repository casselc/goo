// Local presentation ownership over two independent, already observed tails.
// Counts are acquisition starts, never telemetry populations or completeness.
module held_chronology458 {
  const INITIAL: int
  var logsReads: int
  var spansReads: int
  var held: bool
  var away: bool
  var intact: bool

  action init = all {logsReads'=INITIAL, spansReads'=INITIAL, held'=false, away'=false, intact'=true}
  action start = all {not(away), logsReads'=logsReads+1, spansReads'=spansReads+1,
    held'=false, away'=false, intact'=true}
  action resumeLogs = all {not(away), logsReads'=logsReads+1, spansReads'=spansReads,
    held'=false, away'=false, intact'=true}
  action resumeSpans = all {not(away), logsReads'=logsReads, spansReads'=spansReads+1,
    held'=false, away'=false, intact'=true}
  action hold = all {not(away), logsReads'=logsReads, spansReads'=spansReads,
    held'=true, away'=false, intact'=true}
  action inspect = all {held, not(away), logsReads'=logsReads, spansReads'=spansReads,
    held'=held, away'=away, intact'=true}
  action find = inspect
  action page = inspect
  action analyze = all {held, not(away), logsReads'=logsReads, spansReads'=spansReads,
    held'=true, away'=true, intact'=true}
  action back = all {held, away, logsReads'=logsReads, spansReads'=spansReads,
    held'=true, away'=false, intact'=true}
  action lanes = all {not(away), logsReads'=logsReads, spansReads'=spansReads,
    held'=false, away'=false, intact'=true}
  action retire = lanes
  action observe = all {logsReads'=logsReads, spansReads'=spansReads,
    held'=held, away'=away, intact'=true}
  action step = any {start, resumeLogs, resumeSpans, hold, inspect, find, page,
    analyze, back, lanes, retire, observe}
  val safety = intact and logsReads>=0 and spansReads>=0 and (not(away) or held)
}

module held_chronology458_instance {
  import held_chronology458(INITIAL=0).*
}

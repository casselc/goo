// Bounded workload/observation schedule. IDs are bound to real results by the
// browser driver; the model never invents telemetry or application outcomes.
module browser_campaign {
  const REPORTS: int
  const REVISITS: int
  var issued: int
  var checked: int
  var revisits: int
  var target: int
  var done: bool
  action init = all {issued'=0, checked'=0, revisits'=0, target'=0, done'=false}
  action issue = all {
    not(done), issued == checked, issued < REPORTS,
    issued'=issued+1, checked'=checked, revisits'=revisits, target'=issued, done'=false
  }
  action inspect = all {
    not(done), issued == checked+1,
    issued'=issued, checked'=issued, revisits'=revisits, target'=issued-1, done'=false
  }
  action revisit = {
    nondet chosen = 0.to(checked-1).oneOf()
    all {
      not(done), issued == checked, checked > 0, revisits < REVISITS,
      issued'=issued, checked'=checked, revisits'=revisits+1, target'=chosen, done'=false
    }
  }
  action finish = all {
    not(done), issued == REPORTS, checked == REPORTS, revisits == REVISITS,
    issued'=issued, checked'=checked, revisits'=revisits, target'=target, done'=true
  }
  action idle = all {
    done, issued'=issued, checked'=checked, revisits'=revisits, target'=target, done'=true
  }
  action step = any {issue, inspect, revisit, finish, idle}
  val safety = checked >= 0 and checked <= issued and issued <= REPORTS and
    issued-checked <= 1 and revisits >= 0 and revisits <= REVISITS
}
module browserCampaignInstance {
  import browser_campaign(REPORTS=3, REVISITS=2).*
}

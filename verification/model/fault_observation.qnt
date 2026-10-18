module fault_observation {
  const INITIAL: int
  var phase: int
  var subject: str
  var scenario: str
  var signature: str
  var meaning: str
  action init = all {phase'=INITIAL, subject'="", scenario'="", signature'="", meaning'=""}
  action accepted(id: str, kind: str): bool = all {
    phase == 0, id != "", Set("retry", "reject", "missing").contains(kind),
    phase'=1, subject'=id, scenario'=kind, signature'=signature, meaning'=meaning
  }
  action source_ready(id: str, sig: str, retrySeen: bool, stalled: bool, terminalSeen: bool, terminal: str, attempts: int, errors: int): bool = all {
    phase == 1, id == subject, sig != "",
    any {
      scenario == "retry" and retrySeen and not(stalled) and terminalSeen and terminal == "confirmed" and attempts == 2 and errors > 0,
      scenario == "reject" and not(retrySeen) and not(stalled) and terminalSeen and terminal == "rejected" and attempts == 1 and errors == 0,
      scenario == "missing" and not(retrySeen) and stalled and not(terminalSeen) and terminal == "" and attempts == 1 and errors == 0
    },
    phase'=2, subject'=subject, scenario'=scenario, signature'=sig,
    meaning'=if (scenario == "retry") "confirmed-after-retry" else if (scenario == "reject") "rejected" else "unknown-no-terminal"
  }
  action displayed(id: str, sig: str, observed: str): bool = all {
    phase == 2, id == subject, sig == signature, observed == meaning,
    phase'=3, subject'=subject, scenario'=scenario, signature'=signature, meaning'=meaning
  }
  action explored(steps: int, passed: bool): bool = all {
    phase == 3, steps > 0, passed,
    phase'=4, subject'=subject, scenario'=scenario, signature'=signature, meaning'=meaning
  }
  val honest = phase < 2 or Set("confirmed-after-retry", "rejected", "unknown-no-terminal").contains(meaning)
}

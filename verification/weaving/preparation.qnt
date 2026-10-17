module preparation {
  const MIN_REQUESTS: int
  var selected: int
  var facts: int
  action init = all { selected' = 0, facts' = 0 }
  action prepare_http(n: int, f: int): bool = all {
    n >= MIN_REQUESTS, selected' = n, facts' = f
  }
  val safety = selected >= MIN_REQUESTS and facts == selected
}

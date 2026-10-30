// Bounded ownership protocol: only explicit successful Run advances applied data.
// Draft staging, local display/selection and refused widgets do not read sources.
module scoped_view_261 {
  const LIMIT: int
  var reads: int
  var generation: int
  action init = all { reads' = 0, generation' = 0 }
  action applyRun = all { generation < LIMIT, reads' = reads + 1, generation' = generation + 1 }
  action switch = all { generation > 0, reads' = reads, generation' = generation }
  action selectRow = all { generation > 0, reads' = reads, generation' = generation }
  action stage = all { generation > 0, reads' = reads, generation' = generation }
  action refuse = all { generation > 0, reads' = reads, generation' = generation }
  action step = any { applyRun, switch, selectRow, stage, refuse }
  val safety = reads == generation and generation >= 0 and generation <= LIMIT
}

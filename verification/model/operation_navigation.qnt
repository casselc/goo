// A bounded fixture model of captured-evidence navigation, not trace completeness.
module operation_navigation {
 const BASIS: int
 var basis: int
 var selected: int
 action init = all { basis' = BASIS, selected' = 7 }
 // Fixture IDs 1,2,3 encoded as bits 1,2,4. Error IDs are 1,2;
 // one selected duration cell contains ID 2; its trace contains IDs 2,3.
 action select_errors = all { basis' = basis, selected' = 3 }
 action select_cell = all { basis' = basis, selected' = 2 }
 action open_trace = all { basis' = basis, selected' = 6 }
 action follow_subject = all { basis' = basis, selected' = 3 }
 action reset = all { basis' = basis, selected' = 7 }
 val safety = basis == BASIS and Set(2,3,6,7).contains(selected)
}

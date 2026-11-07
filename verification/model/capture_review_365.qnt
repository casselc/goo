module capture_review_365 {
 const INITIAL: int
 var source: int
 var task: bool
 var eligible: bool
 var pinned: bool
 action init = all {source'=INITIAL,task'=false,eligible'=false,pinned'=false}
 action task_ready = all {not(task),task'=true,eligible'=false,pinned'=false,source'=source}
 action label_reviewed = all {task,task'=task,eligible'=true,pinned'=pinned,source'=source}
 action label_unreviewed = all {task,task'=task,eligible'=false,pinned'=pinned,source'=source}
 action uncertain = all {task,task'=task,eligible'=false,pinned'=pinned,source'=source}
 action abstain = all {task,task'=task,eligible'=false,pinned'=pinned,source'=source}
 action disputed = all {task,task'=task,eligible'=false,pinned'=pinned,source'=source}
 action withdrawn = all {task,task'=task,eligible'=false,pinned'=pinned,source'=source}
 action pin = all {task,eligible,task'=task,eligible'=eligible,pinned'=true,source'=source}
 action export_pinned = all {task,pinned,task'=task,eligible'=eligible,pinned'=pinned,source'=source}
 action history = all {task'=task,eligible'=eligible,pinned'=pinned,source'=source}
 val safety = source==INITIAL and (not(pinned) or task)
}

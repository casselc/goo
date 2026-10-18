module investigations {
 const INITIAL: int
 var revisions: int
 var original: bool
 var fixed: bool
 action init = all {revisions'=INITIAL,original'=true,fixed'=true}
 action save = all {revisions==0,revisions'=1,original'=true,fixed'=true}
 action reopen = all {revisions>0,revisions'=revisions,original'=true,fixed'=true}
 action refresh = all {revisions>0,revisions'=revisions+1,original'=true,fixed'=true}
 action reject = all {revisions>0,revisions'=revisions,original'=true,fixed'=true}
 val safety = revisions>=0 and original and fixed
}

module comparison {
 const INITIAL: int
 var comparable: bool
 var exact: bool
 var spread: int
 action init = all {comparable'=false,exact'=false,spread'=INITIAL}
 action aligned = all {comparable'=true,exact'=true,spread'=0}
 action skew = all {comparable'=true,exact'=false,spread'=1}
 action outside = all {comparable'=false,exact'=false,spread' = -1}
 action loss = all {comparable'=false,exact'=false,spread' = -1}
 action reset = all {comparable'=false,exact'=false,spread' = -1}
 action recovered = all {comparable'=true,exact'=false,spread'=1}
 val safety = (not(comparable) and not(exact) and spread == -1) or (comparable and spread>=0 and spread<=2 and (not(exact) or spread==0))
}

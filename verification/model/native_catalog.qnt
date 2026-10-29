module native_catalog {
 const INITIAL: int
 const LIMIT: int
 var count: int
 action init = count' = INITIAL
 action publish = all {count < LIMIT, count' = count + 1}
 action refuse = all {count == LIMIT, count' = count}
 action duplicate = count' = count
 val safety = count >= INITIAL and count <= LIMIT
}

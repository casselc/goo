/* Reported metadata labels only: no readiness, continuity or source policy. */
export const isSumKind=kind=>kind==='counter'||kind==='stock';
export const temporalityLabel=value=>value===1?'Delta':value===2?'Cumulative':value===0?'Unspecified (0)':'Unrecognized ('+value+')';
export const sumReportingLabel=instrument=>temporalityLabel(instrument.temporality)+' · '+(instrument.monotonic?'Monotonic':'Nonmonotonic');

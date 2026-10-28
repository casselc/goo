const {test}=require('node:test');const assert=require('node:assert/strict');const {coverageLabel}=require('../ui/fleet.mjs');
test('fleet coverage labels count supported members without claiming ingestion completeness',()=>{assert.equal(coverageLabel({supported:1,expected:2}),'1/2 members support this exact interval');assert.equal(coverageLabel({supported:0,expected:2}),'0/2 members support this exact interval');});

test('bounded skew labels never imply a combined population',()=>{const {comparisonLabel}=require('../ui/fleet.mjs');assert.equal(comparisonLabel({state:'comparable'}),'Comparable windows · no combined percentage');assert.match(comparisonLabel({state:'unknown'}),/no supported match/);});

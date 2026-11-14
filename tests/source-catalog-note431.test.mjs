import {test} from 'node:test';
import assert from 'node:assert/strict';
import {sourceNote} from '../ui/source-capabilities.mjs';
test('service and instrument catalogs describe excluded families without making them empty',()=>{
 assert.equal(sourceNote([{kind:'logs',state:'compatible'}]),'');
 assert.equal(sourceNote([{kind:'logs',state:'compatible'},{kind:'spans',state:'absent'},{kind:'gauge',state:'unreadable'},{kind:'counter',state:'incompatible'}]),'Sources not included: Spans (table absent), Gauges (source unavailable), Counters (unsupported schema). Available sources only.');
 assert.equal(sourceNote([{kind:'password secret',state:'absent'},{kind:'logs',state:'secret'}]),'');
});

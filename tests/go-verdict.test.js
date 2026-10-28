const {test} = require('node:test');
const assert = require('node:assert/strict');
const fixtures = require('./fixtures/go-test-verdicts.json');
const {validateGoEvents} = require('../journeys/support/go-verdict');
for (const fixture of fixtures) test('required Go verdict: ' + fixture.name, () => {
  const validate = () => validateGoEvents(fixture.output, fixture.expected);
  if (fixture.accept) assert.doesNotThrow(validate);
  else assert.throws(validate);
});

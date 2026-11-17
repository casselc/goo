const assert = require('node:assert/strict');
// The chart preset is explicitly route-only; the live table keeps joint groups.
function routeTemplateQuestion(question) {
  assert.deepEqual(question.groupBy, [
    {kind:'attribute',key:'http.route'},
    {kind:'attribute',key:'http.request.method'},
  ]);
  return {...structuredClone(question),groupBy:[{kind:'attribute',key:'http.route'}]};
}
module.exports = {routeTemplateQuestion};

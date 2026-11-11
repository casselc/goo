'use strict';
const Base=require('./summary400');
function fixture(){const f=Base.fixture();f.environment='summary-template404-'+f.token;for(const r of f.records)r.environment=f.environment;for(const signal of ['logs','spans'])for(const team of ['source','destination'])f.questions[signal][team].environmentSelector.scope.value=f.environment;return f;}
function question(f,kind,team){const q=Base.question(f,kind==='duration'?'duration':'numeric-logs',team);delete q.numeric;if(team==='destination')q.limit=5;return q;}
function composition(q,fields=['count']){return{version:'goo.composition/1',kind:{name:'stack',version:1},children:[{summary:{version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'measure-summary',version:1},display:{fields}}},{records:{version:'goo.view/1',scope:{kind:'record-query',signal:q.signal},kind:{name:'record-table',version:1},display:{columns:['utc','service','summary']}}}]};}
module.exports={...Base,fixture,question,composition};

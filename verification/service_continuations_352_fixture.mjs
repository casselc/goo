import fs from 'node:fs';
import {widerQuestion, edgeQuestion} from '../ui/distributed-services.mjs';
import {model, scopeFields, scopeQuery} from '../ui/activity-views.mjs';
import * as Environment from '../ui/record-environment.mjs';

const {scope, edge, qualified, view, resources, foreignScope} = JSON.parse(fs.readFileSync(0, 'utf8'));
const before = JSON.stringify({scope, edge, view});
const wider = widerQuestion(scope, edge, qualified);
let exact, exactRefused = false, exactReason = '';
try { exact = edgeQuestion(scope, edge); }
catch (error) { exactRefused = true; exactReason = error.message; }
const applied = scopeFields(scope), filter = scopeQuery(scope), saved = model(view);
const ownPolicy = Environment.sameQuestionEnvironment(saved.selector, applied);
const foreignPolicy = Environment.sameQuestionEnvironment(foreignScope, applied);
const environmentMatches = resources.map(resource => Environment.matches(Environment.selection(wider), resource));
const constructed = JSON.stringify({wider, exact, applied, saved});
const unchanged = before === JSON.stringify({scope, edge, view});

if (scope.environmentSelector) scope.environmentSelector.scope = {mode: 'named', value: 'caller mutation'};
else scope.environment = 'caller mutation';
if (scope.namespace) scope.namespace.value = 'caller mutation';
edge.child = 'caller mutation';
edge.childNamespace = 'caller mutation';
view.selector.where.all[0].value = 'caller mutation';
if (view.selector.environmentSelector) view.selector.environmentSelector.scope = {mode: 'named', value: 'caller mutation'};
else view.selector.environment = 'caller mutation';

process.stdout.write(JSON.stringify({
  wider, exact, exactRefused, exactReason, applied, filter, saved,
  ownPolicy, foreignPolicy, environmentMatches, unchanged,
  detached: constructed === JSON.stringify({wider, exact, applied, saved}),
}));

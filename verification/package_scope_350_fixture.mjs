import fs from 'node:fs';
import {packageQuestion, packagePlan, sourceLabel} from '../ui/telemetry-package-intent.mjs';

const {plan, relativePlan, original} = JSON.parse(fs.readFileSync(0, 'utf8'));
const options = {question: original, revision: plan.revision};
const constructed = packageQuestion(plan.definition, original);
const checked = packagePlan(plan, options);
const accepted = JSON.stringify(checked.question) === JSON.stringify(packagePlan(plan, options).question);
// A saved definition is already known independently of the returned plan. Member
// order is irrelevant; the whole definition, not just the claimed revision, binds
// a relative plan to that selection.
const knownDefinition = Object.fromEntries(Object.entries(plan.definition).reverse());
const relativeOptions = {definition: knownDefinition, revision: plan.revision, relative: true};
const relative = packagePlan(relativePlan, relativeOptions);
const negatives = [];
function refuses(name, mutate, opts = options) {
  const changed = structuredClone(plan);
  mutate(changed);
  try {
    packagePlan(changed, opts);
    negatives.push({name, rejected: false});
  } catch {
    negatives.push({name, rejected: true});
  }
}

refuses('question-policy', p => {
  if (p.question.environmentSelector) {
    const s = p.question.environmentSelector;
    if (s.scope.mode === 'named') s.policy = s.policy === 'current-only' ? 'effective' : 'current-only';
    else s.scope = {mode: 'named', value: 'replacement'};
  } else {
    const value = p.question.environment;
    delete p.question.environment;
    p.question.environmentSelector = {version: 'goo/environment-selector/v1', policy: 'effective', scope: {mode: 'named', value}};
  }
});
refuses('question-predicate', p => { delete p.question.where; });
refuses('question-profile', p => {
  if (p.question.numeric) p.question.numeric.unit = 'changed-unit';
  else if (p.question.signal === 'spans') p.question.spanStatus = !p.question.spanStatus;
  else p.question.bucketMs = p.question.bucketMs ? 0 : 1000;
});
refuses('question-grouping', p => { p.question.groupBy = [{kind: 'attribute', key: 'different-group'}]; });
refuses('question-window', p => { p.question.fromMs += 1; });
refuses('revision', p => { p.revision = p.revision === 'f'.repeat(64) ? 'e'.repeat(64) : 'f'.repeat(64); });
refuses('definition-predicate', p => { delete p.definition.where; });
refuses('source-missing', p => { delete p.definition.source.environment; delete p.definition.source.environmentSelector; });
refuses('source-null', p => { delete p.definition.source.environment; p.definition.source.environmentSelector = null; });
refuses('source-both-including-empty', p => {
  p.definition.source.environment = '';
  p.definition.source.environmentSelector ??= {version: 'goo/environment-selector/v1', policy: 'current-only', scope: {mode: 'named', value: 'prod'}};
});
refuses('metadata-views', p => { p.definition.views = ['unsupported-view']; });
refuses('handoff-fields-not-array', p => { p.handoff.fields = {}; });
refuses('handoff-notes-not-array', p => { p.handoff.notes = {}; });

const changedRelative = structuredClone(relativePlan);
changedRelative.definition.lookbackMinutes = changedRelative.definition.lookbackMinutes === 120 ? 119 : changedRelative.definition.lookbackMinutes + 1;
changedRelative.question.fromMs = changedRelative.question.toMs - changedRelative.definition.lookbackMinutes * 60000;
try {
  packagePlan(changedRelative, relativeOptions);
  negatives.push({name: 'relative-known-definition', rejected: false});
} catch {
  negatives.push({name: 'relative-known-definition', rejected: true});
}

const source = plan.definition.source;
const malformedSources = [
  {...source, unknown: true},
  {...source, environment: null},
  {adapter: source.adapter, signal: source.signal, environmentSelector: {version: 'goo/environment-selector/v1', policy: 'effective', scope: {mode: 'named', value: 'bad\ud800'}}},
];
for (const [i, changed] of malformedSources.entries()) {
  refuses('source-invalid-' + i, p => { p.definition.source = changed; });
}

const constructedBefore = JSON.stringify(constructed), checkedBefore = JSON.stringify(checked), relativeBefore = JSON.stringify(relative);
plan.definition.where.all[0].value = 'mutated caller';
plan.question.where.all[0].value = 'mutated returned input';
if (plan.definition.source.environmentSelector) plan.definition.source.environmentSelector.scope.mode = 'unreported';
else plan.definition.source.environment = 'mutated';
if (plan.definition.groupBy?.length) plan.definition.groupBy[0].key = 'mutated';
if (plan.definition.numeric) plan.definition.numeric.unit = 'mutated';
if (plan.definition.numericHistogram) plan.definition.numericHistogram.binning = 'mutated';
if (plan.definition.reportedEdge) plan.definition.reportedEdge.parent.namespace.mode = 'value';

process.stdout.write(JSON.stringify({
  constructed,
  staged: checked.question,
  relativeStaged: relative.question,
  accepted,
  negatives,
  detached: constructedBefore === JSON.stringify(constructed) && checkedBefore === JSON.stringify(checked) && relativeBefore === JSON.stringify(relative),
  label: sourceLabel(checked.definition.source),
}));

const {openFilters}=require('./support/explore');
const { test, expect } = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');

const app = process.env.GOO_SAMPLE_URL || 'http://127.0.0.1:8081';
const investigation = process.env.GOO_QUERY_URL || 'http://127.0.0.1:8088';
const frames = path.resolve('docs/journeys/screenshots');

async function frame(page, name) {
  await fs.mkdir(frames, { recursive: true });
  await page.screenshot({ path: path.join(frames, name + '.png'), fullPage: true });
}

async function query(request, spec) {
  const response = await request.post(investigation + '/api/query', { data: spec });
  expect(response.ok(), await response.text()).toBeTruthy();
  return response.json();
}

async function render(page, spec) {
  const response = page.waitForResponse(r => r.url().endsWith('/api/query') &&
    r.request().method() === 'POST' && r.request().postDataJSON().view === spec.view &&
    (r.request().postDataJSON().reservationId || '') === (spec.reservationId || ''));
  await page.locator('#expression').fill(JSON.stringify(spec, null, 2));
  await page.getByRole('button', { name: 'Run expression', exact: true }).click();
  const data = await (await response).json();
  await expect(page.locator('#run')).toBeEnabled();
  expect(data.error).toBeUndefined();
  await expect(page.locator('#view')).toHaveValue(spec.view);
  return data;
}

async function create(page, scenario) {
  await page.goto(app);
  await page.locator('#scenario').selectOption(scenario);
  await page.getByRole('button', { name: 'Create reservation', exact: true }).click();
  await expect(page.locator('#result')).toContainText('id');
  const result = JSON.parse(await page.locator('#result').innerText());
  expect(result.id).toBeTruthy();
  expect(result.scenario).toBe(scenario);
  return result.id;
}

test.beforeAll(async ({ request }) => {
  await expect.poll(async () => (await request.get(investigation + '/health')).status(),
    { timeout: 60_000, intervals: [1000] }).toBe(200);
});

test('01 retry success crosses the async trace boundary', async ({ page, request }) => {
  const id = await create(page, 'retry');
  await frame(page, '01a-create-retry');
  await page.goto(investigation+'/#examples');await page.locator('#legacyExamples > summary').click();
  await expect(page.locator('#legacyExamples > .warning')).toContainText('unverified');
  await expect.poll(async () => (await query(request, { view: 'completion', reservationId: id }))
    .rows[0]?.classification, { timeout: 100_000, intervals: [1500] }).toBe('good');
  const completion = await render(page, { view: 'completion', reservationId: id, groupBy: 'release' });
  expect(completion.rows).toHaveLength(1);
  expect(completion.summary.good).toBe(1);
  expect(completion.evidence).toBe('demo-unverified');
  await frame(page, '01b-retry-completion');
  await page.locator('#results button').first().click();
  await expect(page.locator('#view')).toHaveValue('traces');
  await expect(page.locator('#selection')).toContainText(id);
  await expect(page.locator('#results')).toContainText('inventory process');
  await expect(page.locator('#status')).not.toContainText('Native metrics');
  const traces = await query(request, { view: 'traces', reservationId: id });
  expect(traces.rows.some(row => Object.entries(row).some(([k, v]) =>
    /link/i.test(k) && Array.isArray(v) && v.length > 0))).toBeTruthy();
  await frame(page, '01c-linked-consumer-spans');
  const logs = await render(page, { view: 'logs', reservationId: id });
  expect(JSON.stringify(logs.rows)).toContain('reservation.retry');
  expect(JSON.stringify(logs.rows)).toContain('reservation.terminal');
  await frame(page, '01d-retry-log-evidence');
});

test('02 compare good, late, missing and rejected outcomes', async ({ page, request }) => {
  const ids = {};
  for (const scenario of ['good', 'late', 'missing', 'reject']) ids[scenario] = await create(page, scenario);
  await page.goto(investigation+'/#examples');await page.locator('#legacyExamples > summary').click();
  for (const scenario of ['good', 'late', 'missing', 'reject']) {
    const want = scenario === 'good' ? 'good' : 'bad';
    await expect.poll(async () => (await query(request, { view: 'completion', reservationId: ids[scenario] }))
      .rows[0]?.classification, { timeout: 100_000, intervals: [1500] }).toBe(want);
    const data = await render(page, { view: 'completion', reservationId: ids[scenario], groupBy: 'scenario' });
    expect(data.rows).toHaveLength(1);
    expect(data.rows[0].classification).toBe(want);
    await frame(page, '02-' + scenario);
  }
  const latency = await render(page, { view: 'latency', scenario: 'missing' });
  expect(latency.rows.length).toBeGreaterThan(0);
  await frame(page, '02-synchronous-latency');
});

test('03 inspect the definition, reject invalid queries, and distinguish metric points', async ({ page }) => {
  await page.goto(investigation+'/#definitions');
  await page.getByText('Application-owned package', { exact: true }).click();
  await expect(page.locator('#package')).toContainText('reservations');
  await frame(page, '03a-owned-definition');
  await page.getByRole('button',{name:'Examples',exact:true}).click();await page.locator('#legacyExamples > summary').click();
  await page.locator('#expression').fill(JSON.stringify({ view: 'arbitrary-sql', sql: 'DROP TABLE otel_logs' }));
  await page.getByRole('button', { name: 'Run expression', exact: true }).click();
  await expect(page.locator('#status')).toContainText(/unsupported|unknown|view/i);
  await frame(page, '03b-rejected-expression');
  const metrics = await render(page, { view: 'metrics' });
  expect(metrics.rows.length).toBeGreaterThan(0);
  await expect(page.locator('#status')).toContainText('not derived rates');
  await page.getByText('Compiled ClickHouse SQL and parameters', { exact: true }).click();
  await expect(page.locator('#explain')).toContainText('SELECT');
  await frame(page, '03c-metric-points-and-explain');
});

test('04 heatmap cell opens raw spans and the linked reservation waterfall', async ({ page, request }) => {
  const errors=[]; page.on('pageerror', e=>errors.push(e.message));
  const id=await create(page,'retry');
  await expect.poll(async()=> {
    const data=await query(request,{view:'traces',reservationId:id});
    return data.rows.some(r=>r.SpanName==='inventory process');
  },{timeout:100_000,intervals:[1500]}).toBeTruthy();
  await page.goto(investigation+'/#examples');await page.locator('#legacyExamples > summary').click();
  const heat=await render(page,{view:'heatmap',reservationId:id,minutes:5});
  expect(heat.possiblyTruncated).toBe(false);
  const total=heat.rows.reduce((n,r)=>n+Number(r.calls),0);
  expect(total).toBeGreaterThan(0);
  await expect(page.locator('#chart')).toContainText(`${total} gRPC server spans`);
  await frame(page,'04a-duration-heatmap');
  const cell=page.locator('.heat-cell:enabled').first();
  const count=Number(await cell.getAttribute('data-count'));
  const response=page.waitForResponse(r=>r.url().endsWith('/api/query')&&r.request().postDataJSON().spanFilter);
  await cell.click();
  const selected=await (await response).json();
  expect(selected.rows).toHaveLength(count);
  await expect(page.locator('#chartTitle')).toHaveText('Trace waterfall');
  await expect(page.locator('#selection')).toContainText('Heatmap selection');
  await expect(page.locator('.waterfall-row')).toHaveCount(count);
  for(const row of selected.rows) {
    expect(row.ServiceName).toBe('lab-grpc');
    const selection=selected.expression.spanFilter;
    expect(BigInt(row.timestamp_ns)>=BigInt(selection.fromMs)*1000000n).toBeTruthy();
    expect(BigInt(row.timestamp_ns)<BigInt(selection.toMs)*1000000n).toBeTruthy();
  }
  await page.locator('.waterfall-row').first().click();
  await expect(page.locator('#spanDetails')).toContainText(id);
  await frame(page,'04b-cell-span-details');
  const full=page.waitForResponse(r=>r.url().endsWith('/api/query')&&r.request().postDataJSON().reservationId===id&&!r.request().postDataJSON().spanFilter);
  await page.getByRole('button',{name:'Open reservation waterfall',exact:true}).click();
  const all=await (await full).json();
  await expect(page.locator('.waterfall-row')).toHaveCount(all.rows.length);
  const consumer=all.rows.find(r=>r.SpanName==='inventory process'&&r.linked_traces.length);
  expect(consumer).toBeTruthy();
  await page.locator(`.waterfall-row[data-span-id="${consumer.SpanId}"]`).click();
  await expect(page.getByRole('button',{name:/Follow async link/})).toBeVisible();
  await frame(page,'04c-reservation-waterfall');
  await page.getByRole('button',{name:/Follow async link/}).click();
  await expect(page.locator('#spanDetails')).toContainText(consumer.linked_spans[0]);
  await frame(page,'04d-follow-async-link');
  expect(errors).toEqual([]);
});

test('05 guided questions preserve a subject and expose USE measurement gaps',async({page,request})=>{
  const id=await create(page,'retry');
  await expect.poll(async()=>(await query(request,{view:'traces',reservationId:id})).rows.length,
    {timeout:100_000,intervals:[1500]}).toBeGreaterThan(0);
  await page.goto(investigation+'/#examples');await page.locator('#legacyExamples > summary').click();
  await render(page,{view:'traces',reservationId:id,scenario:'retry'});
  await page.locator('#workflow').selectOption('retry');
  await expect(page.locator('#workflowCoverage')).toContainText('different denominators');
  await page.getByRole('button',{name:'3. Confirm retry and terminal evidence',exact:true}).click();
  await expect(page.locator('#view')).toHaveValue('logs');
  await expect(page.locator('#selection')).toContainText(id);
  await expect(page.locator('#results')).toContainText('reservation.retry');
  await frame(page,'05a-guided-retry-evidence');
  await page.locator('#workflow').selectOption('use');
  await expect(page.locator('#workflowCoverage')).toContainText('Available: Redis group lag/pending');
  await expect(page.locator('#workflowCoverage')).toContainText('database pool utilization remain NOT INSTRUMENTED');
  await expect(page.locator('#workflowCoverage')).toContainText('unknown, not zero usage');
  await page.getByRole('button',{name:'1. Inspect queue and worker evidence',exact:true}).click();
  await expect(page.locator('#view')).toHaveValue('queue');
  await expect(page.locator('#selection')).toBeEmpty();
  await expect(page.locator('#chart')).toContainText('Worker utilization');
  await frame(page,'05b-use-coverage-gaps');
  await page.locator('#workflow').selectOption('red');
  await expect(page.locator('#workflowCoverage')).toContainText('Missing: a complete service error-rate measure');
  await page.getByRole('button',{name:'1. Explore volume and duration',exact:true}).click();
  await expect(page.locator('#chartTitle')).toHaveText('gRPC duration heatmap');
  await frame(page,'05c-red-volume-and-duration');
});

test('06 real queue pressure and a discoverable semantic package library',async({page,request})=>{
  await create(page,'late');
  await expect.poll(async()=>{
    const data=await query(request,{view:'queue',minutes:5});
    return data.rows.some(r=>r.metric==='lab.worker.active'&&Number(r.value)>0);
  },{timeout:15_000,intervals:[500]}).toBeTruthy();
  await page.goto(investigation+'/#examples');await page.locator('#legacyExamples > summary').click();
  const data=await render(page,{view:'queue',minutes:5});
  const metrics=new Set(data.rows.map(r=>r.metric));
  for(const metric of ['lab.queue.depth','lab.queue.pending','lab.worker.active','lab.worker.capacity','lab.worker.utilization','lab.queue.wait.p95'])expect(metrics.has(metric)).toBeTruthy();
  for(const row of data.rows.filter(r=>r.metric==='lab.worker.utilization')) {
    expect(Number(row.value)).toBeGreaterThanOrEqual(0);expect(Number(row.value)).toBeLessThanOrEqual(1);
    expect(row.instance).toBeTruthy();
  }
  await expect(page.locator('#chart svg')).toHaveCount(4);
  await expect(page.locator('#chart')).toContainText('no averaging across instances');
  await frame(page,'06a-live-queue-and-workers');
  await page.getByRole('button',{name:'Definitions',exact:true}).click();
  await page.getByText('OTel prebuilt library · capability proposals',{exact:true}).click();
  await page.getByText('otel.database.pool · semconv-v1.37.0-development',{exact:true}).click();
  await expect(page.locator('#prebuilts')).toContainText('db.client.connection.wait_time');
  await expect(page.locator('#prebuilts')).toContainText('proposed');
  await frame(page,'06b-semantic-package-library');
});

test('07 live HTTP package renders the shared Go descriptor',async({page,request})=>{
  await create(page,'good');
  let previous;await expect.poll(async()=>{const r=await request.get(investigation+'/api/http-package?minutes=5');if(!r.ok())return false;previous=await r.json();return true},{timeout:45000,intervals:[1000]}).toBe(true);
  const priorFailures=new Set(previous.prepared.facts.filter(f=>f.route==='/demo-failure').map(f=>f.trace_id));
  const failure=await request.get(app+'/demo-failure');expect(failure.status()).toBe(503);
  await expect.poll(async()=>{
    const r=await request.get(investigation+'/api/http-package?minutes=5');
    if(!r.ok())return false;
    const data=await r.json();
    return data.prepared.facts.some(f=>f.route==='/demo-failure'&&!priorFailures.has(f.trace_id))&&data.results.some(row=>row.dimensions.route==='/demo-failure'&&row.measures.failures.count>=1)&&
      data.results.some(row=>row.dimensions.route==='/reserve'&&row.measures.successful.count>=1);
  },{timeout:30_000,intervals:[1000]}).toBeTruthy();
  await page.goto(investigation+'/#definitions');
  await page.locator('#minutes').fill('5');
  const response=page.waitForResponse(r=>r.url().includes('/api/http-package?'));
  await page.getByRole('button',{name:'Evaluate HTTP package',exact:true}).click();
  const data=await(await response).json();
  await expect(page.locator('#runHttpPackage')).toBeEnabled();
  const descriptor=data.expansion.views[0];
  await expect(page.locator('#httpPackageView th')).toHaveText([...descriptor.groupBy,...descriptor.measures]);
  await expect(page.locator('#httpPackageView')).toContainText('/demo-failure');
  await expect(page.locator('#httpPackageView')).toContainText('/reserve');
  await expect(page.locator('#httpPackageStatus')).toContainText('sampling and ingestion completeness unverified');
  await frame(page,'07a-live-http-package-view');
  await page.getByText('HTTP package source, revision and diagnostics',{exact:true}).click();
  await expect(page.locator('#httpPackageExplain')).toContainText(data.expansion.revision);
  await expect(page.locator('#httpPackageExplain')).toContainText('goo.http-sample');
  await expect(page.locator('#httpPackageExplain')).toContainText('failureStatusRanges');
  await frame(page,'07b-http-source-and-policy');
});

test('08 woven Go operations reach the live investigation overview', async ({page,request})=>{
 const operationApp=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090';
 await page.goto(operationApp);
 for(const [name,status] of [['Successful checkout',200],['Fail checkout',503],['Panic checkout',500]]){
  await page.getByRole('button',{name,exact:true}).click();
  await expect(page.locator('#result')).toContainText('HTTP '+status);
 }
 await frame(page,'08a-woven-operation-scenarios');
 await expect.poll(async()=>{
  const r=await request.get(investigation+'/api/operations?minutes=5');if(!r.ok())return false;
  const data=await r.json();const rows=data.conformance.operations;
  return rows.some(r=>r.name==='checkout.reserve'&&r.calls>=2&&r.errors>=1)&&rows.some(r=>r.name==='checkout.panic'&&r.panics>=1);
 },{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');
 await page.locator('#minutes').fill('5');
 const response=page.waitForResponse(r=>r.url().includes('/api/operations?'));
 await page.getByRole('button',{name:'Inspect operations',exact:true}).click();
 const data=await (await response).json();
 await expect(page.locator('#runOperations')).toBeEnabled();
 await expect(page.locator('#operationView')).toContainText('checkout.reserve');
 await expect(page.locator('#operationView')).toContainText('checkout.panic');
 await expect(page.locator('#operationView')).toContainText('Observed · compatible');
 await expect(page.locator('#operationStatus')).toContainText('completeness are unknown');
 await page.locator('#operations').scrollIntoViewIfNeeded();
 await frame(page,'08b-live-operation-overview');
 await page.getByText('Operation binding and evidence basis',{exact:true}).click();
 await expect(page.locator('#operationExplain')).toContainText(data.conformance.manifestRevision);
 await expect(page.locator('#operationExplain')).toContainText('woven-demo');
 await expect(page.locator('#operationExplain')).toContainText('LIMIT 1001');
 await frame(page,'08c-operation-evidence-binding');
});

test('09 operation error heatmap retains its evidence basis through trace drilldown',async({page,request})=>{
 const operationApp=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090';
 expect((await request.post(operationApp+'/invoke?scenario=error')).status()).toBe(503);
 let live;
 await expect.poll(async()=>{
  const r=await request.get(investigation+'/api/operations?minutes=5');if(!r.ok())return false;
  live=await r.json();return live.spans.some(s=>s.operation==='checkout.reserve'&&s.outcome==='error'&&live.spans.some(child=>child.traceID===s.traceID&&child.parentSpanID===s.spanID));
 },{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');await page.locator('#minutes').fill('5');
 let operationRequests=0;page.on('request',r=>{if(r.url().includes('/api/operations?'))operationRequests++});
 const response=page.waitForResponse(r=>r.url().includes('/api/operations?'));
 await page.getByRole('button',{name:'Inspect operations',exact:true}).click();
 const data=await(await response).json();await expect(page.locator('#runOperations')).toBeEnabled();
 const target=data.spans.find(s=>s.operation==='checkout.reserve'&&s.outcome==='error'&&data.spans.some(child=>child.traceID===s.traceID&&child.parentSpanID===s.spanID));expect(target).toBeTruthy();
 await page.getByRole('button',{name:'Inspect errors for checkout.reserve',exact:true}).click();
 await expect(page.locator('#operationSelection')).toContainText('checkout.reserve · server · error');
 await expect(page.locator('#operationMatches')).toContainText('error');
 await frame(page,'09a-operation-error-heatmap');
 // Controls can change, but exploration remains on the response already captured.
 await page.locator('#minutes').fill('1');
 const cells=page.locator('#operationHeatmap button:enabled');let chosen;
 for(let i=0;i<await cells.count();i++){
  const filter=await cells.nth(i).evaluate(n=>({...n.dataset}));
  if(BigInt(target.startNS)>=BigInt(filter.fromNS)&&BigInt(target.startNS)<BigInt(filter.toNS)&&BigInt(target.durationNS)>=BigInt(filter.minNS)&&(filter.maxNS===undefined||BigInt(target.durationNS)<BigInt(filter.maxNS))){chosen={button:cells.nth(i),filter};break;}
 }
 expect(chosen).toBeTruthy();await chosen.button.click();
 const f=chosen.filter;const expected=data.spans.filter(s=>s.operation==='checkout.reserve'&&s.role==='server'&&s.outcome==='error'&&BigInt(s.startNS)>=BigInt(f.fromNS)&&BigInt(s.startNS)<BigInt(f.toNS)&&BigInt(s.durationNS)>=BigInt(f.minNS)&&(f.maxNS===undefined||BigInt(s.durationNS)<BigInt(f.maxNS)));
 expect(expected.length).toBe(Number(f.count));
 await expect(page.locator('#operationMatches button')).toHaveCount(expected.length);
 await expect(page.locator('#operationExplorer')).toHaveAttribute('data-basis',data.conformance.basis);
 await expect(page.locator('#operationEvidenceStatus')).toContainText(new Date(data.fromMs).toISOString());
 await frame(page,'09b-duration-cell-contributors');
 await page.getByRole('button',{name:'Open trace '+target.traceID,exact:true}).first().click();
 const trace=data.spans.filter(s=>s.traceID===target.traceID);
 await expect(page.locator('#operationWaterfall .waterfall-row')).toHaveCount(trace.length);
 await expect(page.locator('#operationWaterfall')).toContainText('checkout.stock');
 const parent=page.locator('#operationWaterfall [data-span-id="'+target.spanID+'"]');await parent.click();
 await expect(page.locator('#operationSpanDetails')).toContainText(target.traceID);
 await expect(page.locator('#operationSpanDetails')).toContainText('"operationOutcome": "error"');
 expect(operationRequests).toBe(1);
 await frame(page,'09c-operation-trace-waterfall');
});

test('10 scoped subject events connect related traces without merging tenants',async({page,request})=>{
 const operationApp=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090';
 const id='subject-'+Date.now();
 await page.goto(operationApp);await page.locator('#subject').fill(id);
 for(const tenant of ['demo','demo','other']){
  await page.locator('#tenant').selectOption(tenant);
  await page.getByRole('button',{name:'Successful checkout',exact:true}).click();
  await expect(page.locator('#result')).toContainText('HTTP 200');
  await expect(page.locator('#result')).toContainText('"tenant":"'+tenant+'"');
  // Ensure the preceding response is complete before starting the next call.
  await expect(page.getByRole('button',{name:'Successful checkout',exact:true})).toBeEnabled();
 }
 await frame(page,'10a-same-subject-separate-tenants');
 await expect.poll(async()=>{
  const r=await request.get(investigation+'/api/operations?minutes=5');if(!r.ok())return false;
  const d=await r.json();return d.spans.filter(s=>(s.events||[]).some(e=>e.subject?.id===id&&e.subject.tenant==='demo')).length===2&&d.spans.filter(s=>(s.events||[]).some(e=>e.subject?.id===id&&e.subject.tenant==='other')).length===1;
 },{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');await page.locator('#minutes').fill('5');
 let operationRequests=0;page.on('request',r=>{if(r.url().includes('/api/operations?'))operationRequests++});
 const response=page.waitForResponse(r=>r.url().includes('/api/operations?'));
 await page.getByRole('button',{name:'Inspect operations',exact:true}).click();
 const data=await(await response).json();await expect(page.locator('#runOperations')).toBeEnabled();
 const related=data.spans.filter(s=>(s.events||[]).some(e=>e.subject?.id===id&&e.subject.tenant==='demo'));
 const other=data.spans.find(s=>(s.events||[]).some(e=>e.subject?.id===id&&e.subject.tenant==='other'));
 expect(related).toHaveLength(2);expect(other).toBeTruthy();expect(related[0].traceID).not.toBe(related[1].traceID);
 await page.getByRole('button',{name:'Inspect calls for checkout.reserve',exact:true}).click();
 await page.getByRole('button',{name:'Open trace '+related[0].traceID,exact:true}).first().click();
 await expect(page.locator('#operationSubjects')).toContainText('reservation.accepted');
 await expect(page.locator('#operationSubjects')).toContainText('reservation.terminal');
 await frame(page,'10b-trace-subject-events');
 await page.getByRole('button',{name:'Follow subject demo/reservation/'+id,exact:true}).first().click();
 await expect(page.locator('#operationSubjectStatus')).toContainText('4 observed events · 2 event-bearing invocations · 2 traces');
 await expect(page.locator('#operationSubjectEvents tr')).toHaveCount(5);
 await expect(page.locator('#operationSubjectEvents')).not.toContainText(other.traceID);
 await expect(page.locator('#operationSubjectEvents')).toContainText(related[0].traceID);
 await expect(page.locator('#operationSubjectEvents')).toContainText(related[1].traceID);
 await expect(page.locator('#operationSubjectStatus')).toContainText('not a parent edge or proof of retry, commit or durable completion');
 await expect(page.locator('#operationExplorer')).toHaveAttribute('data-basis',data.conformance.basis);
 await frame(page,'10c-scoped-subject-related-traces');
 await page.getByRole('button',{name:'Open subject trace '+related[1].traceID,exact:true}).first().click();
 await expect(page.locator('#operationWaterfall')).toContainText(related[1].traceID);
 expect(operationRequests).toBe(1);
});

test('11 a second application loads the same operation views from YAML',async({page,request})=>{
 const reportApp=process.env.GOO_REPORT_URL||'http://127.0.0.1:8091';
 const catalogResponse=await request.get(investigation+'/api/operation-packages');expect(catalogResponse.ok()).toBeTruthy();
 const catalog=await catalogResponse.json();const reports=catalog.find(p=>p.definition.name==='reports');expect(reports).toBeTruthy();
 expect(reports.definition.operations).toEqual([{name:'report.generate',role:'server'}]);
 expect(reports.definition.views).not.toContain('duration-heatmap');
 expect((await request.get(investigation+'/api/operations?package=unknown')).status()).toBe(400);
 await page.goto(reportApp);await page.getByRole('button',{name:'Generate report',exact:true}).click();
 await expect(page.locator('#result')).toContainText('ready');const report=JSON.parse(await page.locator('#result').innerText());expect(report.id).toBeTruthy();
 await frame(page,'11a-second-go-application');
 await expect.poll(async()=>{const r=await request.get(investigation+'/api/operations?package=reports&minutes=5');if(!r.ok())return false;const d=await r.json();return d.spans.some(s=>(s.events||[]).some(e=>e.subject?.id===report.id&&e.subject.namespace==='report'))},{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');await expect(page.locator('#operationPackage option')).toHaveCount(2);
 await page.locator('#operationPackage').selectOption('reports');await page.locator('#minutes').fill('5');
 const response=page.waitForResponse(r=>r.url().includes('/api/operations?')&&r.url().includes('package=reports'));
 await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const data=await(await response).json();await expect(page.locator('#runOperations')).toBeEnabled();
 expect(data.package.revision).toBe(reports.revision);expect(data.source.service).toBe('lab-go-reports');
 await expect(page.locator('#operationPackageTitle')).toHaveText('Report generator operations');
 await expect(page.locator('#operationView')).toContainText('report.generate');await expect(page.locator('#operationView')).not.toContainText('checkout.reserve');
 await expect(page.locator('#operationHeatmap')).toHaveCount(0);
 await frame(page,'11b-yaml-package-shared-overview');
 const span=data.spans.find(s=>s.events.some(e=>e.subject?.id===report.id));
 await page.getByRole('button',{name:'Open trace '+span.traceID,exact:true}).first().click();
 await page.getByRole('button',{name:'Follow subject demo/report/'+report.id,exact:true}).click();
 await expect(page.locator('#operationSubjectStatus')).toContainText('1 observed event · 1 event-bearing invocation · 1 trace');
 await page.getByText('Operation binding and evidence basis',{exact:true}).click();await expect(page.locator('#operationExplain')).toContainText(reports.revision);
 await expect(page.locator('#operationExplain')).toContainText('Sampling and completeness remain unverified');
 await frame(page,'11c-package-contract-and-subject');
 await page.locator('#operationWaterfall').getByRole('button',{name:'Expand trace across services and find linked logs',exact:true}).click();await expect(page.locator('#workspaceTitle')).toHaveText('Explore');await expect(page.locator('#exploreTrace')).toContainText('report.generate');await page.locator('#exploreTrace .waterfall-row').first().click();await expect(page.locator('#exploreSpanDetails')).toContainText(report.id);await expect(page.locator('#exploreSpanDetails')).not.toContainText('reservation');await page.locator('#exploreContextClose').click();await page.getByRole('button',{name:'Operations',exact:true}).click();
 await page.locator('#operationPackage').selectOption('checkout');await expect(page.locator('#operationExplorer')).toBeEmpty();await expect(page.locator('#operationView')).toBeEmpty();
});

test('12 switching packages discards a late response from the previous scope',async({page})=>{
 await page.goto(investigation+'/#operations');await expect(page.locator('#operationPackage option')).toHaveCount(2);
 await page.locator('#operationPackage').selectOption('checkout');await page.locator('#minutes').fill('5');
 let markCaptured,release;
 const captured=new Promise(resolve=>{markCaptured=resolve});const gate=new Promise(resolve=>{release=resolve});
 await page.route('**/api/operations?**',async route=>{
  // Preserve the real upstream response; only its delivery is delayed.
  const response=await route.fetch();
  if(route.request().url().includes('package=checkout')){markCaptured();await gate;}
  await route.fulfill({response});
 });
 await page.getByRole('button',{name:'Inspect operations',exact:true}).click();await captured;
 await page.locator('#operationPackage').selectOption('reports');
 const oldResponse=page.waitForResponse(r=>r.url().includes('/api/operations?')&&r.url().includes('package=checkout'));release();await oldResponse;
 await expect(page.locator('#runOperations')).toBeEnabled();await expect(page.locator('#operationView')).toBeEmpty();await expect(page.locator('#operationExplorer')).toBeEmpty();
 await expect(page.locator('#operationPackage')).toHaveValue('reports');
 await frame(page,'12a-late-package-response-discarded');
 const next=page.waitForResponse(r=>r.url().includes('/api/operations?')&&r.url().includes('package=reports'));
 await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const result=await(await next).json();expect(result.package.definition.name).toBe('reports');
 await expect(page.locator('#operationPackageTitle')).toHaveText('Report generator operations');await expect(page.locator('#operationView')).not.toContainText('checkout.reserve');
 await expect(page.locator('#runOperations')).toBeEnabled();
 await frame(page,'12b-correct-package-evidence');
});

test('13 native counter rates retain unknown boundaries and link only contextual spans',async({page,request})=>{
 // Establish a collected baseline, then make another real invocation to change it.
 expect((await request.post((process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090')+'/invoke?scenario=ok')).status()).toBe(200);
 let scope;
 await expect.poll(async()=>{
  const r=await request.get(investigation+'/api/operations?minutes=2');if(!r.ok())return false;scope=await r.json();
  const rates=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!rates.ok())return false;const data=await rates.json();return data.rates.some(rate=>rate.result.series.some(s=>s.intervals.some(i=>i.state==='ok')));
 },{timeout:60000,intervals:[1000]}).toBe(true);
 expect((await request.post((process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090')+'/invoke?scenario=ok')).status()).toBe(200);
 await expect.poll(async()=>{
  const response=await request.get(investigation+'/api/operations?minutes=2');scope=await response.json();const r=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!r.ok())return false;const d=await r.json();return d.rates.some(rate=>rate.result.series.some(s=>s.intervals.some(i=>i.state==='ok'&&i.perSecond>0)));
 },{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');await expect(page.locator('#operationPackage option')).toHaveCount(2);await page.locator('#operationPackage').selectOption('checkout');await page.locator('#minutes').fill('2');
 const response=page.waitForResponse(r=>r.url().includes('/api/counter-rates?'));
 await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const data=await(await response).json();expect(data.error).toBeUndefined();await expect(page.locator('#runOperations')).toBeEnabled();
 await expect(page.locator('#counterRateView')).toContainText('Native counter rates');await expect(page.locator('#counterRateView')).toContainText('contextual, not exact counter contributors');
 await page.locator('#counterRateView > summary').click();await expect(page.locator('#counterRateView svg').first()).toBeVisible();
 const supported=data.rates.flatMap(r=>r.result.series.flatMap(s=>s.intervals.filter(i=>i.state==='ok')));expect(supported.length).toBeGreaterThan(0);expect(supported.some(i=>i.perSecond>0)).toBe(true);
 for(const rate of data.rates)for(const s of rate.result.series)for(const interval of s.intervals)if(interval.state!=='ok')expect(interval.perSecond).toBeUndefined();
 await frame(page,'13a-native-counter-rate-intervals');
 await page.getByRole('button',{name:'Inspect contextual spans',exact:true}).first().click();await expect(page.locator('#operationSelection')).toContainText('checkout.');
 await page.locator('#counterRateView details').first().locator('summary').click();await expect(page.locator('#counterRateView')).toContainText(data.rates[0].result.basis);
 await frame(page,'13b-counter-basis-and-contextual-spans');
});

test('14 native histogram heatmap preserves bucket bounds and contextual navigation',async({page,request})=>{
 expect((await request.post((process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090')+'/invoke?scenario=ok')).ok()).toBe(true);
 await expect.poll(async()=>{
  const scope=await(await request.get(investigation+'/api/operations?minutes=2')).json();
  const r=await request.get(investigation+'/api/histograms?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!r.ok())return false;
  const d=await r.json();return d.histograms.some(h=>h.result.series.some(s=>s.intervals.some(i=>i.state==='ok'&&BigInt(i.distribution.count)>0n)));
 },{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');await expect(page.locator('#operationPackage option')).toHaveCount(2);await page.locator('#operationPackage').selectOption('checkout');await page.locator('#minutes').fill('2');
 const response=page.waitForResponse(r=>r.url().includes('/api/histograms?'));await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const data=await(await response).json();expect(data.error).toBeUndefined();await expect(page.locator('#runOperations')).toBeEnabled();
 for(const h of data.histograms)for(const s of h.result.series)for(const i of s.intervals){if(i.state!=='ok')expect(i.distribution).toBeUndefined();else expect(i.distribution.bucketCounts.reduce((a,b)=>a+BigInt(b),0n)).toBe(BigInt(i.distribution.count));}
 await page.locator('#histogramView > summary').click();await expect(page.locator('#histogramView table').first()).toBeVisible();
 await page.locator('#histogramView section').filter({has:page.locator('.heat-cell:enabled')}).first().screenshot({path:path.join(frames,'14a-native-histogram-heatmap.png')});
 await page.locator('#histogramView .heat-cell:enabled').first().click();await expect(page.locator('#histogramView .histogramSelection').filter({hasText:'observed bucket population'})).toContainText('no interpolation');
 await expect(page.locator('#histogramView')).toContainText('does not assert bucket membership or exact contributors');
 await page.locator('#histogramView section').filter({has:page.locator('.histogramSelection button')}).screenshot({path:path.join(frames,'14b-histogram-bucket-bounds.png')});
 await page.getByRole('button',{name:'Inspect histogram context',exact:true}).click();await expect(page.locator('#operationSelection')).toContainText('checkout.');
 await page.locator('#operationExplorer').screenshot({path:path.join(frames,'14c-histogram-contextual-spans.png')});
});

test('15 operation error ratios guide contextual error and trace investigation',async({page,request})=>{
 const endpoint=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090';
 const contextual=(i,scope)=>scope.spans.some(s=>s.operation==='checkout.reserve'&&s.role==='server'&&s.outcome==='error'&&BigInt(s.startNS)>=BigInt(i.fromNS)&&BigInt(s.startNS)<BigInt(i.toNS));
 // Establish all explicit zero baselines, then put real calls in a later collection interval.
 expect((await request.post(endpoint+'/invoke?scenario=ok')).status()).toBe(200);
 await expect.poll(async()=>{const scope=await(await request.get(investigation+'/api/operations?minutes=2')).json();const r=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!r.ok())return false;const d=await r.json();return d.ratios?.some(e=>e.result.series.some(s=>s.stream.attributes['goo.operation.name']==='checkout.reserve'&&s.intervals.some(i=>i.state==='ok'||i.state==='no_data')))}, {timeout:60000,intervals:[1000]}).toBe(true);
 expect((await request.post(endpoint+'/invoke?scenario=ok')).status()).toBe(200);expect((await request.post(endpoint+'/invoke?scenario=error')).status()).toBe(503);
 await expect.poll(async()=>{const scope=await(await request.get(investigation+'/api/operations?minutes=2')).json();const r=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!r.ok())return false;const d=await r.json();return d.ratios.some(e=>e.result.series.some(s=>s.stream.attributes['goo.operation.name']==='checkout.reserve'&&s.intervals.some(i=>i.state==='ok'&&BigInt(i.errors)>0n&&contextual(i,scope))))},{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');await expect(page.locator('#operationPackage option')).toHaveCount(2);await page.locator('#operationPackage').selectOption('checkout');await page.locator('#minutes').fill('2');
 const evidence=page.waitForResponse(r=>r.url().includes('/api/operations?'));const response=page.waitForResponse(r=>r.url().includes('/api/counter-rates?'));await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const data=await(await response).json();const scope=await(await evidence).json();await expect(page.locator('#runOperations')).toBeEnabled();expect(data.error).toBeUndefined();
 for(const entry of data.ratios){expect(entry.result.counterBasis).toBe(data.rates.find(r=>r.contract.name===entry.contract.totalRate).result.basis);for(const s of entry.result.series)for(const i of s.intervals){if(i.state==='ok'){expect(BigInt(i.errors)<=BigInt(i.calls)).toBe(true);expect(i.percent).toBeGreaterThanOrEqual(0);expect(i.percent).toBeLessThanOrEqual(100)}else expect(i.percent).toBeUndefined()}}
 await page.locator('#errorRatioView > summary').click();const operation=page.locator('#errorRatioView section').filter({has:page.getByRole('heading',{name:'checkout.reserve / server',exact:true})});await expect(operation).toBeVisible();await expect(operation.locator('svg')).toBeVisible();await operation.screenshot({path:path.join(frames,'15a-operation-error-ratios.png')});
 const positive=data.ratios[0].result.series.find(s=>s.stream.attributes['goo.operation.name']==='checkout.reserve').intervals.filter(i=>i.state==='ok'&&BigInt(i.errors)>0n);const selected=positive.findIndex(i=>contextual(i,scope));expect(selected).toBeGreaterThanOrEqual(0);await operation.getByRole('button',{name:'Why did errors rise?',exact:true}).nth(selected).click();await expect(page.locator('#errorRatioSelection')).toContainText('not an automated increase or cause detector');await page.locator('#errorRatioSelection').screenshot({path:path.join(frames,'15b-guided-error-investigation.png')});
 await page.getByRole('button',{name:'Inspect error context',exact:true}).click();await expect(page.locator('#operationSelection')).toContainText('checkout.reserve · server · error');await expect(page.locator('#operationMatches')).toContainText('error');await page.locator('#operationExplorer').screenshot({path:path.join(frames,'15c-contextual-error-spans.png')});
 await page.locator('#operationMatches').getByRole('button',{name:/^Open trace /}).first().click();await expect(page.locator('#operationWaterfall')).toContainText('checkout.reserve');await page.locator('#operationWaterfall').screenshot({path:path.join(frames,'15d-error-trace-waterfall.png')});
 await page.locator('#operationPackage').selectOption('reports');await expect(page.locator('#errorRatioView')).toBeHidden();await expect(page.locator('#errorRatioView')).toBeEmpty();
});

test('16 declared fleet separates observed members from exact interval coverage',async({page,request})=>{
 for(const url of [process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090',process.env.GOO_PEER_URL||'http://127.0.0.1:8092'])expect((await request.post(url+'/invoke?scenario=ok')).status()).toBe(200);
 await expect.poll(async()=>{const scope=await(await request.get(investigation+'/api/operations?minutes=2')).json();const r=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!r.ok())return false;const d=await r.json();return d.fleet?.result.members.every(m=>m.state==='observed')&&d.fleet.result.series.some(s=>s.intervals.some(i=>i.state==='unknown'))},{timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#operations');await expect(page.locator('#operationPackage option')).toHaveCount(2);await page.locator('#operationPackage').selectOption('checkout');await page.locator('#minutes').fill('2');const response=page.waitForResponse(r=>r.url().includes('/api/counter-rates?'));await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const data=await(await response).json();await expect(page.locator('#runOperations')).toBeEnabled();expect(data.error).toBeUndefined();
 const fleet=data.fleet.result;expect(data.fleet.contract.members).toEqual(['woven-demo','woven-peer']);expect(fleet.members.every(m=>m.state==='observed')).toBe(true);
 for(const s of fleet.series)for(const i of s.intervals)if(i.supported<i.expected){expect(i.state).toBe('unknown');expect(i.calls).toBeUndefined();expect(i.errors).toBeUndefined();expect(i.percent).toBeUndefined()}
 await page.locator('#fleetCoverageView > summary').click();await expect(page.locator('#fleetCoverageView')).toContainText('2/2 declared members observed');await page.locator('#fleetMembers').screenshot({path:path.join(frames,'16a-declared-fleet-members.png')});
 const unsupported=page.locator('#fleetCoverageView .fleet-intervals tr[data-state="unknown"]').first();await expect(unsupported).toContainText('unknown');await unsupported.screenshot({path:path.join(frames,'16b-unsupported-fleet-total.png')});await unsupported.getByRole('button',{name:'Explain fleet coverage',exact:true}).click();await expect(page.locator('#fleetCoverageSelection')).toContainText('Independently timed exports may be unaligned');await page.locator('#fleetCoverageSelection').screenshot({path:path.join(frames,'16c-fleet-coverage-explanation.png')});
 await page.locator('#fleetCoverageView details > summary').click();await expect(page.locator('#fleetCoverageView > details pre')).toContainText(fleet.basis);await page.locator('#operationPackage').selectOption('reports');await expect(page.locator('#fleetCoverageView')).toBeHidden();await expect(page.locator('#fleetCoverageView')).toBeEmpty();
});

async function fleetSnapshot(request){const scope=await(await request.get(investigation+'/api/operations?minutes=2')).json();const response=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!response.ok())throw Error(await response.text());return response.json();}
async function showFleet(page){await page.goto(investigation+'/#operations');await expect(page.locator('#operationPackage option')).toHaveCount(2);await page.locator('#operationPackage').selectOption('checkout');await page.locator('#minutes').fill('2');const response=page.waitForResponse(r=>r.url().includes('/api/counter-rates?'));await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const data=await(await response).json();await expect(page.locator('#runOperations')).toBeEnabled();expect(data.error).toBeUndefined();await page.locator('#fleetCoverageView > summary').click();return data;}
function reserveComparisons(data){return data.fleet.result.comparison.series.find(s=>s.stream.attributes['goo.operation.name']==='checkout.reserve').intervals;}

test('17 live peer imbalance telemetry loss and restart retain honest comparison semantics',async({page,request})=>{
 test.setTimeout(240000);
 const reserveUI=()=>page.locator('#fleetComparison .fleet-comparison-series').filter({has:page.getByRole('heading',{name:'checkout.reserve / server',exact:true})});
 const main=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090',peer=process.env.GOO_PEER_URL||'http://127.0.0.1:8092';
 try{
  expect((await request.post(peer+'/demo/metrics?mode=on')).ok()).toBe(true);
  await expect.poll(async()=>{for(let n=0;n<5;n++)expect((await request.post(main+'/invoke?scenario=ok')).status()).toBe(200);expect((await request.post(peer+'/invoke?scenario=error')).status()).toBe(503);const d=await fleetSnapshot(request);return reserveComparisons(d).some(i=>i.state==='comparable'&&BigInt(i.members.find(m=>m.instance==='woven-demo').calls)>BigInt(i.members.find(m=>m.instance==='woven-peer').calls)&&BigInt(i.members.find(m=>m.instance==='woven-peer').errors)>0n)},{timeout:60000,intervals:[1000]}).toBe(true);
  let data=await showFleet(page);const rows=reserveComparisons(data);const imbalance=rows.find(i=>i.state==='comparable'&&BigInt(i.members.find(m=>m.instance==='woven-demo').calls)>BigInt(i.members.find(m=>m.instance==='woven-peer').calls)&&BigInt(i.members.find(m=>m.instance==='woven-peer').errors)>0n);expect(imbalance).toBeTruthy();expect(BigInt(imbalance.startSpreadNS)<=1000000000n).toBe(true);expect(BigInt(imbalance.endSpreadNS)<=1000000000n).toBe(true);expect(imbalance.percent).toBeUndefined();
  const row=reserveUI().locator('.fleet-comparison-row[data-from-ns="'+imbalance.fromNS+'"]');await row.first().getByRole('button',{name:'Explain member comparison',exact:true}).click();await expect(page.locator('#fleetComparisonSelection pre')).toContainText('\"calls\": \"'+imbalance.members.find(m=>m.instance==='woven-demo').calls+'\"');await row.first().locator('xpath=ancestor::section[1]').screenshot({path:path.join(frames,'17a-actual-member-windows-and-imbalance.png')});
  await page.goto(peer);await page.getByText('Lab telemetry fault injection',{exact:true}).click();await page.getByRole('button',{name:'Pause metric export',exact:true}).click();await expect(page.locator('#metricsStatus')).toContainText('false');const paused=Date.now();await page.getByRole('button',{name:'Successful checkout',exact:true}).click();await expect(page.locator('#result')).toContainText('HTTP 200');await frame(page,'17b-peer-serves-requests-with-metric-loss');
  const duringLoss=i=>BigInt(i.fromNS)>=BigInt(paused)*1000000n&&i.state==='unknown'&&['ok','no_data'].includes(i.members.find(m=>m.instance==='woven-demo').state);
  await expect.poll(async()=>{const d=await fleetSnapshot(request);return Date.now()-paused>=13000&&reserveComparisons(d).some(duringLoss)},{timeout:45000,intervals:[1000]}).toBe(true);
  data=await showFleet(page);const loss=reserveComparisons(data).find(duringLoss);expect(loss).toBeTruthy();const lostPeer=loss.members.find(m=>m.instance==='woven-peer');expect(lostPeer.calls).toBeUndefined();expect(lostPeer.percent).toBeUndefined();await reserveUI().locator('.fleet-comparison-row[data-from-ns="'+loss.fromNS+'"]').first().getByRole('button',{name:'Explain member comparison',exact:true}).click();await page.locator('#fleetComparisonSelection').screenshot({path:path.join(frames,'17c-peer-loss-stays-unknown.png')});
  expect((await request.post(peer+'/demo/metrics?mode=on')).ok()).toBe(true);expect((await request.post(peer+'/invoke?scenario=error')).status()).toBe(503);
  await expect.poll(async()=>{const d=await fleetSnapshot(request);return reserveComparisons(d).some(i=>BigInt(i.fromNS)>=BigInt(paused)*1000000n&&i.members.some(m=>m.instance==='woven-peer'&&m.causes?.includes('excessive-gap')))},{timeout:45000,intervals:[1000]}).toBe(true);
  // Restart only this lab's peer process; volumes and other services are retained.
  const restarted=Date.now();require('node:child_process').execFileSync('docker',['compose','restart','operation-peer'],{cwd:path.resolve('.'),timeout:45000,stdio:'pipe'});
  await expect.poll(async()=>{try{return (await request.get(peer+'/health')).status()}catch{return 0}},{timeout:30000,intervals:[1000]}).toBe(200);
  expect((await request.post(peer+'/invoke?scenario=ok')).status()).toBe(200);
  await expect.poll(async()=>{const d=await fleetSnapshot(request);return reserveComparisons(d).some(i=>BigInt(i.toNS)>BigInt(restarted)*1000000n&&i.members.some(m=>m.instance==='woven-peer'&&m.causes?.includes('counter-epoch-change')))},{timeout:45000,intervals:[1000]}).toBe(true);
  data=await showFleet(page);const reset=reserveComparisons(data).find(i=>BigInt(i.toNS)>BigInt(restarted)*1000000n&&i.members.some(m=>m.instance==='woven-peer'&&m.causes?.includes('counter-epoch-change')));expect(reset.state).toBe('unknown');await reserveUI().locator('.fleet-comparison-row[data-from-ns="'+reset.fromNS+'"]').first().getByRole('button',{name:'Explain member comparison',exact:true}).click();await expect(page.locator('#fleetComparisonSelection pre')).toContainText('counter-epoch-change');await page.locator('#fleetComparisonSelection').screenshot({path:path.join(frames,'17d-restart-epoch-is-not-a-comparison.png')});
 }finally{await request.post(peer+'/demo/metrics?mode=on').catch(()=>{});}
});

test('18 late initial query preserves edited operation lookback',async({page})=>{
 let captured,release;const fetched=new Promise(r=>captured=r),gate=new Promise(r=>release=r);
 await page.route('**/api/query',async route=>{const response=await route.fetch();captured();await gate;await route.fulfill({response})});
 await page.goto(investigation+'/#examples');await page.locator('#legacyExamples > summary').click();await fetched;await page.getByRole('button',{name:'Operations',exact:true}).click();await page.locator('#minutes').fill('2');release();await expect(page.locator('#run')).toBeEnabled();await expect(page.locator('#minutes')).toHaveValue('2');
 await expect(page.locator('#operationPackage option')).toHaveCount(2);const response=page.waitForResponse(r=>r.url().includes('/api/operations?'));await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const result=await(await response).json();expect(result.toMs-result.fromMs).toBe(120000);await expect(page.locator('#runOperations')).toBeEnabled();await expect(page.locator('#minutes')).toHaveValue('2');await frame(page,'18a-edited-lookback-survives-late-response');
});

test('19 saved investigation reopens fixed evidence and refreshes late telemetry',async({page,request})=>{
 const {randomBytes}=require('node:crypto');const {execFileSync}=require('node:child_process');
 const main=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090';
 test.setTimeout(240000);page.setDefaultTimeout(15000);
 // This journey runs alone: establish genuine counter baselines before errors.
 expect((await request.post(main+'/invoke?scenario=ok')).status()).toBe(200);
 await expect.poll(async()=>{const scope=await(await request.get(investigation+'/api/operations?minutes=2')).json();const r=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!r.ok())return false;return (await r.json()).ratios?.some(e=>e.result.series.some(s=>s.stream.attributes['goo.operation.name']==='checkout.reserve'&&s.intervals.some(i=>i.state==='ok'||i.state==='no_data')))}, {timeout:60000,intervals:[1000]}).toBe(true);
 expect((await request.post(main+'/invoke?scenario=error')).status()).toBe(503);
 await expect.poll(async()=>{const r=await request.get(investigation+'/api/operations?minutes=2');if(!r.ok())return false;return (await r.json()).spans.some(s=>s.operation==='checkout.reserve'&&s.outcome==='error')},{timeout:45000,intervals:[1000]}).toBe(true);
 await expect.poll(async()=>{const scope=await(await request.get(investigation+'/api/operations?minutes=2')).json();const r=await request.get(investigation+'/api/counter-rates?package=checkout&fromMs='+scope.fromMs+'&toMs='+scope.toMs);if(!r.ok())return false;return (await r.json()).ratios?.some(e=>e.result.series.some(s=>s.stream.attributes['goo.operation.name']==='checkout.reserve'&&s.intervals.some(i=>i.state==='ok'&&BigInt(i.errors)>0n)))}, {timeout:60000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#investigations');await page.locator('#documentPanel > summary').click();await page.locator('#documentNew').click();await page.locator('#documentTitle').fill('Unpublished ownership note');await page.locator('#documentText').fill('Keep this unsaved note while inspecting operation evidence.');
 await page.goto(investigation+'/#operations');await page.locator('#minutes').fill('2');await expect(page.locator('#operationPackage option[value="checkout"]')).toHaveCount(1);await page.locator('#operationPackage').selectOption('checkout');await page.locator('#runOperations').click();await expect(page.locator('#runOperations')).toBeEnabled();await page.getByRole('button',{name:'Inspect errors for checkout.reserve',exact:true}).click();
 await page.locator('#errorRatioView').evaluate(n=>n.open=true);await page.locator('#errorRatioView section').filter({has:page.getByRole('heading',{name:'checkout.reserve / server',exact:true})}).getByRole('button',{name:'Why did errors rise?',exact:true}).first().click();
 await page.locator('#fleetCoverageView').evaluate(n=>n.open=true);await page.locator('#fleetComparison .fleet-comparison-series').filter({has:page.getByRole('heading',{name:'checkout.reserve / server',exact:true})}).getByRole('button',{name:'Explain member comparison',exact:true}).first().click();
 await page.locator('#operationHeatmap .heat-cell:not([disabled])').first().click();
 const traceButton=page.locator('#operationMatches').getByRole('button',{name:/^Open trace /}).first();const trace=(await traceButton.innerText()).replace('Open trace ','');await traceButton.click();
 const liveSnapshot=await page.locator('#operations').innerHTML();let response=page.waitForResponse(r=>r.url().endsWith('/api/investigations')&&r.request().method()==='POST');await page.locator('#captureSave').click();const first=await(await response).json();expect(first.error).toBeUndefined();expect(first.selection.traceID).toBe(trace);expect(first.selection.filter.fromNS).toBeTruthy();expect(first.selection.filter.minNS).toBeTruthy();expect(first.selection.ratio).toBeTruthy();expect(first.selection.comparison).toBeTruthy();expect(JSON.parse(first.selection.ratio).stream.attributes['goo.operation.name']).toBe('checkout.reserve');expect(JSON.parse(first.selection.comparison).stream.attributes['goo.operation.name']).toBe('checkout.reserve');await expect(page.locator('#saved-operationMode')).toContainText(first.id.slice(0,12));await expect(page.locator('#saved-operationWaterfall')).toContainText(trace);await expect(page.locator('#documentTitle')).toHaveValue('Unpublished ownership note');await expect(page.locator('#documentText')).toHaveValue('Keep this unsaved note while inspecting operation evidence.');await expect(page.locator('#captureSave')).toBeEnabled();await expect(page.locator('#operations')).toBeHidden();expect(await page.locator('#operations').innerHTML()).toBe(liveSnapshot);await page.locator('#savedOperationLive').click();await expect(page.locator('#operations')).toBeVisible();expect(await page.locator('#operations').innerHTML()).toBe(liveSnapshot);await page.getByRole('button',{name:'Saved investigations',exact:true}).click();await expect(page.locator('#savedOperationEvidence')).toBeVisible();await page.locator('#savedInvestigations details').first().evaluate(n=>n.open=true);await page.locator('#savedInvestigations').screenshot({path:path.join(frames,'19a-saved-fixed-evidence-receipt.png')});
 // The real packages process restarts; its named volume retains the immutable receipt.
 execFileSync('docker',['compose','restart','packages'],{cwd:path.resolve('.'),stdio:'pipe'});await expect.poll(async()=>{try{return(await request.get(investigation+'/api/investigations/'+first.id)).status()}catch{return 0}},{timeout:30000}).toBe(200);
 // A late test span is genuinely ingested via OTLP, with event time inside the old question.
 const lateTrace=randomBytes(16).toString('hex'),lateSpan=randomBytes(8).toString('hex');const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));const timestamp=String(BigInt(first.question.toMs-1000)*1000000n);
 const otlp={resourceSpans:[{resource:{attributes:attrs({'service.name':'lab-go-checkout','deployment.environment.name':'local','service.instance.id':'woven-demo'})},scopeSpans:[{scope:{name:'github.com/casselc/goo/instrumentation',version:'0.1.0'},schemaUrl:'goo://instrumentation/v1',spans:[{traceId:lateTrace,spanId:lateSpan,name:'checkout.reserve',kind:2,startTimeUnixNano:timestamp,endTimeUnixNano:String(BigInt(timestamp)+1000000n),attributes:attrs({'goo.operation.name':'checkout.reserve','goo.operation.role':'server','goo.operation.outcome':'error','goo.test.journey':'saved-investigation-late-arrival'})}]}]}]};
 expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:otlp})).ok()).toBe(true);
 await expect.poll(async()=>{const r=await request.get(investigation+'/api/operations?minutes=2');return r.ok()&&(await r.json()).spans.some(s=>s.traceID===lateTrace)},{timeout:45000,intervals:[1000]}).toBe(true);
 let sourceReads=0;await page.route(/\/api\/(operations|counter-rates|histograms)\?/,route=>{sourceReads++;return route.abort()});
 await page.goto(investigation+'/#investigation='+first.id);await page.reload();await expect(page.locator('#saved-operationMode')).toContainText(first.id.slice(0,12));expect(sourceReads).toBe(0);await expect(page.locator('#saved-errorRatioSelection')).toContainText('Investigate observed errors');await expect(page.locator('#saved-fleetComparisonSelection')).toContainText('Actual timestamps and counts are retained');await expect(page.locator('#saved-operationWaterfall')).toContainText(trace);const unchanged=await(await request.get(investigation+'/api/investigations/'+first.id)).json();expect(unchanged).toEqual(first);expect(unchanged.evidence.operations.spans.some(s=>s.traceID===lateTrace)).toBe(false);await page.locator('#saved-operationWaterfall').screenshot({path:path.join(frames,'19b-reopened-trace-context.png')});
 response=page.waitForResponse(r=>r.url().endsWith('/api/investigations')&&r.request().method()==='POST');await page.locator('#captureRefresh').click();const next=await(await response).json();expect(next.error).toBeUndefined();expect(next.parent).toBe(first.id);expect(next.question).toEqual(first.question);expect(next.selection).toEqual(first.selection);expect(next.evidence.operations.spans.some(s=>s.traceID===lateTrace)).toBe(true);expect(next.evidence.operations.conformance.basis).not.toBe(first.evidence.operations.conformance.basis);expect(await(await request.get(investigation+'/api/investigations/'+first.id)).json()).toEqual(first);await page.locator('#captureDiff').evaluate(n=>n.open=true);await expect(page.locator('#captureComparison')).toContainText('"sameQuestion": true');await page.locator('#captureDiff').screenshot({path:path.join(frames,'19c-late-evidence-successor-comparison.png')});
 // A delayed genuine stored response cannot replace a newer package selection.
 let release,seen;const gate=new Promise(r=>release=r),fetched=new Promise(r=>seen=r);
 await page.route('**/api/investigations/'+next.id,async route=>{const upstream=await route.fetch();seen();await gate;await route.fulfill({response:upstream})});
 await page.locator('#captureOpen').click();await fetched;await page.getByRole('button',{name:'Operations',exact:true}).click();await page.locator('#operationPackage').selectOption('reports');release();await expect(page.locator('#captureStatus')).toContainText('saved evidence was not displayed');await expect(page.locator('#operationView')).toBeEmpty();await expect(page.locator('#operationPackage')).toHaveValue('reports');

});

async function explore(request,q){const r=await request.post(investigation+'/api/explore',{data:q});expect(r.ok(),await r.text()).toBe(true);return r.json();}
function column(key){return {kind:'column',key};}function attribute(key){return {kind:'attribute',key};}
async function addExploreFilter(page,field,op,value=''){await openFilters(page);await page.locator('#exploreField').selectOption(JSON.stringify(field));await openFilters(page);await page.locator('#exploreOp').selectOption(op);if(!['exists','missing'].includes(op)){await openFilters(page);await page.locator('#exploreValue').fill(value)};await page.locator('#exploreAdd').click();}
async function exploreClick(page,selector){const promise=page.waitForResponse(r=>r.url().endsWith('/api/explore')&&r.request().method()==='POST');await page.locator(selector).click();const response=await promise;const data=await response.json();expect(response.ok(),JSON.stringify(data)).toBe(true);await expect(page.locator('#exploreRun')).toBeEnabled();return data;}

test('20 general exploration counts full captured populations and preserves missing keys',async({page,request})=>{
 const {randomBytes}=require('node:crypto');const trace=randomBytes(16).toString('hex'),tag=randomBytes(8).toString('hex'),now=Date.now(),attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
 const environment='goo-exploration-journey',key="test.field'] OR 1=1 --\\path\t",special="a\\b\\n\t\n雪";
 const spans=Array.from({length:123},(_,i)=>({traceId:trace,spanId:(BigInt(i)+1n).toString(16).padStart(16,'0'),parentSpanId:i?'0000000000000001':undefined,name:'bounded-exploration-fixture',kind:2,startTimeUnixNano:String(BigInt(now-2000)*1000000n+BigInt(i)*1000000n),endTimeUnixNano:String(BigInt(now-2000)*1000000n+BigInt(i)*1000000n+500000n),attributes:attrs({'goo.test.cohort':tag,'goo.test.fixture':'exploration-population','test.transport':special,...(i<61?{[key]:''}:{})})}));
 const resourceSpans=[0,1].map(n=>({resource:{attributes:attrs({'service.name':n?'fixture-worker':'fixture-edge','deployment.environment.name':environment})},scopeSpans:[{scope:{name:'goo.documentation-fixture',version:'1'},spans:n?spans.slice(90):spans.slice(0,90)}]}));expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans}})).ok()).toBe(true);
 const base={signal:'spans',environment,fromMs:now-60000,toMs:now+1000,limit:100,offset:0,where:{field:attribute('goo.test.cohort'),op:'eq',value:tag}};
 await expect.poll(async()=>{const r=await request.post(investigation+'/api/explore',{data:base});return r.ok()&&(await r.json()).count},{timeout:45000,intervals:[1000]}).toBe('123');
 expect((await explore(request,{...base,where:{all:[base.where,{field:attribute('test.transport'),op:'eq',value:special}]}})).count).toBe('123');
 const grouped=await explore(request,{...base,groupBy:[attribute(key)]});expect(grouped.rows).toHaveLength(100);expect(grouped.hasMore).toBe(true);expect(grouped.groups).toEqual([{g0:'',count:'61'},{g0:null,count:'62'}]);const next=await explore(request,{...base,offset:100,groupBy:[attribute(key)]});expect(next.rows).toHaveLength(23);expect(next.count).toBe('123');expect(next.groups).toEqual(grouped.groups);expect(next.rows.every(r=>!grouped.rows.some(p=>p.SpanId===r.SpanId))).toBe(true);
 const empty=await explore(request,{...base,where:{all:[base.where,{field:attribute(key),op:'eq',value:''}]}});expect(empty.count).toBe('61');const missing=await explore(request,{...base,where:{all:[base.where,{field:attribute(key),op:'missing'}]}});expect(missing.count).toBe('62');const either=await explore(request,{...base,where:{all:[base.where,{any:[{field:attribute(key),op:'eq',value:''},{field:attribute(key),op:'missing'}]}]}});expect(either.count).toBe('123');expect(grouped.plan.rows).not.toContain(key);expect(Object.values(grouped.plan.parameters)).toContain(key);
 await page.goto(investigation+'/#questions');await page.locator('#exploreEnvironment').fill(environment);await exploreClick(page,'#exploreRun');await addExploreFilter(page,attribute('goo.test.cohort'),'eq',tag);await page.locator('#exploreGroup').selectOption(JSON.stringify(attribute(key)));const shown=await exploreClick(page,'#exploreRun');expect(shown.count).toBe('123');await page.getByRole('tab',{name:'Records',exact:true}).click();await expect(page.locator('#exploreStatus')).toContainText('100 displayed');await expect(page.locator('#exploreGroups')).toContainText('Missing key');await expect(page.locator('#exploreGroups')).toContainText('Present empty string');await page.locator('#exploration').screenshot({path:path.join(frames,'20a-full-counts-and-missing-key-groups.png')});await exploreClick(page,'#exploreNext');await expect(page.locator('#exploreStatus')).toContainText('23 displayed');await expect(page.locator('#exploreStatus')).toContainText('123 matching');await page.locator('#exploreRows').screenshot({path:path.join(frames,'20b-bounded-next-page.png')});
 await page.locator('#exploreRows').getByRole('button',{name:'Expand trace '+trace,exact:true}).first().click();await expect(page.locator('#exploreTrace .waterfall-row')).toHaveCount(100);await expect(page.locator('#exploreTraceMore')).toBeEnabled();await page.locator('#exploreTraceMore').click();await expect(page.locator('#exploreTrace .waterfall-row')).toHaveCount(23);await expect(page.locator('#exploreContextStatus')).toContainText('page offset 100');await expect(page.locator('#exploreTrace')).toContainText('parent outside results');await page.locator('#exploreContext').screenshot({path:path.join(frames,'20c-trace-paging-keeps-partial-context-visible.png')});
 // Delay a genuine result and change signal while it is in flight.
 await page.locator('#exploreContextClose').click();
 let release,seen;const gate=new Promise(r=>release=r),fetched=new Promise(r=>seen=r);await page.route('**/api/explore',async route=>{const real=await route.fetch();seen();await gate;await route.fulfill({response:real})});await page.locator('#exploreRun').click();await fetched;await page.locator('#exploreSignal').selectOption('logs');release();await expect(page.locator('#exploreRows')).toBeEmpty();await expect(page.locator('#exploreStatus')).toContainText('Question changed');await page.unroute('**/api/explore');
 const none=await explore(request,{...base,where:{field:attribute('goo.test.cohort'),op:'eq',value:'absent-'+tag}});expect(none.state).toBe('no_data');expect(none.count).toBe('0');
});

test('21 real distributed trace expansion finds exact linked logs without narrowing the trace',async({page,request})=>{
 const reservation=await create(page,'good');const now=Date.now();const q={signal:'spans',environment:'local',fromMs:now-60000,toMs:now+1000,limit:100,offset:0,where:{all:[{field:column('service'),op:'eq',value:'lab-grpc'},{field:attribute('app.reservation.id'),op:'eq',value:reservation}]}};
 let initial;await expect.poll(async()=>{initial=await explore(request,q);return initial.rows.length>0},{timeout:45000,intervals:[1000]}).toBe(true);const trace=initial.rows[0].TraceId;
 await expect.poll(async()=>{const data=await explore(request,{...q,where:{field:column('trace_id'),op:'eq',value:trace}});return new Set(data.rows.map(r=>r.ServiceName)).size>=2},{timeout:45000,intervals:[1000]}).toBe(true);
 await expect.poll(async()=>{const data=await explore(request,{...q,signal:'logs',where:{field:column('trace_id'),op:'eq',value:trace}});return data.rows.length>0},{timeout:45000,intervals:[1000]}).toBe(true);
 await page.goto(investigation+'/#questions');await exploreClick(page,'#exploreRun');await addExploreFilter(page,column('service'),'eq','lab-grpc');await addExploreFilter(page,attribute('app.reservation.id'),'eq',reservation);const search=await exploreClick(page,'#exploreRun');expect(search.rows.every(r=>r.ServiceName==='lab-grpc')).toBe(true);const origin=search.basis;await page.locator('#exploreRows').getByRole('button',{name:'Expand trace '+trace,exact:true}).first().click();await expect(page.locator('#exploreContextStatus')).toContainText('across all services');await expect(page.locator('#exploreTrace')).toContainText('lab-web');await expect(page.locator('#exploreTrace')).toContainText('lab-grpc');await expect(page.locator('#exploreLogStatus')).toContainText('captured logs matching exact trace');await expect(page.locator('#exploreLogs')).toContainText('reservation.accepted');await expect(page.locator('#exploreExplain')).toContainText(origin);let context=JSON.parse(await page.locator('#exploreContextExplain').textContent());expect(context.trace.question.where.all).toEqual([{field:column('trace_id'),op:'eq',value:trace}]);expect(context.trace.question.fromMs).toBe(search.question.fromMs);expect(context.originBasis).toBe(origin);await page.locator('#exploreContext').screenshot({path:path.join(frames,'21a-cross-service-waterfall-and-linked-logs.png')});
 const logOrigin=context.logs.basis;await page.locator('#exploreLogs').getByRole('button',{name:'Expand trace '+trace,exact:true}).first().click();await expect(page.locator('#exploreContextExplain')).toContainText(logOrigin);context=JSON.parse(await page.locator('#exploreContextExplain').textContent());expect(context.originBasis).toBe(logOrigin);expect(context.origin.signal).toBe('logs');
 const spanButton=page.locator('#exploreTrace .waterfall-row').filter({hasText:'lab-web'}).first();const span=await spanButton.getAttribute('data-span-id');await spanButton.click();await expect(page.locator('#exploreLogStatus')).toContainText('and span');const linked=JSON.parse(await page.locator('#exploreContextExplain').textContent());expect(linked.logs.question.where.all).toEqual([{field:column('trace_id'),op:'eq',value:trace},{field:column('span_id'),op:'eq',value:span}]);expect(linked.trace.basis).toBe(context.trace.basis);await page.locator('#exploreLogPanel').screenshot({path:path.join(frames,'21b-exact-span-log-scope.png')});
});

test('22 saved operation evidence expands live context without changing the receipt',async({page,request})=>{
 const main=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090';expect((await request.post(main+'/invoke?scenario=error')).status()).toBe(503);let live;
 await expect.poll(async()=>{const r=await request.get(investigation+'/api/operations?minutes=2&package=checkout');if(!r.ok())return false;live=await r.json();return live.spans.some(s=>s.operation==='checkout.reserve'&&s.outcome==='error')},{timeout:45000,intervals:[1000]}).toBe(true);
 const target=live.spans.filter(s=>s.operation==='checkout.reserve'&&s.outcome==='error').at(-1);const q={package:'checkout',packageRevision:live.package.revision,fromMs:live.fromMs,toMs:live.toMs};const response=await request.post(investigation+'/api/investigations',{data:{question:q,selection:{filter:{operation:'checkout.reserve',role:'server',outcome:'error'},traceID:target.traceID}}});expect(response.ok(),await response.text()).toBe(true);const saved=await response.json();
 const expandedReads=[];page.on('request',r=>{if(r.method()==='POST'&&r.url().endsWith('/api/explore'))expandedReads.push(r.postDataJSON())});await page.goto(investigation+'/#investigation='+saved.id);await expect(page.locator('#captureStatus')).toContainText(saved.id.slice(0,12));await expect(page.locator('#exploreRun')).toHaveCount(0);expect(expandedReads).toHaveLength(0);await page.locator('#saved-operationWaterfall').getByRole('button',{name:'Expand trace across services and find linked logs',exact:true}).click();await expect(page.locator('#exploreContextStatus')).toContainText('Expanded context is a fresh read');await expect(page.locator('#exploreContextStatus')).toContainText('Use Record a finding to capture it separately.');await expect(page.locator('#exploreTrace')).toContainText('checkout.reserve');await expect(page.locator('#exploreContextExplain')).toContainText('"logs"');await expect.poll(()=>expandedReads.length).toBe(2);await expect.poll(async()=>JSON.parse(await page.locator('#exploreContextExplain').textContent()).logs?.question.environment).toBe(saved.evidence.operations.source.environment);await expect(page.locator('#exploreRun')).toHaveCount(1);const expanded=JSON.parse(await page.locator('#exploreContextExplain').textContent());expect(expanded.originBasis).toBe(saved.evidence.operations.conformance.basis);expect(expanded.trace.question.fromMs).toBe(saved.question.fromMs);expect(expanded.trace.question.toMs).toBe(saved.question.toMs);expect(expanded.trace.question.environment).toBe(saved.evidence.operations.source.environment);expect(await(await request.get(investigation+'/api/investigations/'+saved.id)).json()).toEqual(saved);await page.locator('#exploreContext').screenshot({path:path.join(frames,'22a-saved-question-and-independent-live-context.png')});
});

test('23 compare captured distributions in one workspace on desktop and mobile',async({page,request})=>{
 const {randomBytes}=require('node:crypto'),tag=randomBytes(8).toString('hex'),environment='goo-cohort-journey',now=Date.now();
 const attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
 const spans=Array.from({length:12},(_,i)=>({traceId:randomBytes(16).toString('hex'),spanId:randomBytes(8).toString('hex'),name:'index.refresh',kind:2,startTimeUnixNano:String(BigInt(now-2000)*1000000n),endTimeUnixNano:String(BigInt(now-1999)*1000000n),attributes:attrs({'goo.test.cohort':tag,'test.outcome':i<4?'error':'ok',...(i<8?{'test.region':i<3?'east':'west'}:i<10?{'test.region':''}:{})})}));
 expect((await request.post('http://127.0.0.1:4318/v1/traces',{data:{resourceSpans:[{resource:{attributes:attrs({'service.name':'search-indexer','deployment.environment.name':environment})},scopeSpans:[{scope:{name:'goo.documentation-fixture'},spans}]}]}})).ok()).toBe(true);
 const q={signal:'spans',environment,fromMs:now-60000,toMs:now+1000,limit:100,offset:0,where:{field:attribute('goo.test.cohort'),op:'eq',value:tag}};
 await expect.poll(async()=>(await explore(request,q)).count,{timeout:45000,intervals:[1000]}).toBe('12');
 const errors=[];page.on('pageerror',e=>errors.push(e.message));await page.goto(investigation+'/#questions');
 await expect(page.locator('#workspaceTitle')).toHaveText('Explore');await expect(page.locator('#operations')).toBeHidden();await expect(page.locator('#guided')).toBeHidden();
 await page.locator('#exploreEnvironment').fill(environment);await exploreClick(page,'#exploreRun');await addExploreFilter(page,attribute('goo.test.cohort'),'eq',tag);await page.locator('#exploreGroup').selectOption(JSON.stringify(attribute('test.region')));const base=await exploreClick(page,'#exploreRun');expect(base.count).toBe('12');await expect(page.locator('#distributionContent')).toContainText('Missing key');await frame(page,'23a-distribution-workspace');
 let compared=page.waitForResponse(r=>r.url().endsWith('/api/compare'));await page.locator('#baselinePin').click();await compared;
 await addExploreFilter(page,attribute('test.outcome'),'eq','error');compared=page.waitForResponse(r=>r.url().endsWith('/api/compare'));const selected=await exploreClick(page,'#exploreRun');expect(selected.count).toBe('4');expect(selected.question.fromMs).toBe(base.question.fromMs);expect(selected.question.toMs).toBe(base.question.toMs);const comparison=await(await compared).json();expect(comparison.state).toBe('ok');const east=comparison.rows.find(r=>r.value==='east');expect(east).toMatchObject({baselineCount:'3',selectionCount:'3',baselinePercent:'25.00',selectionPercent:'75.00',differencePP:'50.00'});
 await page.getByRole('tab',{name:'Compare',exact:true}).click();await expect(page.locator('#cohortValues')).toContainText('+50.00');await expect(page.locator('#cohortStatus')).toContainText('4 selected records vs 12 baseline');await frame(page,'23b-pinned-baseline-comparison');
 await page.getByRole('button',{name:'Definitions',exact:true}).click();await expect(page.locator('#definitionCatalog .definition-card')).toHaveCount(2);await frame(page,'23c-definition-workspace');await page.getByRole('button',{name:'Explore',exact:true}).click();await expect(page.locator('#cohortValues')).toContainText('+50.00');
 await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await frame(page,'23d-mobile-comparison');
 await page.getByRole('tab',{name:'Compare',exact:true}).focus();await page.keyboard.press('Home');await expect(page.getByRole('tab',{name:'Records',exact:true})).toBeFocused();await page.locator('#exploreRows').getByRole('button',{name:/Expand trace/}).first().click();await expect(page.locator('#exploreTrace .waterfall-row')).toHaveCount(1);await page.screenshot({path:path.join(frames,'23e-mobile-trace-inspector.png')});await page.keyboard.press('Escape');await expect(page.locator('#exploreContext')).toBeHidden();
 await page.getByRole('tab',{name:'Distribution',exact:true}).click();const refinedResponse=page.waitForResponse(r=>r.url().endsWith('/api/explore'));compared=page.waitForResponse(r=>r.url().endsWith('/api/compare'));await page.locator('#distributionContent').getByRole('button',{name:'Filter and run attribute:test.region equals "east"',exact:true}).click();const refined=await(await refinedResponse).json();expect(refined.count).toBe('3');expect(refined.question.fromMs).toBe(base.question.fromMs);expect((await(await compared).json()).rows.find(r=>r.value==='east').selectionPercent).toBe('100.00');
 await addExploreFilter(page,attribute('test.outcome'),'eq','absent');compared=page.waitForResponse(r=>r.url().endsWith('/api/compare'));await exploreClick(page,'#exploreRun');expect((await(await compared).json()).state).toBe('no_data');await page.getByRole('tab',{name:'Compare',exact:true}).click();await expect(page.locator('#cohortValues')).toBeEmpty();await expect(page.locator('#cohortStatus')).toContainText('no normalized comparison');
 await page.locator('#exploreSignal').selectOption('logs');await expect(page.locator('#baselineClear')).toBeHidden();await expect(page.locator('#cohortValues')).toBeEmpty();expect(errors).toEqual([]);
});

test('24 build save reopen and reuse a package over ordinary Python OTel spans',async({page,request})=>{
 const {randomBytes}=require('node:crypto'),{execFileSync}=require('node:child_process');const tag=randomBytes(6).toString('hex'),name='search-'+tag,environment='goo-package-journey';
 const script=`from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource
p=TracerProvider(resource=Resource.create({'service.name':'python-indexer','deployment.environment.name':'${environment}'}))
p.add_span_processor(BatchSpanProcessor(OTLPSpanExporter(endpoint='http://127.0.0.1:4318/v1/traces')))
t=p.get_tracer('acme.search.indexer','1.0')
for i in range(6):
 with t.start_as_current_span('index.refresh') as s:
  s.set_attribute('test.cohort','${tag}')
  s.set_attribute('pipeline.stage','index')
  if i<5: s.set_attribute('region','east' if i<3 else 'west')
p.force_flush()
p.shutdown()
`;
 execFileSync(process.env.GOO_PYTHON||'/workspace/.venvs/goo/bin/python',['-c',script],{stdio:'pipe'});
 const now=Date.now(),q={signal:'spans',environment,fromMs:now-60000,toMs:now+1000,limit:100,offset:0,where:{field:attribute('test.cohort'),op:'eq',value:tag}};let captured;
 await expect.poll(async()=>{captured=await explore(request,q);return captured.count},{timeout:45000,intervals:[1000]}).toBe('6');expect(captured.rows.every(r=>r.ResourceAttributes['telemetry.sdk.language']==='python'&&!r.SpanAttributes['goo.operation.name'])).toBe(true);
 const errors=[];page.on('pageerror',e=>errors.push(e.message));await page.goto(investigation+'/#questions');await page.locator('#exploreEnvironment').fill(environment);await exploreClick(page,'#exploreRun');await addExploreFilter(page,attribute('test.cohort'),'eq',tag);await page.locator('#exploreGroup').selectOption(JSON.stringify(attribute('region')));const original=await exploreClick(page,'#exploreRun');expect(original.count).toBe('6');await page.locator('#exploreCreatePackage').click();await expect(page.locator('#workspaceTitle')).toHaveText('Definitions');await expect(page.locator('#packageName')).toBeFocused();await page.locator('#packageName').fill(name);await page.locator('#packageTitle').fill('Search refreshes');
 let draftRelease,draftSeen;const draftGate=new Promise(r=>draftRelease=r),draftFetched=new Promise(r=>draftSeen=r);await page.route('**/api/telemetry-packages/draft',async route=>{const real=await route.fetch();draftSeen();await draftGate;await route.fulfill({response:real})});await page.locator('#packageTitle').fill('Initial title');const staleDraftReply=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/draft'));await page.locator('#packageBuild').click();await draftFetched;await page.locator('#packageTitle').fill('Search refreshes');draftRelease();await staleDraftReply;await page.unroute('**/api/telemetry-packages/draft');await expect(page.locator('#packageDraftSetup')).toBeVisible();await expect(page.locator('#packageSave')).toBeDisabled();
 let response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/draft'));await page.locator('#packageBuild').click();const plan=await(await response).json();await expect(page.locator('#packageSave')).toBeEnabled();expect(plan.question.fromMs).toBe(original.question.fromMs);expect(plan.yaml).not.toContain('sourceDirectory');expect(plan.yaml).not.toContain('fromMs');expect(plan.yaml).toContain('groupBy:');
 response=page.waitForResponse(r=>r.url().endsWith('/api/explore'));await page.locator('#packagePreview').click();const preview=await(await response).json();expect(preview.count).toBe('6');expect(preview.groups).toEqual(original.groups);await expect(page.locator('#packagePreviewResult')).toContainText('6 matching captured spans');await frame(page,'24a-author-existing-python-telemetry');
 response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages')&&r.request().method()==='POST');await page.locator('#packageSave').click();const saved=await(await response).json();await expect(page.locator('#packageStatus')).toContainText(saved.revision);const old=await(await request.get(investigation+'/api/telemetry-packages/'+saved.revision)).json();
 const download=page.waitForEvent('download');await page.locator('#packageExport').click();const file=await download;expect(file.suggestedFilename()).toBe(name+'.yaml');const exportText=await fs.readFile(await file.path(),'utf8');expect(exportText).toBe(plan.yaml);
 await page.locator('#packageHandoff summary').click();await page.locator('#packageAnnotationName').fill('search.refresh');await page.locator('#packageAnnotationRole').selectOption('consumer');response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/plan'));await page.locator('#packageAnnotation').click();const promotion=await(await response).json();expect(promotion.handoff.goExample).toContain('//goo:operation name:search.refresh role:consumer');await expect(page.locator('#packageHandoffResult')).toContainText('not a patch');const templateDownload=page.waitForEvent('download');await page.getByRole('button',{name:'Download Go template',exact:true}).click();const template=await templateDownload;expect(await fs.readFile(await template.path(),'utf8')).toBe(promotion.handoff.goExample);await frame(page,'24b-optional-instrumentation-handoff');
 execFileSync('docker',['compose','restart','packages'],{cwd:path.resolve('.'),stdio:'pipe'});await expect.poll(async()=>{try{return(await request.get(investigation+'/api/telemetry-packages/'+saved.revision)).status()}catch{return 0}},{timeout:30000}).toBe(200);
 await page.goto(investigation+'/#definitions');await page.getByRole('button',{name:'Explore '+name,exact:true}).click();await expect(page.locator('#workspaceTitle')).toHaveText('Explore');await expect(page.locator('#exploreStatus')).toContainText('6 matching captured spans');await expect(page.locator('#explorePackageScope')).toContainText(saved.revision);await expect(page.locator('#distributionContent')).toContainText('Missing key');await frame(page,'24c-reuse-package-without-source-manifest');
 const fresh=JSON.parse(await page.locator('#exploreExplain').textContent());expect(fresh.question.fromMs).toBeGreaterThanOrEqual(original.question.fromMs);expect(fresh.question.where).toEqual(plan.definition.where);
 await addExploreFilter(page,attribute('region'),'eq','east');const refined=await exploreClick(page,'#exploreRun');expect(refined.count).toBe('3');expect(refined.question.where.all[0]).toEqual(plan.definition.where);expect(await(await request.get(investigation+'/api/telemetry-packages/'+saved.revision)).json()).toEqual(old);
 await page.getByRole('button',{name:'Definitions',exact:true}).click();response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/plan'));await page.locator('#packageFile').setInputFiles({name:name+'.yaml',mimeType:'application/yaml',buffer:Buffer.from(exportText)});await response;await expect(page.locator('#packageSave')).toBeEnabled();await page.locator('#packageYAML').fill(exportText.replace('title: Search refreshes','title: Search refreshes revised'));await expect(page.locator('#packageSave')).toBeDisabled();response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/plan'));await page.locator('#packageValidate').click();await response;response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages')&&r.request().method()==='POST');await page.locator('#packageSave').click();const successor=await(await response).json();expect(successor.revision).not.toBe(saved.revision);expect(await(await request.get(investigation+'/api/telemetry-packages/'+saved.revision)).json()).toEqual(old);await frame(page,'24d-import-and-revise-definition');await page.setViewportSize({width:390,height:844});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);await page.locator('#packageEditor').screenshot({path:path.join(frames,'24e-mobile-package-authoring.png')});
 let release,seen;const gate=new Promise(r=>release=r),fetched=new Promise(r=>seen=r);await page.route('**/api/telemetry-packages/plan',async route=>{const real=await route.fetch();seen();await gate;await route.fulfill({response:real})});await page.locator('#packageValidate').click();await fetched;await page.locator('#packageYAML').fill('apiVersion: invalid');release();await expect(page.locator('#packageSave')).toBeDisabled();await expect(page.locator('#packageDefinitionSummary')).toContainText('Definition changed');await page.unroute('**/api/telemetry-packages/plan');expect(errors).toEqual([]);
});

test('25 import a nested log package and preserve two breakdowns while refining',async({page,request})=>{
 const {randomBytes}=require('node:crypto'),tag=randomBytes(6).toString('hex'),environment='goo-log-package-journey',now=Date.now(),attrs=o=>Object.entries(o).map(([key,stringValue])=>({key,value:{stringValue}}));
 const logRecords=Array.from({length:4},(_,i)=>({timeUnixNano:String(BigInt(now-1000)*1000000n),severityText:i<2?'ERROR':'INFO',severityNumber:i<2?17:9,body:{stringValue:'job.finished '+tag},attributes:attrs({'test.cohort':tag,...(i%2?{'job.queue':'batch'}:{})})}));expect((await request.post('http://127.0.0.1:4318/v1/logs',{data:{resourceLogs:[{resource:{attributes:attrs({'service.name':'node-job-runner','deployment.environment.name':environment,'telemetry.sdk.language':'nodejs'})},scopeLogs:[{scope:{name:'pino-documentation-fixture'},logRecords}]}]}})).ok()).toBe(true);
 const name='logs-'+tag,yaml=`apiVersion: goo/telemetry-package/v1alpha1\nname: ${name}\ntitle: Job logs\nsource: {adapter: clickhouse.otel/v1, signal: logs, environment: ${environment}}\nmeasure: captured-record-count\nlookbackMinutes: 2\nviews: [records, trace-context]\nwhere:\n  all:\n    - field: {kind: attribute, key: test.cohort}\n      op: eq\n      value: ${tag}\n    - any:\n        - field: {kind: column, key: body}\n          op: contains\n          value: job.finished\n        - field: {kind: attribute, key: absent}\n          op: exists\ngroupBy:\n  - {kind: column, key: severity}\n  - {kind: attribute, key: job.queue}\n`;
 const q={signal:'logs',environment,fromMs:now-60000,toMs:now+1000,limit:100,offset:0,where:{field:attribute('test.cohort'),op:'eq',value:tag}};await expect.poll(async()=>(await explore(request,q)).count,{timeout:45000,intervals:[1000]}).toBe('4');
 await page.goto(investigation+'/#definitions');let response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/plan'));await page.locator('#packageFile').setInputFiles({name:name+'.yaml',mimeType:'application/yaml',buffer:Buffer.from(yaml)});const plan=await(await response).json();await expect(page.locator('#packageUse')).toBeEnabled();response=page.waitForResponse(r=>r.url().endsWith('/api/explore'));await page.locator('#packageUse').click();const result=await(await response).json();expect(result.count).toBe('4');expect(result.question.groupBy).toEqual(plan.definition.groupBy);await expect(page.locator('#exploreGroup')).toHaveValue(JSON.stringify(plan.definition.groupBy));await expect(page.locator('#exploreRows')).toContainText('No valid reported trace ID');
 await addExploreFilter(page,column('severity'),'eq','ERROR');const selected=await exploreClick(page,'#exploreRun');expect(selected.count).toBe('2');expect(selected.question.groupBy).toEqual(plan.definition.groupBy);expect(selected.question.where.all[0]).toEqual(plan.definition.where);await page.getByRole('tab',{name:'Records',exact:true}).click();await frame(page,'25a-nested-log-package-and-two-breakdowns');
 // An absent field is retained in an imported package, without claiming records exist.
 await page.getByRole('button',{name:'Definitions',exact:true}).click();await page.locator('#packageYAML').fill(yaml.replace('op: contains\n          value: job.finished','op: contains\n          value: absent-'+tag));response=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/plan'));await page.locator('#packageValidate').click();await response;response=page.waitForResponse(r=>r.url().endsWith('/api/explore'));await page.locator('#packagePreview').click();expect((await(await response).json()).state).toBe('no_data');await expect(page.locator('#packagePreviewResult')).toContainText('0 matching captured logs');
 let release,seen;const gate=new Promise(r=>release=r),fetched=new Promise(r=>seen=r);await page.route('**/api/telemetry-packages/plan',async route=>{const upstream=await route.fetch();seen();await gate;await route.fulfill({response:upstream})});const late=page.waitForResponse(r=>r.url().endsWith('/api/telemetry-packages/plan'));await page.locator('#packageUse').click();await fetched;await page.getByRole('button',{name:'Explore',exact:true}).click();await page.locator('#exploreEnvironment').fill('newer-environment');release();await(await late).finished();await page.evaluate(()=>new Promise(resolve=>requestAnimationFrame(()=>requestAnimationFrame(resolve))));await expect(page.locator('#exploreEnvironment')).toHaveValue('newer-environment');await page.unroute('**/api/telemetry-packages/plan');
});

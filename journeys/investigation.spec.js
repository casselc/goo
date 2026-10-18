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
  await page.goto(investigation);
  await expect(page.locator('.warning')).toContainText('unverified');
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
  await page.goto(investigation);
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
  await page.goto(investigation);
  await page.getByText('Application-owned package', { exact: true }).click();
  await expect(page.locator('#package')).toContainText('reservations');
  await frame(page, '03a-owned-definition');
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
  await page.goto(investigation);
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
  await page.goto(investigation);
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
  await page.goto(investigation);
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
  await page.getByText('OTel prebuilt library · capability proposals',{exact:true}).click();
  await page.getByText('otel.database.pool · semconv-v1.37.0-development',{exact:true}).click();
  await expect(page.locator('#prebuilts')).toContainText('db.client.connection.wait_time');
  await expect(page.locator('#prebuilts')).toContainText('proposed');
  await frame(page,'06b-semantic-package-library');
});

test('07 live HTTP package renders the shared Go descriptor',async({page,request})=>{
  await create(page,'good');
  const failure=await request.get(app+'/demo-failure');expect(failure.status()).toBe(503);
  await expect.poll(async()=>{
    const r=await request.get(investigation+'/api/http-package?minutes=5');
    if(!r.ok())return false;
    const data=await r.json();
    return data.results.some(row=>row.dimensions.route==='/demo-failure'&&row.measures.failures.count>=1)&&
      data.results.some(row=>row.dimensions.route==='/reserve'&&row.measures.successful.count>=1);
  },{timeout:30_000,intervals:[1000]}).toBeTruthy();
  await page.goto(investigation);
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
 const main=process.env.GOO_OPERATION_URL||'http://127.0.0.1:8090',peer=process.env.GOO_PEER_URL||'http://127.0.0.1:8092';
 try{
  expect((await request.post(peer+'/demo/metrics?mode=on')).ok()).toBe(true);
  await expect.poll(async()=>{for(let n=0;n<5;n++)expect((await request.post(main+'/invoke?scenario=ok')).status()).toBe(200);expect((await request.post(peer+'/invoke?scenario=error')).status()).toBe(503);const d=await fleetSnapshot(request);return reserveComparisons(d).some(i=>i.state==='comparable'&&BigInt(i.members.find(m=>m.instance==='woven-demo').calls)>BigInt(i.members.find(m=>m.instance==='woven-peer').calls)&&BigInt(i.members.find(m=>m.instance==='woven-peer').errors)>0n)},{timeout:60000,intervals:[1000]}).toBe(true);
  let data=await showFleet(page);const rows=reserveComparisons(data);const imbalance=rows.find(i=>i.state==='comparable'&&BigInt(i.members.find(m=>m.instance==='woven-demo').calls)>BigInt(i.members.find(m=>m.instance==='woven-peer').calls)&&BigInt(i.members.find(m=>m.instance==='woven-peer').errors)>0n);expect(imbalance).toBeTruthy();expect(BigInt(imbalance.startSpreadNS)<=1000000000n).toBe(true);expect(BigInt(imbalance.endSpreadNS)<=1000000000n).toBe(true);expect(imbalance.percent).toBeUndefined();
  const row=page.locator('#fleetComparison .fleet-comparison-row[data-from-ns="'+imbalance.fromNS+'"]');await row.first().getByRole('button',{name:'Explain member comparison',exact:true}).click();await row.first().locator('xpath=ancestor::section[1]').screenshot({path:path.join(frames,'17a-actual-member-windows-and-imbalance.png')});
  await page.goto(peer);await page.getByText('Lab telemetry fault injection',{exact:true}).click();await page.getByRole('button',{name:'Pause metric export',exact:true}).click();await expect(page.locator('#metricsStatus')).toContainText('false');const paused=Date.now();await page.getByRole('button',{name:'Successful checkout',exact:true}).click();await expect(page.locator('#result')).toContainText('HTTP 200');await frame(page,'17b-peer-serves-requests-with-metric-loss');
  const duringLoss=i=>BigInt(i.fromNS)>=BigInt(paused)*1000000n&&i.state==='unknown'&&['ok','no_data'].includes(i.members.find(m=>m.instance==='woven-demo').state);
  await expect.poll(async()=>{const d=await fleetSnapshot(request);return Date.now()-paused>=13000&&reserveComparisons(d).some(duringLoss)},{timeout:45000,intervals:[1000]}).toBe(true);
  data=await showFleet(page);const loss=reserveComparisons(data).find(duringLoss);expect(loss).toBeTruthy();const lostPeer=loss.members.find(m=>m.instance==='woven-peer');expect(lostPeer.calls).toBeUndefined();expect(lostPeer.percent).toBeUndefined();await page.locator('#fleetComparison .fleet-comparison-row[data-from-ns="'+loss.fromNS+'"]').first().getByRole('button',{name:'Explain member comparison',exact:true}).click();await page.locator('#fleetComparisonSelection').screenshot({path:path.join(frames,'17c-peer-loss-stays-unknown.png')});
  expect((await request.post(peer+'/demo/metrics?mode=on')).ok()).toBe(true);expect((await request.post(peer+'/invoke?scenario=error')).status()).toBe(503);
  await expect.poll(async()=>{const d=await fleetSnapshot(request);return reserveComparisons(d).some(i=>i.members.some(m=>m.instance==='woven-peer'&&m.causes?.includes('excessive-gap')))},{timeout:45000,intervals:[1000]}).toBe(true);
  // Restart only this lab's peer process; volumes and other services are retained.
  require('node:child_process').execFileSync('docker',['compose','restart','operation-peer'],{cwd:path.resolve('.'),timeout:45000,stdio:'pipe'});
  await expect.poll(async()=>{try{return (await request.get(peer+'/health')).status()}catch{return 0}},{timeout:30000,intervals:[1000]}).toBe(200);
  expect((await request.post(peer+'/invoke?scenario=ok')).status()).toBe(200);
  await expect.poll(async()=>{const d=await fleetSnapshot(request);return reserveComparisons(d).some(i=>i.members.some(m=>m.instance==='woven-peer'&&m.causes?.includes('counter-epoch-change')))},{timeout:45000,intervals:[1000]}).toBe(true);
  data=await showFleet(page);const reset=reserveComparisons(data).find(i=>i.members.some(m=>m.instance==='woven-peer'&&m.causes?.includes('counter-epoch-change')));expect(reset.state).toBe('unknown');await page.locator('#fleetComparison .fleet-comparison-row[data-from-ns="'+reset.fromNS+'"]').first().getByRole('button',{name:'Explain member comparison',exact:true}).click();await page.locator('#fleetComparisonSelection').screenshot({path:path.join(frames,'17d-restart-epoch-is-not-a-comparison.png')});
 }finally{await request.post(peer+'/demo/metrics?mode=on').catch(()=>{});}
});

test('18 late initial query preserves edited operation lookback',async({page})=>{
 let captured,release;const fetched=new Promise(r=>captured=r),gate=new Promise(r=>release=r);
 await page.route('**/api/query',async route=>{const response=await route.fetch();captured();await gate;await route.fulfill({response})});
 await page.goto(investigation+'/#operations');await fetched;await page.locator('#minutes').fill('2');release();await expect(page.locator('#run')).toBeEnabled();await expect(page.locator('#minutes')).toHaveValue('2');
 await expect(page.locator('#operationPackage option')).toHaveCount(2);const response=page.waitForResponse(r=>r.url().includes('/api/operations?'));await page.getByRole('button',{name:'Inspect operations',exact:true}).click();const result=await(await response).json();expect(result.toMs-result.fromMs).toBe(120000);await expect(page.locator('#runOperations')).toBeEnabled();await expect(page.locator('#minutes')).toHaveValue('2');await frame(page,'18a-edited-lookback-survives-late-response');
});

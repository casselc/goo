const {test, expect} = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const {randomBytes} = require('node:crypto');
const {startRuntime} = require('./191-generic-runtime');
const {startBombadil} = require('./support/cross-app');

const labels = {logs:'Logs', spans:'Spans', gauge:'Gauges', counter:'Counters', histogram:'Histograms', 'exponential-histogram':'Exponential histograms'};
const tables = {logs:'otel_logs', spans:'otel_traces', gauge:'otel_metrics_gauge', counter:'otel_metrics_sum', histogram:'otel_metrics_histogram', 'exponential-histogram':'otel_metrics_exp_histogram'};

// All mutations are confined to a newly created database carrying this run's
// random identifier. The existing collector/exporter tables are read only.
test('431 available telemetry families remain useful in partial ClickHouse installations', async ({page, request}) => {
  test.setTimeout(240000);
  if (!process.env.GOO_GENERIC_BINARY) throw Error('Set GOO_GENERIC_BINARY to the compiled Go binary');
  const token = randomBytes(6).toString('hex');
  const directory = path.resolve('artifacts/source-capabilities431', token);
  await fs.mkdir(directory, {recursive:true});
  const report = {version:'goo.sourceCapabilities431/1', token, cases:[], errors:[]};
  const base = new URL(process.env.GOO_TEST_CLICKHOUSE_URL || 'http://127.0.0.1:8123');
  const auth = 'Basic '+Buffer.from((process.env.GOO_TEST_CLICKHOUSE_USER || 'lab')+':'+(process.env.GOO_TEST_CLICKHOUSE_PASSWORD || 'lab-local')).toString('base64');
  async function sql(text) {
    const url = new URL(base); url.searchParams.set('date_time_input_format','best_effort');
    const response = await request.post(url.href, {data:text, headers:{Authorization:auth}});
    expect(response.ok(), 'fixture SQL failed: HTTP '+response.status()).toBe(true);
    return response.text();
  }
  const environment = 'source431-'+token;
  const service = 'ordinary-source431-'+token;
  const timestamp = new Date(Date.now()-1000).toISOString();
  const fromMs = Date.now()-60000, toMs = Date.now()+10000;
  const question = {fromMs,toMs};
  const resource = {'deployment.environment.name':environment, 'telemetry.sdk.language':'python'};
  const rows = {
    logs:{Timestamp:timestamp,ServiceName:service,Body:'ad-hoc logs '+token,SeverityText:'INFO',ResourceAttributes:resource,LogAttributes:{'source431.token':token},ScopeName:'source431',ScopeVersion:'1'},
    spans:{Timestamp:timestamp,ServiceName:service,TraceId:randomBytes(16).toString('hex'),SpanId:randomBytes(8).toString('hex'),SpanName:'ad-hoc span '+token,SpanKind:'Server',Duration:1000000,StatusCode:'Ok',ResourceAttributes:resource,SpanAttributes:{'source431.token':token},ScopeName:'source431',ScopeVersion:'1'},
    gauge:{TimeUnix:timestamp,StartTimeUnix:timestamp,ServiceName:service,MetricName:'ad_hoc_gauge',MetricUnit:'1',Value:7,ResourceAttributes:resource,Attributes:{'source431.token':token},ScopeName:'source431',ScopeVersion:'1'},
  };
  async function assertCapabilityLayout(box, entry, width) {
    const layout = await box.locator('.source-capability-table').evaluate(viewport => {
      const container = viewport.getBoundingClientRect();
      const caption = viewport.querySelector('caption').getBoundingClientRect();
      const first = viewport.querySelector('tbody tr').getBoundingClientRect();
      return {containerWidth:container.width, captionWidth:caption.width, captionHeight:caption.height, firstRowGap:first.top-caption.bottom};
    });
    expect(layout.containerWidth).toBeGreaterThan(0);
    expect(layout.captionWidth).toBeGreaterThanOrEqual(layout.containerWidth*0.7);
    expect(layout.captionHeight).toBeGreaterThan(0);
    expect(layout.captionHeight).toBeLessThan(120);
    expect(layout.firstRowGap).toBeGreaterThanOrEqual(-1);
    expect(layout.firstRowGap).toBeLessThanOrEqual(160);
    (entry.layouts ||= []).push({width,...layout});
  }
  page.on('pageerror', error => report.errors.push(error.message));
  try {
  for (const profile of ['logs','spans','gauge','incompatible','absent']) {
    const database = `goo431_${token}_${profile}`;
    if (!/^goo431_[a-f0-9]{12}_(logs|spans|gauge|incompatible|absent)$/.test(database)) throw Error('Fixture database ownership mismatch');
    let runtime, created = false;
    const entry = {profile,database,calls:[]}; report.cases.push(entry);
    const onRequest = req => { if (req.method()==='POST' && new URL(req.url()).pathname.startsWith('/api/')) entry.calls.push({path:new URL(req.url()).pathname,body:req.postDataJSON()}); };
    page.on('request',onRequest);
    try {
      await sql(`CREATE DATABASE ${database}`); created=true;
      if (profile==='incompatible') {
        await sql(`CREATE TABLE ${database}.otel_logs (Timestamp DateTime64(9), Body String) ENGINE=Memory`);
      } else if (tables[profile]) {
        await sql(`CREATE TABLE ${database}.${tables[profile]} AS default.${tables[profile]}`);
        await sql(`INSERT INTO ${database}.${tables[profile]} FORMAT JSONEachRow\n${JSON.stringify(rows[profile])}`);
      }
      const endpoint = new URL(base); endpoint.searchParams.set('database',database);
      runtime = await startRuntime(process.env.GOO_GENERIC_BINARY,path.join(directory,profile),{clickHouseURL:endpoint.href});
      const response = await request.post(runtime.base+'/api/source-capabilities',{data:question});
      expect(response.ok(), await response.text()).toBe(true);
      const result = await response.json(); entry.capabilities=result;
      expect(result.version).toBe('goo/source-capabilities/v1'); expect(result.question).toEqual(question);
      expect(result.families.map(family=>family.kind)).toEqual(Object.keys(tables));
      for (const family of result.families) {
        expect(family.table).toBe(tables[family.kind]);
        expect(family.observation.unit).toBe(['logs','spans'].includes(family.kind)?'records':'exported-points');
        if (family.kind===profile) {
          expect(family.state).toBe('compatible'); expect(family.observation.state).toBe('observed');
        } else if (profile==='incompatible' && family.kind==='logs') {
          expect(family.state).toBe('incompatible'); expect(family.observation.state).toBe('unknown');
        } else {
          expect(family.state).toBe('absent'); expect(family.observation.state).toBe('unknown');
        }
      }
      // A successful schema/readiness check is separate from whether this exact
      // event-time window contains records. No table is truncated to get empty.
      if (tables[profile]) {
        const emptyResponse=await request.post(runtime.base+'/api/source-capabilities',{data:{fromMs:toMs+1000,toMs:toMs+61000}});
        expect(emptyResponse.ok()).toBe(true); const empty=await emptyResponse.json();
        const family=empty.families.find(family=>family.kind===profile);
        expect(family.state).toBe('compatible'); expect(family.observation.state).toBe('empty'); entry.empty=family;
      }
      const inventoryResponse=await request.post(runtime.base+'/api/discovery',{data:{environment,fromMs,toMs}});
      if (tables[profile]) {
        expect(inventoryResponse.ok(),await inventoryResponse.text()).toBe(true);
        const inventory=await inventoryResponse.json(); entry.inventory=inventory;
        expect(inventory.signals).toHaveLength(1);
        expect(inventory.signals[0]).toMatchObject({service,signal:profile==='gauge'?'metrics':profile,count:'1'});
        expect(inventory.dependencies).toEqual([]);
        expect(inventory.sources.filter(source=>source.state==='compatible').map(source=>source.kind)).toEqual([profile]);
      } else {
        expect(inventoryResponse.status()).toBe(502);
      }
      if (profile==='gauge') {
        const metrics=await request.post(runtime.base+'/api/metric-discovery',{data:{environment,service,fromMs,toMs}});
        expect(metrics.ok(),await metrics.text()).toBe(true); const discovered=await metrics.json(); entry.instruments=discovered;
        expect(discovered.records).toBe(1); expect(discovered.instruments).toHaveLength(1);
        expect(discovered.instruments[0].metric).toBe('ad_hoc_gauge');
        expect(discovered.sources.filter(source=>source.state==='compatible').map(source=>source.kind)).toEqual(['gauge']);
      }
      await page.goto(runtime.base+'/'); await expect(page.locator('html')).toHaveAttribute('data-profile','generic');
      const box=page.locator('.environment-adoption');
      await box.locator(':scope>summary').click();
      const before=entry.calls.length;
      await box.getByRole('button',{name:'Read available signals',exact:true}).click();
      const capabilityTable=box.locator('.source-capabilities table');
      await expect(capabilityTable.locator('tbody tr')).toHaveCount(6);
      expect(entry.calls.slice(before).map(call=>call.path)).toEqual(['/api/source-capabilities']);
      if (['logs','spans'].includes(profile)) {
        for (const width of [1440,390,320]) {
          await page.setViewportSize({width,height:1000});
          expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
          await assertCapabilityLayout(box,entry,width);
          await page.screenshot({path:`docs/journeys/screenshots/source-capabilities-${profile}-${width}.png`});
        }
        const choose=box.getByRole('button',{name:'Choose '+profile,exact:true});
        const beforeChoose=entry.calls.length;
        await choose.focus(); await page.keyboard.press('Enter');
        expect(entry.calls).toHaveLength(beforeChoose);
        await expect(page.getByLabel('Live signal',{exact:true})).toHaveValue(profile);
        await expect(box.getByRole('button',{name:'Read environment scopes',exact:true})).toBeFocused();
        entry.localChoose=true;
        // End in records through the existing checked environment adoption flow.
        await box.getByRole('button',{name:'Read environment scopes',exact:true}).click();
        const group=box.locator('.environment-label-table tbody tr').filter({hasText:`Named environment: "${environment}"`});
        await expect(group).toBeVisible(); await group.getByRole('button').click();
        await box.getByRole('button',{name:'Stage in Activity',exact:true}).click();
        const shell=page.locator('.stream-entry'); await shell.getByRole('button',{name:'Start stream',exact:true}).click();
        await expect(shell.locator('.live-tail')).toContainText(token);
        await shell.getByRole('button',{name:'Pause stream',exact:true}).click();
        // Choosing from new availability results must stay local even when a
        // clean applied selector exists (the ordinary signal change restarts it).
        await box.locator(':scope>summary').click();
        await box.getByRole('button',{name:'Read available signals',exact:true}).click();
        await expect(box.getByRole('button',{name:'Choose '+profile,exact:true})).toBeVisible();
        const beforeRechoose=entry.calls.length;
        await box.getByRole('button',{name:'Choose '+profile,exact:true}).click();
        expect(entry.calls).toHaveLength(beforeRechoose);
        await expect(box.getByRole('button',{name:'Read environment scopes',exact:true})).toBeFocused();
        entry.appliedChooseLocal=true;
      } else {
        expect(await box.getByRole('button',{name:/^Choose (logs|spans)$/}).count()).toBe(0);
        await page.setViewportSize({width:390,height:1000});
        expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
        await assertCapabilityLayout(box,entry,390);
        await page.screenshot({path:`docs/journeys/screenshots/source-capabilities-${profile}-390.png`});
      }
      if (profile==='spans' && process.env.GOO_SOURCE431_BOMBADIL==='1') {
        // Refresh this owned source after the deterministic flow: the live tail
        // is an event-time window, not a replay of the earlier source snapshot.
        const fresh={...rows.spans,Timestamp:new Date(Date.now()-500).toISOString(),TraceId:randomBytes(16).toString('hex'),SpanId:randomBytes(8).toString('hex'),SpanName:'actor '+token};
        await sql(`INSERT INTO ${database}.otel_traces FORMAT JSONEachRow\n${JSON.stringify(fresh)}`);
        const actorURL=runtime.base+'/?source431Kind=spans&source431Env='+encodeURIComponent(environment);
        const actor=await startBombadil({dir:path.join(directory,'bombadil-'+token),url:actorURL,specification:'verification/bombadil/source-capabilities431.ts',seconds:30,width:390,height:844});
        try {
          const campaign=await actor.finish({snapshot:{name:'state',version:'goo.sourceCapabilities431Actor/1'}});
          const complete=campaign.states.find(state=>state.phase==='complete');
          expect(complete).toBeTruthy();
          expect([complete.capabilityReads,complete.environmentReads,complete.previewReads,complete.tailReads]).toEqual([1,1,1,1]);
          expect(complete.capabilityObserved).toBe(true);expect(complete.scopeSafe).toBe(true);expect(complete.focusSafe).toBe(true);expect(complete.localSafe).toBe(true);expect(complete.forbidden).toEqual([]);
          entry.bombadil={complete,snapshots:campaign.states.length};
        } finally {await actor.stop();}
      }
      entry.status='passed';
    } finally {
      page.off('request',onRequest);
      await page.goto('about:blank',{timeout:5000}).catch(()=>{});
      try { await runtime?.stop(); }
      finally { if (created) { await sql(`DROP DATABASE ${database} SYNC`); entry.cleaned=true; } }
    }
  }
  expect(report.errors).toEqual([]); report.status='passed';
  } catch (error) { report.status='failed';report.error=error.message;throw error; }
  finally { await fs.writeFile(path.join(directory,'report.json'),JSON.stringify(report,null,2)); }
  console.log('431 source capability report:',path.join(directory,'report.json'));
});

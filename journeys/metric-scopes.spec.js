const {test, expect} = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const {randomBytes} = require('node:crypto');
const {startRuntime} = require('./191-generic-runtime');
const {startBombadil} = require('./support/cross-app');
const ServiceIdentity = require('../ui/service-identity.mjs');
const RecordEnvironment = require('../ui/record-environment.mjs');

const normalizedNamespace = value => ({mode:value.mode,...(value.mode==='value'&&value.value?{value:value.value}:{})});
const selector = environment => ({version:'goo/environment-selector/v1',policy:'effective',scope:environment});

// This is a source-adapter acceptance fixture. It inserts exporter-shaped points
// only into its uniquely owned database, without replacing the lab's tables.
test('432 metrics-only users discover exact environment and service scopes without guessed labels', async ({page, request}) => {
  test.setTimeout(240000);
  if (!process.env.GOO_GENERIC_BINARY) throw Error('Set GOO_GENERIC_BINARY to the compiled Go binary');
  const token=randomBytes(6).toString('hex'),database='goo432_'+token;
  if (!/^goo432_[a-f0-9]{12}$/.test(database)) throw Error('Fixture ownership mismatch');
  const directory=path.resolve('artifacts/metric-scopes432',token);
  await fs.mkdir(directory,{recursive:true});
  const report={version:'goo.metricScopes432/1',token,database,calls:[],facts:[],errors:[]};
  const base=new URL(process.env.GOO_TEST_CLICKHOUSE_URL||'http://127.0.0.1:8123');
  const auth='Basic '+Buffer.from((process.env.GOO_TEST_CLICKHOUSE_USER||'lab')+':'+(process.env.GOO_TEST_CLICKHOUSE_PASSWORD||'lab-local')).toString('base64');
  async function sql(text){const url=new URL(base);url.searchParams.set('date_time_input_format','best_effort');const response=await request.post(url.href,{data:text,headers:{Authorization:auth}});expect(response.ok(),'fixture SQL failed: HTTP '+response.status()).toBe(true);return response.text();}
  const service='ordinary-metrics432-'+token,named='observed-'+token,legacy='legacy-'+token;
  const expected=[
    {environment:{mode:'unreported'},namespace:{mode:'missing'},points:'2'},
    {environment:{mode:'empty'},namespace:{mode:'value'},points:'2'},
    {environment:{mode:'named',value:named},namespace:{mode:'value',value:'api'},points:'4'},
    {environment:{mode:'named',value:legacy},namespace:{mode:'value',value:'workers'},points:'2'},
    {environment:{mode:'named',value:named},namespace:{mode:'missing'},points:'2'},
    {environment:{mode:'named',value:named},namespace:{mode:'value'},points:'2'},
  ];
  const resources=[
    {},
    {'deployment.environment.name':'','deployment.environment':'ignored-'+token,'service.namespace':''},
    {'deployment.environment.name':named,'deployment.environment':'conflicting-'+token,'service.namespace':'api'},
    {'deployment.environment':named,'service.namespace':'api'},
    {'deployment.environment':legacy,'service.namespace':'workers'},
    {'deployment.environment.name':named},
    {'deployment.environment.name':named,'service.namespace':''},
    {missingService:true},
  ];
  const sourcePaths=new Set(['/api/source-capabilities','/api/metric-scopes','/api/metric-discovery','/api/measurements']);
  const sourceCalls=()=>report.calls.filter(call=>sourcePaths.has(call.path));
  const beforeCalls=()=>sourceCalls().length;
  const recordFact=(name,before)=>report.facts.push({name,sourceRequests:beforeCalls()-before});
  let runtime,created=false;
  page.on('pageerror',error=>report.errors.push(error.message));
  page.on('request',req=>{if(req.method()==='POST'&&new URL(req.url()).pathname.startsWith('/api/'))report.calls.push({path:new URL(req.url()).pathname,body:req.postDataJSON()});});
  try {
    await sql('CREATE DATABASE '+database);created=true;
    await sql(`CREATE TABLE ${database}.otel_metrics_gauge AS default.otel_metrics_gauge`);
    // Explicit instance identity is fixture telemetry, required by the existing
    // service measurement kernel; scope discovery itself does not invent it.
    const now=Date.now(),rows=resources.flatMap((resource,index)=>[20000,10000].map((age,point)=>({
      TimeUnix:new Date(now-age).toISOString(),StartTimeUnix:new Date(now-30000).toISOString(),ServiceName:resource.missingService?'':service,
      MetricName:'queue.depth',MetricUnit:'1',Value:7+point,ScopeName:'ordinary.instrumentation',ScopeVersion:'1',
      ResourceAttributes:{'telemetry.sdk.language':'python','service.instance.id':'instance-'+index,...Object.fromEntries(Object.entries(resource).filter(([key])=>key!=='missingService'))},Attributes:{'fixture.token':token},
    })));
    await sql(`INSERT INTO ${database}.otel_metrics_gauge FORMAT JSONEachRow\n`+rows.map(JSON.stringify).join('\n'));
    const endpoint=new URL(base);endpoint.searchParams.set('database',database);
    runtime=await startRuntime(process.env.GOO_GENERIC_BINARY,directory,{clickHouseURL:endpoint.href});
    await page.goto(runtime.base+'/');
    await expect(page.locator('html')).toHaveAttribute('data-profile','generic');
    const findData=page.locator('.environment-adoption');
    await findData.locator(':scope>summary').click();expect(beforeCalls()).toBe(0);
    const availableResponse=page.waitForResponse(response=>response.url().endsWith('/api/source-capabilities')&&response.request().method()==='POST');
    await findData.getByRole('button',{name:'Read available signals',exact:true}).click();
    const available=await (await availableResponse).json();report.capabilities=available;
    expect(available.families.filter(family=>family.state==='compatible').map(family=>family.kind)).toEqual(['gauge']);
    expect(available.families.find(family=>family.kind==='gauge').observation.state).toBe('observed');
    const capabilityQuestion=report.calls.filter(call=>call.path==='/api/source-capabilities').at(-1).body;
    const beforeHandoff=beforeCalls();
    await findData.locator('.source-capabilities').getByRole('button',{name:'Find metric scopes',exact:true}).click();
    const workbench=page.locator('#measurementWorkbench'),discovery=workbench.locator('#metricDiscovery'),scopes=discovery.locator('.metric-scopes');
    await expect(page.getByRole('tab',{name:'Measurements',exact:true})).toHaveAttribute('aria-selected','true');
    await expect(workbench.locator('#metricEnvironment')).toHaveValue('');
    await expect(workbench.locator('#metricService')).toHaveValue('');
    await expect(scopes.getByRole('button',{name:'Read metric scopes',exact:true})).toBeFocused();
    await page.waitForTimeout(100);expect(beforeCalls()).toBe(beforeHandoff);recordFact('metric-handoff',beforeHandoff);
    const read=scopes.getByRole('button',{name:'Read metric scopes',exact:true});
    const pending=page.waitForResponse(response=>response.url().endsWith('/api/metric-scopes')&&response.request().method()==='POST');
    await read.focus();await page.keyboard.press('Enter');
    const response=await pending;expect(response.ok(),await response.text()).toBe(true);
    const found=await response.json();report.discovery=found;
    const scopeQuestion=structuredClone(found.question);expect(scopeQuestion).toEqual(capabilityQuestion);
    expect(report.calls.filter(call=>call.path==='/api/metric-scopes').at(-1).body).toEqual(scopeQuestion);
    expect(found.version).toBe('goo/metric-scopes/v1');expect(found.unit).toBe('exported-points');
    expect(found.sources.filter(source=>source.state==='compatible').map(source=>source.kind)).toEqual(['gauge']);
    expect(found.choices).toHaveLength(7);
    const key=choice=>JSON.stringify([choice.service,normalizedNamespace(choice.namespace),choice.environment]);
    for (const fixture of expected) {
      const choice=found.choices.find(choice=>key(choice)===key({service,...fixture}));
      expect(choice).toBeTruthy();expect(choice.selectable).toBe(true);expect(choice.points).toBe(fixture.points);
      expect(choice.environmentSelector).toEqual(selector(fixture.environment));
    }
    const missingService=found.choices.find(choice=>choice.service==='');
    expect(missingService).toBeTruthy();expect(missingService.selectable).toBe(false);expect(missingService.reason).toBeTruthy();
    await expect(scopes.locator('.metric-scope-table tbody tr')).toHaveCount(7);
    expect(await scopes.getByRole('button',{name:/^Choose metric scope/}).evaluateAll(buttons=>buttons.filter(button=>!button.disabled).length)).toBe(6);
    async function layout(width){
      const height=width<=390?900:1000;
      await page.setViewportSize({width,height});
      const geometry=await scopes.locator('.metric-scope-table').evaluate(host=>{
        const c=host.getBoundingClientRect(),caption=host.querySelector('caption').getBoundingClientRect(),first=host.querySelector('tbody tr').getBoundingClientRect();
        const choice=host.querySelector('tbody button:not(:disabled)').getBoundingClientRect();
        const input=host.closest('.metric-scopes').querySelector('input[aria-label="Metric scope window (minutes)"]').getBoundingClientRect();
        return {containerWidth:c.width,captionWidth:caption.width,captionHeight:caption.height,firstRowGap:first.top-caption.bottom,firstChoiceTop:choice.top,firstChoiceBottom:choice.bottom,inputHeight:input.height};
      });
      expect(geometry.captionWidth).toBeGreaterThanOrEqual(geometry.containerWidth*0.7);expect(geometry.captionHeight).toBeGreaterThan(0);expect(geometry.captionHeight).toBeLessThan(120);expect(geometry.firstRowGap).toBeGreaterThanOrEqual(-1);expect(geometry.firstRowGap).toBeLessThanOrEqual(160);
      expect(geometry.inputHeight).toBeGreaterThan(0);expect(geometry.inputHeight).toBeLessThanOrEqual(44);
      if(width<=390){expect(geometry.firstChoiceTop).toBeGreaterThanOrEqual(0);expect(geometry.firstChoiceBottom).toBeLessThanOrEqual(900);}
      expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
      (report.layouts ||= []).push({width,height,...geometry});
      await page.screenshot({path:`docs/journeys/screenshots/metric-scopes-${width}.png`});
    }
    for (const width of [1440,390,320]) await layout(width);
    await page.setViewportSize({width:1440,height:1100});
    await read.evaluate(element=>window.__metric432Read=element);
    await scopes.getByRole('button',{name:/^Choose metric scope/}).filter({hasText:'Choose metric scope'}).first().evaluate(element=>window.__metric432Choice=element);
    await scopes.locator(':scope>summary').click();
    const beforeWithdrawn=beforeCalls();
    await page.evaluate(()=>{window.__metric432Read.click();window.__metric432Choice.click();});
    await page.waitForTimeout(100);expect(beforeCalls()).toBe(beforeWithdrawn);
    await expect(workbench.locator('#metricService')).toHaveValue('');recordFact('withdrawn-controls',beforeWithdrawn);
    let finalChoice;
    for (const fixture of expected) {
      if (!await scopes.evaluate(node=>node.open)) await scopes.locator(':scope>summary').click();
      if (!await scopes.locator('.metric-scope-table tbody tr').count()) {
        const refresh=page.waitForResponse(response=>response.url().endsWith('/api/metric-scopes')&&response.request().method()==='POST');
        await read.click();const refreshed=await (await refresh).json();expect(refreshed.choices).toHaveLength(7);
        // A new explicit read creates a new fixed discovery window.
        Object.assign(scopeQuestion,refreshed.question);
      }
      const label='Choose metric scope · '+ServiceIdentity.name(service,fixture.namespace)+' · '+RecordEnvironment.label(selector(fixture.environment));
      const choose=scopes.getByRole('button',{name:label,exact:true});
      const before=beforeCalls();await choose.focus();await page.keyboard.press('Enter');
      await expect(workbench.locator('#metricDiscover')).toBeFocused();await page.waitForTimeout(100);
      expect(beforeCalls()).toBe(before);recordFact('choose-'+JSON.stringify([fixture.environment,fixture.namespace]),before);
      const instrumentResponse=page.waitForResponse(response=>response.url().endsWith('/api/metric-discovery')&&response.request().method()==='POST');
      await page.keyboard.press('Enter');const instruments=await (await instrumentResponse).json();
      const question=report.calls.filter(call=>call.path==='/api/metric-discovery').at(-1).body;
      expect(question).toMatchObject({service,fromMs:scopeQuestion.fromMs,toMs:scopeQuestion.toMs,environmentSelector:selector(fixture.environment)});
      expect(normalizedNamespace(question.namespace)).toEqual(normalizedNamespace(fixture.namespace));
      expect(instruments.records).toBe(Number(fixture.points));expect(instruments.instruments).toHaveLength(1);
      expect(instruments.instruments[0].metric).toBe('queue.depth');
      (report.instrumentReads ||= []).push({question,result:instruments});
      finalChoice={fixture,question};
    }
    await page.screenshot({path:'docs/journeys/screenshots/metric-scopes-instruments-1440.png'});
    const beforeBuild=beforeCalls(),planResponse=page.waitForResponse(response=>response.url().endsWith('/api/measurement-packages/plan')&&response.request().method()==='POST');
    await workbench.getByRole('button',{name:'Build measurement package',exact:true}).click();
    const planned=await (await planResponse).json();report.plan=planned;
    await expect(workbench.locator('#measureStatus')).toContainText('Question validated');
    expect(beforeCalls()).toBe(beforeBuild);recordFact('build-preview',beforeBuild);
    expect(planned.question).toMatchObject({fromMs:finalChoice.question.fromMs,toMs:finalChoice.question.toMs,definition:{environmentSelector:finalChoice.question.environmentSelector,source:{services:[service]}}});
    expect(normalizedNamespace(planned.question.definition.source.namespace)).toEqual(normalizedNamespace(finalChoice.fixture.namespace));
    await page.setViewportSize({width:390,height:1000});expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
    await page.screenshot({path:'docs/journeys/screenshots/metric-scopes-staged-390.png'});
    const measurementResponse=page.waitForResponse(response=>response.url().endsWith('/api/measurements')&&response.request().method()==='POST');
    await workbench.locator('#measureRun').click();const measurementReply=await measurementResponse;const evaluated=await measurementReply.json();report.result=evaluated;expect(measurementReply.ok(),JSON.stringify(evaluated)).toBe(true);
    expect(evaluated.count).toBe('2');expect(evaluated.rows).toHaveLength(2);expect(evaluated.gauges.series).toHaveLength(1);
    expect(evaluated.question).toEqual(planned.question);
    await page.screenshot({path:'docs/journeys/screenshots/metric-scopes-evaluated-390.png'});
    if (process.env.GOO_METRIC_SCOPES432_BOMBADIL==='1') {
      const actorNow=Date.now();
      const actorRows=rows.map((row,index)=>({...row,TimeUnix:new Date(actorNow-1000+(index%2)*100).toISOString(),StartTimeUnix:new Date(actorNow-2000).toISOString()}));
      await sql(`INSERT INTO ${database}.otel_metrics_gauge FORMAT JSONEachRow\n`+actorRows.map(JSON.stringify).join('\n'));
      report.actorFixture={points:actorRows.length,readWindow:'fresh independent scope read'};
      const actor=await startBombadil({dir:path.join(directory,'bombadil'),url:runtime.base+'/?metric432Service='+encodeURIComponent(service),specification:'verification/bombadil/metric-scopes432.ts',seconds:30,width:390,height:1000});
      try {const campaign=await actor.finish({snapshot:{name:'state',version:'goo.metricScopes432Actor/1'}}),complete=campaign.states.find(state=>state.phase==='complete');expect(complete).toBeTruthy();expect(complete.localSafe).toBe(true);expect(complete.scopeSafe).toBe(true);expect(complete.forbidden).toEqual([]);expect([complete.scopeReads,complete.instrumentReads,complete.plans]).toEqual([1,1,1]);report.bombadil={complete,snapshots:campaign.states.length};}finally{await actor.stop();}
    }
    expect(report.errors).toEqual([]);
    expect(report.calls.every(call=>sourcePaths.has(call.path)||call.path==='/api/measurement-packages/plan'||/\/api\/queries\/[^/]+\/cancel$/.test(call.path))).toBe(true);
    report.status='passed';
  } catch(error) {report.status='failed';report.failure=error.message;throw error;}
  finally {
    try {await page.goto('about:blank',{timeout:5000}).catch(()=>{});}finally{try{await runtime?.stop();}finally{if(created){await sql('DROP DATABASE '+database+' SYNC');report.cleaned=true;}await fs.writeFile(path.join(directory,'report.json'),JSON.stringify(report,null,2));}}
  }
});

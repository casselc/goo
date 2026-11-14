const {test, expect} = require('@playwright/test');
const fs = require('node:fs/promises');
const path = require('node:path');
const {randomBytes} = require('node:crypto');
const {startRuntime} = require('./191-generic-runtime');
const {source, startBombadil} = require('./support/cross-app');

test('430 discover unfamiliar OTLP scopes before starting Activity, with keyboard-safe pause', async ({page, request}) => {
  test.setTimeout(180000);
  if (!process.env.GOO_GENERIC_BINARY) throw Error('Set GOO_GENERIC_BINARY to the compiled Go binary');
  const token = randomBytes(5).toString('hex');
  const directory = path.resolve('artifacts/unknown-data-entry430');
  const report = {version:'goo.unknownDataEntry430/1', token, calls:[], facts:[], errors:[]};
  const runtime = await startRuntime(process.env.GOO_GENERIC_BINARY, directory);
  const attrs = object => Object.entries(object).map(([key, value]) => ({key, value:{stringValue:value}}));
  const scopes = [
    {mode:'unreported', resource:{}},
    {mode:'empty', resource:{'deployment.environment.name':'', 'deployment.environment':'ignored-'+token}},
    {mode:'named', name:'ordinary-'+token, resource:{'deployment.environment.name':'ordinary-'+token, 'deployment.environment':'conflict-'+token}},
    {mode:'legacy', name:'legacy-'+token, resource:{'deployment.environment':'legacy-'+token}},
  ];
  const calls = () => report.calls.length;
  const fact = (name, before) => report.facts.push({name, posts:calls()-before});
  page.on('pageerror', error => report.errors.push(error.message));
  page.on('request', req => {
    if (req.method() === 'POST' && new URL(req.url()).pathname.startsWith('/api/')) {
      report.calls.push({path:new URL(req.url()).pathname, body:req.postDataJSON()});
    }
  });
  try {
    const timestamp = BigInt(Date.now()-500)*1000000n;
    for (const signal of ['logs','spans']) {
      const resources = scopes.map((scope, index) => {
        const resource = {attributes:attrs({'service.name':'entry430-'+token+'-'+scope.mode, 'telemetry.sdk.language':'python', ...scope.resource})};
        const common = {traceId:randomBytes(16).toString('hex'), spanId:randomBytes(8).toString('hex'), attributes:attrs({'entry430.token':token})};
        return signal === 'logs' ? {resource, scopeLogs:[{logRecords:[{...common, timeUnixNano:String(timestamp+BigInt(index)), severityText:'INFO', body:{stringValue:token+' '+scope.mode}}]}]} :
          {resource, scopeSpans:[{spans:[{...common, name:token+' '+scope.mode, kind:2, startTimeUnixNano:String(timestamp+BigInt(index)), endTimeUnixNano:String(timestamp+1000000n+BigInt(index))}]}]};
      });
      const response = await request.post('http://127.0.0.1:4318/v1/'+(signal==='logs'?'logs':'traces'), {data:signal==='logs'?{resourceLogs:resources}:{resourceSpans:resources}});
      expect(response.ok()).toBe(true);
    }
    for (const table of ['otel_logs','otel_traces']) {
      await expect.poll(async () => Number((await source(request, `SELECT count() AS n FROM default.${table} WHERE startsWith(ServiceName,{service:String}) FORMAT JSONEachRow`, {param_service:'entry430-'+token+'-'}))[0].n)).toBe(4);
    }
    await page.goto(runtime.base+'/');
    await expect(page.locator('html')).toHaveAttribute('data-profile','generic');
    const shell = page.locator('.stream-entry');
    const chooser = shell.locator('.environment-adoption');
    const summary = chooser.locator(':scope>summary');
    const tail = shell.locator('.live-tail');
    const signal = shell.getByLabel('Live signal',{exact:true});
    await expect(shell.getByRole('button',{name:'Start stream',exact:true})).toBeVisible();
    expect(calls()).toBe(0);
    expect(await chooser.evaluate(node => !!node.closest('.stream-entry-toolbar'))).toBe(true);
    await shell.getByLabel('Live environment',{exact:true}).fill('not-applied');
    await expect(shell.getByRole('button',{name:'Start stream',exact:true})).toBeVisible();
    await shell.getByLabel('Live environment',{exact:true}).fill('');
    expect(calls()).toBe(0);
    for (const width of [1440,390,320]) {
      await page.setViewportSize({width,height:900});
      expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
      await page.screenshot({path:`docs/journeys/screenshots/unknown-data-entry-${width}.png`});
    }
    await page.setViewportSize({width:1440,height:1100});
    for (const scope of scopes) {
      if (!await chooser.evaluate(node=>node.open)) { await summary.focus(); await page.keyboard.press('Enter'); }
      const read = chooser.getByRole('button',{name:'Read environment scopes',exact:true});
      let before = calls();
      await read.focus(); await page.keyboard.press('Enter');
      const row = chooser.locator('.environment-label-table tbody tr').filter({hasText:scope.name?`Named environment: "${scope.name}"`:scope.mode==='empty'?'Empty environment label':'Unreported environment'}).first();
      await expect(row).toBeVisible();
      fact('discover-'+scope.mode,before);
      const preview = row.getByRole('button');
      await expect(preview).toHaveAccessibleName(scope.name?'Preview named logs · '+JSON.stringify(scope.name):'Preview '+scope.mode+' logs');
      expect(await row.locator('th[scope=row]').count()).toBe(1);
      before=calls(); await preview.focus(); await page.keyboard.press('Enter');
      const stage=chooser.getByRole('button',{name:'Stage in Activity',exact:true});
      await expect(stage).toBeVisible();
      fact('preview-'+scope.mode,before);
      if (scope.mode==='unreported') {
        await read.evaluate(node=>window.__entry430Read=node);
        await summary.click(); before=calls();
        await page.evaluate(()=>window.__entry430Read.click());
        expect(calls()).toBe(before); fact('closed-read-refused',before);
        await summary.click(); await read.click(); await expect(row).toBeVisible(); await row.getByRole('button').click(); await expect(stage).toBeVisible();
      }
      before=calls(); await stage.focus(); await page.keyboard.press('Enter');
      fact('stage-'+scope.mode,before); expect(calls()).toBe(before);
      const start=shell.getByRole('button',{name:'Start stream',exact:true});
      await expect(start).toBeFocused();
      await expect(tail.locator('tbody tr')).toHaveCount(0);
      await page.keyboard.press('Enter');
      await expect(tail).toContainText(token+' '+scope.mode);
      const selector=report.calls.filter(call=>call.path==='/api/tail').at(-1).body.environmentSelector;
      expect(selector).toEqual({version:'goo/environment-selector/v1',policy:'effective',scope:{mode:scope.mode==='legacy'?'named':scope.mode,...(scope.name?{value:scope.name}:{})}});
      const pause=shell.getByRole('button',{name:'Pause stream',exact:true});
      await pause.focus(); await page.keyboard.press('Enter');
      const resume=shell.getByRole('button',{name:'Resume with recent logs',exact:true});
      await expect(resume).toBeFocused();
      if (scope.mode==='unreported') {
        for (const width of [1440,390,320]) {
          await page.setViewportSize({width,height:900});
          expect(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);
          await page.screenshot({path:`docs/journeys/screenshots/unknown-data-paused-${width}.png`});
        }
        await page.setViewportSize({width:1440,height:1100});
        await resume.focus(); await page.keyboard.press('Enter'); await expect(pause).toBeFocused();
        await expect(tail).toContainText(token+' '+scope.mode);
        await page.keyboard.press('Enter'); await expect(resume).toBeFocused();
        fact('pause-resume-focus',calls());
      }
    }
    const beforeSignal=calls(); await signal.selectOption('spans');
    await expect.poll(calls).toBe(beforeSignal+1);
    expect(report.calls.at(-1).path).toBe('/api/tail');
    expect(report.calls.at(-1).body.signal).toBe('spans');
    fact('switch-applied-signal',beforeSignal);
    if (!await chooser.evaluate(node=>node.open)) await summary.click();
    await chooser.getByRole('button',{name:'Read environment scopes',exact:true}).click();
    const spanRow=chooser.locator('.environment-label-table tbody tr').filter({hasText:`Named environment: "${scopes[2].name}"`});
    await expect(spanRow).toBeVisible(); await spanRow.getByRole('button').click();
    await chooser.getByRole('button',{name:'Stage in Activity',exact:true}).click();
    await shell.getByRole('button',{name:'Start stream',exact:true}).click();
    await expect(tail).toContainText(token+' named');
    expect(report.calls.filter(call=>call.path==='/api/tail').at(-1).body.signal).toBe('spans');
    await shell.getByRole('button',{name:'Pause stream',exact:true}).click();
    if (process.env.GOO_ENTRY430_BOMBADIL==='1') {
      const fresh=String(BigInt(Date.now()-500)*1000000n);
      const actorLog={resourceLogs:[{resource:{attributes:attrs({'service.name':'entry430-'+token+'-actor',...scopes[2].resource})},scopeLogs:[{logRecords:[{timeUnixNano:fresh,severityText:'INFO',body:{stringValue:'actor '+token}}]}]}]};
      expect((await request.post('http://127.0.0.1:4318/v1/logs',{data:actorLog})).ok()).toBe(true);
      await expect.poll(async()=>Number((await source(request,'SELECT count() AS n FROM default.otel_logs WHERE ServiceName={service:String} FORMAT JSONEachRow',{param_service:'entry430-'+token+'-actor'}))[0].n)).toBe(1);
      const actor=await startBombadil({dir:path.join(directory,'bombadil-'+token),url:runtime.base+'/?entry430='+scopes[2].name,specification:'verification/bombadil/unknown-data-entry430.ts',seconds:30,width:390,height:844});
      try {
        const campaign=await actor.finish({snapshot:{name:'state',version:'goo.unknownDataEntry430Actor/1'}});
        const complete=campaign.states.find(state=>state.phase==='complete');
        expect(complete).toBeTruthy(); expect(complete.focusSafe).toBe(true); expect(complete.scopeSafe).toBe(true); expect(complete.localSafe).toBe(true); expect(complete.forbidden).toEqual([]);
        expect([complete.discoveryReads,complete.previewReads,complete.tailReads]).toEqual([1,1,2]);
        report.bombadil={complete,snapshots:campaign.states.length};
      } finally { await actor.stop(); }
    }
    expect(report.errors).toEqual([]);
    expect(report.calls.every(call=>['/api/environment-scopes','/api/environment-scopes/records','/api/tail','/api/queries/cancel'].includes(call.path)||call.path.endsWith('/cancel'))).toBe(true);
    report.status='passed';
  } finally {
    await fs.writeFile(path.join(directory,'report.json'),JSON.stringify(report,null,2));
    await runtime.stop();
  }
});

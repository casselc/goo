const { test, expect } = require('@playwright/test');
const { observe, seed, discover, canonical, write } = require('./202-pod-memory-support');
test.use({ actionTimeout:15000,navigationTimeout:45000 });
test('202 actual pod metadata and native replies cannot publish after panel ownership changes', async ({ page,request }) => {
  test.setTimeout(180000); const facts=await seed(request), observation=observe(page), cases=[];
  for(const item of [{phase:'metadata',reason:'filter'},{phase:'native',reason:'close'},{phase:'native',reason:'row'},{phase:'metadata',reason:'tab'},{phase:'native',reason:'scope'}]) {
    const initial=await discover(page,facts), primary=initial.response.observations.find(r=>canonical(r.resource)===canonical(facts.primary)&&r.resourceSchemaURL===''); expect(primary).toBeTruthy();
    let release,entered,finished,held; const gate=new Promise(r=>{release=r}),started=new Promise(r=>{entered=r}),done=new Promise(r=>{finished=r});
    const pattern=item.phase==='metadata'?'**/api/pod-metric-discovery':'**/api/measurements';
    await page.route(pattern,async route=>{
      const q=route.request().postDataJSON(); if(item.phase==='native'&&(q.definition.source.metric!=='k8s.pod.memory.usage'||q.definition.source.resource.id!=='uid-a')) return route.continue();
      const response=await route.fetch(); held={request:q,status:response.status(),response:await response.json()}; entered(); await gate; try{await route.fulfill({response})}catch{}finally{finished()}
    });
    const beforeOpen={source:observation.source(),plans:observation.plans()}; await page.locator('[data-pod-observe="'+primary.key+'"]').click(); await started; expect(held.status).toBe(200);
    if(item.reason==='filter') await page.getByLabel('Find reported pod resource',{exact:true}).fill('uid-b');
    else if(item.reason==='scope') await page.getByLabel('Kubernetes environment',{exact:true}).fill(facts.environment+'-draft');
    else if(item.reason==='tab') await page.locator('#entitiesServices').click();
    else if(item.reason==='close') await page.locator('.pod-metrics-panel[data-pod-metrics="'+primary.key+'"]').getByRole('button',{name:'Close pod metrics',exact:true}).click();
    else {
      const other=initial.response.observations.find(r=>r.resource['k8s.pod.uid']==='uid-b'&&r.resourceSchemaURL===''); expect(other).toBeTruthy();
      await page.locator('[data-pod-observe="'+other.key+'"]').click();
      await expect(page.locator('.pod-metrics-panel[data-pod-metrics="'+other.key+'"] [role=status]')).toContainText('Independent pod memory observations');
    }
    const snapshot=async()=>({url:page.url(),source:observation.source(),plans:observation.plans(),activeTab:await page.locator('#entitiesKubernetes').getAttribute('aria-selected'),kubernetesHidden:await page.locator('.kubernetes-browser').isHidden(),rows:await page.locator('tr[data-pod-variant]').count(),
      panels:await page.locator('.pod-metrics-panel').evaluateAll(nodes=>nodes.map(n=>({key:n.dataset.podMetrics,hidden:n.hidden,text:n.textContent,widgets:n.querySelectorAll('.pod-metric-widget').length}))) });
    const beforeRelease=await snapshot(); release(); await done; await page.unroute(pattern); await observation.settle(); const afterRelease=await snapshot(); expect(afterRelease).toEqual(beforeRelease);
    if(item.phase==='metadata') expect(afterRelease.plans).toBe(beforeOpen.plans);
    if(item.reason==='row') {expect(afterRelease.panels).toHaveLength(1); expect(afterRelease.panels[0].key).not.toBe(primary.key);expect(afterRelease.panels[0].widgets).toBe(1)}
    cases.push({...item,initial,primary,held,beforeOpen,beforeRelease,afterRelease});
  }
  expect(observation.errors).toEqual([]); await write(facts,'ownership.json',{facts,cases,requests:observation.requests,responses:observation.responses});
});

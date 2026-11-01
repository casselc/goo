const {test,expect}=require('@playwright/test');
const base=process.env.GOO_COMPONENT_URL||'http://127.0.0.1:8088';
// Synthetic supplied data exercises the public component, not source guarantees.
test('categorical facet SSR and detached properties retain controlled focus and local Find without IO',async({page})=>{
 const errors=[],requests=[];page.on('pageerror',error=>errors.push(error.message));page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/'))requests.push(r.url())});
 await page.goto(base+'/ui-components');const ssr=page.locator('#component-categorical-facet'),spa=page.locator('#component-categorical-facet-spa goo-categorical-facet');
 await expect(ssr).toHaveAttribute('data-state','ready');await expect(spa).toHaveAttribute('data-state','ready');
 expect(await ssr.evaluate(e=>e.data.values.total)).toBe('130');expect(await spa.evaluate(e=>e.data)).toEqual(await ssr.evaluate(e=>e.data));
 await expect(ssr.getByRole('button',{name:'Include attribute:http.route Missing key',exact:true})).toBeEnabled();await expect(ssr.getByRole('button',{name:'Include attribute:http.route Present empty string',exact:true})).toBeEnabled();
 const search=ssr.getByRole('searchbox',{name:'Search loaded facet values'});await search.fill('/orders');await expect(ssr.locator('[data-facet-item]:visible')).toHaveCount(1);
 await ssr.getByRole('button',{name:'Include attribute:http.route "/orders"',exact:true}).click();await expect(search).toHaveValue('/orders');await expect(ssr.locator('[data-facet-item]:visible')).toHaveCount(1);
 await ssr.getByRole('button',{name:'Clear staged facet',exact:true}).click();await expect(search).toHaveValue('/orders');await expect(ssr.getByRole('button',{name:'Read values',exact:true})).toBeFocused();
 await ssr.getByRole('button',{name:'Read values',exact:true}).press('Enter');await expect(ssr.getByRole('button',{name:'Read values',exact:true})).toBeFocused();
 await ssr.evaluate(e=>{e.fixtureReady=e.data;e.data={declaration:e.data.declaration,values:{state:'loading'}}});await expect(ssr.getByRole('status')).toBeFocused();await ssr.evaluate(e=>{e.data=e.fixtureReady});await expect(ssr.getByRole('button',{name:'Read values',exact:true})).toBeFocused();
 await ssr.evaluate(e=>{e.data={declaration:e.data.declaration,values:{state:'loading'}}});const heading=page.getByRole('heading',{name:'UI component reference',exact:true});await heading.evaluate(e=>{e.tabIndex=-1;e.focus()});await ssr.evaluate(e=>{e.data=e.fixtureReady});await expect(heading).toBeFocused();
 await ssr.evaluate(e=>{const data=e.data;data.values.total='131';e.setAttribute('data',JSON.stringify(data))});await expect(ssr).toHaveAttribute('data-state','unavailable');await expect(ssr.locator('[data-facet-item]')).toHaveCount(0);
 expect(requests).toEqual([]);expect(errors).toEqual([]);
});
test('legitimate counted values remain visible when query limits prevent selection',async({page})=>{
 const requests=[],errors=[];page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/'))requests.push(r.url())});page.on('pageerror',e=>errors.push(e.message));await page.goto(base+'/ui-components');
 const facet=page.locator('#component-categorical-facet');await expect(facet).toHaveAttribute('data-state','ready');
 for(const value of ['x'.repeat(1025),'a\0b','😀'.repeat(257),'invalid-pattern']){
  await facet.evaluate((e,value)=>{const data=e.data;data.declaration.scope.signal=value==='invalid-pattern'?'logs':'spans';data.declaration.field=value==='invalid-pattern'?{kind:'column',key:'body_pattern_v1'}:{kind:'attribute',key:'http.route'};data.values.field=data.declaration.field;data.values.items=[{value,count:'1'}];data.values.total='1';e.data=data},value);
  await expect(facet).toHaveAttribute('data-state','ready');await expect(facet.locator('[data-facet-item]')).toHaveCount(1);await expect(facet.getByRole('button',{name:/^Include /})).toBeDisabled();await expect(facet.getByRole('button',{name:/^Exclude /})).toBeDisabled();expect(await facet.getByRole('button',{name:/^Include /}).getAttribute('title')).toBeTruthy();
 }
 expect(requests).toEqual([]);expect(errors).toEqual([]);
});

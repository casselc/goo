// Explicitly supplied synthetic primitive data, separate from the live OTLP journey.
const {test,expect}=require('@playwright/test');
const base=process.env.GOO_QUERY_URL||'http://127.0.0.1:8088';
test('209 generic multi-series timeline SSR and detached properties stay passive',async({page})=>{
 const reads=[],errors=[];page.on('request',r=>{if(new URL(r.url()).pathname.startsWith('/api/'))reads.push(r.url());});page.on('pageerror',e=>errors.push(e.message));await page.goto(base+'/ui-components');
 const handle=await page.evaluateHandle(()=>{const f=document.createElement('iframe');document.body.append(f);return f;}),frame=await handle.asElement().contentFrame();
 await frame.setContent('<h1>Supplied synthetic observation dots</h1><goo-value-timeline id="ssr" brush></goo-value-timeline><goo-value-timeline id="spa"></goo-value-timeline>');
 const data={fromNS:'1',toNS:'1000',domain:{min:'0',max:'1'},points:[{key:'a@10',series:'usage',fromNS:'10',toNS:'10',value:'0.1',state:'observed',label:'Supplied first role'},{key:'b@20',series:'request',fromNS:'20',toNS:'20',state:'unknown',label:'Supplied unknown point'},{key:'c@20',series:'limit',fromNS:'20',toNS:'20',value:'0.2',state:'stale',label:'Supplied stale point'}]};
 await frame.locator('#ssr').evaluate((e,data)=>e.setAttribute('data',JSON.stringify(data)),data);await frame.locator('#spa').evaluate((e,data)=>{e.data=data;},data);await frame.addStyleTag({url:base+'/value-timeline.css'});await frame.addScriptTag({type:'module',url:base+'/value-timeline.mjs'});
 for(const id of ['ssr','spa']){await expect(frame.locator('#'+id+' .value-timeline-mark')).toHaveCount(3);expect(await frame.locator('#'+id).evaluate(e=>e.data.domain)).toEqual(data.domain);}
 const unknown=frame.locator('#ssr .value-timeline-mark[data-key="b@20"]');await expect(unknown).toHaveText('?');expect(await unknown.getAttribute('data-interval')).toBe('false');expect(await unknown.evaluate(e=>e.getBoundingClientRect().width)).toBeLessThanOrEqual(18);
 await frame.locator('#ssr').evaluate(e=>{e.addEventListener('goo-timeline-select',event=>{e.selectedKey=event.detail.key;});});await unknown.focus();await frame.page().keyboard.press('Enter');await expect(unknown).toHaveAttribute('aria-pressed','true');
 await frame.locator('#spa').evaluate(e=>{const old=e.data;old.points[0].value='999';const parent=e.parentNode;e.remove();e.data={...e.data,points:[{key:'detached',series:'usage',fromNS:'300',toNS:'300',value:'0.3',state:'observed',label:'Detached supplied value'}]};parent.append(e);});await expect(frame.locator('#spa .value-timeline-mark')).toHaveCount(1);await expect(frame.locator('#spa .value-timeline-mark')).toHaveAttribute('data-value','0.3');
 await page.setViewportSize({width:390,height:844});expect(await frame.evaluate(()=>document.documentElement.scrollWidth<=innerWidth)).toBe(true);expect(reads).toEqual([]);expect(errors).toEqual([]);
});

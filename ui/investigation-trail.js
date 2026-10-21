/* Application-owned, in-memory view retention. No telemetry or durable evidence store. */
(function(global){
 function mount(host){
  const node=global.GooUI.node,bar=node('nav',undefined,'investigation-trail controls'),back=node('button','Back to question'),forward=node('button','Forward'),label=node('span'),body=node('div');bar.setAttribute('aria-label','Investigation trail');label.setAttribute('role','status');bar.append(back,forward,label);host.replaceChildren(bar,body);
  const entries=[];let index=-1,serial=0;
  function update(){host.dataset.questionPhase=entries[index]?.view.dataset.questionPhase||'initial';back.disabled=index<=0;forward.disabled=index>=entries.length-1;label.textContent='Visit '+(index+1)+' of '+entries.length+' · '+(entries[index]?.title||'New question');bar.title='Up to 8 views retained in this tab. Returning does not read telemetry; reload discards this trail. Capture evidence for durable storage.'}
  function leave(){const entry=entries[index];if(!entry)return;entry.scroll=global.scrollY;entry.hash=location.hash;entry.api.suspend();entry.scrolls=[...entry.view.querySelectorAll('*')].filter(element=>element.scrollTop||element.scrollLeft).map(element=>({element,top:element.scrollTop,left:element.scrollLeft}));entry.displays=[...entry.view.querySelectorAll('*')].filter(e=>e.localName.startsWith('goo-')).map(element=>({element,children:[...element.childNodes]}));entry.view.remove()}
  function select(next){if(next<0||next>=entries.length||next===index)return;leave();index=next;const entry=entries[index];body.append(entry.view);for(const {element,children}of entry.displays||[])element.replaceChildren(...children);global.dispatchEvent(new CustomEvent('goo:navigate-explore'));update();history.replaceState(null,'',entry.hash||'#explore');for(const {element,top,left}of entry.scrolls||[]){element.scrollTop=top;element.scrollLeft=left}global.scrollTo({top:entry.scroll||0,behavior:'instant'})}
  function create(q,fields=[],options={}){const replaceBlank=!!q&&entries[index]?.view.dataset.questionPhase==='initial';leave();if(replaceBlank){entries[index].api.dispose();entries.splice(index,1);index--}for(const entry of entries.splice(index+1))entry.api.dispose();const view=node('div'),entry={view,title:q?q.environment+' · '+q.signal:'New question',scroll:0,id:++serial};body.append(view);entries.push(entry);index=entries.length-1;if(entries.length>8){entries.shift().api.dispose();index--}entry.api=global.GooExploration.mount(view,{managed:true,initialBaseline:options.baseline,active:()=>entries[index]===entry&&view.isConnected});update();if(q)entry.api.openQuestion(q,fields,options);return entry.api}
  back.onclick=()=>select(index-1);forward.onclick=()=>select(index+1);
  const openQuestion=(q,fields=[],options={})=>create(q,fields,options);
  global.addEventListener('goo:open-population',e=>openQuestion(e.detail.question,[],{baseline:e.detail.baseline}));
  global.addEventListener('goo:open-question',e=>openQuestion(e.detail));
  document.addEventListener('click',event=>{const link=event.target.closest?.('a[href]');if(!link||event.button!==0||event.ctrlKey||event.metaKey||event.shiftKey||event.altKey||link.target)return;const url=new URL(link.href,location.href);if(url.origin!==location.origin||url.pathname!==location.pathname||!/^#(?:step|question)=/.test(url.hash))return;try{const value=global.GooWorkspace.readQuestionLink(url.hash);if(!value)return;event.preventDefault();openQuestion(value.question,[],{stage:true});history.pushState(null,'',url.hash)}catch{}});
  global.addEventListener('goo:question-state',event=>{update();host.dataset.questionPhase=event.detail.phase});
  create();
  return {openQuestion,openTrace:(...args)=>entries[index].api.openTrace(...args),run:q=>entries[index].api.run(q)};
 }
 global.GooInvestigationTrail={mount};
})(window);

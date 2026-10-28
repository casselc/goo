/* Small application-owned entity tab adapter. Switching tabs never reads sources. */

const global=globalThis;
function mount(tabs,{initial='services',onSelect=()=>{}}={}){let current=initial;const select=key=>{if(!tabs.some(t=>t.key===key))return;const previous=current;current=key;for(const tab of tabs){tab.panel.hidden=tab.key!==key;tab.button.setAttribute('aria-selected',String(tab.key===key));tab.button.tabIndex=tab.key===key?0:-1;if(tab.key===previous&&previous!==key)tab.leave?.()}onSelect(key)};for(const [index,tab]of tabs.entries()){tab.button.onclick=()=>select(tab.key);tab.button.onkeydown=e=>{if(!['ArrowLeft','ArrowRight','Home','End'].includes(e.key))return;e.preventDefault();const next=e.key==='Home'?0:e.key==='End'?tabs.length-1:(index+(e.key==='ArrowRight'?1:tabs.length-1))%tabs.length;select(tabs[next].key);tabs[next].button.focus()}}select(initial);return {select,current:()=>current}}

export {mount};

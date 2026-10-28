/* Internal request generation and listener lifetime mechanics. */
class RequestOwner{constructor(){this.generation=0;this.controller=null}cancel(){this.generation++;this.controller?.abort();this.controller=null}begin(){this.cancel();this.controller=new AbortController();return {generation:this.generation,controller:this.controller}}current(generation){return generation===this.generation}finish(generation){if(this.current(generation))this.controller=null}}
class PendingScope{
  constructor({events=[],observe=()=>()=>{},onRemove=()=>{}}={}){this.events=events.map(e=>({...e,listener:event=>e.callback(event)}));this.observe=observe;this.onRemove=onRemove;this.active=false}
  start(){this.stop();this.active=true;for(const e of this.events)e.target.addEventListener(e.type,e.listener,e.capture||false);this.disconnect=this.observe(()=>{if(!this.active)return;this.stop();this.onRemove()});return this}
  stop(){if(!this.active)return;this.active=false;for(const e of this.events)e.target.removeEventListener(e.type,e.listener,e.capture||false);this.disconnect?.();this.disconnect=null}
 }
export {RequestOwner,PendingScope};

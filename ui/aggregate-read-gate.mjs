/* Pure ownership gate shared by optional aggregate readers. No source or timers. */
class ReadGate{
  constructor(){this.scope=null;this.generation=0;this.loading=false;this.frozen=false}
  bind(scope){if(scope===this.scope)return false;this.scope=scope;this.generation++;this.loading=false;this.frozen=false;return true}
  start(){if(!this.scope||this.frozen||this.loading)return null;this.loading=true;return ++this.generation}
  accept(token,scope){if(token!==this.generation||!this.loading||this.frozen||scope!==this.scope)return false;this.loading=false;return true}
  pause(){this.generation++;this.loading=false;this.frozen=true}
  resume(){this.generation++;this.loading=false;this.frozen=false}
 }

export {ReadGate};

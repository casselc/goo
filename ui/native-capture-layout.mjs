/* Passive ordering of existing Saved DOM. Capture owners decide when data is
 * accepted; this adapter owns no reads, navigation, evidence or action leases. */
export function mount({capture,library,toolbar,recovery,scope}){
 let homes;
 function home(node){const marker=(node.ownerDocument||document).createElement('span');marker.hidden=true;node.before(marker);return marker}
 function initialize(){if(!capture.parentElement||library?.parentElement!==capture.parentElement||!scope?.parentElement)return false;homes={capture:home(capture),scope:home(scope),recovery:recovery?.parentElement?home(recovery):null};return true}
 function move(node,parent,before=null){
  if(!node||!parent||node===before||node.parentElement===parent&&node.nextSibling===before)return;
  // State-preserving moves keep connected custom-element plots and focus intact.
  if(typeof parent.moveBefore==='function')parent.moveBefore(node,before);else parent.insertBefore(node,before);
 }
 return {setNative(native){
  const value=String(!!native);if(capture.getAttribute('data-native-measurement')!==value)capture.setAttribute('data-native-measurement',value);
  if(!homes&&!initialize())return;
  if(native){const top=toolbar?.parentElement===library.parentElement?toolbar:library;move(capture,top.parentElement,top);move(scope,capture);if(recovery)move(recovery,capture,capture.firstChild)}
  else{move(capture,homes.capture.parentElement,homes.capture.nextSibling);move(scope,homes.scope.parentElement,homes.scope.nextSibling);if(homes.recovery)move(recovery,homes.recovery.parentElement,homes.recovery.nextSibling)}
 }};
}

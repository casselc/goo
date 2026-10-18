/* Scope repeated investigation components without sharing DOM identities. */
(function(global){
 function namespace(root,prefix){if(!prefix)return;for(const node of root.querySelectorAll('[id]'))node.id=prefix+node.id;for(const node of root.querySelectorAll('[for],[aria-controls],[aria-labelledby],[aria-describedby]'))for(const attr of ['for','aria-controls','aria-labelledby','aria-describedby'])if(node.hasAttribute(attr))node.setAttribute(attr,node.getAttribute(attr).split(/\s+/).map(id=>prefix+id).join(' '));}
 global.GooDOM={namespace};
})(globalThis);

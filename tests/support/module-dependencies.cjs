/* Native module mocks for the existing small fake-DOM owner fixtures.
 * Production modules never consult this fixture state or publish namespaces. */
const {registerHooks}=require('node:module');
const {mock}=require('node:test'),fs=require('node:fs'),path=require('node:path');
const mappings={};
for(const[name,file]of Object.entries({UI:'components',Visuals:'visualizations',Queries:'queries',PodMetrics:'pod-metrics',ResourceScope:'resource-scope',ResourceNativeView:'resource-native-view',HostDashboard:'host-dashboard',Attributes:'attributes',EvidenceSummary:'evidence-summary',ComparisonVisual:'comparison-visual',Histograms:'histograms',Measurements:'measurements',Workspace:'workspace',NumericDistributionComparison:'numeric-distribution-comparison',NumericMeasures:'numeric-measures'}))mappings['Goo'+name]=path.resolve(__dirname,'../../ui/'+file+'.mjs');
let loadingOriginal=false;
registerHooks({resolve(specifier,context,next){const result=next(specifier,context);if((loadingOriginal||context.parentURL?.endsWith('?goo-fixture-original'))&&result.url.startsWith('file://'+path.resolve(__dirname,'../../ui')+'/'))return {...result,url:result.url.split('?')[0]+'?goo-fixture-original'};return result}});
const registered=new Set(),seen=new Set();
function keys(source){const names=new Set();for(const m of source.matchAll(/export\s*\{([^}]+)\}/g))for(const field of m[1].split(',')){const name=field.trim().split(/\s+as\s+/).at(-1);if(name)names.add(name)}for(const m of source.matchAll(/export\s+(?:async\s+)?(?:function|class|const)\s+(\w+)/g))names.add(m[1]);return [...names]}
function install(filename){
 const names=new Set();
 function collect(file){file=path.resolve(file);if(seen.has(file))return;seen.add(file);const source=fs.readFileSync(file,'utf8');for(const m of source.matchAll(/global\.(Goo\w+)\s*=/g))if(mappings[m[1]]&&!registered.has(m[1]))names.add(m[1]);for(const m of source.matchAll(/require\(['"](\.[^'"]+)['"]\)/g)){let child=path.resolve(path.dirname(file),m[1]);if(!path.extname(child))child+='.js';if(child.startsWith(__dirname+path.sep)&&child!==__filename&&fs.existsSync(child))collect(child)}}
 collect(filename);
 // Real fallbacks use a separate native URL graph. Load them before installing
 // mocks so import cycles cannot bind a production owner to a fixture namespace.
 const originals={};loadingOriginal=true;try{for(const name of names)originals[name]=require(mappings[name])}finally{loadingOriginal=false}
 for(const name of names){
  const url=mappings[name],actual=originals[name],exports={};
  for(const key of keys(fs.readFileSync(url,'utf8'))){let wrapper;
   const resolve=()=>{const value=global[name]?.[key];return value===wrapper||value===undefined?actual[key]:value};
   wrapper=new Proxy(function(...args){return Reflect.apply(resolve(),this,args)},{construct(_target,args){const ctor=resolve();return Reflect.construct(ctor,args)}});exports[key]=wrapper;
  }
  mock.module(url,{exports});registered.add(name);
 }
}
module.exports={install};

/* Generated wrapper characterization, not a browser or source simulator.
   Execute the production producer/root closures and real question Session;
   the actual journey separately covers mounted population renewal and Save. */
import fs from 'node:fs';
import assert from 'node:assert/strict';
import vm from 'node:vm';
import * as Distributions from '../ui/distributions.mjs';
import * as QuestionSession from '../ui/question-session.mjs';
import * as Composition from '../ui/view-composition.mjs';
import {stagedView} from '../ui/view-template-handoff.mjs';
import {adoptExploration} from '../ui/source-snapshots.mjs';
const producer=fs.readFileSync(new URL('../ui/distributions.mjs',import.meta.url),'utf8'),root=fs.readFileSync(new URL('../ui/exploration.mjs',import.meta.url),'utf8');
const parent=fs.readFileSync(new URL('../ui/population-workspace.mjs',import.meta.url),'utf8');
const closure=(text,name)=>{const line=text.split('\n').find(line=>line.trimStart().startsWith('function '+name+'('));assert.ok(line,'production closure '+name+' missing');return line.trim()};
const options=root.split('\n').find(line=>line.includes('onCustomizeDistribution:(view,origin,opener)=>'));
const start=options.indexOf('onCustomizeDistribution:')+'onCustomizeDistribution:'.length,callback=options.slice(start,options.indexOf(',stagePopulation:',start));
function element(tag,text){return {localName:tag,textContent:text,disabled:false,hidden:false,inert:false,fieldset:false,isConnected:true,attributes:{},dataset:{},append(...children){(this.children??=[]).push(...children)},matches(){return this.fieldset},getAttribute(key){return this.attributes[key]??null},closest(){return this.inert?this:null}}}
function rawData(input,grouped){const data=structuredClone(input);if(grouped)return data;const q=data.question;delete q.groupTimeSeries;delete data.groupedTimeDataset;delete data.groupedTimeSeries;data.viewDataset={version:'goo.dataset/1',id:'ranked396',revision:data.basis,origin:'source-read',scope:{id:'ranked396scope',generation:data.basis,signal:q.signal,selector:structuredClone(q)},shape:'ranked',dimensions:['category'],fields:[{key:'category',label:'region',type:'string',nullable:true},{key:'count',label:'Captured records',type:'number',unit:'records'}],rows:data.groups.map((group,index)=>({key:'group:'+index,values:{category:group.g0,count:group.count}})),quality:{extent:'complete',basis:'captured-query',value:'derived',sampling:'unknown',freshness:'fixed'},completeness:{basis:'captured-query',revision:data.basis,scopeId:'ranked396scope',aspects:['groups']},partition:{kind:'disjoint-record-groups',measure:'count',total:data.count},capturedAt:data.capturedAt};return data}
globalThis.fetch=()=>{throw Error('Source-free staging attempted IO')};
let total=0,accepted=0,rejected=0,stale=0;
for(const input of JSON.parse(fs.readFileSync(0,'utf8'))){
 const original=rawData(input.data,input.grouped),before=JSON.stringify(original),events=[],classes=new Set(),buttons=[],session=new QuestionSession.Session();let data=adoptExploration(original,original.question),token=session.start(data.question);session.publish(token,data);const tokens=[token];
 const env={localQuery:'雪',readOnly:false,customizeOpener:null,current:data,lastResult:data,distributionView:input.grouped?null:{presentation:()=>({kind:'table',value:'count',category:'category',query:env.localQuery,order:'label',scale:'population',requiresComplete:false})},timelineView:input.grouped?{presentation:()=>({version:'goo.view/1',scope:{kind:'record-query',signal:data.question.signal},kind:{name:'table',version:1},display:{profile:'grouped-count-v1',value:'count',query:env.localQuery,order:'label'}})}:null,distributionCustomization:Distributions.distributionCustomization,Distributions,QuestionSession,stagedView,node:element,session,active:()=>env.activeFlag,activeFlag:true,filterError:null,container:{isConnected:true,hidden:false,classList:{contains:name=>classes.has(name)}},UI:{actionVisible:node=>node.isConnected&&!node.hidden},structuredClone,CustomEvent:class{constructor(type,options){this.type=type;this.detail=options.detail}},global:{dispatchEvent:event=>{events.push(event);classes.add('population-open')}}};
 vm.createContext(env);env.canCustomizeDistribution=vm.runInContext(closure(root,'plotReady')+'\n'+closure(root,'canCustomizeDistribution')+'\ncanCustomizeDistribution',env);env.onCustomizeDistribution=vm.runInContext('('+callback+')',env);vm.runInContext(closure(producer,'refreshCustomization')+'\n'+closure(producer,'mountCustomization'),env);
 // Exercise the actual local publication guard independently of acquisition epochs.
 const target=element('section'),other=element('section'),key='current';
 const local={views:new Map([[key,{data,host:target}]]),results:[{result:data}],queries:[data.question],active:()=>local.enabled,enabled:true,panel:{hidden:false},childOwner:null,busy:false,composition:{children:[{timeline:{},hidden:false}]},QuestionSession,Composition,actionVisible:node=>node.isConnected&&!node.hidden};
 vm.createContext(local);const localCurrent=vm.runInContext(closure(parent,'localPresentationCurrent')+'\nlocalPresentationCurrent',local);
 const resetLocal=()=>{local.views.set(key,{data,host:target});local.results=[{result:data}];local.queries=[data.question];local.enabled=true;local.panel.hidden=false;local.childOwner=null;local.busy=false;local.composition.children[0].hidden=false;target.isConnected=true;target.hidden=false;target.inert=false};
 resetLocal();assert.equal(localCurrent(key,data,target,'timeline'),true);
 for(const code of input.schedule){resetLocal();switch(code%11){case 0:local.views.set(key,{data,host:other});break;case 1:local.views.set(key,{data:{...data},host:target});break;case 2:local.results=[{result:data,error:'failed replacement'}];break;case 3:local.queries=[{...data.question,limit:data.question.limit+1}];break;case 4:local.enabled=false;break;case 5:local.panel.hidden=true;break;case 6:local.childOwner={};break;case 7:local.busy=true;break;case 8:target.isConnected=false;break;case 9:target.inert=true;break;case 10:local.composition.children[0].hidden=true;break;}assert.equal(localCurrent(key,data,target,'timeline'),false,'retired/hidden local presentation admitted publication');}
 resetLocal();assert.equal(localCurrent(key,data,target,'timeline'),true,'current held presentation did not recover');
 const mount=()=>{const host=element('div');env.mountCustomization(host,data);if(host.children?.[0]){buttons.push(host.children[0]);return host.children[0]}return null};mount();
 function attempt(old=false){const button=old?buttons[0]:buttons.at(-1);const prior=events.length,blocked=env.readOnly||!env.activeFlag||env.filterError||classes.has('population-open')||classes.has('recent-mode')||!env.container.isConnected||env.container.hidden||session.snapshot().phase!=='applied'||button!==env.customizeOpener||button.disabled||button.hidden||button.inert||button.fieldset||button.attributes['aria-disabled']==='true';button.onclick();total++;const delta=events.length-prior;if(blocked)assert.equal(delta,0,'retired/suspended producer emitted configuration');assert.ok(delta===0||delta===1);if(delta){accepted++;const event=events.at(-1);assert.equal(event.detail.stage,true);assert.deepEqual(event.detail.question,data.question);assert.deepEqual(event.detail.view.queries,[data.question]);assert.equal(Object.hasOwn(event.detail.view,'basis'),false);assert.equal(Object.hasOwn(event.detail.view,'viewDataset'),false);}else rejected++;}
 // Positive controls ensure the randomized schedule is not only refusals.
 attempt();classes.delete('population-open');attempt();classes.delete('population-open');
 const retired=token;token=session.start(data.question);tokens.push(token);const pending=session.snapshot();assert.equal(session.publish(retired,data),false);assert.deepEqual(session.snapshot(),pending);stale++;session.fail(token);attempt();token=session.start(data.question);tokens.push(token);session.publish(token,data);mount();
 for(const code of input.schedule){const action=code%18;switch(action){
 case 0:attempt(code%2===0);break;
 case 1:classes.delete('population-open');break;
 case 2:token=session.start(data.question);tokens.push(token);break;
 case 3:session.fail(token);break;
 case 4:{const old=tokens[0],state=session.snapshot();if(old!==state.generation||state.phase!=='running'){assert.equal(session.publish(old,data),false);assert.deepEqual(session.snapshot(),state);stale++}break;}
 case 5:{const raw=structuredClone(original),next=adoptExploration(raw,raw.question);if(session.publish(token,next)){data=next;env.current=data;env.lastResult=data;mount()}break;}
 case 6:buttons.at(-1).hidden=!buttons.at(-1).hidden;break;
 case 7:buttons.at(-1).inert=!buttons.at(-1).inert;break;
 case 8:env.readOnly=!env.readOnly;break;
 case 9:mount();break;
 case 10:env.activeFlag=!env.activeFlag;break;
 case 11:env.filterError=env.filterError?null:'invalid draft';break;
 case 12:buttons.at(-1).disabled=!buttons.at(-1).disabled;break;
 case 13:buttons.at(-1).fieldset=!buttons.at(-1).fieldset;break;
 case 14:env.container.hidden=!env.container.hidden;break;
 case 15:classes.has('recent-mode')?classes.delete('recent-mode'):classes.add('recent-mode');break;
 case 16:env.localQuery='雪'.repeat(86);env.refreshCustomization(data);assert.equal(env.customizeOpener.disabled,true,'invalid live Find left Customize enabled');break;
 case 17:env.localQuery='雪';env.refreshCustomization(data);break;
 }attempt(action%2===0)}
 env.localQuery='雪';env.readOnly=false;env.activeFlag=true;env.filterError=null;env.container.hidden=false;env.container.isConnected=true;classes.clear();token=session.start(data.question);session.publish(token,data);mount();const final=events.length;attempt();assert.equal(events.length,final+1,'current accepted producer did not recover safe staging');assert.equal(JSON.stringify(original),before,'wrappers mutated supplied source evidence');
}
process.stdout.write(JSON.stringify({attempts:total,accepted,rejected,stale,complete:true}));

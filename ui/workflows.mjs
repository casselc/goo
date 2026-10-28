/* Question-led recipes. These invoke validated query templates, never SQL or
 * executable author text. Coverage describes this demo, not inferred health. */

const global=globalThis;
const recipes=[
  {id:'delay',name:'Why is completion delayed?',intro:'Start with business outcomes, then test whether the synchronous request or later asynchronous work explains the delay.',
   coverage:'Business outcomes are demo previews. Fast acceptance is not proof of completion; missing telemetry remains ambiguous.',
   steps:[['Compare outcomes','completion','Group by scenario and select a reservation from the table.','scenario'],
          ['Inspect acceptance latency','heatmap','Look for a duration band, then select a cell and open its reservation waterfall.'],
          ['Follow asynchronous work','traces','Select a consumer span and follow its producer link. Links are not parent edges.'],
          ['Check lifecycle events','logs','Look for accepted, retry, stalled and terminal records for the selected reservation.']]},
  {id:'retry',name:'Did a retry recover?',intro:'Distinguish a failed attempt from a failed business operation.',scenario:'retry',
   coverage:'An error span can belong to a reservation that ultimately succeeds. Attempt counts and subject outcomes have different denominators.',
   steps:[['Find retry outcomes','completion','Select a retry reservation after it matures.','scenario'],
          ['Inspect attempt spans','traces','Compare the error attempt with later work; select spans to see their status.'],
          ['Confirm retry and terminal evidence','logs','Check that retry and terminal events refer to the same reservation.']]},
  {id:'red',name:'Is the service slow? (RED)',intro:'RED asks about rate, errors and duration for a service. Start with observed request volume and duration, then inspect error evidence.',
   coverage:'Available: raw server-span counts per minute and duration bands. Missing: a complete service error-rate measure. Exported counters are point values, not reset-aware rates; sampling/completeness are unverified.',
   steps:[['Explore volume and duration','heatmap','Each cell counts raw gRPC server spans. This is not a fleet p95 or a guaranteed traffic rate.'],
          ['Inspect error evidence','traces','Select spans to inspect Error status; a bounded span list cannot establish the service error rate.'],
          ['Inspect exported counters','metrics','Service scope: this clears reservation/scenario filters. Do not sum cumulative points.']]},
  {id:'use',name:'Could resource pressure explain it? (USE)',intro:'USE checks utilization, saturation and errors for each resource. Start by identifying which measurements actually exist.',
   coverage:'Available: Redis group lag/pending, observed queue wait, active workers and explicit dispatch capacity. CPU, memory and database pool utilization remain NOT INSTRUMENTED. Pending includes active work and stale entries; lag is undelivered work. Missing or stale observations are unknown, not zero usage.',
   steps:[['Inspect queue and worker evidence','queue','Resource scope: clears reservation/scenario filters. Inspect lag, pending, worker utilization and observed queue wait together.'],
          ['Inspect work and error evidence','traces','Consumer duration and Error status can guide a hypothesis; they do not measure queue depth or prove saturation.'],
          ['Check supporting events','logs','Look for retries/stalls. Correlate waiting time with processing evidence; this does not prove causation.']]}
 ];
function mount(container,run,current){
  const label=document.createElement('label'); label.textContent='Investigation question';
  const select=document.createElement('select'); select.id='workflow';
  const empty=document.createElement('option');empty.value='';empty.textContent='Choose a guided investigation…';select.append(empty);
  for(const r of recipes){const o=document.createElement('option');o.value=r.id;o.textContent=r.name;select.append(o)}
  label.append(select);container.append(label);
  const body=document.createElement('div');body.id='workflowBody';container.append(body);
  let active;
  function update(expression){
   if(!active)return;
   for(const b of body.querySelectorAll('button[data-view]')){
    if(b.dataset.view===expression.view)b.setAttribute('aria-current','step'); else b.removeAttribute('aria-current');
   }
   const context=body.querySelector('.workflow-context');
   context.textContent=expression.reservationId?'Following reservation '+expression.reservationId+'. Query steps retain this selection where supported.':'Population scope. Select a reservation in the results or from a heatmap span to narrow subsequent steps.';
  }
  select.onchange=()=>{
   active=recipes.find(r=>r.id===select.value);body.replaceChildren();if(!active)return;
   const intro=document.createElement('p');intro.textContent=active.intro;body.append(intro);
   const coverage=document.createElement('p');coverage.id='workflowCoverage';coverage.className='warning';coverage.textContent=active.coverage;body.append(coverage);
   const steps=document.createElement('div');steps.className='workflow-steps';
   active.steps.forEach(([title,view,reason,group],index)=>{
    const card=document.createElement('div'),button=document.createElement('button');
    button.textContent=`${index+1}. ${title}`;button.dataset.view=view;
    const why=document.createElement('p');why.textContent=reason;
    button.onclick=()=>{
     const context=current();
     const spec={...context,view,groupBy:group||context.groupBy};delete spec.spanFilter;
     if(active.scenario){if(context.scenario!==active.scenario)spec.reservationId='';spec.scenario=active.scenario;}
     if(view==='metrics'||view==='queue'){spec.reservationId='';spec.scenario='';}
     run(spec);
    };
    card.append(button,why);steps.append(card);
   });
   const context=document.createElement('p');context.className='workflow-context';body.append(steps,context);update(current());
  };
  return {update};
 }

export {recipes,mount};

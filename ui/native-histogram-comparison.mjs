/* App composition over a held native pair. No source reads or range actions. */
import './aligned-histogram.mjs';
import {asHistogramComparison} from './snapshots.mjs';
import {buildNativeHistogramComparisonFromEvidence} from './native-histogram-view.mjs';
import {node} from './components.mjs';

// The capture comparison's global decision remains authoritative. A locally
// complete stream must not bypass another member's withheld comparison.
export function comparisonForWindow(result,window){
 if(result?.state!=='comparable'||window?.state!=='comparable'||!result.nativeWindows?.includes(window))return null;
 try{return buildNativeHistogramComparisonFromEvidence(result,window)}catch{return null}
}

// Build a retained stream window only when opened. Reopening keeps its local
// display choices and avoids repeating source-membership scans or DOM work.
export function renderWindow(detail,build){
 let built=false;
 const render=()=>{if(!detail.open||built)return;built=true;build()};
 detail.addEventListener('toggle',render);render();
}

export function mount(host,comparison){
 if(!comparison)return null;
 let pair;try{pair=asHistogramComparison(comparison)}catch{return null}
 if(pair.baseline.histogram.kind!=='otel-explicit'||pair.selection.histogram.kind!=='otel-explicit')return null;
 const section=node('section',undefined,'native-histogram-comparison'),label=node('label','Show '),picker=node('select'),chart=node('goo-aligned-histogram');
 picker.setAttribute('aria-label','Native distribution scale');
 for(const [value,text] of [['count','Observation counts'],['observation-share','Share of each observation population']]){const option=node('option',text);option.value=value;picker.append(option)}
 picker.value='count';label.className='native-time-view';label.append(picker);chart.comparison=pair;
 const draw=()=>{chart.display={scale:picker.value}};
 picker.onchange=draw;draw();section.append(label,chart);host.append(section);
 return {element:section,chart,comparison:pair};
}

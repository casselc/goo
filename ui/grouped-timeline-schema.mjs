/* Finite supplied grouped geometry. Exact values remain the scope keys. */
export const groupedMeanQuery=q=>q?.signal==='spans'&&q.summary==='span-duration';
export function groupedQuery(q){return !!q&&['logs','spans'].includes(q.signal)&&q.groupTimeSeries===true&&q.groupBy?.length===1&&!q.numeric&&!q.numericHistogram&&(!q.summary||groupedMeanQuery(q))&&!q.spanStatus&&!q.reportedEdge&&!(q.groupBy[0].kind==='column'&&['body_pattern_v1','duration_bucket'].includes(q.groupBy[0].key));}
export function validateGrouped(s,fields,reasons){
 const fail=text=>reasons.push(text),q=s.scope?.selector,i=s.interval,mean=groupedMeanQuery(q);
 if(!groupedQuery(q)||s.dimensions?.length!==1||s.dimensions[0]!=='category'||fields.size!==(mean?6:4)||fields.get('category')?.type!=='string'||fields.get('category')?.nullable!==true||fields.get('count')?.type!=='number'||fields.get('count')?.unit!=='records'||fields.get('count')?.nullable!==true||fields.get('count')?.aggregation||s.partition!==undefined)fail('Grouped timelines require the pinned query and typed category/count fields');
 if(mean)for(const [key,kind,method]of [['duration.sum','sum','exact'],['duration.mean','mean','exact-rational']]){const field=fields.get(key),a=field?.aggregation;if(field?.type!=='rational'||field.unit!=='ns'||field.nullable!==true||!a||Object.keys(a).length!==3||a.family!=='duration'||a.kind!==kind||a.method!==method)fail('Grouped mean requires exact duration sum and rational mean fields');}
 for(const key of ['fromMs','toMs'])if(fields.get(key)?.type!=='timestamp'||fields.get(key)?.unit!=='ms')fail('Grouped interval bounds require millisecond timestamps');
 if(!i||i.fromMs!==String(q?.fromMs)||i.toMs!==String(q?.toMs)||i.bucketMs!==String(q?.bucketMs)||![1000,10000,60000,300000].includes(q?.bucketMs)||!Number.isSafeInteger(q?.fromMs)||!Number.isSafeInteger(q?.toMs)||q.toMs<=q.fromMs||Math.ceil((q.toMs-q.fromMs)/q.bucketMs)>240){fail('Grouped timeline geometry differs from its fixed question');return}
 if(!Array.isArray(s.rows))return;
 if(s.rows.length>1920)fail('Grouped timeline exceeds 1920 derived cells');
 const groups=new Map(),tuples=new Set(),bins=Math.ceil((q.toMs-q.fromMs)/q.bucketMs);let positive=0;
 for(const row of s.rows){const v=row?.values;if(!v||!(v.category===null||typeof v.category==='string')||!/^group-interval:[a-f0-9]{64}$/.test(row.key||'')){fail('Grouped cells require canonical typed group/interval identities');continue}
  const key=JSON.stringify(v.category);if(!groups.has(key))groups.set(key,[]);groups.get(key).push(row);const from=Number(v.fromMs),to=Number(v.toMs),tuple=JSON.stringify([v.category,v.fromMs,v.toMs]);
  if(tuples.has(tuple))fail('Duplicate grouped interval');tuples.add(tuple);
  if(!Number.isSafeInteger(from)||!Number.isSafeInteger(to)||String(from)!==v.fromMs||String(to)!==v.toMs||from<q.fromMs||from>=q.toMs||(from-q.fromMs)%q.bucketMs||to!==Math.min(from+q.bucketMs,q.toMs))fail('Grouped intervals must follow the fixed half-open grid');
  if(row.state==='unknown'){if(v.count!==null)fail('Unavailable grouped counts must be null')}
  else if(row.state==='observed'){if(typeof v.count!=='string'||!/^[1-9][0-9]{0,19}$/.test(v.count)||BigInt(v.count)>18446744073709551615n)fail('Observed grouped cells require a positive count');else positive++}
  else if(row.state==='no_data'){if(v.count!=='0'||s.quality?.extent!=='complete')fail('Grouped zero requires coherent exhaustive coverage')}
  else fail('Invalid grouped cell state');
  if(mean){
   const sum=v['duration.sum'],value=v['duration.mean'],sumCell=row.cells?.['duration.sum'],meanCell=row.cells?.['duration.mean'];
   if(row.state==='observed'){
    const parts=typeof value==='string'&&/^(0|[1-9]\d*)(?:\/([1-9]\d*))?$/.exec(value);
    if(typeof sum!=='string'||!/^(0|[1-9]\d*)$/.test(sum)||!parts||typeof v.count!=='string'||!/^([1-9]\d*)$/.test(v.count)||BigInt(parts[1])*BigInt(v.count)!==BigInt(sum)*BigInt(parts[2]||'1'))fail('Grouped duration mean must equal exact sum divided by its captured count');
    for(const cell of [sumCell,meanCell])if(cell?.state!=='derived'||cell.support?.total!==v.count||cell.support?.valid!==v.count||['missing','empty','invalid'].some(key=>cell.support?.[key]!=='0'))fail('Grouped raw-span duration requires exact contributing-count support');
   }else{
    if(sum!==null||value!==null||sumCell?.state!==row.state||meanCell?.state!==row.state)fail('Unavailable grouped duration must be null with its explicit cell state');
    if(row.state==='no_data')for(const cell of [sumCell,meanCell])if(['total','valid','missing','empty','invalid'].some(key=>cell.support?.[key]!=='0'))fail('Empty grouped duration has zero contributing support');
   }
  }
 }
 if(groups.size>8||positive>1000)fail('Grouped source limits exceed eight categories or 1000 observed cells');
 for(const rows of groups.values()){if(rows.length!==bins)fail('Every grouped category requires the complete fixed interval grid');if(!rows.some(row=>row.state==='observed'))fail('All-zero or unavailable invented categories are refused')}
}

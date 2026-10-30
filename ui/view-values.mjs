/* IO-free cell presentation. Exact source values stay separate from rounded display. */
export function categoryText(value){
  return value===undefined?'Not supplied':value===null?'Missing':value===''?'Present empty string':typeof value==='string'?JSON.stringify(value):String(value);
}

function fraction(value){
  if(typeof value!=='string'&&typeof value!=='number')return null;
  const text=String(value),r=/^(-?\d+)(?:\/(\d+))?$/.exec(text);
  if(r){const d=BigInt(r[2]||'1');return d>0n?{n:BigInt(r[1]),d}:null}
  const decimal=/^(-?)(\d+)\.(\d+)$/.exec(text);
  if(decimal)return {n:BigInt(decimal[2]+decimal[3])*(decimal[1]?-1n:1n),d:10n**BigInt(decimal[3].length)};
  const exponential=/^(-?)(\d+)(?:\.(\d+))?[eE]([+-]?\d+)$/.exec(text);if(!exponential)return null;
  const shift=Number(exponential[4])-(exponential[3]||'').length;if(!Number.isSafeInteger(shift)||Math.abs(shift)>1024)return null;
  const numerator=BigInt(exponential[2]+(exponential[3]||''))*(exponential[1]?-1n:1n);
  return shift<0?{n:numerator,d:10n**BigInt(-shift)}:{n:numerator*10n**BigInt(shift),d:1n};
}

function rounded(n,d){
  const negative=n<0n;if(negative)n=-n;
  const quotient=n*1000n/d,remainder=n*1000n%d,scaled=quotient+(remainder*2n>=d?1n:0n);
  if(scaled===0n&&n!==0n)return (negative?'−':'')+'<0.001';
  const digits=String(scaled%1000n).padStart(3,'0').replace(/0+$/,'');
  return (remainder?'≈ ':'')+(negative?'−':'')+String(scaled/1000n)+(digits?'.'+digits:'');
}

export const isFraction=field=>field?.aggregation?.family==='ratio'&&field.aggregation.kind==='fraction';
export function fractionContextText(facts){
  if(!facts)return '';
  const label=fact=>fact.label+': '+(fact.value===null?'Unknown':fact.value);
  return [label(facts.numerator)+' / '+label(facts.denominator),...facts.partition.filter(fact=>fact.field!==facts.numerator.field&&fact.field!==facts.denominator.field).map(label)].join(' · ');
}
function percent(n,d){
  if(n<0n||n>d)return 'Unavailable';
  const scaled=(n*10000n+d/2n)/d;
  if(n===0n)return '0%';if(n===d)return '100%';
  if(scaled===0n)return '<0.01%';if(scaled>=10000n)return '>99.99%';
  const digits=String(scaled%100n).padStart(2,'0').replace(/0+$/,'');
  return (n*10000n%d?'≈ ':'')+String(scaled/100n)+(digits?'.'+digits:'')+'%';
}

export function supportText(cell){
  const s=cell?.support;if(!s)return '';
  return `${s.valid}/${s.total} valid · missing ${s.missing} · empty ${s.empty} · invalid ${s.invalid}`;
}

export function formatCell(field,value,cell){
  if(cell?.state==='no_data')return 'No data';
  if(cell?.state==='no_valid_values')return 'No valid values';
  if(cell?.state==='unknown')return 'Unknown';
  if(value===null)return 'Missing';if(value===undefined)return 'Not supplied';if(value==='')return 'Present empty string';
  if(field?.type==='timestamp'){const time=new Date(Number(value));return Number.isFinite(time.getTime())?time.toISOString():'Unavailable time'}
  if(field?.type!=='number'&&field?.type!=='rational')return field?.type==='string'?categoryText(value):String(value);
  const rational=fraction(value);if(!rational)return 'Unavailable';
  let {n,d}=rational,unit=field.unit||'';
  if(isFraction(field))return percent(n,d);
  if(unit==='ns'){
    const magnitude=n<0n?-n:n;
    if(magnitude>=1000000000n*d){d*=1000000000n;unit='s'}
    else if(magnitude>=1000000n*d){d*=1000000n;unit='ms'}
    else {d*=1000n;unit='µs'}
  }
  return rounded(n,d)+(unit?' '+unit:'');
}

export function exactText(field,value,cell,facts){
  const parts=[value===null||value===undefined?'No supplied value':String(value)+(field?.unit?' '+field.unit:'')];
  if(field?.type==='timestamp')parts[0]+=' since Unix epoch';
  if(field?.aggregation){const a=field.aggregation;parts.push(a.family+' '+a.kind,a.method+(a.quantile!==undefined?' · q='+a.quantile:''))}
  if(isFraction(field)){parts.push(field.aggregation.numerator+' / '+field.aggregation.denominator+' · bounded fraction');const context=fractionContextText(facts);if(context)parts.push(context)}
  if(cell?.state)parts.push(cell.state);
  const support=supportText(cell);if(support)parts.push(support);
  return parts.join(' · ');
}

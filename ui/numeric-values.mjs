/* Exact numeric presentation; no owners, source clients or declarations. */
 function scaleFactor(scale){if(scale===undefined||scale===0)return 1n;if(!Number.isInteger(scale)||scale<1||scale>9)throw Error('Invalid decimal scale.');return 10n**BigInt(scale)}
 function coefficient(raw){if(typeof raw!=='string'||raw.length>256||/\s/u.test(raw)||! /^(0|[1-9]\d*|-[1-9]\d*)$/.test(raw))throw Error('Invalid exact coefficient.');return BigInt(raw)}
 function coefficientText(raw,scale){const factor=scaleFactor(scale),v=coefficient(raw);if(factor===1n)return raw;const n=v<0n?-v:v;return (v<0n?'-':'')+String(n/factor)+'.'+String(n%factor).padStart(scale,'0')}
 function rationalText(raw,scale){if(typeof raw!=='string'||raw.length>256||/\s/u.test(raw)||! /^-?(0|[1-9]\d*)(?:\/[1-9]\d*)?$/.test(raw))throw Error('Invalid exact numeric value.');if(/^-0(?:\/|$)/.test(raw))throw Error('Invalid exact numeric value.');const [n,d='1']=raw.split('/');let a=BigInt(n),b=BigInt(d)*scaleFactor(scale),x=a<0n?-a:a,y=b;while(y){const t=x%y;x=y;y=t}if(x){a/=x;b/=x}return b===1n?String(a):a+'/'+b}
 function rationalDisplay(raw){const reduced=rationalText(raw,0),[numerator,denominator='1']=reduced.split('/');if(denominator==='1')return numerator;let d=BigInt(denominator),twos=0,fives=0;while(d%2n===0n&&twos<=256){d/=2n;twos++}while(d%5n===0n&&fives<=256){d/=5n;fives++}const digits=Math.max(twos,fives);if(d!==1n||digits>256)return reduced;const n=BigInt(numerator),absolute=n<0n?-n:n,scaled=absolute*(2n**BigInt(digits-twos))*(5n**BigInt(digits-fives)),factor=10n**BigInt(digits),fraction=String(scaled%factor).padStart(digits,'0').replace(/0+$/,'');return (n<0n?'-':'')+String(scaled/factor)+(fraction?'.'+fraction:'')}
 function value(summary,metric='mean') {if(!summary||summary.valid==='0')return null;if(typeof summary.valid!=='string'||summary.valid.length>20||/\s/u.test(summary.valid)||!/^[1-9]\d*$/.test(summary.valid))throw Error('Invalid valid-record count.');return {n:coefficient(summary[metric==='mean'?'sum':metric]),d:(metric==='mean'?BigInt(summary.valid):1n)*scaleFactor(summary.scale)}}
 function format(summary,metric='mean') {const v=value(summary,metric);if(!v)return 'Unavailable';if(metric!=='mean')return coefficientText(summary[metric],summary.scale)+' '+summary.unit;const digits=Math.max(3,summary.scale||0),factor=10n**BigInt(digits),negative=v.n<0n,n=negative?-v.n:v.n,rounded=(n*factor+v.d/2n)/v.d;return (negative&&rounded!==0n?'−':'')+String(rounded/factor)+'.'+String(rounded%factor).padStart(digits,'0')+' '+summary.unit}
 function magnitude(v){return v?{n:v.n<0n?-v.n:v.n,d:v.d}:null}
 function coverage(s){return s.valid+' valid · '+s.missing+' missing · '+s.empty+' empty · '+s.invalid+' invalid'}
 // Presentation only: keep the original exact value for disclosure and never
 // round a nonzero scalar into a displayed zero. Prefixes scale numbers, not units.
 function compactScalar(raw){
  if(typeof raw==='number'){if(!Number.isFinite(raw))return null;raw=String(raw)}
  if(typeof raw!=='string'||raw.length>512)return null;
  let n,d;const rational=raw.match(/^(-?\d+)\/([1-9]\d*)$/),decimal=raw.match(/^(-?)(\d+)(?:\.(\d+))?(?:e([+-]?\d+))?$/i);
  if(rational){n=BigInt(rational[1]);d=BigInt(rational[2])}else if(decimal){const exponent=Number(decimal[4]||0)-(decimal[3]||'').length;if(Math.abs(exponent)>600)return null;n=BigInt(decimal[2]+(decimal[3]||''))*(decimal[1]?-1n:1n)*(exponent>0?10n**BigInt(exponent):1n);d=exponent<0?10n**BigInt(-exponent):1n}else return null;
  if(n===0n)return {text:'0',exact:raw,rounded:false};
  const negative=n<0n,a=negative?-n:n;let exponent=String(a).length-String(d).length;
  if(exponent>=0?a<d*10n**BigInt(exponent):a*10n**BigInt(-exponent)<d)exponent--;
  const places=2-exponent,factor=10n**BigInt(Math.abs(places)),numerator=places>=0?a*factor:a,denominator=places>=0?d:d*factor;
  let rounded=(numerator+denominator/2n)/denominator;if(rounded>=1000n){rounded/=10n;exponent++}
  const shift=exponent-2,exactN=shift>=0?rounded*10n**BigInt(shift):rounded,exactD=shift>=0?1n:10n**BigInt(-shift),approximate=exactN*d!==a*exactD;
  const groups=['','k','M','B','T'],group=exponent>=3&&exponent<15?Math.floor(exponent/3):0;
  let text;
  if(exponent<-3||exponent>=15){const digits=String(rounded).padStart(3,'0');text=digits[0]+(digits.slice(1).replace(/0+$/,'')?'.'+digits.slice(1).replace(/0+$/,''):'')+'e'+(exponent>=0?'+':'')+exponent}
  else{const decimalPlaces=2-exponent+group*3,divisor=10n**BigInt(Math.max(0,decimalPlaces));text=decimalPlaces>0?String(rounded/divisor)+'.'+String(rounded%divisor).padStart(decimalPlaces,'0'):String(rounded*10n**BigInt(-decimalPlaces));if(text.includes('.'))text=text.replace(/0+$/,'').replace(/\.$/,'');text+=groups[group]}
  return {text:(approximate?'≈':'')+(negative?'-':'')+text,exact:raw,rounded:approximate};
 }
export {scaleFactor,coefficientText,rationalText,rationalDisplay,value,format,magnitude,coverage,compactScalar};

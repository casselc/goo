/* Source scalar parsing shared by raw observation profiles; no measurement inference. */
export function scalarValue(raw){
 if(raw===null||raw===undefined)return {value:null,state:'absent'};
 if(typeof raw!=='number'&&typeof raw!=='string')return {value:null,state:'invalid'};
 const text=String(raw);
 if(/^nan$/i.test(text)||/^[+-]?inf(?:inity)?$/i.test(text))return {value:null,state:'nonfinite'};
 const digit='[0-9](?:_?[0-9])*',hex='[0-9a-f](?:_?[0-9a-f])*';
 const decimal=new RegExp('^[+-]?(?:'+digit+'(?:\\.(?:'+digit+')?)?|\\.'+digit+')(?:[eE][+-]?'+digit+')?$');
 let value;
 if(decimal.test(text))value=Number(text.replaceAll('_',''));
 else {
  const match=new RegExp('^([+-]?)0[xX]('+hex+'(?:\\.(?:'+hex+')?)?|\\.'+hex+')[pP]([+-]?'+digit+')$','i').exec(text);
  if(!match)return {value:null,state:'invalid'};
  const [whole,fraction='']=match[2].replaceAll('_','').split('.');let mantissa=Number.parseInt(whole||'0',16);for(let i=0;i<fraction.length;i++)mantissa+=Number.parseInt(fraction[i],16)*16**(-i-1);value=(match[1]==='-'?-1:1)*mantissa*2**Number(match[3].replaceAll('_',''));
 }
 return Number.isFinite(value)?{value,state:'finite'}:{value:null,state:'nonfinite'};
}

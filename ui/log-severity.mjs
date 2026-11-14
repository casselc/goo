/* OTel SeverityNumber semantics over the exporter's UInt8 column. No I/O or UI. */
import {object} from './intent-shape.mjs';
export const bands=Object.freeze([
 {key:'trace',label:'Trace',min:1,max:4},
 {key:'debug',label:'Debug',min:5,max:8},
 {key:'info',label:'Info',min:9,max:12},
 {key:'warn',label:'Warn',min:13,max:16},
 {key:'error',label:'Error',min:17,max:20},
 {key:'fatal',label:'Fatal',min:21,max:24},
 {key:'unspecified',label:'Unspecified',min:0,max:0},
 {key:'invalid',label:'Invalid',min:25,max:255}
].map(Object.freeze));
export const keys=Object.freeze(bands.map(band=>band.key));
const summaryKeys=Object.freeze(['count',...keys]);
const uint=value=>typeof value==='string'&&value.length<=20&&/^(0|[1-9][0-9]*)$/.test(value)&&BigInt(value)<(1n<<64n);
export function classify(value){
 if(typeof value==='string'){if(!/^(0|[1-9][0-9]{0,2})$/.test(value))throw Error('SeverityNumber must be an exact UInt8 value.');value=Number(value);}
 if(!Number.isInteger(value)||value<0||value>255)throw Error('SeverityNumber must be an exact UInt8 value.');
 return bands.find(band=>value>=band.min&&value<=band.max).key;
}
export function validateSummary(summary,count){
 object(summary,summaryKeys,'Invalid numeric log severity summary.');
 if(arguments.length===1)count=summary.count;
 if(!summaryKeys.every(key=>uint(summary[key]))||summary.count!==count||keys.reduce((total,key)=>total+BigInt(summary[key]),0n)!==BigInt(summary.count))throw Error('Numeric log severity counts differ from their population.');
 return summary;
}
export function validateOptional(requested,item){
 if(requested)validateSummary(item.logSeverity,item.count);
 else if(Object.hasOwn(item,'logSeverity'))throw Error('Unrequested numeric log severity summary.');
 return item;
}
export function validateConservation(summary,items){
 validateSummary(summary);for(const item of items)validateSummary(item.logSeverity,item.count);
 for(const key of summaryKeys)if(items.reduce((total,item)=>total+BigInt(item.logSeverity[key]),0n)!==BigInt(summary[key]))throw Error('Independent numeric log severity reads disagree.');
 return summary;
}
// Raw ordinary exploration admission; this grants no Dataset/source authority.
export function validateProjection(question,data){
 validateOptional(question.logSeverity===true,data);
 for(const key of ['groups','timeSeries'])if(Array.isArray(data[key]))for(const item of data[key])validateOptional(question.logSeverity===true,item);
 if(question.logSeverity){
  const projections=[];
  if(question.groupBy?.length){if(!Array.isArray(data.groups))throw Error('Numeric severity requires the requested groups.');projections.push(data.groups);}
  if(question.bucketMs){if(!Array.isArray(data.timeSeries))throw Error('Numeric severity requires the requested intervals.');projections.push(data.timeSeries);}
  if(data.state!=='changed_between_reads')for(const items of projections)validateConservation(data.logSeverity,items);
 }
 return data;
}
export function bandPredicate(key){
 const band=bands.find(item=>item.key===key);if(!band)throw Error('Choose an OTel severity band.');
 const field=()=>({kind:'column',key:'severity_number'});
 if(key==='unspecified')return {field:field(),op:'eq',value:'0'};
 if(key==='invalid')return {field:field(),op:'gte',value:'25'};
 return {all:[{field:field(),op:'gte',value:String(band.min)},{field:field(),op:'lte',value:String(band.max)}]};
}

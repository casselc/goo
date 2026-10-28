/* Supplied numeric-bin geometry and exact labels. */
import {coefficientText} from './numeric-values.mjs';
 function edges(type){if(type==='decimal64')type='int64';if(!['int64','uint64'].includes(type))throw Error('Unsupported integer type.');const values=[];if(type==='int64')for(let i=63;i>=0;i--)values.push(String(-(1n<<BigInt(i))));values.push('0');for(let i=0;i<=(type==='int64'?63:64);i++)values.push(String(1n<<BigInt(i)));return values}
 function displayBound(value,binding){return coefficientText(value,binding?.scale)}
export {edges,displayBound};

/* Safe convenience projection from an arbitrary supplied bank. */
import {prepare} from './raw-observation-dataset.mjs';
import {canonical} from './raw-observation-identity.mjs';
export function dataset(data,occurrence){
 const bank=prepare(data),key=typeof occurrence==='string'?occurrence:occurrence?.occurrenceKey;
 if(typeof occurrence==='object'&&canonical(bank.data.rows.find(row=>row.occurrenceKey===key))!==canonical(occurrence))throw Error('Choose one occurrence from this held histogram bank.');
 return bank.histogram(key);
}

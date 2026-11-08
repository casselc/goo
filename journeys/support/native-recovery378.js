const {fixture:exemplars}=require('./histogram-exemplars377');
async function fixture(){const retained=await exemplars(),question=structuredClone(retained.question);question.definition.name='native-recovery378';question.definition.title='Histogram source recovery';return{...retained,question}}
const canonical=value=>JSON.stringify(value,(_key,item)=>item&&typeof item==='object'&&!Array.isArray(item)?Object.fromEntries(Object.entries(item).sort(([a],[b])=>a.localeCompare(b))):item);
module.exports={fixture,canonical};

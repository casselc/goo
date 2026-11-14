// Destination presentation does not own capture publication or investigation state.
export function captureDestination({target='',questions=[],questionKey=''}={}){
 const current=target==='current',showQuestion=current&&questions.length>0;
 const selected=showQuestion?questions.find(question=>question.key===questionKey):undefined;
 return {showQuestion,questionKey:selected?.key||'',selected,
  note:!target?'':current?'Adds to the current draft; save its revision to keep the reference.':'Adds a new revision to the selected investigation.'};
}

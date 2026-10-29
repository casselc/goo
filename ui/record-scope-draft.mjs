/* Application-owned draft scope. No DOM, source reads, navigation or publication. */
import * as Environment from './record-environment.mjs';
import * as Session from './question-session.mjs';
const copy=value=>structuredClone(value);
class RecordScopeDraft {
  #fields;
  constructor(environment=''){this.#fields={environment}}
  // Text editing deliberately replaces any prior selector, including invalid drafts.
  setName(environment){this.#fields={environment}}
  restore(question){const fields=Environment.questionFields(question);this.#fields=fields}
  snapshot(){return copy(this.#fields)}
  matches(question){try{return Session.identity(Environment.questionFields(question))===Session.identity(Environment.questionFields(this.#fields))}catch{return false}}
}
export {RecordScopeDraft};

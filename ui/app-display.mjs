/* Application-only bridge while legacy owners migrate to explicit imports.
 * Embedders import display.mjs or individual features; this is not a public entry. */
import * as display from './display.mjs';
for (const [name, api] of Object.entries(display)) globalThis['Goo'+name] = api;
// Temporary bindings for pure helpers extracted from legacy application owners.
import * as NumericValues from './numeric-values.mjs';
import * as NumericBins from './numeric-bins.mjs';
import * as LogPatternIdentity from './log-pattern-identity.mjs';
Object.assign(globalThis,{GooNumericValues:NumericValues,GooNumericBins:NumericBins,GooLogPatternIdentity:LogPatternIdentity});

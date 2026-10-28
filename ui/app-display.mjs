/* Application-only bridge while legacy owners migrate to explicit imports.
 * Embedders import display.mjs or individual features; this is not a public entry. */
import * as display from './display.mjs';
for (const [name, api] of Object.entries(display)) globalThis['Goo'+name] = api;

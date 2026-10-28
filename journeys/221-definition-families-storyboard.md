# Definition families

An investigator opens Definitions to reuse a record definition. Records is the
initial family; native measurements and application contracts have their own
panels rather than continuing below the editor.

1. Enter an unsaved record draft and filter the record library.
2. Use ArrowRight, Home or End on the family tabs. Only the selected panel is
   visible; selecting a family does not read telemetry or publish a definition.
3. Filter Measurements, visit App contracts, leave the workspace and return.
   The selected family, both filters and record draft remain intact.
4. Evaluate an HTTP package in App contracts. Switching families cancels that
   read, rejects late results, and leaves an explicit retry available.
5. A draft/preview handed off from exploration or Examples selects Records.

Executable checks: `221-definition-families.spec.js`, the retained-revision and
annotation journey `definition-library.spec.js`, and the updated definition-library
and legacy-owner Bombadil actors. The 219 workflow journey checks cross-workspace
preview. No new persistence, identity or source semantics are introduced.

Current narrow-family frame:

![Definition family tabs](../docs/journeys/screenshots/221-definition-families-mobile.png)

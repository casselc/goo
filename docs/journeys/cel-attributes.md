# Filter reported attributes without conflating missing and empty

An operator has ordinary logs or spans with resource and record attributes.
Some producers omit a key, some explicitly report an empty string, and others
report values such as `"0"`, `"7"` or `"west"`. The operator needs to isolate one
population without accidentally treating an absent key as empty or different.

## Executable storyboard

`journeys/cel-attributes.spec.js` sends five in-scope real OTLP logs and five spans through
the collector. Resource and record attributes intentionally differ, so selecting
a qualified namespace changes the population. The journey independently reads
the exported ClickHouse maps and compares every explicit query result with that
source bank. Each signal also has one record outside the selected environment
and one outside the parent predicate, proving that the explicit reads retain
both restrictions. Successful product responses are never synthesized.

1. Stage the fixed environment, event-time window and fixture population. Opening
   Filters and choosing Expression reads no telemetry.
2. Check `"example.zone" in resource.attributes` to select reported keys,
   including explicitly empty values. Check remains source-free; **Run** reads.
3. Check `!("example.zone" in record.attributes)` to select omitted record keys.
   Resource and record namespaces stay independent.
4. Check `resource.attributes.eq("example.zone", "")` for reported empty values.
   Missing keys do not match.
5. Check `record.attributes.ne("example.zone", "")` for reported nonempty values.
   Missing keys still do not match; this is not the complement of equality.
6. Use `.contains("example.zone", "7")` or `.eq("example.zone", "7")` to inspect
   string values. Run retains the same environment, window and parent predicate.
7. Build the same nonempty condition with ordinary visual controls. Its predicate
   leaf and resulting population agree with CEL. Switching modes does not imply
   an automatic whole-question conversion.
8. Try bare indexing, a dynamic key, a numeric comparison value or general
   negation. Checking refuses without acquiring telemetry or publishing state.
   A valid replacement clears the diagnostic and restores explicit Run.

The source additionally reports an SDK integer attribute. Its exporter column
contains the string `"7"`. This journey verifies the physical string projection;
it does not claim to reconstruct SDK types or implement numeric coercion.

## Runtime frames

The final accepted run generated the shared expression editor with its native
**Attribute examples** disclosure opened using the keyboard. The disclosure
reads no telemetry and uses wrapping code examples at narrow widths.

- [Desktop attribute condition](screenshots/440-attribute-condition-1440.png)
- [390 px attribute condition](screenshots/440-attribute-condition-390.png)
- [320 px attribute condition](screenshots/440-attribute-condition-320.png)

## Validation

Final real-app journey passes **42.1s** (test body 40.1s) against the final compiled
Go runtime. The run contains **46 recorded steps**, **12 CEL explicit reads** and
**two visual reads**, with all returned span IDs checked against the independent
exporter oracle. **22 source-free condition checks** include eight refusals and
two valid recoveries. Both signals preserve the complete fixed parent question;
outside-environment and outside-parent records remain excluded. No browser
console errors occur, and the owned runtime cleans up successfully.

Independent examples start from a fresh document with the same fixed question.
This avoids treating an intentionally retained ordinary filter or draft from an
earlier case as part of the next semantic example. Empty predicate values use
the normal Go wire omission for an empty string; absence of an attribute key is
checked separately through reported-map presence. No control guard was relaxed
for acceptance.

Run receipt: `artifacts/cel-attributes440/79ea9b6700eb/report.json`.

The existing mounted CEL editor storyboard also passes **38.7s** with the
updated 25-second Bombadil campaign. Its valid-draft choices now use qualified
resource equality, record inequality or standard membership. This run selected
`record.attributes.ne("http.route", "")`, completed all ten phases in **487
snapshots**, checked one context and two plans, and made **zero source reads or
forbidden POSTs**. Hidden callbacks remained ineligible; no overflow or browser
errors occurred. The owned runtime cleaned up. The campaign does not claim that
all three random choices were executed.

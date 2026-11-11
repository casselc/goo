# Authored numeric histograms (407)

The generated carriers use the existing histogram projection/independent bin
fixture across signed, unsigned and fixed-decimal bindings, including empty and
invalid-only populations. They preserve complete source/view/store/template
structures through Go and JavaScript. These synthetic values establish contract
behavior, not additional source observations.
The reused synthetic source returns an empty record page: positive aggregates
therefore retain unknown coverage after the ordinary coherence check. Complete-only
display is refused for those fixtures and admitted for the coherent empty case.
Importing either form grants no source guarantee.

The actual journey validator matches all 46 persisted OTel records to the fixture,
recomputes membership, integer parsing, coverage and bins independently, checks
full destination queries and current dataset clocks, and matches physical SQL to
the existing compiled plans. The observed-effect model checks all 21 interaction
boundaries, including local changes and source-free planning. Its planted local
read must fail. The separate actor validator binds Bombadil's one accepted read
and DOM bins to the same source bank.

Run serially from the repository root with its configured toolchain:

```sh
(cd verification && go test . -run TestHegelHistogramView407 -count=1 -v)
python3 tests/histogram_view_407_api.py
GOO_GENERIC_BINARY=/path/to/goo npx playwright test journeys/407-authored-histogram.spec.js --workers=1
GOO_GENERIC_BINARY=/path/to/goo npx playwright test journeys/407-authored-histogram-actor.spec.js --workers=1
(cd verification && GOO_HISTOGRAM_407_REPORT=$PWD/../artifacts/histogram407/summary.json GOO_HISTOGRAM_407_ACTOR_REPORT=$PWD/../artifacts/histogram407/actor.json go test . -run 'Test(QuintHistogramViewObserved407|HistogramViewActorObserved407)$' -count=1 -v)
```

Missing actual reports skip the opt-in observed checks. A skipped check does not
establish source, browser or model coverage. The API harness checks wire shape;
exact binding equality and range authority remain semantic/runtime obligations.

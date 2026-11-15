# Typed numeric CEL 443 acceptance

Accepted after independent source review and the cost-estimator follow-up for the
bounded [numeric attribute condition contract](../CEL_NUMERIC_ATTRIBUTES.md).
This review covers CEL declarations, lowering, reference evaluation, inverse
printing and the existing contextual editor integration. It does not claim a
whole-product accessibility audit or an independently rerun live browser campaign.

The declared `int64`, `uint64` and `decimal64` methods return an internal opaque
handle. A directly nested terminal method lowers to the existing ordinary
`NumericBinding` and predicate. Literal keys and units, literal decimal scale,
and exact canonical threshold strings are required. The runtime terminal delegates
to `exploration.Match`; there is no independent CEL numeric parser, float path,
source executor or inferred SDK type. Resource and record maps retain separate
activation maps, including when the same key is interpreted with different
explicit numeric declarations. Missing, empty and invalid never become zero;
numeric inequality matches valid unequal values only.

Compiler admission checks resolved overload identities and the literal constructor
shape before constructing a usable program. Computed arguments, standalone
handles, general arithmetic/indexing and native handle comparison have no source
lowering. Existing expression, predicate and cost budgets remain in force.
Inverse printing recompiles and compares complete typed JSON identity, retaining
branch structure, type, unit, scale and threshold spelling; it returns no partial
expression when a structure cannot be represented.

The initial review identified an overly conservative constructor cost that made
ordinary multi-clause conditions refuse early. The follow-up distinguishes exact
constructor overloads, bounded to 256 units for admitted metadata, from terminal
operations retaining the conservative `MaxFieldBytes` estimate. A regression now
compiles/evaluates twelve numeric clauses within the ordinary/CEL budgets and
confirms that a 64 KiB noncanonical numeric source value remains invalid. Inputs
beyond the reference text envelope remain errors, as documented, rather than
silently gaining an invalid-state verdict. No work-budget relaxation was added.

The UI adds examples inside the existing collapsed Attribute examples disclosure;
it introduces no new execution mode or automatic source read. Existing condition
ownership and explicit Apply/Run remain authoritative. Numeric conditions lower
to the same editable ordinary filter used by visual controls. Unit labels do not
rescale data, and a dotted OTel key remains one literal qualified key.

The reviewer independently ran the full `querycel` tests before and after the
cost follow-up; both runs passed without skips. The implementation owner separately
reports passing `celhttp` and `cmd/goo` tests, the 1,856-test UI suite, and independent
Hegel and real ClickHouse parity checks. Live OTLP browser and Bombadil acceptance,
responsive frames and their final receipts belong in [validation](../../VALIDATION.md)
and the owning storyboard; they were not independently executed by this reviewer.

No blocking source-review findings remain within 443. Numeric aggregation,
arbitrary CEL arithmetic, semantic-convention type/unit inference and broader
query-language expansion retain their separate contracts.

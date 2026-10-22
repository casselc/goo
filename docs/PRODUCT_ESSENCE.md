# What goo is for

**Goo turns telemetry into shared definitions of system behavior, and uses those
definitions to investigate what happened.**

An engineer should be able to ask a meaningful question about an application,
inspect the evidence behind the answer, and make that understanding reusable by
someone else. The application can be written in any language. A goo manifest can
help, but existing OpenTelemetry data must be a useful starting point.

**Every conclusion should point to evidence, and every recurring question should
be reusable.**

The [recent-activity design](LIVE_EXPLORATION_DESIGN.md) applies this essence to the
likely first question: “What is happening now in the system I care about?” It
proposes one Explore canvas for narrowing, selecting, comparing and preserving,
with reusable definitions available progressively rather than required at entry.

## The product loop

1. **Ask.** Start with a behavior that matters: a slow operation, stalled work,
   an outcome, a resource constraint. Discover the relevant telemetry and make
   the population, time window and meaning explicit.
2. **Investigate.** Evaluate that question, move between summaries and individual
   traces/logs/metric streams, and see what the evidence supports or leaves unknown.
3. **Preserve.** Capture the fixed question and its evidence; record an explanation
   that cites it. Reopening must not silently turn old evidence into a new answer.
4. **Reuse.** Save the definition so the next person can ask the same question of
   new evidence. Where useful, bring the definition back into instrumentation and
   code review. This is distinct from reusing an old evidence capture.

The durable objects are **definitions**, **evidence captures** and
**investigations** that explain the evidence. Signals and visualizations are ways
to answer a question. A package is the versioned authoring/sharing format for a
definition; users should not need to learn its implementation before investigating.

## The distinctive promise

The same explicit meaning should survive a move from code to interactive
exploration to saved evidence, and eventually to another data platform. It should
be possible to inspect why an answer was produced: which population, which time
semantics, which operation/outcome or instrument identity, and which source data.
A missing observation must stay distinguishable from a healthy or failed outcome.

This is the organizing direction. Today the repository implements a local
workbench, typed definitions, bounded source adapters, immutable captures and
ordered investigation notes. Production collaboration and full warehouse
translation remain separate work; this statement does not imply they are complete.

## A coherence test for features

A feature earns its place when it helps someone ask a better question, inspect its
evidence, explain a finding, or reuse the meaning. Its scope and time semantics
should survive every handoff. Navigation should describe the user's task;
implementation details belong in progressive disclosure.

For example, Services should open an investigation with its service/environment
already selected. Definitions should contain the same packages the workbench
uses. A preview should never look like evidence, and leaving a view should not
publish an abandoned read. A trace waterfall and a native metric histogram can
have different controls while sharing the same capture and notes workflow.

Use this test before adding another top-level destination, preset or editor.
Improve the common loop before adding another isolated demonstration.

The proposed [whole-product annotation model](PRODUCT_ANNOTATIONS_DESIGN.md)
adds reusable meaning to this loop through scoped entity context, versioned
definitions, investigation assertions and exact evidence targets. Training
examples are a reviewed projection of that meaning, not the sole purpose of
annotation or a separate mandatory workflow.

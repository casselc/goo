# Native capacity architecture and Go review — 246

Architecture reviewed the design and initial implementation; the Go specialist
reviewed the implementation read-only. Both found the bounded shared policy sound
for cooperating writers on a stable local directory. Neither review ran tests or
edited files; executed checks are in VALIDATION.

Architecture caught a first-publication race: a reader that saw a missing root
could later scan the newly created root without a lock. Missing-root reads now
use a builtin-only snapshot for that invocation. The duplicate winner verifier
was moved into the shared internal owner instead of retaining a second cmd copy.

Review follow-ups now have coverage: independent Store instances and subprocesses
contending for the last slot, physical aliases versus logical union, builtin
ownership, canonical YAML size admission, corrupt winner precedence and lock-wait
cancellation. Exact saved-ID lookup is connected to the existing workbench
Package definition disclosure, not just exposed as an unused API.

The final Go review found no blocker. Directory replacement, external writers
that ignore locks and unsuitable network filesystems remain outside the contract;
unsupported platforms fail closed. No public synchronization API, new dependency,
generic storage framework, deletion-based recovery or catalog truncation was added.

# Journey 205 — reported workload context and container identity

Status: source implemented; live acceptance pending.
[Contract and scope diagram](../docs/KUBERNETES_WORKLOAD_CONTEXT.md).

1. Find pod-bearing resource observations in an explicit environment/fixed UTC
   window. See reported workload kinds and compact container identity without
   inferring a controller hierarchy.
2. Use the local workload facet to isolate one qualified reported context.
   Same names with different UID/namespace/cluster remain distinct; missing and
   empty values stay separate. Filtering/disclosure add no source reads.
3. Open a workload's logs or spans as a staged record question. The original
   environment/window and all five qualifier/pair states remain exact; pod,
   container, service and unrelated workload attributes are intentionally dropped.
4. Explicit Run finds sibling-pod matching records and excludes differently
   qualified contexts. Inspect/capture through existing record owners.
5. Back restores the held workload facet, resource rows and opener focus without
   reading inventory/records again. Hidden or superseded actions cannot open
   work. Missing pod UID does not disable valid workload context.
6. Observe exact container readiness and distinguish original/replacement
   container IDs in compact cards. Native/frozen owners share the identity cue.
7. Exercise a bounded Bombadil campaign and independent raw membership/ownership
   checks with rejecting controls. Capture actual compact desktop/mobile frames.

The generated fixture travels through ordinary OTLP and ClickHouse. It is not a
running Kubernetes API/receiver, verified ownership or full inventory test.
Actual results and artifact paths will be recorded after the source freeze.

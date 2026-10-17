# Sources and design influences

Reviewed for this project on 2026-10-16. Links point to primary documentation or author publications. Current documentation can differ from the pinned collector and application versions; validate installed schemas and behavior.

| Source | Contribution |
| --- | --- |
| [Hadley Wickham A layered grammar of graphics](https://vita.had.co.nz/papers/layered-grammar.html) | Decompose data, statistical transformation, encodings, scales, coordinates, and facets |
| [Plotje poses](https://scicloj.github.io/plotje/plotje_book.pose_model.html) | Plain Clojure values and composable transformations before rendering |
| [Plotje composition](https://scicloj.github.io/plotje/plotje_book.composition.html) | Composition as a design influence for view structure |
| [Observability Engineering second edition](https://www.oreilly.com/library/view/observability-engineering-2nd/9781098179915/) | Structured events, production validation, analysis workflows, and SLO practices |
| [Liz Fong-Jones Structured Events Are the Basis of Observability](https://www.honeycomb.io/blog/structured-events-basis-observability) | Rich context and high-cardinality exploration |
| [Charity Majors Observability 1.0 and 2.0](https://charity.wtf/p/there-is-only-one-key-difference-between-observability-1-0-and-2-0) | Unified event context as a foundation for exploration |
| [George Miranda Core Analysis Loop](https://www.honeycomb.io/blog/debugging-with-the-core-analysis-loop-and-what-to-build-vs-buy) | Iterative hypothesis and evidence refinement |
| [Brendan Gregg USE method](https://www.brendangregg.com/usemethod.html) | Resource-oriented questions and explicit measurement gaps |
| [Google SRE Alerting on SLOs](https://sre.google/workbook/alerting-on-slos/) | Burn rates, multiple windows, and low-traffic considerations |
| [Vega-Lite selection parameters](https://vega.github.io/vega-lite/docs/selection.html) | Declarative interaction and linked selection |
| [ClickStack tables and schemas](https://clickhouse.com/docs/clickstack/ingesting-data/schemas) | Backend physical schema and correlation fields |
| [ClickStack Compose](https://clickhouse.com/docs/clickstack/deployment/docker-compose) | Local component deployment |
| [ClickStack API reference](https://clickhouse.com/docs/clickstack/api-reference) | Potential dashboard and alarm export adapter |
| [Pinned collector exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/v0.123.0/exporter/clickhouseexporter) | Configuration for the bundled upstream collector |
| [OTel metrics data model](https://opentelemetry.io/docs/specs/otel/metrics/data-model/) | Temporality, aggregation, metric identity, and histogram semantics |
| [OTel messaging spans](https://opentelemetry.io/docs/specs/semconv/messaging/messaging-spans/) | Batch/message distinction and asynchronous span links |

The second-edition book's public description and contents informed the design; the complete book was not analyzed. The proposed grammar and RFCs are a project synthesis, not a standard attributed to any of these authors. Honeycomb authors' event-first approach is one design influence; native OTel metrics retain their distinct semantics in the proposed language.

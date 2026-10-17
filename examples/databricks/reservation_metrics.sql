-- Illustrative compiler target; not deployed or executed against Databricks.
-- Requires prepared subject facts, not raw otel_logs.
-- Unique key: (evaluation_id, environment, tenant_id, service_name, reservation_id).
-- Each evaluation_id pins T, D, L, W, evidence basis, and definition revision.
-- Unknown/missing identities and empty scopes also require evaluation receipts.
CREATE OR REPLACE VIEW telemetry.semantic.reservation_metrics
WITH METRICS
LANGUAGE YAML
AS $$
version: "1.1"
source: telemetry.semantic.reservation_subject_facts
comment: "Reservation outcomes at a pinned evaluation basis; select one Evaluation ID."
filter: in_cohort = true
dimensions:
  - name: Evaluation ID
    expr: evaluation_id
  - name: Environment
    expr: environment
  - name: Tenant
    expr: tenant_id
  - name: Service
    expr: service_name
  - name: Release
    expr: release
  - name: Accepted At
    expr: accepted_at_utc
measures:
  - name: Eligible Subjects
    expr: COUNT(1)
  - name: Good Subjects
    expr: COUNT(1) FILTER (WHERE classification = 'good')
  - name: Bad Subjects
    expr: COUNT(1) FILTER (WHERE classification = 'bad')
  - name: Unknown Subjects
    expr: COUNT(1) FILTER (WHERE classification = 'unknown')
  - name: Impaired Subjects
    expr: COUNT(1) FILTER (WHERE quality_impaired = true)
  - name: Success Ratio
    expr: |
      CASE
        WHEN MEASURE(`Unknown Subjects`) > 0
          OR MEASURE(`Impaired Subjects`) > 0 THEN NULL
        ELSE MEASURE(`Good Subjects`) * 1.0
          / NULLIF(MEASURE(`Eligible Subjects`), 0)
      END
$$;

-- Caller binds a single compatible evaluation ID. No outcome filter may change
-- the objective denominator. Details/inspection use a separate query surface.
SELECT
  `Service`,
  `Release`,
  MEASURE(`Eligible Subjects`) AS eligible,
  MEASURE(`Good Subjects`) AS good,
  MEASURE(`Bad Subjects`) AS bad,
  MEASURE(`Unknown Subjects`) AS unknown,
  MEASURE(`Impaired Subjects`) AS impaired,
  CASE WHEN :evidence_impaired THEN NULL
       ELSE MEASURE(`Success Ratio`) END AS success_ratio,
  CASE WHEN :evidence_impaired
         OR MEASURE(`Unknown Subjects`) > 0
         OR MEASURE(`Impaired Subjects`) > 0 THEN 'unknown'
       WHEN MEASURE(`Eligible Subjects`) = 0 THEN 'no_data'
       ELSE 'ok' END AS evidence_state
FROM telemetry.semantic.reservation_metrics
WHERE `Evaluation ID` = :evaluation_id
GROUP BY ALL;

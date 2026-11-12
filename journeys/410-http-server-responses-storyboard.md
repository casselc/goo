# 410 · Incoming HTTP response context

An operator investigates reported response values without turning them into an
inferred failure policy.

1. Reopen a fixed inventory capture, browse live discovery and open the exact
   namespace/service. Its current Incoming HTTP profile selects reported Server
   spans with nonempty `http.request.method`. Nine records qualify. Foreign
   namespace, conflicting environment, other-service, Client and empty/legacy-only
   method controls do not qualify; a legacy-only environment row is admitted under
   the explicit effective policy.
2. Open **Response context**. This native disclosure starts no telemetry reads.
   The supplied response widget contains literal `200` (2), `503` (2), missing (3),
   empty (1) and `5xx` (1) groups. A `200` span with recorded Error and a `503` span
   with recorded OK demonstrate that HTTP response status and span status remain
   separate. A legacy `http.status_code` value is not an alias for the current key.
3. Open the missing, empty and `503` populations in turn. Each link stages the
   original full scope plus one exact qualified missing/equality predicate.
   **Run** explicitly reads the selected records (3, 1 and 2 respectively).
   Return through **Services & resources** to the identical held widget nodes and
   read evidence; no service or widget source read is repeated.
4. Add the full response widget question to an investigation and save it. Run and
   capture explicitly rereads the fixed population; the resulting immutable
   evidence retains all nine records and its own read basis. Reopen that capture
   without a source read. Saving a question and capturing evidence are distinct.

The fixture sends fifteen real OTel spans to the collector. The journey reads
persisted ClickHouse rows and independently evaluates scope, membership and literal
group counts. This is source-backed acceptance, not a mocked positive response.
The separate Bombadil campaign repeats discovery → missing response population →
Run → held Services return at 390px, permitting no publication.

Executable checks:

- `410-http-server-responses.spec.js`
- `410-http-server-responses-actor.spec.js`
- `verification/bombadil/http-server410.ts`

The main journey generates actual desktop/390px documentation frames:
`docs/journeys/screenshots/410-incoming-http-responses.png` and
`docs/journeys/screenshots/410-incoming-http-responses-390.png`.
Preparation is not runtime acceptance; final results belong in `VALIDATION.md`.

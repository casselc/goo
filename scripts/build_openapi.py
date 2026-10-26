"""Assemble the reviewed HTTP surface plus reflected Go wire schemas.
Run after go run ./scripts/openapi-schema api/go-schemas.json.
"""
import json
import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
ref = lambda name: {'$ref': '#/components/schemas/' + name}
schemas = json.loads((ROOT / 'api/go-schemas.json').read_text())
for name in ('workbench.ActivityView','workbench.ActivityViewInput'):
 schemas[name]['properties']['baselineCapture']['description']='Optional exact immutable raw-span duration capture ID. Publication verifies stored evidence and same environment; opening a view stages configuration without loading evidence or starting telemetry. Omitted/empty preserves legacy identity.'

schemas['annotations.OperationSpan']['properties']['reportedStatus']['description']='Optional exact exporter StatusCode, independent of declared outcome. Omission means not retained in this evidence; empty and future values remain explicit unknown values. Legacy saved bytes are unchanged.'

schemas['Error'] = {'type': 'object', 'required': ['error'], 'properties': {'error': {'type': 'string'}, 'hint': {'type': 'string'}}, 'additionalProperties': True}
schemas['Cancellation'] = {'type': 'object', 'required': ['state'], 'properties': {'state': {'const': 'cancellation-requested'}}}
# Counted values are two explicit observations, never a transactional snapshot.
counted=schemas['exploration.CountedValuesResult']['properties']
counted['version']['const']='goo/counted-values/v1'
counted['state']['enum']=['observed','no_data','changed_between_reads']
counted['count']['pattern']='^(0|[1-9][0-9]*)$'
counted['basis']['pattern']='^[a-f0-9]{64}$'
counted['capturedAt']['format']='date-time'
schemas['exploration.CountedValuesResult']['description']='Exactly two bounded aggregate reads: complete count then complete string groups, with separate read clocks and content bases. changed_between_reads retains diagnostic observations but must not drive population actions. No source rows, result token, capture, native arithmetic or transactional snapshot.'
schemas['exploration.CountedValue']['properties']['count']['pattern']='^[1-9][0-9]*$'
read=schemas['exploration.CountedValuesRead']['properties']
read['basis']['pattern']='^[a-f0-9]{64}$';read['readAt']['format']='date-time'
schemas['exploration.RecordAggregatesResult']['additionalProperties']=False
aggregate=schemas['exploration.RecordAggregatesResult']['properties']
aggregate['version']['const']='goo/record-aggregates/v1'
aggregate['state']['enum']=['observed','no_data','changed_between_reads']
aggregate['count']['pattern']='^(0|[1-9][0-9]*)$';aggregate['basis']['pattern']='^[a-f0-9]{64}$';aggregate['capturedAt']['format']='date-time'
schemas['RecordAggregateGroup']={'type':'object','required':['g0','count'],'properties':{'g0':{'type':['string','null']},'count':{'type':'string','pattern':'^[1-9][0-9]*$'},'duration':{'$ref':'#/components/schemas/exploration.DurationSummary'},'spanStatus':{'$ref':'#/components/schemas/exploration.SpanStatusSummary'}},'additionalProperties':True}
aggregate['groups']={'type':'array','items':{'$ref':'#/components/schemas/RecordAggregateGroup'}}
schemas['exploration.RecordAggregatesResult']['description']='Aggregate-only record projection: exactly count and groups reads plus one optional time-series read. Separate clocks/content bases do not establish a transactional snapshot. Changed aggregates are diagnostic and non-actionable. Never includes source rows, fields, plan, pagination or retention token.'
# Additive cumulative nonmonotonic sum observations, distinct from counter rates.
schemas['metrics.StockResult']['description']='Per-stream cumulative nonmonotonic sum observations. Signed finite fractional values and accumulation epochs are preserved; decreases are not resets. No delta integration, fleet total, rate, capacity or utilization is inferred.'
for name in ('metrics.StockSample','metrics.StockSeries'):
 schemas[name]['properties']['state']['enum']=['observed','stale','unknown']
stock_sample=schemas['metrics.StockSample']
stock_sample['properties']['value']={'type':'number','description':'Finite observed cumulative tally, possibly signed/fractional. Omitted when unknown; stale observations retain their value.'}
stock_sample['properties']['atNS']['pattern']='^(0|[1-9][0-9]*)$'
stock_sample['allOf']=[{'if':{'properties':{'state':{'enum':['observed','stale']}},'required':['state']},'then':{'required':['value']}},{'if':{'properties':{'state':{'const':'unknown'}},'required':['state']},'then':{'not':{'required':['value']}}}]
# stockDomain is optional; legacy generic stocks keep signed fractional semantics.
for name in ('measurements.Definition','measurements.DefinitionInput'):
 schemas[name]['properties']['stockDomain']['description']='Optional stock-only nonnegative-integer policy. Recorded values must be integers >=0 and <2^53; missing-value flags remain unknown. Empty/omitted keeps generic signed finite fractional stock semantics.'
schemas['measurements.Definition']['properties']['stockDomain']['const']='nonnegative-integer'
schemas['measurements.Result']['allOf']=[{'if':{'properties':{'question':{'properties':{'definition':{'properties':{'stockDomain':{'const':'nonnegative-integer'}},'required':['stockDomain']}},'required':['definition']}},'required':['question']},'then':{'properties':{'stocks':{'properties':{'series':{'type':'array','items':{'properties':{'samples':{'type':'array','items':{'properties':{'value':{'type':'integer','minimum':0,'maximum':9007199254740991}}}}}}}}}}}}]
# Runtime metadata is encoded by small Go/Python maps; schemas are authored here.
schemas['RuntimeCapabilityState']={'type':'object','required':['state'],'properties':{'state':{'enum':['configured','unconfigured']}},'additionalProperties':False}
schemas['RuntimeCapabilities']={'type':'object','required':['version','profile','preparedHTTP','declaredOperations'],'properties':{'version':{'const':'goo/runtime-capabilities/v1'},'profile':{'enum':['lab','generic']},'preparedHTTP':{'$ref':'#/components/schemas/RuntimeCapabilityState'},'declaredOperations':{'$ref':'#/components/schemas/RuntimeCapabilityState'}},'additionalProperties':False,'description':'Source-free startup configuration only. Configured does not establish source availability, authorization, telemetry coverage, ingestion completeness or service health.'}
schemas['RuntimeCapabilityRefusal']={'type':'object','required':['error','code','capability'],'properties':{'error':{'type':'string'},'code':{'const':'capability_unconfigured'},'capability':{'enum':['prepared-http','declared-operations','example-grammar','example-catalog']}},'additionalProperties':False}
schemas['UIConfig']={'type':'object','required':['version','profile','examples'],'properties':{'version':{'const':'goo/ui-config/v1'},'profile':{'enum':['lab','generic']},'examples':{'$ref':'#/components/schemas/RuntimeCapabilityState'}},'additionalProperties':False,'description':'Pure Python UI startup configuration. Does not query or report readiness of ClickHouse or Go.'}
# Additive dashboard profiles opt in; omitted/empty/null retain the legacy plan.
schemas['exploration.DashboardRequestInput']['properties']['profile']['description']='Optional auto, all-spans, http-server, rpc-server, db-client, messaging-consumer or messaging-producer selection. Omitted/empty/null preserves the legacy v2 response. auto selects HTTP only with a positive supplied server-observation hint plus current-key metadata, or a compatible native HTTP instrument. Automatic precedence is HTTP, legacy RPC, DB client, messaging consumer, messaging producer, then All spans. RPC admission requires both legacy rpc.system/rpc.method keys plus a positive rpcServerObserved hint, or an eligible legacy RPC duration instrument. Flat keys alone leave All spans selected; hints are advisory, not authenticated conformance.'
schemas['exploration.DashboardRequestInput']['properties']['httpServerObserved']['description']='Optional advisory positive hint from already-held, same-scope span observations. false/omitted means no positive hint, not HTTP absence; no new probe or semantic-convention compliance claim is implied.'
schemas['exploration.DashboardRequestInput']['properties']['rpcServerObserved']['description']='Optional advisory positive hint from already-held, same-scope Server span observations with nonempty legacy rpc.system and rpc.method on the same row. false/omitted is unknown, not RPC absence or conformance. HTTP admission has automatic precedence; this does not authorize a probe.'
schemas['exploration.DashboardRequestInput']['properties']['dbClientObserved']['description']='Optional advisory positive hint from already-held, same-scope Client span observations with a nonempty current db.system.name on the same row. false/omitted is unknown, not database absence or health. Auto prefers eligible HTTP then RPC before DB; no new source probe is implied.'
schemas['exploration.DashboardRequestInput']['properties']['messagingConsumerObserved']['description']='Optional advisory positive hint from already-held, same-scope Consumer spans with nonempty messaging.system on the same row. false/omitted is unknown, not absence, delivery success, acknowledgement or broker health. A compatible processing-duration histogram can independently admit the consumer family; it remains a separate population.'
schemas['exploration.DashboardRequestInput']['properties']['messagingProducerObserved']['description']='Optional advisory positive hint from already-held, same-scope Producer spans with nonempty messaging.system on the same row. false/omitted is unknown, not absence or delivery success. Consumer has automatic precedence; processing metrics never admit the producer family.'
schemas['exploration.DashboardPlan']['properties']['profile']['description']='Selected all-spans, http-server, rpc-server, db-client, messaging-consumer or messaging-producer family for opt-in requests; absent on legacy requests.'
schemas['exploration.DashboardPlan']['properties']['families']['description']='Opt-in bounded family descriptors. HTTP/RPC/DB/messaging candidate metadata must be confirmed by the executable question before treating counts as observed; no hidden source read or fallback is performed by this planner.'
schemas['exploration.DashboardFamily']['properties']['state']['description']='Metadata capability state, including candidate. Not a service-health verdict, observed server count or semantic-convention conformance assertion.'
schemas['exploration.DashboardWidget']['properties']['family']['description']='Optional display/planning family identity; not a telemetry field or a shared native/span/log denominator.'
# API policy is deliberately authored, not inferred from HTTP verbs.
reads = {'/http-package', '/operations', '/counter-rates', '/histograms', '/inventory', '/discovery', '/explore', '/explore/counted-values', '/record-aggregates', '/trace-search', '/population/explain', '/metric-discovery', '/measurements', '/investigations', '/captures', '/tail', '/query'}
plans = {'/measurement-packages/plan', '/telemetry-packages/draft', '/telemetry-packages/numeric-plan', '/telemetry-packages/plan', '/recipes', '/dashboards/plan', '/population/plan', '/population/explain/plan', '/population/explain/filter', '/numeric-histograms/selection', '/numeric-histograms/comparison-selection', '/compare', '/explain'}
publications = {'/activity-views','/measurement-packages', '/telemetry-packages', '/annotations', '/exploration-views', '/deployments', '/documents', '/captures/comparisons'}
requests = {'/activity-views':'ActivityViewRequest',
 '/retained-evidence/tail':'TailIntent','/compare':'CompareRequest','/inventory':'InventoryRequest','/discovery':'InventoryRequest','/explore':'RecordQuery','/explore/counted-values':'RecordQuery','/record-aggregates':'RecordQuery','/trace-search':'TraceSearchRequest','/tail':'TailSelector',
 '/measurements':'NativeQuestion','/metric-discovery':'MetricDiscoveryRequest','/measurement-packages/plan':'NativePlanRequest','/measurement-packages':'YAMLRequest',
 '/telemetry-packages':'YAMLRequest','/telemetry-packages/draft':'TelemetryDraftRequest','/telemetry-packages/numeric-plan':'NumericPlanRequest','/telemetry-packages/plan':'TelemetryPlanRequest',
 '/recipes':'RecipeRequest','/dashboards/plan':'DashboardRequest','/population/plan':'PopulationRequest','/numeric-histograms/selection':'NumericHistogramSelectionRequest','/numeric-histograms/comparison-selection':'DistributionSelectionRequest','/annotations':'MeaningInput','/exploration-views':'ExplorationView','/deployments':'DeploymentInput','/documents':'DocumentRequest',
 '/population/explain':'ExplanationRequest','/population/explain/plan':'ExplanationRequest','/population/explain/filter':'ExplanationFilterRequest','/captures':'CaptureRequest','/captures/live':'LiveComparisonRequest','/captures/comparisons':'SaveComparisonRequest','/investigations':'InvestigationRequest','/retained-evidence':'RetainRequest','/retained-evidence/{id}/annotations':'CorrectRequest',
}
responses = {('get','/capabilities'):'RuntimeCapabilities',('post','/population/explain'):'ExplanationResult',('post','/population/explain/plan'):'ExplanationPlan',('post','/population/explain/filter'):'ExplanationStagedQuestion',('post','/activity-views'):'ActivityView',('get','/activity-views/{id}'):'ActivityView',('get','/activity-views'):'ActivityViewList',
 ('post','/compare'):'CompareResult',('post','/inventory'):'InventoryResult',('post','/discovery'):'ServiceDiscovery',('post','/explore'):'RecordResult',('post','/explore/counted-values'):'CountedValuesResult',('post','/record-aggregates'):'RecordAggregatesResult',('post','/trace-search'):'TraceSearchResult',
 ('post','/measurements'):'NativeResult',('post','/metric-discovery'):'MetricDiscoveryResult',('post','/telemetry-packages/draft'):'PackagePlan',('post','/telemetry-packages/plan'):'PackagePlan',('post','/telemetry-packages/numeric-plan'):'NumericPlan',('post','/telemetry-packages'):'SavedPackage',('get','/telemetry-packages/{revision}'):'SavedPackage',
 ('post','/recipes'):'RecipePlan',('post','/dashboards/plan'):'DashboardPlan',('post','/population/plan'):'PopulationPlan',('post','/numeric-histograms/selection'):'NumericHistogramSelection',('post','/numeric-histograms/comparison-selection'):'NumericHistogramSelection',
 ('get','/http-package'):'LiveHTTPResult',('get','/operations'):'LiveOperations',('get','/counter-rates'):'LiveRates',('get','/histograms'):'LiveHistograms',
 ('get','/captures'):'CaptureList',('post','/captures'):'CapturePage',('get','/captures/{id}'):'CapturePage',('get','/investigations/{id}'):'InvestigationCapture',('post','/investigations'):'InvestigationCapture',
 ('post','/annotations'):'MeaningRevision',('get','/annotations/{id}'):'MeaningRevision',('post','/deployments'):'Deployment',('get','/deployments/{id}'):'Deployment',('post','/documents'):'Document',('get','/documents/{id}'):'Document',
 ('post','/retained-evidence'):'RetainedReceipt',('post','/retained-evidence/{id}/annotations'):'CorrectionReceipt',('get','/retained-evidence'):'RetainedCatalog',
}

responses['post','/retained-evidence/tail']='TailReceipt'
responses.update({
 ('get','/annotations'):'MeaningPage',('get','/exploration-views'):'ExplorationViewList',('post','/exploration-views'):'ExplorationViewResult',('get','/exploration-views/{id}'):'ExplorationViewResult',
 ('get','/deployments'):'DeploymentContext',('get','/documents'):'DocumentList',('post','/captures/comparisons'):'ComparisonReceipt',('get','/captures/comparisons/{id}/receipt'):'ComparisonReceipt',
 ('get','/captures/{id}/followups'):'FollowupSet',('get','/captures/{id}/compare'):'EvidenceComparison',('get','/captures/{id}/windows'):'CapturedWindowDiscovery',('get','/captures/{id}/receipt'):'CaptureReceiptMetadata',
 ('get','/investigation-workflows'):'WorkflowList',('get','/operation-packages'):'OperationPackageList',('get','/recipes'):'RecipeList',('get','/telemetry-packages'):'TelemetryPackageList',
 ('get','/retained-evidence/{id}'):'RetainedManifest',('get','/retained-evidence/{id}/annotations'):'MeaningPage',('get','/retained-evidence/{id}/annotations/{revision}'):'MeaningRevision',('post','/captures/live'):'LiveComparisonDraft',
})
# The reflected nullable fields are constrained to a discriminated origin union.
schemas['retention.EvidenceContent']['oneOf']=[
 {'properties':{'origin':{'enum':['saved-capture','result-page']},'capture':{'not':{'type':'null'}}},'required':['capture'],'not':{'required':['tail']}},
 {'properties':{'origin':{'const':'tail-selection'},'tail':{'not':{'type':'null'}},'retainedAt':{'type':'string','minLength':1,'format':'date-time'}},'required':['tail','retainedAt'],'not':{'required':['capture']}}
]
schemas['RetentionStatus']={'type':'object','required':['configured'],'properties':{'configured':{'type':'boolean'}}}
responses['get','/retained-evidence/status']='RetentionStatus'
responses['get','/retained-evidence/catalog']='EvidenceCatalog'
responses['get','/retained-evidence/{id}/content']='EvidenceContent'
# Native registry responses are envelopes rather than Definition directly.
schemas['NativePackage'] = {'type':'object','required':['definition','revision','yaml'],'properties':{'definition':ref('measurements.Definition'),'revision':{'type':'string'},'yaml':{'type':'string'}}}
schemas['NativePlan'] = {'allOf':[ref('NativePackage'),{'type':'object','required':['question','plan'],'properties':{'question':ref('measurements.Question'),'plan':ref('measurements.Plan')}}]}
responses['post','/measurement-packages'] = 'NativePackage'
responses['post','/measurement-packages/plan'] = 'NativePlan'
schemas['NativePackageList']={'type':'array','items':{'allOf':[ref('NativePackage'),{'type':'object','required':['builtin'],'properties':{'builtin':{'type':'boolean'}}}]}}
responses['get','/measurement-packages']='NativePackageList'
# These roots are reached by reflected response types; fail loudly if renamed.
for name in ['measurements.Definition','measurements.Question','measurements.Plan']:
 if name not in schemas: raise ValueError(f'Missing reflected type: {name}')

def go_limit(path):
 if path in ('/compare','/numeric-histograms/comparison-selection'): return 1048576
 if path == '/dashboards/plan': return 262144
 if path.startswith(('/telemetry-packages','/measurement-packages')) or path in ('/population/explain','/population/explain/plan','/population/explain/filter','/measurements','/metric-discovery','/trace-search'): return 65536
 if path in ('/annotations','/exploration-views','/deployments','/documents','/population/plan', '/numeric-histograms/selection','/captures/comparisons'): return 131072
 return 16384

def browser_limit(path):
 if path in ('/population/plan', '/numeric-histograms/selection','/exploration-views','/documents','/captures/comparisons'): return 131072
 if path in ('/compare','/numeric-histograms/comparison-selection'): return 1048576
 if path in ('/population/explain','/population/explain/plan','/population/explain/filter','/trace-search') or path.startswith(('/telemetry-packages','/measurement-packages','/measurements')): return 65536
 return 16384

def effect(method,path):
 if path == '/capabilities': return 'configuration-only'
 if path == '/queries/{id}/cancel': return 'cancel-capability'
 if path == '/captures': return 'source-read-and-capture-or-held-explanation-publication'
 if path == '/captures/live': return 'source-read-or-held-draft-publication'
 if path.startswith('/retained-evidence'): return 'configuration-only' if path.endswith('/status') else 'xtdb-read' if method=='get' else 'xtdb-publication-or-recovery'
 if method=='post' and path in plans: return 'pure-plan-or-supplied-summary'
 if path in reads and (method=='post' or path in ('/http-package','/operations','/counter-rates','/histograms')):
  return 'source-read-with-possible-custody-or-capture' if path in ('/measurements','/captures','/investigations') else 'source-read'
 if method=='post' and path in publications: return 'file-publication-with-reference-validation'
 return 'saved-evidence-or-definition-read'

spec = {
 'openapi':'3.1.0',
 'info':{'title':'goo observability API','version':'0.1.0','description':'Current development API. Browser server prefixes paths with /api; direct Go server uses unprefixed paths. Python-only operations override servers. Wire-shape schemas describe encoding, not all domain validation. Semantic constraints, custody and source effects remain authoritative in Go. Unknown input fields are rejected; null and omitted fields can decode and then fail domain validation. No authentication is implemented by this lab; cancellation IDs are local capabilities, not authorization.'},
 'servers':[{'url':'http://localhost:8088/api','description':'Browser Python proxy (default local lab)'},{'url':'http://localhost:8089','description':'Direct Go service; Go-specific limits/headers apply'}],
 'paths':{},'components':{'schemas':schemas},
 'x-contract-notes':['Direct Go accepts any request Content-Type; application/json is the supported client encoding.','Python reparses/reserializes JSON and enforces Content-Length before Go; fixture byte equality describes direct Go, not browser serialization.','Browser proxy: dashboard plan 16 KiB vs Go 256 KiB; metric discovery 16 KiB vs Go 64 KiB; annotations/deployments 16 KiB vs Go 128 KiB.','Direct operation GET validation uses http.Error text/plain with JSON-looking text; browser re-encodes errors as JSON.','Direct tail early errors omit Cache-Control; proxy early JSON errors include no-store.','Python GET prefix routing can forward unsupported suffixes; these are not supported operations.','Malformed non-JSON upstream HTTP errors now return JSON 502 instead of disconnecting; valid JSON errors preserve status.', 'No generated handlers, semantic validator replacement, API stability guarantee or new authorization layer is introduced.'],
}
route_source={}
for file in sorted((ROOT/'cmd/goo').glob('*http.go')):
 text=file.read_text()
 matches=list(re.finditer(r'HandleFunc\("(GET|POST) ([^" ]+)"',text))
 for index,m in enumerate(matches):
  method,path=m[1].lower(),m[2]
  route_source[method,path]=(str(file.relative_to(ROOT)),text[m.end(): matches[index+1].start() if index+1<len(matches) else len(text)])
route_source['post','/queries/{id}/cancel']=('cmd/goo/query_budget.go','')
for (method,path),(file,block) in sorted(route_source.items()):
 e=effect(method,path)
 op={'operationId':method+'_'+re.sub(r'[^a-zA-Z0-9]+','_',path).strip('_'),'summary':method.upper()+' '+path,'tags':[path.split('/')[1]],'description':f'Current owner: {file}. Effect: {e}. See docs/HTTP_BOUNDARIES.md for lifecycle and limits.','x-go-source':file,'x-effects':e,'parameters':[],'responses':{}}
 for p in re.findall(r'\{([^}]+)\}',path):op['parameters'].append({'name':p,'in':'path','required':True,'schema':{'type':'string'},'description':'Exact revision, receipt ID or process-local capability; not an authorization credential.'})
 # Enumerate explicit query getters; allowlists and cross-field validation stay in Go.
 for p in sorted(set(re.findall(r'(?:q|v|values|r\.URL\.Query\(\))\.(?:Get|Has)\("([^"]+)"',block))):
  op['parameters'].append({'name':p,'in':'query','schema':{'type':'string'},'description':'Single value; route-specific allowlists and cross-field rules apply. Targets are JSON-encoded typed objects where supported.'})
 for p in {'/captures':['cursor'],'/captures/{id}':['cursor'],'/captures/{id}/followups':['row'],'/captures/{id}/compare':['selection','distribution']}.get(path,[]):
  if method=='get' and not any(v['name']==p and v['in']=='query' for v in op['parameters']):op['parameters'].append({'name':p,'in':'query','schema':{'type':'string'},'description':'Single canonical value; saved evidence is not reread from the telemetry source.'})
 controlled=method=='post' and path in ('/metric-discovery','/measurements','/explore','/explore/counted-values','/record-aggregates','/trace-search','/population/explain','/inventory','/discovery','/captures','/captures/live')
 if controlled:op['parameters'].append({'name':'X-Goo-Query-ID','in':'header','schema':{'type':'string','pattern':'^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$'},'description':'Optional short-lived cancellation capability; repeated/cancelled IDs are refused.'})
 if method=='post' and path in requests:
  op['requestBody']={'required':True,'content':{'application/json':{'schema':ref(requests[path])}}}
  op['x-go-body-limit-bytes']=go_limit(path);op['x-browser-body-limit-bytes']=browser_limit(path)
 success='201' if method=='post' and (path in publications or path in ('/retained-evidence/tail','/captures','/investigations','/retained-evidence','/retained-evidence/{id}/annotations')) else '200'
 model=responses.get((method,path));body=ref(model) if model else {}
 op['x-schema-coverage']='typed' if model else 'response-envelope-pending'
 op['responses'][success]={'description':'Current success response. Empty/unknown evidence does not imply healthy or complete telemetry.','content':{'application/json':{'schema':body}}}
 codes = {int(n) for n in re.findall(r'(?:fail|send|reply)\(\s*w,\s*(\d{3})|fail\(\s*(\d{3})|WriteHeader\((\d{3})\)|code\s*:?=\s*(\d{3})',block) for n in n if n}
 codes.update(int(n) for n in re.findall(r'http.Error\(w,.*?,\s*(\d{3})\)',block))
 codes.add(502)  # Python proxy connection/JSON forwarding failure.
 if method=='post': codes.add(400)
 if controlled: codes.update((400,409,429))
 if path.startswith('/retained-evidence') and not path.endswith('/status'): codes.add(503)
 if method=='get' and path in ('/http-package','/operation-packages','/operations','/counter-rates','/histograms'): codes.add(503)
 if path=='/queries/{id}/cancel':codes.update((400,429))
 for status in map(str,sorted(n for n in codes if n>=400)):
  op['responses'][status]={'description':'Existing route or proxy refusal. Semantic validation and effect ordering remain in the owner.','content':{'application/json':{'schema':ref('Error')}}}
 if method=='get' and path in ('/http-package','/operation-packages','/operations','/counter-rates','/histograms'):
  op['responses']['503']['content']['application/json']['schema']=ref('RuntimeCapabilityRefusal')
  op['responses']['400']['content']['text/plain']={'schema':{'type':'string'},'example':'{"error":"unsupported query parameter"}\n'}
 if path=='/queries/{id}/cancel':
  op['responses']['200']['content']['application/json']['schema']=ref('Cancellation');op['x-schema-coverage']='typed'
  op['requestBody']={'required':True,'description':'Send {} through the browser proxy, which requires a positive Content-Length; direct Go ignores the cancellation body.','content':{'application/json':{'schema':{},'example':{}}}}
  op['x-browser-body-limit-bytes']=16384
 if path=='/captures/live': op['responses']['201']={'description':'Freeze held draft without rereading source.','content':{'application/json':{'schema':ref('CapturePage')}}}
 spec['paths'].setdefault(path,{})[method]=op

schemas['TailReset']={'type':'object','required':['connection','windowMs','selector','continuity','replay'],'properties':{'connection':{'type':'string','pattern':'^[a-f0-9]{64}$'},'windowMs':{'const':60000},'selector':ref('TailSelector'),'continuity':{'const':'unknown'},'replay':{'const':False}}}
schemas['TailGap']={'type':'object','required':['reason'],'properties':{'reason':{'enum':['source_unavailable','identity_budget','connection_ended']}}}
schemas['TailBatch']={'type':'object','required':['rows','limited','fromMs','toMs','observedAt','membershipState'],'properties':{'membershipState':{'enum':['available','unavailable','empty']},'membership':ref('TailMembershipCapability'),'rows':{'type':'array','items':ref('TailRow')},'limited':{'type':'boolean'},'fromMs':{'type':'integer'},'toMs':{'type':'integer'},'observedAt':{'type':'string','format':'date-time'}}}
schemas['TailBatch']['allOf']=[{'if':{'properties':{'membershipState':{'const':'available'}}},'then':{'required':['membership']},'else':{'not':{'required':['membership']}}}]
tail=spec['paths']['/tail']['post']
tail['responses']['200']={'description':'SSE frames: per-connection numeric id, event (reset/batch/gap), JSON data, blank line. POST via fetch, not browser EventSource. No resumable cursor or replay guarantee.','headers':{'Cache-Control':{'schema':{'const':'no-store'}},'X-Accel-Buffering':{'schema':{'const':'no'}}},'content':{'text/event-stream':{'schema':{'type':'string'},'example':'id: 1\nevent: reset\ndata: {"connection":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","windowMs":60000,"selector":{"environment":"prod"},"continuity":"unknown","replay":false}\n\n'}}}
tail['x-sse-events']={'reset':ref('TailReset'),'batch':ref('TailBatch'),'gap':ref('TailGap')}
tail['x-schema-coverage']='typed-events'
tail['description']+=' Rolling 60-second event-time window; polling every 2 seconds; four local slots; 60-second Go lifetime; 5-second write deadlines. Proxy upstream socket timeout is 45 seconds. Disconnect releases slot. Last-Event-ID is not replay support. Gap frames after headers retain HTTP 200; abort can prevent a final frame.'
# Python-only API: keep these out of the direct-Go server entry.
legacy_props={'view':{'enum':list(json.loads((ROOT/'definitions/reservations.json').read_text())['views'])},'minutes':{'type':'integer','minimum':1,'maximum':120},'groupBy':{'enum':['release','region','scenario']},'reservationId':{'type':'string','maxLength':100},'scenario':{'enum':['','good','retry','late','missing','reject']},'spanFilter':{'type':'object','required':['fromMs','toMs','minMs','maxMs'],'properties':{k:{'type':'integer','minimum':0,'maximum':9007199254740991} for k in ('fromMs','toMs','minMs','maxMs')},'additionalProperties':False}}
schemas['LegacyQuery']={'type':'object','properties':legacy_props,'additionalProperties':False}
for method,path in [('get','/package'),('get','/example-questions'),('get','/prebuilts'),('post','/query'),('post','/explain')]:
 op={'operationId':method+'_python_'+path[1:].replace('-','_'),'summary':method.upper()+' '+path,'tags':['Python lab'],'servers':[spec['servers'][0]],'x-python-source':'server.py','x-effects':'source-read' if path=='/query' else 'pure-plan' if path=='/explain' else 'definition-read','x-schema-coverage':'response-envelope-pending','responses':{'200':{'description':'Existing Python lab response; detailed result schema pending.','content':{'application/json':{'schema':{}}}},'400':{'description':'Invalid input.','content':{'application/json':{'schema':ref('Error')}}},'502':{'description':'Source/proxy failure.','content':{'application/json':{'schema':ref('Error')}}}}}
 if method=='post':op['requestBody']={'required':True,'content':{'application/json':{'schema':ref('LegacyQuery')}}};op['x-browser-body-limit-bytes']=8192
 spec['paths'][path]={method:op}

schemas['ExampleQuestions']={'type':'array','items':{'type':'object','required':['name','yaml'],'properties':{'name':{'type':'string'},'yaml':{'type':'string'}}}}
schemas['LegacyPackage']={'type':'object','required':['apiVersion','id','views'],'additionalProperties':True,'description':'Legacy demo package declaration; grammar-specific attributes are preserved.'}
schemas['LegacyResult']={'type':'object','required':['expression','sql','parameters','evaluatedAt','evidence','limit'],'properties':{'expression':ref('LegacyQuery'),'sql':{'type':'string'},'parameters':{'type':'object','additionalProperties':{'type':'string'}},'evaluatedAt':{'type':'integer'},'evidence':{'const':'demo-unverified'},'limit':{'type':['integer','null']},'rows':{'type':'array','items':{'type':'object','additionalProperties':True}},'summary':{'type':['object','null']},'possiblyTruncated':{'type':'boolean'}}}
for method,path,model in [('get','/package','LegacyPackage'),('get','/prebuilts','PrebuiltRegistry'),('get','/example-questions','ExampleQuestions'),('post','/query','LegacyResult'),('post','/explain','LegacyResult')]:
 op=spec['paths'][path][method];op['responses']['200']['content']['application/json']['schema']=ref(model);op['responses']['200']['description']='Current Python lab response; record rows remain adapter-shaped objects.';op['x-schema-coverage']='typed-envelope'
ui_config={'operationId':'get_python_ui_config','summary':'UI startup configuration','tags':['Python configuration'],'servers':[spec['servers'][0]],'x-python-source':'server.py','x-effects':'configuration-only','x-schema-coverage':'typed','responses':{'200':{'description':'Pure startup metadata; no source probe.','content':{'application/json':{'schema':ref('UIConfig')}}}}}
spec['paths']['/ui-config']={'get':ui_config}
for method,path in [('get','/package'),('get','/example-questions'),('get','/prebuilts'),('post','/query'),('post','/explain')]:
 spec['paths'][path][method]['responses']['503']={'description':'Optional lab example capability is unconfigured; no source read.','content':{'application/json':{'schema':ref('RuntimeCapabilityRefusal')}}}
health={'operationId':'get_python_health','summary':'ClickHouse lab health probe','tags':['Python lab'],'servers':[{'url':'http://localhost:8088','description':'Python root (no /api prefix)'}],'x-python-source':'server.py','x-effects':'source-read','x-schema-coverage':'typed','responses':{}}
for status in ('200','503'):health['responses'][status]={'description':'Readonly SELECT 1 probe; does not establish ingestion completeness.','content':{'application/json':{'schema':{'type':'object','required':['ok'],'properties':{'ok':{'type':'boolean'},'error':{'type':'string'}}}}}}
spec['paths']['/health']={'get':health}
(ROOT/'api/openapi.json').write_text(json.dumps(spec,indent=2,ensure_ascii=False)+'\n')
print(f'Built OpenAPI: {sum(len(p) for p in spec["paths"].values())} operations, {len(schemas)} schemas')

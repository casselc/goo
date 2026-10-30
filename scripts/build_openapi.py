"""Assemble the reviewed HTTP surface plus reflected Go wire schemas.
Run after go run ./scripts/openapi-schema api/go-schemas.json.
"""
import json
import copy
import re
from pathlib import Path
from go_http_routes import route_sources

ROOT = Path(__file__).resolve().parents[1]
ref = lambda name: {'$ref': '#/components/schemas/' + name}
schemas = json.loads((ROOT / 'api/go-schemas.json').read_text())
for name in ('workbench.ActivityView','workbench.ActivityViewInput'):
 schemas[name]['properties']['baselineCapture']['description']='Optional exact immutable raw-span duration capture ID. Publication verifies stored evidence and same environment; opening a view stages configuration without loading evidence or starting telemetry. Omitted/empty preserves legacy identity.'

schemas['annotations.OperationSpan']['properties']['reportedStatus']['description']='Optional exact exporter StatusCode, independent of declared outcome. Omission means not retained in this evidence; empty and future values remain explicit unknown values. Legacy saved bytes are unchanged.'

search=schemas['exploration.CaptureSearchPage']['properties']
search['items']={'type':'array','maxItems':20,'items':ref('exploration.CaptureSummary')}
search['version']['const']='goo/capture-search/v2'
summary=schemas['exploration.CaptureSummary']
summary['description']='Bounded derived discovery metadata, not a partial receipt. No full question or evidence. Exact-ID reads verify authoritative contents. Environment null means unknown, distinct from exact empty.'
summary['additionalProperties']=False
sp=summary['properties']
for name in ('id','parent'):sp[name]['pattern']='^[a-f0-9]{64}$'
sp['title']['maxLength']=120
sp['capturedAt']['format']='date-time'
sp['kind']['enum']=['explore','inventory','comparison','explanation','measurement']
sp['measurementKind']['enum']=['gauge','counter','histogram','lifecycle','stock','state']
sp['signal']['enum']=['','logs','spans','metrics']
sp['environment']['anyOf'][0]['maxLength']=128
for name in ('matched','total'):search[name].update({'minimum':0,'maximum':10000})
search['order']['enum']=['newest','oldest']
# Effective-environment adoption is a distinct readonly contract, not Query.
for name in ('exploration.AdoptionWindowInput','exploration.AdoptionRecordsQuestionInput'):
 props=schemas[name]['properties'];props['signal']['enum']=['logs','spans'];props['fromMs']['minimum']=0;props['toMs']['maximum']=4102444800000
schemas['exploration.AdoptionRecordsQuestionInput']['properties']['limit'].update(minimum=1,maximum=100)
scope=schemas['identity.EnvironmentScopeInput'];scope['properties']['mode']['enum']=['named','empty','unreported'];scope['properties']['value'].update(maxLength=128)
scope['allOf']=[{'if':{'properties':{'mode':{'const':'named'}},'required':['mode']},'then':{'required':['value'],'properties':{'value':{'minLength':1}}},'else':{'properties':{'value':{'const':''}}}}]
schemas['identity.EnvironmentScope']['properties']['mode']['enum']=['named','empty','unreported']
for name,version in [('exploration.EnvironmentScopesResult','goo/environment-scopes/v1'),('exploration.AdoptionRecordsResult','goo/adoption-records/v1')]:
 schemas[name]['properties']['version']['const']=version;schemas[name]['properties']['readAt']['format']='date-time'
schemas['exploration.EnvironmentScopesResult']['properties']['groups']['maxItems']=64
schemas['exploration.EnvironmentScopeGroup']['properties']['count'].update(pattern='^[1-9][0-9]*$',maxLength=20)
schemas['exploration.AdoptionRecordsResult']['properties']['rows']['maxItems']=100

selector=schemas['identity.EnvironmentSelector']
selector['additionalProperties']=False
selector['properties']['version']['const']='goo/environment-selector/v1'
selector['properties']['policy']['enum']=['current-only','effective']
selector['properties']['scope']={'type':'object','additionalProperties':False,'required':['mode'],'properties':{'mode':{'type':'string','enum':['named','empty','unreported']},'value':{'type':'string','maxLength':128,'description':'At most 128 UTF-8 bytes; runtime enforces the byte limit.'}},'allOf':[{'if':{'properties':{'mode':{'const':'named'}},'required':['mode']},'then':{'required':['value'],'properties':{'value':{'minLength':1}}},'else':{'properties':{'value':{'const':''}}}}]}
selector['allOf']=[{'if':{'properties':{'policy':{'const':'current-only'}},'required':['policy']},'then':{'properties':{'scope':{'properties':{'mode':{'const':'named'}}}}}}]
schemas['exploration.AdoptionRecordsResult']['properties']['selector']={'allOf':[{'$ref':'#/components/schemas/identity.EnvironmentSelector'},{'properties':{'policy':{'const':'effective'}}}]}

schemas['identity.EnvironmentSelectorInput']=copy.deepcopy(selector)
for wire in ('exploration.Query','exploration.QueryInput','exploration.TailSelector','exploration.TailSelectorInput'):
 shape=schemas[wire];input_wire=wire.endswith('Input');target='identity.EnvironmentSelectorInput' if input_wire else 'identity.EnvironmentSelector'
 if input_wire:shape['required']=[key for key in shape.get('required',[]) if key!='environment']
 shape['oneOf']=[{'required':['environment'],'properties':{'environment':{'type':'string','minLength':1},'environmentSelector':{'type':'null'}}},{'required':['environmentSelector'],'properties':{'environment':{'const':''},'environmentSelector':{'$ref':'#/components/schemas/'+target}}}]
 shape['description']='Choose the historical current-only environment name or a versioned environmentSelector with an empty/omitted name. No implicit wildcard. Reported-edge record questions do not support the selector yet.'
summary['description']='Bounded discovery metadata, not a receipt. Legacy environment null means unknown or an explicitly versioned selector. environmentSelector preserves policy/scope; legacy name filters do not match selector captures.'
summary['properties']['environmentSelector']['description']='Exact versioned policy/scope, separate from legacy current-only environment. Read with version 250 or newer.'


# Immutable selected observations, independent of source and annotation storage.
keep=schemas['exploration.KeepTailIntentInput']
keep['required']=['version','requestID','connection','selector','refs']
keep['properties']['version']={'const':'goo/keep-tail/v1'}
for name in ('requestID','connection'):keep['properties'][name]={'type':'string','pattern':'^[a-f0-9]{64}$'}
keep['properties']['refs'].update(type='array',minItems=1,maxItems=20)
receipt=schemas['exploration.KeptObservations']
receipt['description']='Exact selected server observations, not a population capture or classification. Atomic immutable local file; retries reuse the same intent-bound request ID, including after proof expiry. No telemetry reread or XTDB required.'
receipt['properties']['version']['const']='goo/kept-observations/v1'
for name in ('id','digest','intentDigest','connection'):receipt['properties'][name]['pattern']='^[a-f0-9]{64}$'
receipt['properties']['members'].update(type='array',minItems=1,maxItems=20)

keepRecords=schemas['exploration.KeepRecordsIntentInput']
keepRecords['required']=['version','requestID','indexes']
keepRecords['properties']['version']={'const':'goo/keep-records/v1'}
for name in ('requestID','captureID','resultToken'):keepRecords['properties'][name]={'type':'string','pattern':'^[a-f0-9]{64}$'}
keepRecords['properties']['indexes']={'type':'array','minItems':1,'maxItems':20,'uniqueItems':True,'items':{'type':'integer','minimum':0,'maximum':9999}}
keepRecords['oneOf']=[{'required':['resultToken'],'not':{'required':['captureID']}},{'required':['captureID'],'not':{'required':['resultToken']}}]
keptRecords=schemas['exploration.KeptRecords'];keptRecords['properties']['version']['const']='goo/kept-records/v1'
for name in ('id','intentDigest','digest','captureID'):keptRecords['properties'][name]['pattern']='^[a-f0-9]{64}$'
keptRecords['properties']['origin']['enum']=['result-page','saved-capture']
keptRecords['properties']['members'].update(type='array',minItems=1,maxItems=20)
keptRecords['properties']['returnedRecords'].update(minimum=1,maximum=1000)
schemas['exploration.KeepRecordExpectationInput']['required']=['question','basis']
schemas['exploration.KeepRecordExpectationInput']['properties']['basis']={'type':'string','pattern':'^[a-f0-9]{64}$'}
keptRecords['oneOf']=[{'properties':{'origin':{'const':'saved-capture'}},'required':['captureID','captureAt']},{'properties':{'origin':{'const':'result-page'},'returnedRecords':{'maximum':100}},'not':{'anyOf':[{'required':['captureID']},{'required':['captureAt']}]}}]
schemas['KeptResource']={'oneOf':[ref('exploration.KeptObservations'),ref('exploration.KeptRecords')]}

keptSearch=schemas['exploration.KeptSearchPage'];keptSearch['properties']['version']['const']='goo/kept-search/v1'
keptSearch['properties']['items'].update(type='array',maxItems=20)
for name in ('total','matched'):keptSearch['properties'][name].update(minimum=0,maximum=1000)
keptSummary=schemas['exploration.KeptSummary'];keptSummary['additionalProperties']=False
keptSummary['properties']['observations'].update(minimum=1,maximum=20)
keptSummary['properties']['origin']['enum']=['tail-selection','result-page','saved-capture']
keptSummary['properties']['signal']['enum']=['logs','spans']
for name in ('id','digest'):keptSummary['properties'][name]['pattern']='^[a-f0-9]{64}$'
keptSummary['oneOf']=[{'required':['environment'],'not':{'required':['environmentSelector']}},{'required':['environmentSelector'],'not':{'required':['environment']}}]
keptSearch['description']='Bounded derived metadata only. Exact reopen independently verifies records. Twenty selections per page,1000 physical root entries; cursor binds filters and metadata state. Missing/corrupt metadata refuses rather than truncating. Repair is explicit bounded source-free maintenance.'
repairPage=schemas['exploration.KeptRepairPage']['properties']
repairPage['indexed'].update(minimum=0,maximum=64);repairPage['bytes'].update(minimum=0,maximum=33554432)
for name in ('completed','total'):repairPage[name].update(minimum=0,maximum=1000)
schemas['exploration.KeptRepairRequestInput']['properties']['cursor'].update(maxLength=1024)

schemas['KeptSearchError']={'type':'object','required':['error','repairAvailable'],'properties':{'error':{'type':'string'},'repairAvailable':{'type':'boolean','description':'True only when explicit index repair may address discovery failure; false for capacity and cursor refusals.'}},'additionalProperties':False}
schemas['Error'] = {'type': 'object', 'required': ['error'], 'properties': {'error': {'type': 'string'}, 'hint': {'type': 'string'}}, 'additionalProperties': True}
schemas['Cancellation'] = {'type': 'object', 'required': ['state'], 'properties': {'state': {'const': 'cancellation-requested'}}}
# Canonical supplied display snapshots: exact field maps, with exhaustive
# ranked-group and interval projections owned by the Go source data path.
for name in ('ViewDataset','ViewScope','ViewField','ViewRow','ViewQuality','ViewCompleteness','ViewInterval','ViewPartition'):
 schemas['exploration.'+name]['additionalProperties']=False
exact_count={'type':'string','pattern':'^(0|[1-9][0-9]*)$','maxLength':20,'description':'Exact canonical uint64 decimal; Go enforces maximum18446744073709551615. Plot coordinates never replace this value.'}
partition=schemas['exploration.ViewPartition']
partition['properties']['kind']={'const':'disjoint-record-groups'}
partition['properties']['measure']={'const':'count'}
partition['properties']['total']=copy.deepcopy(exact_count)
partition['description']='Exact disjoint captured-record partition from the same exhaustive GROUP BY read, including missing and empty groups; never the separately read count or record page. Runtime verifies the row-count sum and source-adoption authority before full-population shares. Zero total makes shares unavailable. The descriptor does not prove complete upstream collection.'
epoch_ms={'type':'string','pattern':'^(0|[1-9][0-9]*)$','maxLength':16,'description':'Exact nonnegative epoch milliseconds, Date-safe; Go checked queries are bounded before2100.'}
category={'type':['string','null'],'maxLength':1024,'description':'Null is missing; empty string is present empty. Runtime enforces1024 UTF-8 bytes; Schema maxLength counts characters.'}
rational_value={'type':'string','pattern':'^(0|-?[1-9][0-9]*)(/[1-9][0-9]*)?$','maxLength':1024,'description':'Exact canonical reduced rational or integer in declared human units. Runtime rejects non-reduced fractions and denominator1; no float conversion defines the value.'}
measure_specs={}
for family,kinds in [('numeric',['mean','sum','min','max']),('duration',['mean','min','max','p50','p95','p99'])]:
 for kind in kinds:
  quantile={'p50':'0.5','p95':'0.95','p99':'0.99'}.get(kind)
  measure_specs[family+'.'+kind]={'family':family,'kind':'quantile' if quantile else kind,'method':'quantilesExactHigh' if quantile else 'exact-rational' if kind=='mean' else 'exact',**({'quantile':quantile} if quantile else {})}
measure_values={key:{**rational_value,'type':['string','null']} for key in measure_specs}
for key in measure_values:
 if key.startswith('duration.'):
  measure_values[key]['pattern']='^(0|[1-9][0-9]*)(/[1-9][0-9]*)?$' if key=='duration.mean' else '^(0|[1-9][0-9]*)$'
  if key!='duration.mean':measure_values[key]['maxLength']=20
for name in ('ViewAggregation','ViewCell','ViewSupport'):schemas['exploration.'+name]['additionalProperties']=False
aggregation=schemas['exploration.ViewAggregation'];agg_props=aggregation['properties']
agg_props['family']['enum']=['numeric','duration'];agg_props['kind']['enum']=['mean','sum','min','max','quantile'];agg_props['method']['enum']=['exact-rational','exact','quantilesExactHigh'];agg_props['quantile']['enum']=['0.5','0.95','0.99']
aggregation['oneOf']=[{'properties':{key:{'const':value} for key,value in spec.items()},**({'not':{'required':['quantile']}} if 'quantile' not in spec else {'required':['quantile']})} for spec in measure_specs.values()]
# Different keys may share identical aggregation descriptions; de-duplicate the
# candidates so oneOf expresses semantic alternatives rather than field names.
aggregation['oneOf']=list({json.dumps(x,sort_keys=True):x for x in aggregation['oneOf']}.values())
support=schemas['exploration.ViewSupport']
for key in ('total','valid','missing','empty','invalid'):support['properties'][key]=copy.deepcopy(exact_count)
support['description']='Exact supporting-record partition. Runtime verifies total=valid+missing+empty+invalid and total equals the row record count.'
cell=schemas['exploration.ViewCell'];cell['properties']['state']['enum']=['derived','estimated','no_data','no_valid_values','unknown']
cell['allOf']=[
 {'if':{'properties':{'state':{'const':'unknown'}},'required':['state']},'then':{'not':{'required':['support']}},'else':{'required':['support'],'properties':{'support':{'type':'object'}}}},
 {'if':{'properties':{'state':{'enum':['derived','estimated']}},'required':['state']},'then':{'properties':{'support':{'properties':{'valid':{**exact_count,'pattern':'^[1-9][0-9]*$'}}}}}},
 {'if':{'properties':{'state':{'const':'no_data'}},'required':['state']},'then':{'properties':{'support':{'properties':{key:{'const':'0'} for key in ('total','valid','missing','empty','invalid')}}}}},
 {'if':{'properties':{'state':{'const':'no_valid_values'}},'required':['state']},'then':{'properties':{'support':{'properties':{'valid':{'const':'0'},'total':{**exact_count,'pattern':'^[1-9][0-9]*$'}}}}}}
]
schemas['exploration.ViewValues']={'type':'object','additionalProperties':{'type':['string','null'],'maxLength':1024},'description':'Exact strings or explicit missing values. Dataset fields and shape constrain every cell; no vendor-specific opaque payload.'}
view=schemas['exploration.ViewDataset'];view_props=view['properties']
view_props['partition']=ref('exploration.ViewPartition')
view['description']='Supplied immutable captured-record snapshot: ranked groups, explicit intervals or population summaries with optional exact numeric/duration measures. An optional ranked partition derives its denominator from the same exhaustive group read. Result-aspect completeness never certifies raw-record paging, upstream ingestion/sampling or a transactional count/page pair. Scope/revision equality, partition sums and interval coverage are runtime checked.'
view_props['version']['const']='goo.dataset/1';view_props['shape']['enum']=['ranked','time-series','summary']
view_props['id']['pattern']='^[a-f0-9]{64}:(groups|intervals|summary)$';view_props['revision']['pattern']='^[a-f0-9]{64}$';view_props['capturedAt']['format']='date-time'
view_props['origin']['enum']=['source-read','browser-held']
view_props['sourceBasis'].update(minLength=1,maxLength=128,description='Unverified caller-claimed source basis, separate from the content-hashed held revision. Not a signed receipt.')
view_props['fields']={'type':'array','minItems':2,'maxItems':13,'items':ref('exploration.ViewField')}
view_props['rows']={'type':'array','maxItems':1000,'items':ref('exploration.ViewRow')}
view_props['dimensions']={'type':'array','minItems':1,'maxItems':2,'uniqueItems':True,'items':{'enum':['category','category1']}}
view_scope=schemas['exploration.ViewScope']['properties']
for name in ('id','generation'):view_scope[name]['pattern']='^[a-f0-9]{64}$'
view_scope['signal']['enum']=['logs','spans'];view_scope['selector']={**ref('exploration.Query'),'description':'Exact applied checked question including environment policy, window, predicates and grouping.'}
view_field=schemas['exploration.ViewField']['properties'];view_field['key']['enum']=['category','category1','count','fromMs','toMs']+list(measure_specs);view_field['type']['enum']=['string','number','timestamp','rational']
schemas['exploration.ViewField']['allOf']=[{'if':{'properties':{'key':{'const':key}},'required':['key']},'then':{'required':['aggregation','unit','nullable'],'properties':{'nullable':{'const':True},'type':{'const':'rational'},'unit':{'const':'ns'} if spec['family']=='duration' else {'type':'string','minLength':1},'aggregation':{'type':'object','properties':{k:{'const':v} for k,v in spec.items()}}}}} for key,spec in measure_specs.items()]
schemas['exploration.ViewField']['allOf'].append({'if':{'properties':{'key':{'enum':['category','category1','count','fromMs','toMs']}},'required':['key']},'then':{'not':{'required':['aggregation']}}})
view_quality=schemas['exploration.ViewQuality']['properties'];view_quality['extent']['enum']=['complete','unknown']
for name,value in [('basis','captured-query'),('value','derived'),('sampling','unknown'),('freshness','fixed')]:view_quality[name]['const']=value
view_quality['sampling']['description']='Upstream sampling and collection completeness remain unverified even for complete result aspects.'
view_quality['extent']['description']='Only groups/intervals as named by shape; disagreement or unproved coverage remains unknown.'
view_complete=schemas['exploration.ViewCompleteness']['properties'];view_complete['basis']['const']='captured-query';view_complete['aspects']={'type':'array','minItems':1,'maxItems':1,'items':{'enum':['groups','intervals','summary']}}
for name in ('revision','scopeId'):view_complete[name]['pattern']='^[a-f0-9]{64}$'
schemas['exploration.ViewCompleteness']['description']='Bound to enclosing revision/scope ID; runtime enforces equality. Intervals require exhaustive contiguous known-state coverage. Never inferred from page length or a separate count.'
interval_props=schemas['exploration.ViewInterval']['properties']
for name in ('fromMs','toMs'):interval_props[name]=copy.deepcopy(epoch_ms)
interval_props['bucketMs']={'type':'string','enum':['1000','10000','60000','300000']}
row=schemas['exploration.ViewRow'];row['properties']['state']['enum']=['observed','no_data','unknown']
row['properties']['key'].update(minLength=1,maxLength=128,description='Stable snapshot row identity; never a source-record member reference.')
row['properties']['cells']={'type':'object','maxProperties':10,'propertyNames':{'enum':list(measure_specs)},'additionalProperties':ref('exploration.ViewCell')}
row['allOf']=[]
for key in measure_specs:
 row['allOf'].append({'if':{'properties':{'values':{'required':[key]}},'required':['values']},'then':{'required':['cells'],'properties':{'cells':{'required':[key]}}}})
 row['allOf'].append({'if':{'properties':{'cells':{'required':[key]}},'required':['cells']},'then':{
  'properties':{'values':{'required':[key]}},
  'allOf':[{'if':{'properties':{'cells':{'properties':{key:{'properties':{'state':{'enum':['derived','estimated']}}}}}}},
   'then':{'properties':{'values':{'properties':{key:rational_value}}}},
   'else':{'properties':{'values':{'properties':{key:{'type':'null'}}}}}}]}})


def view_field_spec(key,kind,unit=None,nullable=False):
 props={'key':{'const':key},'type':{'const':kind}}
 required=[]
 if unit:props['unit']={'const':unit};required.append('unit')
 if nullable:props['nullable']={'const':True};required.append('nullable')
 else:props['nullable']={'const':False}
 return {'allOf':[ref('exploration.ViewField'),{'properties':props,'required':required}]}
ranked=[]
for dims in (['category'],['category','category1']):
 fields=[view_field_spec(key,'string',nullable=True) for key in dims]+[view_field_spec('count','number','records')]
 values={'type':'object','required':dims+['count'],'properties':{**{key:category for key in dims},'count':exact_count,**measure_values},'additionalProperties':False}
 ranked_row={'allOf':[ref('exploration.ViewRow'),{
  'not':{'required':['state']},
  'properties':{'key':{'pattern':'^category:[a-f0-9]{64}$'},'values':values}}]}
 ranked.append({'required':['dimensions'],'properties':{
  'dimensions':{'const':dims},
  'fields':{'minItems':len(fields),'maxItems':len(fields)+10,'prefixItems':fields,'items':{'allOf':[ref('exploration.ViewField'),{'properties':{'key':{'enum':list(measure_specs)}}}]}},
  'rows':{'items':ranked_row}}})
time_values={'type':'object','required':['fromMs','toMs'],'properties':{'fromMs':epoch_ms,'toMs':epoch_ms,'count':{'type':['string','null']},**measure_values},'additionalProperties':False}
time_states=[{'properties':{'state':{'const':'observed'},'values':{'required':['count'],'properties':{'count':{**exact_count,'pattern':'^[1-9][0-9]*$'}}}}},{'properties':{'state':{'const':'no_data'},'values':{'required':['count'],'properties':{'count':{'const':'0'}}}}},{'properties':{'state':{'const':'unknown'},'values':{'properties':{'count':{'type':'null'}}}}}]
time_fields=[view_field_spec('fromMs','timestamp','ms'),view_field_spec('toMs','timestamp','ms'),view_field_spec('count','number','records',nullable=True)]
ranked_shape={'properties':{'shape':{'const':'ranked'},'id':{'pattern':'^[a-f0-9]{64}:groups$'},'completeness':{'properties':{'aspects':{'const':['groups']}}}},'not':{'required':['interval']},'oneOf':ranked}
time_row={'allOf':[ref('exploration.ViewRow'),{
 'required':['state'],
 'properties':{'key':{'pattern':'^interval:[0-9]+:[0-9]+$'},'values':time_values},
 'oneOf':time_states}]}
time_shape={'required':['interval'],'not':{'anyOf':[{'required':['dimensions']},{'required':['partition']}]},'properties':{
 'shape':{'const':'time-series'},'id':{'pattern':'^[a-f0-9]{64}:intervals$'},'interval':{'type':'object'},
 'fields':{'minItems':3,'maxItems':13,'prefixItems':time_fields,'items':{'allOf':[ref('exploration.ViewField'),{'properties':{'key':{'enum':list(measure_specs)}}}]}},
 'rows':{'maxItems':240,'items':time_row},
 'completeness':{'properties':{'aspects':{'const':['intervals']}}}}}
summary_values={'type':'object','required':['count'],'properties':{'count':exact_count,**measure_values},'additionalProperties':False}
summary_row={'allOf':[ref('exploration.ViewRow'),{'not':{'required':['state']},'properties':{'key':{'const':'population'},'values':summary_values}}]}
summary_fields=[view_field_spec('count','number','records')]
summary_shape={'not':{'anyOf':[{'required':['dimensions']},{'required':['interval']},{'required':['partition']}]},'properties':{
 'shape':{'const':'summary'},'id':{'pattern':'^[a-f0-9]{64}:summary$'},
 'fields':{'minItems':1,'maxItems':11,'prefixItems':summary_fields,'items':{'allOf':[ref('exploration.ViewField'),{'properties':{'key':{'enum':list(measure_specs)}}}]}},
 'rows':{'minItems':1,'maxItems':1,'items':summary_row},
 'completeness':{'properties':{'aspects':{'const':['summary']}}}}}
view_props['fields']['minItems']=1
view['oneOf']=[ranked_shape,time_shape,summary_shape]
view['allOf']=[{'if':{'properties':{'quality':{'properties':{'extent':{'const':'complete'}},'required':['extent']}},'required':['quality']},'then':{'required':['completeness'],'properties':{'completeness':{'type':'object'},'rows':{'items':{'properties':{'state':{'enum':['observed','no_data']}}}}}},'else':{'properties':{'completeness':{'type':'null'}}}}]

view['allOf'].extend([
 {'if':{'properties':{'origin':{'const':'browser-held'}},'required':['origin']},'then':{'not':{'required':['completeness']},'properties':{'shape':{'const':'summary'},'quality':{'properties':{'extent':{'const':'unknown'}}}}}},
 {'if':{'properties':{'origin':{'const':'source-read'}},'required':['origin']},'then':{'required':['capturedAt']}}
])
for name in ('exploration.Cohort','exploration.CohortInput'):
 if name in schemas and 'capturedAt' in schemas[name]['properties']:
  schemas[name]['properties']['capturedAt']={'type':'string','description':'Optional unverified caller-held capture clock. Omission/empty means unknown; never replaced with current time. Nonempty RFC3339 is runtime checked.'}
change=schemas['exploration.ViewChange'];change['additionalProperties']=False
change['properties']['state']['enum']=['derived','unavailable'];change['properties']['value']={**rational_value,'type':['string','null']}
change['properties']['unit'].update(minLength=1,maxLength=128)
change['allOf']=[{'if':{'properties':{'state':{'const':'derived'}},'required':['state']},'then':{'properties':{'value':rational_value}},'else':{'required':['reason'],'properties':{'reason':{'minLength':1},'value':{'type':'null'}}}}]
comparison=schemas['exploration.ViewComparison'];comparison['additionalProperties']=False
comparison['description']='Independent browser-held summary snapshots with exact selection-minus-baseline differences. No combined population, normalized rate, signed source authority, pooled percentile or invented shared capture clock.'
cmp_props=comparison['properties'];cmp_props['version']['const']='goo.comparison/1';cmp_props['state']['enum']=['observed','no_data','diagnostic','incompatible']
for side in ('baseline','selection'):
 cmp_props[side]={'allOf':[ref('exploration.ViewDataset'),{'properties':{'shape':{'const':'summary'},'origin':{'const':'browser-held'}},'required':['origin']}]}
cmp_props['reason'].update(minLength=1,maxLength=128)
cmp_props['changes']={'type':'object','minProperties':1,'maxProperties':11,'propertyNames':{'enum':['count']+list(measure_specs)},'additionalProperties':ref('exploration.ViewChange')}
comparison['allOf']=[{'if':{'properties':{'state':{'enum':['diagnostic','incompatible']}},'required':['state']},'then':{'properties':{'changes':{'additionalProperties':{'properties':{'state':{'const':'unavailable'}}}}}},'else':{'properties':{'changes':{'properties':{'count':{'properties':{'state':{'const':'derived'}}}},'required':['count']}}}}]
for key in ['count']+list(measure_specs):
 has_field={'anyOf':[{'properties':{side:{'properties':{'fields':{'contains':{'properties':{'key':{'const':key}},'required':['key']}}}}},'required':[side]} for side in ('baseline','selection')]}
 comparison['allOf'].append({'if':has_field,'then':{'properties':{'changes':{'required':[key]}}},'else':{'properties':{'changes':{'not':{'required':[key]}}}}})
 if key=='count':
  delta={'properties':{'unit':{'const':'records'}},'not':{'required':['aggregation']}}
 else:
  spec=measure_specs[key]
  delta={'required':['aggregation'],'properties':{'aggregation':{'type':'object','properties':{k:{'const':v} for k,v in spec.items()}},**({'unit':{'const':'ns'}} if spec['family']=='duration' else {})}}
 comparison['allOf'].append({'properties':{'changes':{'properties':{key:delta}}}})

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
# Fractional counters are explicit; omitted domain preserves integer contracts.
for name in ('measurements.Definition','measurements.DefinitionInput'):
 schemas[name]['properties']['counterDomain'].update({'enum':['nonnegative-decimal'],'description':'Optional per-stream counter-only exact exported decimal domain. Omitted keeps integer counter semantics. Host authority requires this opt-in; Pinned CPU-time defaults use cumulative seconds; explicitly expanded network/disk profiles certify whole byte counts below 2^53 and retain cumulative temporality.'})
for name in ('metrics.CounterResult',):
 schemas[name]['properties']['valueDomain'].update({'enum':['nonnegative-decimal'],'description':'Explicit bounded nonnegative decimal arithmetic. Omitted retains uint64 input/count semantics.'})
schemas['metrics.CounterInterval']['properties']['delta'].update({'maxLength':1024,'description':'Exact canonical base-ten interval difference, present only for ok intervals. Decimal differences can be longer than individual input values.'})
schemas['metrics.CounterInterval']['properties']['exactPerSecond'].update({'maxLength':2048,'pattern':'^(0|[1-9][0-9]*)(/[1-9][0-9]*)?$','description':'Reduced exact nonnegative rational rate for opted-in decimal counters, in source units per second. perSecond is its finite floating-point display approximation.'})
schemas['metrics.CounterResult']['description']='Per-stream native counter intervals. Explicit decimal domain adds exact decimal deltas/rational rates; count ratios/fleet and window-comparison contracts do not accept that domain. Unknown intervals have no rate; resets, gaps, missing values and clipped boundaries are not interpolated.'
# Additive cumulative nonmonotonic sum observations, distinct from counter rates.
schemas['metrics.StockResult']['description']='Per-stream cumulative nonmonotonic sum observations. Signed finite fractional values and accumulation epochs are preserved; decreases are not resets. No delta integration, fleet total, rate, capacity or utilization is inferred.'
for name in ('metrics.StockSample','metrics.StockSeries'):
 schemas[name]['properties']['state']['enum']=['observed','stale','unknown']
stock_sample=schemas['metrics.StockSample']
stock_sample['properties']['value']={'type':'number','description':'Finite observed cumulative tally, possibly signed/fractional. Omitted when unknown; stale observations retain their value.'}
stock_sample['properties']['atNS']['pattern']='^(0|[1-9][0-9]*)$'
stock_sample['allOf']=[{'if':{'properties':{'state':{'enum':['observed','stale']}},'required':['state']},'then':{'required':['value']}},{'if':{'properties':{'state':{'const':'unknown'}},'required':['state']},'then':{'not':{'required':['value']}}}]
# Categorical native states preserve observation quality separately from domain codes.
for name in ('measurements.Definition','measurements.DefinitionInput'):
 schemas[name]['properties']['stateProfile']['description']='Required only for kind state: pinned k8scluster-pod-phase/0.137.0 or k8scluster-container-ready/0.137.0 with matching exact resource authority. Unitless receiver gauges; no numeric arithmetic, probe-existence or health inference.'
schemas['metrics.StateResult']['description']='Exact-stream categorical observations. state is observation quality; codeState is supported/unsupported domain decoding. Fresh reported phase Unknown is observed. Unsupported finite codes retain raw code and quality without a label.'
for name in ('metrics.StateSample','metrics.StateSeries'):
 schemas[name]['properties']['state']['enum']=['observed','stale','unknown']
schemas['metrics.StateSample']['properties']['codeState']['enum']=['supported','unsupported']
schemas['metrics.StateSample']['properties']['label']['enum']=['Pending','Running','Succeeded','Failed','Unknown','Not ready','Ready']
# stockDomain is optional; legacy generic stocks keep signed fractional semantics.
for name in ('measurements.Definition','measurements.DefinitionInput'):
 schemas[name]['properties']['stockDomain']['description']='Optional stock-only nonnegative-integer policy. Recorded values must be integers >=0 and <2^53; missing-value flags remain unknown. Pinned filesystem stocks additionally require reported cumulative nonmonotonic export and strict source metadata. Empty/omitted keeps generic signed finite fractional stock semantics.'
schemas['measurements.Definition']['properties']['stockDomain']['const']='nonnegative-integer'
schemas['measurements.Result']['allOf']=[{'if':{'properties':{'question':{'properties':{'definition':{'properties':{'stockDomain':{'const':'nonnegative-integer'}},'required':['stockDomain']}},'required':['definition']}},'required':['question']},'then':{'properties':{'stocks':{'properties':{'series':{'type':'array','items':{'properties':{'samples':{'type':'array','items':{'properties':{'value':{'type':'integer','minimum':0,'maximum':9007199254740991}}}}}}}}}}}}]
pool=schemas['measurements.PoolSignalsPlan']['properties']
pool['version']['const']='goo/pool-signals/v1';pool['state']['enum']=['candidates','unconfirmed']
for name in ('capacity','wait'):
 pool[name]={'type':'array','maxItems':2,'items':{'$ref':'#/components/schemas/measurements.PoolSignalOption'}}
for name in ('capacityOmitted','waitOmitted','streamsOmitted'):pool[name]['minimum']=0
pool['streamsOmitted']['maximum']=160000
schemas['measurements.PoolSignalOption']['properties']['key']['pattern']='^[a-f0-9]{64}$'
schemas['measurements.PoolSignalsRequestInput']['properties']['instruments']['anyOf'][0]['maxItems']=16
schemas['measurements.PoolSignalsPlan']['description']='Pure supplied-metadata planning. Capacity/wait options are independent exact-stream questions, not fresh evidence or alignment/health/utilization claims. Preserves parent bounds; no source read, custody or publication.'
view=schemas['measurements.PoolView']['properties']
view['version']['const']='goo/pool-view/v1'
schemas['measurements.PoolView']['additionalProperties']=False
schemas['measurements.PoolView']['description']='Pure composition of replay-validated held native evidence. Full identity, fixed bounds and observed stock times/epochs govern rational utilization; separate read clocks remain visible. No source read, custody, capture or transactional observation is created.'
util=schemas['measurements.PoolViewUtilization'];util['properties']['state']['enum']=['observed','unavailable','inconsistent']
util['properties']['numerator']['pattern']='^(0|[1-9][0-9]*)$';util['properties']['denominator']['pattern']='^[1-9][0-9]*$'
util['allOf']=[{'if':{'properties':{'state':{'const':'observed'}},'required':['state']},'then':{'required':['numerator','denominator']},'else':{'not':{'anyOf':[{'required':['numerator']},{'required':['denominator']}]}}}]
evidence=schemas['measurements.PoolViewEvidence']['properties'];evidence['basis']['pattern']='^[a-f0-9]{64}$';evidence['readAt']['format']='date-time'
schemas['measurements.ResourceBinding']['properties']['kind']['enum']=['host','kubernetes-pod','kubernetes-container']
schemas['measurements.ResourceBindingInput']['properties']['kind']={'type':'string','enum':['host','kubernetes-pod','kubernetes-container']}
schemas['measurements.ResourceBinding']['properties']['id'].update({'minLength':1,'maxLength':512,'description':'Exact reported host.id, k8s.pod.uid or container.id for the selected resource kind; never a name, normalized runtime ID or manufactured service/instance identity.'})
host=schemas['measurements.HostDiscovery']['properties'];host['version']['const']='goo/host-discovery/v1';host['hosts']['anyOf'][0]['maxItems']=100
schemas['measurements.HostDiscovery']['description']='One bounded host-bearing metadata read over gauge/sum/histogram tables. Raw full resource maps/schemas remain separate identities. Unidentified source rows are not queried/countable; decoder missingHostIDRecords does not establish source absence. Metadata is not health/completeness.'
for name in ('records','missingHostIDRecords','hostsOmitted'):host[name]['minimum']=0
plan=schemas['measurements.HostDashboardPlan']['properties'];plan['version']['const']='goo/host-dashboard/v1';plan['state']['enum']=['candidates','unconfirmed'];plan['widgets']['anyOf'][0]['maxItems']=2;plan['candidates']['anyOf'][0]['maxItems']=16
for name in ('candidatesOmitted','streamsOmitted','instrumentsOmitted'):plan[name]['minimum']=0
schemas['measurements.HostWidget']['properties']['state']['const']='candidate'
schemas['measurements.HostWidget']['properties']['grain']['const']='host-exact-stream'
schemas['measurements.HostEntry']['properties']['key']['pattern']='^[a-f0-9]{64}$'
schemas['measurements.HostEntry']['properties']['resource'].update({'required':['host.id'],'properties':{'host.id':{'type':'string','minLength':1,'maxLength':512}}})
schemas['measurements.HostEntry']['properties']['instruments']['anyOf'][0]['maxItems']=16
schemas['measurements.HostWidget']['properties']['key']['pattern']='^[a-f0-9]{64}$'
schemas['measurements.HostWidget']['properties']['question']={'allOf':[{'$ref':'#/components/schemas/measurements.Question'},{'properties':{'definition':{'properties':{'source':{'required':['resource','exactStream'],'properties':{'resource':{'type':'object'},'exactStream':{'type':'object'},'services':{'type':['array','null'],'maxItems':0},'namespace':{'type':'null'}}}}}}}]}
schemas['measurements.HostDashboardPlan']['description']='Pure retained-metadata planning. At most two independently evaluated default questions, with up to16 explicit exact-stream candidates. CPU logical cores/modes and memory states remain separate. No whole-host/fleet aggregation, derived capacity, saturation or health is inferred.'
pod=schemas['exploration.PodDiscovery']['properties'];pod['version']['const']='goo/pod-discovery/v1';pod['observations']['anyOf'][0]['maxItems']=100;pod['sourceGroups'].update({'minimum':0,'maximum':1000});pod['observationsOmitted'].update({'minimum':0,'maximum':900});pod['basis']['pattern']='^[a-f0-9]{64}$';pod['readAt']['format']='date-time'
schemas['exploration.PodDiscovery']['description']='One bounded UNION of five source-family resource-map/schema groups, not physical pod inventory or Kubernetes API state. More than1000groups refuses the entire request;100resource variants retained with omissions. No native authority, health, capacity, ownership or exact-map record join is inferred.'
observation=schemas['exploration.PodResourceObservation'];observation['properties']['key']['pattern']='^[a-f0-9]{64}$'
for name in ('podUIDState','clusterUIDState'):observation['properties'][name]['enum']=['present','empty','missing']
observation['properties']['families']['anyOf'][0]['maxItems']=5
observation['properties']['resourceSchemaURL']['description']='Explicit null for unrecorded trace-table resource schema; string including empty for known log/metric schema. These states remain distinct identities.'
observation['allOf']=[{'if':{'properties':{'podUIDState':{'const':'present'}},'required':['podUIDState']},'then':{'properties':{'actionable':{'const':True}}},'else':{'properties':{'actionable':{'const':False}}}}]
family=schemas['exploration.PodFamilyObservation']['properties'];family['family']['enum']=['logs','spans','gauge','sum','histogram'];family['count']['pattern']='^[1-9][0-9]*$'
for name in ('firstNS','lastNS'):family[name]['pattern']='^(0|[1-9][0-9]*)$'
# Runtime metadata is encoded by small Go maps; schemas are authored here.
schemas['RuntimeCapabilityState']={'type':'object','required':['state'],'properties':{'state':{'enum':['configured','unconfigured']}},'additionalProperties':False}
schemas['RuntimeCapabilities']={'type':'object','required':['version','profile','preparedHTTP','declaredOperations'],'properties':{'version':{'const':'goo/runtime-capabilities/v1'},'profile':{'enum':['lab','generic']},'preparedHTTP':{'$ref':'#/components/schemas/RuntimeCapabilityState'},'declaredOperations':{'$ref':'#/components/schemas/RuntimeCapabilityState'}},'additionalProperties':False,'description':'Source-free startup configuration only. Configured does not establish source availability, authorization, telemetry coverage, ingestion completeness or service health.'}
schemas['RuntimeCapabilityRefusal']={'type':'object','required':['error','code','capability'],'properties':{'error':{'type':'string'},'code':{'const':'capability_unconfigured'},'capability':{'enum':['prepared-http','declared-operations','example-grammar','example-catalog']}},'additionalProperties':False}
schemas['UIConfig']={'type':'object','required':['version','profile','examples'],'properties':{'version':{'const':'goo/ui-config/v1'},'profile':{'enum':['lab','generic']},'examples':{'$ref':'#/components/schemas/RuntimeCapabilityState'}},'additionalProperties':False,'description':'Pure Go startup configuration. Does not query or report readiness of ClickHouse.'}
# Additive dashboard profiles opt in; omitted/empty/null retain the legacy plan.
schemas['exploration.DashboardRequestInput']['properties']['profile']['description']='Optional auto, all-spans, http-server, rpc-server, db-client, messaging-consumer or messaging-producer selection. Omitted/empty/null preserves the legacy v2 response. auto selects HTTP only with a positive supplied server-observation hint plus current-key metadata, or a compatible native HTTP instrument. Automatic precedence is HTTP, legacy RPC, DB client, messaging consumer, messaging producer, then All spans. RPC admission requires both legacy rpc.system/rpc.method keys plus a positive rpcServerObserved hint, or an eligible legacy RPC duration instrument. Flat keys alone leave All spans selected; hints are advisory, not authenticated conformance.'
schemas['exploration.DashboardRequestInput']['properties']['httpServerObserved']['description']='Optional advisory positive hint from already-held, same-scope span observations. false/omitted means no positive hint, not HTTP absence; no new probe or semantic-convention compliance claim is implied.'
schemas['exploration.DashboardRequestInput']['properties']['rpcServerObserved']['description']='Optional advisory positive hint from already-held, same-scope Server span observations with nonempty legacy rpc.system and rpc.method on the same row. false/omitted is unknown, not RPC absence or conformance. HTTP admission has automatic precedence; this does not authorize a probe.'
schemas['exploration.DashboardRequestInput']['properties']['dbClientObserved']['description']='Optional advisory positive hint from already-held, same-scope Client span observations with a nonempty current db.system.name on the same row. false/omitted is unknown, not database absence or health. Auto prefers eligible HTTP then RPC before DB; no new source probe is implied.'
schemas['exploration.DashboardRequestInput']['properties']['messagingConsumerObserved']['description']='Optional advisory positive hint from already-held, same-scope Consumer spans with nonempty messaging.system on the same row. false/omitted is unknown, not absence, delivery success, acknowledgement or broker health. A compatible processing-duration histogram can independently admit the consumer family; it remains a separate population.'
schemas['exploration.DashboardRequestInput']['properties']['messagingProducerObserved']['description']='Optional advisory positive hint from already-held, same-scope Producer spans with nonempty messaging.system on the same row. false/omitted is unknown, not absence or delivery success. Consumer has automatic precedence; processing metrics never admit the producer family.'
schemas['exploration.DashboardPlan']['properties']['profile']['description']='Selected all-spans, http-server, rpc-server, db-client, messaging-consumer or messaging-producer family for opt-in requests; absent on legacy requests.'
schemas['exploration.DashboardPlan']['properties']['nativeOptionsOmitted'].update({'minimum':0,'description':'Optional positive count of supported native options excluded by the 128-option display bound after exact-stream expansion. Omitted/zero preserves legacy wire bytes.'})
schemas['exploration.DashboardPlan']['properties']['runtimeStreamsOmitted'].update({'minimum':0,'description':'Optional positive count of runtime preview streams clipped beyond first25 plus supplied bounded discovery omissions. This is not source absence or a count of validated executable options.'})
schemas['exploration.DashboardPlan']['properties']['families']['description']='Opt-in bounded family descriptors. HTTP/RPC/DB/messaging candidate metadata must be confirmed by the executable question before treating counts as observed; no hidden source read or fallback is performed by this planner.'
schemas['exploration.DashboardFamily']['properties']['state']['description']='Metadata capability state, including candidate. Not a service-health verdict, observed server count or semantic-convention conformance assertion.'
schemas['exploration.DashboardWidget']['properties']['family']['description']='Optional display/planning family identity; not a telemetry field or a shared native/span/log denominator.'
# API policy is deliberately authored, not inferred from HTTP verbs.
for name in ('measurements.PodMetricDiscoveryQuestion','measurements.PodMetricDiscoveryQuestionInput'):
 question=schemas[name];question['required']=sorted(set(question.get('required',[])+['resourceSchemaURL']))
 question['properties']['resourceSchemaURL']={'type':'string','maxLength':1024,'description':'Required reported resource schema including known empty string; unknown/null trace schema is refused before source access.'}
for name,version in (('measurements.PodMetricDiscovery','goo/pod-metric-discovery/v1'),('measurements.PodDashboardPlan','goo/pod-dashboard/v1')):
 schemas[name]['properties']['version']['const']=version
podplan=schemas['measurements.PodDashboardPlan']['properties'];podplan['widgets']['anyOf'][0]['maxItems']=2;podplan['candidates']['anyOf'][0]['maxItems']=16;podplan['state']['enum']=['candidates','unconfirmed']
podwidget=schemas['measurements.PodMetricWidget']['properties'];podwidget['state']['const']='candidate';podwidget['grain']['enum']=['pod-resource-stream','container-resource-stream'];podwidget['profile']['enum']=['collector-kubeletstats-v0.137.0','otel-k8s-v1.37.0','collector-k8scluster-v0.137.0','container-memory-kubeletstats-v0.137.0','container-memory-k8scluster-v0.137.0','container-cpu-kubeletstats-v0.137.0','container-cpu-k8scluster-v0.137.0']
podwidget['role']={'type':'string','enum':['working-set','usage','request','limit']}
podwidget['question']={'allOf':[{'$ref':'#/components/schemas/measurements.Question'},{'properties':{'definition':{'properties':{'kind':{'enum':['gauge','state']},'source':{'required':['resource','exactStream'],'properties':{'resource':{'type':'object','properties':{'kind':{'enum':['kubernetes-pod','kubernetes-container']}}},'exactStream':{'type':'object'},'services':{'type':['array','null'],'maxItems':0},'namespace':{'type':'null'}}}}}}}]}
podwidget_schema=schemas['measurements.PodMetricWidget']
podwidget_schema['allOf']=[{'if':{'properties':{'profile':{'const':'collector-k8scluster-v0.137.0'}},'required':['profile']},'then':{'properties':{'question':{'properties':{'definition':{'required':['stateProfile'],'properties':{'kind':{'const':'state'},'stateProfile':{'enum':['k8scluster-pod-phase/0.137.0','k8scluster-container-ready/0.137.0']}}}}}}},'else':{'properties':{'question':{'properties':{'definition':{'properties':{'kind':{'const':'gauge'}}}}}}}}]
# Exact container readiness is a separate authority/grain, never a pod alias.
podwidget_schema['allOf'].append({'if': {'anyOf':[{'properties':{'role':{'enum':['working-set','usage','request','limit']}},'required':['role']},{'properties':{'question':{'properties':{'definition':{'properties':{'stateProfile':{'const':'k8scluster-container-ready/0.137.0'}},'required':['stateProfile']}},'required':['definition']}},'required':['question']}]}, 'then':{'properties':{'grain':{'const':'container-resource-stream'},'question':{'properties':{'definition':{'properties':{'source':{'properties':{'resource':{'properties':{'kind':{'const':'kubernetes-container'}}}}}}}}}}},'else':{'properties':{'grain':{'const':'pod-resource-stream'},'question':{'properties':{'definition':{'properties':{'source':{'properties':{'resource':{'properties':{'kind':{'const':'kubernetes-pod'}}}}}}}}}}}})

podwidget_schema['allOf'].append({'if':{'required':['role']},'then':{'properties':{'profile':{'enum':['container-memory-kubeletstats-v0.137.0','container-memory-k8scluster-v0.137.0','container-cpu-kubeletstats-v0.137.0','container-cpu-k8scluster-v0.137.0']},'question':{'properties':{'definition':{'properties':{'kind':{'const':'gauge'}}}}}}}})
podwidget_schema['allOf'].append({'if':{'properties':{'profile':{'enum':['container-memory-kubeletstats-v0.137.0','container-memory-k8scluster-v0.137.0','container-cpu-kubeletstats-v0.137.0','container-cpu-k8scluster-v0.137.0']}},'required':['profile']},'then':{'required':['role']}})
poddiscovery=schemas['measurements.PodMetricDiscovery']['properties'];poddiscovery['instruments']['anyOf'][0]['maxItems']=16
for name in ('records','instrumentsOmitted'):poddiscovery[name].update({'minimum':0,'maximum':10000})


reads = {'/environment-scopes','/environment-scopes/records','/pod-metric-discovery','/host-discovery','/pod-discovery','/http-package', '/operations', '/counter-rates', '/histograms', '/inventory', '/discovery', '/explore', '/explore/counted-values', '/record-aggregates', '/trace-search', '/population/explain', '/metric-discovery', '/measurements', '/investigations', '/captures', '/tail', '/query'}
for profile,roles in [('container-memory-kubeletstats-v0.137.0',['working-set']),('container-memory-k8scluster-v0.137.0',['request','limit']),('container-cpu-kubeletstats-v0.137.0',['usage']),('container-cpu-k8scluster-v0.137.0',['request','limit'])]:
 podwidget_schema['allOf'].append({'if':{'properties':{'profile':{'const':profile}},'required':['profile']},'then':{'properties':{'role':{'enum':roles}}}})
plans = {'/pod-dashboard/plan','/host-dashboard/plan','/pool-signals/plan','/pool-signals/view','/container-memory/view','/container-cpu/view','/container-resources/history','/measurement-packages/plan', '/telemetry-packages/draft', '/telemetry-packages/numeric-plan', '/telemetry-packages/plan', '/recipes', '/dashboards/plan', '/population/plan', '/population/explain/plan', '/population/explain/filter', '/numeric-histograms/selection', '/numeric-histograms/comparison-selection', '/compare', '/explain'}
publications = {'/kept-observations/records','/kept-observations/tail','/activity-views','/measurement-packages', '/telemetry-packages', '/annotations', '/exploration-views', '/deployments', '/documents', '/captures/comparisons'}
requests = {'/kept-observations/repair':'KeptRepairRequest','/kept-observations/records':'KeepRecordsIntent','/kept-observations/tail':'KeepTailIntent','/pod-metric-discovery':'PodMetricDiscoveryQuestion','/pod-dashboard/plan':'PodDashboardRequest','/host-discovery':'HostDiscoveryQuestion','/pod-discovery':'PodDiscoveryQuestion','/host-dashboard/plan':'HostDashboardRequest','/pool-signals/plan':'PoolSignalsRequest','/pool-signals/view':'PoolViewRequest','/container-memory/view':'ContainerMemoryViewRequest','/container-cpu/view':'ContainerCPUViewRequest','/container-resources/history':'ContainerResourceHistoryRequest','/activity-views':'ActivityViewRequest',
 '/retained-evidence/tail':'TailIntent','/compare':'CompareRequest','/inventory':'InventoryRequest','/discovery':'InventoryRequest','/explore':'RecordQuery','/explore/counted-values':'RecordQuery','/record-aggregates':'RecordQuery','/trace-search':'TraceSearchRequest','/tail':'TailSelector',
 '/environment-scopes':'AdoptionWindow','/environment-scopes/records':'AdoptionRecordsQuestion','/measurements':'NativeQuestion','/metric-discovery':'MetricDiscoveryRequest','/measurement-packages/plan':'NativePlanRequest','/measurement-packages':'YAMLRequest',
 '/telemetry-packages':'YAMLRequest','/telemetry-packages/draft':'TelemetryDraftRequest','/telemetry-packages/numeric-plan':'NumericPlanRequest','/telemetry-packages/plan':'TelemetryPlanRequest',
 '/recipes':'RecipeRequest','/dashboards/plan':'DashboardRequest','/population/plan':'PopulationRequest','/numeric-histograms/selection':'NumericHistogramSelectionRequest','/numeric-histograms/comparison-selection':'DistributionSelectionRequest','/annotations':'MeaningInput','/exploration-views':'ExplorationView','/deployments':'DeploymentInput','/documents':'DocumentRequest',
 '/population/explain':'ExplanationRequest','/population/explain/plan':'ExplanationRequest','/population/explain/filter':'ExplanationFilterRequest','/captures':'CaptureRequest','/captures/live':'LiveComparisonRequest','/captures/comparisons':'SaveComparisonRequest','/investigations':'InvestigationRequest','/retained-evidence':'RetainRequest','/retained-evidence/{id}/annotations':'CorrectRequest',
}
responses = {('get','/kept-observations/search'):'KeptSearchPage',('post','/kept-observations/repair'):'KeptRepairPage',('post','/kept-observations/records'):'KeptRecords',('post','/kept-observations/tail'):'KeptObservations',('get','/kept-observations/{id}'):'KeptResource',('post','/pod-metric-discovery'):'PodMetricDiscovery',('post','/pod-dashboard/plan'):'PodDashboardPlan',('post','/host-discovery'):'HostDiscovery',('post','/pod-discovery'):'PodDiscovery',('post','/host-dashboard/plan'):'HostDashboardPlan',('post','/pool-signals/plan'):'PoolSignalsPlan',('post','/pool-signals/view'):'PoolView',('post','/container-memory/view'):'ContainerMemoryView',('post','/container-cpu/view'):'ContainerCPUView',('post','/container-resources/history'):'ContainerResourceHistory',('get','/capabilities'):'RuntimeCapabilities',('post','/population/explain'):'ExplanationResult',('post','/population/explain/plan'):'ExplanationPlan',('post','/population/explain/filter'):'ExplanationStagedQuestion',('post','/activity-views'):'ActivityView',('get','/activity-views/{id}'):'ActivityView',('get','/activity-views'):'ActivityViewList',
 ('post','/compare'):'CompareResult',('post','/inventory'):'InventoryResult',('post','/discovery'):'ServiceDiscovery',('post','/explore'):'RecordResult',('post','/explore/counted-values'):'CountedValuesResult',('post','/record-aggregates'):'RecordAggregatesResult',('post','/trace-search'):'TraceSearchResult',
 ('post','/measurements'):'NativeResult',('post','/metric-discovery'):'MetricDiscoveryResult',('post','/telemetry-packages/draft'):'PackagePlan',('post','/telemetry-packages/plan'):'PackagePlan',('post','/telemetry-packages/numeric-plan'):'NumericPlan',('post','/telemetry-packages'):'SavedPackage',('get','/telemetry-packages/{revision}'):'SavedPackage',
 ('post','/recipes'):'RecipePlan',('post','/dashboards/plan'):'DashboardPlan',('post','/population/plan'):'PopulationPlan',('post','/numeric-histograms/selection'):'NumericHistogramSelection',('post','/numeric-histograms/comparison-selection'):'NumericHistogramSelection',
 ('get','/http-package'):'LiveHTTPResult',('get','/operations'):'LiveOperations',('get','/counter-rates'):'LiveRates',('get','/histograms'):'LiveHistograms',
 ('get','/captures'):'CaptureList',('get','/captures/search'):'CaptureSearchPage',('post','/captures'):'CapturePage',('get','/captures/{id}'):'CapturePage',('get','/investigations/{id}'):'InvestigationCapture',('post','/investigations'):'InvestigationCapture',
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
# Pure container memory composition preserves independent evidence and exact rational percentages.
cm=schemas['measurements.ContainerMemoryView']['properties'];cm['version']['const']='goo/container-memory-view/v1'
cm['signals']={'type':'array','minItems':3,'maxItems':3,'items':{'$ref':'#/components/schemas/measurements.ContainerMemorySignal'}}
signal=schemas['measurements.ContainerMemorySignal'];signal['additionalProperties']=False
signal['properties']['role']['enum']=['working-set','request','limit'];signal['properties']['state']['enum']=['observed','stale','unknown','invalid','no_data','unavailable']
signal['properties']['value']['pattern']='^(0|[1-9][0-9]*)$'
signal['properties']['observationState']['enum']=['observed','stale','unknown']
signal['properties']['basis']['pattern']='^[a-f0-9]{64}$';signal['properties']['readAt']['format']='date-time'
percentage=schemas['measurements.ContainerMemoryPercentage'];percentage['additionalProperties']=False
percentage['properties']['state']['enum']=['observed','unavailable']
percentage['properties']['value']['pattern']='^(0|[1-9][0-9]*)(/[1-9][0-9]*)?$'
percentage['allOf']=[{'if':{'properties':{'state':{'const':'observed'}},'required':['state']},'then':{'required':['value','text']}},{'if':{'properties':{'state':{'const':'unavailable'}},'required':['state']},'then':{'required':['reason'],'not':{'anyOf':[{'required':['value']},{'required':['text']}]}}}]
schemas['measurements.ContainerMemoryView']['additionalProperties']=False
schemas['measurements.ContainerMemoryView']['description']='Replay-validated held container gauge evidence; independently useful signals, strict exact-resource/schema/attribute/observation/freshness alignment for rational percentages. No source reads, new custody, inferred health or missing-as-zero semantics.'
# Pure container CPU composition preserves independent evidence and fractional cores.
cm=schemas['measurements.ContainerCPUView']['properties'];cm['version']['const']='goo/container-cpu-view/v1'
cm['signals']={'type':'array','minItems':3,'maxItems':3,'items':{'$ref':'#/components/schemas/measurements.ContainerCPUSignal'}}
signal=schemas['measurements.ContainerCPUSignal'];signal['additionalProperties']=False
signal['properties']['role']['enum']=['usage','request','limit'];signal['properties']['state']['enum']=['observed','stale','unknown','invalid','no_data','unavailable']
signal['properties']['value'].update(pattern=r'^(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$',maxLength=512)
signal['properties']['observationState']['enum']=['observed','stale','unknown']
signal['properties']['basis']['pattern']='^[a-f0-9]{64}$';signal['properties']['readAt']['format']='date-time'
percentage=schemas['measurements.ContainerCPUPercentage'];percentage['additionalProperties']=False
percentage['properties']['state']['enum']=['observed','unavailable']
percentage['properties']['value']['pattern']='^(0|[1-9][0-9]*)(/[1-9][0-9]*)?$'
percentage['allOf']=[{'if':{'properties':{'state':{'const':'observed'}},'required':['state']},'then':{'required':['value','text']}},{'if':{'properties':{'state':{'const':'unavailable'}},'required':['state']},'then':{'required':['reason'],'not':{'anyOf':[{'required':['value']},{'required':['text']}]}}}]
schemas['measurements.ContainerCPUView']['additionalProperties']=False
schemas['measurements.ContainerCPUView']['description']='Replay-validated held container gauge evidence; independently useful signals, strict exact-resource/schema/attribute/observation/freshness alignment for rational percentages. No source reads, new custody, inferred health or missing-as-zero semantics.'
schemas['measurements.ContainerCPUPercentage']['properties']['value']['maxLength']=1024
schemas['measurements.ContainerCPUPercentage']['properties']['text']['maxLength']=1024
# Held resource history is a pure observation projection, never a continuous signal.
history=schemas['measurements.ContainerResourceHistory'];history['additionalProperties']=False
history['properties']['version']['const']='goo/container-resources-history/v1'
history['properties']['series'].update(maxItems=6)
hr=schemas['measurements.ContainerResourceHistoryRequestInput'];hr['anyOf']=[{'required':['memory'],'properties':{'memory':{'not':{'type':'null'}}}},{'required':['cpu'],'properties':{'cpu':{'not':{'type':'null'}}}}]
hs=schemas['measurements.ContainerResourceHistorySeries'];hs['additionalProperties']=False
hs['properties']['key']['enum']=['memory/working-set','memory/request','memory/limit','cpu/usage','cpu/request','cpu/limit']
hs['properties']['family']['enum']=['memory','cpu'];hs['properties']['role']['enum']=['working-set','usage','request','limit']
hs['properties']['unit']['enum']=['By','{cpu}'];hs['properties']['state']['enum']=['unavailable','no_data','observations']
hs['properties']['points'].update(maxItems=10000)
hs['allOf']=[{'if':{'properties':{'family':{'const':'memory'}},'required':['family']},'then':{'properties':{'key':{'enum':['memory/working-set','memory/request','memory/limit']},'role':{'enum':['working-set','request','limit']},'unit':{'const':'By'},'points':{'items':{'properties':{'value':{'pattern':'^(0|[1-9][0-9]*)$'}}}}}},'else':{'properties':{'key':{'enum':['cpu/usage','cpu/request','cpu/limit']},'role':{'enum':['usage','request','limit']},'unit':{'const':'{cpu}'}}}},{'if':{'properties':{'state':{'const':'observations'}},'required':['state']},'then':{'required':['question','readAt','basis','fromNS','toNS','maxGapNS'],'properties':{'points':{'minItems':1}}},'else':{'properties':{'points':{'maxItems':0}}}}]
hp=schemas['measurements.ContainerResourceHistoryPoint'];hp['additionalProperties']=False
hp['properties']['state']['enum']=['observed','stale','unknown','invalid'];hp['properties']['observationState']['enum']=['observed','stale','unknown']
hp['properties']['atNS'].update(pattern='^(0|[1-9][0-9]*)$',maxLength=20)
hp['properties']['value'].update(pattern=r'^(0|[1-9][0-9]*)(\.[0-9]*[1-9])?$',maxLength=512)
hp['properties']['rawRowIndices'].update(minItems=1,maxItems=10000,uniqueItems=True,items={'type':'integer','minimum':0,'maximum':9999})
hp['allOf']=[{'if':{'properties':{'state':{'enum':['unknown','invalid']}},'required':['state']},'then':{'not':{'required':['value']}}},{'if':{'properties':{'state':{'enum':['observed','stale']}},'required':['state']},'then':{'required':['value']}}]
# Native registry responses are envelopes rather than Definition directly.
schemas['NativePackage'] = {'type':'object','required':['definition','revision','yaml'],'properties':{'definition':ref('measurements.Definition'),'revision':{'type':'string'},'yaml':{'type':'string'}}}
schemas['NativePlan'] = {'allOf':[ref('NativePackage'),{'type':'object','required':['question','plan'],'properties':{'question':ref('measurements.Question'),'plan':ref('measurements.Plan')}}]}
responses['post','/environment-scopes']='EnvironmentScopesResult'
responses['post','/environment-scopes/records']='AdoptionRecordsResult'
responses['post','/measurement-packages'] = 'NativePackage'
responses['post','/measurement-packages/plan'] = 'NativePlan'
schemas['NativePackageList']={'type':'array','maxItems':256,'items':{'allOf':[ref('NativePackage'),{'type':'object','required':['builtin'],'properties':{'builtin':{'type':'boolean'}}}]}}
responses['get','/measurement-packages']='NativePackageList'
responses['get','/measurement-packages/revisions/{revision}']='NativePackage'
# These roots are reached by reflected response types; fail loudly if renamed.
for name in ['measurements.Definition','measurements.Question','measurements.Plan']:
 if name not in schemas: raise ValueError(f'Missing reflected type: {name}')

def go_limit(path):
 if path in ('/environment-scopes','/environment-scopes/records'): return 4096
 if path in ('/container-memory/view','/container-cpu/view','/container-resources/history'): return 8388608
 if path == '/pool-signals/view': return 2097152
 if path in ('/compare','/numeric-histograms/comparison-selection'): return 1048576
 if path == '/dashboards/plan': return 262144
 if path == '/pod-dashboard/plan': return 8389632
 if path.startswith(('/telemetry-packages','/measurement-packages')) or path in ('/pod-metric-discovery','/host-dashboard/plan','/pool-signals/plan','/population/explain','/population/explain/plan','/population/explain/filter','/measurements','/metric-discovery','/trace-search'): return 65536
 if path in ('/annotations','/exploration-views','/deployments','/documents','/population/plan', '/numeric-histograms/selection','/captures/comparisons'): return 131072
 return 16384

def effect(method,path):
 if path in ('/capabilities','/ui-config'): return 'configuration-only'
 if path == '/health': return 'liveness-only'
 if path in ('/example-questions','/prebuilts'): return 'embedded-definition-read'
 if path == '/queries/{id}/cancel': return 'cancel-capability'
 if path == '/captures': return 'source-read-and-capture-or-held-explanation-publication'
 if path == '/captures/live': return 'source-read-or-held-draft-publication'
 if path.startswith('/retained-evidence'): return 'configuration-only' if path.endswith('/status') else 'xtdb-read' if method=='get' else 'xtdb-publication-or-recovery'
 if method=='post' and path in plans: return 'pure-plan-or-supplied-summary'
 if path in reads and (method=='post' or path in ('/http-package','/operations','/counter-rates','/histograms')):
  return 'source-read-with-possible-custody-or-capture' if path in ('/measurements','/captures','/investigations') else 'source-read'
 if path=='/kept-observations/repair': return 'bounded-derived-index-repair'
 if method=='post' and path in ('/kept-observations/tail','/kept-observations/records'): return 'file-publication-from-attested-selection'
 if method=='post' and path in publications: return 'file-publication-with-reference-validation'
 return 'saved-evidence-or-definition-read'

spec = {
 'openapi':'3.1.0',
 'info':{'title':'goo observability API','version':'0.1.0','description':'Current development API served by one Go process. Browser paths under /api and unprefixed API paths dispatch to the same handlers, admission controls and state. No proxy or JSON reserialization boundary. Wire-shape schemas describe encoding, not all domain validation. Semantic constraints, custody and source effects remain authoritative in Go. Unknown input fields are rejected; null and omitted fields can decode and then fail domain validation. No authentication is implemented by this lab; cancellation IDs are local capabilities, not authorization.'},
 'servers':[{'url':'/api','description':'Go browser API on this origin'},{'url':'/','description':'Same Go handlers without /api prefix'}],
 'paths':{},'components':{'schemas':schemas},
 'x-contract-notes':['Go accepts any request Content-Type; application/json is the supported client encoding.','Browser /api and unprefixed API requests have the same byte limits, headers, cancellation and streaming semantics.','JSON receive is bounded to 15 seconds and interrupted by request cancellation; successful decoding clears the receive deadline before source/response work.','Operation GET validation uses http.Error text/plain with JSON-looking text; /api preserves that status, type and body.','Tail early JSON errors omit Cache-Control; streaming responses set no-store and X-Accel-Buffering: no.','Health is process liveness only; it does not query ClickHouse or certify telemetry ingestion.','No generated handlers, semantic validator replacement, API stability guarantee or new authorization layer is introduced.'],
}
route_source=route_sources(ROOT)
for (method,path),(file,block) in sorted(route_source.items()):
 e=effect(method,path)
 op={'operationId':method+'_'+re.sub(r'[^a-zA-Z0-9]+','_',path).strip('_'),'summary':method.upper()+' '+path,'tags':[path.split('/')[1]],'description':f'Current owner: {file}. Effect: {e}. See docs/HTTP_BOUNDARIES.md for lifecycle and limits.','x-go-source':file,'x-effects':e,'parameters':[],'responses':{}}
 for p in re.findall(r'\{([^}]+)\}',path):op['parameters'].append({'name':p,'in':'path','required':True,'schema':{'type':'string'},'description':'Exact revision, receipt ID or process-local capability; not an authorization credential.'})
 # Enumerate explicit query getters; allowlists and cross-field validation stay in Go.
 for p in sorted(set(re.findall(r'(?:q|v|values|r\.URL\.Query\(\))\.(?:Get|Has)\("([^"]+)"',block))):
  op['parameters'].append({'name':p,'in':'query','schema':{'type':'string'},'description':'Single value; route-specific allowlists and cross-field rules apply. Targets are JSON-encoded typed objects where supported.'})
 if path=='/kept-observations/search':op['parameters']=[{'name':name,'in':'query','schema':({'type':'string','enum':['','tail-selection','result-page','saved-capture']} if name=='origin' else {'type':'string','enum':['','logs','spans']} if name=='signal' else {'type':'string','maxLength':1024})} for name in ('origin','signal','cursor')]
 for p in {'/captures':['cursor'],'/captures/{id}':['cursor'],'/captures/{id}/followups':['row'],'/captures/{id}/compare':['selection','distribution']}.get(path,[]):
  if method=='get' and not any(v['name']==p and v['in']=='query' for v in op['parameters']):op['parameters'].append({'name':p,'in':'query','schema':{'type':'string'},'description':'Single canonical value; saved evidence is not reread from the telemetry source.'})
 controlled=method=='post' and path in ('/environment-scopes','/environment-scopes/records','/metric-discovery','/pod-metric-discovery','/host-discovery','/pod-discovery','/measurements','/explore','/explore/counted-values','/record-aggregates','/trace-search','/population/explain','/inventory','/discovery','/captures','/captures/live')
 controlled = controlled or method=='get' and path in ('/http-package','/operations','/counter-rates','/histograms')
 if controlled:op['parameters'].append({'name':'X-Goo-Query-ID','in':'header','schema':{'type':'string','pattern':'^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$'},'description':'Optional short-lived cancellation capability; repeated/cancelled IDs are refused.'})
 if method=='post' and path in requests:
  op['requestBody']={'required':True,'content':{'application/json':{'schema':ref(requests[path])}}}
  op['x-go-body-limit-bytes']=go_limit(path)
 success='202' if path=='/queries/{id}/cancel' else '201' if method=='post' and (path in publications or path in ('/retained-evidence/tail','/captures','/investigations','/retained-evidence','/retained-evidence/{id}/annotations')) else '200'
 model=responses.get((method,path));body=ref(model) if model else {}
 op['x-schema-coverage']='typed' if model else 'response-envelope-pending'
 op['responses'][success]={'description':'Current success response. Empty/unknown evidence does not imply healthy or complete telemetry.','content':{'application/json':{'schema':body}}}
 codes = {int(n) for n in re.findall(r'(?:fail|send|reply)\(\s*w,\s*(\d{3})|fail\(\s*(\d{3})|WriteHeader\((\d{3})\)|code\s*:?=\s*(\d{3})',block) for n in n if n}
 codes.update(int(n) for n in re.findall(r'http.Error\(w,.*?,\s*(\d{3})\)',block))
 if 'source-read' in e or e.startswith('xtdb-'): codes.add(502)
 if method=='post': codes.add(400)
 if controlled: codes.update((400,409,429))
 if path in ('/environment-scopes','/environment-scopes/records'): codes.update((400,409,429,502))
 if path.startswith('/retained-evidence') and not path.endswith('/status'): codes.add(503)
 if method=='get' and path in ('/http-package','/operation-packages','/operations','/counter-rates','/histograms'): codes.add(503)
 if path=='/queries/{id}/cancel':codes.update((400,429))
 if path=='/captures/search':codes.update((400,500,503))
 if path in ('/measurement-packages','/measurement-packages/revisions/{revision}'): codes.add(400)
 if path=='/measurement-packages/revisions/{revision}':
  op['description']='Source-free exact native revision lookup. Verifies the saved hash-named YAML directly, even when catalog capacity is exceeded. Corrupt saved winners are refused before bounded builtin fallback. No query parameters, telemetry read, publication or full saved-catalog scan.'
  op['parameters'][0]['schema']={'type':'string','pattern':'^[a-f0-9]{64}$'}
 if path=='/captures/search':
  op['description']='Derived metadata search over immutable captures, literal case-insensitive title containment, optional exact kind/signal/environment, capture-time order. No telemetry reads or payload bodies. 10,000 nonreserved directory entries (two known index directories separately reserved), 16 MiB aggregate metadata, 64 KiB per metadata entry and 8 MiB retained summary JSON; missing index/budget refusal is 503. Metadata checksum and payload stat stamps do not establish payload content integrity; reopen and explicit rebuild verify full contents. Cursors bind query, admitted catalog IDs and metadata digests; changed scope/catalog is 400. Omitted environment matches any; present-empty matches exact empty. Direct ID reopen keeps its independent limits.'
  for p in op['parameters']:
   if p.get('name')=='order':p['schema']={'type':'string','enum':['newest','oldest'],'default':'newest'}
   if p.get('name')=='signal':p['schema']={'type':'string','enum':['','logs','spans','metrics']}
   if p.get('name')=='title':p['schema']={'type':'string','maxLength':120}
   if p.get('name')=='environment':p['schema']={'type':'string','maxLength':256}
   if p.get('name')=='kind':p['schema']={'type':'string','enum':['','explore','inventory','comparison','explanation','measurement']}
 for status in map(str,sorted(n for n in codes if n>=400)):
  op['responses'][status]={'description':'Existing route refusal. Semantic validation and effect ordering remain in the owner.','content':{'application/json':{'schema':ref('Error')}}}
 if method=='get' and path=='/kept-observations/search':op['responses']['503']['content']['application/json']['schema']=ref('KeptSearchError')
 if method=='get' and path in ('/http-package','/operation-packages','/operations','/counter-rates','/histograms'):
  op['responses']['503']['content']['application/json']['schema']=ref('RuntimeCapabilityRefusal')
  op['responses']['400']['content']['text/plain']={'schema':{'type':'string'},'example':'{"error":"unsupported query parameter"}\n'}
 if path=='/queries/{id}/cancel':
  op['responses']['202']['content']['application/json']['schema']=ref('Cancellation');op['x-schema-coverage']='typed'
  op['requestBody']={'required':False,'description':'Cancellation ignores the request body. An empty body or {} is supported.','content':{'application/json':{'schema':{},'example':{}}}}
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
tail['description']+=' Rolling 60-second event-time window; polling every 2 seconds; four local slots; 60-second Go lifetime; 5-second write deadlines. Disconnect releases slot. Last-Event-ID is not replay support. Gap frames after headers retain HTTP 200; abort can prevent a final frame.'
# Pure embedded metadata is registered by the same Go mux as observations.
schemas['ExampleQuestions']={'type':'array','items':{'type':'object','required':['name','yaml'],'properties':{'name':{'type':'string'},'yaml':{'type':'string'}},'additionalProperties':False}}
schemas['Liveness']={'type':'object','required':['ok'],'properties':{'ok':{'const':True}},'additionalProperties':False,'description':'The Go HTTP process is running; no source or ingestion check.'}
for path,model in [('/ui-config','UIConfig'),('/example-questions','ExampleQuestions'),('/prebuilts','PrebuiltRegistry'),('/health','Liveness')]:
 op=spec['paths'][path]['get']
 op['responses']['200']['content']['application/json']['schema']=ref(model)
 op['x-schema-coverage']='typed'
 op['responses']['400']={'description':'Query parameters are unsupported.','content':{'application/json':{'schema':ref('Error')}}}
 if path in ('/example-questions','/prebuilts'):
  op['responses']['503']={'description':'Optional example catalog is unconfigured; no source read.','content':{'application/json':{'schema':ref('RuntimeCapabilityRefusal')}}}
(ROOT/'api/openapi.json').write_text(json.dumps(spec,indent=2,ensure_ascii=False)+'\n')
print(f'Built OpenAPI: {sum(len(p) for p in spec["paths"].values())} operations, {len(schemas)} schemas')

"""Shape checks over generated complete continuation carriers.

Hash/scope equality and witness membership are Go/JavaScript semantics, not JSON
Schema proofs. These cases exercise the real enclosing API carriers as well as
request/result leaves and deliberately malformed wire envelopes.
"""
import copy
import json
import sys
from pathlib import Path
from jsonschema import Draft202012Validator, FormatChecker
ROOT=Path(__file__).resolve().parents[1]

def main():
    path=Path(sys.argv[1]) if len(sys.argv)>1 else ROOT/'artifacts/verification/trace-continuation403-carriers.json'
    cases=json.loads(path.read_text());spec=json.loads((ROOT/'api/openapi.json').read_text())
    def validator(name):
        return Draft202012Validator({'$ref':'#/components/schemas/'+name,'components':spec['components']},format_checker=FormatChecker())
    page_check=validator('exploration.CapturePage')
    result_check=validator('exploration.TraceSearchResult')
    request_check=validator('exploration.TraceSearchPageRequestInput')
    controls=0;pages=0
    for c in cases:
        page_check.validate(c['page']);result_check.validate(c['result'])
        position=c['result'].get('position') or c['result'].get('nextPosition')
        if position:
            request={'request':c['request'],'position':position};request_check.validate(request);pages+=1
            for mutation in ('null-position','bad-hash','bad-id','extra-position','request-cursor'):
                bad=copy.deepcopy(request)
                if mutation=='null-position':bad['position']=None
                elif mutation=='bad-hash':bad['position']['requestHash']='bad'
                elif mutation=='bad-id':bad['position']['lastTraceId']='bad'
                elif mutation=='extra-position':bad['position']['extra']=True
                else:bad['request']['position']=position
                assert not request_check.is_valid(bad),mutation
                controls+=1
        for mutation in ('missing-remaining','null-remaining','mixed-payload','null-receipt','null-position'):
            bad=copy.deepcopy(c['page'])
            if mutation=='missing-remaining':del bad['traceSearch']['remainingTraces']
            elif mutation=='null-remaining':bad['traceSearch']['remainingTraces']=None
            elif mutation=='mixed-payload':bad['explore']={}
            elif mutation=='null-receipt':bad['receipt']=None
            else:bad['receipt']['question']['traceSearchPosition']=None
            assert not page_check.is_valid(bad),mutation
            controls+=1
    print(f'{len(cases)} complete trace-page API carriers, {pages} page inputs, {controls} malformed-wire controls pass')

if __name__=='__main__':
    try:
        main()
    except Exception as error:
        if hasattr(error,'absolute_path'):
            location='/'.join(str(x) for x in error.absolute_path) or '<root>'
            print(f'{type(error).__name__}: {location}: schema rule {error.validator}',file=sys.stderr)
        else:
            print(f'{type(error).__name__}: {str(error)[:500]}',file=sys.stderr)
        raise SystemExit(1)

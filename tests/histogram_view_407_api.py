"""407 full carriers: wire admission only; runtime owns binding and support proof."""
import copy
import json
import sys
from pathlib import Path
from jsonschema import Draft202012Validator, FormatChecker
ROOT=Path(__file__).resolve().parents[1]

def main():
    corpus=Path(sys.argv[1]) if len(sys.argv)>1 else ROOT/'artifacts/verification/histogram407-carriers.json'
    vectors=json.loads(corpus.read_text());spec=json.loads((ROOT/'api/openapi.json').read_text());controls=0
    def check(name):
        return Draft202012Validator({'$ref':'#/components/schemas/'+name,'components':spec['components']},format_checker=FormatChecker())
    for v in vectors:
        for suffix in ('','Input'):
            validator=check('workbench.ExplorationView'+suffix);validator.validate(v['source'])
            bad=copy.deepcopy(v['source']);leaf=next(c['histogram'] for c in bad['composition']['children'] if 'histogram' in c);leaf['display']['normalization']='density'
            assert not validator.is_valid(bad);controls+=1
            bad=copy.deepcopy(v['source']);bad['queries'][0]['offset']=1
            assert not validator.is_valid(bad);controls+=1
        check('scopedview.Dataset').validate(v['dataset'])
        check('workbench.ViewTemplateExpansion').validate(v['expansion'])
        check('workbench.ViewTemplateExpansionRequestInput').validate({'source':v['source'],'query':v['destination']})
    print(f'{len(vectors)} complete histogram wire vectors; {controls} closed-display/root controls')

if __name__=='__main__':
    try: main()
    except Exception as error:
        if hasattr(error,'absolute_path'): print(f'{type(error).__name__}: {"/".join(map(str,error.absolute_path))}: {error.validator}',file=sys.stderr)
        else: print(f'{type(error).__name__}: {str(error)[:500]}',file=sys.stderr)
        raise SystemExit(1)

"""Rate carrier wire shapes; arithmetic and eligibility are Go/JS proof obligations."""
import copy
import json
import sys
from pathlib import Path
from jsonschema import Draft202012Validator, FormatChecker
ROOT = Path(__file__).resolve().parents[1]

def main():
    corpus = Path(sys.argv[1]) if len(sys.argv)>1 else ROOT/'artifacts/verification/record-rate406-carriers.json'
    vectors = json.loads(corpus.read_text())
    spec = json.loads((ROOT/'api/openapi.json').read_text())
    def check(name):
        return Draft202012Validator({'$ref':'#/components/schemas/'+name,'components':spec['components']},format_checker=FormatChecker())
    controls=0
    for v in vectors:
        for suffix in ('','Input'):
            view=check('workbench.ExplorationView'+suffix)
            for key in ('summary','timeline'):
                view.validate(v[key])
                bad=copy.deepcopy(v[key]);bad['composition']['children'][0][key]['display']['unexpected']=True
                assert not view.is_valid(bad);controls+=1
        check('workbench.ViewTemplateExpansion').validate(v['expansion'])
        check('workbench.ViewTemplateExpansionRequestInput').validate({'source':v['summary'],'query':v['destination']})
        # Complete exploration response includes both datasets and support facts.
        check('exploration.Result').validate(v['data'])
        for slot in ('summaryDataset','timeDataset'):
            for mutation in ('null-count','number-elapsed','unknown-support'):
                bad=copy.deepcopy(v['data']);facts=bad[slot]['rows'][0]['cells']['record.rate']['rate']
                if mutation=='null-count': facts['count']=None
                elif mutation=='number-elapsed': facts['elapsedMs']=1000
                else: facts['guess']=True
                assert not check('exploration.Result').is_valid(bad),mutation
                controls+=1
    print(f'{len(vectors)} full rate response/view/template wire vectors; {controls} malformed-wire controls')

if __name__=='__main__':
    try: main()
    except Exception as error:
        if hasattr(error,'absolute_path'):
            print(f'{type(error).__name__}: {"/".join(map(str,error.absolute_path))}: {error.validator}',file=sys.stderr)
        else: print(f'{type(error).__name__}: {str(error)[:500]}',file=sys.stderr)
        raise SystemExit(1)

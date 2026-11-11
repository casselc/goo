"""Input/output shapes for generated whole record-summary template carriers.

Destination compatibility, page-zero/profile eligibility, field availability and
source leases remain Go/JavaScript semantics. These checks do not infer them from
JSON Schema or from a standalone summary leaf.
"""
import copy
import json
import sys
from pathlib import Path
from jsonschema import Draft202012Validator, FormatChecker
ROOT = Path(__file__).resolve().parents[1]


def main():
    corpus = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / 'artifacts/verification/summary-template404-carriers.json'
    vectors = json.loads(corpus.read_text())
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    def check(name):
        return Draft202012Validator({'$ref': '#/components/schemas/' + name, 'components': spec['components']}, format_checker=FormatChecker())
    controls = 0
    for suffix in ('', 'Input'):
        view_check = check('workbench.ExplorationView' + suffix)
        binding_check = check('workbench.ViewTemplateBinding' + suffix)
        for v in vectors:
            view_check.validate(v['source'])
            view_check.validate(v['expansion']['view'])
            binding_check.validate(v['source']['template'])
            for mutation in ('null-template', 'bad-group', 'bad-profile', 'null-baseline', 'unknown-binding', 'null-summary', 'unknown-summary', *(('snapshot', 'basis') if suffix == 'Input' else ())):
                bad = copy.deepcopy(v['source'])
                if mutation in ('snapshot', 'basis'):
                    bad[mutation] = {} if mutation == 'snapshot' else 'donor'
                elif mutation == 'null-template': bad['template'] = None
                elif mutation == 'bad-group': bad['template']['group'] = 'guess'
                elif mutation == 'bad-profile': bad['template']['profile'] = 'guess'
                elif mutation == 'null-baseline': bad['template']['baseline'] = None
                elif mutation == 'unknown-binding': bad['template']['query'] = v['destination']
                else:
                    child = next(x for x in bad['composition']['children'] if 'summary' in x)
                    if mutation == 'null-summary': child['summary'] = None
                    else: child['summary']['snapshot'] = {}
                assert not view_check.is_valid(bad), (suffix, mutation)
                controls += 1
    expansion_check = check('workbench.ViewTemplateExpansion')
    request_check = check('workbench.ViewTemplateExpansionRequestInput')
    for v in vectors:
        expansion_check.validate(v['expansion'])
        request = {'source': v['source'], 'query': v['destination']}
        request_check.validate(request)
        for mutation in ('null-source', 'null-query', 'mixed-arm', 'extra-basis'):
            bad = copy.deepcopy(request)
            if mutation == 'null-source': bad['source'] = None
            elif mutation == 'null-query': bad['query'] = None
            elif mutation == 'mixed-arm': bad['template'] = 'otel.http-server-routes'
            else: bad['basis'] = 'donor'
            assert not request_check.is_valid(bad), mutation
            controls += 1
    print(f'{len(vectors)} complete record-summary templates, expansions and input/output carriers; {controls} malformed-wire controls pass')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        if hasattr(error, 'absolute_path'):
            location = '/'.join(str(x) for x in error.absolute_path) or '<root>'
            print(f'{type(error).__name__}: {location}: schema rule {error.validator}', file=sys.stderr)
        else:
            print(f'{type(error).__name__}: {str(error)[:500]}', file=sys.stderr)
        raise SystemExit(1)

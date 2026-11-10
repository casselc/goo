"""Finite 398 corpus: API shape validation complements Go/JS semantic admission.

Byte budgets, control/surrogate rejection and nonblank titles are explicitly
runtime checks, not claims about JSON Schema's Unicode character maxLength.
"""
import copy
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / 'scripts'))
from check_openapi import validator


def main():
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    vectors = json.loads((ROOT / 'tests/fixtures/grouped-mean-398-wire.json').read_text())
    check = validator(spec, {'$ref': '#/components/schemas/workbench.TimelineViewDeclarationInput'})
    for vector in vectors:
        accepted = check.is_valid(vector['declaration'])
        assert accepted == vector['apiAccepted'], (vector['name'], accepted, list(check.iter_errors(vector['declaration'])))
    # Check the enclosing stored-view arm as well as the leaf declaration.
    base = next(vector for vector in vectors if vector['accepted'])
    declaration, question = copy.deepcopy(base['declaration']), copy.deepcopy(base['question'])
    question.setdefault('offset', 0)
    view = {'version': 1, 'title': 'Mean duration', 'queries': [question],
            'composition': {'version': 'goo.composition/1', 'kind': {'name': 'stack', 'version': 1},
                            'children': [{'timeline': declaration}, {'records': {
                                'version': 'goo.view/1', 'scope': {'kind': 'record-query', 'signal': 'spans'},
                                'kind': {'name': 'record-table', 'version': 1},
                                'display': {'columns': ['utc', 'service', 'summary']}}}]}}
    carriers = 0
    for suffix in ('', 'Input'):
        carrier = validator(spec, {'$ref': '#/components/schemas/workbench.ExplorationView' + suffix})
        for kind in ('small-multiples', 'table', 'heatmap'):
            accepted = copy.deepcopy(view)
            accepted['composition']['children'][0]['timeline']['kind']['name'] = kind
            carrier.validate(accepted)
            carriers += 1
        for change in ('missing-summary', 'count-summary', 'logs', 'count-profile'):
            bad = copy.deepcopy(view)
            if change == 'missing-summary':
                bad['queries'][0].pop('summary')
            elif change == 'count-summary':
                bad['queries'][0]['summary'] = ''
            elif change == 'logs':
                bad['queries'][0]['signal'] = 'logs'
            else:
                bad['composition']['children'][0]['timeline']['display'].update(profile='grouped-count-v1', value='count')
            assert not carrier.is_valid(bad), (suffix, change)
            carriers += 1
    print(f'{len(vectors)} grouped mean API shape vectors and {carriers} enclosing-view cases passed; semantic-only text restrictions remain runtime-owned')


if __name__ == '__main__':
    main()

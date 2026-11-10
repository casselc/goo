"""Finite summary leaves and enclosing-view schema admission.

Runtime checks own byte budgets, invalid Unicode/control text, actual query field
availability and source evidence; apiAccepted records only schema-shape claims.
"""
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / 'scripts'))
from check_openapi import validator


def main():
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    leaves = json.loads((ROOT / 'tests/fixtures/summary-view-400-wire.json').read_text())
    carriers = json.loads((ROOT / 'tests/fixtures/summary-view-400-carriers.json').read_text())
    for suffix in ('', 'Input'):
        leaf = validator(spec, {'$ref': '#/components/schemas/workbench.SummaryViewDeclaration' + suffix})
        for vector in leaves:
            accepted = leaf.is_valid(vector['declaration'])
            assert accepted == vector['apiAccepted'], (suffix, vector['name'], accepted, list(leaf.iter_errors(vector['declaration'])))
        carrier = validator(spec, {'$ref': '#/components/schemas/workbench.ExplorationView' + suffix})
        for vector in carriers:
            accepted = carrier.is_valid(vector['view'])
            assert accepted == vector['apiAccepted'], (suffix, vector['name'], accepted, list(carrier.iter_errors(vector['view'])))
    print(f'{len(leaves)} leaf and {len(carriers)} enclosing-view API cases pass for input/output shapes; runtime semantics remain separate')


if __name__ == '__main__':
    main()

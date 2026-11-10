"""Finite 397 corpus: API shape validation complements Go/JS semantic admission.

Byte budgets, control/surrogate rejection and nonblank titles are explicitly
runtime checks, not claims about JSON Schema's Unicode character maxLength.
"""
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / 'scripts'))
from check_openapi import validator


def main():
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    vectors = json.loads((ROOT / 'tests/fixtures/quantitative-view-397-wire.json').read_text())
    check = validator(spec, {'$ref': '#/components/schemas/workbench.ViewDeclarationInput'})
    for vector in vectors:
        accepted = check.is_valid(vector['declaration'])
        assert accepted == vector['apiAccepted'], (vector['name'], accepted, list(check.iter_errors(vector['declaration'])))
    print(f'{len(vectors)} quantitative API shape vectors passed; semantic-only text restrictions remain runtime-owned')


if __name__ == '__main__':
    main()

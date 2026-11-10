"""Existing full CapturePage carriers for retained gauge context.

The generated Go/JS property corpus is local semantic exploration. These checks
cover API shape only; expected-visit ID, stream/sample membership and custody hash
are independently admitted by runtime owners and observed-record verification.
"""
import copy
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT / 'scripts'))
from jsonschema import Draft202012Validator, FormatChecker


def main():
    corpus = Path(sys.argv[1]) if len(sys.argv) > 1 else ROOT / 'artifacts/verification/retained-gauge401-carriers.json'
    vectors = json.loads(corpus.read_text())
    spec = json.loads((ROOT / 'api/openapi.json').read_text())
    check = Draft202012Validator({'allOf': [{'$ref': '#/components/schemas/exploration.CapturePage'}], 'components': spec['components']}, format_checker=FormatChecker())
    controls = 0
    for vector in vectors:
        page = vector['page']
        check.validate(page)
        for kind in ('null-measurement', 'mixed-arm', 'bad-clock', 'bad-receipt-id', 'null-receipt'):
            bad = copy.deepcopy(page)
            if kind == 'null-measurement':
                bad['measurement'] = None
            elif kind == 'mixed-arm':
                bad['explore'] = {}
            elif kind == 'bad-clock':
                bad['receipt']['capturedAt'] = 'not-a-clock'
            elif kind == 'bad-receipt-id':
                bad['receipt']['id'] = 'bad'
            else:
                bad['receipt'] = None
            assert not check.is_valid(bad), kind
            controls += 1
    print(f'{len(vectors)} generated retained gauge full API carriers and {controls} malformed-envelope shape controls pass')


if __name__ == '__main__':
    main()

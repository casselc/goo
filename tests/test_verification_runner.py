"""Negative controls for the required verification verdict boundary."""
import importlib.util
import json
from pathlib import Path
import unittest
from unittest import mock
import tempfile

spec = importlib.util.spec_from_file_location('verify_optional', Path(__file__).resolve().parents[1] / 'scripts/verify_optional.py')
runner = importlib.util.module_from_spec(spec)
spec.loader.exec_module(runner)


def events(*rows):
    return '\n'.join(json.dumps({'Package': 'pkg', **row}) for row in rows) + '\n'


class RequiredVerdict(unittest.TestCase):
    def test_exact_required_pass(self):
        runner.validate_events(events({'Action': 'run', 'Test': 'Required'}, {'Action': 'pass', 'Test': 'Required'}, {'Action': 'pass', 'Package': 'pkg'}), ['Required'])

    def test_empty_or_misspelled_selection_fails(self):
        for output in ['', events({'Action': 'pass', 'Test': 'Other'}), events({'Action': 'pass', 'Package': 'pkg'})]:
            with self.subTest(output=output), self.assertRaises(RuntimeError):
                runner.validate_events(output, ['Required'])

    def test_skip_or_failure_cannot_be_hidden_by_parent_pass(self):
        for action in ['skip', 'fail']:
            with self.subTest(action=action), self.assertRaises(RuntimeError):
                runner.validate_events(events({'Action': action, 'Test': 'Required/child'}, {'Action': 'pass', 'Test': 'Required'}), ['Required'])

    def test_non_json_fails(self):
        with self.assertRaises(RuntimeError):
            runner.validate_events('not a test event', ['Required'])

    def test_shared_protocol_cases(self):
        cases = json.loads((Path(__file__).parent / 'fixtures/go-test-verdicts.json').read_text())
        for case in cases:
            with self.subTest(case=case['name']):
                if case['accept']:
                    runner.validate_events(case['output'], case['expected'])
                else:
                    with self.assertRaises(RuntimeError):
                        runner.validate_events(case['output'], case['expected'])

    def test_missing_quint_has_install_guidance(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            (root / 'verification').mkdir()
            (root / 'verification/toolchain.json').write_text(json.dumps({'node': '24.19.0'}))
            (root / '.go-version').write_text('1.27.2')
            with mock.patch.object(runner, 'ROOT', root), mock.patch.object(runner, 'command', side_effect=['go1.27.2', 'v24.19.0']):
                with self.assertRaisesRegex(RuntimeError, 'Quint missing: run npm ci'):
                    runner.preflight()

    def test_pinned_sibling_mismatch_refused(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / 'goo'
            (root / 'verification').mkdir(parents=True)
            (root / 'verification/toolchain.json').write_text(json.dumps({'node': '24.19.0', 'observabilityCommit': 'pinned'}))
            (root / '.go-version').write_text('1.27.2')
            (root / 'package.json').write_text(json.dumps({'devDependencies': {'@informalsystems/quint': '0.32.0'}}))
            (root / 'node_modules/.bin').mkdir(parents=True)
            (root / 'node_modules/.bin/quint').touch()
            (root.parent / 'observability/quintgo').mkdir(parents=True)
            (root.parent / 'observability/quintgo/go.mod').touch()
            with mock.patch.object(runner, 'ROOT', root), mock.patch.object(runner, 'command', side_effect=['go1.27.2', 'v24.19.0', '0.32.0', 'other']):
                with self.assertRaisesRegex(RuntimeError, 'observability commit: expected pinned, found other'):
                    runner.preflight()

    def test_tool_version_mismatch_fails(self):
        with self.assertRaisesRegex(RuntimeError, 'expected 1, found 2'):
            runner.require_version('tool', '2', '1')


if __name__ == '__main__':
    unittest.main()

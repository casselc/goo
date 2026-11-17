"""Exercise the real shell boundary with unavailable lab dependencies."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

REPO = Path(__file__).resolve().parents[1]

class JourneyRunner(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        for folder in ['scripts', 'node_modules/.bin', 'bin']:
            (self.root / folder).mkdir(parents=True)
        for name in ['run_journeys.sh', 'run_reservation_journeys.sh']:
            shutil.copy(REPO / 'scripts' / name, self.root / 'scripts' / name)
        self.log = self.root / 'calls'
        self.env = {**os.environ, 'PATH': str(self.root / 'bin') + ':' + os.environ['PATH'], 'CALLS': str(self.log)}
        for name in ['docker', 'python3']:
            self.executable(self.root / 'bin' / name, '#!/bin/bash\necho forbidden >> "$CALLS"\nexit 91\n')
        self.executable(self.root / 'node_modules/.bin/playwright', '#!/bin/bash\nprintf "%s\\n" "$@" >> "$CALLS"\nexit 7\n')

    def executable(self, path, text):
        path.write_text(text)
        path.chmod(0o755)

    def run_wrapper(self, *args):
        return subprocess.run(['bash', str(self.root / 'scripts/run_journeys.sh'), *args], env=self.env, capture_output=True, text=True)

    def test_generic_forwards_exact_arguments_and_failure_without_lab_or_pruning(self):
        result = self.run_wrapper('journeys/a file.spec.js', '--grep', 'a b')
        self.assertEqual(result.returncode, 7, result.stderr)
        self.assertEqual(self.log.read_text().splitlines(), ['test', 'journeys/a file.spec.js', '--grep', 'a b'])

    def test_help_and_missing_selection_have_no_side_effects(self):
        self.assertEqual(self.run_wrapper('--help').returncode, 0)
        self.assertEqual(self.run_wrapper().returncode, 2)
        self.assertEqual(self.run_wrapper('--smoke').returncode, 2)
        self.assertFalse(self.log.exists())

    def test_reservation_is_an_explicit_dispatch_with_unchanged_arguments(self):
        self.executable(self.root / 'scripts/run_reservation_journeys.sh', '#!/bin/bash\nprintf "%s\\n" "$@" >> "$CALLS"\nexit 8\n')
        result = self.run_wrapper('--reservation', '--smoke', 'chosen')
        self.assertEqual(result.returncode, 8)
        self.assertEqual(self.log.read_text().splitlines(), ['--smoke', 'chosen'])

if __name__ == '__main__':
    unittest.main()

import gzip
import importlib.util
import json
import os
from pathlib import Path
import tempfile
import subprocess
import unittest

spec = importlib.util.spec_from_file_location('retention', Path(__file__).parents[1] / 'scripts/retain_diagnostics.py')
r = importlib.util.module_from_spec(spec)
spec.loader.exec_module(r)


class DiagnosticRetention(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.now = 1_800_000_000

    def put(self, relative, age=100, content='diagnostic'):
        p = self.root / relative
        p.parent.mkdir(parents=True, exist_ok=True)
        if relative.endswith('trace.jsonl') and content == 'diagnostic':
            content = json.dumps(dict(timestamp=0,action=None,state={},snapshots=[],violations=[]))+'\n'
        p.write_text(content)
        os.utime(p, (self.now-age*3600, self.now-age*3600))
        return p

    def apply(self, plans):
        audit = self.root / 'audit.jsonl.gz'
        result = r.apply(self.root, plans, audit)
        return result, [json.loads(line) for line in gzip.open(audit, 'rt')]

    def test_expiry_preserves_compact_evidence_and_recent_failures(self):
        old = self.put('legacy/run/trace.jsonl')
        frame = self.put('legacy/run/screenshots/a.jpeg')
        ledger = self.put('legacy/run/ledger.json')
        model = self.put('legacy/run/model-trace.jsonl')
        ambiguous = self.put('model/trace.jsonl', content='{}\n')
        log = self.put('legacy/run/process.log')
        failure = self.put('failure/trace.zip')
        self.put('failure/run-result.json', content='{"status":"failed"}')
        plans = r.inventory(self.root, self.now)
        self.assertTrue(old.exists())  # planning is read-only
        result, audit = self.apply(plans)
        self.assertEqual(result['removed_files'], 2)
        self.assertFalse(old.exists())
        self.assertFalse(frame.exists())
        for p in [ledger, model, log, failure, ambiguous]: self.assertTrue(p.exists())
        self.assertEqual({a['path'] for a in audit}, {'legacy/run/trace.jsonl', 'legacy/run/screenshots/a.jpeg'})

    def test_pins_active_runs_fresh_logs_and_symlink_targets_survive(self):
        for name, marker in [('pinned','.keep'), ('active','.active')]:
            self.put(name+'/trace.zip')
            self.put(name+'/'+marker)
        self.put('fresh/trace.zip')
        self.put('fresh/process.log', age=0)
        outside = self.put('outside.bin')
        (self.root/'linked.png').symlink_to(outside)
        (self.root/'linked-directory').symlink_to(self.root/'pinned', target_is_directory=True)
        result, _ = self.apply(r.inventory(self.root, self.now))
        self.assertEqual(result['removed_files'], 0)
        self.assertTrue(outside.exists())

    def test_changed_bundle_and_new_pin_are_rechecked(self):
        trace = self.put('run/trace.zip')
        plans = r.inventory(self.root, self.now)
        self.put('run/process.log', age=0)
        self.assertEqual(self.apply(plans)[0]['removed_files'], 0)
        self.assertTrue(trace.exists())
        plans = r.inventory(self.root, self.now+200*3600)
        self.put('run/.keep')
        self.assertEqual(self.apply(plans)[0]['removed_files'], 0)

    def test_hardlinks_report_only_reclaimable_blocks(self):
        a = self.put('old/trace.zip')
        b = self.root/'new/trace.zip'; b.parent.mkdir()
        os.link(a,b)
        self.put('new/.keep')
        plans = r.inventory(self.root,self.now)
        self.assertEqual(r.summary(plans)['reclaimable_allocated_bytes'],0)
        self.apply(plans)
        self.assertTrue(b.exists())
        self.assertFalse(a.exists())

    def test_budget_removes_quiet_success_but_protects_failures(self):
        self.put('success/trace.zip', age=3)
        self.put('success/run-result.json', age=3, content='{"status":"passed"}')
        self.put('failure/trace.zip', age=3)
        self.put('failure/run-result.json', age=3, content='{"status":"failed"}')
        budget = r.BUDGET_BYTES
        try:
            r.BUDGET_BYTES=0
            plans=r.inventory(self.root,self.now)
            self.assertEqual([p['owner'].name for p in plans if p['remove']],['success'])
            self.assertTrue(r.summary(plans)['budget_exceeded'])
        finally: r.BUDGET_BYTES=budget

    def test_outcome_deadlines(self):
        for name, status, age in [('pass', 'passed', 24), ('young', 'passed', 23),
                                  ('failure', 'failed', 167), ('old-failure', 'failed', 169),
                                  ('legacy', 'unknown', 72)]:
            self.put(name+'/trace.jsonl', age=age)
            if status != 'unknown':
                self.put(name+'/run-result.json', age=age, content=json.dumps({'status':status}))
        removed={p['owner'].name for p in r.inventory(self.root,self.now) if p['remove']}
        self.assertEqual(removed, {'pass','old-failure','legacy'})

    def test_wrapper_restores_traffic_and_preserves_test_failure(self):
        binary = self.root/'bin'
        binary.mkdir()
        log = self.root/'wrapper.log'
        commands = {
            'docker': '#!/bin/bash\nprintf "%s\\n" "$*" >> "$WRAPPER_LOG"\nif [[ "$*" == "compose ps --status running --services" ]]; then echo traffic; fi\n',
            'python3': '#!/bin/bash\nif [[ "$1" == "-" ]]; then cat >/dev/null; exit 0; fi\nif [[ "$1" == "scripts/smoke.py" ]]; then exit 7; fi\necho cleanup >> "$WRAPPER_LOG"\nexit 3\n',
        }
        for name, content in commands.items():
            path=binary/name;path.write_text(content);path.chmod(0o755)
        result=subprocess.run(['bash',str(r.REPO/'scripts/run_journeys.sh'),'--smoke'],
                              env={**os.environ,'PATH':str(binary)+':'+os.environ['PATH'],
                                   'WRAPPER_LOG':str(log)},capture_output=True,text=True)
        self.assertEqual(result.returncode,7,result.stderr)
        self.assertIn('test exit status is preserved',result.stderr)
        self.assertEqual(log.read_text().splitlines()[-3:],
                         ['compose stop traffic','compose start traffic','cleanup'])


if __name__ == '__main__':
    unittest.main()

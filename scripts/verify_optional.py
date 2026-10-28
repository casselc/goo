#!/usr/bin/env python3
"""Bounded opt-in verification; no installation, checkout edits or lab reset."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
SUITES = {
    "offline": [
        ("instrumentation", ".", ["TestRuntimeInitializationErrors", "TestOwnedRuntimesConcurrentAccounting", "TestOutcomeInitializationBound", "TestDefaultBinding"]),
        ("verification", ".", ["TestOwnedRuntimeSDKIsolation", "TestHegelOperationAccountingStateSwarm", "TestHegelExactDeadlineAndDuplicateDeliveries", "TestQuintModelTracesDriveRealPreparation", "TestQuintValidatesObservedTraceAndRejectsMutation"]),
    ],
    "woven": [("verification/weaving", "./scenario", ["TestWovenRealHTTPPreparation", "TestGooAnnotationsEmitActualOTel"])],
}


def command(args, cwd=ROOT, env=None, strip=True):
    result = subprocess.run(args, cwd=cwd, env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    if result.returncode:
        raise RuntimeError(f"{' '.join(args)} failed:\n{result.stdout}{result.stderr}")
    return result.stdout.strip() if strip else result.stdout


def require_version(label, actual, expected):
    if actual != expected:
        raise RuntimeError(f"{label}: expected {expected}, found {actual}")


def preflight():
    pins = json.loads((ROOT / 'verification/toolchain.json').read_text())
    env = dict(os.environ, GOWORK='off', GOTOOLCHAIN='local', QUINTGO_BACKEND='typescript')
    require_version('Go', command(['go', 'env', 'GOVERSION'], env=env), 'go' + (ROOT / '.go-version').read_text().strip())
    require_version('Node', command(['node', '--version'], env=env), 'v' + pins['node'])
    # Both direct tests and quintgo must resolve this same pinned executable.
    quint = (ROOT / 'node_modules/.bin/quint').absolute()
    if not quint.is_file():
        raise RuntimeError('Quint missing: run npm ci in the repository root')
    version = json.loads((ROOT / 'package.json').read_text())['devDependencies']['@informalsystems/quint']
    require_version('Quint', command([str(quint), '--version'], env=env), version)
    env['QUINT_BIN'] = str(quint)
    env['PATH'] = str(quint.parent) + os.pathsep + env.get('PATH', '')
    sibling = ROOT.parent / 'observability'
    if not (sibling / 'quintgo/go.mod').is_file():
        raise RuntimeError('Missing ../observability/quintgo; see verification/README.md#reproducible-entry-point')
    require_version('observability commit', command(['git', 'rev-parse', 'HEAD'], cwd=sibling), pins['observabilityCommit'])
    if command(['git', 'status', '--porcelain', '--untracked-files=all', '--', 'quintgo'], cwd=sibling):
        raise RuntimeError('Sibling quintgo has local changes; use a clean pinned checkout')
    print(f"Prerequisites: Go, Node, Quint {version}, pinned quintgo verified", flush=True)
    return env


def validate_events(output, expected):
    """Require every named test to pass, and refuse any skipped subtest."""
    if not output.endswith('\n'):
        raise RuntimeError('Incomplete Go test event stream')
    passed, ran, packages, skipped, failed = {}, set(), set(), [], []
    actions = {'start', 'run', 'pause', 'cont', 'pass', 'bench', 'fail', 'output', 'skip', 'build-output', 'build-fail'}
    for line in output.splitlines():
        try:
            event = json.loads(line)
        except json.JSONDecodeError as exc:
            raise RuntimeError('Non-JSON output in Go test event stream: ' + line[:200]) from exc
        if not isinstance(event, dict) or not isinstance(event.get('Action'), str) or event['Action'] not in actions:
            raise RuntimeError('Malformed Go test event')
        action, name, package = event.get('Action'), event.get('Test'), event.get('Package')
        if action in ('run', 'pass') and (not isinstance(package, str) or not package or ('Test' in event and (not isinstance(name, str) or not name))):
            raise RuntimeError('Malformed Go test verdict')
        if action == 'run' and name:
            ran.add((package, name))
        if action == 'skip':
            skipped.append(name or event.get('Package', 'package'))
        elif action in ('fail', 'build-fail'):
            failed.append(name or event.get('Package', 'package'))
        elif action == 'pass':
            if name:
                passed[name] = package
            else:
                packages.add(package)
    missing = {name for name in expected if name not in passed or (passed[name], name) not in ran or passed[name] not in packages}
    if skipped or failed or missing:
        raise RuntimeError(f"Required verdict missing: skipped={skipped}, failed={failed}, missing={sorted(missing)}")


def run_suite(suite, env):
    for directory, package, tests in SUITES[suite]:
        prefix = ['go', 'run', '-mod=readonly', 'github.com/DataDog/orchestrion', 'go'] if suite == 'woven' else ['go']
        args = prefix + ['test', '-mod=readonly', '-json', '-count=1', '-timeout=5m', '-run', '^(' + '|'.join(tests) + ')$', package]
        print(f"Running {suite}: {directory} ({len(tests)} required tests)", flush=True)
        output = command(args, cwd=ROOT / directory, env=env, strip=False)
        try:
            validate_events(output, tests)
        except RuntimeError:
            print(output, file=sys.stderr)
            raise
        print(f"Passed: {', '.join(tests)}", flush=True)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--suite', choices=[*SUITES, 'lab'], default='offline')
    parser.add_argument('--check', action='store_true', help='check prerequisites without executing tests')
    args = parser.parse_args()
    try:
        if args.suite == 'lab':
            # Uses this Python's requirements-api.txt dependencies; never starts or resets Compose.
            for module in ['jsonschema', 'openapi_spec_validator']:
                __import__(module)
            if not args.check:
                print(command([sys.executable, str(ROOT / 'scripts/check_deployed_api.py')]))
        else:
            env = preflight()
            if not args.check:
                run_suite(args.suite, env)
    except (RuntimeError, OSError, ImportError) as exc:
        print(f'Verification refused: {exc}', file=sys.stderr)
        return 1
    return 0


if __name__ == '__main__':
    sys.exit(main())

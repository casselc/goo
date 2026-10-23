#!/usr/bin/env python3
"""Expire browser diagnostics only; dry-run unless --apply is supplied."""
import argparse
import collections
import gzip
import json
import os
from pathlib import Path
import stat
import time
import uuid

REPO = Path(__file__).resolve().parents[1]
HOURS = {'passed': 24, 'failed': 168, 'unknown': 72}
BUDGET_BYTES = 512 * 1024 * 1024
RAW_NAMES = {'trace.jsonl', 'trace.zip'}
RAW_SUFFIXES = {'.png', '.jpg', '.jpeg', '.webm'}


def files(root):
    for directory, dirs, names in os.walk(root, followlinks=False):
        dirs[:] = sorted(d for d in dirs if not (Path(directory) / d).is_symlink()
                         and d != 'retention')
        for name in sorted(names):
            path = Path(directory) / name
            info = path.lstat()
            if stat.S_ISREG(info.st_mode):
                yield path, info


def fingerprint(info):
    return (info.st_dev, info.st_ino, info.st_size, info.st_mtime_ns)


def raw_recording(path):
    if path.suffix.lower() in RAW_SUFFIXES or path.name == 'trace.zip':
        return True
    if path.name != 'trace.jsonl':
        return False
    # Do not confuse model/action ledgers with Bombadil browser recordings.
    try:
        with path.open() as stream:
            first = json.loads(stream.readline(1024 * 1024))
        return isinstance(first, dict) and {'timestamp', 'action', 'state', 'snapshots', 'violations'} <= first.keys()
    except (OSError, ValueError, UnicodeError):
        return False


def inventory(root, now=None):
    now = time.time() if now is None else now
    entries = list(files(root))
    # A trace's directory owns its frames, even when nested under screenshots/.
    owners = {p.parent for p, _ in entries if p.name in RAW_NAMES}
    owners.update(p.parent for p, _ in entries if p.name == 'run-result.json')
    groups = collections.defaultdict(list)
    for path, info in entries:
        if not raw_recording(path):
            continue
        owner = next((p for p in path.parents if p in owners), None)
        if owner is None:
            # Legacy Playwright attachments: siblings form one diagnostic bundle.
            owner = path.parent.parent if path.parent.name == 'screenshots' else path.parent
        groups[owner].append((path, info))
    newest_by_dir = collections.defaultdict(float)
    for path, info in entries:
        for directory in path.parents:
            if directory == root or directory.is_relative_to(root):
                newest_by_dir[directory] = max(newest_by_dir[directory], info.st_mtime)
            else:
                break
    plans = []
    for owner, raw in sorted(groups.items()):
        # Fresh logs/ledgers also protect an otherwise old recording.
        newest = newest_by_dir[owner]
        ancestors = [owner, *owner.parents]
        pinned = any((p / '.keep').exists() or (p / '.active').exists()
                     for p in ancestors if p == root or p.is_relative_to(root))
        outcome = 'unknown'
        try:
            result_file = owner / 'run-result.json'
            if result_file.is_symlink():
                raise ValueError('Symlink outcome is not trusted')
            result = json.loads(result_file.read_text())
            if result.get('status') in ('passed', 'failed'):
                outcome = result['status']
        except (OSError, ValueError, AttributeError):
            pass
        expired = now - newest >= HOURS[outcome] * 3600
        plans.append(dict(owner=owner, files=raw, status=outcome, newest=newest,
                          remove=expired and not pinned, pinned=pinned, reason='expired'))
    # Budget pressure may expire successful recordings early, after two quiet
    # hours. Recent failures/unknown runs and pins take precedence over the cap.
    eligible = [p for p in plans if not p['remove'] and not p['pinned']
                and p['status'] == 'passed' and now-p['newest'] >= 7200]
    for plan in sorted(eligible, key=lambda p: p['newest']):
        if summary(plans)['retained_allocated_bytes'] <= BUDGET_BYTES:
            break
        plan.update(remove=True, reason='budget')
    return plans


def summary(plans):
    selected = [p for p in plans if p['remove']]
    all_links = collections.Counter()
    sizes = {}
    selected_links = collections.Counter()
    retained = set()
    for plan in plans:
        for _, info in plan['files']:
            key = (info.st_dev, info.st_ino)
            all_links[key] = info.st_nlink
            sizes[key] = info.st_blocks * 512
            if plan['remove']:
                selected_links[key] += 1
            else:
                retained.add(key)
    return dict(bundles=len(plans), expired_bundles=len(selected),
                files=sum(len(p['files']) for p in selected),
                logical_bytes=sum(s.st_size for p in selected for _, s in p['files']),
                reclaimable_allocated_bytes=sum(sizes[k] for k, n in selected_links.items()
                                                if n == all_links[k]),
                pinned_bundles=sum(p['pinned'] for p in plans),
                budget_bytes=BUDGET_BYTES, retained_allocated_bytes=sum(sizes[k] for k in retained),
                budget_exceeded=sum(sizes[k] for k in retained) > BUDGET_BYTES)


def apply(root, plans, audit):
    removed = skipped = 0
    with gzip.open(audit, 'wt') as log:
        for plan in plans:
            if not plan['remove']:
                continue
            owner = plan['owner']
            # Revalidate freshness/pins before removing a bundle.
            ancestors = [owner, *owner.parents]
            safe = not any(p.is_symlink() or (p / '.keep').exists() or (p / '.active').exists()
                           for p in ancestors if p == root or p.is_relative_to(root))
            if safe:
                safe = all(s.st_mtime <= plan['newest'] for _, s in files(owner))
            for path, before in plan['files']:
                try:
                    unchanged = fingerprint(path.lstat()) == fingerprint(before)
                    parents_safe = not any(p.is_symlink() for p in path.parents
                                           if p == root or p.is_relative_to(root))
                    if not safe or not unchanged or not parents_safe:
                        skipped += 1
                        continue
                    path.unlink()
                    removed += 1
                    log.write(json.dumps(dict(path=str(path.relative_to(root)),
                                              bytes=before.st_size, mtime_ns=before.st_mtime_ns,
                                              status=plan['status'], reason=plan['reason'])) + '\n')
                except FileNotFoundError:
                    skipped += 1
    return dict(removed_files=removed, skipped_files=skipped, audit=str(audit))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--apply', action='store_true', help='delete expired browser recordings')
    args = parser.parse_args()
    # Deliberately no arbitrary --root option: application stores are out of scope.
    root = (REPO / 'artifacts').resolve()
    if root == REPO or REPO.is_relative_to(root) or root.exists() and not root.is_dir():
        parser.error('Expected the dedicated artifacts directory or its symlink target.')
    if not root.exists():
        print(json.dumps(dict(mode='apply' if args.apply else 'dry-run', root=str(root), **summary([])), indent=2))
        return
    import fcntl
    lock_fd = os.open(root / '.retention.lock', os.O_CREAT | os.O_RDWR | os.O_NOFOLLOW, 0o600)
    with os.fdopen(lock_fd, 'r+') as lock:
        try:
            fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
        except BlockingIOError:
            parser.exit(1, 'Another diagnostic retention command is running.\n')
        plans = inventory(root)
        result = dict(mode='apply' if args.apply else 'dry-run', root=str(root), **summary(plans))
        if args.apply:
            audits = root / 'retention'
            audits.mkdir(exist_ok=True)
            if audits.is_symlink():
                parser.error('Audit directory must not be a symlink.')
            audit = audits / (time.strftime('%Y%m%dT%H%M%SZ', time.gmtime()) + '-' + uuid.uuid4().hex[:8] + '.jsonl.gz')
            result.update(apply(root, plans, audit))
        print(json.dumps(result, indent=2))


if __name__ == '__main__':
    main()

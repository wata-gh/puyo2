#!/usr/bin/env python3
"""Publish a verified patch without executing candidate code. Actions-only."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import tempfile
import time

from maintenance import (ROOT, Failure, archive_files, command, existing_pr,
                         github_get, http, validate_diff)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--input', type=Path, required=True)
    args = parser.parse_args()
    if os.environ.get('GITHUB_ACTIONS') != 'true' or os.environ.get('GITHUB_REPOSITORY') != 'wata-gh/puyo2':
        raise Failure('publication is restricted to puyo2 GitHub Actions')
    report = json.loads((args.input / 'report.json').read_text())
    if report['status'] in ('no_updates', 'skipped_existing_pr'):
        print(report['status'])
        return
    if report['status'] != 'verified' or report['mode'] != 'publish':
        raise Failure('not a verified publication run')
    if report['run'] != os.environ['GITHUB_RUN_ID']:
        raise Failure('artifact belongs to another workflow run')
    patch = (args.input / 'update.patch').read_bytes()
    if hashlib.sha256(patch).hexdigest() != report['patch_sha256']:
        raise Failure('patch digest mismatch')
    config = json.loads((ROOT / 'ops/maintenance/config.json').read_text())
    if len(patch) > config['max_patch_bytes']:
        raise Failure('patch size limit')
    kind = report['kind']
    if kind not in ('dependencies', 'rust'):
        raise Failure('unexpected kind')
    if existing_pr(kind):
        print('Existing maintenance PR; no overwrite')
        return
    base = report['base_sha']
    if not re.fullmatch('[0-9a-f]{40}', base):
        raise Failure('invalid base SHA')
    if github_get('git/ref/heads/main')['object']['sha'] != base:
        raise Failure('main advanced; rerun maintenance against current main')
    branch = 'codex/maintenance-' + kind
    refs = github_get('git/matching-refs/heads/' + branch)
    if any(r['ref'] == 'refs/heads/' + branch for r in refs):
        raise Failure('orphan update branch exists; inspect it before deleting or opening a PR')
    before = archive_files(command(['git', 'archive', base], cwd=ROOT), config['max_archive_bytes'])
    with tempfile.TemporaryDirectory(prefix='puyo2-publish-') as tmp:
        command(['git', 'clone', '--no-hardlinks', str(ROOT), tmp])
        command(['git', 'checkout', '--detach', base], cwd=tmp)
        command(['git', 'apply', '--check', '-'], cwd=tmp, data=patch)
        command(['git', 'apply', '--index', '-'], cwd=tmp, data=patch)
        tree = command(['git', 'write-tree'], cwd=tmp).decode().strip()
        after = archive_files(command(['git', 'archive', tree], cwd=tmp), config['max_archive_bytes'])
        changed = validate_diff(before, after, kind)
        if not changed:
            print('No changes')
            return
        if changed != report['changed'] or not report.get('validation') or report['validation'][-1]['result'] != 'passed':
            raise Failure('validation evidence mismatch')
        command(['git', '-c', 'user.name=puyo2-maintenance[bot]', '-c', 'user.email=maintenance@users.noreply.github.com',
                 'commit', '--no-gpg-sign', '-qm', 'Update ' + kind], cwd=tmp)
        sha = command(['git', 'rev-parse', 'HEAD'], cwd=tmp).decode().strip()
        # gh authenticates through GH_TOKEN; no credential file or token URL.
        # Empty expected ref means create-only, including races with human pushes.
        command(['git', '-c', 'credential.helper=', '-c', 'credential.helper=!gh auth git-credential',
                 'push', '--force-with-lease=refs/heads/' + branch + ':',
                 'https://github.com/wata-gh/puyo2.git', 'HEAD:refs/heads/' + branch], cwd=tmp)
    url = f'https://github.com/wata-gh/puyo2/actions/runs/{report["run"]}'
    body = '\n'.join([
        'Automated, verified ' + kind + ' update. Human review and merge are required.', '',
        'Candidates:', *['- ' + n for n in report['candidates']], '',
        'Validation: pinned Rust and unchanged MSRV build/test; release binaries; pnsolve levels 1–5; cargo package/install.',
        'Existing tests, edition, MSRV, dependency features and workflows are protected.',
        f'Compatibility repair turns: {report.get("repairs", 0)}. Observed tokens (best effort): {report.get("observed_tokens", "not used")}.',
        'puyo-rsrch-engine was not executed or modified. Review public API and downstream compatibility.',
        f'[Controller logs and report]({url}). Normal Rust CI must pass for commit `{sha}`.',
        'CI is pending at PR creation; this draft is not a claim of merge readiness.',
    ])
    pr = http('https://api.github.com/repos/wata-gh/puyo2/pulls', method='POST',
              key=os.environ['GH_TOKEN'], body={'title': 'Update ' + kind, 'head': branch, 'base': 'main', 'body': body, 'draft': True})
    print('Created draft: ' + pr['html_url'])
    # Verify that normal pull_request CI actually ran on the exact published head.
    deadline = time.monotonic() + 1800
    while time.monotonic() < deadline:
        runs = github_get('actions/workflows/rust.yml/runs?event=pull_request&head_sha=' + sha)['workflow_runs']
        matching = [r for r in runs if r['head_sha'] == sha]
        if matching:
            latest = max(matching, key=lambda r: r['id'])
            if latest['status'] == 'completed':
                if latest['conclusion'] != 'success':
                    raise Failure('normal Rust CI did not succeed: ' + latest['html_url'])
                http('https://api.github.com/repos/wata-gh/puyo2/pulls/' + str(pr['number']),
                     method='PATCH', key=os.environ['GH_TOKEN'],
                     body={'body': body + '\n\nVerified normal Rust CI: ' + latest['html_url']})
                print('Verified normal Rust CI: ' + latest['html_url'])
                return
        time.sleep(20)
    raise Failure('normal Rust CI missing or incomplete after 30 minutes; PR remains draft')


if __name__ == '__main__':
    main()
